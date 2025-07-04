package bulk_toggle_therapist_timeslots

import (
	"github.com/mishkahtherapy/brain/core/domain"
	"github.com/mishkahtherapy/brain/core/domain/timeslot"
	"github.com/mishkahtherapy/brain/core/ports"
	"github.com/mishkahtherapy/brain/core/usecases/common"
)

type Input struct {
	TherapistID domain.TherapistID
	IsActive    bool
}

type Usecase interface {
	Execute(input Input) error
}

type usecase struct {
	therapistRepo ports.TherapistRepository
	timeslotRepo  ports.TimeSlotRepository
}

func NewUsecase(therapistRepo ports.TherapistRepository, timeslotRepo ports.TimeSlotRepository) Usecase {
	return &usecase{
		therapistRepo: therapistRepo,
		timeslotRepo:  timeslotRepo,
	}
}

func (u *usecase) Execute(input Input) error {
	// Validate input
	if input.TherapistID == "" {
		return timeslot.ErrTherapistIDRequired
	}

	// Check if therapist exists
	_, err := u.therapistRepo.GetByID(input.TherapistID)
	if err != nil {
		if err == common.ErrTherapistNotFound {
			return timeslot.ErrTherapistNotFound
		}
		return err
	}

	// Bulk toggle all timeslots for the therapist
	err = u.timeslotRepo.BulkToggleByTherapistID(string(input.TherapistID), input.IsActive)
	if err != nil {
		return err
	}

	return nil
}
