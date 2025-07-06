package testutils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/mishkahtherapy/brain/core/domain"
	"github.com/mishkahtherapy/brain/core/domain/booking"
	"github.com/mishkahtherapy/brain/core/ports"
)

// HTTPTestUtils provides helper functions for HTTP testing
type HTTPTestUtils struct {
	mux *http.ServeMux
}

// NewHTTPTestUtils creates a new HTTPTestUtils instance
func NewHTTPTestUtils(mux *http.ServeMux) *HTTPTestUtils {
	return &HTTPTestUtils{mux: mux}
}

// MakeRequest creates and executes an HTTP request, returning the response
func (h *HTTPTestUtils) MakeRequest(method, path string, body interface{}) *httptest.ResponseRecorder {
	var bodyBytes []byte
	var err error

	if body != nil {
		bodyBytes, err = json.Marshal(body)
		if err != nil {
			panic(err)
		}
	}

	req := httptest.NewRequest(method, path, bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	h.mux.ServeHTTP(rec, req)
	return rec
}

// AssertStatus asserts that the response has the expected HTTP status code
func (h *HTTPTestUtils) AssertStatus(t *testing.T, rec *httptest.ResponseRecorder, expectedStatus int) {
	if rec.Code != expectedStatus {
		t.Errorf("Expected status %d, got %d. Body: %s", expectedStatus, rec.Code, rec.Body.String())
	}
}

// ParseResponse parses the response body into the given type
func (h *HTTPTestUtils) ParseResponse(t *testing.T, rec *httptest.ResponseRecorder, v interface{}) {
	if err := json.Unmarshal(rec.Body.Bytes(), v); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}
}

// DatabaseTestUtils provides helper functions for database operations in tests
type DatabaseTestUtils struct {
	db ports.SQLDatabase
}

// NewDatabaseTestUtils creates a new DatabaseTestUtils instance
func NewDatabaseTestUtils(db ports.SQLDatabase) *DatabaseTestUtils {
	return &DatabaseTestUtils{db: db}
}

// CreateTestTherapist creates a test therapist and returns its ID
func (d *DatabaseTestUtils) CreateTestTherapist(t *testing.T, name, email, phone, whatsapp string) domain.TherapistID {
	therapistID := domain.NewTherapistID()
	now := time.Now().UTC()

	_, err := d.db.Exec(`
		INSERT INTO therapists (id, name, email, phone_number, whatsapp_number, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`, therapistID, name, email, phone, whatsapp, now, now)

	if err != nil {
		t.Fatalf("Failed to create test therapist: %v", err)
	}

	return therapistID
}

// CreateTestClient creates a test client and returns its ID
func (d *DatabaseTestUtils) CreateTestClient(t *testing.T, name, whatsapp, timezone string) domain.ClientID {
	clientID := domain.NewClientID()
	now := time.Now().UTC()

	_, err := d.db.Exec(`
		INSERT INTO clients (id, name, whatsapp_number, timezone, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?)
	`, clientID, name, whatsapp, timezone, now, now)

	if err != nil {
		t.Fatalf("Failed to create test client: %v", err)
	}

	return clientID
}

// CreateTestTimeSlot creates a test time slot and returns its ID
func (d *DatabaseTestUtils) CreateTestTimeSlot(t *testing.T, therapistID domain.TherapistID, dayOfWeek, startTime string, durationMinutes int) domain.TimeSlotID {
	timeSlotID := domain.NewTimeSlotID()
	now := time.Now().UTC()

	_, err := d.db.Exec(`
		INSERT INTO time_slots (id, therapist_id, day_of_week, start_time, duration_minutes, pre_session_buffer, post_session_buffer, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, timeSlotID, therapistID, dayOfWeek, startTime, durationMinutes, 0, 0, now, now)

	if err != nil {
		t.Fatalf("Failed to create test time slot: %v", err)
	}

	return timeSlotID
}

// CreateTestSpecialization creates a test specialization and returns its ID
func (d *DatabaseTestUtils) CreateTestSpecialization(t *testing.T, name string) domain.SpecializationID {
	specializationID := domain.NewSpecializationID()
	now := time.Now().UTC()

	_, err := d.db.Exec(`
		INSERT INTO specializations (id, name, created_at, updated_at)
		VALUES (?, ?, ?, ?)
	`, specializationID, name, now, now)

	if err != nil {
		t.Fatalf("Failed to create test specialization: %v", err)
	}

	return specializationID
}

// LinkTherapistToSpecialization links a therapist to a specialization
func (d *DatabaseTestUtils) LinkTherapistToSpecialization(t *testing.T, therapistID domain.TherapistID, specializationID domain.SpecializationID) {
	now := time.Now().UTC()

	_, err := d.db.Exec(`
		INSERT INTO therapist_specializations (id, therapist_id, specialization_id, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?)
	`, "therapist_spec_test_"+string(therapistID), therapistID, specializationID, now, now)

	if err != nil {
		t.Fatalf("Failed to link therapist to specialization: %v", err)
	}
}

// BookingTestData represents the test data needed for booking tests
type BookingTestData struct {
	TherapistID      domain.TherapistID
	ClientID         domain.ClientID
	TimeSlotID       domain.TimeSlotID
	SpecializationID domain.SpecializationID
}

// CreateBookingTestData creates all necessary test data for booking tests
func (d *DatabaseTestUtils) CreateBookingTestData(t *testing.T) *BookingTestData {
	// Create specialization with unique name
	specializationID := d.CreateTestSpecialization(t, fmt.Sprintf("Test Specialization %d", rand.Intn(10000)))

	// Create therapist with unique email
	therapistID := d.CreateTestTherapist(t, "Dr. Test Therapist", fmt.Sprintf("test.therapist+%d@example.com", rand.Intn(10000)), "+1234567890", "+1234567890")

	// Link therapist to specialization
	d.LinkTherapistToSpecialization(t, therapistID, specializationID)

	// Create client with unique WhatsApp number
	clientID := d.CreateTestClient(t, "Test Client", fmt.Sprintf("+1234567%d", rand.Intn(10000)), "UTC")

	// Create time slot
	timeSlotID := d.CreateTestTimeSlot(t, therapistID, "Monday", "10:00", 60)

	return &BookingTestData{
		TherapistID:      therapistID,
		ClientID:         clientID,
		TimeSlotID:       timeSlotID,
		SpecializationID: specializationID,
	}
}

// BookingTestUtils provides comprehensive utilities for booking tests
type BookingTestUtils struct {
	HTTP     *HTTPTestUtils
	Database *DatabaseTestUtils
}

// NewBookingTestUtils creates a new BookingTestUtils instance
func NewBookingTestUtils(mux *http.ServeMux, db ports.SQLDatabase) *BookingTestUtils {
	return &BookingTestUtils{
		HTTP:     NewHTTPTestUtils(mux),
		Database: NewDatabaseTestUtils(db),
	}
}

// CreateBookingRequest creates a booking request with the given data
func (b *BookingTestUtils) CreateBookingRequest(therapistID domain.TherapistID, clientID domain.ClientID, timeSlotID domain.TimeSlotID, startTime, timezone string) map[string]interface{} {
	return map[string]interface{}{
		"therapistId": therapistID,
		"clientId":    clientID,
		"timeSlotId":  timeSlotID,
		"startTime":   startTime,
		"timezone":    timezone,
	}
}

// CreateBooking makes a booking creation request and returns the response
func (b *BookingTestUtils) CreateBooking(t *testing.T, requestData map[string]interface{}) (*httptest.ResponseRecorder, *booking.Booking) {
	rec := b.HTTP.MakeRequest("POST", "/api/v1/bookings", requestData)

	var createdBooking booking.Booking
	if rec.Code == http.StatusCreated {
		b.HTTP.ParseResponse(t, rec, &createdBooking)
	}

	return rec, &createdBooking
}

// AssertBookingCreated asserts that a booking was successfully created
func (b *BookingTestUtils) AssertBookingCreated(t *testing.T, rec *httptest.ResponseRecorder, expectedTherapistID domain.TherapistID, expectedClientID domain.ClientID, expectedTimeSlotID domain.TimeSlotID) {
	b.HTTP.AssertStatus(t, rec, http.StatusCreated)

	var createdBooking booking.Booking
	b.HTTP.ParseResponse(t, rec, &createdBooking)

	if createdBooking.TherapistID != expectedTherapistID {
		t.Errorf("Expected therapist ID %s, got %s", expectedTherapistID, createdBooking.TherapistID)
	}
	if createdBooking.ClientID != expectedClientID {
		t.Errorf("Expected client ID %s, got %s", expectedClientID, createdBooking.ClientID)
	}
	if createdBooking.TimeSlotID != expectedTimeSlotID {
		t.Errorf("Expected timeslot ID %s, got %s", expectedTimeSlotID, createdBooking.TimeSlotID)
	}
}

// AssertBookingError asserts that a booking request resulted in an error with the expected status
func (b *BookingTestUtils) AssertBookingError(t *testing.T, rec *httptest.ResponseRecorder, expectedStatus int) {
	b.HTTP.AssertStatus(t, rec, expectedStatus)
}

// CreateIsolatedBookingData creates completely isolated test data for booking tests
func (b *BookingTestUtils) CreateIsolatedBookingData(t *testing.T, timezone string) *BookingTestData {
	// Create specialization with unique name
	specializationID := b.Database.CreateTestSpecialization(t, fmt.Sprintf("Isolated Test Specialization %d", rand.Intn(10000)))

	// Create therapist with unique email
	therapistID := b.Database.CreateTestTherapist(t, "Isolated Test Therapist", "isolated.therapist@example.com", "+1234567899", "+1234567899")

	// Link therapist to specialization
	b.Database.LinkTherapistToSpecialization(t, therapistID, specializationID)

	// Create client with unique WhatsApp number
	clientID := b.Database.CreateTestClient(t, "Isolated Test Client", "+1234567898", timezone)

	// Create time slot
	timeSlotID := b.Database.CreateTestTimeSlot(t, therapistID, "Sunday", "15:00", 60)

	return &BookingTestData{
		TherapistID:      therapistID,
		ClientID:         clientID,
		TimeSlotID:       timeSlotID,
		SpecializationID: specializationID,
	}
}
