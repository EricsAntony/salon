package model

// BookingStatus represents the status of a booking
type BookingStatus string

const (
	BookingStatusInitiated  BookingStatus = "initiated"
	BookingStatusConfirmed  BookingStatus = "confirmed"
	BookingStatusRescheduled BookingStatus = "rescheduled"
	BookingStatusCanceled   BookingStatus = "canceled"
	BookingStatusCompleted  BookingStatus = "completed"
)

// PaymentStatus represents the payment status of a booking
type PaymentStatus string

const (
	PaymentStatusPending PaymentStatus = "pending"
	PaymentStatusPaid    PaymentStatus = "paid"
	PaymentStatusFailed  PaymentStatus = "failed"
	PaymentStatusRefunded PaymentStatus = "refunded"
)

// BookingAction represents actions performed on bookings for history tracking
type BookingAction string

const (
	BookingActionCreated     BookingAction = "created"
	BookingActionConfirmed   BookingAction = "confirmed"
	BookingActionRescheduled BookingAction = "rescheduled"
	BookingActionCanceled    BookingAction = "canceled"
	BookingActionCompleted   BookingAction = "completed"
)

// IsValid checks if the booking status is valid
func (bs BookingStatus) IsValid() bool {
	switch bs {
	case BookingStatusInitiated, BookingStatusConfirmed, BookingStatusRescheduled, BookingStatusCanceled, BookingStatusCompleted:
		return true
	default:
		return false
	}
}

// IsValid checks if the payment status is valid
func (ps PaymentStatus) IsValid() bool {
	switch ps {
	case PaymentStatusPending, PaymentStatusPaid, PaymentStatusFailed, PaymentStatusRefunded:
		return true
	default:
		return false
	}
}

// IsValid checks if the booking action is valid
func (ba BookingAction) IsValid() bool {
	switch ba {
	case BookingActionCreated, BookingActionConfirmed, BookingActionRescheduled, BookingActionCanceled, BookingActionCompleted:
		return true
	default:
		return false
	}
}
