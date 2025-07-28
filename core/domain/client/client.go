package client

import (
	"github.com/mishkahtherapy/brain/core/domain"
	"github.com/mishkahtherapy/brain/core/domain/booking"
)

type Client struct {
	ID             domain.ClientID       `json:"id"`
	Name           string                `json:"name"`
	WhatsAppNumber domain.WhatsAppNumber `json:"whatsAppNumber"`
	TimezoneOffset domain.TimezoneOffset `json:"timezoneOffset"` // Frontend hint for timezone adjustments
	Bookings       []booking.Booking     `json:"bookings"`
	CreatedAt      domain.UTCTimestamp   `json:"createdAt"`
	UpdatedAt      domain.UTCTimestamp   `json:"updatedAt"`

	// TODO: Add chat messages (?) / roomId (?)
}
