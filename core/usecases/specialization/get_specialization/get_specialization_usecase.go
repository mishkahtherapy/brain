package get_specialization

import (
	"errors"

	"github.com/mishkahtherapy/brain/core/domain"
	"github.com/mishkahtherapy/brain/core/ports"
)

var ErrSpecializationNotFound = errors.New("specialization not found")

type Usecase struct {
	specializationRepo ports.SpecializationRepository
}

func NewUsecase(specializationRepo ports.SpecializationRepository) *Usecase {
	return &Usecase{specializationRepo: specializationRepo}
}

func (u *Usecase) Execute(id domain.SpecializationID) (*domain.Specialization, error) {
	specialization, err := u.specializationRepo.GetByID(id)
	if err != nil {
		return nil, ErrSpecializationNotFound
	}
	return specialization, nil
}
