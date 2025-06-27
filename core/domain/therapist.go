package domain

type Therapist struct {
	ID             string   `json:"id"`
	Name           string   `json:"name"`
	Email          string   `json:"email"`
	PhoneNumber    string   `json:"phoneNumber"`
	WhatsAppNumber string   `json:"whatsAppNumber"`
	TimeSlotIDs    []string `json:"timeSlotIds"`
	BookingIDs     []string `json:"bookingIds"`
}
