package domain

import (
	"errors"
	"time"
)

var (
	ErrTimezoneIsRequired = errors.New("timezone is required")
	ErrInvalidTimezone    = errors.New("invalid timezone format")
)

// Timezone represents an IANA timezone identifier
// Used for validation only - no timezone conversions happen on the backend
type Timezone string

// IsValid checks if the timezone string is a valid IANA timezone identifier
func (tz Timezone) IsValid() bool {
	_, err := time.LoadLocation(string(tz))
	return err == nil
}

// ToLocation converts the timezone string to a time.Location (for validation purposes)
func (tz Timezone) ToLocation() (*time.Location, error) {
	return time.LoadLocation(string(tz))
}
