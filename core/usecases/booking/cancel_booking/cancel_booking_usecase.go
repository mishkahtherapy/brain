package cancel_booking

import (
	"github.com/mishkahtherapy/brain/core/domain"
	"github.com/mishkahtherapy/brain/core/domain/booking"
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

func (u *Usecase) Execute(input Input) (*ports.BookingResponse, error) {
	// Validate required fields
	if input.BookingID == "" {
		return nil, common.ErrBookingIDIsRequired
	}

	// Get existing booking
	existingBooking, err := u.bookingRepo.GetByID(input.BookingID)
	if err != nil || existingBooking == nil {
		return nil, common.ErrBookingNotFound
	}

	// Validate booking can be cancelled (not already cancelled)
	if existingBooking.State == booking.BookingStateCancelled {
		return nil, common.ErrInvalidStateTransition
	}

	// Change state to Cancelled
	updatedAt := domain.NewUTCTimestamp().Time()
	err = u.bookingRepo.UpdateState(
		existingBooking.ID,
		booking.BookingStateCancelled,
		updatedAt,
	)
	if err != nil {
		return nil, common.ErrFailedToCancelBooking
	}

	return &ports.BookingResponse{
		RegularBookingID:     existingBooking.ID,
		TherapistID:          existingBooking.TherapistID,
		ClientID:             existingBooking.ClientID,
		State:                existingBooking.State,
		StartTime:            existingBooking.StartTime,
		Duration:             existingBooking.Duration,
		ClientTimezoneOffset: existingBooking.ClientTimezoneOffset,
	}, nil
}
