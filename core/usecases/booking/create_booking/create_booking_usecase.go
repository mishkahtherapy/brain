package create_booking

import (
	"errors"
	"time"

	"github.com/mishkahtherapy/brain/core/domain"
	"github.com/mishkahtherapy/brain/core/ports"
)

var ErrFailedToCreateBooking = errors.New("failed to create booking")
var ErrTherapistIDIsRequired = errors.New("therapist ID is required")
var ErrClientIDIsRequired = errors.New("client ID is required")
var ErrTimeSlotIDIsRequired = errors.New("timeslot ID is required")
var ErrStartTimeIsRequired = errors.New("start time is required")
var ErrTherapistNotFound = errors.New("therapist not found")
var ErrClientNotFound = errors.New("client not found")
var ErrTimeSlotNotFound = errors.New("timeslot not found")
var ErrTimeSlotAlreadyBooked = errors.New("timeslot is already booked")

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

func (u *Usecase) Execute(input Input) (*domain.Booking, error) {
	// Validate required fields
	if err := validateInput(input); err != nil {
		return nil, err
	}

	// Check if therapist exists
	therapist, err := u.therapistRepo.GetByID(input.TherapistID)
	if err != nil || therapist == nil {
		return nil, ErrTherapistNotFound
	}

	// Check if client exists
	client, err := u.clientRepo.GetByID(string(input.ClientID))
	if err != nil || client == nil {
		return nil, ErrClientNotFound
	}

	// Check if timeslot exists
	timeSlot, err := u.timeSlotRepo.GetByID(string(input.TimeSlotID))
	if err != nil || timeSlot == nil {
		return nil, ErrTimeSlotNotFound
	}

	// Check if timeslot is already booked (no existing confirmed/pending bookings)
	therapistBookings, err := u.bookingRepo.ListByTherapist(input.TherapistID)
	if err != nil {
		return nil, ErrFailedToCreateBooking
	}

	for _, booking := range therapistBookings {
		if booking.TimeSlotID == input.TimeSlotID &&
			(booking.State == domain.BookingStateConfirmed || booking.State == domain.BookingStatePending) {
			return nil, ErrTimeSlotAlreadyBooked
		}
	}

	// Create booking with Pending state
	now := domain.NewUTCTimestamp()
	booking := &domain.Booking{
		ID:          domain.NewBookingID(),
		TherapistID: input.TherapistID,
		ClientID:    input.ClientID,
		TimeSlotID:  input.TimeSlotID,
		StartTime:   input.StartTime,
		State:       domain.BookingStatePending,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	err = u.bookingRepo.Create(booking)
	if err != nil {
		return nil, ErrFailedToCreateBooking
	}

	return booking, nil
}

func validateInput(input Input) error {
	if input.TherapistID == "" {
		return ErrTherapistIDIsRequired
	}
	if input.ClientID == "" {
		return ErrClientIDIsRequired
	}
	if input.TimeSlotID == "" {
		return ErrTimeSlotIDIsRequired
	}
	if time.Time(input.StartTime).IsZero() {
		return ErrStartTimeIsRequired
	}
	return nil
}
