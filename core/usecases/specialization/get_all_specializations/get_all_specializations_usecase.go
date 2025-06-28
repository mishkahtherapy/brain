package get_all_specializations

import (
	"github.com/mishkahtherapy/brain/core/domain"
	"github.com/mishkahtherapy/brain/core/ports"
)

type Usecase struct {
	specializationRepo ports.SpecializationRepository
}

func NewUsecase(specializationRepo ports.SpecializationRepository) *Usecase {
	return &Usecase{specializationRepo: specializationRepo}
}

func (u *Usecase) Execute() ([]*domain.Specialization, error) {
	return u.specializationRepo.GetAll()
}
