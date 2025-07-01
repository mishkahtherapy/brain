package create_booking

import (
	"time"

	"github.com/mishkahtherapy/brain/core/domain"
	"github.com/mishkahtherapy/brain/core/domain/booking"
	"github.com/mishkahtherapy/brain/core/ports"
	"github.com/mishkahtherapy/brain/core/usecases/common"
)

type Input struct {
	TherapistID domain.TherapistID  `json:"therapistId"`
	ClientID    domain.ClientID     `json:"clientId"`
	TimeSlotID  domain.TimeSlotID   `json:"timeSlotId"`
	StartTime   domain.UTCTimestamp `json:"startTime"`
}

type Usecase struct {
	bookingRepo   ports.BookingRepository
	therapistRepo ports.TherapistRepository
	clientRepo    ports.ClientRepository
	timeSlotRepo  ports.TimeSlotRepository
}

func NewUsecase(
	bookingRepo ports.BookingRepository,
	therapistRepo ports.TherapistRepository,
	clientRepo ports.ClientRepository,
	timeSlotRepo ports.TimeSlotRepository,
) *Usecase {
	return &Usecase{
		bookingRepo:   bookingRepo,
		therapistRepo: therapistRepo,
		clientRepo:    clientRepo,
		timeSlotRepo:  timeSlotRepo,
	}
}

func (u *Usecase) Execute(input Input) (*booking.Booking, error) {
	// Validate required fields
	if err := validateInput(input); err != nil {
		return nil, err
	}

	// Check if therapist exists
	therapist, err := u.therapistRepo.GetByID(input.TherapistID)
	if err != nil || therapist == nil {
		return nil, common.ErrTherapistNotFound
	}

	// Check if client exists
	client, err := u.clientRepo.GetByID(input.ClientID)
	if err != nil || client == nil {
		return nil, common.ErrClientNotFound
	}

	// Check if timeslot exists
	timeSlot, err := u.timeSlotRepo.GetByID(string(input.TimeSlotID))
	if err != nil || timeSlot == nil {
		return nil, common.ErrTimeSlotNotFound
	}

	// Check if timeslot is already booked (no existing confirmed/pending bookings)
	therapistBookings, err := u.bookingRepo.ListByTherapist(input.TherapistID)
	if err != nil {
		return nil, common.ErrFailedToCreateBooking
	}

	for _, existingBooking := range therapistBookings {
		if existingBooking.TimeSlotID == input.TimeSlotID &&
			(existingBooking.State == booking.BookingStateConfirmed || existingBooking.State == booking.BookingStatePending) {
			return nil, common.ErrTimeSlotAlreadyBooked
		}
	}

	// Create booking with Pending state
	now := domain.NewUTCTimestamp()
	createdBooking := &booking.Booking{
		ID:          domain.NewBookingID(),
		TherapistID: input.TherapistID,
		ClientID:    input.ClientID,
		TimeSlotID:  input.TimeSlotID,
		StartTime:   input.StartTime,
		State:       booking.BookingStatePending,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	err = u.bookingRepo.Create(createdBooking)
	if err != nil {
		return nil, common.ErrFailedToCreateBooking
	}

	return createdBooking, nil
}

func validateInput(input Input) error {
	if input.TherapistID == "" {
		return common.ErrTherapistIDIsRequired
	}
	if input.ClientID == "" {
		return common.ErrClientIDIsRequired
	}
	if input.TimeSlotID == "" {
		return common.ErrTimeSlotIDIsRequired
	}
	if time.Time(input.StartTime).IsZero() {
		return common.ErrStartTimeIsRequired
	}
	return nil
}
