package model

import (
	"time"

	"github.com/google/uuid"
)

// Notification status constants
const (
	NotificationStatusPending   = "pending"
	NotificationStatusSent      = "sent"
	NotificationStatusFailed    = "failed"
	NotificationStatusRetrying  = "retrying"
	NotificationStatusCanceled  = "canceled"
)

// Notification channel constants
const (
	ChannelEmail = "email"
	ChannelSMS   = "sms"
	ChannelPush  = "push"
)

// Notification priority constants
const (
	PriorityLow    = "low"
	PriorityNormal = "normal"
	PriorityHigh   = "high"
	PriorityUrgent = "urgent"
)

// Notification represents a notification record
type Notification struct {
	ID          uuid.UUID  `json:"id" db:"id"`
	EventType   string     `json:"event_type" db:"event_type"`
	Channel     string     `json:"channel" db:"channel"`
	Recipient   string     `json:"recipient" db:"recipient"`
	Subject     *string    `json:"subject,omitempty" db:"subject"`
	Content     string     `json:"content" db:"content"`
	TemplateID  *uuid.UUID `json:"template_id,omitempty" db:"template_id"`
	Status      string     `json:"status" db:"status"`
	Priority    string     `json:"priority" db:"priority"`
	Metadata    *string    `json:"metadata,omitempty" db:"metadata"`
	ProviderID  *string    `json:"provider_id,omitempty" db:"provider_id"`
	ErrorMsg    *string    `json:"error_message,omitempty" db:"error_message"`
	RetryCount  int        `json:"retry_count" db:"retry_count"`
	ScheduledAt *time.Time `json:"scheduled_at,omitempty" db:"scheduled_at"`
	SentAt      *time.Time `json:"sent_at,omitempty" db:"sent_at"`
	ExpiresAt   *time.Time `json:"expires_at,omitempty" db:"expires_at"`
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at" db:"updated_at"`
}

// NotificationTemplate represents a notification template
type NotificationTemplate struct {
	ID          uuid.UUID `json:"id" db:"id"`
	Name        string    `json:"name" db:"name"`
	EventType   string    `json:"event_type" db:"event_type"`
	Channel     string    `json:"channel" db:"channel"`
	Subject     *string   `json:"subject,omitempty" db:"subject"`
	Content     string    `json:"content" db:"content"`
	Variables   *string   `json:"variables,omitempty" db:"variables"`
	IsActive    bool      `json:"is_active" db:"is_active"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

// NotificationAttempt represents a notification delivery attempt
type NotificationAttempt struct {
	ID             uuid.UUID `json:"id" db:"id"`
	NotificationID uuid.UUID `json:"notification_id" db:"notification_id"`
	AttemptNumber  int       `json:"attempt_number" db:"attempt_number"`
	Status         string    `json:"status" db:"status"`
	Provider       string    `json:"provider" db:"provider"`
	ProviderID     *string   `json:"provider_id,omitempty" db:"provider_id"`
	ErrorMessage   *string   `json:"error_message,omitempty" db:"error_message"`
	ResponseData   *string   `json:"response_data,omitempty" db:"response_data"`
	AttemptedAt    time.Time `json:"attempted_at" db:"attempted_at"`
	CreatedAt      time.Time `json:"created_at" db:"created_at"`
}

// BookingEvent represents a booking-related event
type BookingEvent struct {
	EventID     uuid.UUID              `json:"event_id"`
	EventType   string                 `json:"event_type"`
	BookingID   uuid.UUID              `json:"booking_id"`
	UserID      uuid.UUID              `json:"user_id"`
	SalonID     uuid.UUID              `json:"salon_id"`
	BranchID    uuid.UUID              `json:"branch_id"`
	Timestamp   time.Time              `json:"timestamp"`
	Data        map[string]interface{} `json:"data"`
}

// NotificationRequest represents a request to send a notification
type NotificationRequest struct {
	EventType   string                 `json:"event_type" validate:"required"`
	Channel     string                 `json:"channel" validate:"required,oneof=email sms push"`
	Recipient   string                 `json:"recipient" validate:"required"`
	Subject     *string                `json:"subject,omitempty"`
	Content     string                 `json:"content" validate:"required"`
	TemplateID  *uuid.UUID             `json:"template_id,omitempty"`
	Priority    string                 `json:"priority" validate:"oneof=low normal high urgent"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	ScheduledAt *time.Time             `json:"scheduled_at,omitempty"`
}

// NotificationResponse represents a notification response
type NotificationResponse struct {
	Notification *Notification `json:"notification"`
	Message      string        `json:"message"`
}

// NotificationListResponse represents a list of notifications
type NotificationListResponse struct {
	Notifications []*Notification `json:"notifications"`
	TotalCount    int             `json:"total_count"`
	Page          int             `json:"page"`
	PageSize      int             `json:"page_size"`
}

// TemplateRequest represents a request to create/update a template
type TemplateRequest struct {
	Name      string                 `json:"name" validate:"required"`
	EventType string                 `json:"event_type" validate:"required"`
	Channel   string                 `json:"channel" validate:"required,oneof=email sms push"`
	Subject   *string                `json:"subject,omitempty"`
	Content   string                 `json:"content" validate:"required"`
	Variables map[string]interface{} `json:"variables,omitempty"`
	IsActive  bool                   `json:"is_active"`
}

// TemplateResponse represents a template response
type TemplateResponse struct {
	Template *NotificationTemplate `json:"template"`
	Message  string                `json:"message"`
}

// HealthResponse represents a health check response
type HealthResponse struct {
	Status    string `json:"status"`
	Timestamp string `json:"timestamp"`
	Version   string `json:"version"`
	Database  string `json:"database"`
	Providers string `json:"providers"`
}

// SendNotificationRequest represents a request to send a notification
type SendNotificationRequest struct {
	UserID    *uuid.UUID             `json:"user_id,omitempty"`
	Type      string                 `json:"type" validate:"required,oneof=email sms push"`
	Recipient string                 `json:"recipient" validate:"required"`
	Subject   string                 `json:"subject,omitempty"`
	Content   string                 `json:"content" validate:"required"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// Event represents an event from the message broker
type Event struct {
	Type string                 `json:"type"`
	Data map[string]interface{} `json:"data"`
}

// IsExpired checks if the notification has expired
func (n *Notification) IsExpired() bool {
	return n.ExpiresAt != nil && time.Now().After(*n.ExpiresAt)
}

// CanRetry checks if the notification can be retried
func (n *Notification) CanRetry(maxRetries int) bool {
	return n.Status == NotificationStatusFailed && n.RetryCount < maxRetries
}

// IsTerminalStatus checks if the notification is in a terminal status
func (n *Notification) IsTerminalStatus() bool {
	return n.Status == NotificationStatusSent ||
		n.Status == NotificationStatusCanceled
}
