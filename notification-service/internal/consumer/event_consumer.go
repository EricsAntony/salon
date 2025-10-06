package consumer

import (
	"context"
	"time"

	"notification-service/internal/model"
	"notification-service/internal/service"

	"github.com/rs/zerolog/log"
)

type EventConsumer struct {
	notificationService *service.NotificationService
}

func NewEventConsumer(notificationService *service.NotificationService) *EventConsumer {
	return &EventConsumer{
		notificationService: notificationService,
	}
}

// Start starts the event consumer (mock implementation)
func (c *EventConsumer) Start(ctx context.Context) error {
	log.Info().Msg("Starting event consumer")

	// Mock event consumption - in a real implementation, this would connect to Kafka/RabbitMQ
	go c.mockEventConsumer(ctx)

	return nil
}

// mockEventConsumer simulates consuming events from a message broker
func (c *EventConsumer) mockEventConsumer(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Info().Msg("Event consumer stopped")
			return
		case <-ticker.C:
			// Simulate processing an event
			c.processEvent(ctx, &model.Event{
				Type: "booking.confirmed",
				Data: map[string]interface{}{
					"booking_id": "123e4567-e89b-12d3-a456-426614174000",
					"user_id":    "456e7890-e89b-12d3-a456-426614174001",
					"email":      "user@example.com",
					"phone":      "+1234567890",
				},
			})
		}
	}
}

// processEvent processes an incoming event and sends appropriate notifications
func (c *EventConsumer) processEvent(ctx context.Context, event *model.Event) {
	log.Info().Str("event_type", event.Type).Msg("Processing event")

	switch event.Type {
	case "booking.confirmed":
		c.handleBookingConfirmed(ctx, event)
	case "booking.cancelled":
		c.handleBookingCancelled(ctx, event)
	case "payment.completed":
		c.handlePaymentCompleted(ctx, event)
	case "payment.failed":
		c.handlePaymentFailed(ctx, event)
	default:
		log.Warn().Str("event_type", event.Type).Msg("Unknown event type")
	}
}

// handleBookingConfirmed handles booking confirmation events
func (c *EventConsumer) handleBookingConfirmed(ctx context.Context, event *model.Event) {
	email, ok := event.Data["email"].(string)
	if !ok {
		log.Error().Msg("Missing email in booking confirmed event")
		return
	}

	userID, ok := event.Data["user_id"].(string)
	if !ok {
		log.Error().Msg("Missing user_id in booking confirmed event")
		return
	}

	// Send email notification
	emailRequest := &model.SendNotificationRequest{
		Type:      "email",
		Recipient: email,
		Subject:   "Booking Confirmed",
		Content:   "Your booking has been confirmed. We look forward to seeing you!",
		Metadata: map[string]interface{}{
			"event_type": event.Type,
			"user_id":    userID,
		},
	}

	_, err := c.notificationService.SendNotification(ctx, emailRequest)
	if err != nil {
		log.Error().Err(err).Msg("Failed to send booking confirmation email")
	}

	// Send SMS notification if phone number is available
	if phone, ok := event.Data["phone"].(string); ok && phone != "" {
		smsRequest := &model.SendNotificationRequest{
			Type:      "sms",
			Recipient: phone,
			Content:   "Your booking has been confirmed. We look forward to seeing you!",
			Metadata: map[string]interface{}{
				"event_type": event.Type,
				"user_id":    userID,
			},
		}

		_, err := c.notificationService.SendNotification(ctx, smsRequest)
		if err != nil {
			log.Error().Err(err).Msg("Failed to send booking confirmation SMS")
		}
	}
}

// handleBookingCancelled handles booking cancellation events
func (c *EventConsumer) handleBookingCancelled(ctx context.Context, event *model.Event) {
	email, ok := event.Data["email"].(string)
	if !ok {
		log.Error().Msg("Missing email in booking cancelled event")
		return
	}

	// Send email notification
	emailRequest := &model.SendNotificationRequest{
		Type:      "email",
		Recipient: email,
		Subject:   "Booking Cancelled",
		Content:   "Your booking has been cancelled. If you have any questions, please contact us.",
	}

	_, err := c.notificationService.SendNotification(ctx, emailRequest)
	if err != nil {
		log.Error().Err(err).Msg("Failed to send booking cancellation email")
	}
}

// handlePaymentCompleted handles payment completion events
func (c *EventConsumer) handlePaymentCompleted(ctx context.Context, event *model.Event) {
	email, ok := event.Data["email"].(string)
	if !ok {
		log.Error().Msg("Missing email in payment completed event")
		return
	}

	// Send email notification
	emailRequest := &model.SendNotificationRequest{
		Type:      "email",
		Recipient: email,
		Subject:   "Payment Received",
		Content:   "Thank you! Your payment has been successfully processed.",
	}

	_, err := c.notificationService.SendNotification(ctx, emailRequest)
	if err != nil {
		log.Error().Err(err).Msg("Failed to send payment completion email")
	}
}

// handlePaymentFailed handles payment failure events
func (c *EventConsumer) handlePaymentFailed(ctx context.Context, event *model.Event) {
	email, ok := event.Data["email"].(string)
	if !ok {
		log.Error().Msg("Missing email in payment failed event")
		return
	}

	// Send email notification
	emailRequest := &model.SendNotificationRequest{
		Type:      "email",
		Recipient: email,
		Subject:   "Payment Failed",
		Content:   "We were unable to process your payment. Please try again or contact us for assistance.",
	}

	_, err := c.notificationService.SendNotification(ctx, emailRequest)
	if err != nil {
		log.Error().Err(err).Msg("Failed to send payment failure email")
	}
}
