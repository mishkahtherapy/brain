package domain

type Therapist struct {
	ID             TherapistID    `json:"id"`
	Name           string         `json:"name"`
	Email          Email          `json:"email"`
	PhoneNumber    PhoneNumber    `json:"phoneNumber"`
	WhatsAppNumber WhatsAppNumber `json:"whatsAppNumber"`

	// TODO: don't return IDs, return the time slots
	TimeSlots []TimeSlot `json:"timeSlots"`

	// TODO: don't return IDs, return the specializations
	Specializations []Specialization `json:"specializations"`

	CreatedAt UTCTimestamp `json:"createdAt"`
	UpdatedAt UTCTimestamp `json:"updatedAt"`
}
