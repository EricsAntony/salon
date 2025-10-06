package provider

import (
	"context"
	"fmt"

	"notification-service/internal/config"
	"notification-service/internal/model"
)

// providerManager implements ProviderManager interface
type providerManager struct {
	providers map[string]NotificationProvider
	config    *config.Config
}

// NewProviderManager creates a new provider manager
func NewProviderManager(cfg *config.Config) ProviderManager {
	manager := &providerManager{
		providers: make(map[string]NotificationProvider),
		config:    cfg,
	}

	// Initialize email provider
	if cfg.EmailProvider != "" {
		emailProvider := NewEmailProvider(cfg)
		manager.providers[model.ChannelEmail] = emailProvider
	}

	// Initialize SMS provider
	if cfg.SMSProvider != "" {
		smsProvider := NewSMSProvider(cfg)
		manager.providers[model.ChannelSMS] = smsProvider
	}

	// Initialize push provider
	if cfg.PushProvider != "" {
		pushProvider := NewPushProvider(cfg)
		manager.providers[model.ChannelPush] = pushProvider
	}

	return manager
}

// GetProvider returns a notification provider for the specified channel
func (m *providerManager) GetProvider(channel string) (NotificationProvider, error) {
	provider, exists := m.providers[channel]
	if !exists {
		return nil, fmt.Errorf("notification provider for channel '%s' not found or not configured", channel)
	}
	return provider, nil
}

// GetAvailableChannels returns list of available notification channels
func (m *providerManager) GetAvailableChannels() []string {
	var channels []string
	for channel := range m.providers {
		channels = append(channels, channel)
	}
	return channels
}

// HealthCheck performs health checks on all providers
func (m *providerManager) HealthCheck(ctx context.Context) map[string]error {
	results := make(map[string]error)
	
	for channel, provider := range m.providers {
		if err := provider.IsHealthy(ctx); err != nil {
			results[channel] = err
		} else {
			results[channel] = nil // Healthy
		}
	}
	
	return results
}
