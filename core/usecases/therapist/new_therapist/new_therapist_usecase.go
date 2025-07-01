package new_therapist

import (
	"errors"
	"regexp"

	"github.com/mishkahtherapy/brain/core/domain"
	"github.com/mishkahtherapy/brain/core/ports"
)

var ErrFailedToCreateTherapist = errors.New("failed to create therapist")
var ErrFailedToGetTherapist = errors.New("failed to get therapist")
var ErrTherapistAlreadyExists = errors.New("therapist already exists")
var ErrEmailIsRequired = errors.New("email is required")
var ErrNameIsRequired = errors.New("name is required")
var ErrPhoneNumberIsRequired = errors.New("phone number is required")
var ErrWhatsAppNumberIsRequired = errors.New("whatsapp number is required")
var ErrInvalidPhoneNumber = errors.New("invalid phone number: must be in the format +1234567890")
var ErrInvalidWhatsAppNumber = errors.New("invalid whatsapp number: must be in the format +1234567890")
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
	therapist := &domain.Therapist{
		Name:           input.Name,
		Email:          input.Email,
		PhoneNumber:    input.PhoneNumber,
		WhatsAppNumber: input.WhatsAppNumber,
		SpeaksEnglish:  input.SpeaksEnglish,
	}

	specializations := make([]domain.Specialization, 0)
	for _, specID := range input.SpecializationIDs {
		specializations = append(specializations, domain.Specialization{ID: specID})
	}
	therapist.Specializations = specializations

	err := validateTherapist(therapist)
	if err != nil {
		return nil, err
	}

	err = validateSpecializations(u.specializationRepo, input.SpecializationIDs)
	if err != nil {
		return nil, err
	}

	existingTherapist, err := u.therapistRepo.GetByEmail(therapist.Email)
	if err != nil {
		return nil, err
	}
	if existingTherapist != nil {
		return nil, ErrTherapistAlreadyExists
	}

	existingTherapist, err = u.therapistRepo.GetByWhatsAppNumber(therapist.WhatsAppNumber)
	if err != nil {
		return nil, err
	}
	if existingTherapist != nil {
		return nil, ErrTherapistAlreadyExists
	}

	therapist.ID = domain.NewTherapistID()
	timestamp := domain.NewUTCTimestamp()
	therapist.CreatedAt = timestamp
	therapist.UpdatedAt = timestamp
	err = u.therapistRepo.Create(therapist)
	if err != nil {
		return nil, ErrFailedToCreateTherapist
	}
	return therapist, nil
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

func validateTherapist(therapist *domain.Therapist) error {
	if therapist.Email == domain.Email("") {
		return ErrEmailIsRequired
	}
	if therapist.Name == "" {
		return ErrNameIsRequired
	}
	if therapist.PhoneNumber == domain.PhoneNumber("") {
		return ErrPhoneNumberIsRequired
	}
	// Validate phone number
	if !isValidPhoneNumber(string(therapist.PhoneNumber)) {
		return ErrInvalidPhoneNumber
	}
	// Validate whatsapp number
	if therapist.WhatsAppNumber == domain.WhatsAppNumber("") {
		return ErrWhatsAppNumberIsRequired
	}
	if !isValidPhoneNumber(string(therapist.WhatsAppNumber)) {
		return ErrInvalidWhatsAppNumber
	}
	return nil
}

func isValidPhoneNumber(phoneNumber string) bool {
	re := regexp.MustCompile(`^\+?[1-9]\d{1,14}$`)
	return re.MatchString(phoneNumber)
}
