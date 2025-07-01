package get_booking

import (
	"github.com/mishkahtherapy/brain/core/domain"
	"github.com/mishkahtherapy/brain/core/domain/booking"
	"github.com/mishkahtherapy/brain/core/ports"
	"github.com/mishkahtherapy/brain/core/usecases/common"
)

type Usecase struct {
	bookingRepo ports.BookingRepository
}

func NewUsecase(bookingRepo ports.BookingRepository) *Usecase {
	return &Usecase{bookingRepo: bookingRepo}
}

func (u *Usecase) Execute(id domain.BookingID) (*booking.Booking, error) {
	existingBooking, err := u.bookingRepo.GetByID(id)
	if err != nil || existingBooking == nil {
		return nil, common.ErrBookingNotFound
	}
	return existingBooking, nil
}
