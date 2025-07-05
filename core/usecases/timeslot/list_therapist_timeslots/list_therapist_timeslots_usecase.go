package list_therapist_timeslots

import (
	"github.com/mishkahtherapy/brain/core/domain"

	"github.com/mishkahtherapy/brain/core/domain/timeslot"
	"github.com/mishkahtherapy/brain/core/ports"
)

type Input struct {
	TherapistID domain.TherapistID `json:"therapistId"`
}

type Output struct {
	Timeslots []timeslot.TimeSlot `json:"timeslots"`
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

func (u *Usecase) Execute(input Input) (*Output, error) {
	// Validate input
	if err := u.validateInput(input); err != nil {
		return nil, err
	}

	// Verify therapist exists
	if _, err := u.therapistRepo.GetByID(input.TherapistID); err != nil {
		return nil, timeslot.ErrTherapistNotFound
	}

	var timeslots []*timeslot.TimeSlot
	var err error
	timeslots, err = u.timeslotRepo.ListByTherapist(string(input.TherapistID))

	if err != nil {
		return nil, err
	}

	// Convert []*domain.TimeSlot to []domain.TimeSlot for output
	result := make([]timeslot.TimeSlot, len(timeslots))
	for i, ts := range timeslots {
		result[i] = *ts
	}

	return &Output{
		Timeslots: result,
	}, nil
}

func (u *Usecase) validateInput(input Input) error {
	if input.TherapistID == "" {
		return timeslot.ErrTherapistIDRequired
	}

	return nil
}
