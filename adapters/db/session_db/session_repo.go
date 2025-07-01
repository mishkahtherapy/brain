package session_db

import (
	"database/sql"
	"errors"
	"log/slog"
	"time"

	"github.com/mishkahtherapy/brain/core/domain"
	"github.com/mishkahtherapy/brain/core/ports"
)

type SessionRepository struct {
	db ports.SQLDatabase
}

// Error definitions
var ErrSessionNotFound = errors.New("session not found")
var ErrSessionAlreadyExists = errors.New("session already exists")
var ErrSessionIDIsRequired = errors.New("session id is required")
var ErrSessionBookingIDIsRequired = errors.New("session booking id is required")
var ErrSessionTherapistIDIsRequired = errors.New("session therapist id is required")
var ErrSessionClientIDIsRequired = errors.New("session client id is required")
var ErrSessionTimeSlotIDIsRequired = errors.New("session timeslot id is required")
var ErrSessionStartTimeIsRequired = errors.New("session start time is required")
var ErrSessionPaidAmountIsRequired = errors.New("session paid amount is required")
var ErrSessionLanguageIsRequired = errors.New("session language is required")
var ErrSessionStateIsRequired = errors.New("session state is required")
var ErrSessionCreatedAtIsRequired = errors.New("session created at is required")
var ErrSessionUpdatedAtIsRequired = errors.New("session updated at is required")
var ErrFailedToGetSession = errors.New("failed to get session")
var ErrFailedToCreateSession = errors.New("failed to create session")
var ErrFailedToUpdateSession = errors.New("failed to update session")
var ErrInvalidDateRange = errors.New("invalid date range")

// NewSessionRepository creates a new session repository
func NewSessionRepository(db ports.SQLDatabase) *SessionRepository {
	return &SessionRepository{db: db}
}

// CreateSession creates a new session in the database
func (r *SessionRepository) CreateSession(session *domain.Session) error {
	// Validate required fields
	if session.ID == "" {
		return ErrSessionIDIsRequired
	}
	if session.BookingID == "" {
		return ErrSessionBookingIDIsRequired
	}
	if session.TherapistID == "" {
		return ErrSessionTherapistIDIsRequired
	}
	if session.ClientID == "" {
		return ErrSessionClientIDIsRequired
	}
	if session.TimeSlotID == "" {
		return ErrSessionTimeSlotIDIsRequired
	}
	if session.StartTime == (domain.UTCTimestamp{}) {
		return ErrSessionStartTimeIsRequired
	}
	if session.PaidAmount <= 0 {
		return ErrSessionPaidAmountIsRequired
	}
	if session.Language == "" {
		return ErrSessionLanguageIsRequired
	}
	if session.State == "" {
		return ErrSessionStateIsRequired
	}
	if session.CreatedAt == (domain.UTCTimestamp{}) {
		return ErrSessionCreatedAtIsRequired
	}
	if session.UpdatedAt == (domain.UTCTimestamp{}) {
		return ErrSessionUpdatedAtIsRequired
	}

	query := `
		INSERT INTO sessions (
			id, booking_id, therapist_id, client_id, timeslot_id, 
			start_time, paid_amount, language, state, notes, 
			meeting_url, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err := r.db.Exec(
		query,
		session.ID,
		session.BookingID,
		session.TherapistID,
		session.ClientID,
		session.TimeSlotID,
		session.StartTime,
		session.PaidAmount,
		session.Language,
		session.State,
		session.Notes,
		session.MeetingURL,
		session.CreatedAt,
		session.UpdatedAt,
	)

	if err != nil {
		slog.Error("error creating session", "error", err)
		return ErrFailedToCreateSession
	}

	return nil
}

// GetSessionByID retrieves a session by its ID
func (r *SessionRepository) GetSessionByID(id domain.SessionID) (*domain.Session, error) {
	if id == "" {
		return nil, ErrSessionIDIsRequired
	}

	query := `
		SELECT id, booking_id, therapist_id, client_id, timeslot_id, 
		       start_time, paid_amount, language, state, notes, 
		       meeting_url, created_at, updated_at
		FROM sessions
		WHERE id = ?
	`

	row := r.db.QueryRow(query, id)
	session := &domain.Session{}
	err := row.Scan(
		&session.ID,
		&session.BookingID,
		&session.TherapistID,
		&session.ClientID,
		&session.TimeSlotID,
		&session.StartTime,
		&session.PaidAmount,
		&session.Language,
		&session.State,
		&session.Notes,
		&session.MeetingURL,
		&session.CreatedAt,
		&session.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrSessionNotFound
		}
		slog.Error("error getting session by id", "error", err)
		return nil, ErrFailedToGetSession
	}

	return session, nil
}

// UpdateSessionState updates a session's state
func (r *SessionRepository) UpdateSessionState(id domain.SessionID, state domain.SessionState) error {
	if id == "" {
		return ErrSessionIDIsRequired
	}
	if state == "" {
		return ErrSessionStateIsRequired
	}

	// First get the session to check if the state transition is valid
	session, err := r.GetSessionByID(id)
	if err != nil {
		return err
	}

	// Check if the state transition is valid
	if !session.IsValidStateTransition(state) {
		return errors.New("invalid state transition")
	}

	// Update the state and timestamp
	updatedAt := domain.NewUTCTimestamp()
	query := `
		UPDATE sessions
		SET state = ?, updated_at = ?
		WHERE id = ?
	`

	result, err := r.db.Exec(query, state, updatedAt, id)
	if err != nil {
		slog.Error("error updating session state", "error", err)
		return ErrFailedToUpdateSession
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		slog.Error("error getting rows affected after update", "error", err)
		return ErrFailedToUpdateSession
	}

	if rowsAffected == 0 {
		return ErrSessionNotFound
	}

	return nil
}

// UpdateSessionNotes updates a session's notes
func (r *SessionRepository) UpdateSessionNotes(id domain.SessionID, notes string) error {
	if id == "" {
		return ErrSessionIDIsRequired
	}

	// First get the session to properly update the notes
	session, err := r.GetSessionByID(id)
	if err != nil {
		return err
	}

	// Check if notes can be updated
	if !session.CanUpdateField("notes") {
		return errors.New("cannot update notes in current state")
	}

	// Update the notes and timestamp
	updatedAt := domain.NewUTCTimestamp()
	query := `
		UPDATE sessions
		SET notes = ?, updated_at = ?
		WHERE id = ?
	`

	result, err := r.db.Exec(query, notes, updatedAt, id)
	if err != nil {
		slog.Error("error updating session notes", "error", err)
		return ErrFailedToUpdateSession
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		slog.Error("error getting rows affected after update", "error", err)
		return ErrFailedToUpdateSession
	}

	if rowsAffected == 0 {
		return ErrSessionNotFound
	}

	return nil
}

// UpdateMeetingURL updates a session's meeting URL
func (r *SessionRepository) UpdateMeetingURL(id domain.SessionID, meetingURL string) error {
	if id == "" {
		return ErrSessionIDIsRequired
	}

	// First get the session to check if the meeting URL can be updated
	session, err := r.GetSessionByID(id)
	if err != nil {
		return err
	}

	// Check if meeting URL can be updated
	if !session.CanUpdateField("meetingUrl") {
		return errors.New("cannot update meeting URL in current state")
	}

	// Update the meeting URL and timestamp
	updatedAt := domain.NewUTCTimestamp()
	query := `
		UPDATE sessions
		SET meeting_url = ?, updated_at = ?
		WHERE id = ?
	`

	result, err := r.db.Exec(query, meetingURL, updatedAt, id)
	if err != nil {
		slog.Error("error updating session meeting URL", "error", err)
		return ErrFailedToUpdateSession
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		slog.Error("error getting rows affected after update", "error", err)
		return ErrFailedToUpdateSession
	}

	if rowsAffected == 0 {
		return ErrSessionNotFound
	}

	return nil
}

// ListSessionsByTherapist lists all sessions for a therapist
func (r *SessionRepository) ListSessionsByTherapist(therapistID domain.TherapistID) ([]*domain.Session, error) {
	if therapistID == "" {
		return nil, ErrSessionTherapistIDIsRequired
	}

	query := `
		SELECT id, booking_id, therapist_id, client_id, timeslot_id, 
		       start_time, paid_amount, language, state, notes, 
		       meeting_url, created_at, updated_at
		FROM sessions
		WHERE therapist_id = ?
		ORDER BY start_time ASC
	`

	rows, err := r.db.Query(query, therapistID)
	if err != nil {
		slog.Error("error listing sessions by therapist", "error", err)
		return nil, ErrFailedToGetSession
	}
	defer rows.Close()

	return r.scanSessions(rows)
}

// ListSessionsByClient lists all sessions for a client
func (r *SessionRepository) ListSessionsByClient(clientID domain.ClientID) ([]*domain.Session, error) {
	if clientID == "" {
		return nil, ErrSessionClientIDIsRequired
	}

	query := `
		SELECT id, booking_id, therapist_id, client_id, timeslot_id, 
		       start_time, paid_amount, language, state, notes, 
		       meeting_url, created_at, updated_at
		FROM sessions
		WHERE client_id = ?
		ORDER BY start_time ASC
	`

	rows, err := r.db.Query(query, clientID)
	if err != nil {
		slog.Error("error listing sessions by client", "error", err)
		return nil, ErrFailedToGetSession
	}
	defer rows.Close()

	return r.scanSessions(rows)
}

// ListSessionsAdmin lists all sessions within a date range for admin purposes
func (r *SessionRepository) ListSessionsAdmin(startDate, endDate time.Time) ([]*domain.Session, error) {
	// Validate date range
	if startDate.After(endDate) {
		return nil, ErrInvalidDateRange
	}

	query := `
		SELECT id, booking_id, therapist_id, client_id, timeslot_id, 
		       start_time, paid_amount, language, state, notes, 
		       meeting_url, created_at, updated_at
		FROM sessions
		WHERE start_time >= ? AND start_time <= ?
		ORDER BY start_time ASC
	`

	rows, err := r.db.Query(query, startDate, endDate)
	if err != nil {
		slog.Error("error listing sessions for admin", "error", err)
		return nil, ErrFailedToGetSession
	}
	defer rows.Close()

	return r.scanSessions(rows)
}

// Helper method to scan multiple session rows
func (r *SessionRepository) scanSessions(rows *sql.Rows) ([]*domain.Session, error) {
	sessions := make([]*domain.Session, 0)
	for rows.Next() {
		session := &domain.Session{}
		err := rows.Scan(
			&session.ID,
			&session.BookingID,
			&session.TherapistID,
			&session.ClientID,
			&session.TimeSlotID,
			&session.StartTime,
			&session.PaidAmount,
			&session.Language,
			&session.State,
			&session.Notes,
			&session.MeetingURL,
			&session.CreatedAt,
			&session.UpdatedAt,
		)
		if err != nil {
			slog.Error("error scanning session", "error", err)
			return nil, ErrFailedToGetSession
		}
		sessions = append(sessions, session)
	}
	return sessions, nil
}
