package list_therapist_timeslots

import (
	"github.com/mishkahtherapy/brain/core/domain"
	"github.com/mishkahtherapy/brain/core/ports"
	"github.com/mishkahtherapy/brain/core/usecases/timeslot"
)

type Input struct {
	TherapistID domain.TherapistID `json:"therapistId"`
	DayOfWeek   *domain.DayOfWeek  `json:"dayOfWeek,omitempty"` // Optional filter by day
}

type Output struct {
	Timeslots []domain.TimeSlot `json:"timeslots"`
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

	var timeslots []*domain.TimeSlot
	var err error

	// If day filter is provided, use ListByDay, otherwise get all therapist timeslots
	if input.DayOfWeek != nil {
		timeslots, err = u.timeslotRepo.ListByDay(string(input.TherapistID), string(*input.DayOfWeek))
	} else {
		timeslots, err = u.timeslotRepo.ListByTherapist(string(input.TherapistID))
	}

	if err != nil {
		return nil, err
	}

	// Convert []*domain.TimeSlot to []domain.TimeSlot for output
	result := make([]domain.TimeSlot, len(timeslots))
	for i, ts := range timeslots {
		result[i] = *ts
	}

	return &Output{
		Timeslots: result,
	}, nil
}

func (u *Usecase) validateInput(input Input) error {
	if input.TherapistID == "" {
		return timeslot.ErrTherapistIDIsRequired
	}

	// Validate day of week if provided
	if input.DayOfWeek != nil {
		if !timeslot.IsValidDayOfWeek(*input.DayOfWeek) {
			return timeslot.ErrInvalidDayOfWeek
		}
	}

	return nil
}
