package common

import "errors"

// Entity Not Found Errors - Centralized to avoid duplication
var (
	ErrBookingNotFound        = errors.New("booking not found")
	ErrTherapistNotFound      = errors.New("therapist not found")
	ErrClientNotFound         = errors.New("client not found")
	ErrSessionNotFound        = errors.New("session not found")
	ErrTimeSlotNotFound       = errors.New("timeslot not found")
	ErrSpecializationNotFound = errors.New("specialization not found")
)

// Operation Failed Errors - Centralized patterns
var (
	ErrFailedToCreateBooking  = errors.New("failed to create booking")
	ErrFailedToCancelBooking  = errors.New("failed to cancel booking")
	ErrFailedToConfirmBooking = errors.New("failed to confirm booking")
	ErrFailedToListBookings   = errors.New("failed to list bookings")

	ErrFailedToCreateSession      = errors.New("failed to create session")
	ErrFailedToListSessions       = errors.New("failed to list sessions")
	ErrFailedToUpdateSession      = errors.New("failed to update session")
	ErrFailedToUpdateSessionState = errors.New("failed to update session state")
	ErrFailedToUpdateSessionNotes = errors.New("failed to update session notes")
	ErrFailedToUpdateMeetingURL   = errors.New("failed to update meeting URL")

	ErrFailedToCreateTherapist = errors.New("failed to create therapist")
	ErrFailedToUpdateTherapist = errors.New("failed to update therapist")

	ErrFailedToCreateSpecialization = errors.New("failed to create specialization")
	ErrFailedToGetSpecializations   = errors.New("failed to get specializations")
)

// Business Logic Errors
var (
	ErrInvalidStateTransition = errors.New("invalid state transition")
	ErrInvalidBookingState    = errors.New("booking must be in pending state to be confirmed")
	ErrTimeSlotAlreadyBooked  = errors.New("timeslot is already booked")
	ErrMeetingURLNotSet       = errors.New("meeting URL is not set for this session")
	ErrInvalidMeetingURL      = errors.New("invalid meeting URL format")
)

// Common validation errors that appear in multiple usecases
var (
	ErrInvalidDateRange = errors.New("invalid date range")
)

// Required Field Errors - ID validations
var (
	ErrBookingIDIsRequired        = errors.New("booking ID is required")
	ErrTherapistIDIsRequired      = errors.New("therapist ID is required")
	ErrClientIDIsRequired         = errors.New("client ID is required")
	ErrSessionIDIsRequired        = errors.New("session ID is required")
	ErrTimeSlotIDIsRequired       = errors.New("timeslot ID is required")
	ErrSpecializationIDIsRequired = errors.New("specialization ID is required")
)

// Required Field Errors - Other common validations
var (
	ErrStartTimeIsRequired            = errors.New("start time is required")
	ErrDurationIsRequired             = errors.New("duration is required")
	ErrClientTimezoneOffsetIsRequired = errors.New("client timezone offset is required")
	ErrPaidAmountIsRequired           = errors.New("paid amount is required")
	ErrLanguageIsRequired             = errors.New("language is required")
	ErrStateIsRequired                = errors.New("state is required")
	ErrNotesIsRequired                = errors.New("notes is required")
	ErrMeetingURLIsRequired           = errors.New("meeting URL is required")
	ErrNameIsRequired                 = errors.New("name is required")
)
