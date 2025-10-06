package repository

import (
	"context"
	"fmt"
	"time"

	"booking-service/internal/model"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// BookingRepository defines the interface for booking data operations
type BookingRepository interface {
	// Booking operations
	Create(ctx context.Context, booking *model.Booking) error
	GetByID(ctx context.Context, id uuid.UUID) (*model.Booking, error)
	GetByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*model.Booking, error)
	Update(ctx context.Context, booking *model.Booking) error
	UpdateStatus(ctx context.Context, id uuid.UUID, status model.BookingStatus) error
	
	// Booking service operations
	CreateBookingService(ctx context.Context, service *model.BookingService) error
	GetBookingServices(ctx context.Context, bookingID uuid.UUID) ([]*model.BookingService, error)
	UpdateBookingService(ctx context.Context, service *model.BookingService) error
	DeleteBookingServices(ctx context.Context, bookingID uuid.UUID) error
	
	// Availability operations
	GetStylistBookings(ctx context.Context, stylistID uuid.UUID, startTime, endTime time.Time) ([]*model.BookingService, error)
	CheckStylistAvailability(ctx context.Context, stylistID uuid.UUID, startTime, endTime time.Time) (bool, error)
	
	// History operations
	CreateHistory(ctx context.Context, history *model.BookingHistory) error
	GetBookingHistory(ctx context.Context, bookingID uuid.UUID) ([]*model.BookingHistory, error)
	
	// Configuration operations
	GetBranchConfiguration(ctx context.Context, branchID uuid.UUID) (*model.BranchConfiguration, error)
	CreateBranchConfiguration(ctx context.Context, config *model.BranchConfiguration) error
	UpdateBranchConfiguration(ctx context.Context, config *model.BranchConfiguration) error
}

type bookingRepository struct {
	db *pgxpool.Pool
}

// NewBookingRepository creates a new booking repository
func NewBookingRepository(db *pgxpool.Pool) BookingRepository {
	return &bookingRepository{db: db}
}

// Create creates a new booking
func (r *bookingRepository) Create(ctx context.Context, booking *model.Booking) error {
	query := `
		INSERT INTO bookings (id, user_id, salon_id, branch_id, status, total_amount, gst, booking_fee, payment_status, payment_id, notes)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		RETURNING created_at, updated_at
	`
	
	err := r.db.QueryRow(ctx, query,
		booking.ID, booking.UserID, booking.SalonID, booking.BranchID,
		booking.Status, booking.TotalAmount, booking.GST, booking.BookingFee,
		booking.PaymentStatus, booking.PaymentID, booking.Notes,
	).Scan(&booking.CreatedAt, &booking.UpdatedAt)
	
	if err != nil {
		return fmt.Errorf("failed to create booking: %w", err)
	}
	
	return nil
}

// GetByID retrieves a booking by ID with its services and history
func (r *bookingRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.Booking, error) {
	query := `
		SELECT id, user_id, salon_id, branch_id, status, total_amount, gst, booking_fee,
		       payment_status, payment_id, notes, created_at, updated_at
		FROM bookings
		WHERE id = $1
	`
	
	booking := &model.Booking{}
	err := r.db.QueryRow(ctx, query, id).Scan(
		&booking.ID, &booking.UserID, &booking.SalonID, &booking.BranchID,
		&booking.Status, &booking.TotalAmount, &booking.GST, &booking.BookingFee,
		&booking.PaymentStatus, &booking.PaymentID, &booking.Notes,
		&booking.CreatedAt, &booking.UpdatedAt,
	)
	
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("booking not found")
		}
		return nil, fmt.Errorf("failed to get booking: %w", err)
	}
	
	// Load services
	services, err := r.GetBookingServices(ctx, booking.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to load booking services: %w", err)
	}
	booking.Services = make([]model.BookingService, len(services))
	for i, service := range services {
		booking.Services[i] = *service
	}
	
	// Load history
	history, err := r.GetBookingHistory(ctx, booking.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to load booking history: %w", err)
	}
	booking.History = make([]model.BookingHistory, len(history))
	for i, h := range history {
		booking.History[i] = *h
	}
	
	return booking, nil
}

// GetByUserID retrieves bookings for a specific user
func (r *bookingRepository) GetByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*model.Booking, error) {
	query := `
		SELECT id, user_id, salon_id, branch_id, status, total_amount, gst, booking_fee,
		       payment_status, payment_id, notes, created_at, updated_at
		FROM bookings
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`
	
	rows, err := r.db.Query(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get user bookings: %w", err)
	}
	defer rows.Close()
	
	var bookings []*model.Booking
	for rows.Next() {
		booking := &model.Booking{}
		err := rows.Scan(
			&booking.ID, &booking.UserID, &booking.SalonID, &booking.BranchID,
			&booking.Status, &booking.TotalAmount, &booking.GST, &booking.BookingFee,
			&booking.PaymentStatus, &booking.PaymentID, &booking.Notes,
			&booking.CreatedAt, &booking.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan booking: %w", err)
		}
		
		// Load services for each booking
		services, err := r.GetBookingServices(ctx, booking.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to load booking services: %w", err)
		}
		booking.Services = make([]model.BookingService, len(services))
		for i, service := range services {
			booking.Services[i] = *service
		}
		
		bookings = append(bookings, booking)
	}
	
	return bookings, rows.Err()
}

// Update updates a booking
func (r *bookingRepository) Update(ctx context.Context, booking *model.Booking) error {
	query := `
		UPDATE bookings
		SET status = $2, total_amount = $3, gst = $4, booking_fee = $5,
		    payment_status = $6, payment_id = $7, notes = $8, updated_at = NOW()
		WHERE id = $1
	`
	
	result, err := r.db.Exec(ctx, query,
		booking.ID, booking.Status, booking.TotalAmount, booking.GST,
		booking.BookingFee, booking.PaymentStatus, booking.PaymentID, booking.Notes,
	)
	
	if err != nil {
		return fmt.Errorf("failed to update booking: %w", err)
	}
	
	if result.RowsAffected() == 0 {
		return fmt.Errorf("booking not found")
	}
	
	return nil
}

// UpdateStatus updates only the booking status
func (r *bookingRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status model.BookingStatus) error {
	query := `UPDATE bookings SET status = $2, updated_at = NOW() WHERE id = $1`
	
	result, err := r.db.Exec(ctx, query, id, status)
	if err != nil {
		return fmt.Errorf("failed to update booking status: %w", err)
	}
	
	if result.RowsAffected() == 0 {
		return fmt.Errorf("booking not found")
	}
	
	return nil
}

// CreateBookingService creates a new booking service
func (r *bookingRepository) CreateBookingService(ctx context.Context, service *model.BookingService) error {
	query := `
		INSERT INTO booking_services (id, booking_id, service_id, stylist_id, start_time, end_time, price)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING created_at, updated_at
	`
	
	err := r.db.QueryRow(ctx, query,
		service.ID, service.BookingID, service.ServiceID, service.StylistID,
		service.StartTime, service.EndTime, service.Price,
	).Scan(&service.CreatedAt, &service.UpdatedAt)
	
	if err != nil {
		return fmt.Errorf("failed to create booking service: %w", err)
	}
	
	return nil
}

// GetBookingServices retrieves all services for a booking
func (r *bookingRepository) GetBookingServices(ctx context.Context, bookingID uuid.UUID) ([]*model.BookingService, error) {
	query := `
		SELECT id, booking_id, service_id, stylist_id, start_time, end_time, price, created_at, updated_at
		FROM booking_services
		WHERE booking_id = $1
		ORDER BY start_time
	`
	
	rows, err := r.db.Query(ctx, query, bookingID)
	if err != nil {
		return nil, fmt.Errorf("failed to get booking services: %w", err)
	}
	defer rows.Close()
	
	var services []*model.BookingService
	for rows.Next() {
		service := &model.BookingService{}
		err := rows.Scan(
			&service.ID, &service.BookingID, &service.ServiceID, &service.StylistID,
			&service.StartTime, &service.EndTime, &service.Price,
			&service.CreatedAt, &service.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan booking service: %w", err)
		}
		services = append(services, service)
	}
	
	return services, rows.Err()
}

// UpdateBookingService updates a booking service
func (r *bookingRepository) UpdateBookingService(ctx context.Context, service *model.BookingService) error {
	query := `
		UPDATE booking_services
		SET service_id = $2, stylist_id = $3, start_time = $4, end_time = $5, price = $6, updated_at = NOW()
		WHERE id = $1
	`
	
	result, err := r.db.Exec(ctx, query,
		service.ID, service.ServiceID, service.StylistID,
		service.StartTime, service.EndTime, service.Price,
	)
	
	if err != nil {
		return fmt.Errorf("failed to update booking service: %w", err)
	}
	
	if result.RowsAffected() == 0 {
		return fmt.Errorf("booking service not found")
	}
	
	return nil
}

// DeleteBookingServices deletes all services for a booking
func (r *bookingRepository) DeleteBookingServices(ctx context.Context, bookingID uuid.UUID) error {
	query := `DELETE FROM booking_services WHERE booking_id = $1`
	
	_, err := r.db.Exec(ctx, query, bookingID)
	if err != nil {
		return fmt.Errorf("failed to delete booking services: %w", err)
	}
	
	return nil
}

// GetStylistBookings retrieves all bookings for a stylist in a time range
func (r *bookingRepository) GetStylistBookings(ctx context.Context, stylistID uuid.UUID, startTime, endTime time.Time) ([]*model.BookingService, error) {
	query := `
		SELECT bs.id, bs.booking_id, bs.service_id, bs.stylist_id, bs.start_time, bs.end_time, bs.price, bs.created_at, bs.updated_at
		FROM booking_services bs
		JOIN bookings b ON bs.booking_id = b.id
		WHERE bs.stylist_id = $1
		  AND bs.start_time < $3
		  AND bs.end_time > $2
		  AND b.status IN ('confirmed', 'rescheduled')
		ORDER BY bs.start_time
	`
	
	rows, err := r.db.Query(ctx, query, stylistID, startTime, endTime)
	if err != nil {
		return nil, fmt.Errorf("failed to get stylist bookings: %w", err)
	}
	defer rows.Close()
	
	var bookings []*model.BookingService
	for rows.Next() {
		booking := &model.BookingService{}
		err := rows.Scan(
			&booking.ID, &booking.BookingID, &booking.ServiceID, &booking.StylistID,
			&booking.StartTime, &booking.EndTime, &booking.Price,
			&booking.CreatedAt, &booking.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan stylist booking: %w", err)
		}
		bookings = append(bookings, booking)
	}
	
	return bookings, rows.Err()
}

// CheckStylistAvailability checks if a stylist is available for a time slot
func (r *bookingRepository) CheckStylistAvailability(ctx context.Context, stylistID uuid.UUID, startTime, endTime time.Time) (bool, error) {
	query := `
		SELECT COUNT(*)
		FROM booking_services bs
		JOIN bookings b ON bs.booking_id = b.id
		WHERE bs.stylist_id = $1
		  AND bs.start_time < $3
		  AND bs.end_time > $2
		  AND b.status IN ('confirmed', 'rescheduled')
	`
	
	var count int
	err := r.db.QueryRow(ctx, query, stylistID, startTime, endTime).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("failed to check stylist availability: %w", err)
	}
	
	return count == 0, nil
}

// CreateHistory creates a new booking history entry
func (r *bookingRepository) CreateHistory(ctx context.Context, history *model.BookingHistory) error {
	query := `
		INSERT INTO booking_history (id, booking_id, action, old_values, new_values, user_id, reason)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING timestamp
	`
	
	err := r.db.QueryRow(ctx, query,
		history.ID, history.BookingID, history.Action,
		history.OldValues, history.NewValues, history.UserID, history.Reason,
	).Scan(&history.Timestamp)
	
	if err != nil {
		return fmt.Errorf("failed to create booking history: %w", err)
	}
	
	return nil
}

// GetBookingHistory retrieves the history for a booking
func (r *bookingRepository) GetBookingHistory(ctx context.Context, bookingID uuid.UUID) ([]*model.BookingHistory, error) {
	query := `
		SELECT id, booking_id, action, old_values, new_values, user_id, reason, timestamp
		FROM booking_history
		WHERE booking_id = $1
		ORDER BY timestamp DESC
	`
	
	rows, err := r.db.Query(ctx, query, bookingID)
	if err != nil {
		return nil, fmt.Errorf("failed to get booking history: %w", err)
	}
	defer rows.Close()
	
	var history []*model.BookingHistory
	for rows.Next() {
		h := &model.BookingHistory{}
		err := rows.Scan(
			&h.ID, &h.BookingID, &h.Action,
			&h.OldValues, &h.NewValues, &h.UserID, &h.Reason, &h.Timestamp,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan booking history: %w", err)
		}
		history = append(history, h)
	}
	
	return history, rows.Err()
}

// GetBranchConfiguration retrieves configuration for a branch
func (r *bookingRepository) GetBranchConfiguration(ctx context.Context, branchID uuid.UUID) (*model.BranchConfiguration, error) {
	query := `
		SELECT branch_id, buffer_time_minutes, cancellation_cutoff_hours, reschedule_window_hours,
		       max_advance_booking_days, booking_fee_amount, gst_percentage, created_at, updated_at
		FROM branch_configurations
		WHERE branch_id = $1
	`
	
	config := &model.BranchConfiguration{}
	err := r.db.QueryRow(ctx, query, branchID).Scan(
		&config.BranchID, &config.BufferTimeMinutes, &config.CancellationCutoffHours,
		&config.RescheduleWindowHours, &config.MaxAdvanceBookingDays,
		&config.BookingFeeAmount, &config.GSTPercentage,
		&config.CreatedAt, &config.UpdatedAt,
	)
	
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("branch configuration not found")
		}
		return nil, fmt.Errorf("failed to get branch configuration: %w", err)
	}
	
	return config, nil
}

// CreateBranchConfiguration creates a new branch configuration
func (r *bookingRepository) CreateBranchConfiguration(ctx context.Context, config *model.BranchConfiguration) error {
	query := `
		INSERT INTO branch_configurations (branch_id, buffer_time_minutes, cancellation_cutoff_hours,
		                                 reschedule_window_hours, max_advance_booking_days,
		                                 booking_fee_amount, gst_percentage)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING created_at, updated_at
	`
	
	err := r.db.QueryRow(ctx, query,
		config.BranchID, config.BufferTimeMinutes, config.CancellationCutoffHours,
		config.RescheduleWindowHours, config.MaxAdvanceBookingDays,
		config.BookingFeeAmount, config.GSTPercentage,
	).Scan(&config.CreatedAt, &config.UpdatedAt)
	
	if err != nil {
		return fmt.Errorf("failed to create branch configuration: %w", err)
	}
	
	return nil
}

// UpdateBranchConfiguration updates a branch configuration
func (r *bookingRepository) UpdateBranchConfiguration(ctx context.Context, config *model.BranchConfiguration) error {
	query := `
		UPDATE branch_configurations
		SET buffer_time_minutes = $2, cancellation_cutoff_hours = $3, reschedule_window_hours = $4,
		    max_advance_booking_days = $5, booking_fee_amount = $6, gst_percentage = $7, updated_at = NOW()
		WHERE branch_id = $1
	`
	
	result, err := r.db.Exec(ctx, query,
		config.BranchID, config.BufferTimeMinutes, config.CancellationCutoffHours,
		config.RescheduleWindowHours, config.MaxAdvanceBookingDays,
		config.BookingFeeAmount, config.GSTPercentage,
	)
	
	if err != nil {
		return fmt.Errorf("failed to update branch configuration: %w", err)
	}
	
	if result.RowsAffected() == 0 {
		return fmt.Errorf("branch configuration not found")
	}
	
	return nil
}
