package domain

type Client struct {
	ID             ClientID       `json:"id"`
	Name           string         `json:"name"`
	WhatsAppNumber WhatsAppNumber `json:"whatsAppNumber"`
	BookingIDs     []BookingID    `json:"bookingIds"`
	CreatedAt      UTCTimestamp   `json:"createdAt"`
	UpdatedAt      UTCTimestamp   `json:"updatedAt"`
}
