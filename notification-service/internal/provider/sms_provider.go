package provider

import (
	"context"
	"fmt"

	"notification-service/internal/config"
	"notification-service/internal/model"

	"github.com/twilio/twilio-go"
	twilioApi "github.com/twilio/twilio-go/rest/api/v2010"
)

// SMSProvider implements NotificationProvider for SMS notifications
type SMSProvider struct {
	config   *config.Config
	provider string // "twilio"
	client   *twilio.RestClient
}

// NewSMSProvider creates a new SMS provider
func NewSMSProvider(cfg *config.Config) *SMSProvider {
	var client *twilio.RestClient
	
	if cfg.SMSProvider == "twilio" && cfg.TwilioAccountSID != "" && cfg.TwilioAuthToken != "" {
		client = twilio.NewRestClientWithParams(twilio.ClientParams{
			Username: cfg.TwilioAccountSID,
			Password: cfg.TwilioAuthToken,
		})
	}

	return &SMSProvider{
		config:   cfg,
		provider: cfg.SMSProvider,
		client:   client,
	}
}

// GetName returns the provider name
func (s *SMSProvider) GetName() string {
	return fmt.Sprintf("sms-%s", s.provider)
}

// GetChannel returns the notification channel
func (s *SMSProvider) GetChannel() string {
	return model.ChannelSMS
}

// Send sends an SMS notification
func (s *SMSProvider) Send(ctx context.Context, notification *model.Notification) (*SendResponse, error) {
	switch s.provider {
	case "twilio":
		return s.sendWithTwilio(ctx, notification)
	default:
		return nil, fmt.Errorf("unsupported SMS provider: %s", s.provider)
	}
}

// sendWithTwilio sends SMS using Twilio
func (s *SMSProvider) sendWithTwilio(ctx context.Context, notification *model.Notification) (*SendResponse, error) {
	if s.client == nil {
		return nil, fmt.Errorf("Twilio client not configured")
	}

	if s.config.TwilioFromNumber == "" {
		return nil, fmt.Errorf("Twilio from number not configured")
	}

	// Create message parameters
	params := &twilioApi.CreateMessageParams{}
	params.SetFrom(s.config.TwilioFromNumber)
	params.SetTo(notification.Recipient)
	params.SetBody(notification.Content)

	// Send SMS
	resp, err := s.client.Api.CreateMessage(params)
	if err != nil {
		return &SendResponse{
			Status:       StatusFailed,
			ErrorMessage: err.Error(),
		}, err
	}

	// Check response
	if resp.Sid != nil {
		status := StatusSent
		if resp.Status != nil && (*resp.Status == "failed" || *resp.Status == "undelivered") {
			status = StatusFailed
		}

		response := &SendResponse{
			ProviderID: *resp.Sid,
			Status:     status,
			Message:    "SMS sent successfully via Twilio",
			Metadata: map[string]interface{}{
				"sid": *resp.Sid,
			},
		}

		if resp.Status != nil {
			response.Metadata["twilio_status"] = *resp.Status
		}
		if resp.ErrorCode != nil {
			response.Metadata["error_code"] = *resp.ErrorCode
		}
		if resp.ErrorMessage != nil {
			response.ErrorMessage = *resp.ErrorMessage
		}

		return response, nil
	}

	return &SendResponse{
		Status:       StatusFailed,
		ErrorMessage: "No SID returned from Twilio",
	}, fmt.Errorf("Twilio API error: no SID returned")
}

// IsHealthy checks if the SMS provider is healthy
func (s *SMSProvider) IsHealthy(ctx context.Context) error {
	switch s.provider {
	case "twilio":
		if s.config.TwilioAccountSID == "" {
			return fmt.Errorf("Twilio Account SID not configured")
		}
		if s.config.TwilioAuthToken == "" {
			return fmt.Errorf("Twilio Auth Token not configured")
		}
		if s.config.TwilioFromNumber == "" {
			return fmt.Errorf("Twilio from number not configured")
		}
		if s.client == nil {
			return fmt.Errorf("Twilio client not initialized")
		}
		// Could add a test API call here
		return nil
	default:
		return fmt.Errorf("unsupported SMS provider: %s", s.provider)
	}
}
