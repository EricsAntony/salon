package service

import (
	"context"
	"encoding/json"
	"time"

	"notification-service/internal/model"
	"notification-service/internal/provider"
	"notification-service/internal/repository"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

type NotificationService struct {
	notificationRepo *repository.NotificationRepository
	providerManager  provider.ProviderManager
}

func NewNotificationService(notificationRepo *repository.NotificationRepository, providerManager provider.ProviderManager) *NotificationService {
	return &NotificationService{
		notificationRepo: notificationRepo,
		providerManager:  providerManager,
	}
}

// SendNotification sends a notification using the appropriate provider
func (s *NotificationService) SendNotification(ctx context.Context, request *model.SendNotificationRequest) (*model.Notification, error) {
	// Create notification record
	notification := &model.Notification{
		ID:        uuid.New(),
		EventType: "manual_send",
		Channel:   request.Type,
		Recipient: request.Recipient,
		Subject:   &request.Subject,
		Content:   request.Content,
		Status:    "pending",
		Priority:  "normal",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Add metadata if provided
	if request.Metadata != nil {
		metadataJSON, _ := json.Marshal(request.Metadata)
		metadataStr := string(metadataJSON)
		notification.Metadata = &metadataStr
	}

	// Save to database
	if err := s.notificationRepo.CreateNotification(ctx, notification); err != nil {
		log.Error().Err(err).Msg("Failed to create notification record")
		return nil, err
	}

	// Send notification asynchronously
	go s.sendNotificationAsync(ctx, notification)

	return notification, nil
}

// GetNotification retrieves a notification by ID
func (s *NotificationService) GetNotification(ctx context.Context, id uuid.UUID) (*model.Notification, error) {
	return s.notificationRepo.GetNotificationByID(ctx, id)
}

// GetNotifications retrieves notifications with optional filters
func (s *NotificationService) GetNotifications(ctx context.Context, userID *uuid.UUID, notificationType, status string) ([]*model.Notification, error) {
	return s.notificationRepo.GetNotifications(ctx, userID, notificationType, status)
}

// sendNotificationAsync sends the notification using the appropriate provider
func (s *NotificationService) sendNotificationAsync(ctx context.Context, notification *model.Notification) {
	// Get the appropriate provider
	notificationProvider, err := s.providerManager.GetProvider(notification.Channel)
	if err != nil {
		log.Error().Err(err).Str("channel", notification.Channel).Msg("Failed to get provider")
		s.updateNotificationStatus(ctx, notification.ID, "failed", nil, stringPtr(err.Error()))
		return
	}

	// Send the notification
	response, err := notificationProvider.Send(ctx, notification)
	if err != nil {
		log.Error().Err(err).Str("notification_id", notification.ID.String()).Msg("Failed to send notification")
		s.updateNotificationStatus(ctx, notification.ID, "failed", nil, stringPtr(err.Error()))
	} else {
		log.Info().Str("notification_id", notification.ID.String()).Str("provider_id", response.ProviderID).Msg("Notification sent successfully")
		s.updateNotificationStatus(ctx, notification.ID, "sent", stringPtr(response.ProviderID), nil)
	}
}

// updateNotificationStatus updates the notification status in the database
func (s *NotificationService) updateNotificationStatus(ctx context.Context, id uuid.UUID, status string, providerMessageID, errorMessage *string) {
	if err := s.notificationRepo.UpdateNotificationStatus(ctx, id, status, providerMessageID, errorMessage); err != nil {
		log.Error().Err(err).Str("notification_id", id.String()).Msg("Failed to update notification status")
	}
}

// getProviderName returns the provider name for a notification type (deprecated - using provider manager now)
func (s *NotificationService) getProviderName(notificationType string) string {
	switch notificationType {
	case "email":
		return "sendgrid" // or "smtp"
	case "sms":
		return "twilio"
	case "push":
		return "fcm" // or "apns"
	default:
		return "unknown"
	}
}

// stringPtr returns a pointer to a string
func stringPtr(s string) *string {
	return &s
}
