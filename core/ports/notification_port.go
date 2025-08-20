package ports

import (
	"errors"

	"github.com/mishkahtherapy/brain/core/domain"
)

type NotificationID string

type Notification struct {
	Title    string `json:"title,omitempty"`
	Body     string `json:"body,omitempty"`
	ImageURL string `json:"image,omitempty"`
	Link     string `json:"link,omitempty"`
}

var ErrNotificationFailed = errors.New("notification failed")

type NotificationPort interface {
	SendNotification(deviceID domain.DeviceID, notification Notification) (*NotificationID, error)
}

type NotificationRepository interface {
	CreateNotification(therapistID domain.TherapistID, firebaseNotificationID NotificationID, notification Notification) error
}
