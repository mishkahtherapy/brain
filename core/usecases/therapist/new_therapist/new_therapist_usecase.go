package new_therapist

import (
	"errors"

	"github.com/mishkahtherapy/brain/core/domain"
	"github.com/mishkahtherapy/brain/core/ports"
	therapistvalidation "github.com/mishkahtherapy/brain/core/usecases/therapist"
)

var ErrFailedToCreateTherapist = errors.New("failed to create therapist")
var ErrFailedToGetTherapist = errors.New("failed to get therapist")
var ErrSpecializationNotFound = errors.New("specialization not found")
var ErrFailedToGetSpecializations = errors.New("failed to get specializations")

type Input struct {
	Name              string                    `json:"name"`
	Email             domain.Email              `json:"email"`
	PhoneNumber       domain.PhoneNumber        `json:"phoneNumber"`
	WhatsAppNumber    domain.WhatsAppNumber     `json:"whatsAppNumber"`
	SpeaksEnglish     bool                      `json:"speaksEnglish"`
	SpecializationIDs []domain.SpecializationID `json:"specializationIds"`
}

type Usecase struct {
	therapistRepo      ports.TherapistRepository
	specializationRepo ports.SpecializationRepository
}

func NewUsecase(therapistRepo ports.TherapistRepository, specializationRepo ports.SpecializationRepository) *Usecase {
	return &Usecase{therapistRepo: therapistRepo, specializationRepo: specializationRepo}
}

func (u *Usecase) Execute(input Input) (*domain.Therapist, error) {
	// Validate required fields
	if err := therapistvalidation.ValidateRequiredFields(input.Name, input.Email, input.PhoneNumber, input.WhatsAppNumber); err != nil {
		return nil, err
	}

	// Validate phone number formats
	if err := therapistvalidation.ValidatePhoneNumbers(input.PhoneNumber, input.WhatsAppNumber); err != nil {
		return nil, err
	}

	// Validate specializations exist
	if err := validateSpecializations(u.specializationRepo, input.SpecializationIDs); err != nil {
		return nil, err
	}

	// Validate email and WhatsApp uniqueness
	if err := therapistvalidation.ValidateUniquenessForCreate(u.therapistRepo, input.Email, input.WhatsAppNumber); err != nil {
		return nil, err
	}

	// Create therapist entity
	newTherapist := &domain.Therapist{
		ID:             domain.NewTherapistID(),
		Name:           input.Name,
		Email:          input.Email,
		PhoneNumber:    input.PhoneNumber,
		WhatsAppNumber: input.WhatsAppNumber,
		SpeaksEnglish:  input.SpeaksEnglish,
	}

	// Add specializations
	specializations := make([]domain.Specialization, 0)
	for _, specID := range input.SpecializationIDs {
		specializations = append(specializations, domain.Specialization{ID: specID})
	}
	newTherapist.Specializations = specializations

	// Set timestamps
	timestamp := domain.NewUTCTimestamp()
	newTherapist.CreatedAt = timestamp
	newTherapist.UpdatedAt = timestamp

	// Save therapist
	if err := u.therapistRepo.Create(newTherapist); err != nil {
		return nil, ErrFailedToCreateTherapist
	}

	return newTherapist, nil
}

func validateSpecializations(specializationRepo ports.SpecializationRepository, specializationIDs []domain.SpecializationID) error {
	dbSpecializations, err := specializationRepo.BulkGetByIds(specializationIDs)
	if err != nil {
		return ErrFailedToGetSpecializations
	}
	for _, specializationID := range specializationIDs {
		if _, ok := dbSpecializations[specializationID]; !ok {
			return ErrSpecializationNotFound
		}
	}
	return nil
}
