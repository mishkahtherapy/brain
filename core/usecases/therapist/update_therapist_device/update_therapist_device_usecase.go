package update_therapist_device

import (
	"errors"

	"github.com/mishkahtherapy/brain/core/domain"
	"github.com/mishkahtherapy/brain/core/ports"
)

var ErrTherapistIDIsRequired = errors.New("therapist id is required")
var ErrDeviceIDIsRequired = errors.New("device id is required")

type Input struct {
	TherapistID domain.TherapistID `json:"therapistId"`
	DeviceID    domain.DeviceID    `json:"deviceId"`
}

type Usecase struct {
	therapistRepo    ports.TherapistRepository
	notificationPort ports.NotificationPort
}

func NewUsecase(therapistRepo ports.TherapistRepository, notificationPort ports.NotificationPort) *Usecase {
	return &Usecase{
		therapistRepo:    therapistRepo,
		notificationPort: notificationPort,
	}
}

func (u *Usecase) Execute(input Input) error {
	if input.TherapistID == "" {
		return ErrTherapistIDIsRequired
	}

	if input.DeviceID == "" {
		return ErrDeviceIDIsRequired
	}

	deviceIDUpdatedAt := domain.NewUTCTimestamp()
	err := u.therapistRepo.UpdateDevice(input.TherapistID, input.DeviceID, deviceIDUpdatedAt)
	if err != nil {
		return err
	}

	notification := ports.Notification{
		Title: "Device updated",
		Body:  "Your device has been updated",
	}
	_, err = u.notificationPort.SendNotification(input.DeviceID, notification)

	if err != nil {
		return err
	}

	return nil
}
