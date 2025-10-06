package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"notification-service/internal/model"

	"github.com/google/uuid"
)

type NotificationRepository struct {
	db *sql.DB
}

func NewNotificationRepository(db *sql.DB) *NotificationRepository {
	return &NotificationRepository{
		db: db,
	}
}

// CreateNotification creates a new notification record
func (r *NotificationRepository) CreateNotification(ctx context.Context, notification *model.Notification) error {
	query := `
		INSERT INTO notifications (id, event_type, channel, recipient, subject, content, status, priority, metadata, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`

	_, err := r.db.ExecContext(ctx, query,
		notification.ID,
		notification.EventType,
		notification.Channel,
		notification.Recipient,
		notification.Subject,
		notification.Content,
		notification.Status,
		notification.Priority,
		notification.Metadata,
		notification.CreatedAt,
		notification.UpdatedAt,
	)

	return err
}

// GetNotificationByID retrieves a notification by ID
func (r *NotificationRepository) GetNotificationByID(ctx context.Context, id uuid.UUID) (*model.Notification, error) {
	query := `
		SELECT id, event_type, channel, recipient, subject, content, status, priority, 
		       provider_id, error_msg, metadata, sent_at, created_at, updated_at
		FROM notifications 
		WHERE id = $1
	`

	notification := &model.Notification{}
	var sentAt sql.NullTime
	var subject sql.NullString
	var providerID sql.NullString
	var errorMsg sql.NullString
	var metadata sql.NullString

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&notification.ID,
		&notification.EventType,
		&notification.Channel,
		&notification.Recipient,
		&subject,
		&notification.Content,
		&notification.Status,
		&notification.Priority,
		&providerID,
		&errorMsg,
		&metadata,
		&sentAt,
		&notification.CreatedAt,
		&notification.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	if subject.Valid {
		notification.Subject = &subject.String
	}
	if providerID.Valid {
		notification.ProviderID = &providerID.String
	}
	if errorMsg.Valid {
		notification.ErrorMsg = &errorMsg.String
	}
	if metadata.Valid {
		notification.Metadata = &metadata.String
	}
	if sentAt.Valid {
		notification.SentAt = &sentAt.Time
	}

	return notification, nil
}

// GetNotifications retrieves notifications with optional filters
func (r *NotificationRepository) GetNotifications(ctx context.Context, userID *uuid.UUID, notificationType, status string) ([]*model.Notification, error) {
	query := `
		SELECT id, event_type, channel, recipient, subject, content, status, priority, 
		       provider_id, error_msg, metadata, sent_at, created_at, updated_at
		FROM notifications 
		WHERE 1=1
	`
	args := []interface{}{}
	argCount := 0

	if notificationType != "" {
		argCount++
		query += fmt.Sprintf(" AND channel = $%d", argCount)
		args = append(args, notificationType)
	}

	if status != "" {
		argCount++
		query += fmt.Sprintf(" AND status = $%d", argCount)
		args = append(args, status)
	}

	query += " ORDER BY created_at DESC LIMIT 100"

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var notifications []*model.Notification
	for rows.Next() {
		notification := &model.Notification{}
		var sentAt sql.NullTime
		var subject sql.NullString
		var providerID sql.NullString
		var errorMsg sql.NullString
		var metadata sql.NullString

		err := rows.Scan(
			&notification.ID,
			&notification.EventType,
			&notification.Channel,
			&notification.Recipient,
			&subject,
			&notification.Content,
			&notification.Status,
			&notification.Priority,
			&providerID,
			&errorMsg,
			&metadata,
			&sentAt,
			&notification.CreatedAt,
			&notification.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		if subject.Valid {
			notification.Subject = &subject.String
		}
		if providerID.Valid {
			notification.ProviderID = &providerID.String
		}
		if errorMsg.Valid {
			notification.ErrorMsg = &errorMsg.String
		}
		if metadata.Valid {
			notification.Metadata = &metadata.String
		}
		if sentAt.Valid {
			notification.SentAt = &sentAt.Time
		}

		notifications = append(notifications, notification)
	}

	return notifications, rows.Err()
}

// UpdateNotificationStatus updates the status of a notification
func (r *NotificationRepository) UpdateNotificationStatus(ctx context.Context, id uuid.UUID, status string, providerMessageID, errorMessage *string) error {
	query := `
		UPDATE notifications 
		SET status = $2, provider_id = $3, error_msg = $4, 
		    sent_at = CASE WHEN $2 = 'sent' THEN $5 ELSE sent_at END,
		    updated_at = $5
		WHERE id = $1
	`

	now := time.Now()
	_, err := r.db.ExecContext(ctx, query, id, status, providerMessageID, errorMessage, now)
	return err
}
