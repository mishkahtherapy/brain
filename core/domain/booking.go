package domain

type Booking struct {
	ID          BookingID    `json:"id"`
	TimeSlotID  TimeSlotID   `json:"timeSlotId"`
	TherapistID TherapistID  `json:"therapistId"`
	ClientID    ClientID     `json:"clientId"`
	StartTime   UTCTimestamp `json:"startTime"` // ISO 8601 datetime, e.g. "2024-06-01T09:00:00Z"
	CreatedAt   UTCTimestamp `json:"createdAt"`
	UpdatedAt   UTCTimestamp `json:"updatedAt"`
	// Duration is always 1 hour
}
