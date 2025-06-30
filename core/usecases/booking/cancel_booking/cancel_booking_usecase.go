package cancel_booking

import (
	"errors"

	"github.com/mishkahtherapy/brain/core/domain"
	"github.com/mishkahtherapy/brain/core/ports"
)

var ErrFailedToCancelBooking = errors.New("failed to cancel booking")
var ErrBookingIDIsRequired = errors.New("booking ID is required")
var ErrBookingNotFound = errors.New("booking not found")
var ErrInvalidStateTransition = errors.New("booking cannot be cancelled from current state")

type Input struct {
	BookingID domain.BookingID `json:"bookingId"`
}

type Usecase struct {
	bookingRepo ports.BookingRepository
}

func NewUsecase(bookingRepo ports.BookingRepository) *Usecase {
	return &Usecase{bookingRepo: bookingRepo}
}

func (u *Usecase) Execute(input Input) (*domain.Booking, error) {
	// Validate required fields
	if input.BookingID == "" {
		return nil, ErrBookingIDIsRequired
	}

	// Get existing booking
	booking, err := u.bookingRepo.GetByID(input.BookingID)
	if err != nil || booking == nil {
		return nil, ErrBookingNotFound
	}

	// Validate booking can be cancelled (not already cancelled)
	if booking.State == domain.BookingStateCancelled {
		return nil, ErrInvalidStateTransition
	}

	// Change state to Cancelled
	booking.State = domain.BookingStateCancelled
	booking.UpdatedAt = domain.NewUTCTimestamp()

	err = u.bookingRepo.Update(booking)
	if err != nil {
		return nil, ErrFailedToCancelBooking
	}

	return booking, nil
}
