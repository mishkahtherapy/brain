package get_session

import (
	"errors"

	"github.com/mishkahtherapy/brain/core/domain"
	"github.com/mishkahtherapy/brain/core/ports"
)

// Error definitions
var ErrSessionNotFound = errors.New("session not found")
var ErrSessionIDIsRequired = errors.New("session ID is required")

// Usecase struct with required dependencies
type Usecase struct {
	sessionRepo ports.SessionRepository
}

// NewUsecase creates a new instance of the get session usecase
func NewUsecase(sessionRepo ports.SessionRepository) *Usecase {
	return &Usecase{sessionRepo: sessionRepo}
}

// Execute retrieves a session by its ID
func (u *Usecase) Execute(id domain.SessionID) (*domain.Session, error) {
	if id == "" {
		return nil, ErrSessionIDIsRequired
	}

	session, err := u.sessionRepo.GetSessionByID(id)
	if err != nil {
		return nil, ErrSessionNotFound
	}

	return session, nil
}
