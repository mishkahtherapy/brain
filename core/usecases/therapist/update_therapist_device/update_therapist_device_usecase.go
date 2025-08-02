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
	therapistRepo ports.TherapistRepository
}

func NewUsecase(therapistRepo ports.TherapistRepository) *Usecase {
	return &Usecase{
		therapistRepo: therapistRepo,
	}
}

func (u *Usecase) Execute(input Input) error {
	if input.TherapistID == "" {
		return ErrTherapistIDIsRequired
	}

	if input.DeviceID == "" {
		return ErrDeviceIDIsRequired
	}

	err := u.therapistRepo.UpdateDevice(input.TherapistID, input.DeviceID)
	if err != nil {
		return err
	}

	return nil
}
