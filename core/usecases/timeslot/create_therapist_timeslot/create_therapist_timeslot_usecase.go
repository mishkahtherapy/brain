package create_therapist_timeslot

import (
	"time"

	"github.com/mishkahtherapy/brain/core/domain"
	"github.com/mishkahtherapy/brain/core/ports"
	"github.com/mishkahtherapy/brain/core/usecases/timeslot"
)

type Input struct {
	TherapistID       domain.TherapistID `json:"therapistId"`
	DayOfWeek         domain.DayOfWeek   `json:"dayOfWeek"`
	StartTime         string             `json:"startTime"`         // "09:00"
	EndTime           string             `json:"endTime"`           // "17:00"
	PreSessionBuffer  int                `json:"preSessionBuffer"`  // minutes
	PostSessionBuffer int                `json:"postSessionBuffer"` // minutes
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

func (u *Usecase) Execute(input Input) (*domain.TimeSlot, error) {
	// Validate input
	if err := u.validateInput(input); err != nil {
		return nil, err
	}

	// Verify therapist exists
	if _, err := u.therapistRepo.GetByID(input.TherapistID); err != nil {
		return nil, timeslot.ErrTherapistNotFound
	}

	// Check for overlapping timeslots
	if err := u.checkForOverlaps(input); err != nil {
		return nil, err
	}

	// Generate unique ID for the timeslot
	timeslotID := domain.NewTimeSlotID()

	// Create timeslot
	now := domain.UTCTimestamp(time.Now().UTC())
	newTimeslot := &domain.TimeSlot{
		ID:                timeslotID,
		TherapistID:       input.TherapistID,
		DayOfWeek:         input.DayOfWeek,
		StartTime:         input.StartTime,
		EndTime:           input.EndTime,
		PreSessionBuffer:  input.PreSessionBuffer,
		PostSessionBuffer: input.PostSessionBuffer,
		BookingIDs:        make([]domain.BookingID, 0),
		CreatedAt:         now,
		UpdatedAt:         now,
	}

	// Save to repository
	if err := u.timeslotRepo.Create(newTimeslot); err != nil {
		return nil, err
	}

	return newTimeslot, nil
}

func (u *Usecase) validateInput(input Input) error {
	// Validate required fields
	if input.TherapistID == "" {
		return timeslot.ErrTherapistIDIsRequired
	}

	if input.DayOfWeek == "" {
		return timeslot.ErrDayOfWeekIsRequired
	}

	if input.StartTime == "" {
		return timeslot.ErrStartTimeIsRequired
	}

	if input.EndTime == "" {
		return timeslot.ErrEndTimeIsRequired
	}

	// Validate day of week
	if !timeslot.IsValidDayOfWeek(input.DayOfWeek) {
		return timeslot.ErrInvalidDayOfWeek
	}

	// Validate time format and range
	if err := timeslot.ValidateTimeRange(input.StartTime, input.EndTime); err != nil {
		return err
	}

	// Validate buffer times
	if err := timeslot.ValidateBufferTimes(input.PreSessionBuffer, input.PostSessionBuffer); err != nil {
		return err
	}

	return nil
}

func (u *Usecase) checkForOverlaps(input Input) error {
	// Get existing timeslots for this therapist on this day
	existingSlots, err := u.timeslotRepo.ListByDay(string(input.TherapistID), string(input.DayOfWeek))
	if err != nil {
		return err
	}

	// Parse new timeslot times
	newStart, _ := timeslot.ParseTimeString(input.StartTime)
	newEnd, _ := timeslot.ParseTimeString(input.EndTime)

	// Check for overlaps with existing timeslots
	for _, existing := range existingSlots {
		existingStart, _ := timeslot.ParseTimeString(existing.StartTime)
		existingEnd, _ := timeslot.ParseTimeString(existing.EndTime)

		// Check if time ranges overlap
		if timeslot.TimesOverlap(newStart, newEnd, existingStart, existingEnd) {
			return timeslot.ErrOverlappingTimeslot
		}
	}

	return nil
}
