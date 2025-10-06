package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

// NotificationClient handles communication with the notification service
type NotificationClient struct {
	baseURL    string
	httpClient *http.Client
}

// NewNotificationClient creates a new notification service client
func NewNotificationClient(baseURL string) *NotificationClient {
	return &NotificationClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// Notification request/response structures
type SendNotificationRequest struct {
	UserID    *uuid.UUID             `json:"user_id,omitempty"`
	Type      string                 `json:"type"`
	Recipient string                 `json:"recipient"`
	Subject   string                 `json:"subject,omitempty"`
	Content   string                 `json:"content"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

type SendNotificationResponse struct {
	ID        uuid.UUID `json:"id"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

// BookingEvent represents a booking event for notifications
type BookingEvent struct {
	Type      string                 `json:"type"`
	BookingID uuid.UUID              `json:"booking_id"`
	UserID    uuid.UUID              `json:"user_id"`
	SalonID   uuid.UUID              `json:"salon_id"`
	BranchID  uuid.UUID              `json:"branch_id"`
	Data      map[string]interface{} `json:"data"`
	Timestamp time.Time              `json:"timestamp"`
}

// SendBookingConfirmationNotification sends booking confirmation notifications
func (c *NotificationClient) SendBookingConfirmationNotification(ctx context.Context, bookingEvent *BookingEvent) error {
	// Get user details from the event data
	userEmail, _ := bookingEvent.Data["user_email"].(string)
	userPhone, _ := bookingEvent.Data["user_phone"].(string)
	userName, _ := bookingEvent.Data["user_name"].(string)
	salonName, _ := bookingEvent.Data["salon_name"].(string)
	totalAmount, _ := bookingEvent.Data["total_amount"].(float64)
	bookingTime, _ := bookingEvent.Data["booking_time"].(string)

	if userEmail == "" && userPhone == "" {
		return fmt.Errorf("no contact information available for user")
	}

	// Prepare notification metadata
	metadata := map[string]interface{}{
		"booking_id":   bookingEvent.BookingID.String(),
		"user_id":      bookingEvent.UserID.String(),
		"salon_name":   salonName,
		"total_amount": totalAmount,
		"booking_time": bookingTime,
		"event_type":   bookingEvent.Type,
	}

	// Send email notification
	if userEmail != "" {
		emailRequest := &SendNotificationRequest{
			UserID:    &bookingEvent.UserID,
			Type:      "email",
			Recipient: userEmail,
			Subject:   fmt.Sprintf("Booking Confirmed - %s", salonName),
			Content: fmt.Sprintf(`Dear %s,

Your booking has been confirmed!

Salon: %s
Date & Time: %s
Total Amount: ₹%.2f

Thank you for choosing our services. We look forward to serving you!

Best regards,
%s Team`, userName, salonName, bookingTime, totalAmount, salonName),
			Metadata: metadata,
		}

		if err := c.SendNotification(ctx, emailRequest); err != nil {
			log.Error().Err(err).Msg("Failed to send booking confirmation email")
		}
	}

	// Send SMS notification
	if userPhone != "" {
		smsRequest := &SendNotificationRequest{
			UserID:    &bookingEvent.UserID,
			Type:      "sms",
			Recipient: userPhone,
			Content: fmt.Sprintf("Hi %s! Your booking at %s on %s is confirmed. Amount: ₹%.2f. Thank you!",
				userName, salonName, bookingTime, totalAmount),
			Metadata: metadata,
		}

		if err := c.SendNotification(ctx, smsRequest); err != nil {
			log.Error().Err(err).Msg("Failed to send booking confirmation SMS")
		}
	}

	return nil
}

// SendBookingCancellationNotification sends booking cancellation notifications
func (c *NotificationClient) SendBookingCancellationNotification(ctx context.Context, bookingEvent *BookingEvent) error {
	userEmail, _ := bookingEvent.Data["user_email"].(string)
	userPhone, _ := bookingEvent.Data["user_phone"].(string)
	userName, _ := bookingEvent.Data["user_name"].(string)
	salonName, _ := bookingEvent.Data["salon_name"].(string)
	bookingTime, _ := bookingEvent.Data["booking_time"].(string)
	reason, _ := bookingEvent.Data["reason"].(string)

	if userEmail == "" && userPhone == "" {
		return fmt.Errorf("no contact information available for user")
	}

	metadata := map[string]interface{}{
		"booking_id":   bookingEvent.BookingID.String(),
		"user_id":      bookingEvent.UserID.String(),
		"salon_name":   salonName,
		"booking_time": bookingTime,
		"reason":       reason,
		"event_type":   bookingEvent.Type,
	}

	// Send email notification
	if userEmail != "" {
		emailRequest := &SendNotificationRequest{
			UserID:    &bookingEvent.UserID,
			Type:      "email",
			Recipient: userEmail,
			Subject:   fmt.Sprintf("Booking Cancelled - %s", salonName),
			Content: fmt.Sprintf(`Dear %s,

Your booking has been cancelled.

Salon: %s
Date & Time: %s
Reason: %s

If you have any questions, please contact us.

Best regards,
%s Team`, userName, salonName, bookingTime, reason, salonName),
			Metadata: metadata,
		}

		if err := c.SendNotification(ctx, emailRequest); err != nil {
			log.Error().Err(err).Msg("Failed to send booking cancellation email")
		}
	}

	// Send SMS notification
	if userPhone != "" {
		smsRequest := &SendNotificationRequest{
			UserID:    &bookingEvent.UserID,
			Type:      "sms",
			Recipient: userPhone,
			Content: fmt.Sprintf("Hi %s! Your booking at %s on %s has been cancelled. Reason: %s",
				userName, salonName, bookingTime, reason),
			Metadata: metadata,
		}

		if err := c.SendNotification(ctx, smsRequest); err != nil {
			log.Error().Err(err).Msg("Failed to send booking cancellation SMS")
		}
	}

	return nil
}

// SendPaymentConfirmationNotification sends payment confirmation notifications
func (c *NotificationClient) SendPaymentConfirmationNotification(ctx context.Context, bookingEvent *BookingEvent) error {
	userEmail, _ := bookingEvent.Data["user_email"].(string)
	userPhone, _ := bookingEvent.Data["user_phone"].(string)
	userName, _ := bookingEvent.Data["user_name"].(string)
	salonName, _ := bookingEvent.Data["salon_name"].(string)
	totalAmount, _ := bookingEvent.Data["total_amount"].(float64)
	paymentID, _ := bookingEvent.Data["payment_id"].(string)

	if userEmail == "" && userPhone == "" {
		return fmt.Errorf("no contact information available for user")
	}

	metadata := map[string]interface{}{
		"booking_id":   bookingEvent.BookingID.String(),
		"user_id":      bookingEvent.UserID.String(),
		"salon_name":   salonName,
		"total_amount": totalAmount,
		"payment_id":   paymentID,
		"event_type":   bookingEvent.Type,
	}

	// Send email notification
	if userEmail != "" {
		emailRequest := &SendNotificationRequest{
			UserID:    &bookingEvent.UserID,
			Type:      "email",
			Recipient: userEmail,
			Subject:   fmt.Sprintf("Payment Received - %s", salonName),
			Content: fmt.Sprintf(`Dear %s,

Thank you! Your payment has been successfully processed.

Salon: %s
Amount Paid: ₹%.2f
Payment ID: %s

Your booking is now confirmed. We look forward to serving you!

Best regards,
%s Team`, userName, salonName, totalAmount, paymentID, salonName),
			Metadata: metadata,
		}

		if err := c.SendNotification(ctx, emailRequest); err != nil {
			log.Error().Err(err).Msg("Failed to send payment confirmation email")
		}
	}

	// Send SMS notification
	if userPhone != "" {
		smsRequest := &SendNotificationRequest{
			UserID:    &bookingEvent.UserID,
			Type:      "sms",
			Recipient: userPhone,
			Content: fmt.Sprintf("Hi %s! Payment of ₹%.2f for %s received successfully. Booking confirmed!",
				userName, totalAmount, salonName),
			Metadata: metadata,
		}

		if err := c.SendNotification(ctx, smsRequest); err != nil {
			log.Error().Err(err).Msg("Failed to send payment confirmation SMS")
		}
	}

	return nil
}

// SendNotification sends a notification via the notification service
func (c *NotificationClient) SendNotification(ctx context.Context, request *SendNotificationRequest) error {
	url := fmt.Sprintf("%s/api/v1/notifications/send", c.baseURL)
	
	body, err := json.Marshal(request)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("notification service returned status %d", resp.StatusCode)
	}

	var response SendNotificationResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	log.Info().
		Str("notification_id", response.ID.String()).
		Str("type", request.Type).
		Str("recipient", request.Recipient).
		Msg("Notification sent successfully")

	return nil
}
