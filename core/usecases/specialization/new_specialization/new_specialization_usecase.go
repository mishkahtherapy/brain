package new_specialization

import (
	"errors"
	"strings"

	"github.com/mishkahtherapy/brain/core/domain"
	"github.com/mishkahtherapy/brain/core/ports"
	"github.com/mishkahtherapy/brain/core/usecases/common"
)

// Domain-specific errors that are not common across usecases
var ErrSpecializationAlreadyExists = errors.New("specialization already exists")

type Input struct {
	Name string `json:"name"`
}

type Usecase struct {
	specializationRepo ports.SpecializationRepository
}

func NewUsecase(specializationRepo ports.SpecializationRepository) *Usecase {
	return &Usecase{specializationRepo: specializationRepo}
}

func (u *Usecase) Execute(input Input) (*domain.Specialization, error) {
	if input.Name == "" {
		return nil, common.ErrNameIsRequired
	}

	existingSpecialization, err := u.specializationRepo.GetByName(input.Name)
	if err != nil {
		return nil, common.ErrFailedToGetSpecializations
	}
	if existingSpecialization != nil {
		return nil, ErrSpecializationAlreadyExists
	}

	now := domain.NewUTCTimestamp()
	specialization := &domain.Specialization{
		ID:        domain.NewSpecializationID(),
		Name:      cleanUpName(input.Name),
		CreatedAt: now,
		UpdatedAt: now,
	}

	err = u.specializationRepo.Create(specialization)
	if err != nil {
		return nil, common.ErrFailedToCreateSpecialization
	}

	return specialization, nil
}

func cleanUpName(name string) string {
	return strings.TrimSpace(strings.ToLower(name))
}
