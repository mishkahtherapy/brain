package domain

type SessionState string
type SessionLanguage string

const (
	SessionStatePlanned     SessionState = "planned"
	SessionStateDone        SessionState = "done"
	SessionStateRescheduled SessionState = "rescheduled"
	SessionStateCancelled   SessionState = "cancelled"
	SessionStateRefunded    SessionState = "refunded"
)

const (
	SessionLanguageArabic  SessionLanguage = "arabic"
	SessionLanguageEnglish SessionLanguage = "english"
)

// IsFinalState returns true if the session state is a final state
// (done, rescheduled, cancelled, refunded)
func (s SessionState) IsFinalState() bool {
	return s == SessionStateDone ||
		s == SessionStateRescheduled ||
		s == SessionStateCancelled ||
		s == SessionStateRefunded
}

// Session represents a confirmed therapy session derived from a booking
type Session struct {
	ID                   SessionID       `json:"id"`
	BookingID            BookingID       `json:"bookingId"`
	TherapistID          TherapistID     `json:"therapistId"`
	ClientID             ClientID        `json:"clientId"`
	TimeSlotID           TimeSlotID      `json:"timeSlotId"`
	StartTime            UTCTimestamp    `json:"startTime"`
	Duration             DurationMinutes `json:"duration"`
	ClientTimezoneOffset TimezoneOffset  `json:"clientTimezoneOffset"`
	PaidAmount           int             `json:"paidAmount"` // USD cents
	Language             SessionLanguage `json:"language"`
	State                SessionState    `json:"state"`
	Notes                string          `json:"notes"` // delays, special notes, ...etc.
	MeetingURL           string          `json:"meetingUrl,omitempty"`
	CreatedAt            UTCTimestamp    `json:"createdAt"`
	UpdatedAt            UTCTimestamp    `json:"updatedAt"`
}

// IsValidStateTransition checks if a state transition is valid based on the rules:
// - Any state can only transition once to a final state
// - Final states cannot transition to other states
func (s *Session) IsValidStateTransition(newState SessionState) bool {
	// If current state is final, only allow transition to itself (for notes updates)
	if s.State.IsFinalState() {
		return s.State == newState
	}

	// From planned, can transition to any state
	if s.State == SessionStatePlanned {
		return true
	}

	return false
}

// CanUpdateField checks if the given field can be updated based on the session state
// In final states, only notes and meetingUrl can be updated
func (s *Session) CanUpdateField(field string) bool {
	if !s.State.IsFinalState() {
		return true
	}

	return field == "notes" || field == "meetingUrl"
}

// AppendNote adds a note with timestamp, preserving previous notes
func (s *Session) AppendNote(note string) {
	timestamp := NewUTCTimestamp()
	if s.Notes != "" {
		s.Notes += "\n\n"
	}
	s.Notes += timestamp.String() + ": " + note
	s.UpdatedAt = timestamp
}
