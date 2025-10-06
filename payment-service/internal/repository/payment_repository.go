package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"payment-service/internal/model"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

// PaymentRepository defines the interface for payment data operations
type PaymentRepository interface {
	// Payment operations
	Create(ctx context.Context, payment *model.Payment) error
	GetByID(ctx context.Context, id uuid.UUID) (*model.Payment, error)
	GetByBookingID(ctx context.Context, bookingID uuid.UUID) ([]*model.Payment, error)
	GetByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*model.Payment, error)
	Update(ctx context.Context, payment *model.Payment) error
	UpdateStatus(ctx context.Context, id uuid.UUID, status string) error

	// Refund operations
	CreateRefund(ctx context.Context, refund *model.Refund) error
	GetRefundByID(ctx context.Context, id uuid.UUID) (*model.Refund, error)
	GetRefundsByPaymentID(ctx context.Context, paymentID uuid.UUID) ([]*model.Refund, error)
	UpdateRefund(ctx context.Context, refund *model.Refund) error

	// Payment attempt operations
	CreateAttempt(ctx context.Context, attempt *model.PaymentAttempt) error
	GetAttemptsByPaymentID(ctx context.Context, paymentID uuid.UUID) ([]*model.PaymentAttempt, error)

	// Idempotency operations
	CreateIdempotencyRecord(ctx context.Context, record *model.IdempotencyRecord) error
	GetIdempotencyRecord(ctx context.Context, key string) (*model.IdempotencyRecord, error)
	CleanupExpiredIdempotencyRecords(ctx context.Context) error

	// Analytics and reporting
	GetPaymentStats(ctx context.Context, from, to time.Time) (map[string]interface{}, error)
}

type paymentRepository struct {
	db *sql.DB
}

// NewPaymentRepository creates a new payment repository
func NewPaymentRepository(db *sql.DB) PaymentRepository {
	return &paymentRepository{db: db}
}

// Create creates a new payment
func (r *paymentRepository) Create(ctx context.Context, payment *model.Payment) error {
	query := `
		INSERT INTO payments (
			id, booking_id, user_id, amount, currency, status, gateway,
			gateway_payment_id, gateway_order_id, payment_method, payment_url,
			idempotency_key, metadata, failure_reason, processed_at, expires_at,
			created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18
		)`

	_, err := r.db.ExecContext(ctx, query,
		payment.ID, payment.BookingID, payment.UserID, payment.Amount, payment.Currency,
		payment.Status, payment.Gateway, payment.GatewayPaymentID, payment.GatewayOrderID,
		payment.PaymentMethod, payment.PaymentURL, payment.IdempotencyKey, payment.Metadata,
		payment.FailureReason, payment.ProcessedAt, payment.ExpiresAt,
		payment.CreatedAt, payment.UpdatedAt,
	)

	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
			return fmt.Errorf("payment with idempotency key already exists")
		}
		return fmt.Errorf("failed to create payment: %w", err)
	}

	return nil
}

// GetByID retrieves a payment by ID
func (r *paymentRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.Payment, error) {
	query := `
		SELECT id, booking_id, user_id, amount, currency, status, gateway,
			   gateway_payment_id, gateway_order_id, payment_method, payment_url,
			   idempotency_key, metadata, failure_reason, processed_at, expires_at,
			   created_at, updated_at
		FROM payments WHERE id = $1`

	payment := &model.Payment{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&payment.ID, &payment.BookingID, &payment.UserID, &payment.Amount, &payment.Currency,
		&payment.Status, &payment.Gateway, &payment.GatewayPaymentID, &payment.GatewayOrderID,
		&payment.PaymentMethod, &payment.PaymentURL, &payment.IdempotencyKey, &payment.Metadata,
		&payment.FailureReason, &payment.ProcessedAt, &payment.ExpiresAt,
		&payment.CreatedAt, &payment.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("payment not found")
		}
		return nil, fmt.Errorf("failed to get payment: %w", err)
	}

	return payment, nil
}

// GetByBookingID retrieves payments by booking ID
func (r *paymentRepository) GetByBookingID(ctx context.Context, bookingID uuid.UUID) ([]*model.Payment, error) {
	query := `
		SELECT id, booking_id, user_id, amount, currency, status, gateway,
			   gateway_payment_id, gateway_order_id, payment_method, payment_url,
			   idempotency_key, metadata, failure_reason, processed_at, expires_at,
			   created_at, updated_at
		FROM payments WHERE booking_id = $1 ORDER BY created_at DESC`

	rows, err := r.db.QueryContext(ctx, query, bookingID)
	if err != nil {
		return nil, fmt.Errorf("failed to get payments by booking ID: %w", err)
	}
	defer rows.Close()

	var payments []*model.Payment
	for rows.Next() {
		payment := &model.Payment{}
		err := rows.Scan(
			&payment.ID, &payment.BookingID, &payment.UserID, &payment.Amount, &payment.Currency,
			&payment.Status, &payment.Gateway, &payment.GatewayPaymentID, &payment.GatewayOrderID,
			&payment.PaymentMethod, &payment.PaymentURL, &payment.IdempotencyKey, &payment.Metadata,
			&payment.FailureReason, &payment.ProcessedAt, &payment.ExpiresAt,
			&payment.CreatedAt, &payment.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan payment: %w", err)
		}
		payments = append(payments, payment)
	}

	return payments, nil
}

// GetByUserID retrieves payments by user ID with pagination
func (r *paymentRepository) GetByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*model.Payment, error) {
	query := `
		SELECT id, booking_id, user_id, amount, currency, status, gateway,
			   gateway_payment_id, gateway_order_id, payment_method, payment_url,
			   idempotency_key, metadata, failure_reason, processed_at, expires_at,
			   created_at, updated_at
		FROM payments WHERE user_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`

	rows, err := r.db.QueryContext(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get payments by user ID: %w", err)
	}
	defer rows.Close()

	var payments []*model.Payment
	for rows.Next() {
		payment := &model.Payment{}
		err := rows.Scan(
			&payment.ID, &payment.BookingID, &payment.UserID, &payment.Amount, &payment.Currency,
			&payment.Status, &payment.Gateway, &payment.GatewayPaymentID, &payment.GatewayOrderID,
			&payment.PaymentMethod, &payment.PaymentURL, &payment.IdempotencyKey, &payment.Metadata,
			&payment.FailureReason, &payment.ProcessedAt, &payment.ExpiresAt,
			&payment.CreatedAt, &payment.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan payment: %w", err)
		}
		payments = append(payments, payment)
	}

	return payments, nil
}

// Update updates a payment
func (r *paymentRepository) Update(ctx context.Context, payment *model.Payment) error {
	query := `
		UPDATE payments SET
			status = $2, gateway_payment_id = $3, gateway_order_id = $4,
			payment_method = $5, payment_url = $6, metadata = $7,
			failure_reason = $8, processed_at = $9, expires_at = $10, updated_at = $11
		WHERE id = $1`

	payment.UpdatedAt = time.Now()

	_, err := r.db.ExecContext(ctx, query,
		payment.ID, payment.Status, payment.GatewayPaymentID, payment.GatewayOrderID,
		payment.PaymentMethod, payment.PaymentURL, payment.Metadata,
		payment.FailureReason, payment.ProcessedAt, payment.ExpiresAt, payment.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to update payment: %w", err)
	}

	return nil
}

// UpdateStatus updates only the payment status
func (r *paymentRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status string) error {
	query := `UPDATE payments SET status = $2, updated_at = $3 WHERE id = $1`

	_, err := r.db.ExecContext(ctx, query, id, status, time.Now())
	if err != nil {
		return fmt.Errorf("failed to update payment status: %w", err)
	}

	return nil
}

// CreateRefund creates a new refund
func (r *paymentRepository) CreateRefund(ctx context.Context, refund *model.Refund) error {
	query := `
		INSERT INTO refunds (
			id, payment_id, amount, currency, status, gateway, gateway_refund_id,
			reason, idempotency_key, metadata, failure_reason, processed_at,
			created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14
		)`

	_, err := r.db.ExecContext(ctx, query,
		refund.ID, refund.PaymentID, refund.Amount, refund.Currency, refund.Status,
		refund.Gateway, refund.GatewayRefundID, refund.Reason, refund.IdempotencyKey,
		refund.Metadata, refund.FailureReason, refund.ProcessedAt,
		refund.CreatedAt, refund.UpdatedAt,
	)

	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
			return fmt.Errorf("refund with idempotency key already exists")
		}
		return fmt.Errorf("failed to create refund: %w", err)
	}

	return nil
}

// GetRefundByID retrieves a refund by ID
func (r *paymentRepository) GetRefundByID(ctx context.Context, id uuid.UUID) (*model.Refund, error) {
	query := `
		SELECT id, payment_id, amount, currency, status, gateway, gateway_refund_id,
			   reason, idempotency_key, metadata, failure_reason, processed_at,
			   created_at, updated_at
		FROM refunds WHERE id = $1`

	refund := &model.Refund{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&refund.ID, &refund.PaymentID, &refund.Amount, &refund.Currency, &refund.Status,
		&refund.Gateway, &refund.GatewayRefundID, &refund.Reason, &refund.IdempotencyKey,
		&refund.Metadata, &refund.FailureReason, &refund.ProcessedAt,
		&refund.CreatedAt, &refund.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("refund not found")
		}
		return nil, fmt.Errorf("failed to get refund: %w", err)
	}

	return refund, nil
}

// GetRefundsByPaymentID retrieves refunds by payment ID
func (r *paymentRepository) GetRefundsByPaymentID(ctx context.Context, paymentID uuid.UUID) ([]*model.Refund, error) {
	query := `
		SELECT id, payment_id, amount, currency, status, gateway, gateway_refund_id,
			   reason, idempotency_key, metadata, failure_reason, processed_at,
			   created_at, updated_at
		FROM refunds WHERE payment_id = $1 ORDER BY created_at DESC`

	rows, err := r.db.QueryContext(ctx, query, paymentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get refunds by payment ID: %w", err)
	}
	defer rows.Close()

	var refunds []*model.Refund
	for rows.Next() {
		refund := &model.Refund{}
		err := rows.Scan(
			&refund.ID, &refund.PaymentID, &refund.Amount, &refund.Currency, &refund.Status,
			&refund.Gateway, &refund.GatewayRefundID, &refund.Reason, &refund.IdempotencyKey,
			&refund.Metadata, &refund.FailureReason, &refund.ProcessedAt,
			&refund.CreatedAt, &refund.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan refund: %w", err)
		}
		refunds = append(refunds, refund)
	}

	return refunds, nil
}

// UpdateRefund updates a refund
func (r *paymentRepository) UpdateRefund(ctx context.Context, refund *model.Refund) error {
	query := `
		UPDATE refunds SET
			status = $2, gateway_refund_id = $3, metadata = $4,
			failure_reason = $5, processed_at = $6, updated_at = $7
		WHERE id = $1`

	refund.UpdatedAt = time.Now()

	_, err := r.db.ExecContext(ctx, query,
		refund.ID, refund.Status, refund.GatewayRefundID, refund.Metadata,
		refund.FailureReason, refund.ProcessedAt, refund.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to update refund: %w", err)
	}

	return nil
}

// CreateAttempt creates a new payment attempt
func (r *paymentRepository) CreateAttempt(ctx context.Context, attempt *model.PaymentAttempt) error {
	query := `
		INSERT INTO payment_attempts (
			id, payment_id, attempt_number, gateway, status,
			error_message, response_data, attempted_at, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`

	_, err := r.db.ExecContext(ctx, query,
		attempt.ID, attempt.PaymentID, attempt.AttemptNumber, attempt.Gateway,
		attempt.Status, attempt.ErrorMessage, attempt.ResponseData,
		attempt.AttemptedAt, attempt.CreatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create payment attempt: %w", err)
	}

	return nil
}

// GetAttemptsByPaymentID retrieves payment attempts by payment ID
func (r *paymentRepository) GetAttemptsByPaymentID(ctx context.Context, paymentID uuid.UUID) ([]*model.PaymentAttempt, error) {
	query := `
		SELECT id, payment_id, attempt_number, gateway, status,
			   error_message, response_data, attempted_at, created_at
		FROM payment_attempts WHERE payment_id = $1 ORDER BY attempt_number ASC`

	rows, err := r.db.QueryContext(ctx, query, paymentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get payment attempts: %w", err)
	}
	defer rows.Close()

	var attempts []*model.PaymentAttempt
	for rows.Next() {
		attempt := &model.PaymentAttempt{}
		err := rows.Scan(
			&attempt.ID, &attempt.PaymentID, &attempt.AttemptNumber, &attempt.Gateway,
			&attempt.Status, &attempt.ErrorMessage, &attempt.ResponseData,
			&attempt.AttemptedAt, &attempt.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan payment attempt: %w", err)
		}
		attempts = append(attempts, attempt)
	}

	return attempts, nil
}

// CreateIdempotencyRecord creates a new idempotency record
func (r *paymentRepository) CreateIdempotencyRecord(ctx context.Context, record *model.IdempotencyRecord) error {
	query := `
		INSERT INTO idempotency_records (
			id, idempotency_key, payment_id, request_hash, response_data, expires_at, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7)`

	_, err := r.db.ExecContext(ctx, query,
		record.ID, record.IdempotencyKey, record.PaymentID, record.RequestHash,
		record.ResponseData, record.ExpiresAt, record.CreatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create idempotency record: %w", err)
	}

	return nil
}

// GetIdempotencyRecord retrieves an idempotency record by key
func (r *paymentRepository) GetIdempotencyRecord(ctx context.Context, key string) (*model.IdempotencyRecord, error) {
	query := `
		SELECT id, idempotency_key, payment_id, request_hash, response_data, expires_at, created_at
		FROM idempotency_records WHERE idempotency_key = $1 AND expires_at > NOW()`

	record := &model.IdempotencyRecord{}
	err := r.db.QueryRowContext(ctx, query, key).Scan(
		&record.ID, &record.IdempotencyKey, &record.PaymentID, &record.RequestHash,
		&record.ResponseData, &record.ExpiresAt, &record.CreatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Not found, but not an error
		}
		return nil, fmt.Errorf("failed to get idempotency record: %w", err)
	}

	return record, nil
}

// CleanupExpiredIdempotencyRecords removes expired idempotency records
func (r *paymentRepository) CleanupExpiredIdempotencyRecords(ctx context.Context) error {
	query := `DELETE FROM idempotency_records WHERE expires_at <= NOW()`

	_, err := r.db.ExecContext(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to cleanup expired idempotency records: %w", err)
	}

	return nil
}

// GetPaymentStats retrieves payment statistics for a date range
func (r *paymentRepository) GetPaymentStats(ctx context.Context, from, to time.Time) (map[string]interface{}, error) {
	query := `
		SELECT 
			COUNT(*) as total_payments,
			COUNT(CASE WHEN status = 'success' THEN 1 END) as successful_payments,
			COUNT(CASE WHEN status = 'failed' THEN 1 END) as failed_payments,
			COALESCE(SUM(CASE WHEN status = 'success' THEN amount END), 0) as total_amount,
			COALESCE(AVG(CASE WHEN status = 'success' THEN amount END), 0) as avg_amount
		FROM payments 
		WHERE created_at >= $1 AND created_at <= $2`

	var totalPayments, successfulPayments, failedPayments int
	var totalAmount, avgAmount float64

	err := r.db.QueryRowContext(ctx, query, from, to).Scan(
		&totalPayments, &successfulPayments, &failedPayments, &totalAmount, &avgAmount,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get payment stats: %w", err)
	}

	stats := map[string]interface{}{
		"total_payments":     totalPayments,
		"successful_payments": successfulPayments,
		"failed_payments":    failedPayments,
		"total_amount":       totalAmount,
		"average_amount":     avgAmount,
		"success_rate":       float64(successfulPayments) / float64(totalPayments) * 100,
	}

	return stats, nil
}
