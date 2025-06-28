package domain

type Client struct {
	ID             ClientID     `json:"id"`
	Name           string       `json:"name"`
	Email          string       `json:"email"`
	WhatsAppNumber string       `json:"whatsAppNumber"`
	BookingIDs     []BookingID  `json:"bookingIds"`
	CreatedAt      UTCTimestamp `json:"createdAt"`
	UpdatedAt      UTCTimestamp `json:"updatedAt"`
}
