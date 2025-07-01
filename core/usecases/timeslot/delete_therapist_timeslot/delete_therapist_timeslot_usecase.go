package delete_therapist_timeslot

import (
	"github.com/mishkahtherapy/brain/core/domain"
	"github.com/mishkahtherapy/brain/core/ports"
	"github.com/mishkahtherapy/brain/core/usecases/timeslot"
)

type Input struct {
	TherapistID domain.TherapistID `json:"therapistId"`
	TimeslotID  domain.TimeSlotID  `json:"timeslotId"`
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

func (u *Usecase) Execute(input Input) error {
	// Validate input
	if err := u.validateInput(input); err != nil {
		return err
	}

	// Verify therapist exists
	if _, err := u.therapistRepo.GetByID(input.TherapistID); err != nil {
		return timeslot.ErrTherapistNotFound
	}

	// Get the timeslot
	timeslotResult, err := u.timeslotRepo.GetByID(string(input.TimeslotID))
	if err != nil {
		// Check if it's the repository's not found error
		if err.Error() == "timeslot not found" {
			return timeslot.ErrTimeslotNotFound
		}
		return err
	}

	// Verify the timeslot belongs to the specified therapist
	if timeslotResult.TherapistID != input.TherapistID {
		return timeslot.ErrTimeslotNotOwned
	}

	// Check if timeslot has active bookings
	if len(timeslotResult.BookingIDs) > 0 {
		return timeslot.ErrTimeslotHasActiveBookings
	}

	// Delete the timeslot
	if err := u.timeslotRepo.Delete(string(input.TimeslotID)); err != nil {
		return err
	}

	return nil
}

func (u *Usecase) validateInput(input Input) error {
	if input.TherapistID == "" {
		return timeslot.ErrTherapistIDIsRequired
	}

	if input.TimeslotID == "" {
		return timeslot.ErrTimeslotIDIsRequired
	}

	return nil
}
