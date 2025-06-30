package update_meeting_url

import (
	"errors"
	"net/url"

	"github.com/mishkahtherapy/brain/core/domain"
	"github.com/mishkahtherapy/brain/core/ports"
)

// Error definitions
var ErrSessionNotFound = errors.New("session not found")
var ErrSessionIDIsRequired = errors.New("session ID is required")
var ErrMeetingURLIsRequired = errors.New("meeting URL is required")
var ErrInvalidMeetingURL = errors.New("invalid meeting URL format")
var ErrFailedToUpdateMeetingURL = errors.New("failed to update meeting URL")

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
		return nil, ErrSessionIDIsRequired
	}
	if input.MeetingURL == "" {
		return nil, ErrMeetingURLIsRequired
	}

	// Validate meeting URL format
	if _, err := url.ParseRequestURI(input.MeetingURL); err != nil {
		return nil, ErrInvalidMeetingURL
	}

	// Get the current session
	session, err := u.sessionRepo.GetSessionByID(input.SessionID)
	if err != nil {
		return nil, ErrSessionNotFound
	}

	// Update meeting URL and timestamp
	session.MeetingURL = input.MeetingURL
	session.UpdatedAt = domain.NewUTCTimestamp()

	// Persist the change
	err = u.sessionRepo.UpdateMeetingURL(input.SessionID, input.MeetingURL)
	if err != nil {
		return nil, ErrFailedToUpdateMeetingURL
	}

	return session, nil
}
