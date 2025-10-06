package gateway

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	"payment-service/internal/model"

	"github.com/razorpay/razorpay-go"
)

// RazorpayGateway implements PaymentGateway for Razorpay
type RazorpayGateway struct {
	client        *razorpay.Client
	keyID         string
	webhookSecret string
}

// NewRazorpayGateway creates a new Razorpay gateway instance
func NewRazorpayGateway(keyID, keySecret, webhookSecret string) *RazorpayGateway {
	client := razorpay.NewClient(keyID, keySecret)
	return &RazorpayGateway{
		client:        client,
		keyID:         keyID,
		webhookSecret: webhookSecret,
	}
}

// GetName returns the gateway name
func (r *RazorpayGateway) GetName() string {
	return model.GatewayRazorpay
}

// InitiatePayment creates a Razorpay Order
func (r *RazorpayGateway) InitiatePayment(ctx context.Context, request *PaymentRequest) (*PaymentResponse, error) {
	// Convert amount to paise (Razorpay uses smallest currency unit)
	amountPaise := int(request.Amount * 100)

	data := map[string]interface{}{
		"amount":   amountPaise,
		"currency": request.Currency,
		"receipt":  request.OrderID,
	}

	// Add notes (metadata)
	if request.Metadata != nil {
		notes := make(map[string]interface{})
		for key, value := range request.Metadata {
			notes[key] = value
		}
		notes["customer_id"] = request.CustomerID
		notes["customer_email"] = request.CustomerEmail
		notes["customer_phone"] = request.CustomerPhone
		data["notes"] = notes
	}

	// Create order
	order, err := r.client.Order.Create(data, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create Razorpay order: %w", err)
	}

	orderID, ok := order["id"].(string)
	if !ok {
		return nil, fmt.Errorf("invalid order ID from Razorpay")
	}

	response := &PaymentResponse{
		GatewayPaymentID: orderID,
		GatewayOrderID:   orderID,
		Status:           StatusPending,
		PaymentURL:       "", // Razorpay uses client-side integration
		Amount:           request.Amount,
		Currency:         request.Currency,
		Metadata: map[string]interface{}{
			"razorpay_order_id": orderID,
			"key_id":           r.keyID,
		},
	}

	return response, nil
}

// ConfirmPayment confirms a Razorpay payment
func (r *RazorpayGateway) ConfirmPayment(ctx context.Context, gatewayPaymentID string) (*PaymentResponse, error) {
	// Fetch payment details
	payment, err := r.client.Payment.Fetch(gatewayPaymentID, nil, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch Razorpay payment: %w", err)
	}

	status, ok := payment["status"].(string)
	if !ok {
		return nil, fmt.Errorf("invalid payment status from Razorpay")
	}

	amount, ok := payment["amount"].(float64)
	if !ok {
		return nil, fmt.Errorf("invalid payment amount from Razorpay")
	}

	currency, ok := payment["currency"].(string)
	if !ok {
		return nil, fmt.Errorf("invalid payment currency from Razorpay")
	}

	orderID, _ := payment["order_id"].(string)

	response := &PaymentResponse{
		GatewayPaymentID: gatewayPaymentID,
		GatewayOrderID:   orderID,
		Status:           MapGatewayStatus(model.GatewayRazorpay, status),
		Amount:           amount / 100, // Convert from paise
		Currency:         currency,
		Metadata: map[string]interface{}{
			"razorpay_payment_id": gatewayPaymentID,
			"razorpay_order_id":   orderID,
		},
	}

	// Extract payment method
	if method, ok := payment["method"].(string); ok {
		response.PaymentMethod = method
	}

	// Set processed time for successful payments
	if status == "captured" || status == "authorized" {
		if createdAt, ok := payment["created_at"].(float64); ok {
			processedAt := time.Unix(int64(createdAt), 0).Format(time.RFC3339)
			response.ProcessedAt = &processedAt
		}
	}

	return response, nil
}

// RefundPayment processes a refund through Razorpay
func (r *RazorpayGateway) RefundPayment(ctx context.Context, request *RefundRequest) (*RefundResponse, error) {
	// Convert amount to paise
	amountPaise := int(request.Amount * 100)

	data := map[string]interface{}{
		"amount": amountPaise,
	}

	// Add notes (metadata)
	if request.Metadata != nil {
		notes := make(map[string]interface{})
		for key, value := range request.Metadata {
			notes[key] = value
		}
		notes["reason"] = request.Reason
		data["notes"] = notes
	}

	// Create refund
	refund, err := r.client.Payment.Refund(request.GatewayPaymentID, amountPaise, data, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create Razorpay refund: %w", err)
	}

	refundID, ok := refund["id"].(string)
	if !ok {
		return nil, fmt.Errorf("invalid refund ID from Razorpay")
	}

	status, ok := refund["status"].(string)
	if !ok {
		return nil, fmt.Errorf("invalid refund status from Razorpay")
	}

	amount, ok := refund["amount"].(float64)
	if !ok {
		return nil, fmt.Errorf("invalid refund amount from Razorpay")
	}

	currency, ok := refund["currency"].(string)
	if !ok {
		return nil, fmt.Errorf("invalid refund currency from Razorpay")
	}

	response := &RefundResponse{
		GatewayRefundID: refundID,
		Status:          MapGatewayStatus(model.GatewayRazorpay, status),
		Amount:          amount / 100, // Convert from paise
		Currency:        currency,
		Metadata: map[string]interface{}{
			"razorpay_refund_id": refundID,
		},
	}

	// Set processed time for processed refunds
	if status == "processed" {
		if createdAt, ok := refund["created_at"].(float64); ok {
			processedAt := time.Unix(int64(createdAt), 0).Format(time.RFC3339)
			response.ProcessedAt = &processedAt
		}
	}

	return response, nil
}

// VerifyWebhook verifies Razorpay webhook signature and returns event data
func (r *RazorpayGateway) VerifyWebhook(ctx context.Context, payload []byte, signature string) (*WebhookEvent, error) {
	// Verify webhook signature
	if !r.verifyWebhookSignature(payload, signature) {
		return nil, fmt.Errorf("invalid Razorpay webhook signature")
	}

	// Parse webhook payload
	var webhookData map[string]interface{}
	if err := json.Unmarshal(payload, &webhookData); err != nil {
		return nil, fmt.Errorf("failed to parse Razorpay webhook payload: %w", err)
	}

	eventType, ok := webhookData["event"].(string)
	if !ok {
		return nil, fmt.Errorf("invalid event type in Razorpay webhook")
	}

	webhookEvent := &WebhookEvent{
		EventType: eventType,
		Metadata:  make(map[string]interface{}),
	}

	// Parse event data based on type
	switch eventType {
	case "payment.captured", "payment.failed", "payment.authorized":
		paymentData, ok := webhookData["payload"].(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("invalid payment data in Razorpay webhook")
		}

		payment, ok := paymentData["payment"].(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("invalid payment entity in Razorpay webhook")
		}

		if paymentID, ok := payment["id"].(string); ok {
			webhookEvent.GatewayPaymentID = paymentID
		}

		if status, ok := payment["status"].(string); ok {
			webhookEvent.Status = MapGatewayStatus(model.GatewayRazorpay, status)
		}

		if amount, ok := payment["amount"].(float64); ok {
			webhookEvent.Amount = amount / 100 // Convert from paise
		}

		if currency, ok := payment["currency"].(string); ok {
			webhookEvent.Currency = currency
		}

		if method, ok := payment["method"].(string); ok {
			webhookEvent.PaymentMethod = method
		}

		if createdAt, ok := payment["created_at"].(float64); ok {
			webhookEvent.ProcessedAt = time.Unix(int64(createdAt), 0).Format(time.RFC3339)
		}

		// Add notes as metadata
		if notes, ok := payment["notes"].(map[string]interface{}); ok {
			for key, value := range notes {
				webhookEvent.Metadata[key] = value
			}
		}

	case "refund.processed", "refund.failed":
		paymentData, ok := webhookData["payload"].(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("invalid refund data in Razorpay webhook")
		}

		refund, ok := paymentData["refund"].(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("invalid refund entity in Razorpay webhook")
		}

		if paymentID, ok := refund["payment_id"].(string); ok {
			webhookEvent.GatewayPaymentID = paymentID
		}

		webhookEvent.Status = StatusRefunded
		if eventType == "refund.failed" {
			webhookEvent.Status = StatusFailed
		}

		if amount, ok := refund["amount"].(float64); ok {
			webhookEvent.Amount = amount / 100 // Convert from paise
		}

		if currency, ok := refund["currency"].(string); ok {
			webhookEvent.Currency = currency
		}

	default:
		return nil, fmt.Errorf("unsupported Razorpay webhook event type: %s", eventType)
	}

	return webhookEvent, nil
}

// GetPaymentStatus retrieves current payment status from Razorpay
func (r *RazorpayGateway) GetPaymentStatus(ctx context.Context, gatewayPaymentID string) (*PaymentResponse, error) {
	return r.ConfirmPayment(ctx, gatewayPaymentID)
}

// verifyWebhookSignature verifies Razorpay webhook signature
func (r *RazorpayGateway) verifyWebhookSignature(payload []byte, signature string) bool {
	expectedSignature := r.generateWebhookSignature(payload)
	return hmac.Equal([]byte(signature), []byte(expectedSignature))
}

// generateWebhookSignature generates expected webhook signature
func (r *RazorpayGateway) generateWebhookSignature(payload []byte) string {
	h := hmac.New(sha256.New, []byte(r.webhookSecret))
	h.Write(payload)
	return hex.EncodeToString(h.Sum(nil))
}
