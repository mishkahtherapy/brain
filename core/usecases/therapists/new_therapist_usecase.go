package therapists

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

type Usecase struct {
	therapistRepo ports.TherapistRepository
}

func NewUsecase(therapistRepo ports.TherapistRepository) *Usecase {
	return &Usecase{therapistRepo: therapistRepo}
}

func (u *Usecase) Execute(therapist *domain.Therapist) (*domain.Therapist, error) {
	err := validateTherapist(therapist)
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

func validateTherapist(therapist *domain.Therapist) error {
	if therapist.Email == "" {
		return ErrEmailIsRequired
	}
	if therapist.Name == "" {
		return ErrNameIsRequired
	}
	if therapist.PhoneNumber == "" {
		return ErrPhoneNumberIsRequired
	}
	// Validate phone number
	if !isValidPhoneNumber(therapist.PhoneNumber) {
		return ErrInvalidPhoneNumber
	}
	// Validate whatsapp number
	if therapist.WhatsAppNumber == "" {
		return ErrWhatsAppNumberIsRequired
	}
	if !isValidPhoneNumber(therapist.WhatsAppNumber) {
		return ErrInvalidWhatsAppNumber
	}
	return nil
}

func isValidPhoneNumber(phoneNumber string) bool {
	re := regexp.MustCompile(`^\+?[1-9]\d{1,14}$`)
	return re.MatchString(phoneNumber)
}
