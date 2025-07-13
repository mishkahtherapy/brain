package update_therapist_timeslot

import (
	"time"

	"github.com/mishkahtherapy/brain/core/domain"
	"github.com/mishkahtherapy/brain/core/domain/therapist"
	"github.com/mishkahtherapy/brain/core/domain/timeslot"
	"github.com/mishkahtherapy/brain/core/ports"
	timeslot_usecase "github.com/mishkahtherapy/brain/core/usecases/timeslot"
)

type Input struct {
	TherapistID       domain.TherapistID     `json:"therapistId"`
	TimeslotID        domain.TimeSlotID      `json:"timeslotId"`
	DayOfWeek         timeslot.DayOfWeek     `json:"dayOfWeek"`
	Start             domain.Time24h         `json:"start"`             // "09:00"
	Duration          domain.DurationMinutes `json:"duration"`          // Duration in minutes
	PreSessionBuffer  domain.DurationMinutes `json:"preSessionBuffer"`  // minutes
	PostSessionBuffer domain.DurationMinutes `json:"postSessionBuffer"` // minutes
	IsActive          bool                   `json:"isActive"`
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

	// Get existing timeslot
	existingTimeslot, err := u.timeslotRepo.GetByID(input.TimeslotID)
	if err != nil {
		// Check if it's the repository's not found error
		if err.Error() == "timeslot not found" {
			return nil, timeslot.ErrTimeslotNotFound
		}
		return nil, err
	}

	// Verify the timeslot belongs to the specified therapist
	if existingTimeslot.TherapistID != input.TherapistID {
		return nil, timeslot.ErrTimeslotNotOwned
	}

	// Check for overlapping timeslots (excluding the current one)
	if err := u.checkForOverlaps(input); err != nil {
		return nil, err
	}

	// Update the timeslot
	updatedTimeslot := &timeslot.TimeSlot{
		ID:                input.TimeslotID,
		TherapistID:       input.TherapistID,
		DayOfWeek:         input.DayOfWeek,
		Start:             input.Start,
		Duration:          input.Duration,
		PreSessionBuffer:  input.PreSessionBuffer,
		PostSessionBuffer: input.PostSessionBuffer,
		IsActive:          input.IsActive,
		BookingIDs:        existingTimeslot.BookingIDs, // Preserve existing bookings
		CreatedAt:         existingTimeslot.CreatedAt,  // Preserve creation time
		UpdatedAt:         domain.UTCTimestamp(time.Now().UTC()),
	}

	// Save to repository
	if err := u.timeslotRepo.Update(updatedTimeslot); err != nil {
		return nil, err
	}

	return updatedTimeslot, nil
}

func (u *Usecase) validateInput(input Input) error {
	// Validate required fields
	if input.TherapistID == "" {
		return therapist.ErrTherapistIDRequired
	}

	if input.TimeslotID == "" {
		return timeslot.ErrTimeslotIDIsRequired
	}

	if input.DayOfWeek == "" {
		return timeslot.ErrDayOfWeekIsRequired
	}

	if input.Start == "" {
		return timeslot.ErrStartTimeIsRequired
	}

	if input.Duration == 0 {
		return timeslot.ErrDurationIsRequired
	}

	// Validate day of week
	if !timeslot_usecase.IsValidDayOfWeek(input.DayOfWeek) {
		return timeslot.ErrInvalidDayOfWeek
	}

	// Validate time format
	if _, err := timeslot_usecase.ParseTimeString(input.Start); err != nil {
		return err
	}

	// Validate duration
	if err := timeslot_usecase.ValidateDuration(input.Duration); err != nil {
		return err
	}

	// Validate buffer times
	if err := timeslot_usecase.ValidateBufferTimes(input.PreSessionBuffer, input.PostSessionBuffer); err != nil {
		return err
	}

	return nil
}

func (u *Usecase) checkForOverlaps(input Input) error {
	// Get all existing timeslots for this therapist
	existingSlots, err := u.timeslotRepo.ListByTherapist(input.TherapistID)
	if err != nil {
		return err
	}

	// Create a temporary timeslot for the update
	newSlot := timeslot.TimeSlot{
		ID:                input.TimeslotID,
		TherapistID:       input.TherapistID,
		DayOfWeek:         input.DayOfWeek,
		Start:             input.Start,
		Duration:          input.Duration,
		PreSessionBuffer:  input.PreSessionBuffer,
		PostSessionBuffer: input.PostSessionBuffer,
	}

	// Check for conflicts and insufficient gaps
	for _, existing := range existingSlots {
		// Skip the timeslot we're updating
		if existing.ID == input.TimeslotID {
			continue
		}

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
