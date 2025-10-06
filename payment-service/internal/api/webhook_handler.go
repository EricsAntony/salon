package api

import (
	"io"
	"net/http"

	"payment-service/internal/service"

	"github.com/rs/zerolog/log"
	"salon-shared/errors"
	"salon-shared/utils"
)

// WebhookHandler handles webhook requests from payment gateways
type WebhookHandler struct {
	paymentService service.PaymentService
}

// NewWebhookHandler creates a new webhook handler
func NewWebhookHandler(paymentService service.PaymentService) *WebhookHandler {
	return &WebhookHandler{
		paymentService: paymentService,
	}
}

// StripeWebhook handles POST /webhooks/stripe
func (h *WebhookHandler) StripeWebhook(w http.ResponseWriter, r *http.Request) {
	// Read the request body
	payload, err := io.ReadAll(r.Body)
	if err != nil {
		log.Error().Err(err).Msg("Failed to read Stripe webhook payload")
		errors.WriteAPIError(w, errors.MapToAPIError(errors.NewValidationError("request_body", "Failed to read request body")))
		return
	}

	// Get Stripe signature from header
	signature := r.Header.Get("Stripe-Signature")
	if signature == "" {
		log.Warn().Msg("Missing Stripe signature header")
		errors.WriteAPIError(w, errors.MapToAPIError(errors.NewValidationError("signature", "Missing Stripe-Signature header")))
		return
	}

	// Process webhook
	err = h.paymentService.ProcessWebhook(r.Context(), "stripe", payload, signature)
	if err != nil {
		log.Error().Err(err).Msg("Failed to process Stripe webhook")
		errors.WriteAPIError(w, errors.MapToAPIError(err))
		return
	}

	// Return success response
	utils.WriteJSON(w, http.StatusOK, map[string]string{
		"status": "success",
	})

	log.Info().Msg("Stripe webhook processed successfully")
}

// RazorpayWebhook handles POST /webhooks/razorpay
func (h *WebhookHandler) RazorpayWebhook(w http.ResponseWriter, r *http.Request) {
	// Read the request body
	payload, err := io.ReadAll(r.Body)
	if err != nil {
		log.Error().Err(err).Msg("Failed to read Razorpay webhook payload")
		errors.WriteAPIError(w, errors.MapToAPIError(errors.NewValidationError("request_body", "Failed to read request body")))
		return
	}

	// Get Razorpay signature from header
	signature := r.Header.Get("X-Razorpay-Signature")
	if signature == "" {
		log.Warn().Msg("Missing Razorpay signature header")
		errors.WriteAPIError(w, errors.MapToAPIError(errors.NewValidationError("signature", "Missing X-Razorpay-Signature header")))
		return
	}

	// Process webhook
	err = h.paymentService.ProcessWebhook(r.Context(), "razorpay", payload, signature)
	if err != nil {
		log.Error().Err(err).Msg("Failed to process Razorpay webhook")
		errors.WriteAPIError(w, errors.MapToAPIError(err))
		return
	}

	// Return success response
	utils.WriteJSON(w, http.StatusOK, map[string]string{
		"status": "success",
	})

	log.Info().Msg("Razorpay webhook processed successfully")
}
