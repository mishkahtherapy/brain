package list_bookings_by_therapist

import (
	"github.com/mishkahtherapy/brain/core/domain"
	"github.com/mishkahtherapy/brain/core/ports"
	"github.com/mishkahtherapy/brain/core/usecases/common"
)

type Input struct {
	TherapistID domain.TherapistID   `json:"therapistId"`
	State       *domain.BookingState `json:"state,omitempty"` // Optional state filter
}

type Usecase struct {
	bookingRepo ports.BookingRepository
}

func NewUsecase(bookingRepo ports.BookingRepository) *Usecase {
	return &Usecase{bookingRepo: bookingRepo}
}

func (u *Usecase) Execute(input Input) ([]*domain.Booking, error) {
	// Validate required fields
	if input.TherapistID == "" {
		return nil, common.ErrTherapistIDIsRequired
	}

	var bookings []*domain.Booking
	var err error

	// If state filter is provided, use the specific method
	if input.State != nil {
		bookings, err = u.bookingRepo.ListByTherapistAndState(input.TherapistID, *input.State)
	} else {
		bookings, err = u.bookingRepo.ListByTherapist(input.TherapistID)
	}

	if err != nil {
		return nil, common.ErrFailedToListBookings
	}

	return bookings, nil
}
