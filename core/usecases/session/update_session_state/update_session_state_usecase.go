package update_session_state

import (
	"errors"

	"github.com/mishkahtherapy/brain/core/domain"
	"github.com/mishkahtherapy/brain/core/ports"
)

// Error definitions
var ErrSessionNotFound = errors.New("session not found")
var ErrSessionIDIsRequired = errors.New("session ID is required")
var ErrInvalidStateTransition = errors.New("invalid state transition")
var ErrStateIsRequired = errors.New("state is required")
var ErrFailedToUpdateSessionState = errors.New("failed to update session state")

// Input struct defines parameters for updating a session state
type Input struct {
	SessionID domain.SessionID    `json:"sessionId"`
	NewState  domain.SessionState `json:"newState"`
}

// Usecase struct with required dependencies
type Usecase struct {
	sessionRepo ports.SessionRepository
}

// NewUsecase creates a new instance of the update session state usecase
func NewUsecase(sessionRepo ports.SessionRepository) *Usecase {
	return &Usecase{sessionRepo: sessionRepo}
}

// Execute updates a session's state if the transition is valid
func (u *Usecase) Execute(input Input) (*domain.Session, error) {
	// Validate input
	if input.SessionID == "" {
		return nil, ErrSessionIDIsRequired
	}
	if input.NewState == "" {
		return nil, ErrStateIsRequired
	}

	// Get the current session
	session, err := u.sessionRepo.GetSessionByID(input.SessionID)
	if err != nil {
		return nil, ErrSessionNotFound
	}

	// Validate state transition
	if !session.IsValidStateTransition(input.NewState) {
		return nil, ErrInvalidStateTransition
	}

	// Update the session state
	session.State = input.NewState
	session.UpdatedAt = domain.NewUTCTimestamp()

	// Persist the change
	err = u.sessionRepo.UpdateSessionState(input.SessionID, input.NewState)
	if err != nil {
		return nil, ErrFailedToUpdateSessionState
	}

	return session, nil
}
