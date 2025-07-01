package get_therapist

import (
	"github.com/mishkahtherapy/brain/core/domain"
	"github.com/mishkahtherapy/brain/core/ports"
	"github.com/mishkahtherapy/brain/core/usecases/common"
)

type Usecase struct {
	therapistRepo ports.TherapistRepository
}

func NewUsecase(therapistRepo ports.TherapistRepository) *Usecase {
	return &Usecase{therapistRepo: therapistRepo}
}

func (u *Usecase) Execute(id domain.TherapistID) (*domain.Therapist, error) {
	therapist, err := u.therapistRepo.GetByID(id)
	if err != nil {
		return nil, common.ErrTherapistNotFound
	}
	return therapist, nil
}
