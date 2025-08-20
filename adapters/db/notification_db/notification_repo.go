package notification_db

import (
	"errors"
	"log/slog"

	"github.com/mishkahtherapy/brain/core/domain"
	"github.com/mishkahtherapy/brain/core/ports"
)

var ErrFailedToCreateNotification = errors.New("failed to create notification")
var ErrTherapistIDIsRequired = errors.New("therapist id is required")

type NotificationRepository struct {
	db ports.SQLDatabase
}

func NewNotificationRepository(db ports.SQLDatabase) ports.NotificationRepository {
	return &NotificationRepository{db: db}
}

func (r *NotificationRepository) CreateNotification(
	therapistID domain.TherapistID,
	firebaseNotificationID ports.NotificationID,
	notification ports.Notification,
) error {
	if therapistID == "" {
		return ErrTherapistIDIsRequired
	}

	query := `INSERT INTO push_notifications (therapist_id, firebase_notification_id, title, body, image_url) VALUES (?, ?, ?, ?, ?)`
	_, err := r.db.Exec(query, therapistID, string(firebaseNotificationID), notification.Title, notification.Body, notification.ImageURL)
	if err != nil {
		slog.Error("error creating notification", "error", err)
		return ErrFailedToCreateNotification
	}

	return nil
}
