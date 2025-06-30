package get_booking

import (
	"errors"

	"github.com/mishkahtherapy/brain/core/domain"
	"github.com/mishkahtherapy/brain/core/ports"
)

var ErrBookingNotFound = errors.New("booking not found")

type Usecase struct {
	bookingRepo ports.BookingRepository
}

func NewUsecase(bookingRepo ports.BookingRepository) *Usecase {
	return &Usecase{bookingRepo: bookingRepo}
}

func (u *Usecase) Execute(id domain.BookingID) (*domain.Booking, error) {
	booking, err := u.bookingRepo.GetByID(id)
	if err != nil || booking == nil {
		return nil, ErrBookingNotFound
	}
	return booking, nil
}
