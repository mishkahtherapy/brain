package create_adhoc_booking

import (
	"time"

	"github.com/mishkahtherapy/brain/core/domain"
	"github.com/mishkahtherapy/brain/core/domain/booking"
	"github.com/mishkahtherapy/brain/core/domain/timeslot"
	"github.com/mishkahtherapy/brain/core/ports"
	"github.com/mishkahtherapy/brain/core/usecases/common"
	"github.com/mishkahtherapy/brain/core/usecases/common/overlap_detector"
)

type Input struct {
	TherapistID          domain.TherapistID     `json:"therapistId"`
	ClientID             domain.ClientID        `json:"clientId"`
	StartTime            domain.UTCTimestamp    `json:"startTime"`
	Duration             domain.DurationMinutes `json:"duration"`
	ClientTimezoneOffset domain.TimezoneOffset  `json:"clientTimezoneOffset"`
}

type Usecase struct {
	bookingRepo      ports.BookingRepository
	adhocBookingRepo ports.AdhocBookingRepository
	timeSlotRepo     ports.TimeSlotRepository
	therapistRepo    ports.TherapistRepository
	clientRepo       ports.ClientRepository
}

func NewUsecase(
	bookingRepo ports.BookingRepository,
	adhocBookingRepo ports.AdhocBookingRepository,
	timeSlotRepo ports.TimeSlotRepository,
	therapistRepo ports.TherapistRepository,
	clientRepo ports.ClientRepository,
) *Usecase {
	return &Usecase{
		bookingRepo:      bookingRepo,
		adhocBookingRepo: adhocBookingRepo,
		timeSlotRepo:     timeSlotRepo,
		therapistRepo:    therapistRepo,
		clientRepo:       clientRepo,
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

	// Get booking day of week
	bookingDayOfWeek := time.Time(input.StartTime).Weekday()

	// Get therapist
	therapist, err := u.therapistRepo.GetByID(input.TherapistID)
	if err != nil {
		return nil, common.ErrTherapistNotFound
	}

	// Get day timeslots
	timeslots, err := u.timeSlotRepo.ListByTherapist(therapist.ID)
	if err != nil {
		return nil, common.ErrFailedToCreateBooking
	}

	// Filter timeslots by day of week
	dayTimeslots := make([]*timeslot.TimeSlot, 0)
	for _, slot := range timeslots {
		if !slot.IsActive {
			continue
		}

		if slot.DayOfWeek == timeslot.MapToDayOfWeek(bookingDayOfWeek) {
			dayTimeslots = append(dayTimeslots, slot)
		}
	}

	// Make sure booking doesn't intersect with any of the therapist's timeslots.
	for _, slot := range dayTimeslots {
		slotStart, _ := slot.ApplyToDate(time.Time(input.StartTime))
		if hasOverlap(
			slotStart, slot.Duration,
			input.StartTime, input.Duration,
		) {
			return nil, timeslot.ErrBookingShouldBeMadeInTimeslot
		}
	}

	// Get regular bookings on the same day. This handles the case of a therapist
	// modifying their timeslot after a booking has been made using the old timeslot range.
	regularBookings, err := u.bookingRepo.BulkListByTherapistForDateRange(
		[]domain.TherapistID{input.TherapistID},
		[]booking.BookingState{booking.BookingStateConfirmed},
		time.Time(input.StartTime),
		time.Time(input.StartTime),
	)
	if err != nil {
		return nil, common.ErrFailedToCreateBooking
	}

	// Make sure booking doesn't intersect with any of the regular bookings
	for _, booking := range regularBookings[input.TherapistID] {
		if hasOverlap(
			booking.StartTime, booking.Duration,
			input.StartTime, input.Duration,
		) {
			return nil, timeslot.ErrOverlappingBooking
		}
	}

	// Get adhoc bookings on the same day
	adhocBookingMap, err := u.adhocBookingRepo.BulkListByTherapistForDateRange(
		[]domain.TherapistID{input.TherapistID},
		[]booking.BookingState{booking.BookingStateConfirmed},
		time.Time(input.StartTime),
		time.Time(input.StartTime),
	)
	if err != nil {
		return nil, common.ErrFailedToCreateBooking
	}

	adhocBookings := adhocBookingMap[input.TherapistID]
	// Make sure booking doesn't intersect with any of the adhoc bookings
	for _, booking := range adhocBookings {
		if hasOverlap(
			booking.StartTime, booking.Duration,
			input.StartTime, input.Duration,
		) {
			return nil, timeslot.ErrOverlappingBooking
		}
	}

	// Create adhoc booking
	now := domain.NewUTCTimestamp()
	adhocBooking := &booking.AdhocBooking{
		ID:                   domain.NewAdhocBookingID(),
		TherapistID:          input.TherapistID,
		ClientID:             input.ClientID,
		StartTime:            input.StartTime,
		Duration:             input.Duration,
		State:                booking.BookingStatePending,
		ClientTimezoneOffset: input.ClientTimezoneOffset,
		CreatedAt:            now,
		UpdatedAt:            now,
	}

	// Save to repository
	err = u.adhocBookingRepo.Create(adhocBooking)
	if err != nil {
		return nil, err
	}

	return &ports.BookingResponse{
		AdhocBookingID:       adhocBooking.ID,
		TherapistID:          adhocBooking.TherapistID,
		ClientID:             adhocBooking.ClientID,
		State:                adhocBooking.State,
		StartTime:            adhocBooking.StartTime,
		Duration:             adhocBooking.Duration,
		ClientTimezoneOffset: adhocBooking.ClientTimezoneOffset,
	}, nil
}

func validateInput(input Input) error {
	if input.TherapistID == "" {
		return common.ErrTherapistIDIsRequired
	}
	if input.ClientID == "" {
		return common.ErrClientIDIsRequired
	}
	if input.StartTime == (domain.UTCTimestamp{}) {
		return common.ErrStartTimeIsRequired
	}
	if input.Duration == 0 {
		return common.ErrDurationIsRequired
	}
	if input.ClientTimezoneOffset == 0 {
		return common.ErrClientTimezoneOffsetIsRequired
	}
	return nil
}

func hasOverlap(
	start domain.UTCTimestamp,
	duration domain.DurationMinutes,
	otherStart domain.UTCTimestamp,
	otherDuration domain.DurationMinutes,
) bool {
	end := start.Add(time.Duration(duration) * time.Minute)
	otherEnd := otherStart.Add(time.Duration(otherDuration) * time.Minute)
	overlapDetector := overlap_detector.New(
		start.Time(), end.Time(),
	)
	return overlapDetector.HasOverlap(
		otherStart.Time(), otherEnd.Time(),
	)
}
