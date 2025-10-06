package api

import (
	"encoding/json"
	"net/http"
	"strconv"

	"payment-service/internal/model"
	"payment-service/internal/service"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"salon-shared/errors"
	"salon-shared/utils"
)

// PaymentHandler handles payment-related HTTP requests
type PaymentHandler struct {
	paymentService service.PaymentService
}

// NewPaymentHandler creates a new payment handler
func NewPaymentHandler(paymentService service.PaymentService) *PaymentHandler {
	return &PaymentHandler{
		paymentService: paymentService,
	}
}

// InitiatePayment handles POST /api/v1/payments/initiate
func (h *PaymentHandler) InitiatePayment(w http.ResponseWriter, r *http.Request) {
	var request model.InitiatePaymentRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		errors.WriteAPIError(w, errors.MapToAPIError(errors.NewValidationError("request_body", "Invalid request body")))
		return
	}

	// Validate required fields
	var validationErrors errors.ValidationErrors
	
	if request.BookingID == uuid.Nil {
		validationErrors = errors.AppendValidationError(validationErrors, "booking_id", "Booking ID is required")
	}
	
	if request.UserID == uuid.Nil {
		validationErrors = errors.AppendValidationError(validationErrors, "user_id", "User ID is required")
	}
	
	if request.Amount <= 0 {
		validationErrors = errors.AppendValidationError(validationErrors, "amount", "Amount must be greater than 0")
	}
	
	if request.Currency == "" {
		validationErrors = errors.AppendValidationError(validationErrors, "currency", "Currency is required")
	}
	
	if request.IdempotencyKey == "" {
		validationErrors = errors.AppendValidationError(validationErrors, "idempotency_key", "Idempotency key is required")
	}

	if len(validationErrors) > 0 {
		errors.WriteAPIError(w, errors.MapToAPIError(validationErrors))
		return
	}

	// Get idempotency key from header if not in body
	if request.IdempotencyKey == "" {
		request.IdempotencyKey = r.Header.Get("X-Idempotency-Key")
		if request.IdempotencyKey == "" {
			errors.WriteAPIError(w, errors.MapToAPIError(errors.NewValidationError("idempotency_key", "Idempotency key required in body or X-Idempotency-Key header")))
			return
		}
	}

	response, err := h.paymentService.InitiatePayment(r.Context(), &request)
	if err != nil {
		log.Error().Err(err).Msg("Failed to initiate payment")
		errors.WriteAPIError(w, errors.MapToAPIError(err))
		return
	}

	utils.WriteJSON(w, http.StatusCreated, response)
}

// ConfirmPayment handles POST /api/v1/payments/confirm
func (h *PaymentHandler) ConfirmPayment(w http.ResponseWriter, r *http.Request) {
	var request model.ConfirmPaymentRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		errors.WriteAPIError(w, errors.MapToAPIError(errors.NewValidationError("request_body", "Invalid request body")))
		return
	}

	// Validate required fields
	var validationErrors errors.ValidationErrors
	
	if request.PaymentID == uuid.Nil {
		validationErrors = errors.AppendValidationError(validationErrors, "payment_id", "Payment ID is required")
	}
	
	if request.GatewayPaymentID == "" {
		validationErrors = errors.AppendValidationError(validationErrors, "gateway_payment_id", "Gateway payment ID is required")
	}

	if len(validationErrors) > 0 {
		errors.WriteAPIError(w, errors.MapToAPIError(validationErrors))
		return
	}

	response, err := h.paymentService.ConfirmPayment(r.Context(), &request)
	if err != nil {
		log.Error().Err(err).Str("payment_id", request.PaymentID.String()).Msg("Failed to confirm payment")
		errors.WriteAPIError(w, errors.MapToAPIError(err))
		return
	}

	utils.WriteJSON(w, http.StatusOK, response)
}

// GetPayment handles GET /api/v1/payments/{paymentID}
func (h *PaymentHandler) GetPayment(w http.ResponseWriter, r *http.Request) {
	paymentIDStr := chi.URLParam(r, "paymentID")
	paymentID, err := uuid.Parse(paymentIDStr)
	if err != nil {
		errors.WriteAPIError(w, errors.MapToAPIError(errors.NewValidationError("payment_id", "Invalid payment ID")))
		return
	}

	payment, err := h.paymentService.GetPayment(r.Context(), paymentID)
	if err != nil {
		log.Error().Err(err).Str("payment_id", paymentID.String()).Msg("Failed to get payment")
		errors.WriteAPIError(w, errors.MapToAPIError(err))
		return
	}

	utils.WriteJSON(w, http.StatusOK, payment)
}

// GetPaymentsByBooking handles GET /api/v1/bookings/{bookingID}/payments
func (h *PaymentHandler) GetPaymentsByBooking(w http.ResponseWriter, r *http.Request) {
	bookingIDStr := chi.URLParam(r, "bookingID")
	bookingID, err := uuid.Parse(bookingIDStr)
	if err != nil {
		errors.WriteAPIError(w, errors.MapToAPIError(errors.NewValidationError("booking_id", "Invalid booking ID")))
		return
	}

	payments, err := h.paymentService.GetPaymentsByBooking(r.Context(), bookingID)
	if err != nil {
		log.Error().Err(err).Str("booking_id", bookingID.String()).Msg("Failed to get payments by booking")
		errors.WriteAPIError(w, errors.MapToAPIError(err))
		return
	}

	utils.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"payments": payments,
		"count":    len(payments),
	})
}

// GetPaymentsByUser handles GET /api/v1/users/{userID}/payments
func (h *PaymentHandler) GetPaymentsByUser(w http.ResponseWriter, r *http.Request) {
	userIDStr := chi.URLParam(r, "userID")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		errors.WriteAPIError(w, errors.MapToAPIError(errors.NewValidationError("user_id", "Invalid user ID")))
		return
	}

	// Parse pagination parameters
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	limit := 20 // default
	if limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 && parsedLimit <= 100 {
			limit = parsedLimit
		}
	}

	offset := 0 // default
	if offsetStr != "" {
		if parsedOffset, err := strconv.Atoi(offsetStr); err == nil && parsedOffset >= 0 {
			offset = parsedOffset
		}
	}

	response, err := h.paymentService.GetPaymentsByUser(r.Context(), userID, limit, offset)
	if err != nil {
		log.Error().Err(err).Str("user_id", userID.String()).Msg("Failed to get payments by user")
		errors.WriteAPIError(w, errors.MapToAPIError(err))
		return
	}

	utils.WriteJSON(w, http.StatusOK, response)
}

// RefundPayment handles POST /api/v1/payments/{paymentID}/refund
func (h *PaymentHandler) RefundPayment(w http.ResponseWriter, r *http.Request) {
	paymentIDStr := chi.URLParam(r, "paymentID")
	paymentID, err := uuid.Parse(paymentIDStr)
	if err != nil {
		errors.WriteAPIError(w, errors.MapToAPIError(errors.NewValidationError("payment_id", "Invalid payment ID")))
		return
	}

	var request model.RefundPaymentRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		errors.WriteAPIError(w, errors.MapToAPIError(errors.NewValidationError("request_body", "Invalid request body")))
		return
	}

	// Set payment ID from URL
	request.PaymentID = paymentID

	// Validate required fields
	var validationErrors errors.ValidationErrors
	
	if request.Reason == "" {
		validationErrors = errors.AppendValidationError(validationErrors, "reason", "Refund reason is required")
	}
	
	if request.IdempotencyKey == "" {
		validationErrors = errors.AppendValidationError(validationErrors, "idempotency_key", "Idempotency key is required")
	}

	if len(validationErrors) > 0 {
		errors.WriteAPIError(w, errors.MapToAPIError(validationErrors))
		return
	}

	response, err := h.paymentService.RefundPayment(r.Context(), &request)
	if err != nil {
		log.Error().Err(err).Str("payment_id", paymentID.String()).Msg("Failed to refund payment")
		errors.WriteAPIError(w, errors.MapToAPIError(err))
		return
	}

	utils.WriteJSON(w, http.StatusCreated, response)
}

// GetRefunds handles GET /api/v1/payments/{paymentID}/refunds
func (h *PaymentHandler) GetRefunds(w http.ResponseWriter, r *http.Request) {
	paymentIDStr := chi.URLParam(r, "paymentID")
	paymentID, err := uuid.Parse(paymentIDStr)
	if err != nil {
		errors.WriteAPIError(w, errors.MapToAPIError(errors.NewValidationError("payment_id", "Invalid payment ID")))
		return
	}

	refunds, err := h.paymentService.GetRefundsByPayment(r.Context(), paymentID)
	if err != nil {
		log.Error().Err(err).Str("payment_id", paymentID.String()).Msg("Failed to get refunds")
		errors.WriteAPIError(w, errors.MapToAPIError(err))
		return
	}

	utils.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"refunds": refunds,
		"count":   len(refunds),
	})
}

// GetRefund handles GET /api/v1/refunds/{refundID}
func (h *PaymentHandler) GetRefund(w http.ResponseWriter, r *http.Request) {
	refundIDStr := chi.URLParam(r, "refundID")
	refundID, err := uuid.Parse(refundIDStr)
	if err != nil {
		errors.WriteAPIError(w, errors.MapToAPIError(errors.NewValidationError("refund_id", "Invalid refund ID")))
		return
	}

	refund, err := h.paymentService.GetRefund(r.Context(), refundID)
	if err != nil {
		log.Error().Err(err).Str("refund_id", refundID.String()).Msg("Failed to get refund")
		errors.WriteAPIError(w, errors.MapToAPIError(err))
		return
	}

	utils.WriteJSON(w, http.StatusOK, refund)
}

// RetryPayment handles POST /api/v1/payments/{paymentID}/retry
func (h *PaymentHandler) RetryPayment(w http.ResponseWriter, r *http.Request) {
	paymentIDStr := chi.URLParam(r, "paymentID")
	paymentID, err := uuid.Parse(paymentIDStr)
	if err != nil {
		errors.WriteAPIError(w, errors.MapToAPIError(errors.NewValidationError("payment_id", "Invalid payment ID")))
		return
	}

	response, err := h.paymentService.RetryFailedPayment(r.Context(), paymentID)
	if err != nil {
		log.Error().Err(err).Str("payment_id", paymentID.String()).Msg("Failed to retry payment")
		errors.WriteAPIError(w, errors.MapToAPIError(err))
		return
	}

	utils.WriteJSON(w, http.StatusOK, response)
}
