package api

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"booking-service/internal/db"
	"booking-service/internal/service"

	"github.com/EricsAntony/salon/salon-shared/auth"
	"github.com/EricsAntony/salon/salon-shared/errors"
	"github.com/EricsAntony/salon/salon-shared/utils"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

// Handlers contains all HTTP handlers for the booking service
type Handlers struct {
	bookingService service.BookingService
}

// NewHandlers creates a new handlers instance
func NewHandlers(bookingService service.BookingService) *Handlers {
	return &Handlers{
		bookingService: bookingService,
	}
}

// Health returns the health status of the service
func (h *Handlers) Health(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"status":    "healthy",
		"service":   "booking-service",
		"timestamp": time.Now().UTC(),
	}
	utils.WriteJSON(w, http.StatusOK, response)
}

// Ready returns the readiness status of the service
func (h *Handlers) Ready(w http.ResponseWriter, r *http.Request) {
	// Check database connectivity
	if err := db.HealthCheck(nil); err != nil {
		response := map[string]interface{}{
			"status": "not ready",
			"error":  "database connection failed",
		}
		utils.WriteJSON(w, http.StatusServiceUnavailable, response)
		return
	}

	response := map[string]interface{}{
		"status":    "ready",
		"service":   "booking-service",
		"timestamp": time.Now().UTC(),
	}
	utils.WriteJSON(w, http.StatusOK, response)
}

// InitiateBooking handles POST /bookings/initiate
func (h *Handlers) InitiateBooking(w http.ResponseWriter, r *http.Request) {
	var request service.InitiateBookingRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		errors.WriteAPIError(w, errors.NewValidationError("request_body", "invalid JSON format"))
		return
	}

	// Get user ID from context (set by middleware)
	userID, ok := r.Context().Value(auth.CtxUserID).(string)
	if !ok {
		errors.WriteAPIError(w, errors.NewAuthError("authentication", "user not authenticated"))
		return
	}

	// Parse user ID
	parsedUserID, err := uuid.Parse(userID)
	if err != nil {
		errors.WriteAPIError(w, errors.NewValidationError("user_id", "invalid user ID format"))
		return
	}
	request.UserID = parsedUserID

	// Validate request
	if err := h.validateInitiateBookingRequest(&request); err != nil {
		errors.WriteAPIError(w, err)
		return
	}

	// Create booking
	booking, err := h.bookingService.InitiateBooking(r.Context(), &request)
	if err != nil {
		log.Error().Err(err).Msg("Failed to initiate booking")
		errors.WriteAPIError(w, &errors.ConflictError{Resource: "booking", Detail: err.Error()})
		return
	}

	utils.WriteJSON(w, http.StatusCreated, booking)
}

// GetStylistAvailability handles GET /stylists/{stylistId}/availability
func (h *Handlers) GetStylistAvailability(w http.ResponseWriter, r *http.Request) {
	stylistIDStr := chi.URLParam(r, "stylistId")
	stylistID, err := uuid.Parse(stylistIDStr)
	if err != nil {
		errors.WriteAPIError(w, errors.NewValidationError("stylist_id", "invalid stylist ID format"))
		return
	}

	salonIDStr := r.URL.Query().Get("salon_id")
	if salonIDStr == "" {
		errors.WriteAPIError(w, errors.NewValidationError("salon_id", "salon_id parameter is required"))
		return
	}

	salonID, err := uuid.Parse(salonIDStr)
	if err != nil {
		errors.WriteAPIError(w, errors.NewValidationError("salon_id", "invalid salon ID format"))
		return
	}

	dateStr := r.URL.Query().Get("date")
	if dateStr == "" {
		errors.WriteAPIError(w, errors.NewValidationError("date", "date parameter is required"))
		return
	}

	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		errors.WriteAPIError(w, errors.NewValidationError("date", "invalid date format, use YYYY-MM-DD"))
		return
	}

	slots, err := h.bookingService.GetStylistAvailability(r.Context(), salonID, stylistID, date)
	if err != nil {
		log.Error().Err(err).Str("stylist_id", stylistID.String()).Msg("Failed to get stylist availability")
		errors.WriteAPIError(w, &errors.ConflictError{Resource: "availability", Detail: err.Error()})
		return
	}

	utils.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"salon_id":   salonID,
		"stylist_id": stylistID,
		"date":       date.Format("2006-01-02"),
		"slots":      slots,
	})
}

// CalculateBookingSummary handles POST /bookings/summary
func (h *Handlers) CalculateBookingSummary(w http.ResponseWriter, r *http.Request) {
	var request service.BookingSummaryRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		errors.WriteAPIError(w, errors.NewValidationError("request_body", "invalid JSON format"))
		return
	}

	// Validate request
	if err := h.validateBookingSummaryRequest(&request); err != nil {
		errors.WriteAPIError(w, err)
		return
	}

	summary, err := h.bookingService.CalculateBookingSummary(r.Context(), &request)
	if err != nil {
		log.Error().Err(err).Msg("Failed to calculate booking summary")
		errors.WriteAPIError(w, &errors.ConflictError{Resource: "booking_summary", Detail: err.Error()})
		return
	}

	utils.WriteJSON(w, http.StatusOK, summary)
}

// ConfirmBooking handles POST /bookings/confirm
func (h *Handlers) ConfirmBooking(w http.ResponseWriter, r *http.Request) {
	var request struct {
		BookingID uuid.UUID `json:"booking_id"`
		PaymentID string    `json:"payment_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		errors.WriteAPIError(w, errors.NewValidationError("request_body", "invalid JSON format"))
		return
	}

	if request.BookingID == uuid.Nil {
		errors.WriteAPIError(w, errors.NewValidationError("booking_id", "booking_id is required"))
		return
	}

	if request.PaymentID == "" {
		errors.WriteAPIError(w, errors.NewValidationError("payment_id", "payment_id is required"))
		return
	}

	booking, err := h.bookingService.ConfirmBooking(r.Context(), request.BookingID, request.PaymentID)
	if err != nil {
		log.Error().Err(err).Str("booking_id", request.BookingID.String()).Msg("Failed to confirm booking")
		errors.WriteAPIError(w, &errors.ConflictError{Resource: "booking", Detail: err.Error()})
		return
	}

	utils.WriteJSON(w, http.StatusOK, booking)
}

// GetBooking handles GET /bookings/{bookingId}
func (h *Handlers) GetBooking(w http.ResponseWriter, r *http.Request) {
	bookingIDStr := chi.URLParam(r, "bookingId")
	bookingID, err := uuid.Parse(bookingIDStr)
	if err != nil {
		errors.WriteAPIError(w, errors.NewValidationError("booking_id", "invalid booking ID format"))
		return
	}

	booking, err := h.bookingService.GetBooking(r.Context(), bookingID)
	if err != nil {
		log.Error().Err(err).Str("booking_id", bookingID.String()).Msg("Failed to get booking")
		errors.WriteAPIError(w, &errors.NotFoundError{Resource: "booking", ID: bookingID.String()})
		return
	}

	utils.WriteJSON(w, http.StatusOK, booking)
}

// GetUserBookings handles GET /bookings/user/{userId}
func (h *Handlers) GetUserBookings(w http.ResponseWriter, r *http.Request) {
	userIDStr := chi.URLParam(r, "userId")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		errors.WriteAPIError(w, errors.NewValidationError("user_id", "invalid user ID format"))
		return
	}

	// Get authenticated user ID from context
	authUserID, ok := r.Context().Value(auth.CtxUserID).(string)
	if !ok {
		errors.WriteAPIError(w, errors.NewAuthError("authentication", "user not authenticated"))
		return
	}

	// Ensure user can only access their own bookings
	if userID.String() != authUserID {
		errors.WriteAPIError(w, errors.NewAuthError("authorization", "unauthorized access"))
		return
	}

	// Parse pagination parameters
	limit := 20 // default
	offset := 0 // default

	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 && parsedLimit <= 100 {
			limit = parsedLimit
		}
	}

	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		if parsedOffset, err := strconv.Atoi(offsetStr); err == nil && parsedOffset >= 0 {
			offset = parsedOffset
		}
	}

	bookings, err := h.bookingService.GetUserBookings(r.Context(), userID, limit, offset)
	if err != nil {
		log.Error().Err(err).Str("user_id", userID.String()).Msg("Failed to get user bookings")
		errors.WriteAPIError(w, &errors.ConflictError{Resource: "bookings", Detail: err.Error()})
		return
	}

	utils.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"bookings": bookings,
		"limit":    limit,
		"offset":   offset,
	})
}

// CancelBooking handles PATCH /bookings/{bookingId}/cancel
func (h *Handlers) CancelBooking(w http.ResponseWriter, r *http.Request) {
	bookingIDStr := chi.URLParam(r, "bookingId")
	bookingID, err := uuid.Parse(bookingIDStr)
	if err != nil {
		errors.WriteAPIError(w, errors.NewValidationError("booking_id", "invalid booking ID format"))
		return
	}

	var request struct {
		Reason string `json:"reason"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		errors.WriteAPIError(w, errors.NewValidationError("request_body", "invalid JSON format"))
		return
	}

	// Get user ID from context
	userIDStr, ok := r.Context().Value(auth.CtxUserID).(string)
	if !ok {
		errors.WriteAPIError(w, errors.NewAuthError("authentication", "user not authenticated"))
		return
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		errors.WriteAPIError(w, errors.NewValidationError("user_id", "invalid user ID format"))
		return
	}

	if err := h.bookingService.CancelBooking(r.Context(), bookingID, userID, request.Reason); err != nil {
		log.Error().Err(err).Str("booking_id", bookingID.String()).Msg("Failed to cancel booking")
		errors.WriteAPIError(w, &errors.ConflictError{Resource: "booking", Detail: err.Error()})
		return
	}

	utils.WriteJSON(w, http.StatusOK, map[string]string{
		"message": "Booking canceled successfully",
	})
}

// RescheduleBooking handles PATCH /bookings/{bookingId}/reschedule
func (h *Handlers) RescheduleBooking(w http.ResponseWriter, r *http.Request) {
	bookingIDStr := chi.URLParam(r, "bookingId")
	bookingID, err := uuid.Parse(bookingIDStr)
	if err != nil {
		errors.WriteAPIError(w, errors.NewValidationError("booking_id", "invalid booking ID format"))
		return
	}

	var request service.RescheduleBookingRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		errors.WriteAPIError(w, errors.NewValidationError("request_body", "invalid JSON format"))
		return
	}

	// Get user ID from context
	userIDStr, ok := r.Context().Value(auth.CtxUserID).(string)
	if !ok {
		errors.WriteAPIError(w, errors.NewAuthError("authentication", "user not authenticated"))
		return
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		errors.WriteAPIError(w, errors.NewValidationError("user_id", "invalid user ID format"))
		return
	}

	request.BookingID = bookingID
	request.UserID = userID

	// Validate request
	if err := h.validateRescheduleBookingRequest(&request); err != nil {
		errors.WriteAPIError(w, err)
		return
	}

	booking, err := h.bookingService.RescheduleBooking(r.Context(), &request)
	if err != nil {
		log.Error().Err(err).Str("booking_id", bookingID.String()).Msg("Failed to reschedule booking")
		errors.WriteAPIError(w, &errors.ConflictError{Resource: "booking", Detail: err.Error()})
		return
	}

	utils.WriteJSON(w, http.StatusOK, booking)
}

// GetBranchConfig handles GET /branches/{branchId}/config
func (h *Handlers) GetBranchConfig(w http.ResponseWriter, r *http.Request) {
	branchIDStr := chi.URLParam(r, "branchId")
	branchID, err := uuid.Parse(branchIDStr)
	if err != nil {
		errors.WriteAPIError(w, errors.NewValidationError("branch_id", "invalid branch ID format"))
		return
	}

	config, err := h.bookingService.GetBranchConfiguration(r.Context(), branchID)
	if err != nil {
		log.Error().Err(err).Str("branch_id", branchID.String()).Msg("Failed to get branch configuration")
		errors.WriteAPIError(w, &errors.NotFoundError{Resource: "branch_configuration", ID: branchID.String()})
		return
	}

	utils.WriteJSON(w, http.StatusOK, config)
}

// Validation helper functions
func (h *Handlers) validateInitiateBookingRequest(request *service.InitiateBookingRequest) error {
	if request.UserID == uuid.Nil {
		return errors.NewValidationError("user_id", "user_id is required")
	}
	if request.SalonID == uuid.Nil {
		return errors.NewValidationError("salon_id", "salon_id is required")
	}
	if request.BranchID == uuid.Nil {
		return errors.NewValidationError("branch_id", "branch_id is required")
	}
	if len(request.Services) == 0 {
		return errors.NewValidationError("services", "at least one service is required")
	}

	for i, service := range request.Services {
		if service.ServiceID == uuid.Nil {
			return errors.NewValidationError("services", "service_id is required for service "+strconv.Itoa(i))
		}
		if service.StylistID == uuid.Nil {
			return errors.NewValidationError("services", "stylist_id is required for service "+strconv.Itoa(i))
		}
		if service.StartTime.IsZero() {
			return errors.NewValidationError("services", "start_time is required for service "+strconv.Itoa(i))
		}
		if service.StartTime.Before(time.Now()) {
			return errors.NewValidationError("services", "start_time cannot be in the past for service "+strconv.Itoa(i))
		}
	}

	return nil
}

func (h *Handlers) validateBookingSummaryRequest(request *service.BookingSummaryRequest) error {
	if request.SalonID == uuid.Nil {
		return errors.NewValidationError("salon_id", "salon_id is required")
	}
	if request.BranchID == uuid.Nil {
		return errors.NewValidationError("branch_id", "branch_id is required")
	}
	if len(request.Services) == 0 {
		return errors.NewValidationError("services", "at least one service is required")
	}

	for i, service := range request.Services {
		if service.ServiceID == uuid.Nil {
			return errors.NewValidationError("services", "service_id is required for service "+strconv.Itoa(i))
		}
		if service.StylistID == uuid.Nil {
			return errors.NewValidationError("services", "stylist_id is required for service "+strconv.Itoa(i))
		}
	}

	return nil
}

func (h *Handlers) validateRescheduleBookingRequest(request *service.RescheduleBookingRequest) error {
	if request.BookingID == uuid.Nil {
		return errors.NewValidationError("booking_id", "booking_id is required")
	}
	if request.UserID == uuid.Nil {
		return errors.NewValidationError("user_id", "user_id is required")
	}
	if len(request.Services) == 0 {
		return errors.NewValidationError("services", "at least one service is required")
	}

	for i, service := range request.Services {
		if service.ServiceID == uuid.Nil {
			return errors.NewValidationError("services", "service_id is required for service "+strconv.Itoa(i))
		}
		if service.StylistID == uuid.Nil {
			return errors.NewValidationError("services", "stylist_id is required for service "+strconv.Itoa(i))
		}
		if service.StartTime.IsZero() {
			return errors.NewValidationError("services", "start_time is required for service "+strconv.Itoa(i))
		}
		if service.StartTime.Before(time.Now()) {
			return errors.NewValidationError("services", "start_time cannot be in the past for service "+strconv.Itoa(i))
		}
	}

	return nil
}
