package ports

import (
	"time"

	"github.com/mishkahtherapy/brain/core/domain"
	"github.com/mishkahtherapy/brain/core/domain/booking"
)

type BookingRepository interface {
	GetByID(id domain.BookingID) (*booking.Booking, error)
	Create(booking *booking.Booking) error
	Update(booking *booking.Booking) error
	Delete(id domain.BookingID) error
	ListByTherapist(therapistID domain.TherapistID) ([]*booking.Booking, error)
	ListByClient(clientID domain.ClientID) ([]*booking.Booking, error)
	ListByState(state booking.BookingState) ([]*booking.Booking, error)
	ListByTherapistAndState(therapistID domain.TherapistID, state booking.BookingState) ([]*booking.Booking, error)
	ListByClientAndState(clientID domain.ClientID, state booking.BookingState) ([]*booking.Booking, error)
	ListConfirmedByTherapistForDateRange(therapistID domain.TherapistID, startDate, endDate time.Time) ([]*booking.Booking, error)
	Search(startDate, endDate time.Time, state *booking.BookingState) ([]*booking.Booking, error)
}
