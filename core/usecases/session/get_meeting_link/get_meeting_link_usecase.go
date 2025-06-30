package get_meeting_link

import (
	"errors"

	"github.com/mishkahtherapy/brain/core/domain"
	"github.com/mishkahtherapy/brain/core/ports"
)

// Error definitions
var ErrSessionNotFound = errors.New("session not found")
var ErrSessionIDIsRequired = errors.New("session ID is required")
var ErrMeetingURLNotSet = errors.New("meeting URL is not set for this session")

// Input struct defines parameters for getting a meeting link
type Input struct {
	SessionID domain.SessionID `json:"sessionId"`
}

// Output struct defines the result of the operation
type Output struct {
	MeetingURL string `json:"meetingUrl"`
}

// Usecase struct with required dependencies
type Usecase struct {
	sessionRepo ports.SessionRepository
}

// NewUsecase creates a new instance of the get meeting link usecase
func NewUsecase(sessionRepo ports.SessionRepository) *Usecase {
	return &Usecase{sessionRepo: sessionRepo}
}

// Execute retrieves the meeting URL for a specific session
func (u *Usecase) Execute(input Input) (*Output, error) {
	// Validate input
	if input.SessionID == "" {
		return nil, ErrSessionIDIsRequired
	}

	// Retrieve session from repository
	session, err := u.sessionRepo.GetSessionByID(input.SessionID)
	if err != nil {
		return nil, ErrSessionNotFound
	}

	// Check if meeting URL is set
	if session.MeetingURL == "" {
		return nil, ErrMeetingURLNotSet
	}

	// Return meeting URL
	return &Output{
		MeetingURL: session.MeetingURL,
	}, nil
}
