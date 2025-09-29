package create_booking

import (
	"time"

	"github.com/mishkahtherapy/brain/core/domain"
	"github.com/mishkahtherapy/brain/core/domain/booking"
	"github.com/mishkahtherapy/brain/core/domain/schedule"
	"github.com/mishkahtherapy/brain/core/ports"
	"github.com/mishkahtherapy/brain/core/usecases/common"
	"github.com/mishkahtherapy/brain/core/usecases/common/overlap_detector"
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

func (u *Usecase) Execute(input Input) (*ports.BookingResponse, error) {
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

	var matchingAvailability *schedule.AvailableTimeRange
	for _, availability := range availabilities {
		matches := checkIfAvailabilityMatches(availability, input)

		if matches {
			matchingAvailability = &availability
			break
		}
	}

	if matchingAvailability == nil {
		return nil, common.ErrInvalidBookingTime
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

	return &ports.BookingResponse{
		RegularBookingID:     createdBooking.ID,
		TherapistID:          createdBooking.TherapistID,
		ClientID:             createdBooking.ClientID,
		State:                createdBooking.State,
		StartTime:            createdBooking.StartTime,
		Duration:             createdBooking.Duration,
		ClientTimezoneOffset: createdBooking.ClientTimezoneOffset,
	}, nil
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

func checkIfAvailabilityMatches(availability schedule.AvailableTimeRange, input Input) bool {
	// Make sure the booked timeslot is within the availability
	availabilityStartTime := time.Time(availability.From)
	availabilityEndTime := availabilityStartTime.Add(time.Duration(availability.Duration) * time.Minute)

	inputStartTime := time.Time(input.StartTime)
	inputEndTime := inputStartTime.Add(time.Duration(input.Duration) * time.Minute)

	return overlap_detector.New(
		availabilityStartTime,
		availabilityEndTime,
	).HasOverlap(
		inputStartTime,
		inputEndTime,
	)
}
