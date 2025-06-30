package list_sessions_by_therapist

import (
	"errors"

	"github.com/mishkahtherapy/brain/core/domain"
	"github.com/mishkahtherapy/brain/core/ports"
)

// Error definitions
var ErrTherapistIDIsRequired = errors.New("therapist ID is required")
var ErrFailedToListSessions = errors.New("failed to list sessions")

// Input struct defines parameters for listing sessions by therapist
type Input struct {
	TherapistID domain.TherapistID `json:"therapistId"`
}

// Usecase struct with required dependencies
type Usecase struct {
	sessionRepo ports.SessionRepository
}

// NewUsecase creates a new instance of the list sessions by therapist usecase
func NewUsecase(sessionRepo ports.SessionRepository) *Usecase {
	return &Usecase{sessionRepo: sessionRepo}
}

// Execute retrieves all sessions for a specific therapist
func (u *Usecase) Execute(input Input) ([]*domain.Session, error) {
	// Validate input
	if input.TherapistID == "" {
		return nil, ErrTherapistIDIsRequired
	}

	// Retrieve sessions from repository
	sessions, err := u.sessionRepo.ListSessionsByTherapist(input.TherapistID)
	if err != nil {
		return nil, ErrFailedToListSessions
	}

	// Return sessions (empty slice if none found)
	return sessions, nil
}
