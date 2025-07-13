package schedule

import (
	"github.com/mishkahtherapy/brain/core/domain"
	"github.com/mishkahtherapy/brain/core/domain/specialization"
)

type TherapistInfo struct {
	ID              domain.TherapistID              `json:"id"`
	Name            string                          `json:"name"`
	Specializations []specialization.Specialization `json:"specializations"`
	SpeaksEnglish   bool                            `json:"speaksEnglish"`
	TimeSlotID      domain.TimeSlotID               `json:"timeSlotId"`
}

type AvailableTimeRange struct {
	DayOfWeek       string              `json:"dayOfWeek"`       // The day of the week for this availability
	From            domain.UTCTimestamp `json:"from"`            // Start of available range
	To              domain.UTCTimestamp `json:"to"`              // End of available range
	DurationMinutes int                 `json:"durationMinutes"` // Duration in minutes
	Therapists      []TherapistInfo     `json:"therapists"`      // List of therapists available in this time range
}
