package list_sessions_by_client

import (
	"errors"

	"github.com/mishkahtherapy/brain/core/domain"
	"github.com/mishkahtherapy/brain/core/ports"
)

// Error definitions
var ErrClientIDIsRequired = errors.New("client ID is required")
var ErrFailedToListSessions = errors.New("failed to list sessions")

// Input struct defines parameters for listing sessions by client
type Input struct {
	ClientID domain.ClientID `json:"clientId"`
}

// Usecase struct with required dependencies
type Usecase struct {
	sessionRepo ports.SessionRepository
}

// NewUsecase creates a new instance of the list sessions by client usecase
func NewUsecase(sessionRepo ports.SessionRepository) *Usecase {
	return &Usecase{sessionRepo: sessionRepo}
}

// Execute retrieves all sessions for a specific client
func (u *Usecase) Execute(input Input) ([]*domain.Session, error) {
	// Validate input
	if input.ClientID == "" {
		return nil, ErrClientIDIsRequired
	}

	// Retrieve sessions from repository
	sessions, err := u.sessionRepo.ListSessionsByClient(input.ClientID)
	if err != nil {
		return nil, ErrFailedToListSessions
	}

	// Return sessions (empty slice if none found)
	return sessions, nil
}
