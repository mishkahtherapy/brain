package therapist

import (
	"github.com/mishkahtherapy/brain/core/domain"
	"github.com/mishkahtherapy/brain/core/domain/specialization"
)

type Therapist struct {
	ID              domain.TherapistID              `json:"id"`
	Name            string                          `json:"name"`
	Email           domain.Email                    `json:"email"`
	PhoneNumber     domain.PhoneNumber              `json:"phoneNumber"`
	WhatsAppNumber  domain.WhatsAppNumber           `json:"whatsAppNumber"`
	SpeaksEnglish   bool                            `json:"speaksEnglish"`
	DeviceID        domain.DeviceID                 `json:"-"` // Not exposed to client
	Specializations []specialization.Specialization `json:"specializations"`
	TimezoneOffset  domain.TimezoneOffset           `json:"timezoneOffset"`

	CreatedAt domain.UTCTimestamp `json:"createdAt"`
	UpdatedAt domain.UTCTimestamp `json:"updatedAt"`
}
