package update_timezone

import (
	"errors"

	"github.com/mishkahtherapy/brain/core/domain"
	"github.com/mishkahtherapy/brain/core/ports"
	timeslot_usecase "github.com/mishkahtherapy/brain/core/usecases/timeslot"
)

var (
	ErrInvalidTimezoneOffset = errors.New("invalid timezoneOffset")
	ErrClientNotFound        = errors.New("client not found")
)

type Input struct {
	ClientID       domain.ClientID       `json:"clientId"`
	TimezoneOffset domain.TimezoneOffset `json:"timezoneOffset"`
}

type Usecase struct {
	clientRepo ports.ClientRepository
}

func NewUsecase(clientRepo ports.ClientRepository) *Usecase {
	return &Usecase{
		clientRepo: clientRepo,
	}
}

func (u *Usecase) Execute(input Input) error {
	// Validate offset
	if err := timeslot_usecase.ValidateTimezoneOffset(input.TimezoneOffset); err != nil {
		return ErrInvalidTimezoneOffset
	}

	// Check if client exists
	client, err := u.clientRepo.GetByID(input.ClientID)
	if err != nil {
		return err
	}
	if client == nil {
		return ErrClientNotFound
	}

	// Update client timezone offset
	return u.clientRepo.UpdateTimezoneOffset(input.ClientID, input.TimezoneOffset)
}
