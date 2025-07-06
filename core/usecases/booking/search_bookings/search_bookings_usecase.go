package search_bookings

import (
	"time"

	"github.com/mishkahtherapy/brain/core/domain/booking"
	"github.com/mishkahtherapy/brain/core/ports"
	"github.com/mishkahtherapy/brain/core/usecases/common"
)

// Input represents the parameters accepted by the Search Bookings use-case.
// Start and End define the inclusive UTC time range to search within.
// When State is nil no filtering by booking state is applied.
// If provided, State must be one of the valid booking.BookingState constants.
// Validation is performed inside Execute.

type Input struct {
	Start time.Time
	End   time.Time
	State *booking.BookingState
}

type Usecase struct {
	bookingRepo ports.BookingRepository
}

func NewUsecase(bookingRepo ports.BookingRepository) *Usecase {
	return &Usecase{bookingRepo: bookingRepo}
}

func (u *Usecase) Execute(input Input) ([]*booking.Booking, error) {
	// Validate date range only if both dates are provided
	if !input.Start.IsZero() && !input.End.IsZero() && input.End.Before(input.Start) {
		return nil, common.ErrInvalidDateRange
	}

	// Delegate to repository
	bookings, err := u.bookingRepo.Search(input.Start, input.End, input.State)
	if err != nil {
		return nil, common.ErrFailedToListBookings
	}

	return bookings, nil
}
