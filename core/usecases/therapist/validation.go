package therapist

import (
	"regexp"

	"github.com/mishkahtherapy/brain/core/domain"
	"github.com/mishkahtherapy/brain/core/ports"
)

// ValidateRequiredFields validates that all required therapist fields are provided
func ValidateRequiredFields(name string, email domain.Email, phoneNumber domain.PhoneNumber, whatsAppNumber domain.WhatsAppNumber) error {
	if name == "" {
		return domain.ErrTherapistNameRequired
	}

	if email == domain.Email("") {
		return domain.ErrTherapistEmailRequired
	}

	if phoneNumber == domain.PhoneNumber("") {
		return domain.ErrTherapistPhoneRequired
	}

	if whatsAppNumber == domain.WhatsAppNumber("") {
		return domain.ErrTherapistWhatsAppRequired
	}

	return nil
}

// ValidatePhoneNumbers validates the format of phone and WhatsApp numbers
func ValidatePhoneNumbers(phoneNumber domain.PhoneNumber, whatsAppNumber domain.WhatsAppNumber) error {
	if !IsValidPhoneNumber(string(phoneNumber)) {
		return domain.ErrTherapistInvalidPhone
	}

	if !IsValidPhoneNumber(string(whatsAppNumber)) {
		return domain.ErrTherapistInvalidWhatsApp
	}

	return nil
}

// IsValidPhoneNumber validates a phone number format using international format
func IsValidPhoneNumber(phoneNumber string) bool {
	re := regexp.MustCompile(`^\+?[1-9]\d{1,14}$`)
	return re.MatchString(phoneNumber)
}

// ValidateEmailUniqueness checks if an email is already in use by another therapist
// skipTherapistID allows skipping a specific therapist (useful for updates)
func ValidateEmailUniqueness(repo ports.TherapistRepository, email domain.Email, skipTherapistID *domain.TherapistID) error {
	existingTherapist, err := repo.GetByEmail(email)
	if err != nil {
		// If error is "not found", email is available
		return nil
	}

	if existingTherapist != nil {
		// If we're updating and this is the same therapist, allow it
		if skipTherapistID != nil && existingTherapist.ID == *skipTherapistID {
			return nil
		}
		return domain.ErrTherapistEmailExists
	}

	return nil
}

// ValidateWhatsAppUniqueness checks if a WhatsApp number is already in use by another therapist
// skipTherapistID allows skipping a specific therapist (useful for updates)
func ValidateWhatsAppUniqueness(repo ports.TherapistRepository, whatsAppNumber domain.WhatsAppNumber, skipTherapistID *domain.TherapistID) error {
	existingTherapist, err := repo.GetByWhatsAppNumber(whatsAppNumber)
	if err != nil {
		// If error is "not found", WhatsApp is available
		return nil
	}

	if existingTherapist != nil {
		// If we're updating and this is the same therapist, allow it
		if skipTherapistID != nil && existingTherapist.ID == *skipTherapistID {
			return nil
		}
		return domain.ErrTherapistWhatsAppExists
	}

	return nil
}

// ValidateUniquenessForCreate validates email and WhatsApp uniqueness for creating a new therapist
func ValidateUniquenessForCreate(repo ports.TherapistRepository, email domain.Email, whatsAppNumber domain.WhatsAppNumber) error {
	if err := ValidateEmailUniqueness(repo, email, nil); err != nil {
		// Map the specific error to the general "already exists" for create operations
		if err == domain.ErrTherapistEmailExists {
			return domain.ErrTherapistAlreadyExists
		}
		return err
	}

	if err := ValidateWhatsAppUniqueness(repo, whatsAppNumber, nil); err != nil {
		// Map the specific error to the general "already exists" for create operations
		if err == domain.ErrTherapistWhatsAppExists {
			return domain.ErrTherapistAlreadyExists
		}
		return err
	}

	return nil
}

// ValidateUniquenessForUpdate validates email and WhatsApp uniqueness for updating an existing therapist
func ValidateUniquenessForUpdate(repo ports.TherapistRepository, therapistID domain.TherapistID, email domain.Email, whatsAppNumber domain.WhatsAppNumber) error {
	if err := ValidateEmailUniqueness(repo, email, &therapistID); err != nil {
		return err
	}

	if err := ValidateWhatsAppUniqueness(repo, whatsAppNumber, &therapistID); err != nil {
		return err
	}

	return nil
}
