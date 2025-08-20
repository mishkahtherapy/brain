package create_booking

import (
	"errors"
	"log/slog"
	"time"

	"github.com/mishkahtherapy/brain/core/domain"
	"github.com/mishkahtherapy/brain/core/domain/booking"
	"github.com/mishkahtherapy/brain/core/domain/timeslot"
	"github.com/mishkahtherapy/brain/core/ports"
	"github.com/mishkahtherapy/brain/core/usecases/common"
)

var (
	ErrTimezoneIsRequired = errors.New("timezone is required")
	ErrInvalidTimezone    = errors.New("invalid timezone format")
)

type Input struct {
	TherapistID domain.TherapistID     `json:"therapistId"`
	ClientID    domain.ClientID        `json:"clientId"`
	TimeSlotID  domain.TimeSlotID      `json:"timeSlotId"`
	StartTime   domain.UTCTimestamp    `json:"startTime"`
	Duration    domain.DurationMinutes `json:"duration"`
	// TimezoneOffset domain.TimezoneOffset `json:"timezoneOffset"`
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
	client, err := u.clientRepo.FindByIDs([]domain.ClientID{input.ClientID})
	if err != nil || client == nil {
		return nil, common.ErrClientNotFound
	}

	// Check if timeslot exists
	timeSlot, err := u.timeSlotRepo.GetByID(input.TimeSlotID)
	if err != nil || timeSlot == nil {
		return nil, common.ErrTimeSlotNotFound
	}

	// Fetch all timeslots for the therapist once to avoid repeated DB hits in the overlap check loop.
	therapistTimeSlots, err := u.timeSlotRepo.ListByTherapist(input.TherapistID)
	if err != nil {
		return nil, common.ErrFailedToCreateBooking
	}

	// Build a quick lookup map: timeSlotID ➜ TimeSlot
	therapistTimeSlotMap := make(map[domain.TimeSlotID]*timeslot.TimeSlot, len(therapistTimeSlots))
	for _, ts := range therapistTimeSlots {
		therapistTimeSlotMap[ts.ID] = ts
	}

	// Gather existing bookings for therapist
	therapistBookings, err := u.bookingRepo.ListByTherapist(input.TherapistID)
	if err != nil {
		return nil, common.ErrFailedToCreateBooking
	}

	// Compute the occupied interval for the *new* booking. Note that the new booking's
	// own post-session buffer does NOT affect conflict detection – only the existing
	// booking's buffer matters as per business rules.
	newBookingStart := time.Time(input.StartTime)
	newBookingEnd := newBookingStart.Add(time.Duration(timeSlot.Duration) * time.Minute)

	for _, existingBooking := range therapistBookings {
		if existingBooking.State != booking.BookingStateConfirmed {
			continue
		}

		existingTimeSlot, ok := therapistTimeSlotMap[existingBooking.TimeSlotID]
		if !ok {
			slog.Error("timeslot not found", "existing_booking_id", existingBooking.ID, "existing_booking_time_slot_id", existingBooking.TimeSlotID)
			// Timeslot not found; conservative conflict
			return nil, common.ErrTimeSlotAlreadyBooked
		}

		// Occupied interval of the *existing* booking, extended forward by its
		// post-session buffer.
		existingStartTime := time.Time(existingBooking.StartTime)
		existingEndTime := existingStartTime.Add(time.Duration(existingTimeSlot.Duration) * time.Minute)
		existingEndWithBuffer := existingEndTime.Add(time.Duration(existingTimeSlot.PostSessionBuffer) * time.Minute)

		// Ranges overlap if:
		//   newStart < existingEndWithBuffer  &&  newEnd > existingStart
		// This captures full and partial overlaps, including intrusion into the
		// existing booking's post-session buffer.
		if newBookingStart.Before(existingEndWithBuffer) && newBookingEnd.After(existingStartTime) {
			return nil, common.ErrTimeSlotAlreadyBooked
		}
	}

	// Create booking with Pending state and timezone (no conversion, just store as hint)
	now := domain.NewUTCTimestamp()
	createdBooking := &booking.Booking{
		ID:          domain.NewBookingID(),
		TherapistID: input.TherapistID,
		ClientID:    input.ClientID,
		TimeSlotID:  input.TimeSlotID,
		StartTime:   input.StartTime, // Always in UTC
		Duration:    input.Duration,
		// TimezoneOffset: input.TimezoneOffset, // Store as frontend hint, no conversion
		State:     booking.BookingStatePending,
		CreatedAt: now,
		UpdatedAt: now,
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
