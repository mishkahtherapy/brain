package firebase_notifier

import (
	"context"
	"fmt"
	"log"
	"log/slog"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/messaging"
	"github.com/mishkahtherapy/brain/core/domain"
	"github.com/mishkahtherapy/brain/core/ports"
	"google.golang.org/api/option"
)

type FirebaseNotifier struct {
	messagingClient *messaging.Client
}

func NewFirebaseNotifier(firebaseServiceAccountPath string) ports.NotificationPort {
	firebaseApp, err := firebase.NewApp(context.Background(), nil, option.WithCredentialsFile(firebaseServiceAccountPath))
	if err != nil {
		log.Fatalf("error initializing app: %v\n", err)
	}
	messagingClient, err := firebaseApp.Messaging(context.Background())
	if err != nil {
		log.Fatalf("error initializing messaging client: %v\n", err)
	}
	return &FirebaseNotifier{messagingClient: messagingClient}
}

func (f *FirebaseNotifier) SendNotification(
	deviceID domain.DeviceID,
	notification ports.Notification,
) (*ports.NotificationID, error) {

	webPushConfig := &messaging.WebpushConfig{
		Notification: &messaging.WebpushNotification{
			Title: notification.Title,
			Body:  notification.Body,
			Icon:  notification.ImageURL,
		},
	}
	if notification.Link != "" {
		webPushConfig.FCMOptions = &messaging.WebpushFCMOptions{
			Link: notification.Link,
		}
	}

	message := &messaging.Message{
		Token: string(deviceID),
		Notification: &messaging.Notification{
			Title:    notification.Title,
			Body:     notification.Body,
			ImageURL: notification.ImageURL,
		},
		Android: &messaging.AndroidConfig{
			Priority: "high",
		},
		// APNS: &messaging.APNSConfig{

		// 	Headers: map[string]string{
		// 		"apns-priority": "5",
		// 	},
		// 	Payload: &messaging.APNSPayload{
		// 		Aps: &messaging.Aps{
		// 			ContentAvailable: true,
		// 		},
		// 	},
		// },
		Webpush: webPushConfig,
	}

	firebaseNotificationId, err := f.messagingClient.Send(context.Background(), message)
	if err != nil {
		slog.Error("error sending notification", slog.String("error", err.Error()), slog.String("device_id", string(deviceID)), slog.String("notification", fmt.Sprintf("%+v", notification)))
		return nil, ports.ErrNotificationFailed
	}

	slog.Info("sent notification",
		slog.Group("notification",
			slog.String("device_id", string(deviceID)),
			slog.String("notification", fmt.Sprintf("%+v", notification)),
			slog.String("firebase_notification_id", string(firebaseNotificationId)),
		),
	)

	notificationID := ports.NotificationID(firebaseNotificationId)
	return &notificationID, nil
}
