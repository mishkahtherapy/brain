package get_all_therapists

import (
	"github.com/mishkahtherapy/brain/core/domain"
	"github.com/mishkahtherapy/brain/core/ports"
)

type Usecase struct {
	therapistRepo ports.TherapistRepository
}

func NewUsecase(therapistRepo ports.TherapistRepository) *Usecase {
	return &Usecase{therapistRepo: therapistRepo}
}

func (u *Usecase) Execute() ([]*domain.Therapist, error) {
	return u.therapistRepo.List()
}
