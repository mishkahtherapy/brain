package ports

import (
	"errors"
	"time"

	"github.com/mishkahtherapy/brain/core/domain"
	"github.com/mishkahtherapy/brain/core/domain/booking"
)

var ErrBookingNotFound = errors.New("booking not found")
var ErrBookingAlreadyExists = errors.New("booking already exists")
var ErrBookingIDIsRequired = errors.New("booking id is required")
var ErrBookingTimeSlotIDIsRequired = errors.New("booking timeslot id is required")
var ErrBookingTherapistIDIsRequired = errors.New("booking therapist id is required")
var ErrBookingClientIDIsRequired = errors.New("booking client id is required")
var ErrBookingStateIsRequired = errors.New("booking state is required")
var ErrBookingStartTimeIsRequired = errors.New("booking start time is required")
var ErrBookingCreatedAtIsRequired = errors.New("booking created at is required")
var ErrBookingUpdatedAtIsRequired = errors.New("booking updated at is required")
var ErrBookingDurationIsRequired = errors.New("booking duration is required")
var ErrFailedToGetBookings = errors.New("failed to get bookings")
var ErrFailedToCreateBooking = errors.New("failed to create booking")
var ErrFailedToUpdateBooking = errors.New("failed to update booking")
var ErrFailedToDeleteBooking = errors.New("failed to delete booking")
var ErrInvalidBookingFilters = errors.New("invalid booking filters")
var ErrInvalidDateRange = errors.New("invalid date range")

type BookingFilters struct {
	TherapistID domain.TherapistID
	ClientID    domain.ClientID
	State       booking.BookingState
}

func (f *BookingFilters) IsValid() bool {
	if f.ClientID != "" {
		return true
	}
	if f.TherapistID != "" {
		return true
	}
	if f.State != "" {
		return true
	}
	return false
}

type BookingRepository interface {
	GetByID(id domain.BookingID) (*booking.Booking, error)
	Create(booking *booking.Booking) error
	UpdateState(bookingID domain.BookingID, state booking.BookingState, updatedAt time.Time) error
	UpdateStateTx(sqlExec SQLExec, bookingID domain.BookingID, state booking.BookingState, updatedAt time.Time) error
	Delete(id domain.BookingID) error
	List(filters BookingFilters) ([]*booking.Booking, error)
	ListByTherapistForDateRange(
		therapistID domain.TherapistID,
		states []booking.BookingState,
		startDate, endDate time.Time,
	) ([]*booking.Booking, error)
	BulkListByTherapistForDateRange(
		therapistIDs []domain.TherapistID,
		states []booking.BookingState,
		startDate, endDate time.Time,
	) (map[domain.TherapistID][]*booking.Booking, error)
	BulkCancel(tx SQLTx, bookingIDs []domain.BookingID) error
	Search(startDate, endDate time.Time, states []booking.BookingState) ([]*booking.Booking, error)
}

type BookingResponse struct {
	RegularBookingID     domain.BookingID       `json:"regularBookingId,omitempty"`
	AdhocBookingID       domain.AdhocBookingID  `json:"adhocBookingId,omitempty"`
	TherapistID          domain.TherapistID     `json:"therapistId"`
	ClientID             domain.ClientID        `json:"clientId"`
	State                booking.BookingState   `json:"state"`
	StartTime            domain.UTCTimestamp    `json:"startTime"` // ISO 8601 datetime, e.g. "2024-06-01T09:00:00Z"
	Duration             domain.DurationMinutes `json:"duration"`
	ClientTimezoneOffset domain.TimezoneOffset  `json:"clientTimezoneOffset"` // Frontend hint for timezone adjustments. TODO: add an offset for therapist and an offset for patient
}
