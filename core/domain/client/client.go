package client

import "github.com/mishkahtherapy/brain/core/domain"

type Client struct {
	ID             domain.ClientID       `json:"id"`
	Name           string                `json:"name"`
	WhatsAppNumber domain.WhatsAppNumber `json:"whatsAppNumber"`
	TimezoneOffset domain.TimezoneOffset `json:"timezoneOffset"` // Frontend hint for timezone adjustments
	// Deprecated: kept temporarily for backward compatibility with old tests.
	Timezone   domain.Timezone     `json:"timezone,omitempty"`
	BookingIDs []domain.BookingID  `json:"bookingIds"`
	CreatedAt  domain.UTCTimestamp `json:"createdAt"`
	UpdatedAt  domain.UTCTimestamp `json:"updatedAt"`
}
