package gateway

import (
	"context"

	"payment-service/internal/model"
)

// PaymentGateway defines the interface for payment gateway implementations
type PaymentGateway interface {
	// GetName returns the gateway name
	GetName() string

	// InitiatePayment creates a payment intent/order and returns payment URL
	InitiatePayment(ctx context.Context, request *PaymentRequest) (*PaymentResponse, error)

	// ConfirmPayment confirms a payment using gateway payment ID
	ConfirmPayment(ctx context.Context, gatewayPaymentID string) (*PaymentResponse, error)

	// RefundPayment processes a refund
	RefundPayment(ctx context.Context, request *RefundRequest) (*RefundResponse, error)

	// VerifyWebhook verifies webhook signature and returns event data
	VerifyWebhook(ctx context.Context, payload []byte, signature string) (*WebhookEvent, error)

	// GetPaymentStatus retrieves current payment status from gateway
	GetPaymentStatus(ctx context.Context, gatewayPaymentID string) (*PaymentResponse, error)
}

// PaymentRequest represents a payment initiation request
type PaymentRequest struct {
	Amount        float64                `json:"amount"`
	Currency      string                 `json:"currency"`
	OrderID       string                 `json:"order_id"`
	CustomerID    string                 `json:"customer_id"`
	CustomerEmail string                 `json:"customer_email"`
	CustomerPhone string                 `json:"customer_phone"`
	Description   string                 `json:"description"`
	CallbackURL   string                 `json:"callback_url"`
	Metadata      map[string]interface{} `json:"metadata"`
}

// PaymentResponse represents a payment response from gateway
type PaymentResponse struct {
	GatewayPaymentID string                 `json:"gateway_payment_id"`
	GatewayOrderID   string                 `json:"gateway_order_id"`
	Status           string                 `json:"status"`
	PaymentURL       string                 `json:"payment_url"`
	PaymentMethod    string                 `json:"payment_method"`
	Amount           float64                `json:"amount"`
	Currency         string                 `json:"currency"`
	ProcessedAt      *string                `json:"processed_at"`
	Metadata         map[string]interface{} `json:"metadata"`
	ErrorMessage     string                 `json:"error_message"`
}

// RefundRequest represents a refund request
type RefundRequest struct {
	GatewayPaymentID string                 `json:"gateway_payment_id"`
	Amount           float64                `json:"amount"`
	Reason           string                 `json:"reason"`
	Metadata         map[string]interface{} `json:"metadata"`
}

// RefundResponse represents a refund response from gateway
type RefundResponse struct {
	GatewayRefundID  string                 `json:"gateway_refund_id"`
	Status           string                 `json:"status"`
	Amount           float64                `json:"amount"`
	Currency         string                 `json:"currency"`
	ProcessedAt      *string                `json:"processed_at"`
	Metadata         map[string]interface{} `json:"metadata"`
	ErrorMessage     string                 `json:"error_message"`
}

// WebhookEvent represents a webhook event from gateway
type WebhookEvent struct {
	EventType        string                 `json:"event_type"`
	GatewayPaymentID string                 `json:"gateway_payment_id"`
	Status           string                 `json:"status"`
	Amount           float64                `json:"amount"`
	Currency         string                 `json:"currency"`
	PaymentMethod    string                 `json:"payment_method"`
	ProcessedAt      string                 `json:"processed_at"`
	Metadata         map[string]interface{} `json:"metadata"`
}

// GatewayManager manages multiple payment gateways
type GatewayManager interface {
	GetGateway(name string) (PaymentGateway, error)
	GetAvailableGateways() []string
	SelectBestGateway(amount float64, currency string) (PaymentGateway, error)
}

// Status mapping constants
const (
	StatusPending   = "pending"
	StatusInitiated = "initiated"
	StatusSuccess   = "success"
	StatusFailed    = "failed"
	StatusCanceled  = "canceled"
	StatusRefunded  = "refunded"
)

// Gateway-specific status mapping
func MapGatewayStatus(gateway, status string) string {
	switch gateway {
	case model.GatewayStripe:
		return mapStripeStatus(status)
	case model.GatewayRazorpay:
		return mapRazorpayStatus(status)
	default:
		return status
	}
}

func mapStripeStatus(status string) string {
	switch status {
	case "requires_payment_method", "requires_confirmation":
		return StatusPending
	case "processing":
		return StatusInitiated
	case "succeeded":
		return StatusSuccess
	case "canceled":
		return StatusCanceled
	case "requires_action":
		return StatusPending
	default:
		return StatusFailed
	}
}

func mapRazorpayStatus(status string) string {
	switch status {
	case "created":
		return StatusPending
	case "authorized", "captured":
		return StatusSuccess
	case "failed":
		return StatusFailed
	case "refunded":
		return StatusRefunded
	default:
		return StatusPending
	}
}
