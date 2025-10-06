package model

import (
	"time"

	"github.com/google/uuid"
)

// Booking represents a booking in the system
type Booking struct {
	ID            uuid.UUID     `json:"id" db:"id"`
	UserID        uuid.UUID     `json:"user_id" db:"user_id"`
	SalonID       uuid.UUID     `json:"salon_id" db:"salon_id"`
	BranchID      uuid.UUID     `json:"branch_id" db:"branch_id"`
	Status        BookingStatus `json:"status" db:"status"`
	TotalAmount   float64       `json:"total_amount" db:"total_amount"`
	GST           float64       `json:"gst" db:"gst"`
	BookingFee    float64       `json:"booking_fee" db:"booking_fee"`
	PaymentStatus PaymentStatus `json:"payment_status" db:"payment_status"`
	PaymentID     *string       `json:"payment_id,omitempty" db:"payment_id"`
	Notes         *string       `json:"notes,omitempty" db:"notes"`
	CreatedAt     time.Time     `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time     `json:"updated_at" db:"updated_at"`
	
	// Related data (loaded via joins)
	Services []BookingService `json:"services,omitempty"`
	History  []BookingHistory `json:"history,omitempty"`
}

// BookingService represents a service within a booking
type BookingService struct {
	ID        uuid.UUID `json:"id" db:"id"`
	BookingID uuid.UUID `json:"booking_id" db:"booking_id"`
	ServiceID uuid.UUID `json:"service_id" db:"service_id"`
	StylistID uuid.UUID `json:"stylist_id" db:"stylist_id"`
	StartTime time.Time `json:"start_time" db:"start_time"`
	EndTime   time.Time `json:"end_time" db:"end_time"`
	Price     float64   `json:"price" db:"price"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
	
	// Related data (loaded via joins)
	ServiceName *string `json:"service_name,omitempty"`
	StylistName *string `json:"stylist_name,omitempty"`
}

// BookingHistory represents the history of changes to a booking
type BookingHistory struct {
	ID        uuid.UUID     `json:"id" db:"id"`
	BookingID uuid.UUID     `json:"booking_id" db:"booking_id"`
	Action    BookingAction `json:"action" db:"action"`
	OldValues *string       `json:"old_values,omitempty" db:"old_values"`
	NewValues *string       `json:"new_values,omitempty" db:"new_values"`
	UserID    *uuid.UUID    `json:"user_id,omitempty" db:"user_id"`
	Reason    *string       `json:"reason,omitempty" db:"reason"`
	Timestamp time.Time     `json:"timestamp" db:"timestamp"`
}

// BranchConfiguration represents configuration settings for a branch
type BranchConfiguration struct {
	BranchID                uuid.UUID `json:"branch_id" db:"branch_id"`
	BufferTimeMinutes       int       `json:"buffer_time_minutes" db:"buffer_time_minutes"`
	CancellationCutoffHours int       `json:"cancellation_cutoff_hours" db:"cancellation_cutoff_hours"`
	RescheduleWindowHours   int       `json:"reschedule_window_hours" db:"reschedule_window_hours"`
	MaxAdvanceBookingDays   int       `json:"max_advance_booking_days" db:"max_advance_booking_days"`
	BookingFeeAmount        float64   `json:"booking_fee_amount" db:"booking_fee_amount"`
	GSTPercentage           float64   `json:"gst_percentage" db:"gst_percentage"`
	CreatedAt               time.Time `json:"created_at" db:"created_at"`
	UpdatedAt               time.Time `json:"updated_at" db:"updated_at"`
}

// TimeSlot represents an available time slot for booking
type TimeSlot struct {
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
	Available bool      `json:"available"`
}

// BookingSummary represents a summary of booking costs
type BookingSummary struct {
	Subtotal   float64 `json:"subtotal"`
	BookingFee float64 `json:"booking_fee"`
	GST        float64 `json:"gst"`
	Total      float64 `json:"total"`
}

// GetDuration returns the total duration of the booking in minutes
func (b *Booking) GetDuration() int {
	if len(b.Services) == 0 {
		return 0
	}
	
	var totalMinutes int
	for _, service := range b.Services {
		duration := service.EndTime.Sub(service.StartTime)
		totalMinutes += int(duration.Minutes())
	}
	
	return totalMinutes
}

// GetEarliestStartTime returns the earliest start time among all services
func (b *Booking) GetEarliestStartTime() *time.Time {
	if len(b.Services) == 0 {
		return nil
	}
	
	earliest := b.Services[0].StartTime
	for _, service := range b.Services[1:] {
		if service.StartTime.Before(earliest) {
			earliest = service.StartTime
		}
	}
	
	return &earliest
}

// GetLatestEndTime returns the latest end time among all services
func (b *Booking) GetLatestEndTime() *time.Time {
	if len(b.Services) == 0 {
		return nil
	}
	
	latest := b.Services[0].EndTime
	for _, service := range b.Services[1:] {
		if service.EndTime.After(latest) {
			latest = service.EndTime
		}
	}
	
	return &latest
}

// CanBeCanceled checks if the booking can be canceled based on cutoff time
func (b *Booking) CanBeCanceled(cutoffHours int) bool {
	if b.Status == BookingStatusCanceled || b.Status == BookingStatusCompleted {
		return false
	}
	
	earliestStart := b.GetEarliestStartTime()
	if earliestStart == nil {
		return false
	}
	
	cutoffTime := time.Now().Add(time.Duration(cutoffHours) * time.Hour)
	return earliestStart.After(cutoffTime)
}

// CanBeRescheduled checks if the booking can be rescheduled based on window
func (b *Booking) CanBeRescheduled(windowHours int) bool {
	if b.Status == BookingStatusCanceled || b.Status == BookingStatusCompleted {
		return false
	}
	
	earliestStart := b.GetEarliestStartTime()
	if earliestStart == nil {
		return false
	}
	
	windowTime := time.Now().Add(time.Duration(windowHours) * time.Hour)
	return earliestStart.After(windowTime)
}
