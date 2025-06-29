package update_therapist_specializations

import (
	"errors"

	"github.com/mishkahtherapy/brain/core/domain"
	"github.com/mishkahtherapy/brain/core/ports"
)

var ErrTherapistNotFound = errors.New("therapist not found")
var ErrFailedToUpdateTherapist = errors.New("failed to update therapist")
var ErrSpecializationNotFound = errors.New("one or more specializations not found")

type Input struct {
	TherapistID       domain.TherapistID        `json:"therapistId"`
	SpecializationIDs []domain.SpecializationID `json:"specializationIds"`
}

type Usecase struct {
	therapistRepo      ports.TherapistRepository
	specializationRepo ports.SpecializationRepository
}

func NewUsecase(therapistRepo ports.TherapistRepository, specializationRepo ports.SpecializationRepository) *Usecase {
	return &Usecase{
		therapistRepo:      therapistRepo,
		specializationRepo: specializationRepo,
	}
}

func (u *Usecase) Execute(input Input) (*domain.Therapist, error) {
	// Get the existing therapist
	therapist, err := u.therapistRepo.GetByID(input.TherapistID)
	if err != nil {
		return nil, ErrTherapistNotFound
	}
	if therapist == nil {
		return nil, ErrTherapistNotFound
	}

	specializations := make([]domain.Specialization, 0)
	// Bulk validate that all specializations exist
	if len(input.SpecializationIDs) > 0 {
		foundSpecializationMap, err := u.specializationRepo.BulkGetByIds(input.SpecializationIDs)
		if err != nil {
			return nil, ErrSpecializationNotFound
		}

		// Check if all requested specializations were found
		for _, specID := range input.SpecializationIDs {
			if _, exists := foundSpecializationMap[specID]; !exists {
				return nil, ErrSpecializationNotFound
			}
		}

		for _, specID := range input.SpecializationIDs {
			specializations = append(specializations, *foundSpecializationMap[specID])
		}
	}

	// Update the therapist's specializations
	therapist.Specializations = specializations
	therapist.UpdatedAt = domain.NewUTCTimestamp()

	// Save the updated therapist
	err = u.therapistRepo.UpdateSpecializations(input.TherapistID, input.SpecializationIDs)
	if err != nil {
		return nil, ErrFailedToUpdateTherapist
	}

	return therapist, nil
}
