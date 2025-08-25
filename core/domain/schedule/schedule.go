package schedule

import (
	"github.com/mishkahtherapy/brain/core/domain"
	"github.com/mishkahtherapy/brain/core/domain/specialization"
)

type TherapistInfo struct {
	TherapistID     domain.TherapistID              `json:"therapistId"`
	Name            string                          `json:"name"`
	Specializations []specialization.Specialization `json:"specializations"`
	SpeaksEnglish   bool                            `json:"speaksEnglish"`
	TimeSlotID      domain.TimeSlotID               `json:"timeSlotId"`
}

// I'm returning available "Time Ranges" not a ready made schedule to cater for timezone conversions on the frotnend.
type AvailableTimeRange struct {
	From       domain.UTCTimestamp    `json:"from"`       // Start of available range
	To         domain.UTCTimestamp    `json:"to"`         // End of available range
	Duration   domain.DurationMinutes `json:"duration"`   // Duration in minutes
	Therapists []TherapistInfo        `json:"therapists"` // List of therapists available in this time range
}
