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
var ErrInvalidDateRange = errors.New("invalid date range")

type BookingRepository interface {
	GetByID(id domain.BookingID) (*booking.Booking, error)
	Create(booking *booking.Booking) error
	Update(booking *booking.Booking) error
	UpdateTx(sqlExec SQLExec, booking *booking.Booking) error
	Delete(id domain.BookingID) error
	ListByTherapist(therapistID domain.TherapistID) ([]*booking.Booking, error)
	ListByClient(clientID domain.ClientID) ([]*booking.Booking, error)
	ListByState(state booking.BookingState) ([]*booking.Booking, error)
	ListByTherapistAndState(therapistID domain.TherapistID, state booking.BookingState) ([]*booking.Booking, error)
	ListByClientAndState(clientID domain.ClientID, state booking.BookingState) ([]*booking.Booking, error)
	BulkListByTherapistForDateRange(
		therapistIDs []domain.TherapistID,
		states []booking.BookingState,
		startDate, endDate time.Time,
	) (map[domain.TherapistID][]*booking.Booking, error)
	BulkCancel(tx SQLTx, bookingIDs []domain.BookingID) error
	Search(startDate, endDate time.Time, states []booking.BookingState) ([]*booking.Booking, error)
}
