package get_all_therapists

import (
	"github.com/mishkahtherapy/brain/core/domain/therapist"
	"github.com/mishkahtherapy/brain/core/ports"
)

type Usecase struct {
	therapistRepo ports.TherapistRepository
}

func NewUsecase(therapistRepo ports.TherapistRepository) *Usecase {
	return &Usecase{therapistRepo: therapistRepo}
}

func (u *Usecase) Execute() ([]*therapist.Therapist, error) {
	return u.therapistRepo.List()
}
