package model

import (
	"time"

	"github.com/google/uuid"
)

// Payment status constants
const (
	PaymentStatusPending   = "pending"
	PaymentStatusInitiated = "initiated"
	PaymentStatusSuccess   = "success"
	PaymentStatusFailed    = "failed"
	PaymentStatusCanceled  = "canceled"
	PaymentStatusRefunded  = "refunded"
)

// Payment gateway constants
const (
	GatewayStripe    = "stripe"
	GatewayRazorpay  = "razorpay"
)

// Payment method constants
const (
	PaymentMethodCard   = "card"
	PaymentMethodUPI    = "upi"
	PaymentMethodNetbanking = "netbanking"
	PaymentMethodWallet = "wallet"
)

// Payment represents a payment transaction
type Payment struct {
	ID                uuid.UUID  `json:"id" db:"id"`
	BookingID         uuid.UUID  `json:"booking_id" db:"booking_id"`
	UserID            uuid.UUID  `json:"user_id" db:"user_id"`
	Amount            float64    `json:"amount" db:"amount"`
	Currency          string     `json:"currency" db:"currency"`
	Status            string     `json:"status" db:"status"`
	Gateway           string     `json:"gateway" db:"gateway"`
	GatewayPaymentID  *string    `json:"gateway_payment_id,omitempty" db:"gateway_payment_id"`
	GatewayOrderID    *string    `json:"gateway_order_id,omitempty" db:"gateway_order_id"`
	PaymentMethod     *string    `json:"payment_method,omitempty" db:"payment_method"`
	PaymentURL        *string    `json:"payment_url,omitempty" db:"payment_url"`
	IdempotencyKey    string     `json:"idempotency_key" db:"idempotency_key"`
	Metadata          *string    `json:"metadata,omitempty" db:"metadata"`
	FailureReason     *string    `json:"failure_reason,omitempty" db:"failure_reason"`
	ProcessedAt       *time.Time `json:"processed_at,omitempty" db:"processed_at"`
	ExpiresAt         *time.Time `json:"expires_at,omitempty" db:"expires_at"`
	CreatedAt         time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at" db:"updated_at"`
}

// Refund represents a payment refund
type Refund struct {
	ID               uuid.UUID  `json:"id" db:"id"`
	PaymentID        uuid.UUID  `json:"payment_id" db:"payment_id"`
	Amount           float64    `json:"amount" db:"amount"`
	Currency         string     `json:"currency" db:"currency"`
	Status           string     `json:"status" db:"status"`
	Gateway          string     `json:"gateway" db:"gateway"`
	GatewayRefundID  *string    `json:"gateway_refund_id,omitempty" db:"gateway_refund_id"`
	Reason           string     `json:"reason" db:"reason"`
	IdempotencyKey   string     `json:"idempotency_key" db:"idempotency_key"`
	Metadata         *string    `json:"metadata,omitempty" db:"metadata"`
	FailureReason    *string    `json:"failure_reason,omitempty" db:"failure_reason"`
	ProcessedAt      *time.Time `json:"processed_at,omitempty" db:"processed_at"`
	CreatedAt        time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at" db:"updated_at"`
}

// PaymentAttempt represents a payment retry attempt
type PaymentAttempt struct {
	ID            uuid.UUID  `json:"id" db:"id"`
	PaymentID     uuid.UUID  `json:"payment_id" db:"payment_id"`
	AttemptNumber int        `json:"attempt_number" db:"attempt_number"`
	Gateway       string     `json:"gateway" db:"gateway"`
	Status        string     `json:"status" db:"status"`
	ErrorMessage  *string    `json:"error_message,omitempty" db:"error_message"`
	ResponseData  *string    `json:"response_data,omitempty" db:"response_data"`
	AttemptedAt   time.Time  `json:"attempted_at" db:"attempted_at"`
	CreatedAt     time.Time  `json:"created_at" db:"created_at"`
}

// IdempotencyRecord represents an idempotency key record
type IdempotencyRecord struct {
	ID             uuid.UUID `json:"id" db:"id"`
	IdempotencyKey string    `json:"idempotency_key" db:"idempotency_key"`
	PaymentID      uuid.UUID `json:"payment_id" db:"payment_id"`
	RequestHash    string    `json:"request_hash" db:"request_hash"`
	ResponseData   string    `json:"response_data" db:"response_data"`
	ExpiresAt      time.Time `json:"expires_at" db:"expires_at"`
	CreatedAt      time.Time `json:"created_at" db:"created_at"`
}

// IsExpired checks if the payment has expired
func (p *Payment) IsExpired() bool {
	return p.ExpiresAt != nil && time.Now().After(*p.ExpiresAt)
}

// CanBeRefunded checks if the payment can be refunded
func (p *Payment) CanBeRefunded() bool {
	return p.Status == PaymentStatusSuccess
}

// IsTerminalStatus checks if the payment is in a terminal status
func (p *Payment) IsTerminalStatus() bool {
	return p.Status == PaymentStatusSuccess ||
		p.Status == PaymentStatusFailed ||
		p.Status == PaymentStatusCanceled ||
		p.Status == PaymentStatusRefunded
}
