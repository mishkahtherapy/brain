package update_therapist_info

import (
	"errors"
	"regexp"
	"time"

	"github.com/mishkahtherapy/brain/core/domain"
	"github.com/mishkahtherapy/brain/core/ports"
)

// Business logic errors
var (
	ErrTherapistNotFound           = errors.New("therapist not found")
	ErrTherapistIDIsRequired       = errors.New("therapist ID is required")
	ErrNameIsRequired              = errors.New("name is required")
	ErrEmailIsRequired             = errors.New("email is required")
	ErrPhoneNumberIsRequired       = errors.New("phone number is required")
	ErrWhatsAppNumberIsRequired    = errors.New("whatsapp number is required")
	ErrInvalidEmail                = errors.New("invalid email format")
	ErrInvalidPhoneNumber          = errors.New("invalid phone number: must be in the format +1234567890")
	ErrInvalidWhatsAppNumber       = errors.New("invalid whatsapp number: must be in the format +1234567890")
	ErrEmailAlreadyExists          = errors.New("email already exists")
	ErrWhatsAppNumberAlreadyExists = errors.New("whatsapp number already exists")
)

type Input struct {
	TherapistID    domain.TherapistID    `json:"therapistId"`
	Name           string                `json:"name"`
	Email          domain.Email          `json:"email"`
	PhoneNumber    domain.PhoneNumber    `json:"phoneNumber"`
	WhatsAppNumber domain.WhatsAppNumber `json:"whatsAppNumber"`
	SpeaksEnglish  bool                  `json:"speaksEnglish"`
}

type Usecase struct {
	therapistRepo ports.TherapistRepository
}

func NewUsecase(therapistRepo ports.TherapistRepository) *Usecase {
	return &Usecase{
		therapistRepo: therapistRepo,
	}
}

func (u *Usecase) Execute(input Input) (*domain.Therapist, error) {
	// Validate input
	if err := u.validateInput(input); err != nil {
		return nil, err
	}

	// Get existing therapist
	existingTherapist, err := u.therapistRepo.GetByID(input.TherapistID)
	if err != nil {
		return nil, ErrTherapistNotFound
	}

	// Validate email uniqueness (if changed)
	if input.Email != existingTherapist.Email {
		emailTherapist, err := u.therapistRepo.GetByEmail(input.Email)
		if err == nil && emailTherapist != nil {
			return nil, ErrEmailAlreadyExists
		}
	}

	// Validate WhatsApp uniqueness (if changed)
	if input.WhatsAppNumber != existingTherapist.WhatsAppNumber {
		whatsappTherapist, err := u.therapistRepo.GetByWhatsAppNumber(input.WhatsAppNumber)
		if err == nil && whatsappTherapist != nil {
			return nil, ErrWhatsAppNumberAlreadyExists
		}
	}

	// Update therapist with new values
	updatedTherapist := &domain.Therapist{
		ID:              input.TherapistID,
		Name:            input.Name,
		Email:           input.Email,
		PhoneNumber:     input.PhoneNumber,
		WhatsAppNumber:  input.WhatsAppNumber,
		SpeaksEnglish:   input.SpeaksEnglish,
		Specializations: existingTherapist.Specializations, // Keep existing specializations
		CreatedAt:       existingTherapist.CreatedAt,       // Keep original creation time
		UpdatedAt:       domain.UTCTimestamp(time.Now().UTC()),
	}

	// Save updated therapist
	if err := u.therapistRepo.Update(updatedTherapist); err != nil {
		return nil, err
	}

	// Return updated therapist (fetch fresh from DB to ensure consistency)
	return u.therapistRepo.GetByID(input.TherapistID)
}

func (u *Usecase) validateInput(input Input) error {
	// Validate required fields
	if input.TherapistID == "" {
		return ErrTherapistIDIsRequired
	}

	if input.Name == "" {
		return ErrNameIsRequired
	}

	if input.Email == domain.Email("") {
		return ErrEmailIsRequired
	}

	if input.PhoneNumber == domain.PhoneNumber("") {
		return ErrPhoneNumberIsRequired
	}

	if input.WhatsAppNumber == domain.WhatsAppNumber("") {
		return ErrWhatsAppNumberIsRequired
	}

	// Validate phone number format
	if !isValidPhoneNumber(string(input.PhoneNumber)) {
		return ErrInvalidPhoneNumber
	}

	// Validate WhatsApp number format
	if !isValidPhoneNumber(string(input.WhatsAppNumber)) {
		return ErrInvalidWhatsAppNumber
	}

	return nil
}

func isValidPhoneNumber(phoneNumber string) bool {
	re := regexp.MustCompile(`^\+?[1-9]\d{1,14}$`)
	return re.MatchString(phoneNumber)
}
