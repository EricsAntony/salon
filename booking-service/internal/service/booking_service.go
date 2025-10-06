package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"booking-service/internal/config"
	"booking-service/internal/model"
	"booking-service/internal/repository"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

// BookingService defines the interface for booking business logic
type BookingService interface {
	// Booking lifecycle
	InitiateBooking(ctx context.Context, request *InitiateBookingRequest) (*model.Booking, error)
	ConfirmBooking(ctx context.Context, bookingID uuid.UUID, paymentID string) (*model.Booking, error)
	CancelBooking(ctx context.Context, bookingID uuid.UUID, userID uuid.UUID, reason string) error
	RescheduleBooking(ctx context.Context, request *RescheduleBookingRequest) (*model.Booking, error)
	
	// Booking queries
	GetBooking(ctx context.Context, bookingID uuid.UUID) (*model.Booking, error)
	GetUserBookings(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*model.Booking, error)
	
	// Payment integration
	InitiatePaymentForBooking(ctx context.Context, bookingID uuid.UUID, gateway string) (*InitiatePaymentResponse, error)
	ProcessPaymentCallback(ctx context.Context, bookingID uuid.UUID, paymentID uuid.UUID, gatewayPaymentID string) error
	RefundBookingPayment(ctx context.Context, bookingID uuid.UUID, reason string, amount *float64) (*RefundPaymentResponse, error)
	
	// Availability and pricing
	GetStylistAvailability(ctx context.Context, salonID, stylistID uuid.UUID, date time.Time) ([]*model.TimeSlot, error)
	CalculateBookingSummary(ctx context.Context, request *BookingSummaryRequest) (*model.BookingSummary, error)
	
	// Configuration
	GetBranchConfiguration(ctx context.Context, branchID uuid.UUID) (*model.BranchConfiguration, error)
}

type bookingService struct {
	repo               repository.BookingRepository
	externalService    ExternalService
	paymentClient      *PaymentClient
	notificationClient *NotificationClient
	config             *config.Config
}

// NewBookingService creates a new booking service
func NewBookingService(repo repository.BookingRepository, cfg *config.Config) BookingService {
	externalService := NewExternalService(cfg.UserServiceURL, cfg.SalonServiceURL)
	paymentClient := NewPaymentClient(cfg.PaymentServiceURL)
	notificationClient := NewNotificationClient(cfg.NotificationServiceURL)
	
	return &bookingService{
		repo:               repo,
		externalService:    externalService,
		paymentClient:      paymentClient,
		notificationClient: notificationClient,
		config:             cfg,
	}
}

// Request/Response structures
type InitiateBookingRequest struct {
	UserID   uuid.UUID                    `json:"user_id"`
	SalonID  uuid.UUID                    `json:"salon_id"`
	BranchID uuid.UUID                    `json:"branch_id"`
	Services []InitiateBookingServiceItem `json:"services"`
	Notes    *string                      `json:"notes,omitempty"`
}

type InitiateBookingServiceItem struct {
	ServiceID uuid.UUID `json:"service_id"`
	StylistID uuid.UUID `json:"stylist_id"`
	StartTime time.Time `json:"start_time"`
}

type RescheduleBookingRequest struct {
	BookingID uuid.UUID                    `json:"booking_id"`
	UserID    uuid.UUID                    `json:"user_id"`
	Services  []InitiateBookingServiceItem `json:"services"`
	Reason    string                       `json:"reason"`
}

type BookingSummaryRequest struct {
	SalonID  uuid.UUID                    `json:"salon_id"`
	BranchID uuid.UUID                    `json:"branch_id"`
	Services []InitiateBookingServiceItem `json:"services"`
}

// InitiateBooking creates a new booking in initiated status
func (s *bookingService) InitiateBooking(ctx context.Context, request *InitiateBookingRequest) (*model.Booking, error) {
	// Validate user exists
	user, err := s.externalService.ValidateUser(ctx, request.UserID)
	if err != nil {
		return nil, fmt.Errorf("invalid user: %w", err)
	}

	// Validate branch exists
	_, err = s.externalService.GetBranch(ctx, request.SalonID, request.BranchID)
	if err != nil {
		return nil, fmt.Errorf("invalid branch: %w", err)
	}

	// Get branch configuration
	branchConfig, err := s.getBranchConfigWithDefaults(ctx, request.BranchID)
	if err != nil {
		return nil, fmt.Errorf("failed to get branch configuration: %w", err)
	}

	// Validate and process services
	var bookingServices []model.BookingService
	var totalAmount float64

	for _, serviceItem := range request.Services {
		// Validate service exists
		serviceInfo, err := s.externalService.GetService(ctx, request.SalonID, serviceItem.ServiceID)
		if err != nil {
			return nil, fmt.Errorf("invalid service %s: %w", serviceItem.ServiceID, err)
		}

		// Validate stylist exists and can perform this service
		stylist, err := s.externalService.GetStylist(ctx, request.SalonID, serviceItem.StylistID)
		if err != nil {
			return nil, fmt.Errorf("invalid stylist %s: %w", serviceItem.StylistID, err)
		}

		// Check if stylist belongs to the same branch
		if stylist.BranchID != request.BranchID {
			return nil, fmt.Errorf("stylist %s does not belong to branch %s", serviceItem.StylistID, request.BranchID)
		}

		// Calculate end time based on service duration and buffer
		endTime := serviceItem.StartTime.Add(time.Duration(serviceInfo.Duration) * time.Minute)
		
		// Check availability with buffer
		available, err := s.checkAvailabilityWithBuffer(ctx, serviceItem.StylistID, serviceItem.StartTime, endTime, branchConfig.BufferTimeMinutes)
		if err != nil {
			return nil, fmt.Errorf("failed to check availability: %w", err)
		}
		if !available {
			return nil, fmt.Errorf("stylist %s is not available at %s", serviceItem.StylistID, serviceItem.StartTime.Format("2006-01-02 15:04"))
		}

		// Create booking service
		bookingService := model.BookingService{
			ID:        uuid.New(),
			ServiceID: serviceItem.ServiceID,
			StylistID: serviceItem.StylistID,
			StartTime: serviceItem.StartTime,
			EndTime:   endTime,
			Price:     serviceInfo.Price,
		}

		bookingServices = append(bookingServices, bookingService)
		totalAmount += serviceInfo.Price
	}

	// Calculate GST and total
	gst := totalAmount * (branchConfig.GSTPercentage / 100)
	finalTotal := totalAmount + branchConfig.BookingFeeAmount + gst

	// Create booking
	booking := &model.Booking{
		ID:            uuid.New(),
		UserID:        request.UserID,
		SalonID:       request.SalonID,
		BranchID:      request.BranchID,
		Status:        model.BookingStatusInitiated,
		TotalAmount:   finalTotal,
		GST:           gst,
		BookingFee:    branchConfig.BookingFeeAmount,
		PaymentStatus: model.PaymentStatusPending,
		Notes:         request.Notes,
	}

	// Save booking
	if err := s.repo.Create(ctx, booking); err != nil {
		return nil, fmt.Errorf("failed to create booking: %w", err)
	}

	// Save booking services
	for i := range bookingServices {
		bookingServices[i].BookingID = booking.ID
		if err := s.repo.CreateBookingService(ctx, &bookingServices[i]); err != nil {
			return nil, fmt.Errorf("failed to create booking service: %w", err)
		}
	}

	// Create history entry
	historyData := map[string]interface{}{
		"user_name": user.Name,
		"services":  len(bookingServices),
		"total":     finalTotal,
	}
	historyJSON, _ := json.Marshal(historyData)
	history := &model.BookingHistory{
		ID:        uuid.New(),
		BookingID: booking.ID,
		Action:    model.BookingActionCreated,
		NewValues: stringPtr(string(historyJSON)),
		UserID:    &request.UserID,
	}
	if err := s.repo.CreateHistory(ctx, history); err != nil {
		log.Warn().Err(err).Msg("Failed to create booking history")
	}

	// Load services for response
	booking.Services = bookingServices

	log.Info().
		Str("booking_id", booking.ID.String()).
		Str("user_id", request.UserID.String()).
		Float64("total", finalTotal).
		Msg("Booking initiated successfully")

	return booking, nil
}

// InitiatePaymentForBooking initiates payment for a booking
func (s *bookingService) InitiatePaymentForBooking(ctx context.Context, bookingID uuid.UUID, gateway string) (*InitiatePaymentResponse, error) {
	// Get booking
	booking, err := s.repo.GetByID(ctx, bookingID)
	if err != nil {
		return nil, fmt.Errorf("booking not found: %w", err)
	}

	// Validate booking status
	if booking.Status != model.BookingStatusInitiated {
		return nil, fmt.Errorf("payment can only be initiated for bookings in initiated status, current status: %s", booking.Status)
	}

	// Validate user exists
	_, err = s.externalService.ValidateUser(ctx, booking.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user details: %w", err)
	}

	// Generate idempotency key
	idempotencyKey := fmt.Sprintf("booking-%s-%d", bookingID.String(), time.Now().Unix())

	// Prepare payment request
	paymentRequest := &InitiatePaymentRequest{
		BookingID:      bookingID,
		UserID:         booking.UserID,
		Amount:         booking.TotalAmount,
		Currency:       "INR", // Default currency
		IdempotencyKey: idempotencyKey,
		Description:    fmt.Sprintf("Payment for booking %s", bookingID.String()),
		Gateway:        gateway,
	}

	// Initiate payment
	paymentResponse, err := s.paymentClient.InitiatePayment(ctx, paymentRequest)
	if err != nil {
		return nil, fmt.Errorf("failed to initiate payment: %w", err)
	}

	log.Info().
		Str("booking_id", bookingID.String()).
		Str("payment_id", paymentResponse.PaymentID.String()).
		Str("gateway", gateway).
		Float64("amount", booking.TotalAmount).
		Msg("Payment initiated for booking")

	return paymentResponse, nil
}

// ProcessPaymentCallback processes payment callback and confirms booking
func (s *bookingService) ProcessPaymentCallback(ctx context.Context, bookingID uuid.UUID, paymentID uuid.UUID, gatewayPaymentID string) error {
	// Confirm payment
	confirmRequest := &ConfirmPaymentRequest{
		PaymentID:        paymentID,
		GatewayPaymentID: gatewayPaymentID,
	}

	_, err := s.paymentClient.ConfirmPayment(ctx, confirmRequest)
	if err != nil {
		return fmt.Errorf("failed to confirm payment: %w", err)
	}

	// Confirm booking
	confirmedBooking, err := s.ConfirmBooking(ctx, bookingID, paymentID.String())
	if err != nil {
		return fmt.Errorf("failed to confirm booking: %w", err)
	}

	// Send confirmation notifications
	go s.sendBookingConfirmationNotifications(ctx, confirmedBooking)

	log.Info().
		Str("booking_id", bookingID.String()).
		Str("payment_id", paymentID.String()).
		Msg("Payment callback processed and booking confirmed")

	return nil
}

// RefundBookingPayment initiates a refund for a booking payment
func (s *bookingService) RefundBookingPayment(ctx context.Context, bookingID uuid.UUID, reason string, amount *float64) (*RefundPaymentResponse, error) {
	// Get booking
	booking, err := s.repo.GetByID(ctx, bookingID)
	if err != nil {
		return nil, fmt.Errorf("booking not found: %w", err)
	}

	// Validate booking has payment
	if booking.PaymentID == nil {
		return nil, fmt.Errorf("booking has no associated payment")
	}

	// Parse payment ID
	paymentID, err := uuid.Parse(*booking.PaymentID)
	if err != nil {
		return nil, fmt.Errorf("invalid payment ID: %w", err)
	}

	// Generate idempotency key for refund
	idempotencyKey := fmt.Sprintf("refund-%s-%d", bookingID.String(), time.Now().Unix())

	// Prepare refund request
	refundRequest := &RefundPaymentRequest{
		PaymentID:      paymentID,
		Amount:         amount, // nil for full refund
		Reason:         reason,
		IdempotencyKey: idempotencyKey,
	}

	// Initiate refund
	refundResponse, err := s.paymentClient.RefundPayment(ctx, refundRequest)
	if err != nil {
		return nil, fmt.Errorf("failed to initiate refund: %w", err)
	}

	// Update booking payment status
	booking.PaymentStatus = model.PaymentStatusRefunded
	if err := s.repo.Update(ctx, booking); err != nil {
		log.Error().Err(err).Msg("Failed to update booking payment status after refund")
	}

	log.Info().
		Str("booking_id", bookingID.String()).
		Str("refund_id", refundResponse.RefundID.String()).
		Float64("amount", refundResponse.Amount).
		Msg("Refund initiated for booking")

	return refundResponse, nil
}

// ConfirmBooking confirms a booking after successful payment
func (s *bookingService) ConfirmBooking(ctx context.Context, bookingID uuid.UUID, paymentID string) (*model.Booking, error) {
	// Get booking
	booking, err := s.repo.GetByID(ctx, bookingID)
	if err != nil {
		return nil, fmt.Errorf("booking not found: %w", err)
	}

	// Validate booking status
	if booking.Status != model.BookingStatusInitiated {
		return nil, fmt.Errorf("booking cannot be confirmed in status: %s", booking.Status)
	}

	// Update booking status
	booking.Status = model.BookingStatusConfirmed
	booking.PaymentStatus = model.PaymentStatusPaid
	booking.PaymentID = &paymentID

	if err := s.repo.Update(ctx, booking); err != nil {
		return nil, fmt.Errorf("failed to update booking: %w", err)
	}

	// Create history entry
	history := &model.BookingHistory{
		ID:        uuid.New(),
		BookingID: booking.ID,
		Action:    model.BookingActionConfirmed,
		NewValues: stringPtr(fmt.Sprintf(`{"payment_id": "%s"}`, paymentID)),
	}
	if err := s.repo.CreateHistory(ctx, history); err != nil {
		log.Warn().Err(err).Msg("Failed to create booking history")
	}

	log.Info().
		Str("booking_id", booking.ID.String()).
		Str("payment_id", paymentID).
		Msg("Booking confirmed successfully")

	return booking, nil
}

// Helper function to get branch configuration with defaults
func (s *bookingService) getBranchConfigWithDefaults(ctx context.Context, branchID uuid.UUID) (*model.BranchConfiguration, error) {
	config, err := s.repo.GetBranchConfiguration(ctx, branchID)
	if err != nil {
		// If configuration doesn't exist, create default one
		config = &model.BranchConfiguration{
			BranchID:                branchID,
			BufferTimeMinutes:       s.config.DefaultBufferTimeMinutes,
			CancellationCutoffHours: s.config.DefaultCancellationCutoffHours,
			RescheduleWindowHours:   s.config.DefaultRescheduleWindowHours,
			MaxAdvanceBookingDays:   s.config.DefaultMaxAdvanceBookingDays,
			BookingFeeAmount:        s.config.DefaultBookingFeeAmount,
			GSTPercentage:           s.config.DefaultGSTPercentage,
		}
		
		if createErr := s.repo.CreateBranchConfiguration(ctx, config); createErr != nil {
			log.Warn().Err(createErr).Str("branch_id", branchID.String()).Msg("Failed to create default branch configuration")
		}
	}
	return config, nil
}

// Helper function to check availability with buffer time
func (s *bookingService) checkAvailabilityWithBuffer(ctx context.Context, stylistID uuid.UUID, startTime, endTime time.Time, bufferMinutes int) (bool, error) {
	// Add buffer time to the requested slot
	bufferedStart := startTime.Add(-time.Duration(bufferMinutes) * time.Minute)
	bufferedEnd := endTime.Add(time.Duration(bufferMinutes) * time.Minute)
	
	return s.repo.CheckStylistAvailability(ctx, stylistID, bufferedStart, bufferedEnd)
}

// CancelBooking cancels a booking
func (s *bookingService) CancelBooking(ctx context.Context, bookingID uuid.UUID, userID uuid.UUID, reason string) error {
	// Get booking
	booking, err := s.repo.GetByID(ctx, bookingID)
	if err != nil {
		return fmt.Errorf("booking not found: %w", err)
	}

	// Validate user owns the booking
	if booking.UserID != userID {
		return fmt.Errorf("unauthorized: user does not own this booking")
	}

	// Check if booking can be canceled
	branchConfig, err := s.getBranchConfigWithDefaults(ctx, booking.BranchID)
	if err != nil {
		return fmt.Errorf("failed to get branch configuration: %w", err)
	}

	if !booking.CanBeCanceled(branchConfig.CancellationCutoffHours) {
		return fmt.Errorf("booking cannot be canceled within %d hours of appointment", branchConfig.CancellationCutoffHours)
	}

	// Update booking status
	if err := s.repo.UpdateStatus(ctx, bookingID, model.BookingStatusCanceled); err != nil {
		return fmt.Errorf("failed to cancel booking: %w", err)
	}

	// Create history entry
	history := &model.BookingHistory{
		ID:        uuid.New(),
		BookingID: bookingID,
		Action:    model.BookingActionCanceled,
		Reason:    &reason,
		UserID:    &userID,
	}
	if err := s.repo.CreateHistory(ctx, history); err != nil {
		log.Warn().Err(err).Msg("Failed to create booking history")
	}

	// Send cancellation notifications
	go s.sendBookingCancellationNotifications(ctx, booking, reason)

	log.Info().
		Str("booking_id", bookingID.String()).
		Str("user_id", userID.String()).
		Str("reason", reason).
		Msg("Booking canceled successfully")

	return nil
}

// GetBooking retrieves a booking by ID
func (s *bookingService) GetBooking(ctx context.Context, bookingID uuid.UUID) (*model.Booking, error) {
	return s.repo.GetByID(ctx, bookingID)
}

// GetUserBookings retrieves bookings for a user
func (s *bookingService) GetUserBookings(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*model.Booking, error) {
	return s.repo.GetByUserID(ctx, userID, limit, offset)
}

// GetBranchConfiguration retrieves branch configuration
func (s *bookingService) GetBranchConfiguration(ctx context.Context, branchID uuid.UUID) (*model.BranchConfiguration, error) {
	return s.getBranchConfigWithDefaults(ctx, branchID)
}

// RescheduleBooking reschedules an existing booking
func (s *bookingService) RescheduleBooking(ctx context.Context, request *RescheduleBookingRequest) (*model.Booking, error) {
	// Get existing booking
	booking, err := s.repo.GetByID(ctx, request.BookingID)
	if err != nil {
		return nil, fmt.Errorf("booking not found: %w", err)
	}

	// Validate user owns the booking
	if booking.UserID != request.UserID {
		return nil, fmt.Errorf("unauthorized: user does not own this booking")
	}

	// Check if booking can be rescheduled
	branchConfig, err := s.getBranchConfigWithDefaults(ctx, booking.BranchID)
	if err != nil {
		return nil, fmt.Errorf("failed to get branch configuration: %w", err)
	}

	if !booking.CanBeRescheduled(branchConfig.RescheduleWindowHours) {
		return nil, fmt.Errorf("booking cannot be rescheduled within %d hours of appointment", branchConfig.RescheduleWindowHours)
	}

	// Store old values for history
	oldServices, _ := json.Marshal(booking.Services)

	// Delete existing booking services
	if err := s.repo.DeleteBookingServices(ctx, request.BookingID); err != nil {
		return nil, fmt.Errorf("failed to delete old booking services: %w", err)
	}

	// Validate and create new services (similar to InitiateBooking)
	var newBookingServices []model.BookingService
	var totalAmount float64

	for _, serviceItem := range request.Services {
		// Validate service and stylist (same logic as InitiateBooking)
		serviceInfo, err := s.externalService.GetService(ctx, booking.SalonID, serviceItem.ServiceID)
		if err != nil {
			return nil, fmt.Errorf("invalid service %s: %w", serviceItem.ServiceID, err)
		}

		stylist, err := s.externalService.GetStylist(ctx, booking.SalonID, serviceItem.StylistID)
		if err != nil {
			return nil, fmt.Errorf("invalid stylist %s: %w", serviceItem.StylistID, err)
		}

		if stylist.BranchID != booking.BranchID {
			return nil, fmt.Errorf("stylist %s does not belong to branch %s", serviceItem.StylistID, booking.BranchID)
		}

		endTime := serviceItem.StartTime.Add(time.Duration(serviceInfo.Duration) * time.Minute)
		
		available, err := s.checkAvailabilityWithBuffer(ctx, serviceItem.StylistID, serviceItem.StartTime, endTime, branchConfig.BufferTimeMinutes)
		if err != nil {
			return nil, fmt.Errorf("failed to check availability: %w", err)
		}
		if !available {
			return nil, fmt.Errorf("stylist %s is not available at %s", serviceItem.StylistID, serviceItem.StartTime.Format("2006-01-02 15:04"))
		}

		bookingService := model.BookingService{
			ID:        uuid.New(),
			BookingID: request.BookingID,
			ServiceID: serviceItem.ServiceID,
			StylistID: serviceItem.StylistID,
			StartTime: serviceItem.StartTime,
			EndTime:   endTime,
			Price:     serviceInfo.Price,
		}

		newBookingServices = append(newBookingServices, bookingService)
		totalAmount += serviceInfo.Price
	}

	// Recalculate totals
	gst := totalAmount * (branchConfig.GSTPercentage / 100)
	finalTotal := totalAmount + branchConfig.BookingFeeAmount + gst

	// Update booking
	booking.Status = model.BookingStatusRescheduled
	booking.TotalAmount = finalTotal
	booking.GST = gst

	if err := s.repo.Update(ctx, booking); err != nil {
		return nil, fmt.Errorf("failed to update booking: %w", err)
	}

	// Create new booking services
	for i := range newBookingServices {
		if err := s.repo.CreateBookingService(ctx, &newBookingServices[i]); err != nil {
			return nil, fmt.Errorf("failed to create new booking service: %w", err)
		}
	}

	// Create history entry
	newServices, _ := json.Marshal(newBookingServices)
	history := &model.BookingHistory{
		ID:        uuid.New(),
		BookingID: request.BookingID,
		Action:    model.BookingActionRescheduled,
		OldValues: stringPtr(string(oldServices)),
		NewValues: stringPtr(string(newServices)),
		UserID:    &request.UserID,
		Reason:    &request.Reason,
	}
	if err := s.repo.CreateHistory(ctx, history); err != nil {
		log.Warn().Err(err).Msg("Failed to create booking history")
	}

	// Load updated services
	booking.Services = make([]model.BookingService, len(newBookingServices))
	for i, service := range newBookingServices {
		booking.Services[i] = service
	}

	log.Info().
		Str("booking_id", request.BookingID.String()).
		Str("user_id", request.UserID.String()).
		Msg("Booking rescheduled successfully")

	return booking, nil
}

// GetStylistAvailability generates available time slots for a stylist
func (s *bookingService) GetStylistAvailability(ctx context.Context, salonID, stylistID uuid.UUID, date time.Time) ([]*model.TimeSlot, error) {
	// Get stylist schedule from salon-service
	schedule, err := s.externalService.GetStylistSchedule(ctx, salonID, stylistID, date)
	if err != nil {
		return nil, fmt.Errorf("failed to get stylist schedule: %w", err)
	}

	// Get stylist's existing bookings for the date
	startOfDay := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	endOfDay := startOfDay.Add(24 * time.Hour)
	
	existingBookings, err := s.repo.GetStylistBookings(ctx, stylistID, startOfDay, endOfDay)
	if err != nil {
		return nil, fmt.Errorf("failed to get existing bookings: %w", err)
	}

	// Generate available slots (30-minute intervals)
	var availableSlots []*model.TimeSlot
	slotDuration := 30 * time.Minute

	for _, workingHour := range schedule.WorkingHours {
		current := workingHour.StartTime
		for current.Add(slotDuration).Before(workingHour.EndTime) || current.Add(slotDuration).Equal(workingHour.EndTime) {
			slotEnd := current.Add(slotDuration)
			
			// Check if slot conflicts with existing bookings
			available := true
			for _, booking := range existingBookings {
				if (current.Before(booking.EndTime) && slotEnd.After(booking.StartTime)) {
					available = false
					break
				}
			}

			// Check if slot conflicts with breaks
			if available {
				for _, breakPeriod := range schedule.Breaks {
					if (current.Before(breakPeriod.EndTime) && slotEnd.After(breakPeriod.StartTime)) {
						available = false
						break
					}
				}
			}

			// Only include future slots (not past)
			if available && current.After(time.Now()) {
				availableSlots = append(availableSlots, &model.TimeSlot{
					StartTime: current,
					EndTime:   slotEnd,
					Available: true,
				})
			}

			current = current.Add(slotDuration)
		}
	}

	return availableSlots, nil
}

// CalculateBookingSummary calculates pricing for a booking
func (s *bookingService) CalculateBookingSummary(ctx context.Context, request *BookingSummaryRequest) (*model.BookingSummary, error) {
	// Get branch configuration
	branchConfig, err := s.getBranchConfigWithDefaults(ctx, request.BranchID)
	if err != nil {
		return nil, fmt.Errorf("failed to get branch configuration: %w", err)
	}

	var subtotal float64

	// Calculate subtotal from services
	for _, serviceItem := range request.Services {
		serviceInfo, err := s.externalService.GetService(ctx, request.SalonID, serviceItem.ServiceID)
		if err != nil {
			return nil, fmt.Errorf("invalid service %s: %w", serviceItem.ServiceID, err)
		}
		subtotal += serviceInfo.Price
	}

	// Calculate GST and total
	gst := subtotal * (branchConfig.GSTPercentage / 100)
	total := subtotal + branchConfig.BookingFeeAmount + gst

	return &model.BookingSummary{
		Subtotal:   subtotal,
		BookingFee: branchConfig.BookingFeeAmount,
		GST:        gst,
		Total:      total,
	}, nil
}

// sendBookingConfirmationNotifications sends confirmation notifications for a booking
func (s *bookingService) sendBookingConfirmationNotifications(ctx context.Context, booking *model.Booking) {
	// Get user details
	user, err := s.externalService.ValidateUser(ctx, booking.UserID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get user details for notification")
		return
	}

	// Get salon details
	salon, err := s.externalService.GetSalon(ctx, booking.SalonID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get salon details for notification")
		return
	}

	// Get branch details
	branch, err := s.externalService.GetBranch(ctx, booking.SalonID, booking.BranchID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get branch details for notification")
		return
	}

	// Prepare booking event
	bookingEvent := &BookingEvent{
		Type:      "booking.confirmed",
		BookingID: booking.ID,
		UserID:    booking.UserID,
		SalonID:   booking.SalonID,
		BranchID:  booking.BranchID,
		Timestamp: time.Now(),
		Data: map[string]interface{}{
			"user_name":     user.Name,
			"user_email":    user.Email,
			"user_phone":    user.Phone,
			"salon_name":    salon.Name,
			"branch_name":   branch.Name,
			"total_amount":  booking.TotalAmount,
			"booking_time":  booking.CreatedAt.Format("2006-01-02 15:04"),
			"payment_id":    booking.PaymentID,
		},
	}

	// Send notifications
	if err := s.notificationClient.SendBookingConfirmationNotification(ctx, bookingEvent); err != nil {
		log.Error().Err(err).Msg("Failed to send booking confirmation notifications")
	}
}

// sendBookingCancellationNotifications sends cancellation notifications for a booking
func (s *bookingService) sendBookingCancellationNotifications(ctx context.Context, booking *model.Booking, reason string) {
	// Get user details
	user, err := s.externalService.ValidateUser(ctx, booking.UserID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get user details for notification")
		return
	}

	// Get salon details
	salon, err := s.externalService.GetSalon(ctx, booking.SalonID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get salon details for notification")
		return
	}

	// Get branch details
	branch, err := s.externalService.GetBranch(ctx, booking.SalonID, booking.BranchID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get branch details for notification")
		return
	}

	// Prepare booking event
	bookingEvent := &BookingEvent{
		Type:      "booking.cancelled",
		BookingID: booking.ID,
		UserID:    booking.UserID,
		SalonID:   booking.SalonID,
		BranchID:  booking.BranchID,
		Timestamp: time.Now(),
		Data: map[string]interface{}{
			"user_name":    user.Name,
			"user_email":   user.Email,
			"user_phone":   user.Phone,
			"salon_name":   salon.Name,
			"branch_name":  branch.Name,
			"booking_time": booking.CreatedAt.Format("2006-01-02 15:04"),
			"reason":       reason,
		},
	}

	// Send notifications
	if err := s.notificationClient.SendBookingCancellationNotification(ctx, bookingEvent); err != nil {
		log.Error().Err(err).Msg("Failed to send booking cancellation notifications")
	}
}

// Helper function to convert string to pointer
func stringPtr(s string) *string {
	return &s
}
