package list_bookings_by_client

import (
	"github.com/mishkahtherapy/brain/core/domain"
	"github.com/mishkahtherapy/brain/core/ports"
	"github.com/mishkahtherapy/brain/core/usecases/common"
)

type Input struct {
	ClientID domain.ClientID      `json:"clientId"`
	State    *domain.BookingState `json:"state,omitempty"` // Optional state filter
}

type Usecase struct {
	bookingRepo ports.BookingRepository
}

func NewUsecase(bookingRepo ports.BookingRepository) *Usecase {
	return &Usecase{bookingRepo: bookingRepo}
}

func (u *Usecase) Execute(input Input) ([]*domain.Booking, error) {
	// Validate required fields
	if input.ClientID == "" {
		return nil, common.ErrClientIDIsRequired
	}

	var bookings []*domain.Booking
	var err error

	// If state filter is provided, use the specific method
	if input.State != nil {
		bookings, err = u.bookingRepo.ListByClientAndState(input.ClientID, *input.State)
	} else {
		bookings, err = u.bookingRepo.ListByClient(input.ClientID)
	}

	if err != nil {
		return nil, common.ErrFailedToListBookings
	}

	return bookings, nil
}
