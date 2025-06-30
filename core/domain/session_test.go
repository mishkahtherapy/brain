package domain

import (
	"strings"
	"testing"
	"time"
)

func TestSessionState_IsFinalState(t *testing.T) {
	tests := []struct {
		name  string
		state SessionState
		want  bool
	}{
		{"Planned state is not final", SessionStatePlanned, false},
		{"Done state is final", SessionStateDone, true},
		{"Rescheduled state is final", SessionStateRescheduled, true},
		{"Cancelled state is final", SessionStateCancelled, true},
		{"Refunded state is final", SessionStateRefunded, true},
		{"Empty state is not final", SessionState(""), false},
		{"Unknown state is not final", SessionState("unknown"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.state.IsFinalState(); got != tt.want {
				t.Errorf("SessionState.IsFinalState() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewSession(t *testing.T) {
	// Create a valid session
	now := NewUTCTimestamp()
	session := &Session{
		ID:          NewSessionID(),
		BookingID:   BookingID("booking_test"),
		TherapistID: TherapistID("therapist_test"),
		ClientID:    ClientID("client_test"),
		TimeSlotID:  TimeSlotID("timeslot_test"),
		StartTime:   now,
		PaidAmount:  5000, // $50.00
		Language:    SessionLanguageEnglish,
		State:       SessionStatePlanned,
		Notes:       "",
		MeetingURL:  "",
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	// Verify the session has the expected initial values
	t.Run("New session has correct initial state", func(t *testing.T) {
		if session.State != SessionStatePlanned {
			t.Errorf("New session should have 'planned' state, got %s", session.State)
		}
		if session.Notes != "" {
			t.Errorf("New session should have empty notes, got %s", session.Notes)
		}
		if session.MeetingURL != "" {
			t.Errorf("New session should have empty meeting URL, got %s", session.MeetingURL)
		}
	})

	// Verify the session ID format
	t.Run("Session ID has correct format", func(t *testing.T) {
		id := string(session.ID)
		if len(id) < 8 || id[:8] != "session_" {
			t.Errorf("Session ID should start with 'session_', got %s", id)
		}
	})
}

func TestSession_IsValidStateTransition(t *testing.T) {
	tests := []struct {
		name           string
		currentState   SessionState
		newState       SessionState
		expectedResult bool
	}{
		{"From Planned to Planned", SessionStatePlanned, SessionStatePlanned, true},
		{"From Planned to Done", SessionStatePlanned, SessionStateDone, true},
		{"From Planned to Rescheduled", SessionStatePlanned, SessionStateRescheduled, true},
		{"From Planned to Cancelled", SessionStatePlanned, SessionStateCancelled, true},
		{"From Planned to Refunded", SessionStatePlanned, SessionStateRefunded, true},

		{"From Done to Done", SessionStateDone, SessionStateDone, true},
		{"From Done to Planned", SessionStateDone, SessionStatePlanned, false},
		{"From Done to Rescheduled", SessionStateDone, SessionStateRescheduled, false},
		{"From Done to Cancelled", SessionStateDone, SessionStateCancelled, false},
		{"From Done to Refunded", SessionStateDone, SessionStateRefunded, false},

		{"From Rescheduled to Rescheduled", SessionStateRescheduled, SessionStateRescheduled, true},
		{"From Rescheduled to Planned", SessionStateRescheduled, SessionStatePlanned, false},
		{"From Rescheduled to Done", SessionStateRescheduled, SessionStateDone, false},
		{"From Rescheduled to Cancelled", SessionStateRescheduled, SessionStateCancelled, false},
		{"From Rescheduled to Refunded", SessionStateRescheduled, SessionStateRefunded, false},

		{"From Cancelled to Cancelled", SessionStateCancelled, SessionStateCancelled, true},
		{"From Cancelled to Planned", SessionStateCancelled, SessionStatePlanned, false},
		{"From Cancelled to Done", SessionStateCancelled, SessionStateDone, false},
		{"From Cancelled to Rescheduled", SessionStateCancelled, SessionStateRescheduled, false},
		{"From Cancelled to Refunded", SessionStateCancelled, SessionStateRefunded, false},

		{"From Refunded to Refunded", SessionStateRefunded, SessionStateRefunded, true},
		{"From Refunded to Planned", SessionStateRefunded, SessionStatePlanned, false},
		{"From Refunded to Done", SessionStateRefunded, SessionStateDone, false},
		{"From Refunded to Rescheduled", SessionStateRefunded, SessionStateRescheduled, false},
		{"From Refunded to Cancelled", SessionStateRefunded, SessionStateCancelled, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Session{State: tt.currentState}
			if got := s.IsValidStateTransition(tt.newState); got != tt.expectedResult {
				t.Errorf("Session.IsValidStateTransition() = %v, want %v", got, tt.expectedResult)
			}
		})
	}
}

func TestSession_CanUpdateField(t *testing.T) {
	tests := []struct {
		name      string
		state     SessionState
		field     string
		canUpdate bool
	}{
		{"Update notes in planned state", SessionStatePlanned, "notes", true},
		{"Update meetingUrl in planned state", SessionStatePlanned, "meetingUrl", true},
		{"Update startTime in planned state", SessionStatePlanned, "startTime", true},
		{"Update paidAmount in planned state", SessionStatePlanned, "paidAmount", true},

		{"Update notes in done state", SessionStateDone, "notes", true},
		{"Update meetingUrl in done state", SessionStateDone, "meetingUrl", true},
		{"Update startTime in done state", SessionStateDone, "startTime", false},
		{"Update paidAmount in done state", SessionStateDone, "paidAmount", false},

		{"Update notes in cancelled state", SessionStateCancelled, "notes", true},
		{"Update meetingUrl in cancelled state", SessionStateCancelled, "meetingUrl", true},
		{"Update startTime in cancelled state", SessionStateCancelled, "startTime", false},
		{"Update paidAmount in cancelled state", SessionStateCancelled, "paidAmount", false},

		{"Update notes in rescheduled state", SessionStateRescheduled, "notes", true},
		{"Update meetingUrl in rescheduled state", SessionStateRescheduled, "meetingUrl", true},
		{"Update startTime in rescheduled state", SessionStateRescheduled, "startTime", false},
		{"Update paidAmount in rescheduled state", SessionStateRescheduled, "paidAmount", false},

		{"Update notes in refunded state", SessionStateRefunded, "notes", true},
		{"Update meetingUrl in refunded state", SessionStateRefunded, "meetingUrl", true},
		{"Update startTime in refunded state", SessionStateRefunded, "startTime", false},
		{"Update paidAmount in refunded state", SessionStateRefunded, "paidAmount", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Session{State: tt.state}
			if got := s.CanUpdateField(tt.field); got != tt.canUpdate {
				t.Errorf("Session.CanUpdateField(%s) = %v, want %v", tt.field, got, tt.canUpdate)
			}
		})
	}
}

func TestSession_AppendNote(t *testing.T) {
	// Setup
	now := NewUTCTimestamp()
	session := &Session{
		ID:        NewSessionID(),
		Notes:     "",
		UpdatedAt: now,
	}

	// Test adding the first note
	t.Run("Append first note", func(t *testing.T) {
		originalUpdatedAt := session.UpdatedAt
		time.Sleep(1 * time.Second) // Ensure timestamp is different

		session.AppendNote("First test note")

		if session.Notes == "" {
			t.Error("Notes should not be empty after appending")
		}

		if !time.Time(session.UpdatedAt).After(time.Time(originalUpdatedAt)) {
			t.Error("UpdatedAt should be updated after appending note")
		}

		if !strings.Contains(session.Notes, "First test note") {
			t.Errorf("Notes should contain 'First test note', got: %s", session.Notes)
		}
	})

	// Test adding a second note
	t.Run("Append second note", func(t *testing.T) {
		originalNotes := session.Notes
		originalUpdatedAt := session.UpdatedAt
		time.Sleep(1 * time.Second) // Ensure timestamp is different

		session.AppendNote("Second test note")

		if len(session.Notes) <= len(originalNotes) {
			t.Error("Notes should be longer after appending second note")
		}

		if !time.Time(session.UpdatedAt).After(time.Time(originalUpdatedAt)) {
			t.Error("UpdatedAt should be updated after appending second note")
		}

		if !strings.Contains(session.Notes, "Second test note") {
			t.Errorf("Notes should contain 'Second test note', got: %s", session.Notes)
		}
	})

	// Verify format contains timestamp
	t.Run("Notes format includes timestamp", func(t *testing.T) {
		session := &Session{Notes: ""}
		session.AppendNote("Test note")

		// RFC3339 format has "T" and "Z" or timezone offset
		if len(session.Notes) < 20 || (session.Notes[10] != 'T' &&
			(session.Notes[19] != 'Z' && session.Notes[19] != '+' && session.Notes[19] != '-')) {
			t.Errorf("Notes format should include RFC3339 timestamp, got: %s", session.Notes)
		}
	})

	// Verify notes are separated properly
	t.Run("Multiple notes are separated properly", func(t *testing.T) {
		session := &Session{Notes: ""}
		session.AppendNote("First note")
		session.AppendNote("Second note")

		if len(session.Notes) < 40 {
			t.Errorf("Multiple notes should have sufficient length, got: %s", session.Notes)
		}

		// Check for double newline separator
		found := false
		for i := 0; i < len(session.Notes)-1; i++ {
			if session.Notes[i] == '\n' && session.Notes[i+1] == '\n' {
				found = true
				break
			}
		}

		if !found {
			t.Errorf("Multiple notes should be separated by double newline, got: %s", session.Notes)
		}
	})
}

func TestSessionLanguage(t *testing.T) {
	t.Run("Supported languages are defined", func(t *testing.T) {
		if SessionLanguageArabic != "arabic" {
			t.Errorf("Arabic language should be 'arabic', got %s", SessionLanguageArabic)
		}
		if SessionLanguageEnglish != "english" {
			t.Errorf("English language should be 'english', got %s", SessionLanguageEnglish)
		}
	})

	// Testing with an invalid language
	t.Run("Session with invalid language", func(t *testing.T) {
		now := NewUTCTimestamp()
		session := &Session{
			ID:          NewSessionID(),
			BookingID:   BookingID("booking_test"),
			TherapistID: TherapistID("therapist_test"),
			ClientID:    ClientID("client_test"),
			TimeSlotID:  TimeSlotID("timeslot_test"),
			StartTime:   now,
			PaidAmount:  5000,
			Language:    SessionLanguage("invalid"),
			State:       SessionStatePlanned,
			CreatedAt:   now,
			UpdatedAt:   now,
		}

		// Session creation should succeed even with invalid language
		// (validation should happen at the usecase level)
		if session == nil {
			t.Error("Session creation should succeed with invalid language")
		}

		if session.Language != "invalid" {
			t.Errorf("Session language should be 'invalid', got %s", session.Language)
		}
	})
}

func TestFullSessionLifecycle(t *testing.T) {
	// Create a new session
	now := NewUTCTimestamp()
	session := &Session{
		ID:          NewSessionID(),
		BookingID:   BookingID("booking_test"),
		TherapistID: TherapistID("therapist_test"),
		ClientID:    ClientID("client_test"),
		TimeSlotID:  TimeSlotID("timeslot_test"),
		StartTime:   now,
		PaidAmount:  5000,
		Language:    SessionLanguageEnglish,
		State:       SessionStatePlanned,
		Notes:       "",
		MeetingURL:  "",
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	// 1. Initial state checks
	t.Run("Initial state", func(t *testing.T) {
		if session.State != SessionStatePlanned {
			t.Errorf("Initial state should be planned, got %s", session.State)
		}
		if !session.CanUpdateField("startTime") {
			t.Error("Should be able to update startTime in planned state")
		}
	})

	// 2. Update meeting URL
	t.Run("Update meeting URL", func(t *testing.T) {
		session.MeetingURL = "https://zoom.us/meeting/123"
		if session.MeetingURL != "https://zoom.us/meeting/123" {
			t.Errorf("Meeting URL not updated correctly, got %s", session.MeetingURL)
		}
	})

	// 3. Add pre-session note
	t.Run("Add pre-session note", func(t *testing.T) {
		originalNotes := session.Notes
		session.AppendNote("Pre-session preparation complete")
		if session.Notes == originalNotes {
			t.Error("Notes should be updated after append")
		}
	})

	// 4. Transition to done state
	t.Run("Transition to done state", func(t *testing.T) {
		if !session.IsValidStateTransition(SessionStateDone) {
			t.Error("Should allow transition from planned to done")
		}
		session.State = SessionStateDone
		if session.State != SessionStateDone {
			t.Errorf("State should be done, got %s", session.State)
		}
	})

	// 5. Try invalid operations after completion
	t.Run("Operations after completion", func(t *testing.T) {
		// Should not allow changing to another state
		if session.IsValidStateTransition(SessionStateCancelled) {
			t.Error("Should not allow transition from done to cancelled")
		}

		// Should not allow updating certain fields
		if session.CanUpdateField("startTime") {
			t.Error("Should not allow updating startTime in done state")
		}
		if session.CanUpdateField("paidAmount") {
			t.Error("Should not allow updating paidAmount in done state")
		}

		// Should allow updating notes
		if !session.CanUpdateField("notes") {
			t.Error("Should allow updating notes in done state")
		}

		// Should allow updating meeting URL
		if !session.CanUpdateField("meetingUrl") {
			t.Error("Should allow updating meetingUrl in done state")
		}
	})

	// 6. Add post-session note
	t.Run("Add post-session note", func(t *testing.T) {
		originalNotes := session.Notes
		session.AppendNote("Session completed successfully")
		if session.Notes == originalNotes {
			t.Error("Notes should be updated after append")
		}
	})
}
