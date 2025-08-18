package service

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"user-service/internal/config"
	appErrors "user-service/internal/errors"
	"user-service/internal/repository"

	"github.com/EricsAntony/salon/salon-shared/auth"
	"github.com/EricsAntony/salon/salon-shared/models"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

type UserService interface {
	RequestOTP(ctx context.Context, phone string) (string, error)
	Register(ctx context.Context, phone, name, gender string, email, location *string, otp string) (*models.User, string, string, error)
	Authenticate(ctx context.Context, phone, otp string) (string, string, error)
	GetUser(ctx context.Context, id string) (*models.User, error)
	Refresh(ctx context.Context, refreshToken string) (string, string, error)
	Revoke(ctx context.Context, userID string) error
	UpdateUser(ctx context.Context, userID string, name, gender, email, location *string) (*models.User, error)
	DeleteUser(ctx context.Context, requesterID, targetID string) error
	HealthCheck(ctx context.Context) error
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

func (s *userService) RequestOTP(ctx context.Context, phone string) (string, error) {
	phone = normalizePhone(phone)
	if phone == "" {
		return "", errors.New("phone required")
	}
	code := fmt.Sprintf("%06d", rand.Intn(1000000))
	exp := time.Now().Add(time.Duration(s.cfg.OTP.ExpiryMinutes) * time.Minute)
	hash := auth.HashString(code)
	if err := s.otps.Create(ctx, phone, hash, exp); err != nil {
		return "", err
	}
	auth.SendOTPViaSMS(phone, code)
	return code, nil
}

func (s *userService) verifyOTP(ctx context.Context, phone, otp string) error {
	// Check rate limiting for failed attempts
	failedCount, err := s.otps.GetFailedAttemptsCount(ctx, phone, s.cfg.OTP.FailureWindowMinutes)
	log.Info().Int("failure_window_minutes", s.cfg.OTP.FailureWindowMinutes).Msg("OTP failure count")
	if err != nil {
		log.Error().Err(err).Str("phone", phone).Msg("failed to check OTP failure count")
		return appErrors.ErrInternalError
	}
	if failedCount >= s.cfg.OTP.MaxFailedAttempts {
		log.Warn().Str("phone", phone).Int("failed_count", failedCount).Msg("OTP rate limit exceeded")
		return appErrors.ErrRateLimited
	}

	rec, err := s.otps.GetLatest(ctx, phone)
	if err != nil {
		return appErrors.ErrInternalError
	}
	if rec == nil {
		return appErrors.ErrOTPNotRequested
	}
	if auth.HashString(otp) != rec.CodeHash {
		_ = s.otps.IncrementAttempts(ctx, rec.ID)
		return appErrors.ErrInvalidOTP
	}
	if time.Now().After(rec.ExpiresAt) {
		return appErrors.ErrOTPExpired
	}
	return nil
}

func (s *userService) Register(ctx context.Context, phone, name, gender string, email, location *string, otp string) (*models.User, string, string, error) {
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
	u := &models.User{ID: uid, PhoneNumber: phone, Name: name, Gender: models.Gender(strings.ToLower(gender)), Email: email, Location: location}
	if u.Gender != models.GenderMale && u.Gender != models.GenderFemale && u.Gender != models.GenderOther {
		return nil, "", "", errors.New("invalid gender")
	}
	if err := s.users.Create(ctx, u); err != nil {
		return nil, "", "", err
	}
	// Tokens
	access, _, err := s.jwt.GenerateAccessToken(uid)
	if err != nil {
		return nil, "", "", err
	}
	refresh, rexp, err := s.jwt.GenerateRefreshToken(uid)
	if err != nil {
		return nil, "", "", err
	}
	// Revoke any existing refresh tokens for this user just in case
	if err := s.tokens.RevokeAllForUser(ctx, uid); err != nil {
		return nil, "", "", err
	}
	if err := s.tokens.Save(ctx, uid, auth.HashString(refresh), rexp); err != nil {
		return nil, "", "", err
	}
	return u, access, refresh, nil
}

func (s *userService) Authenticate(ctx context.Context, phone, otp string) (string, string, error) {
	phone = normalizePhone(phone)
	if err := s.verifyOTP(ctx, phone, otp); err != nil {
		return "", "", err
	}
	u, err := s.users.GetByPhone(ctx, phone)
	if err != nil {
		return "", "", err
	}
	if u == nil {
		return "", "", appErrors.ErrUserNotRegistered
	}
	access, _, err := s.jwt.GenerateAccessToken(u.ID)
	if err != nil {
		return "", "", err
	}
	refresh, rexp, err := s.jwt.GenerateRefreshToken(u.ID)
	if err != nil {
		return "", "", err
	}
	// Revoke existing refresh tokens for this user upon new authentication
	if err := s.tokens.RevokeAllForUser(ctx, u.ID); err != nil {
		return "", "", err
	}
	if err := s.tokens.Save(ctx, u.ID, auth.HashString(refresh), rexp); err != nil {
		return "", "", err
	}
	return access, refresh, nil
}

func (s *userService) GetUser(ctx context.Context, id string) (*models.User, error) {
	u, err := s.users.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if u == nil {
		return nil, errors.New("not found")
	}
	return u, nil
}

func (s *userService) Refresh(ctx context.Context, refreshToken string) (string, string, error) {
	claims, err := s.jwt.ValidateRefreshToken(refreshToken)
	if err != nil {
		return "", "", err
	}
	uid := claims.UserID
	// validate against DB
	valid, err := s.tokens.IsValid(ctx, uid, auth.HashString(refreshToken), time.Now())
	if err != nil {
		return "", "", err
	}
	if !valid {
		return "", "", errors.New("invalid or revoked refresh token")
	}
	// rotate
	if err := s.tokens.RevokeExact(ctx, uid, auth.HashString(refreshToken)); err != nil {
		return "", "", err
	}
	access, _, err := s.jwt.GenerateAccessToken(uid)
	if err != nil {
		return "", "", err
	}
	newRefresh, rexp, err := s.jwt.GenerateRefreshToken(uid)
	if err != nil {
		return "", "", err
	}
	if err := s.tokens.Save(ctx, uid, auth.HashString(newRefresh), rexp); err != nil {
		return "", "", err
	}
	return access, newRefresh, nil
}

func (s *userService) Revoke(ctx context.Context, userID string) error {
	return s.tokens.RevokeAllForUser(ctx, userID)
}

func (s *userService) UpdateUser(ctx context.Context, userID string, name, gender, email, location *string) (*models.User, error) {
	u, err := s.users.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if u == nil {
		return nil, errors.New("not found")
	}
	if name != nil {
		u.Name = strings.TrimSpace(*name)
	}
	if gender != nil {
		u.Gender = models.Gender(strings.ToLower(strings.TrimSpace(*gender)))
	}
	if email != nil {
		u.Email = email
	}
	if location != nil {
		u.Location = location
	}
	if err := s.users.Update(ctx, u); err != nil {
		return nil, err
	}
	return u, nil
}

func (s *userService) DeleteUser(ctx context.Context, requesterID, targetID string) error {
	if requesterID != targetID {
		return errors.New("forbidden")
	}
	return s.users.Delete(ctx, targetID)
}

func (s *userService) HealthCheck(ctx context.Context) error {
	// Check database connectivity by performing a simple query
	return s.users.HealthCheck(ctx)
}

func init() {
	rand.Seed(time.Now().UnixNano())
	log.Debug().Msg("user service package init")
}
