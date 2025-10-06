package provider

import (
	"context"
	"fmt"

	"notification-service/internal/config"
	"notification-service/internal/model"

	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
	"gopkg.in/gomail.v2"
)

// EmailProvider implements NotificationProvider for email notifications
type EmailProvider struct {
	config   *config.Config
	provider string // "sendgrid" or "smtp"
}

// NewEmailProvider creates a new email provider
func NewEmailProvider(cfg *config.Config) *EmailProvider {
	return &EmailProvider{
		config:   cfg,
		provider: cfg.EmailProvider,
	}
}

// GetName returns the provider name
func (e *EmailProvider) GetName() string {
	return fmt.Sprintf("email-%s", e.provider)
}

// GetChannel returns the notification channel
func (e *EmailProvider) GetChannel() string {
	return model.ChannelEmail
}

// Send sends an email notification
func (e *EmailProvider) Send(ctx context.Context, notification *model.Notification) (*SendResponse, error) {
	switch e.provider {
	case "sendgrid":
		return e.sendWithSendGrid(ctx, notification)
	case "smtp":
		return e.sendWithSMTP(ctx, notification)
	default:
		return nil, fmt.Errorf("unsupported email provider: %s", e.provider)
	}
}

// sendWithSendGrid sends email using SendGrid
func (e *EmailProvider) sendWithSendGrid(ctx context.Context, notification *model.Notification) (*SendResponse, error) {
	if e.config.SendGridAPIKey == "" {
		return nil, fmt.Errorf("SendGrid API key not configured")
	}

	// Create SendGrid message
	from := mail.NewEmail(e.config.SMTPFromName, e.config.SMTPFromEmail)
	to := mail.NewEmail("", notification.Recipient)
	
	subject := "Notification"
	if notification.Subject != nil {
		subject = *notification.Subject
	}

	message := mail.NewSingleEmail(from, subject, to, notification.Content, notification.Content)

	// Send email
	client := sendgrid.NewSendClient(e.config.SendGridAPIKey)
	response, err := client.SendWithContext(ctx, message)
	if err != nil {
		return &SendResponse{
			Status:       StatusFailed,
			ErrorMessage: err.Error(),
		}, err
	}

	// Check response status
	if response.StatusCode >= 200 && response.StatusCode < 300 {
		return &SendResponse{
			ProviderID: fmt.Sprintf("sendgrid-%d", response.StatusCode),
			Status:     StatusSent,
			Message:    "Email sent successfully via SendGrid",
			Metadata: map[string]interface{}{
				"status_code": response.StatusCode,
				"headers":     response.Headers,
			},
		}, nil
	}

	return &SendResponse{
		Status:       StatusFailed,
		ErrorMessage: fmt.Sprintf("SendGrid returned status code: %d", response.StatusCode),
		Metadata: map[string]interface{}{
			"status_code": response.StatusCode,
			"body":        response.Body,
		},
	}, fmt.Errorf("SendGrid API error: status %d", response.StatusCode)
}

// sendWithSMTP sends email using SMTP
func (e *EmailProvider) sendWithSMTP(ctx context.Context, notification *model.Notification) (*SendResponse, error) {
	if e.config.SMTPHost == "" {
		return nil, fmt.Errorf("SMTP host not configured")
	}

	// Create email message
	m := gomail.NewMessage()
	m.SetHeader("From", fmt.Sprintf("%s <%s>", e.config.SMTPFromName, e.config.SMTPFromEmail))
	m.SetHeader("To", notification.Recipient)
	
	subject := "Notification"
	if notification.Subject != nil {
		subject = *notification.Subject
	}
	m.SetHeader("Subject", subject)
	m.SetBody("text/html", notification.Content)

	// Configure SMTP dialer
	d := gomail.NewDialer(e.config.SMTPHost, e.config.SMTPPort, e.config.SMTPUsername, e.config.SMTPPassword)

	// Send email
	if err := d.DialAndSend(m); err != nil {
		return &SendResponse{
			Status:       StatusFailed,
			ErrorMessage: err.Error(),
		}, err
	}

	return &SendResponse{
		ProviderID: fmt.Sprintf("smtp-%s", e.config.SMTPHost),
		Status:     StatusSent,
		Message:    "Email sent successfully via SMTP",
		Metadata: map[string]interface{}{
			"smtp_host": e.config.SMTPHost,
			"smtp_port": e.config.SMTPPort,
		},
	}, nil
}

// IsHealthy checks if the email provider is healthy
func (e *EmailProvider) IsHealthy(ctx context.Context) error {
	switch e.provider {
	case "sendgrid":
		if e.config.SendGridAPIKey == "" {
			return fmt.Errorf("SendGrid API key not configured")
		}
		// Could add a test API call here
		return nil
	case "smtp":
		if e.config.SMTPHost == "" {
			return fmt.Errorf("SMTP host not configured")
		}
		// Could add a test connection here
		return nil
	default:
		return fmt.Errorf("unsupported email provider: %s", e.provider)
	}
}
