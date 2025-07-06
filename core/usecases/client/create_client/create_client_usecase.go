package create_client

import (
	"errors"
	"strings"

	"github.com/mishkahtherapy/brain/core/domain"
	"github.com/mishkahtherapy/brain/core/domain/client"
	"github.com/mishkahtherapy/brain/core/ports"
)

var (
	ErrWhatsAppNumberIsRequired = errors.New("whatsapp number is required")
	ErrInvalidWhatsAppNumber    = errors.New("invalid whatsapp number format")
	ErrClientAlreadyExists      = errors.New("client with this whatsapp number already exists")
	ErrTimezoneIsRequired       = errors.New("timezone is required")
	ErrInvalidTimezone          = errors.New("invalid timezone format")
)

type Input struct {
	Name           string                `json:"name"`
	WhatsAppNumber domain.WhatsAppNumber `json:"whatsAppNumber"`
	Timezone       string                `json:"timezone"` // Required field
}

type Usecase struct {
	clientRepo ports.ClientRepository
}

func NewUsecase(clientRepo ports.ClientRepository) *Usecase {
	return &Usecase{
		clientRepo: clientRepo,
	}
}

func (u *Usecase) Execute(input Input) (*client.Client, error) {
	// Validate input
	if err := u.validateInput(input); err != nil {
		return nil, err
	}

	// Check if client already exists with this WhatsApp number
	existingClient, err := u.clientRepo.GetByWhatsAppNumber(input.WhatsAppNumber)
	if err != nil {
		return nil, err
	}
	if existingClient != nil {
		return nil, ErrClientAlreadyExists
	}

	// Create new client
	client := &client.Client{
		ID:             domain.NewClientID(),
		Name:           strings.TrimSpace(input.Name),
		WhatsAppNumber: input.WhatsAppNumber,
		Timezone:       input.Timezone, // Required timezone
		BookingIDs:     []domain.BookingID{},
		CreatedAt:      domain.NewUTCTimestamp(),
		UpdatedAt:      domain.NewUTCTimestamp(),
	}

	// Save to repository
	if err := u.clientRepo.Create(client); err != nil {
		return nil, err
	}

	return client, nil
}

func (u *Usecase) validateInput(input Input) error {
	// Validate WhatsApp number
	if string(input.WhatsAppNumber) == "" {
		return ErrWhatsAppNumberIsRequired
	}

	// Basic WhatsApp number format validation (starts with + and has digits)
	whatsAppStr := string(input.WhatsAppNumber)
	if !strings.HasPrefix(whatsAppStr, "+") || len(whatsAppStr) < 8 {
		return ErrInvalidWhatsAppNumber
	}

	// Check if the rest are digits (after the +)
	for _, char := range whatsAppStr[1:] {
		if char < '0' || char > '9' {
			return ErrInvalidWhatsAppNumber
		}
	}

	// Validate timezone (required)
	if input.Timezone == "" {
		return ErrTimezoneIsRequired
	}

	if !domain.Timezone(input.Timezone).IsValid() {
		return ErrInvalidTimezone
	}

	return nil
}
