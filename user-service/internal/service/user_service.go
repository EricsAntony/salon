package service

import (
	"context"
	"errors"
	"math/rand"
	"strings"
	"time"

	"user-service/internal/config"
	appErrors "user-service/internal/errors"
	"user-service/internal/repository"

	sharedauth "github.com/EricsAntony/salon/salon-shared/auth"
	sharedvalidation "github.com/EricsAntony/salon/salon-shared/validation"
	models "user-service/internal/model"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

// RegisterParams bundles inputs to Register
type RegisterParams struct {
	Phone    string
	Name     string
	Gender   string
	Email    *string
	Location *string
	Lat      *float64
	Lng      *float64
	OTP      string
}

type UpdateUserParams struct {
	UserID   string
	Location *string
	Lat      *float64
	Lng      *float64
}

type UserService interface {
	RequestOTP(ctx context.Context, phone string) (string, error)
	Register(ctx context.Context, p RegisterParams) (*models.User, string, string, error)
	Authenticate(ctx context.Context, phone, otp string) (string, string, string, error)
	GetUser(ctx context.Context, id string) (*models.User, error)
	Refresh(ctx context.Context, refreshToken string) (string, string, error)
	Revoke(ctx context.Context, userID string) error
	UpdateUser(ctx context.Context, p UpdateUserParams) (*models.User, error)
	DeleteUser(ctx context.Context, requesterID, targetID string) error
	HealthCheck(ctx context.Context) error
}

type userService struct {
	users  repository.UserRepository
	otps   repository.OTPRepository
	tokens repository.TokenRepository
	jwt    *sharedauth.JWTManager
	cfg    *config.Config
}

func NewUserService(u repository.UserRepository, o repository.OTPRepository, t repository.TokenRepository, jwt *sharedauth.JWTManager, cfg *config.Config) UserService {
	return &userService{users: u, otps: o, tokens: t, jwt: jwt, cfg: cfg}
}


func (s *userService) RequestOTP(ctx context.Context, phone string) (string, error) {
	// Use shared phone validation and normalization
	normalizedPhone, err := sharedvalidation.ValidatePhone(phone)
	if err != nil {
		return "", err
	}
	
	// Generate OTP using shared utility
	code, err := sharedvalidation.GenerateOTP()
	if err != nil {
		return "", err
	}
	
	exp := time.Now().Add(time.Duration(s.cfg.OTP.ExpiryMinutes) * time.Minute)
	hash := sharedauth.HashString(code)
	if err := s.otps.Create(ctx, normalizedPhone, hash, exp); err != nil {
		return "", err
	}
	sharedauth.SendOTPViaSMS(normalizedPhone, code)
	return code, nil
}

func (s *userService) verifyOTP(ctx context.Context, phone, otp string) error {
	// Validate OTP format using shared validation
	if err := sharedvalidation.ValidateOTP(otp); err != nil {
		return appErrors.ErrInvalidOTP
	}
	
	// Normalize phone using shared validation
	normalizedPhone, err := sharedvalidation.ValidatePhone(phone)
	if err != nil {
		return err
	}
	
	// Check rate limiting for failed attempts
	failedCount, err := s.otps.GetFailedAttemptsCount(ctx, normalizedPhone, s.cfg.OTP.FailureWindowMinutes)
	log.Info().Int("failure_window_minutes", s.cfg.OTP.FailureWindowMinutes).Msg("OTP failure count")
	if err != nil {
		log.Error().Err(err).Str("phone", normalizedPhone).Msg("failed to check OTP failure count")
		return appErrors.ErrInternalError
	}
	if failedCount >= s.cfg.OTP.MaxFailedAttempts {
		log.Warn().Str("phone", normalizedPhone).Int("failed_count", failedCount).Msg("OTP rate limit exceeded")
		return appErrors.ErrRateLimited
	}

	rec, err := s.otps.GetLatest(ctx, normalizedPhone)
	if err != nil {
		return appErrors.ErrInternalError
	}
	if rec == nil {
		return appErrors.ErrOTPNotRequested
	}
	if sharedauth.HashString(otp) != rec.CodeHash {
		_ = s.otps.IncrementAttempts(ctx, rec.ID)
		return appErrors.ErrInvalidOTP
	}
	if time.Now().After(rec.ExpiresAt) {
		return appErrors.ErrOTPExpired
	}
	return nil
}

func (s *userService) Register(ctx context.Context, p RegisterParams) (*models.User, string, string, error) {
	// Use shared phone validation and normalization
	phone, err := sharedvalidation.ValidatePhone(p.Phone)
	if err != nil {
		return nil, "", "", err
	}
	if err := s.verifyOTP(ctx, phone, p.OTP); err != nil {
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
	u := &models.User{ID: uid, PhoneNumber: phone, Name: p.Name, Gender: models.Gender(strings.ToLower(p.Gender)), Email: p.Email, Location: p.Location}
	// Set optional geo fields for response as well
	u.Lat = p.Lat
	u.Lng = p.Lng
	if u.Gender != models.GenderMale && u.Gender != models.GenderFemale && u.Gender != models.GenderOther {
		return nil, "", "", errors.New("invalid gender")
	}
	if err := s.users.Create(ctx, u); err != nil {
		return nil, "", "", err
	}
	// Tokens
	access, _, err := s.jwt.GenerateAccessTokenWithType(uid, sharedauth.UserTypeCustomer)
	if err != nil {
		return nil, "", "", err
	}
	refresh, rexp, err := s.jwt.GenerateRefreshTokenWithType(uid, sharedauth.UserTypeCustomer)
	if err != nil {
		return nil, "", "", err
	}
	// Revoke any existing refresh tokens for this user just in case
	if err := s.tokens.RevokeAllForUser(ctx, uid); err != nil {
		return nil, "", "", err
	}
	if err := s.tokens.Save(ctx, uid, sharedauth.HashString(refresh), rexp); err != nil {
		return nil, "", "", err
	}
	return u, access, refresh, nil
}

func (s *userService) Authenticate(ctx context.Context, phone, otp string) (string, string, string, error) {
	// Use shared phone validation and normalization
	normalizedPhone, err := sharedvalidation.ValidatePhone(phone)
	if err != nil {
		return "", "", "", err
	}
	if err := s.verifyOTP(ctx, normalizedPhone, otp); err != nil {
		return "", "", "", err
	}
	u, err := s.users.GetByPhone(ctx, normalizedPhone)
	if err != nil {
		return "", "", "", err
	}
	if u == nil {
		return "", "", "", appErrors.ErrUserNotRegistered
	}
	access, _, err := s.jwt.GenerateAccessTokenWithType(u.ID, sharedauth.UserTypeCustomer)
	if err != nil {
		return "", "", "", err
	}
	refresh, rexp, err := s.jwt.GenerateRefreshTokenWithType(u.ID, sharedauth.UserTypeCustomer)
	if err != nil {
		return "", "", "", err
	}
	// Revoke existing refresh tokens for this user upon new authentication
	if err := s.tokens.RevokeAllForUser(ctx, u.ID); err != nil {
		return "", "", "", err
	}
	if err := s.tokens.Save(ctx, u.ID, sharedauth.HashString(refresh), rexp); err != nil {
		return "", "", "", err
	}
	return u.ID, access, refresh, nil
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
	valid, err := s.tokens.IsValid(ctx, uid, sharedauth.HashString(refreshToken), time.Now())
	if err != nil {
		return "", "", err
	}
	if !valid {
		return "", "", errors.New("invalid or revoked refresh token")
	}
	// rotate
	if err := s.tokens.RevokeExact(ctx, uid, sharedauth.HashString(refreshToken)); err != nil {
		return "", "", err
	}
	access, _, err := s.jwt.GenerateAccessTokenWithType(uid, sharedauth.UserTypeCustomer)
	if err != nil {
		return "", "", err
	}
	newRefresh, rexp, err := s.jwt.GenerateRefreshTokenWithType(uid, sharedauth.UserTypeCustomer)
	if err != nil {
		return "", "", err
	}
	if err := s.tokens.Save(ctx, uid, sharedauth.HashString(newRefresh), rexp); err != nil {
		return "", "", err
	}
	return access, newRefresh, nil
}

func (s *userService) Revoke(ctx context.Context, userID string) error {
	return s.tokens.RevokeAllForUser(ctx, userID)
}

func (s *userService) UpdateUser(ctx context.Context, p UpdateUserParams) (*models.User, error) {
	u, err := s.users.GetByID(ctx, p.UserID)
	if err != nil {
		return nil, err
	}
	if u == nil {
		return nil, errors.New("not found")
	}
	if p.Location != nil {
		u.Location = p.Location
	}
	// reflect geo updates on returned model
	if p.Lat != nil {
		u.Lat = p.Lat
	}
	if p.Lng != nil {
		u.Lng = p.Lng
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
