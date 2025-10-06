package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

// PaymentClient handles communication with the payment service
type PaymentClient struct {
	baseURL    string
	httpClient *http.Client
}

// NewPaymentClient creates a new payment service client
func NewPaymentClient(baseURL string) *PaymentClient {
	return &PaymentClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// Payment request/response structures
type InitiatePaymentRequest struct {
	BookingID       uuid.UUID `json:"booking_id"`
	UserID          uuid.UUID `json:"user_id"`
	Amount          float64   `json:"amount"`
	Currency        string    `json:"currency"`
	IdempotencyKey  string    `json:"idempotency_key"`
	Description     string    `json:"description,omitempty"`
	Gateway         string    `json:"gateway,omitempty"`
}

type InitiatePaymentResponse struct {
	PaymentID       uuid.UUID              `json:"payment_id"`
	Status          string                 `json:"status"`
	Amount          float64                `json:"amount"`
	Currency        string                 `json:"currency"`
	Gateway         string                 `json:"gateway"`
	GatewayResponse map[string]interface{} `json:"gateway_response"`
	CreatedAt       time.Time              `json:"created_at"`
}

type ConfirmPaymentRequest struct {
	PaymentID         uuid.UUID `json:"payment_id"`
	GatewayPaymentID  string    `json:"gateway_payment_id"`
	GatewaySignature  string    `json:"gateway_signature,omitempty"`
}

type ConfirmPaymentResponse struct {
	PaymentID uuid.UUID `json:"payment_id"`
	Status    string    `json:"status"`
	UpdatedAt time.Time `json:"updated_at"`
}

type RefundPaymentRequest struct {
	PaymentID      uuid.UUID `json:"payment_id"`
	Amount         *float64  `json:"amount,omitempty"` // nil for full refund
	Reason         string    `json:"reason"`
	IdempotencyKey string    `json:"idempotency_key"`
}

type RefundPaymentResponse struct {
	RefundID        uuid.UUID `json:"refund_id"`
	PaymentID       uuid.UUID `json:"payment_id"`
	Amount          float64   `json:"amount"`
	Status          string    `json:"status"`
	Reason          string    `json:"reason"`
	GatewayRefundID string    `json:"gateway_refund_id,omitempty"`
	CreatedAt       time.Time `json:"created_at"`
}

// InitiatePayment initiates a payment for a booking
func (c *PaymentClient) InitiatePayment(ctx context.Context, request *InitiatePaymentRequest) (*InitiatePaymentResponse, error) {
	url := fmt.Sprintf("%s/api/v1/payments/initiate", c.baseURL)
	
	body, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Idempotency-Key", request.IdempotencyKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("payment service returned status %d", resp.StatusCode)
	}

	var response InitiatePaymentResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	log.Info().
		Str("payment_id", response.PaymentID.String()).
		Str("booking_id", request.BookingID.String()).
		Float64("amount", request.Amount).
		Msg("Payment initiated successfully")

	return &response, nil
}

// ConfirmPayment confirms a payment after gateway callback
func (c *PaymentClient) ConfirmPayment(ctx context.Context, request *ConfirmPaymentRequest) (*ConfirmPaymentResponse, error) {
	url := fmt.Sprintf("%s/api/v1/payments/confirm", c.baseURL)
	
	body, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("payment service returned status %d", resp.StatusCode)
	}

	var response ConfirmPaymentResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	log.Info().
		Str("payment_id", response.PaymentID.String()).
		Str("status", response.Status).
		Msg("Payment confirmed successfully")

	return &response, nil
}

// RefundPayment initiates a refund for a payment
func (c *PaymentClient) RefundPayment(ctx context.Context, request *RefundPaymentRequest) (*RefundPaymentResponse, error) {
	url := fmt.Sprintf("%s/api/v1/payments/%s/refund", c.baseURL, request.PaymentID.String())
	
	body, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Idempotency-Key", request.IdempotencyKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("payment service returned status %d", resp.StatusCode)
	}

	var response RefundPaymentResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	log.Info().
		Str("refund_id", response.RefundID.String()).
		Str("payment_id", response.PaymentID.String()).
		Float64("amount", response.Amount).
		Msg("Refund initiated successfully")

	return &response, nil
}

// GetPaymentsByBooking retrieves payments for a booking
func (c *PaymentClient) GetPaymentsByBooking(ctx context.Context, bookingID uuid.UUID) ([]InitiatePaymentResponse, error) {
	url := fmt.Sprintf("%s/api/v1/bookings/%s/payments", c.baseURL, bookingID.String())
	
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("payment service returned status %d", resp.StatusCode)
	}

	var response struct {
		Payments []InitiatePaymentResponse `json:"payments"`
		Count    int                       `json:"count"`
	}
	
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return response.Payments, nil
}
