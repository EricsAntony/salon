package repository

import (
	"context"
	"errors"
	"time"

	models "user-service/internal/model"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog/log"
)

type UserRepository interface {
	Create(ctx context.Context, u *models.User) error
	GetByID(ctx context.Context, id string) (*models.User, error)
	GetByPhone(ctx context.Context, phone string) (*models.User, error)
	Update(ctx context.Context, u *models.User) error
	Delete(ctx context.Context, id string) error
	HealthCheck(ctx context.Context) error
}

func (r *tokenRepository) RevokeExact(ctx context.Context, userID, tokenHash string) error {
	_, err := r.db.Exec(ctx, `UPDATE refresh_tokens SET revoked = true WHERE user_id = $1 AND token_hash = $2 AND revoked = false`, userID, tokenHash)
	return err
}

func (r *userRepository) Update(ctx context.Context, u *models.User) error {
	ct, err := r.db.Exec(ctx, `UPDATE users SET phone_number = $2, name = $3, gender = $4, email = $5, location = $6, lat = COALESCE($7, lat), lng = COALESCE($8, lng), updated_at = NOW() WHERE id = $1`,
		u.ID, u.PhoneNumber, u.Name, u.Gender, u.Email, u.Location, u.Lat, u.Lng)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

func (r *userRepository) Delete(ctx context.Context, id string) error {
	_, err := r.db.Exec(ctx, `DELETE FROM users WHERE id = $1`, id)
	return err
}

type OTPRepository interface {
	Create(ctx context.Context, phone, codeHash string, expiresAt time.Time) error
	GetLatest(ctx context.Context, phone string) (*models.OTP, error)
	IncrementAttempts(ctx context.Context, id int64) error
	GetFailedAttemptsCount(ctx context.Context, phone string, windowMinutes int) (int, error)
	DeleteExpired(ctx context.Context, maxAge time.Duration) (int, error)
}

type TokenRepository interface {
	Save(ctx context.Context, userID, tokenHash string, expiresAt time.Time) error
	RevokeAllForUser(ctx context.Context, userID string) error
	IsValid(ctx context.Context, userID, tokenHash string, now time.Time) (bool, error)
	RevokeExact(ctx context.Context, userID, tokenHash string) error
}

type userRepository struct{ db *pgxpool.Pool }

type otpRepository struct{ db *pgxpool.Pool }

type tokenRepository struct{ db *pgxpool.Pool }

func NewUserRepository(db *pgxpool.Pool) UserRepository   { return &userRepository{db} }
func NewOTPRepository(db *pgxpool.Pool) OTPRepository     { return &otpRepository{db} }
func NewTokenRepository(db *pgxpool.Pool) TokenRepository { return &tokenRepository{db} }

func (r *userRepository) Create(ctx context.Context, u *models.User) error {
	_, err := r.db.Exec(ctx, `INSERT INTO users (id, phone_number, name, gender, email, location, lat, lng, created_at, updated_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8, NOW(), NOW())`,
		u.ID, u.PhoneNumber, u.Name, u.Gender, u.Email, u.Location, u.Lat, u.Lng)
	return err
}

func (r *userRepository) GetByID(ctx context.Context, id string) (*models.User, error) {
	row := r.db.QueryRow(ctx, `SELECT id, phone_number, name, gender, email, location, lat, lng, created_at, updated_at FROM users WHERE id = $1`, id)
	var u models.User
	var email, loc *string
	var lat, lng *float64
	if err := row.Scan(&u.ID, &u.PhoneNumber, &u.Name, &u.Gender, &email, &loc, &lat, &lng, &u.CreatedAt, &u.UpdatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	u.Email = email
	u.Location = loc
	u.Lat = lat
	u.Lng = lng
	return &u, nil
}

func (r *userRepository) GetByPhone(ctx context.Context, phone string) (*models.User, error) {
	row := r.db.QueryRow(ctx, `SELECT id, phone_number, name, gender, email, location, lat, lng, created_at, updated_at FROM users WHERE phone_number = $1`, phone)
	var u models.User
	var email, loc *string
	var lat, lng *float64
	if err := row.Scan(&u.ID, &u.PhoneNumber, &u.Name, &u.Gender, &email, &loc, &lat, &lng, &u.CreatedAt, &u.UpdatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	u.Email = email
	u.Location = loc
	u.Lat = lat
	u.Lng = lng
	return &u, nil
}

func (r *userRepository) HealthCheck(ctx context.Context) error {
	// Simple ping to check database connectivity
	return r.db.Ping(ctx)
}

func (r *otpRepository) Create(ctx context.Context, phone, codeHash string, expiresAt time.Time) error {
	_, err := r.db.Exec(ctx, `INSERT INTO otps (phone_number, code_hash, expires_at, created_at) VALUES ($1,$2,$3,NOW())`, phone, codeHash, expiresAt)
	return err
}

func (r *otpRepository) GetLatest(ctx context.Context, phone string) (*models.OTP, error) {
	row := r.db.QueryRow(ctx, `SELECT id, phone_number, code_hash, expires_at, attempts, created_at FROM otps WHERE phone_number = $1 ORDER BY id DESC LIMIT 1`, phone)
	var o models.OTP
	if err := row.Scan(&o.ID, &o.PhoneNumber, &o.CodeHash, &o.ExpiresAt, &o.Attempts, &o.CreatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &o, nil
}

func (r *otpRepository) IncrementAttempts(ctx context.Context, id int64) error {
	ct, err := r.db.Exec(ctx, `UPDATE otps SET attempts = attempts + 1 WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		log.Warn().Int64("otp_id", id).Msg("no otp row to increment")
	}
	return nil
}

func (r *otpRepository) GetFailedAttemptsCount(ctx context.Context, phone string, windowMinutes int) (int, error) {
	row := r.db.QueryRow(ctx, `
		SELECT COALESCE(SUM(attempts), 0) 
		FROM otps 
		WHERE phone_number = $1 
		AND created_at > NOW() - (INTERVAL '1 minute') * $2
		AND attempts > 0`, phone, windowMinutes)

	var count int
	if err := row.Scan(&count); err != nil {
		return 0, err
	}
	return count, nil
}

func (r *otpRepository) DeleteExpired(ctx context.Context, maxAge time.Duration) (int, error) {
	cutoff := time.Now().Add(-maxAge)
	result, err := r.db.Exec(ctx, `DELETE FROM otps WHERE created_at < $1`, cutoff)
	if err != nil {
		return 0, err
	}
	return int(result.RowsAffected()), nil
}

func (r *tokenRepository) Save(ctx context.Context, userID, tokenHash string, expiresAt time.Time) error {
	_, err := r.db.Exec(ctx, `INSERT INTO refresh_tokens (user_id, token_hash, expires_at, created_at, revoked) VALUES ($1,$2,$3,NOW(), false)`, userID, tokenHash, expiresAt)
	return err
}

func (r *tokenRepository) RevokeAllForUser(ctx context.Context, userID string) error {
	_, err := r.db.Exec(ctx, `UPDATE refresh_tokens SET revoked = true WHERE user_id = $1 AND revoked = false`, userID)
	return err
}

func (r *tokenRepository) IsValid(ctx context.Context, userID, tokenHash string, now time.Time) (bool, error) {
	row := r.db.QueryRow(ctx, `SELECT 1 FROM refresh_tokens WHERE user_id = $1 AND token_hash = $2 AND revoked = false AND expires_at > $3 LIMIT 1`, userID, tokenHash, now)
	var one int
	if err := row.Scan(&one); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}
