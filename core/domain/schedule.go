package domain

import "time"

type TherapistInfo struct {
	ID              TherapistID      `json:"id"`
	Name            string           `json:"name"`
	Specializations []Specialization `json:"specializations"`
	SpeaksEnglish   bool             `json:"speaksEnglish"`
}

type AvailableTimeRange struct {
	Date            time.Time       `json:"date"`            // The specific date for this availability
	StartTime       time.Time       `json:"startTime"`       // Start of available range
	EndTime         time.Time       `json:"endTime"`         // End of available range
	DurationMinutes int             `json:"durationMinutes"` // Duration in minutes
	Therapists      []TherapistInfo `json:"therapists"`      // List of therapists available in this time range
}

type ScheduleResponse struct {
	Availabilities []AvailableTimeRange `json:"availabilities"`
}
