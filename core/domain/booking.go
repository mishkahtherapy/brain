package domain

type Booking struct {
	ID          string `json:"id"`
	TimeSlotID  string `json:"timeSlotId"`
	TherapistID string `json:"therapistId"`
	ClientID    string `json:"clientId"`
	Date        string `json:"date"` // ISO 8601 date, e.g. "2024-06-01"
}
