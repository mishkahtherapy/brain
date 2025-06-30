package list_sessions_admin

import (
	"errors"
	"time"

	"github.com/mishkahtherapy/brain/core/domain"
	"github.com/mishkahtherapy/brain/core/ports"
)

// Error definitions
var ErrInvalidDateRange = errors.New("start date must be before end date")
var ErrFailedToListSessions = errors.New("failed to list sessions")

// Input struct defines parameters for listing sessions with admin privileges
type Input struct {
	StartDate time.Time `json:"startDate"`
	EndDate   time.Time `json:"endDate"`
}

// Usecase struct with required dependencies
type Usecase struct {
	sessionRepo ports.SessionRepository
}

// NewUsecase creates a new instance of the list sessions admin usecase
func NewUsecase(sessionRepo ports.SessionRepository) *Usecase {
	return &Usecase{sessionRepo: sessionRepo}
}

// Execute retrieves all sessions within the specified time range
func (u *Usecase) Execute(input Input) ([]*domain.Session, error) {
	// Validate input
	if input.StartDate.After(input.EndDate) {
		return nil, ErrInvalidDateRange
	}

	// Set default time range if not provided
	// If zero time, use a large range (past 1 year to future 1 year)
	if input.StartDate.IsZero() {
		input.StartDate = time.Now().AddDate(-1, 0, 0) // 1 year ago
	}
	if input.EndDate.IsZero() {
		input.EndDate = time.Now().AddDate(1, 0, 0) // 1 year from now
	}

	// Retrieve sessions from repository
	sessions, err := u.sessionRepo.ListSessionsAdmin(input.StartDate, input.EndDate)
	if err != nil {
		return nil, ErrFailedToListSessions
	}

	// Return sessions (empty slice if none found)
	return sessions, nil
}
