package auth

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/rs/zerolog/log"
	"user-service/internal/config"
)

type JWTManager struct {
	accessSecret  []byte
	refreshSecret []byte
	accessTTL     time.Duration
	refreshTTL    time.Duration
}

type Claims struct {
	UserID string `json:"uid"`
	jwt.RegisteredClaims
}

func NewJWTManager(cfg *config.Config) *JWTManager {
	return &JWTManager{
		accessSecret:  []byte(cfg.JWT.AccessSecret),
		refreshSecret: []byte(cfg.JWT.RefreshSecret),
		accessTTL:     time.Duration(cfg.JWT.AccessTTLMinutes) * time.Minute,
		refreshTTL:    time.Duration(cfg.JWT.RefreshTTLDays) * 24 * time.Hour,
	}
}

func (m *JWTManager) GenerateAccessToken(userID string) (string, time.Time, error) {
	exp := time.Now().Add(m.accessTTL)
	claims := &Claims{UserID: userID, RegisteredClaims: jwt.RegisteredClaims{ExpiresAt: jwt.NewNumericDate(exp)}}
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	s, err := t.SignedString(m.accessSecret)
	return s, exp, err
}

func (m *JWTManager) GenerateRefreshToken(userID string) (string, time.Time, error) {
	exp := time.Now().Add(m.refreshTTL)
	claims := &Claims{UserID: userID, RegisteredClaims: jwt.RegisteredClaims{ExpiresAt: jwt.NewNumericDate(exp)}}
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	s, err := t.SignedString(m.refreshSecret)
	return s, exp, err
}

func (m *JWTManager) ValidateAccessToken(tok string) (*Claims, error) {
	return m.parse(tok, m.accessSecret)
}

func (m *JWTManager) ValidateRefreshToken(tok string) (*Claims, error) {
	return m.parse(tok, m.refreshSecret)
}

func (m *JWTManager) parse(tok string, secret []byte) (*Claims, error) {
	parsed, err := jwt.ParseWithClaims(tok, &Claims{}, func(token *jwt.Token) (interface{}, error) { return secret, nil })
	if err != nil {
		return nil, err
	}
	if !parsed.Valid {
		return nil, errors.New("invalid token")
	}
	claims, ok := parsed.Claims.(*Claims)
	if !ok {
		return nil, errors.New("invalid claims")
	}
	return claims, nil
}

func HashString(s string) string {
	h := sha256.Sum256([]byte(s))
	return hex.EncodeToString(h[:])
}

// For demo purposes, OTP code will be logged; in production integrate with an SMS provider.
func SendOTPViaSMS(phone, code string) {
	log.Info().Str("phone", phone).Str("otp_code", code).Msg("OTP generated (send via SMS in production)")
}
