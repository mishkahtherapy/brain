package domain

type Client struct {
	ID             string   `json:"id"`
	Name           string   `json:"name"`
	Email          string   `json:"email"`
	WhatsAppNumber string   `json:"whatsAppNumber"`
	BookingIDs     []string `json:"bookingIds"`
}
