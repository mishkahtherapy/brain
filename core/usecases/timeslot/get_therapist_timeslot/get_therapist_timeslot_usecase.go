package get_therapist_timeslot

import (
	"github.com/mishkahtherapy/brain/core/domain"
	"github.com/mishkahtherapy/brain/core/domain/therapist"
	"github.com/mishkahtherapy/brain/core/domain/timeslot"
	"github.com/mishkahtherapy/brain/core/ports"
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

func (u *Usecase) Execute(input Input) (*timeslot.TimeSlot, error) {
	// Validate input
	if err := u.validateInput(input); err != nil {
		return nil, err
	}

	// Verify therapist exists
	if _, err := u.therapistRepo.GetByID(input.TherapistID); err != nil {
		return nil, timeslot.ErrTherapistNotFound
	}

	// Get the timeslot
	timeslotResult, err := u.timeslotRepo.GetByID(input.TimeslotID)
	if err != nil {
		// Check if it's the repository's not found error
		if err.Error() == "timeslot not found" {
			return nil, timeslot.ErrTimeslotNotFound
		}
		return nil, err
	}

	// Verify the timeslot belongs to the specified therapist
	if timeslotResult.TherapistID != input.TherapistID {
		return nil, timeslot.ErrTimeslotNotOwned
	}

	return timeslotResult, nil
}

func (u *Usecase) validateInput(input Input) error {
	if input.TherapistID == "" {
		return therapist.ErrTherapistIDRequired
	}

	if input.TimeslotID == "" {
		return timeslot.ErrTimeslotIDIsRequired
	}

	return nil
}
