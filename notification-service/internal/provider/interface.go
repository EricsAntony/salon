package provider

import (
	"context"

	"notification-service/internal/model"
)

// NotificationProvider defines the interface for notification providers
type NotificationProvider interface {
	// GetName returns the provider name
	GetName() string

	// GetChannel returns the notification channel this provider supports
	GetChannel() string

	// Send sends a notification
	Send(ctx context.Context, notification *model.Notification) (*SendResponse, error)

	// IsHealthy checks if the provider is healthy
	IsHealthy(ctx context.Context) error
}

// SendResponse represents a response from sending a notification
type SendResponse struct {
	ProviderID   string                 `json:"provider_id"`
	Status       string                 `json:"status"`
	Message      string                 `json:"message"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
	ErrorMessage string                 `json:"error_message,omitempty"`
}

// ProviderManager manages multiple notification providers
type ProviderManager interface {
	GetProvider(channel string) (NotificationProvider, error)
	GetAvailableChannels() []string
	HealthCheck(ctx context.Context) map[string]error
}

// Provider status constants
const (
	StatusSent   = "sent"
	StatusFailed = "failed"
)
