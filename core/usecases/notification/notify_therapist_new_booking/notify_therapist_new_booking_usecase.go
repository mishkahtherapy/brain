package notify_therapist_new_booking

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/mishkahtherapy/brain/core/domain"
	"github.com/mishkahtherapy/brain/core/ports"
)

type Usecase struct {
	therapistRepo       ports.TherapistRepository
	notificationPort    ports.NotificationPort
	notificationRepo    ports.NotificationRepository
	therapistAppBaseURL string
}

func NewUsecase(
	therapistRepo ports.TherapistRepository,
	notificationPort ports.NotificationPort,
	notificationRepo ports.NotificationRepository,
	therapistAppBaseURL string,
) *Usecase {
	return &Usecase{
		therapistRepo:       therapistRepo,
		notificationPort:    notificationPort,
		notificationRepo:    notificationRepo,
		therapistAppBaseURL: therapistAppBaseURL,
	}
}

func (u *Usecase) Execute(session *domain.Session) {
	therapist, err := u.therapistRepo.GetByID(session.TherapistID)
	if err != nil {
		slog.Warn("failed to get therapist for notification", "therapist_id", therapist.ID, "error", err)
		return
	}

	if therapist.DeviceID == "" {
		slog.Info("therapist has no device id, skipping notification", "therapist_id", therapist.ID)
		return
	}
	therapistTimezoneOffset := int(therapist.TimezoneOffset / 60)
	timezoneLabel := fmt.Sprintf("UTC%+d", therapistTimezoneOffset)
	therapistTimezone := time.FixedZone(timezoneLabel, therapistTimezoneOffset)
	therapistTime := time.Date(session.StartTime.Year(), session.StartTime.Month(), session.StartTime.Day(), 0, 0, 0, 0, therapistTimezone)
	notification := ports.Notification{
		Title:    "Session Confirmed",
		Body:     fmt.Sprintf("Your next session is confirmed on %s", therapistTime.Format(time.DateOnly)),
		ImageURL: "https://therapist.mishkahtherapy.com/mishkah-logo.png",
		// TODO: add session id to the link
		Link: fmt.Sprintf("%s/sessions", u.therapistAppBaseURL),
	}

	firebaseNotificationId, err := u.notificationPort.SendNotification(therapist.DeviceID, notification)
	if err != nil {
		slog.Warn("failed to notify therapist",
			slog.Group(
				"therapist",
				"id", therapist.ID,
				"device_id", therapist.DeviceID,
				"name", therapist.Name,
				"notification", notification.Body,
			),
			"sessionID", session.ID,
			"error", err)
		return
	}

	// Persist the notification
	err = u.notificationRepo.CreateNotification(therapist.ID, *firebaseNotificationId, notification)
	if err != nil {
		slog.Warn("failed to persist notification",
			slog.Group(
				"therapist",
				"id", therapist.ID,
				"device_id", therapist.DeviceID,
				"name", therapist.Name,
				"notification", notification.Body,
			),
			"sessionID", session.ID,
			"error", err)
		return
	}
}
