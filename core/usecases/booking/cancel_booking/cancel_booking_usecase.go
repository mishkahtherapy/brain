package cancel_booking

import (
	"github.com/mishkahtherapy/brain/core/domain"
	"github.com/mishkahtherapy/brain/core/ports"
	"github.com/mishkahtherapy/brain/core/usecases/common"
)

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
		return nil, common.ErrBookingIDIsRequired
	}

	// Get existing booking
	booking, err := u.bookingRepo.GetByID(input.BookingID)
	if err != nil || booking == nil {
		return nil, common.ErrBookingNotFound
	}

	// Validate booking can be cancelled (not already cancelled)
	if booking.State == domain.BookingStateCancelled {
		return nil, common.ErrInvalidStateTransition
	}

	// Change state to Cancelled
	booking.State = domain.BookingStateCancelled
	booking.UpdatedAt = domain.NewUTCTimestamp()

	err = u.bookingRepo.Update(booking)
	if err != nil {
		return nil, common.ErrFailedToCancelBooking
	}

	return booking, nil
}
