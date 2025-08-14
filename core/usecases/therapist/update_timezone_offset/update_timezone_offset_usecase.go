package update_timezone_offset

import (
	"errors"

	"github.com/mishkahtherapy/brain/core/domain"
	"github.com/mishkahtherapy/brain/core/domain/therapist"
	"github.com/mishkahtherapy/brain/core/ports"
)

var ErrTherapistNotFound = errors.New("therapist not found")
var ErrFailedToUpdateTherapist = errors.New("failed to update therapist")
var ErrTherapistIDIsRequired = errors.New("therapist id is required")
var ErrTimezoneOffsetIsRequired = errors.New("timezone offset is required")

type Input struct {
	TherapistID    domain.TherapistID    `json:"therapistId"`
	TimezoneOffset domain.TimezoneOffset `json:"timezoneOffset"`
}

type Usecase struct {
	therapistRepo ports.TherapistRepository
}

func NewUsecase(therapistRepo ports.TherapistRepository) *Usecase {
	return &Usecase{
		therapistRepo: therapistRepo,
	}
}

func (u *Usecase) Execute(input Input) (*therapist.Therapist, error) {
	if input.TherapistID == "" {
		return nil, ErrTherapistIDIsRequired
	}

	therapist, err := u.therapistRepo.GetByID(input.TherapistID)
	if err != nil {
		return nil, ErrTherapistNotFound
	}

	if err := u.therapistRepo.UpdateTimezoneOffset(input.TherapistID, input.TimezoneOffset); err != nil {
		return nil, ErrFailedToUpdateTherapist
	}

	therapist.TimezoneOffset = input.TimezoneOffset
	return therapist, nil
}
