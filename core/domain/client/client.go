package client

import "github.com/mishkahtherapy/brain/core/domain"

type Client struct {
	ID             domain.ClientID       `json:"id"`
	Name           string                `json:"name"`
	WhatsAppNumber domain.WhatsAppNumber `json:"whatsAppNumber"`
	Timezone       string                `json:"timezone"` // Frontend hint for timezone adjustments
	BookingIDs     []domain.BookingID    `json:"bookingIds"`
	CreatedAt      domain.UTCTimestamp   `json:"createdAt"`
	UpdatedAt      domain.UTCTimestamp   `json:"updatedAt"`
}
