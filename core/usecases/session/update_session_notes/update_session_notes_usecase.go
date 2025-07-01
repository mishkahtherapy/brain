package update_session_notes

import (
	"github.com/mishkahtherapy/brain/core/domain"
	"github.com/mishkahtherapy/brain/core/ports"
	"github.com/mishkahtherapy/brain/core/usecases/common"
)

// Input struct defines parameters for updating session notes
type Input struct {
	SessionID domain.SessionID `json:"sessionId"`
	Notes     string           `json:"notes"`
}

// Usecase struct with required dependencies
type Usecase struct {
	sessionRepo ports.SessionRepository
}

// NewUsecase creates a new instance of the update session notes usecase
func NewUsecase(sessionRepo ports.SessionRepository) *Usecase {
	return &Usecase{sessionRepo: sessionRepo}
}

// Execute updates a session's notes by appending the new note with a timestamp
func (u *Usecase) Execute(input Input) (*domain.Session, error) {
	// Validate input
	if input.SessionID == "" {
		return nil, common.ErrSessionIDIsRequired
	}
	if input.Notes == "" {
		return nil, common.ErrNotesIsRequired
	}

	// Get the current session
	session, err := u.sessionRepo.GetSessionByID(input.SessionID)
	if err != nil {
		return nil, common.ErrSessionNotFound
	}

	// Append the new note with timestamp
	session.AppendNote(input.Notes)

	// Persist the change
	err = u.sessionRepo.UpdateSessionNotes(input.SessionID, session.Notes)
	if err != nil {
		return nil, common.ErrFailedToUpdateSessionNotes
	}

	return session, nil
}
