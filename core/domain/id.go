package domain

import (
	"fmt"
	"strings"

	"github.com/google/uuid"
)

type ClientID string
type TherapistID string
type DeviceID string
type TimeSlotID string
type BookingID string
type SessionID string
type SpecializationID string

func NewClientID() ClientID {
	return ClientID(generatePrefixedUUID("client"))
}

func NewTherapistID() TherapistID {
	return TherapistID(generatePrefixedUUID("therapist"))
}

func NewSpecializationID() SpecializationID {
	return SpecializationID(generatePrefixedUUID("specialization"))
}

func NewBookingID() BookingID {
	return BookingID(generatePrefixedUUID("booking"))
}

func NewSessionID() SessionID {
	return SessionID(generatePrefixedUUID("session"))
}

func NewTimeSlotID() TimeSlotID {
	return TimeSlotID(generatePrefixedUUID("timeslot"))
}

func generatePrefixedUUID(prefix string) string {
	return fmt.Sprintf("%s_%s", prefix, strings.ReplaceAll(uuid.NewString(), "-", ""))
}
