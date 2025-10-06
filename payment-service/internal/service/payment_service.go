package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	"payment-service/internal/config"
	"payment-service/internal/gateway"
	"payment-service/internal/model"
	"payment-service/internal/repository"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

// PaymentService defines the interface for payment business logic
type PaymentService interface {
	// Payment operations
	InitiatePayment(ctx context.Context, request *model.InitiatePaymentRequest) (*model.PaymentResponse, error)
	ConfirmPayment(ctx context.Context, request *model.ConfirmPaymentRequest) (*model.PaymentResponse, error)
	GetPayment(ctx context.Context, paymentID uuid.UUID) (*model.Payment, error)
	GetPaymentsByBooking(ctx context.Context, bookingID uuid.UUID) ([]*model.Payment, error)
	GetPaymentsByUser(ctx context.Context, userID uuid.UUID, limit, offset int) (*model.PaymentListResponse, error)

	// Refund operations
	RefundPayment(ctx context.Context, request *model.RefundPaymentRequest) (*model.RefundResponse, error)
	GetRefund(ctx context.Context, refundID uuid.UUID) (*model.Refund, error)
	GetRefundsByPayment(ctx context.Context, paymentID uuid.UUID) ([]*model.Refund, error)

	// Webhook operations
	ProcessWebhook(ctx context.Context, gatewayName string, payload []byte, signature string) error

	// Retry operations
	RetryFailedPayment(ctx context.Context, paymentID uuid.UUID) (*model.PaymentResponse, error)

	// Health check
	HealthCheck(ctx context.Context) (*model.HealthResponse, error)
}

type paymentService struct {
	paymentRepo repository.PaymentRepository
	gatewayRepo repository.GatewayRepository
	gatewayMgr  gateway.GatewayManager
	config      *config.Config
}

// NewPaymentService creates a new payment service
func NewPaymentService(
	paymentRepo repository.PaymentRepository,
	gatewayRepo repository.GatewayRepository,
	cfg *config.Config,
) PaymentService {
	gatewayMgr := gateway.NewGatewayManager(cfg)

	return &paymentService{
		paymentRepo: paymentRepo,
		gatewayRepo: gatewayRepo,
		gatewayMgr:  gatewayMgr,
		config:      cfg,
	}
}

// InitiatePayment creates a new payment and initiates it with the gateway
func (s *paymentService) InitiatePayment(ctx context.Context, request *model.InitiatePaymentRequest) (*model.PaymentResponse, error) {
	// Check for existing idempotency record
	if existingRecord, err := s.paymentRepo.GetIdempotencyRecord(ctx, request.IdempotencyKey); err == nil && existingRecord != nil {
		// Return cached response
		var cachedResponse model.PaymentResponse
		if err := json.Unmarshal([]byte(existingRecord.ResponseData), &cachedResponse); err == nil {
			log.Info().
				Str("idempotency_key", request.IdempotencyKey).
				Str("payment_id", existingRecord.PaymentID.String()).
				Msg("Returning cached payment response")
			return &cachedResponse, nil
		}
	}

	// Get payment gateway
	var paymentGateway gateway.PaymentGateway
	var err error

	if request.Gateway != "" {
		paymentGateway, err = s.gatewayMgr.GetGateway(request.Gateway)
		if err != nil {
			return nil, fmt.Errorf("invalid gateway: %w", err)
		}
	} else {
		// Auto-select best gateway
		paymentGateway, err = s.gatewayMgr.SelectBestGateway(request.Amount, request.Currency)
		if err != nil {
			return nil, fmt.Errorf("no available gateway: %w", err)
		}
	}

	// Create payment record
	payment := &model.Payment{
		ID:             uuid.New(),
		BookingID:      request.BookingID,
		UserID:         request.UserID,
		Amount:         request.Amount,
		Currency:       request.Currency,
		Status:         model.PaymentStatusPending,
		Gateway:        paymentGateway.GetName(),
		IdempotencyKey: request.IdempotencyKey,
		ExpiresAt:      timePtr(time.Now().Add(time.Duration(s.config.PaymentTimeoutMinutes) * time.Minute)),
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	// Add metadata
	if request.Metadata != nil {
		metadataJSON, _ := json.Marshal(request.Metadata)
		payment.Metadata = stringPtr(string(metadataJSON))
	}

	// Save payment to database
	if err := s.paymentRepo.Create(ctx, payment); err != nil {
		return nil, fmt.Errorf("failed to create payment record: %w", err)
	}

	// Create payment attempt record
	attempt := &model.PaymentAttempt{
		ID:            uuid.New(),
		PaymentID:     payment.ID,
		AttemptNumber: 1,
		Gateway:       payment.Gateway,
		Status:        model.PaymentStatusPending,
		AttemptedAt:   time.Now(),
		CreatedAt:     time.Now(),
	}

	// Initiate payment with gateway
	gatewayRequest := &gateway.PaymentRequest{
		Amount:        request.Amount,
		Currency:      request.Currency,
		OrderID:       payment.ID.String(),
		CustomerID:    request.UserID.String(),
		Description:   fmt.Sprintf("Payment for booking %s", request.BookingID.String()),
		Metadata:      request.Metadata,
	}

	gatewayResponse, err := paymentGateway.InitiatePayment(ctx, gatewayRequest)
	if err != nil {
		// Update payment status to failed
		payment.Status = model.PaymentStatusFailed
		payment.FailureReason = stringPtr(err.Error())
		s.paymentRepo.Update(ctx, payment)

		// Update attempt record
		attempt.Status = model.PaymentStatusFailed
		attempt.ErrorMessage = stringPtr(err.Error())
		s.paymentRepo.CreateAttempt(ctx, attempt)

		return nil, fmt.Errorf("failed to initiate payment with gateway: %w", err)
	}

	// Update payment with gateway response
	payment.Status = model.PaymentStatusInitiated
	payment.GatewayPaymentID = &gatewayResponse.GatewayPaymentID
	payment.GatewayOrderID = &gatewayResponse.GatewayOrderID
	payment.PaymentURL = &gatewayResponse.PaymentURL

	if err := s.paymentRepo.Update(ctx, payment); err != nil {
		log.Error().Err(err).Str("payment_id", payment.ID.String()).Msg("Failed to update payment after gateway initiation")
	}

	// Update attempt record
	attempt.Status = model.PaymentStatusInitiated
	if gatewayResponse.Metadata != nil {
		responseData, _ := json.Marshal(gatewayResponse)
		attempt.ResponseData = stringPtr(string(responseData))
	}
	s.paymentRepo.CreateAttempt(ctx, attempt)

	// Create response
	response := &model.PaymentResponse{
		Payment:    payment,
		PaymentURL: &gatewayResponse.PaymentURL,
		Message:    "Payment initiated successfully",
	}

	// Cache response for idempotency
	s.cacheIdempotencyResponse(ctx, request, payment.ID, response)

	log.Info().
		Str("payment_id", payment.ID.String()).
		Str("gateway", payment.Gateway).
		Float64("amount", payment.Amount).
		Msg("Payment initiated successfully")

	return response, nil
}

// ConfirmPayment confirms a payment after successful gateway processing
func (s *paymentService) ConfirmPayment(ctx context.Context, request *model.ConfirmPaymentRequest) (*model.PaymentResponse, error) {
	// Get payment
	payment, err := s.paymentRepo.GetByID(ctx, request.PaymentID)
	if err != nil {
		return nil, fmt.Errorf("payment not found: %w", err)
	}

	// Check if payment is already confirmed
	if payment.Status == model.PaymentStatusSuccess {
		return &model.PaymentResponse{
			Payment: payment,
			Message: "Payment already confirmed",
		}, nil
	}

	// Validate payment can be confirmed
	if payment.Status != model.PaymentStatusInitiated {
		return nil, fmt.Errorf("payment cannot be confirmed in status: %s", payment.Status)
	}

	// Check if payment has expired
	if payment.IsExpired() {
		payment.Status = model.PaymentStatusFailed
		payment.FailureReason = stringPtr("Payment expired")
		s.paymentRepo.Update(ctx, payment)
		return nil, fmt.Errorf("payment has expired")
	}

	// Get gateway and confirm payment
	paymentGateway, err := s.gatewayMgr.GetGateway(payment.Gateway)
	if err != nil {
		return nil, fmt.Errorf("gateway not available: %w", err)
	}

	gatewayResponse, err := paymentGateway.ConfirmPayment(ctx, request.GatewayPaymentID)
	if err != nil {
		// Update payment status to failed
		payment.Status = model.PaymentStatusFailed
		payment.FailureReason = stringPtr(err.Error())
		s.paymentRepo.Update(ctx, payment)

		return nil, fmt.Errorf("failed to confirm payment with gateway: %w", err)
	}

	// Update payment with confirmation details
	payment.Status = gateway.MapGatewayStatus(payment.Gateway, gatewayResponse.Status)
	payment.GatewayPaymentID = &gatewayResponse.GatewayPaymentID
	
	if request.PaymentMethod != "" {
		payment.PaymentMethod = &request.PaymentMethod
	} else if gatewayResponse.PaymentMethod != "" {
		payment.PaymentMethod = &gatewayResponse.PaymentMethod
	}

	if gatewayResponse.ProcessedAt != nil {
		if processedTime, err := time.Parse(time.RFC3339, *gatewayResponse.ProcessedAt); err == nil {
			payment.ProcessedAt = &processedTime
		}
	}

	// Add confirmation metadata
	if request.Metadata != nil {
		existingMetadata := make(map[string]interface{})
		if payment.Metadata != nil {
			json.Unmarshal([]byte(*payment.Metadata), &existingMetadata)
		}
		
		for key, value := range request.Metadata {
			existingMetadata[key] = value
		}
		
		metadataJSON, _ := json.Marshal(existingMetadata)
		payment.Metadata = stringPtr(string(metadataJSON))
	}

	if err := s.paymentRepo.Update(ctx, payment); err != nil {
		return nil, fmt.Errorf("failed to update payment: %w", err)
	}

	response := &model.PaymentResponse{
		Payment: payment,
		Message: "Payment confirmed successfully",
	}

	log.Info().
		Str("payment_id", payment.ID.String()).
		Str("gateway_payment_id", request.GatewayPaymentID).
		Str("status", payment.Status).
		Msg("Payment confirmed successfully")

	return response, nil
}

// GetPayment retrieves a payment by ID
func (s *paymentService) GetPayment(ctx context.Context, paymentID uuid.UUID) (*model.Payment, error) {
	return s.paymentRepo.GetByID(ctx, paymentID)
}

// GetPaymentsByBooking retrieves payments for a booking
func (s *paymentService) GetPaymentsByBooking(ctx context.Context, bookingID uuid.UUID) ([]*model.Payment, error) {
	return s.paymentRepo.GetByBookingID(ctx, bookingID)
}

// GetPaymentsByUser retrieves payments for a user with pagination
func (s *paymentService) GetPaymentsByUser(ctx context.Context, userID uuid.UUID, limit, offset int) (*model.PaymentListResponse, error) {
	payments, err := s.paymentRepo.GetByUserID(ctx, userID, limit, offset)
	if err != nil {
		return nil, err
	}

	// Get total count (simplified - in production, you'd want a separate count query)
	totalCount := len(payments)
	if len(payments) == limit {
		totalCount = limit + 1 // Indicate there might be more
	}

	return &model.PaymentListResponse{
		Payments:   payments,
		TotalCount: totalCount,
		Page:       offset/limit + 1,
		PageSize:   limit,
	}, nil
}

// RefundPayment processes a payment refund
func (s *paymentService) RefundPayment(ctx context.Context, request *model.RefundPaymentRequest) (*model.RefundResponse, error) {
	// Check for existing idempotency record
	if existingRecord, err := s.paymentRepo.GetIdempotencyRecord(ctx, request.IdempotencyKey); err == nil && existingRecord != nil {
		// Return cached response (simplified - you'd want a separate refund idempotency table)
		return &model.RefundResponse{
			Message: "Refund already processed",
		}, nil
	}

	// Get payment
	payment, err := s.paymentRepo.GetByID(ctx, request.PaymentID)
	if err != nil {
		return nil, fmt.Errorf("payment not found: %w", err)
	}

	// Validate payment can be refunded
	if !payment.CanBeRefunded() {
		return nil, fmt.Errorf("payment cannot be refunded in status: %s", payment.Status)
	}

	// Determine refund amount
	refundAmount := payment.Amount
	if request.Amount != nil {
		refundAmount = *request.Amount
		if refundAmount > payment.Amount {
			return nil, fmt.Errorf("refund amount cannot exceed payment amount")
		}
	}

	// Create refund record
	refund := &model.Refund{
		ID:             uuid.New(),
		PaymentID:      payment.ID,
		Amount:         refundAmount,
		Currency:       payment.Currency,
		Status:         model.PaymentStatusPending,
		Gateway:        payment.Gateway,
		Reason:         request.Reason,
		IdempotencyKey: request.IdempotencyKey,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	// Add metadata
	if request.Metadata != nil {
		metadataJSON, _ := json.Marshal(request.Metadata)
		refund.Metadata = stringPtr(string(metadataJSON))
	}

	// Save refund to database
	if err := s.paymentRepo.CreateRefund(ctx, refund); err != nil {
		return nil, fmt.Errorf("failed to create refund record: %w", err)
	}

	// Process refund with gateway
	paymentGateway, err := s.gatewayMgr.GetGateway(payment.Gateway)
	if err != nil {
		return nil, fmt.Errorf("gateway not available: %w", err)
	}

	gatewayRequest := &gateway.RefundRequest{
		GatewayPaymentID: *payment.GatewayPaymentID,
		Amount:           refundAmount,
		Reason:           request.Reason,
		Metadata:         request.Metadata,
	}

	gatewayResponse, err := paymentGateway.RefundPayment(ctx, gatewayRequest)
	if err != nil {
		// Update refund status to failed
		refund.Status = model.PaymentStatusFailed
		refund.FailureReason = stringPtr(err.Error())
		s.paymentRepo.UpdateRefund(ctx, refund)

		return nil, fmt.Errorf("failed to process refund with gateway: %w", err)
	}

	// Update refund with gateway response
	refund.Status = gateway.MapGatewayStatus(payment.Gateway, gatewayResponse.Status)
	refund.GatewayRefundID = &gatewayResponse.GatewayRefundID

	if gatewayResponse.ProcessedAt != nil {
		if processedTime, err := time.Parse(time.RFC3339, *gatewayResponse.ProcessedAt); err == nil {
			refund.ProcessedAt = &processedTime
		}
	}

	if err := s.paymentRepo.UpdateRefund(ctx, refund); err != nil {
		log.Error().Err(err).Str("refund_id", refund.ID.String()).Msg("Failed to update refund after gateway processing")
	}

	response := &model.RefundResponse{
		Refund:  refund,
		Message: "Refund processed successfully",
	}

	log.Info().
		Str("refund_id", refund.ID.String()).
		Str("payment_id", payment.ID.String()).
		Float64("amount", refundAmount).
		Msg("Refund processed successfully")

	return response, nil
}

// GetRefund retrieves a refund by ID
func (s *paymentService) GetRefund(ctx context.Context, refundID uuid.UUID) (*model.Refund, error) {
	return s.paymentRepo.GetRefundByID(ctx, refundID)
}

// GetRefundsByPayment retrieves refunds for a payment
func (s *paymentService) GetRefundsByPayment(ctx context.Context, paymentID uuid.UUID) ([]*model.Refund, error) {
	return s.paymentRepo.GetRefundsByPaymentID(ctx, paymentID)
}

// ProcessWebhook processes webhook events from payment gateways
func (s *paymentService) ProcessWebhook(ctx context.Context, gatewayName string, payload []byte, signature string) error {
	// Get gateway
	paymentGateway, err := s.gatewayMgr.GetGateway(gatewayName)
	if err != nil {
		return fmt.Errorf("gateway not found: %w", err)
	}

	// Verify and parse webhook
	webhookEvent, err := paymentGateway.VerifyWebhook(ctx, payload, signature)
	if err != nil {
		return fmt.Errorf("failed to verify webhook: %w", err)
	}

	// Find payment by gateway payment ID
	// Note: This is simplified - in production, you'd want an index on gateway_payment_id
	// For now, we'll log the webhook event
	log.Info().
		Str("gateway", gatewayName).
		Str("event_type", webhookEvent.EventType).
		Str("gateway_payment_id", webhookEvent.GatewayPaymentID).
		Str("status", webhookEvent.Status).
		Msg("Webhook event received")

	// TODO: Implement payment status update based on webhook
	// This would involve finding the payment and updating its status

	return nil
}

// RetryFailedPayment retries a failed payment
func (s *paymentService) RetryFailedPayment(ctx context.Context, paymentID uuid.UUID) (*model.PaymentResponse, error) {
	// Get payment
	payment, err := s.paymentRepo.GetByID(ctx, paymentID)
	if err != nil {
		return nil, fmt.Errorf("payment not found: %w", err)
	}

	// Check if payment can be retried
	if payment.Status != model.PaymentStatusFailed {
		return nil, fmt.Errorf("payment cannot be retried in status: %s", payment.Status)
	}

	// Check retry attempts
	attempts, err := s.paymentRepo.GetAttemptsByPaymentID(ctx, paymentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get payment attempts: %w", err)
	}

	if len(attempts) >= s.config.MaxRetryAttempts {
		return nil, fmt.Errorf("maximum retry attempts exceeded")
	}

	// Reset payment status and retry
	payment.Status = model.PaymentStatusPending
	payment.FailureReason = nil
	payment.UpdatedAt = time.Now()

	if err := s.paymentRepo.Update(ctx, payment); err != nil {
		return nil, fmt.Errorf("failed to update payment for retry: %w", err)
	}

	// Create new attempt record
	attempt := &model.PaymentAttempt{
		ID:            uuid.New(),
		PaymentID:     payment.ID,
		AttemptNumber: len(attempts) + 1,
		Gateway:       payment.Gateway,
		Status:        model.PaymentStatusPending,
		AttemptedAt:   time.Now(),
		CreatedAt:     time.Now(),
	}

	// Retry with gateway
	paymentGateway, err := s.gatewayMgr.GetGateway(payment.Gateway)
	if err != nil {
		return nil, fmt.Errorf("gateway not available: %w", err)
	}

	gatewayRequest := &gateway.PaymentRequest{
		Amount:     payment.Amount,
		Currency:   payment.Currency,
		OrderID:    payment.ID.String(),
		CustomerID: payment.UserID.String(),
	}

	gatewayResponse, err := paymentGateway.InitiatePayment(ctx, gatewayRequest)
	if err != nil {
		// Update payment and attempt status to failed
		payment.Status = model.PaymentStatusFailed
		payment.FailureReason = stringPtr(err.Error())
		s.paymentRepo.Update(ctx, payment)

		attempt.Status = model.PaymentStatusFailed
		attempt.ErrorMessage = stringPtr(err.Error())
		s.paymentRepo.CreateAttempt(ctx, attempt)

		return nil, fmt.Errorf("failed to retry payment with gateway: %w", err)
	}

	// Update payment with new gateway response
	payment.Status = model.PaymentStatusInitiated
	payment.GatewayPaymentID = &gatewayResponse.GatewayPaymentID
	payment.PaymentURL = &gatewayResponse.PaymentURL

	if err := s.paymentRepo.Update(ctx, payment); err != nil {
		log.Error().Err(err).Str("payment_id", payment.ID.String()).Msg("Failed to update payment after retry")
	}

	// Update attempt record
	attempt.Status = model.PaymentStatusInitiated
	s.paymentRepo.CreateAttempt(ctx, attempt)

	response := &model.PaymentResponse{
		Payment:    payment,
		PaymentURL: &gatewayResponse.PaymentURL,
		Message:    "Payment retry initiated successfully",
	}

	log.Info().
		Str("payment_id", payment.ID.String()).
		Int("attempt_number", attempt.AttemptNumber).
		Msg("Payment retry initiated successfully")

	return response, nil
}

// HealthCheck performs a health check
func (s *paymentService) HealthCheck(ctx context.Context) (*model.HealthResponse, error) {
	// Check database connectivity
	_, err := s.paymentRepo.GetPaymentStats(ctx, time.Now().Add(-24*time.Hour), time.Now())
	dbStatus := "healthy"
	if err != nil {
		dbStatus = "unhealthy"
	}

	return &model.HealthResponse{
		Status:    "healthy",
		Timestamp: time.Now().Format(time.RFC3339),
		Version:   "1.0.0",
		Database:  dbStatus,
	}, nil
}

// Helper functions

func (s *paymentService) cacheIdempotencyResponse(ctx context.Context, request *model.InitiatePaymentRequest, paymentID uuid.UUID, response *model.PaymentResponse) {
	// Create request hash for validation
	requestData, _ := json.Marshal(request)
	hash := sha256.Sum256(requestData)
	requestHash := hex.EncodeToString(hash[:])

	// Cache response
	responseData, _ := json.Marshal(response)
	
	record := &model.IdempotencyRecord{
		ID:             uuid.New(),
		IdempotencyKey: request.IdempotencyKey,
		PaymentID:      paymentID,
		RequestHash:    requestHash,
		ResponseData:   string(responseData),
		ExpiresAt:      time.Now().Add(time.Duration(s.config.IdempotencyTTLHours) * time.Hour),
		CreatedAt:      time.Now(),
	}

	if err := s.paymentRepo.CreateIdempotencyRecord(ctx, record); err != nil {
		log.Warn().Err(err).Str("idempotency_key", request.IdempotencyKey).Msg("Failed to cache idempotency response")
	}
}

func stringPtr(s string) *string {
	return &s
}

func timePtr(t time.Time) *time.Time {
	return &t
}
