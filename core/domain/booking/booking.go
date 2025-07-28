package booking

import "github.com/mishkahtherapy/brain/core/domain"

type BookingState string

const (
	BookingStatePending   BookingState = "pending"
	BookingStateConfirmed BookingState = "confirmed"
	BookingStateCancelled BookingState = "cancelled"
)

type Booking struct {
	ID          domain.BookingID       `json:"id"`
	TimeSlotID  domain.TimeSlotID      `json:"timeSlotId"`
	TherapistID domain.TherapistID     `json:"therapistId"`
	ClientID    domain.ClientID        `json:"clientId"`
	State       BookingState           `json:"state"`
	StartTime   domain.UTCTimestamp    `json:"startTime"` // ISO 8601 datetime, e.g. "2024-06-01T09:00:00Z"
	Duration    domain.DurationMinutes `json:"duration"`
	// TimezoneOffset domain.TimezoneOffset  `json:"timezoneOffset"` // Frontend hint for timezone adjustments. TODO: add an offset for therapist and an offset for patient
	CreatedAt domain.UTCTimestamp `json:"createdAt"`
	UpdatedAt domain.UTCTimestamp `json:"updatedAt"`
	// Duration is always 1 hour
}
