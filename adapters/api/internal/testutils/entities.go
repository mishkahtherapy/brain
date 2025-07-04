package testutils

import (
	"testing"
	"time"

	"github.com/mishkahtherapy/brain/core/domain"
	"github.com/mishkahtherapy/brain/core/ports"
)

// CreateTestTherapist inserts a test therapist and returns the ID
func CreateTestTherapist(t *testing.T, database ports.SQLDatabase) domain.TherapistID {
	return CreateTestTherapistWithName(t, database, "Dr. Test Therapist")
}

// CreateTestTherapistWithName inserts a test therapist with custom name
func CreateTestTherapistWithName(t *testing.T, database ports.SQLDatabase, name string) domain.TherapistID {
	now := time.Now().UTC()
	therapistID := domain.NewTherapistID()

	_, err := database.Exec(`
		INSERT INTO therapists (id, name, email, phone_number, whatsapp_number, speaks_english, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`, therapistID, name, "test@example.com", "+1234567890", "+1234567890", true, now, now)

	if err != nil {
		t.Fatalf("Failed to insert test therapist: %v", err)
	}

	return therapistID
}

// CreateTestClient inserts a test client and returns the ID
func CreateTestClient(t *testing.T, database ports.SQLDatabase) domain.ClientID {
	return CreateTestClientWithName(t, database, "Test Client")
}

// CreateTestClientWithName inserts a test client with custom name
func CreateTestClientWithName(t *testing.T, database ports.SQLDatabase, name string) domain.ClientID {
	now := time.Now().UTC()
	clientID := domain.NewClientID()

	_, err := database.Exec(`
		INSERT INTO clients (id, name, whatsapp_number, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?)
	`, clientID, name, "+1234567891", now, now)

	if err != nil {
		t.Fatalf("Failed to insert test client: %v", err)
	}

	return clientID
}

// CreateTestSpecialization inserts a test specialization and returns the ID
func CreateTestSpecialization(t *testing.T, database ports.SQLDatabase) domain.SpecializationID {
	return CreateTestSpecializationWithName(t, database, "Test Specialization")
}

// CreateTestSpecializationWithName inserts a test specialization with custom name
func CreateTestSpecializationWithName(t *testing.T, database ports.SQLDatabase, name string) domain.SpecializationID {
	now := time.Now().UTC()
	specializationID := domain.NewSpecializationID()

	_, err := database.Exec(`
		INSERT INTO specializations (id, name, created_at, updated_at)
		VALUES (?, ?, ?, ?)
	`, specializationID, name, now, now)

	if err != nil {
		t.Fatalf("Failed to insert test specialization: %v", err)
	}

	return specializationID
}

// CreateTestTimeSlot inserts a test timeslot and returns the ID
func CreateTestTimeSlot(t *testing.T, database ports.SQLDatabase, therapistID domain.TherapistID) domain.TimeSlotID {
	now := time.Now().UTC()
	timeSlotID := domain.NewTimeSlotID()

	_, err := database.Exec(`
		INSERT INTO time_slots (id, therapist_id, day_of_week, start_time, duration_minutes, pre_session_buffer, post_session_buffer, is_active, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, timeSlotID, therapistID, "Monday", "10:00", 60, 0, 0, true, now, now)

	if err != nil {
		t.Fatalf("Failed to insert test time slot: %v", err)
	}

	return timeSlotID
}

// CreateTestTimeSlotCustom inserts a test timeslot with custom parameters
func CreateTestTimeSlotCustom(t *testing.T, database ports.SQLDatabase, therapistID domain.TherapistID, dayOfWeek, startTime string, durationMinutes int, isActive bool) domain.TimeSlotID {
	now := time.Now().UTC()
	timeSlotID := domain.NewTimeSlotID()

	_, err := database.Exec(`
		INSERT INTO time_slots (id, therapist_id, day_of_week, start_time, duration_minutes, pre_session_buffer, post_session_buffer, is_active, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, timeSlotID, therapistID, dayOfWeek, startTime, durationMinutes, 0, 0, isActive, now, now)

	if err != nil {
		t.Fatalf("Failed to insert test time slot: %v", err)
	}

	return timeSlotID
}

// LinkTherapistSpecialization creates a therapist-specialization relationship
func LinkTherapistSpecialization(t *testing.T, database ports.SQLDatabase, therapistID domain.TherapistID, specializationID domain.SpecializationID) {
	now := time.Now().UTC()
	id := "therapist_spec_" + string(therapistID) + "_" + string(specializationID)

	_, err := database.Exec(`
		INSERT INTO therapist_specializations (id, therapist_id, specialization_id, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?)
	`, id, therapistID, specializationID, now, now)

	if err != nil {
		t.Fatalf("Failed to link therapist specialization: %v", err)
	}
}

// FullTestData contains all the basic test entities
type FullTestData struct {
	TherapistID      domain.TherapistID
	ClientID         domain.ClientID
	TimeSlotID       domain.TimeSlotID
	SpecializationID domain.SpecializationID
}

// CreateFullTestData creates a complete set of test entities with relationships
func CreateFullTestData(t *testing.T, database ports.SQLDatabase) *FullTestData {
	therapistID := CreateTestTherapist(t, database)
	clientID := CreateTestClient(t, database)
	specializationID := CreateTestSpecialization(t, database)
	timeSlotID := CreateTestTimeSlot(t, database, therapistID)

	// Link therapist to specialization
	LinkTherapistSpecialization(t, database, therapistID, specializationID)

	return &FullTestData{
		TherapistID:      therapistID,
		ClientID:         clientID,
		TimeSlotID:       timeSlotID,
		SpecializationID: specializationID,
	}
}
