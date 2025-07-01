package update_therapist_info

import (
	"time"

	"github.com/mishkahtherapy/brain/core/domain"
	"github.com/mishkahtherapy/brain/core/domain/therapist"
	"github.com/mishkahtherapy/brain/core/ports"
	therapistvalidation "github.com/mishkahtherapy/brain/core/usecases/therapist"
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

func (u *Usecase) Execute(input Input) (*therapist.Therapist, error) {
	// Validate therapist ID
	if input.TherapistID == "" {
		return nil, therapist.ErrTherapistIDRequired
	}

	// Validate required fields
	if err := therapistvalidation.ValidateRequiredFields(input.Name, input.Email, input.PhoneNumber, input.WhatsAppNumber); err != nil {
		return nil, err
	}

	// Validate phone number formats
	if err := therapistvalidation.ValidatePhoneNumbers(input.PhoneNumber, input.WhatsAppNumber); err != nil {
		return nil, err
	}

	// Get existing therapist
	existingTherapist, err := u.therapistRepo.GetByID(input.TherapistID)
	if err != nil {
		return nil, therapist.ErrTherapistNotFound
	}

	// Validate email and WhatsApp uniqueness for update
	if err := therapistvalidation.ValidateUniquenessForUpdate(u.therapistRepo, input.TherapistID, input.Email, input.WhatsAppNumber); err != nil {
		return nil, err
	}

	// Update therapist with new values
	updatedTherapist := &therapist.Therapist{
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
