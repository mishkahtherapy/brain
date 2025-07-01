package timeslot

import "errors"

// Common error definitions for all timeslot use cases
var (
	// Input validation errors
	ErrTherapistIDIsRequired = errors.New("therapist id is required")
	ErrTimeslotIDIsRequired  = errors.New("timeslot id is required")
	ErrDayOfWeekIsRequired   = errors.New("day of week is required")
	ErrStartTimeIsRequired   = errors.New("start time is required")
	ErrEndTimeIsRequired     = errors.New("end time is required")

	// Business logic errors
	ErrTherapistNotFound        = errors.New("therapist not found")
	ErrTimeslotNotFound         = errors.New("timeslot not found")
	ErrTimeslotNotOwned         = errors.New("timeslot does not belong to this therapist")
	ErrInvalidDayOfWeek         = errors.New("invalid day of week")
	ErrInvalidTimeFormat        = errors.New("invalid time format, use HH:MM")
	ErrInvalidTimeRange         = errors.New("start time must be before end time")
	ErrPreSessionBufferNegative = errors.New("pre-session buffer cannot be negative")
	ErrPostSessionBufferTooLow  = errors.New("post-session buffer must be at least 30 minutes")
	ErrOverlappingTimeslot      = errors.New("timeslot overlaps with existing timeslot for this therapist")

	// Deletion constraints
	ErrTimeslotHasActiveBookings = errors.New("cannot delete timeslot with active bookings")
)
