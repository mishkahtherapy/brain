package domain

import (
	"fmt"

	"github.com/google/uuid"
)

type ClientID string
type TherapistID string
type TimeSlotID string
type BookingID string
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

func generatePrefixedUUID(prefix string) string {
	return fmt.Sprintf("%s_%s", prefix, uuid.NewString())
}
