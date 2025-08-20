package ports

import (
	"github.com/mishkahtherapy/brain/core/domain"
	"github.com/mishkahtherapy/brain/core/domain/therapist"
)

type TherapistRepository interface {
	GetByID(id domain.TherapistID) (*therapist.Therapist, error)
	GetByEmail(email domain.Email) (*therapist.Therapist, error)
	GetByWhatsAppNumber(whatsappNumber domain.WhatsAppNumber) (*therapist.Therapist, error)
	Create(therapist *therapist.Therapist) error
	Update(therapist *therapist.Therapist) error
	UpdateSpecializations(therapistID domain.TherapistID, specializationIDs []domain.SpecializationID) error
	UpdateDevice(therapistID domain.TherapistID, deviceID domain.DeviceID, deviceIDUpdatedAt domain.UTCTimestamp) error
	UpdateTimezoneOffset(therapistID domain.TherapistID, timezoneOffset domain.TimezoneOffset) error
	Delete(id domain.TherapistID) error
	List() ([]*therapist.Therapist, error)
	FindBySpecializationAndLanguage(specializationName string, mustSpeakEnglish bool) ([]*therapist.Therapist, error)
	FindByIDs(therapistIDs []domain.TherapistID) ([]*therapist.Therapist, error)
}
