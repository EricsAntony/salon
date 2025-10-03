package repository

import (
    "context"
    "errors"
    "time"

    "github.com/jackc/pgx/v5"
    "github.com/jackc/pgx/v5/pgconn"
    "github.com/jackc/pgx/v5/pgxpool"
    "salon-service/internal/model"
)

type StaffAuthRepository interface {
    CreateOTP(ctx context.Context, phone, codeHash string, expiresAt time.Time) error
    GetLatestOTP(ctx context.Context, phone string) (*model.StaffOTP, error)
    IncrementOTPAttempts(ctx context.Context, id int64) error
    CreateRefreshToken(ctx context.Context, staffID, tokenHash string, expiresAt time.Time) error
    RevokeRefreshTokens(ctx context.Context, staffID string) error
    IsRefreshTokenValid(ctx context.Context, staffID, tokenHash string, now time.Time) (bool, error)
    RevokeRefreshToken(ctx context.Context, staffID, tokenHash string) error
}

type staffAuthRepository struct {
    db *pgxpool.Pool
}

func NewStaffAuthRepository(db *pgxpool.Pool) StaffAuthRepository {
    return &staffAuthRepository{db: db}
}

func (r *staffAuthRepository) CreateOTP(ctx context.Context, phone, codeHash string, expiresAt time.Time) error {
    _, err := r.db.Exec(ctx, `
        INSERT INTO staff_otps (phone_number, code_hash, expires_at, created_at)
        VALUES ($1, $2, $3, NOW())
    `, phone, codeHash, expiresAt)
    return err
}

func (r *staffAuthRepository) GetLatestOTP(ctx context.Context, phone string) (*model.StaffOTP, error) {
    row := r.db.QueryRow(ctx, `
        SELECT id, phone_number, code_hash, expires_at, attempts, created_at
        FROM staff_otps
        WHERE phone_number = $1
        ORDER BY id DESC
        LIMIT 1
    `, phone)
    var otp model.StaffOTP
    if err := row.Scan(&otp.ID, &otp.PhoneNumber, &otp.CodeHash, &otp.ExpiresAt, &otp.Attempts, &otp.CreatedAt); err != nil {
        if errors.Is(err, pgx.ErrNoRows) {
            return nil, nil
        }
        return nil, err
    }
    return &otp, nil
}

func (r *staffAuthRepository) IncrementOTPAttempts(ctx context.Context, id int64) error {
    _, err := r.db.Exec(ctx, `UPDATE staff_otps SET attempts = attempts + 1 WHERE id = $1`, id)
    return err
}

func (r *staffAuthRepository) CreateRefreshToken(ctx context.Context, staffID, tokenHash string, expiresAt time.Time) error {
    _, err := r.db.Exec(ctx, `
        INSERT INTO staff_refresh_tokens (staff_id, token_hash, expires_at, created_at, revoked)
        VALUES ($1, $2, $3, NOW(), false)
    `, staffID, tokenHash, expiresAt)
    return err
}

func (r *staffAuthRepository) RevokeRefreshTokens(ctx context.Context, staffID string) error {
    _, err := r.db.Exec(ctx, `UPDATE staff_refresh_tokens SET revoked = true WHERE staff_id = $1 AND revoked = false`, staffID)
    return err
}

func (r *staffAuthRepository) IsRefreshTokenValid(ctx context.Context, staffID, tokenHash string, now time.Time) (bool, error) {
    row := r.db.QueryRow(ctx, `
        SELECT 1
        FROM staff_refresh_tokens
        WHERE staff_id = $1 AND token_hash = $2 AND revoked = false AND expires_at > $3
        LIMIT 1
    `, staffID, tokenHash, now)
    var one int
    if err := row.Scan(&one); err != nil {
        if errors.Is(err, pgx.ErrNoRows) {
            return false, nil
        }
        return false, err
    }
    return true, nil
}

func (r *staffAuthRepository) RevokeRefreshToken(ctx context.Context, staffID, tokenHash string) error {
    _, err := r.db.Exec(ctx, `
        UPDATE staff_refresh_tokens SET revoked = true
        WHERE staff_id = $1 AND token_hash = $2 AND revoked = false
    `, staffID, tokenHash)
    return err
}

func IsUniqueViolation(err error) bool {
    var pgErr *pgconn.PgError
    if errors.As(err, &pgErr) {
        return pgErr.Code == "23505"
    }
    return false
}
