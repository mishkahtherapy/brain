package therapist

import "errors"

// Therapist domain errors - business rule violations
var (
	ErrTherapistNotFound         = errors.New("therapist not found")
	ErrTherapistAlreadyExists    = errors.New("therapist already exists")
	ErrTherapistNameRequired     = errors.New("name is required")
	ErrTherapistEmailRequired    = errors.New("email is required")
	ErrTherapistPhoneRequired    = errors.New("phone number is required")
	ErrTherapistWhatsAppRequired = errors.New("whatsapp number is required")
	ErrTherapistInvalidPhone     = errors.New("invalid phone number: must be in the format +1234567890")
	ErrTherapistInvalidWhatsApp  = errors.New("invalid whatsapp number: must be in the format +1234567890")
	ErrTherapistEmailExists      = errors.New("email already exists")
	ErrTherapistWhatsAppExists   = errors.New("whatsapp number already exists")
	ErrTherapistIDRequired       = errors.New("therapist ID is required")
)
