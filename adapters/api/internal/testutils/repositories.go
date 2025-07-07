package testutils

import (
	"time"

	"github.com/mishkahtherapy/brain/core/domain"
	"github.com/mishkahtherapy/brain/core/domain/client"
	"github.com/mishkahtherapy/brain/core/domain/timeslot"
	"github.com/mishkahtherapy/brain/core/ports"
)

// TestSessionRepository is a minimal test implementation of the session repository
type TestSessionRepository struct {
	db ports.SQLDatabase
}

func NewTestSessionRepository(db ports.SQLDatabase) ports.SessionRepository {
	return &TestSessionRepository{db: db}
}

func (r *TestSessionRepository) CreateSession(session *domain.Session) error {
	return nil // Just return success for test
}

func (r *TestSessionRepository) GetSessionByID(id domain.SessionID) (*domain.Session, error) {
	return nil, nil
}

func (r *TestSessionRepository) UpdateSessionState(id domain.SessionID, state domain.SessionState) error {
	return nil
}

func (r *TestSessionRepository) UpdateSessionNotes(id domain.SessionID, notes string) error {
	return nil
}

func (r *TestSessionRepository) UpdateMeetingURL(id domain.SessionID, meetingURL string) error {
	return nil
}

func (r *TestSessionRepository) ListSessionsByTherapist(therapistID domain.TherapistID) ([]*domain.Session, error) {
	return nil, nil
}

func (r *TestSessionRepository) ListSessionsByClient(clientID domain.ClientID) ([]*domain.Session, error) {
	return nil, nil
}

func (r *TestSessionRepository) ListSessionsAdmin(startDate, endDate time.Time) ([]*domain.Session, error) {
	return nil, nil
}

// TestClientRepository is a minimal test implementation that can read clients
type TestClientRepository struct {
	db ports.SQLDatabase
}

func NewTestClientRepository(db ports.SQLDatabase) ports.ClientRepository {
	return &TestClientRepository{db: db}
}

func (r *TestClientRepository) GetByID(id domain.ClientID) (*client.Client, error) {
	query := `SELECT id, name, whatsapp_number, created_at, updated_at FROM clients WHERE id = ?`
	row := r.db.QueryRow(query, id)

	var client client.Client
	err := row.Scan(&client.ID, &client.Name, &client.WhatsAppNumber, &client.CreatedAt, &client.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &client, nil
}

func (r *TestClientRepository) Create(client *client.Client) error { return nil }
func (r *TestClientRepository) Update(client *client.Client) error { return nil }
func (r *TestClientRepository) Delete(id domain.ClientID) error    { return nil }
func (r *TestClientRepository) UpdateTimezoneOffset(id domain.ClientID, offsetMinutes domain.TimezoneOffset) error {
	return nil
}
func (r *TestClientRepository) GetByWhatsAppNumber(whatsappNumber domain.WhatsAppNumber) (*client.Client, error) {
	return nil, nil
}
func (r *TestClientRepository) List() ([]*client.Client, error) { return nil, nil }

// TestTimeSlotRepository is a minimal test implementation that can read timeslots
type TestTimeSlotRepository struct {
	db ports.SQLDatabase
}

func NewTestTimeSlotRepository(db ports.SQLDatabase) ports.TimeSlotRepository {
	return &TestTimeSlotRepository{db: db}
}

func (r *TestTimeSlotRepository) GetByID(id string) (*timeslot.TimeSlot, error) {
	query := `SELECT id, therapist_id, day_of_week, start_time, duration_minutes, pre_session_buffer, post_session_buffer, is_active, created_at, updated_at FROM time_slots WHERE id = ?`
	row := r.db.QueryRow(query, id)

	var timeSlot timeslot.TimeSlot
	err := row.Scan(&timeSlot.ID, &timeSlot.TherapistID, &timeSlot.DayOfWeek, &timeSlot.StartTime, &timeSlot.DurationMinutes, &timeSlot.PreSessionBuffer, &timeSlot.PostSessionBuffer, &timeSlot.IsActive, &timeSlot.CreatedAt, &timeSlot.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &timeSlot, nil
}

func (r *TestTimeSlotRepository) Create(timeslot *timeslot.TimeSlot) error { return nil }
func (r *TestTimeSlotRepository) Update(timeslot *timeslot.TimeSlot) error { return nil }
func (r *TestTimeSlotRepository) Delete(id string) error                   { return nil }
func (r *TestTimeSlotRepository) ListByTherapist(therapistID string) ([]*timeslot.TimeSlot, error) {
	return nil, nil
}
func (r *TestTimeSlotRepository) ListByDay(therapistID string, day string) ([]*timeslot.TimeSlot, error) {
	return nil, nil
}
func (r *TestTimeSlotRepository) BulkToggleByTherapistID(therapistID string, isActive bool) error {
	return nil
}
