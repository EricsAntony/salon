package service

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"user-service/internal/auth"
	"user-service/internal/config"
	"user-service/internal/model"
	"user-service/internal/repository"
)

type UserService interface {
	RequestOTP(ctx context.Context, phone string) error
	Register(ctx context.Context, phone, name, gender string, email, location *string, otp string) (*model.User, string, string, error)
	Authenticate(ctx context.Context, phone, otp string) (string, string, error)
	GetUser(ctx context.Context, id string) (*model.User, error)
	Refresh(ctx context.Context, refreshToken string) (string, string, error)
	Revoke(ctx context.Context, userID string) error
	UpdateUser(ctx context.Context, userID string, name, gender, email, location *string) (*model.User, error)
	DeleteUser(ctx context.Context, requesterID, targetID string) error
}

type userService struct {
	users  repository.UserRepository
	otps   repository.OTPRepository
	tokens repository.TokenRepository
	jwt    *auth.JWTManager
	cfg    *config.Config
}

func NewUserService(u repository.UserRepository, o repository.OTPRepository, t repository.TokenRepository, jwt *auth.JWTManager, cfg *config.Config) UserService {
	return &userService{users: u, otps: o, tokens: t, jwt: jwt, cfg: cfg}
}

func normalizePhone(phone string) string {
	p := strings.TrimSpace(phone)
	p = strings.ReplaceAll(p, " ", "")
	p = strings.ReplaceAll(p, "-", "")
	return p
}

func (s *userService) RequestOTP(ctx context.Context, phone string) error {
	phone = normalizePhone(phone)
	if phone == "" {
		return errors.New("phone required")
	}
	code := fmt.Sprintf("%06d", rand.Intn(1000000))
	exp := time.Now().Add(time.Duration(s.cfg.OTP.ExpiryMinutes) * time.Minute)
	hash := auth.HashString(code)
	if err := s.otps.Create(ctx, phone, hash, exp); err != nil {
		return err
	}
	auth.SendOTPViaSMS(phone, code)
	return nil
}

func (s *userService) verifyOTP(ctx context.Context, phone, otp string) error {
	rec, err := s.otps.GetLatest(ctx, phone)
	if err != nil {
		return err
	}
	if rec == nil {
		return errors.New("no otp requested")
	}
	if time.Now().After(rec.ExpiresAt) {
		return errors.New("otp expired")
	}
	if auth.HashString(otp) != rec.CodeHash {
		_ = s.otps.IncrementAttempts(ctx, rec.ID)
		return errors.New("invalid otp")
	}
	return nil
}

func (s *userService) Register(ctx context.Context, phone, name, gender string, email, location *string, otp string) (*model.User, string, string, error) {
	phone = normalizePhone(phone)
	if err := s.verifyOTP(ctx, phone, otp); err != nil {
		return nil, "", "", err
	}
	existing, err := s.users.GetByPhone(ctx, phone)
	if err != nil {
		return nil, "", "", err
	}
	if existing != nil {
		return nil, "", "", errors.New("user already exists; please authenticate")
	}
	// Create user
	uid := uuid.NewString()
	u := &model.User{ID: uid, PhoneNumber: phone, Name: name, Gender: model.Gender(strings.ToLower(gender)), Email: email, Location: location}
	if u.Gender != model.GenderMale && u.Gender != model.GenderFemale && u.Gender != model.GenderOther {
		return nil, "", "", errors.New("invalid gender")
	}
	if err := s.users.Create(ctx, u); err != nil {
		return nil, "", "", err
	}
	// Tokens
	access, _, err := s.jwt.GenerateAccessToken(uid)
	if err != nil { return nil, "", "", err }
	refresh, rexp, err := s.jwt.GenerateRefreshToken(uid)
	if err != nil { return nil, "", "", err }
	// Revoke any existing refresh tokens for this user just in case
	if err := s.tokens.RevokeAllForUser(ctx, uid); err != nil { return nil, "", "", err }
	if err := s.tokens.Save(ctx, uid, auth.HashString(refresh), rexp); err != nil { return nil, "", "", err }
	return u, access, refresh, nil
}

func (s *userService) Authenticate(ctx context.Context, phone, otp string) (string, string, error) {
	phone = normalizePhone(phone)
	if err := s.verifyOTP(ctx, phone, otp); err != nil {
		return "", "", err
	}
	u, err := s.users.GetByPhone(ctx, phone)
	if err != nil { return "", "", err }
	if u == nil { return "", "", errors.New("user not found; please register") }
	access, _, err := s.jwt.GenerateAccessToken(u.ID)
	if err != nil { return "", "", err }
	refresh, rexp, err := s.jwt.GenerateRefreshToken(u.ID)
	if err != nil { return "", "", err }
	// Revoke existing refresh tokens for this user upon new authentication
	if err := s.tokens.RevokeAllForUser(ctx, u.ID); err != nil { return "", "", err }
	if err := s.tokens.Save(ctx, u.ID, auth.HashString(refresh), rexp); err != nil { return "", "", err }
	return access, refresh, nil
}

func (s *userService) GetUser(ctx context.Context, id string) (*model.User, error) {
	u, err := s.users.GetByID(ctx, id)
	if err != nil { return nil, err }
	if u == nil { return nil, errors.New("not found") }
	return u, nil
}

func (s *userService) Refresh(ctx context.Context, refreshToken string) (string, string, error) {
	claims, err := s.jwt.ValidateRefreshToken(refreshToken)
	if err != nil { return "", "", err }
	uid := claims.UserID
	// validate against DB
	valid, err := s.tokens.IsValid(ctx, uid, auth.HashString(refreshToken), time.Now())
	if err != nil { return "", "", err }
	if !valid { return "", "", errors.New("invalid or revoked refresh token") }
	// rotate
	if err := s.tokens.RevokeExact(ctx, uid, auth.HashString(refreshToken)); err != nil { return "", "", err }
	access, _, err := s.jwt.GenerateAccessToken(uid)
	if err != nil { return "", "", err }
	newRefresh, rexp, err := s.jwt.GenerateRefreshToken(uid)
	if err != nil { return "", "", err }
	if err := s.tokens.Save(ctx, uid, auth.HashString(newRefresh), rexp); err != nil { return "", "", err }
	return access, newRefresh, nil
}

func (s *userService) Revoke(ctx context.Context, userID string) error {
	return s.tokens.RevokeAllForUser(ctx, userID)
}

func (s *userService) UpdateUser(ctx context.Context, userID string, name, gender, email, location *string) (*model.User, error) {
	u, err := s.users.GetByID(ctx, userID)
	if err != nil { return nil, err }
	if u == nil { return nil, errors.New("not found") }
	if name != nil { u.Name = strings.TrimSpace(*name) }
	if gender != nil { u.Gender = model.Gender(strings.ToLower(strings.TrimSpace(*gender))) }
	if email != nil { u.Email = email }
	if location != nil { u.Location = location }
	if err := s.users.Update(ctx, u); err != nil { return nil, err }
	return u, nil
}

func (s *userService) DeleteUser(ctx context.Context, requesterID, targetID string) error {
	if requesterID != targetID { return errors.New("forbidden") }
	return s.users.Delete(ctx, targetID)
}

func init() {
	rand.Seed(time.Now().UnixNano())
	log.Debug().Msg("user service package init")
}
