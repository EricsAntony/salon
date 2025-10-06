package model

import (
	"github.com/google/uuid"
)

// InitiatePaymentRequest represents a request to initiate a payment
type InitiatePaymentRequest struct {
	BookingID      uuid.UUID `json:"booking_id" validate:"required"`
	UserID         uuid.UUID `json:"user_id" validate:"required"`
	Amount         float64   `json:"amount" validate:"required,gt=0"`
	Currency       string    `json:"currency" validate:"required,len=3"`
	Gateway        string    `json:"gateway" validate:"required,oneof=stripe razorpay"`
	IdempotencyKey string    `json:"idempotency_key" validate:"required"`
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
}

// ConfirmPaymentRequest represents a request to confirm a payment
type ConfirmPaymentRequest struct {
	PaymentID        uuid.UUID `json:"payment_id" validate:"required"`
	GatewayPaymentID string    `json:"gateway_payment_id" validate:"required"`
	PaymentMethod    string    `json:"payment_method,omitempty"`
	Metadata         map[string]interface{} `json:"metadata,omitempty"`
}

// RefundPaymentRequest represents a request to refund a payment
type RefundPaymentRequest struct {
	PaymentID      uuid.UUID `json:"payment_id" validate:"required"`
	Amount         *float64  `json:"amount,omitempty"` // nil for full refund
	Reason         string    `json:"reason" validate:"required"`
	IdempotencyKey string    `json:"idempotency_key" validate:"required"`
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
}

// WebhookRequest represents a webhook payload from payment gateways
type WebhookRequest struct {
	Gateway   string                 `json:"gateway"`
	EventType string                 `json:"event_type"`
	Data      map[string]interface{} `json:"data"`
	Signature string                 `json:"signature"`
}

// PaymentResponse represents a payment response
type PaymentResponse struct {
	Payment    *Payment `json:"payment"`
	PaymentURL *string  `json:"payment_url,omitempty"`
	Message    string   `json:"message"`
}

// RefundResponse represents a refund response
type RefundResponse struct {
	Refund  *Refund `json:"refund"`
	Message string  `json:"message"`
}

// PaymentListResponse represents a list of payments
type PaymentListResponse struct {
	Payments   []*Payment `json:"payments"`
	TotalCount int        `json:"total_count"`
	Page       int        `json:"page"`
	PageSize   int        `json:"page_size"`
}

// HealthResponse represents a health check response
type HealthResponse struct {
	Status    string `json:"status"`
	Timestamp string `json:"timestamp"`
	Version   string `json:"version"`
	Database  string `json:"database"`
}
