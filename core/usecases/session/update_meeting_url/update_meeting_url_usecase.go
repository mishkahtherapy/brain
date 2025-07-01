package update_meeting_url

import (
	"net/url"

	"github.com/mishkahtherapy/brain/core/domain"
	"github.com/mishkahtherapy/brain/core/ports"
	"github.com/mishkahtherapy/brain/core/usecases/common"
)

// Input struct defines parameters for updating a meeting URL
type Input struct {
	SessionID  domain.SessionID `json:"sessionId"`
	MeetingURL string           `json:"meetingUrl"`
}

// Usecase struct with required dependencies
type Usecase struct {
	sessionRepo ports.SessionRepository
}

// NewUsecase creates a new instance of the update meeting URL usecase
func NewUsecase(sessionRepo ports.SessionRepository) *Usecase {
	return &Usecase{sessionRepo: sessionRepo}
}

// Execute updates a session's meeting URL
func (u *Usecase) Execute(input Input) (*domain.Session, error) {
	// Validate input
	if input.SessionID == "" {
		return nil, common.ErrSessionIDIsRequired
	}
	if input.MeetingURL == "" {
		return nil, common.ErrMeetingURLIsRequired
	}

	// Validate meeting URL format
	if _, err := url.ParseRequestURI(input.MeetingURL); err != nil {
		return nil, common.ErrInvalidMeetingURL
	}

	// Get the current session
	session, err := u.sessionRepo.GetSessionByID(input.SessionID)
	if err != nil {
		return nil, common.ErrSessionNotFound
	}

	// Update meeting URL and timestamp
	session.MeetingURL = input.MeetingURL
	session.UpdatedAt = domain.NewUTCTimestamp()

	// Persist the change
	err = u.sessionRepo.UpdateMeetingURL(input.SessionID, input.MeetingURL)
	if err != nil {
		return nil, common.ErrFailedToUpdateMeetingURL
	}

	return session, nil
}
