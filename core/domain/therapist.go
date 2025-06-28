package domain

type Therapist struct {
	ID             TherapistID  `json:"id"`
	Name           string       `json:"name"`
	Email          string       `json:"email"`
	PhoneNumber    string       `json:"phoneNumber"`
	WhatsAppNumber string       `json:"whatsAppNumber"`
	TimeSlotIDs    []TimeSlotID `json:"timeSlotIds"`
	BookingIDs     []BookingID  `json:"bookingIds"`

	SpecializationIDs []SpecializationID `json:"specializationIds"`

	CreatedAt UTCTimestamp `json:"createdAt"`
	UpdatedAt UTCTimestamp `json:"updatedAt"`
}
