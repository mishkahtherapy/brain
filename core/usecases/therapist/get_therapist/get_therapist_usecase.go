package get_therapist

import (
	"errors"

	"github.com/mishkahtherapy/brain/core/domain"
	"github.com/mishkahtherapy/brain/core/ports"
)

var ErrTherapistNotFound = errors.New("therapist not found")

type Usecase struct {
	therapistRepo ports.TherapistRepository
}

func NewUsecase(therapistRepo ports.TherapistRepository) *Usecase {
	return &Usecase{therapistRepo: therapistRepo}
}

func (u *Usecase) Execute(id domain.TherapistID) (*domain.Therapist, error) {
	therapist, err := u.therapistRepo.GetByID(id)
	if err != nil {
		return nil, ErrTherapistNotFound
	}
	return therapist, nil
}
