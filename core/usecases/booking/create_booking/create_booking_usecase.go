package create_booking

import (
	"log/slog"
	"time"

	"github.com/mishkahtherapy/brain/core/domain"
	"github.com/mishkahtherapy/brain/core/domain/booking"
	"github.com/mishkahtherapy/brain/core/ports"
	"github.com/mishkahtherapy/brain/core/usecases/common"
	"github.com/mishkahtherapy/brain/core/usecases/schedule/get_schedule"
)

type Input struct {
	TherapistID          domain.TherapistID     `json:"therapistId"`
	ClientID             domain.ClientID        `json:"clientId"`
	TimeSlotID           domain.TimeSlotID      `json:"timeSlotId"`
	StartTime            domain.UTCTimestamp    `json:"startTime"`
	Duration             domain.DurationMinutes `json:"duration"`
	ClientTimezoneOffset domain.TimezoneOffset  `json:"clientTimezoneOffset"`
}

type Usecase struct {
	bookingRepo        ports.BookingRepository
	therapistRepo      ports.TherapistRepository
	clientRepo         ports.ClientRepository
	timeSlotRepo       ports.TimeSlotRepository
	getScheduleUsecase get_schedule.Usecase
}

func NewUsecase(
	bookingRepo ports.BookingRepository,
	therapistRepo ports.TherapistRepository,
	clientRepo ports.ClientRepository,
	timeSlotRepo ports.TimeSlotRepository,
	getScheduleUsecase get_schedule.Usecase,
) *Usecase {
	return &Usecase{
		bookingRepo:        bookingRepo,
		therapistRepo:      therapistRepo,
		clientRepo:         clientRepo,
		timeSlotRepo:       timeSlotRepo,
		getScheduleUsecase: getScheduleUsecase,
	}
}

func (u *Usecase) Execute(input Input) (*booking.Booking, error) {
	// Validate required fields
	if err := validateInput(input); err != nil {
		return nil, err
	}

	// Check if client exists
	client, err := u.clientRepo.FindByIDs([]domain.ClientID{input.ClientID})
	if err != nil || client == nil {
		return nil, common.ErrClientNotFound
	}

	startTime := time.Time(input.StartTime)
	endTime := startTime.Add(time.Duration(input.Duration) * time.Minute)
	availabilities, err := u.getScheduleUsecase.Execute(get_schedule.Input{
		TherapistIDs: []domain.TherapistID{input.TherapistID},
		StartDate:    startTime,
		EndDate:      endTime,
	})

	if err != nil {
		return nil, err
	}

	if len(availabilities) == 0 {
		return nil, common.ErrTimeSlotAlreadyBooked
	}

	if len(availabilities) > 1 {
		slog.Error(
			"multiple availabilities found for the same time slot",
			"timeSlotId", input.TimeSlotID,
			"availabilities", availabilities,
			"startTime", input.StartTime,
			"endTime", endTime,
		)
		return nil, common.ErrTimeSlotAlreadyBooked
	}

	availability := availabilities[0]
	// Make sure the booked timeslot is within the availability
	availabilityStartTime := time.Time(availability.From)
	availabilityEndTime := availabilityStartTime.Add(time.Duration(availability.Duration) * time.Minute)
	if input.StartTime.Time().Before(availabilityStartTime) || input.StartTime.Time().Add(time.Duration(input.Duration)*time.Minute).After(availabilityEndTime) {
		slog.Error(
			"booked timeslot is not within the availability",
			"timeSlotId", input.TimeSlotID,
			"availabilityStartTime", availabilityStartTime,
			"availabilityEndTime", availabilityEndTime,
			"startTime", input.StartTime,
			"endTime", endTime,
		)
		return nil, common.ErrTimeSlotAlreadyBooked
	}

	// Create booking with Pending state and timezone (no conversion, just store as hint)
	now := domain.NewUTCTimestamp()
	createdBooking := &booking.Booking{
		ID:                   domain.NewBookingID(),
		TherapistID:          input.TherapistID,
		ClientID:             input.ClientID,
		TimeSlotID:           input.TimeSlotID,
		StartTime:            input.StartTime, // Always in UTC
		Duration:             input.Duration,
		ClientTimezoneOffset: input.ClientTimezoneOffset,
		State:                booking.BookingStatePending,
		CreatedAt:            now,
		UpdatedAt:            now,
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
