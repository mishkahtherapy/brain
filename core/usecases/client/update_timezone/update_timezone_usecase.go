package update_timezone

import (
	"errors"

	"github.com/mishkahtherapy/brain/core/domain"
	"github.com/mishkahtherapy/brain/core/ports"
)

var (
	ErrTimezoneIsRequired = errors.New("timezone is required")
	ErrInvalidTimezone    = errors.New("invalid timezone format")
	ErrClientNotFound     = errors.New("client not found")
)

type Input struct {
	ClientID domain.ClientID `json:"clientId"`
	Timezone domain.Timezone `json:"timezone"`
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
	// Validate timezone
	if input.Timezone == "" {
		return ErrTimezoneIsRequired
	}

	if !domain.Timezone(input.Timezone).IsValid() {
		return ErrInvalidTimezone
	}

	// Check if client exists
	client, err := u.clientRepo.GetByID(input.ClientID)
	if err != nil {
		return err
	}
	if client == nil {
		return ErrClientNotFound
	}

	// Update client timezone
	return u.clientRepo.UpdateTimezone(input.ClientID, input.Timezone)
}
