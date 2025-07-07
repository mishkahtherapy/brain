package create_therapist_timeslot

import (
	"time"

	"github.com/mishkahtherapy/brain/core/domain"
	"github.com/mishkahtherapy/brain/core/domain/timeslot"
	"github.com/mishkahtherapy/brain/core/ports"
	timeslot_usecase "github.com/mishkahtherapy/brain/core/usecases/timeslot"
)

type Input struct {
	TherapistID       domain.TherapistID    `json:"therapistId"`
	LocalDayOfWeek    string                `json:"dayOfWeek"`         // Local day "Monday"
	LocalStartTime    string                `json:"startTime"`         // Local time "01:30"
	DurationMinutes   int                   `json:"durationMinutes"`   // Duration in minutes
	TimezoneOffset    domain.TimezoneOffset `json:"timezoneOffset"`    // Minutes from UTC (+180 for GMT+3)
	PreSessionBuffer  int                   `json:"preSessionBuffer"`  // minutes
	PostSessionBuffer int                   `json:"postSessionBuffer"` // minutes
}

type Usecase struct {
	therapistRepo ports.TherapistRepository
	timeslotRepo  ports.TimeSlotRepository
}

func NewUsecase(therapistRepo ports.TherapistRepository, timeslotRepo ports.TimeSlotRepository) *Usecase {
	return &Usecase{
		therapistRepo: therapistRepo,
		timeslotRepo:  timeslotRepo,
	}
}

func (u *Usecase) Execute(input Input) (*timeslot.TimeSlot, error) {
	// Validate input
	if err := u.validateInput(input); err != nil {
		return nil, err
	}

	// Verify therapist exists
	if _, err := u.therapistRepo.GetByID(input.TherapistID); err != nil {
		return nil, timeslot.ErrTherapistNotFound
	}

	// Convert local time to UTC
	utcDay, utcStart, err := timeslot_usecase.ConvertLocalToUTC(
		input.LocalDayOfWeek,
		input.LocalStartTime,
		input.TimezoneOffset,
	)
	if err != nil {
		return nil, err
	}

	// Create UTC timeslot for storage
	utcTimeslot := &timeslot.TimeSlot{
		TherapistID:       input.TherapistID,
		DayOfWeek:         timeslot.DayOfWeek(utcDay),
		StartTime:         utcStart,
		DurationMinutes:   input.DurationMinutes,
		PreSessionBuffer:  input.PreSessionBuffer,
		PostSessionBuffer: input.PostSessionBuffer,
		IsActive:          true,
	}

	// Check for overlapping timeslots (using UTC times)
	if err := u.checkForOverlaps(*utcTimeslot); err != nil {
		return nil, err
	}

	// Generate unique ID and timestamps
	timeslotID := domain.NewTimeSlotID()
	now := domain.UTCTimestamp(time.Now().UTC())

	utcTimeslot.ID = timeslotID
	utcTimeslot.BookingIDs = make([]domain.BookingID, 0)
	utcTimeslot.CreatedAt = now
	utcTimeslot.UpdatedAt = now

	// Save to repository
	if err := u.timeslotRepo.Create(utcTimeslot); err != nil {
		return nil, err
	}

	return utcTimeslot, nil
}

func (u *Usecase) validateInput(input Input) error {
	// Validate required fields
	if input.TherapistID == "" {
		return timeslot.ErrTherapistIDRequired
	}

	if input.LocalDayOfWeek == "" {
		return timeslot.ErrDayOfWeekIsRequired
	}

	if input.LocalStartTime == "" {
		return timeslot.ErrStartTimeIsRequired
	}

	if input.DurationMinutes == 0 {
		return timeslot.ErrDurationIsRequired
	}

	// Validate day of week
	dayOfWeek := timeslot.DayOfWeek(input.LocalDayOfWeek)
	if !timeslot_usecase.IsValidDayOfWeek(dayOfWeek) {
		return timeslot.ErrInvalidDayOfWeek
	}

	// Validate local start time format
	if _, err := timeslot_usecase.ParseTimeString(input.LocalStartTime); err != nil {
		return err
	}

	// Validate duration
	if err := timeslot_usecase.ValidateDuration(input.DurationMinutes); err != nil {
		return err
	}

	// Validate timezone offset
	if err := timeslot_usecase.ValidateTimezoneOffset(input.TimezoneOffset); err != nil {
		return err
	}

	// Validate buffer times
	if err := timeslot_usecase.ValidateBufferTimes(input.PreSessionBuffer, input.PostSessionBuffer); err != nil {
		return err
	}

	return nil
}

func (u *Usecase) checkForOverlaps(newSlot timeslot.TimeSlot) error {
	// Get all existing timeslots for this therapist
	existingSlots, err := u.timeslotRepo.ListByTherapist(string(newSlot.TherapistID))
	if err != nil {
		return err
	}

	// Check for conflicts and insufficient gaps
	for _, existing := range existingSlots {
		// Check for overlapping effective time ranges (including buffers)
		if timeslot_usecase.HasEffectiveTimeSlotConflict(newSlot, *existing) {
			return timeslot.ErrOverlappingTimeslot
		}

		// Check for sufficient gap between slots (at least 30 minutes)
		if !timeslot_usecase.HasSufficientGapBetweenSlots(newSlot, *existing) {
			return timeslot.ErrInsufficientGapBetweenSlots
		}
	}

	return nil
}
