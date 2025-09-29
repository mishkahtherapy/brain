package booking

import (
	"fmt"
	"strings"

	"github.com/mishkahtherapy/brain/core/domain"
)

type BookingState string

const (
	BookingStatePending   BookingState = "pending"
	BookingStateConfirmed BookingState = "confirmed"
	BookingStateCancelled BookingState = "cancelled"
)

type BookingType int

const (
	BookingTypeRegular BookingType = 0
	BookingTypeAdhoc   BookingType = 1
)

func GetType(bookingID string) (BookingType, error) {
	if strings.HasPrefix(bookingID, "booking_") {
		return BookingTypeRegular, nil
	}
	if strings.HasPrefix(bookingID, "adhoc_booking_") {
		return BookingTypeAdhoc, nil
	}

	return 0, fmt.Errorf("invalid booking id: %s", bookingID)
}

type Booking struct {
	ID                   domain.BookingID       `json:"id"`
	TimeSlotID           domain.TimeSlotID      `json:"timeSlotId"`
	TherapistID          domain.TherapistID     `json:"therapistId"`
	ClientID             domain.ClientID        `json:"clientId"`
	State                BookingState           `json:"state"`
	StartTime            domain.UTCTimestamp    `json:"startTime"` // ISO 8601 datetime, e.g. "2024-06-01T09:00:00Z"
	Duration             domain.DurationMinutes `json:"duration"`
	ClientTimezoneOffset domain.TimezoneOffset  `json:"clientTimezoneOffset"` // Frontend hint for timezone adjustments. TODO: add an offset for therapist and an offset for patient
	CreatedAt            domain.UTCTimestamp    `json:"createdAt"`
	UpdatedAt            domain.UTCTimestamp    `json:"updatedAt"`
}

// AdhocBooking is a booking that is not associated with a time slot,
// and it doesn't intersect with any of the therapist's timeslots.
type AdhocBooking struct {
	ID                   domain.AdhocBookingID  `json:"id"`
	TherapistID          domain.TherapistID     `json:"therapistId"`
	ClientID             domain.ClientID        `json:"clientId"`
	State                BookingState           `json:"state"`
	StartTime            domain.UTCTimestamp    `json:"startTime"` // ISO 8601 datetime, e.g. "2024-06-01T09:00:00Z"
	Duration             domain.DurationMinutes `json:"duration"`
	ClientTimezoneOffset domain.TimezoneOffset  `json:"clientTimezoneOffset"` // Frontend hint for timezone adjustments. TODO: add an offset for therapist and an offset for patient
	CreatedAt            domain.UTCTimestamp    `json:"createdAt"`
	UpdatedAt            domain.UTCTimestamp    `json:"updatedAt"`
}
