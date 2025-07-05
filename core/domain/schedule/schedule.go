package schedule

import (
	"time"

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
	Date            time.Time       `json:"date"`            // The specific date for this availability
	StartTime       time.Time       `json:"startTime"`       // Start of available range
	EndTime         time.Time       `json:"endTime"`         // End of available range
	DurationMinutes int             `json:"durationMinutes"` // Duration in minutes
	Therapists      []TherapistInfo `json:"therapists"`      // List of therapists available in this time range
}
