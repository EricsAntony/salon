package gateway

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"payment-service/internal/model"

	"github.com/stripe/stripe-go/v76"
	"github.com/stripe/stripe-go/v76/paymentintent"
	"github.com/stripe/stripe-go/v76/refund"
	"github.com/stripe/stripe-go/v76/webhook"
)

// StripeGateway implements PaymentGateway for Stripe
type StripeGateway struct {
	secretKey     string
	webhookSecret string
}

// NewStripeGateway creates a new Stripe gateway instance
func NewStripeGateway(secretKey, webhookSecret string) *StripeGateway {
	stripe.Key = secretKey
	return &StripeGateway{
		secretKey:     secretKey,
		webhookSecret: webhookSecret,
	}
}

// GetName returns the gateway name
func (s *StripeGateway) GetName() string {
	return model.GatewayStripe
}

// InitiatePayment creates a Stripe PaymentIntent
func (s *StripeGateway) InitiatePayment(ctx context.Context, request *PaymentRequest) (*PaymentResponse, error) {
	// Convert amount to cents (Stripe uses smallest currency unit)
	amountCents := int64(request.Amount * 100)

	params := &stripe.PaymentIntentParams{
		Amount:   stripe.Int64(amountCents),
		Currency: stripe.String(request.Currency),
		Metadata: make(map[string]string),
	}

	// Add customer information if provided
	if request.CustomerEmail != "" {
		params.ReceiptEmail = stripe.String(request.CustomerEmail)
	}

	// Add metadata
	if request.Metadata != nil {
		for key, value := range request.Metadata {
			if strValue, ok := value.(string); ok {
				params.Metadata[key] = strValue
			} else {
				// Convert non-string values to JSON
				if jsonValue, err := json.Marshal(value); err == nil {
					params.Metadata[key] = string(jsonValue)
				}
			}
		}
	}

	// Add order ID to metadata
	params.Metadata["order_id"] = request.OrderID
	params.Metadata["customer_id"] = request.CustomerID

	if request.Description != "" {
		params.Description = stripe.String(request.Description)
	}

	// Create PaymentIntent
	pi, err := paymentintent.New(params)
	if err != nil {
		return nil, fmt.Errorf("failed to create Stripe PaymentIntent: %w", err)
	}

	response := &PaymentResponse{
		GatewayPaymentID: pi.ID,
		GatewayOrderID:   pi.ID, // Stripe uses same ID for both
		Status:           MapGatewayStatus(model.GatewayStripe, string(pi.Status)),
		PaymentURL:       "", // Stripe uses client-side integration
		Amount:           request.Amount,
		Currency:         request.Currency,
		Metadata: map[string]interface{}{
			"client_secret": pi.ClientSecret,
			"stripe_payment_intent_id": pi.ID,
		},
	}

	return response, nil
}

// ConfirmPayment confirms a Stripe PaymentIntent
func (s *StripeGateway) ConfirmPayment(ctx context.Context, gatewayPaymentID string) (*PaymentResponse, error) {
	pi, err := paymentintent.Get(gatewayPaymentID, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get Stripe PaymentIntent: %w", err)
	}

	response := &PaymentResponse{
		GatewayPaymentID: pi.ID,
		GatewayOrderID:   pi.ID,
		Status:           MapGatewayStatus(model.GatewayStripe, string(pi.Status)),
		Amount:           float64(pi.Amount) / 100, // Convert from cents
		Currency:         string(pi.Currency),
		Metadata: map[string]interface{}{
			"stripe_payment_intent_id": pi.ID,
		},
	}

	if pi.Status == stripe.PaymentIntentStatusSucceeded {
		processedAt := time.Unix(pi.Created, 0).Format(time.RFC3339)
		response.ProcessedAt = &processedAt

		// Extract payment method information
		if pi.PaymentMethod != nil {
			response.PaymentMethod = string(pi.PaymentMethod.Type)
		}
	}

	return response, nil
}

// RefundPayment processes a refund through Stripe
func (s *StripeGateway) RefundPayment(ctx context.Context, request *RefundRequest) (*RefundResponse, error) {
	// Convert amount to cents
	amountCents := int64(request.Amount * 100)

	params := &stripe.RefundParams{
		PaymentIntent: stripe.String(request.GatewayPaymentID),
		Amount:        stripe.Int64(amountCents),
		Reason:        stripe.String("requested_by_customer"),
		Metadata:      make(map[string]string),
	}

	// Add metadata
	if request.Metadata != nil {
		for key, value := range request.Metadata {
			if strValue, ok := value.(string); ok {
				params.Metadata[key] = strValue
			}
		}
	}
	params.Metadata["reason"] = request.Reason

	// Create refund
	r, err := refund.New(params)
	if err != nil {
		return nil, fmt.Errorf("failed to create Stripe refund: %w", err)
	}

	response := &RefundResponse{
		GatewayRefundID: r.ID,
		Status:          MapGatewayStatus(model.GatewayStripe, string(r.Status)),
		Amount:          float64(r.Amount) / 100, // Convert from cents
		Currency:        string(r.Currency),
		Metadata: map[string]interface{}{
			"stripe_refund_id": r.ID,
		},
	}

	if r.Status == stripe.RefundStatusSucceeded {
		processedAt := time.Unix(r.Created, 0).Format(time.RFC3339)
		response.ProcessedAt = &processedAt
	}

	return response, nil
}

// VerifyWebhook verifies Stripe webhook signature and returns event data
func (s *StripeGateway) VerifyWebhook(ctx context.Context, payload []byte, signature string) (*WebhookEvent, error) {
	event, err := webhook.ConstructEvent(payload, signature, s.webhookSecret)
	if err != nil {
		return nil, fmt.Errorf("failed to verify Stripe webhook: %w", err)
	}

	webhookEvent := &WebhookEvent{}
	webhookEvent.EventType = string(event.Type)
	webhookEvent.Metadata = make(map[string]interface{})

	// Parse event data based on type
	switch event.Type {
	case "payment_intent.succeeded", "payment_intent.payment_failed", "payment_intent.canceled":
		var pi stripe.PaymentIntent
		if err := json.Unmarshal(event.Data.Raw, &pi); err != nil {
			return nil, fmt.Errorf("failed to parse PaymentIntent from webhook: %w", err)
		}

		webhookEvent.GatewayPaymentID = pi.ID
		webhookEvent.Status = MapGatewayStatus(model.GatewayStripe, string(pi.Status))
		webhookEvent.Amount = float64(pi.Amount) / 100
		webhookEvent.Currency = string(pi.Currency)
		webhookEvent.ProcessedAt = time.Unix(pi.Created, 0).Format(time.RFC3339)

		if pi.PaymentMethod != nil {
			webhookEvent.PaymentMethod = string(pi.PaymentMethod.Type)
		}

		// Add metadata
		for key, value := range pi.Metadata {
			webhookEvent.Metadata[key] = value
		}

	case "charge.dispute.created":
		// Handle dispute events if needed
		webhookEvent.EventType = "dispute_created"

	default:
		return nil, fmt.Errorf("unsupported Stripe webhook event type: %s", event.Type)
	}

	return webhookEvent, nil
}

// GetPaymentStatus retrieves current payment status from Stripe
func (s *StripeGateway) GetPaymentStatus(ctx context.Context, gatewayPaymentID string) (*PaymentResponse, error) {
	return s.ConfirmPayment(ctx, gatewayPaymentID)
}
