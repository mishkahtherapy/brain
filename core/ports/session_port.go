package ports

import (
	"time"

	"github.com/mishkahtherapy/brain/core/domain"
)

type SessionRepository interface {
	CreateSession(session *domain.Session) error
	GetSessionByID(id domain.SessionID) (*domain.Session, error)
	UpdateSessionState(id domain.SessionID, state domain.SessionState) error
	UpdateSessionNotes(id domain.SessionID, notes string) error
	UpdateMeetingURL(id domain.SessionID, meetingURL string) error
	ListSessionsByTherapist(therapistID domain.TherapistID) ([]*domain.Session, error)
	ListSessionsByClient(clientID domain.ClientID) ([]*domain.Session, error)
	ListSessionsAdmin(startDate, endDate time.Time) ([]*domain.Session, error)
}
