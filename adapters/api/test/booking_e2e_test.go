package test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/mishkahtherapy/brain/adapters/api"
	"github.com/mishkahtherapy/brain/adapters/db"
	"github.com/mishkahtherapy/brain/adapters/db/booking"
	"github.com/mishkahtherapy/brain/adapters/db/therapist"
	"github.com/mishkahtherapy/brain/core/domain"
	"github.com/mishkahtherapy/brain/core/ports"
	"github.com/mishkahtherapy/brain/core/usecases/booking/cancel_booking"
	"github.com/mishkahtherapy/brain/core/usecases/booking/confirm_booking"
	"github.com/mishkahtherapy/brain/core/usecases/booking/create_booking"
	"github.com/mishkahtherapy/brain/core/usecases/booking/get_booking"
	"github.com/mishkahtherapy/brain/core/usecases/booking/list_bookings_by_client"
	"github.com/mishkahtherapy/brain/core/usecases/booking/list_bookings_by_therapist"

	_ "github.com/glebarez/go-sqlite"
)

// Simple test implementations for missing repositories

// TestSessionRepository is a test implementation of the session repository
type TestSessionRepository struct {
	db ports.SQLDatabase
}

func (r *TestSessionRepository) CreateSession(session *domain.Session) error {
	return nil // Just return success for test
}

func (r *TestSessionRepository) GetSessionByID(id domain.SessionID) (*domain.Session, error) {
	return nil, nil // Not used in test
}

func (r *TestSessionRepository) UpdateSessionState(id domain.SessionID, state domain.SessionState) error {
	return nil // Not used in test
}

func (r *TestSessionRepository) UpdateSessionNotes(id domain.SessionID, notes string) error {
	return nil // Not used in test
}

func (r *TestSessionRepository) UpdateMeetingURL(id domain.SessionID, meetingURL string) error {
	return nil // Not used in test
}

func (r *TestSessionRepository) ListSessionsByTherapist(therapistID domain.TherapistID) ([]*domain.Session, error) {
	return nil, nil // Not used in test
}

func (r *TestSessionRepository) ListSessionsByClient(clientID domain.ClientID) ([]*domain.Session, error) {
	return nil, nil // Not used in test
}

func (r *TestSessionRepository) ListSessionsAdmin(startDate, endDate time.Time) ([]*domain.Session, error) {
	return nil, nil // Not used in test
}

type TestClientRepository struct {
	db ports.SQLDatabase
}

func (r *TestClientRepository) GetByID(id domain.ClientID) (*domain.Client, error) {
	query := `SELECT id, name, whatsapp_number, created_at, updated_at FROM clients WHERE id = ?`
	row := r.db.QueryRow(query, id)

	var client domain.Client
	err := row.Scan(&client.ID, &client.Name, &client.WhatsAppNumber, &client.CreatedAt, &client.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &client, nil
}

func (r *TestClientRepository) Create(client *domain.Client) error {
	return nil // Not used in test
}

func (r *TestClientRepository) Update(client *domain.Client) error {
	return nil // Not used in test
}

func (r *TestClientRepository) Delete(id domain.ClientID) error {
	return nil // Not used in test
}

func (r *TestClientRepository) GetByWhatsAppNumber(whatsappNumber domain.WhatsAppNumber) (*domain.Client, error) {
	return nil, nil // Not used in test
}

func (r *TestClientRepository) List() ([]*domain.Client, error) {
	return nil, nil // Not used in test
}

type TestTimeSlotRepository struct {
	db ports.SQLDatabase
}

func (r *TestTimeSlotRepository) GetByID(id string) (*domain.TimeSlot, error) {
	query := `SELECT id, therapist_id, day_of_week, start_time, end_time, pre_session_buffer, post_session_buffer, created_at, updated_at FROM time_slots WHERE id = ?`
	row := r.db.QueryRow(query, id)

	var timeSlot domain.TimeSlot
	err := row.Scan(&timeSlot.ID, &timeSlot.TherapistID, &timeSlot.DayOfWeek, &timeSlot.StartTime, &timeSlot.EndTime, &timeSlot.PreSessionBuffer, &timeSlot.PostSessionBuffer, &timeSlot.CreatedAt, &timeSlot.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &timeSlot, nil
}

func (r *TestTimeSlotRepository) Create(timeslot *domain.TimeSlot) error {
	return nil // Not used in test
}

func (r *TestTimeSlotRepository) Update(timeslot *domain.TimeSlot) error {
	return nil // Not used in test
}

func (r *TestTimeSlotRepository) Delete(id string) error {
	return nil // Not used in test
}

func (r *TestTimeSlotRepository) ListByTherapist(therapistID string) ([]*domain.TimeSlot, error) {
	return nil, nil // Not used in test
}

func (r *TestTimeSlotRepository) ListByDay(therapistID string, day string) ([]*domain.TimeSlot, error) {
	return nil, nil // Not used in test
}

func TestBookingE2E(t *testing.T) {
	// Setup test database
	database, cleanup := setupBookingTestDB(t)
	defer cleanup()

	// Insert test data directly via SQL
	testData := insertBookingTestData(t, database)

	// Setup repositories
	bookingRepo := booking.NewBookingRepository(database)
	therapistRepo := therapist.NewTherapistRepository(database)
	clientRepo := &TestClientRepository{db: database}
	timeSlotRepo := &TestTimeSlotRepository{db: database}
	sessionRepo := &TestSessionRepository{db: database}

	// Setup usecases
	createBookingUsecase := create_booking.NewUsecase(bookingRepo, therapistRepo, clientRepo, timeSlotRepo)
	getBookingUsecase := get_booking.NewUsecase(bookingRepo)
	confirmBookingUsecase := confirm_booking.NewUsecase(bookingRepo, sessionRepo)
	cancelBookingUsecase := cancel_booking.NewUsecase(bookingRepo)
	listByTherapistUsecase := list_bookings_by_therapist.NewUsecase(bookingRepo)
	listByClientUsecase := list_bookings_by_client.NewUsecase(bookingRepo)

	// Setup handler
	bookingHandler := api.NewBookingHandler(
		*createBookingUsecase,
		*getBookingUsecase,
		*confirmBookingUsecase,
		*cancelBookingUsecase,
		*listByTherapistUsecase,
		*listByClientUsecase,
	)

	// Setup router
	mux := http.NewServeMux()
	bookingHandler.RegisterRoutes(mux)

	t.Run("Complete booking workflow", func(t *testing.T) {
		// Step 1: Create a booking
		bookingData := map[string]interface{}{
			"therapistId": testData.TherapistID,
			"clientId":    testData.ClientID,
			"timeSlotId":  testData.TimeSlotID,
			"startTime":   "2024-12-15T10:00:00Z",
		}
		bookingBody, _ := json.Marshal(bookingData)

		createReq := httptest.NewRequest("POST", "/api/v1/bookings", bytes.NewBuffer(bookingBody))
		createReq.Header.Set("Content-Type", "application/json")
		createRec := httptest.NewRecorder()

		mux.ServeHTTP(createRec, createReq)

		// Verify creation response
		if createRec.Code != http.StatusCreated {
			t.Fatalf("Expected status %d, got %d. Body: %s", http.StatusCreated, createRec.Code, createRec.Body.String())
		}

		// Parse created booking
		var createdBooking domain.Booking
		if err := json.Unmarshal(createRec.Body.Bytes(), &createdBooking); err != nil {
			t.Fatalf("Failed to parse created booking: %v", err)
		}

		// Verify created booking data
		if createdBooking.TherapistID != testData.TherapistID {
			t.Errorf("Expected therapist ID %s, got %s", testData.TherapistID, createdBooking.TherapistID)
		}
		if createdBooking.ClientID != testData.ClientID {
			t.Errorf("Expected client ID %s, got %s", testData.ClientID, createdBooking.ClientID)
		}
		if createdBooking.TimeSlotID != testData.TimeSlotID {
			t.Errorf("Expected timeslot ID %s, got %s", testData.TimeSlotID, createdBooking.TimeSlotID)
		}
		if createdBooking.State != domain.BookingStatePending {
			t.Errorf("Expected state %s, got %s", domain.BookingStatePending, createdBooking.State)
		}
		if createdBooking.ID == "" {
			t.Error("Expected ID to be set")
		}

		// Step 2: Get the booking by ID
		getReq := httptest.NewRequest("GET", "/api/v1/bookings/"+string(createdBooking.ID), nil)
		getRec := httptest.NewRecorder()

		mux.ServeHTTP(getRec, getReq)

		// Verify get response
		if getRec.Code != http.StatusOK {
			t.Fatalf("Expected status %d, got %d. Body: %s", http.StatusOK, getRec.Code, getRec.Body.String())
		}

		// Parse retrieved booking
		var retrievedBooking domain.Booking
		if err := json.Unmarshal(getRec.Body.Bytes(), &retrievedBooking); err != nil {
			t.Fatalf("Failed to parse retrieved booking: %v", err)
		}

		// Verify retrieved booking matches created one
		if retrievedBooking.ID != createdBooking.ID {
			t.Errorf("Expected ID %s, got %s", createdBooking.ID, retrievedBooking.ID)
		}
		if retrievedBooking.State != domain.BookingStatePending {
			t.Errorf("Expected state %s, got %s", domain.BookingStatePending, retrievedBooking.State)
		}

		// Step 3: Confirm the booking
		confirmData := map[string]interface{}{
			"bookingId":  createdBooking.ID,
			"paidAmount": 9999, // $99.99 USD
			"language":   "english",
		}
		confirmBody, _ := json.Marshal(confirmData)
		confirmReq := httptest.NewRequest("PUT", "/api/v1/bookings/"+string(createdBooking.ID)+"/confirm", bytes.NewBuffer(confirmBody))
		confirmReq.Header.Set("Content-Type", "application/json")
		confirmRec := httptest.NewRecorder()

		mux.ServeHTTP(confirmRec, confirmReq)

		// Verify confirm response
		if confirmRec.Code != http.StatusOK {
			t.Fatalf("Expected status %d, got %d. Body: %s", http.StatusOK, confirmRec.Code, confirmRec.Body.String())
		}

		// Parse confirmed booking
		var confirmedBooking domain.Booking
		if err := json.Unmarshal(confirmRec.Body.Bytes(), &confirmedBooking); err != nil {
			t.Fatalf("Failed to parse confirmed booking: %v", err)
		}

		// Verify booking state changed to confirmed
		if confirmedBooking.State != domain.BookingStateConfirmed {
			t.Errorf("Expected state %s, got %s", domain.BookingStateConfirmed, confirmedBooking.State)
		}

		// Step 4: List bookings by therapist
		listByTherapistReq := httptest.NewRequest("GET", "/api/v1/therapists/"+string(testData.TherapistID)+"/bookings", nil)
		listByTherapistRec := httptest.NewRecorder()

		mux.ServeHTTP(listByTherapistRec, listByTherapistReq)

		// Verify list response
		if listByTherapistRec.Code != http.StatusOK {
			t.Fatalf("Expected status %d, got %d. Body: %s", http.StatusOK, listByTherapistRec.Code, listByTherapistRec.Body.String())
		}

		// Parse therapist bookings
		var therapistBookings []*domain.Booking
		if err := json.Unmarshal(listByTherapistRec.Body.Bytes(), &therapistBookings); err != nil {
			t.Fatalf("Failed to parse therapist bookings: %v", err)
		}

		// Verify our booking is in the list
		found := false
		for _, booking := range therapistBookings {
			if booking.ID == createdBooking.ID {
				found = true
				if booking.State != domain.BookingStateConfirmed {
					t.Errorf("Expected booking state %s, got %s", domain.BookingStateConfirmed, booking.State)
				}
				break
			}
		}
		if !found {
			t.Error("Created booking not found in therapist bookings list")
		}

		// Step 5: List bookings by client
		listByClientReq := httptest.NewRequest("GET", "/api/v1/clients/"+string(testData.ClientID)+"/bookings", nil)
		listByClientRec := httptest.NewRecorder()

		mux.ServeHTTP(listByClientRec, listByClientReq)

		// Verify list response
		if listByClientRec.Code != http.StatusOK {
			t.Fatalf("Expected status %d, got %d. Body: %s", http.StatusOK, listByClientRec.Code, listByClientRec.Body.String())
		}

		// Parse client bookings
		var clientBookings []*domain.Booking
		if err := json.Unmarshal(listByClientRec.Body.Bytes(), &clientBookings); err != nil {
			t.Fatalf("Failed to parse client bookings: %v", err)
		}

		// Verify our booking is in the list
		found = false
		for _, booking := range clientBookings {
			if booking.ID == createdBooking.ID {
				found = true
				break
			}
		}
		if !found {
			t.Error("Created booking not found in client bookings list")
		}

		// Step 6: Test state filtering - list confirmed bookings for therapist
		listConfirmedReq := httptest.NewRequest("GET", "/api/v1/therapists/"+string(testData.TherapistID)+"/bookings?state=confirmed", nil)
		listConfirmedRec := httptest.NewRecorder()

		mux.ServeHTTP(listConfirmedRec, listConfirmedReq)

		if listConfirmedRec.Code != http.StatusOK {
			t.Fatalf("Expected status %d, got %d. Body: %s", http.StatusOK, listConfirmedRec.Code, listConfirmedRec.Body.String())
		}

		var confirmedBookings []*domain.Booking
		if err := json.Unmarshal(listConfirmedRec.Body.Bytes(), &confirmedBookings); err != nil {
			t.Fatalf("Failed to parse confirmed bookings: %v", err)
		}

		// Verify only confirmed bookings are returned
		for _, booking := range confirmedBookings {
			if booking.State != domain.BookingStateConfirmed {
				t.Errorf("Expected only confirmed bookings, got booking with state %s", booking.State)
			}
		}

		// Step 7: Cancel the booking
		cancelReq := httptest.NewRequest("PUT", "/api/v1/bookings/"+string(createdBooking.ID)+"/cancel", nil)
		cancelRec := httptest.NewRecorder()

		mux.ServeHTTP(cancelRec, cancelReq)

		// Verify cancel response
		if cancelRec.Code != http.StatusOK {
			t.Fatalf("Expected status %d, got %d. Body: %s", http.StatusOK, cancelRec.Code, cancelRec.Body.String())
		}

		// Parse cancelled booking
		var cancelledBooking domain.Booking
		if err := json.Unmarshal(cancelRec.Body.Bytes(), &cancelledBooking); err != nil {
			t.Fatalf("Failed to parse cancelled booking: %v", err)
		}

		// Verify booking state changed to cancelled
		if cancelledBooking.State != domain.BookingStateCancelled {
			t.Errorf("Expected state %s, got %s", domain.BookingStateCancelled, cancelledBooking.State)
		}
	})

	t.Run("Error cases", func(t *testing.T) {
		// Test get non-existent booking
		nonExistentID := "booking_00000000-0000-0000-0000-000000000000"
		getReq := httptest.NewRequest("GET", "/api/v1/bookings/"+nonExistentID, nil)
		getRec := httptest.NewRecorder()

		mux.ServeHTTP(getRec, getReq)

		if getRec.Code != http.StatusNotFound {
			t.Errorf("Expected status %d for non-existent booking, got %d", http.StatusNotFound, getRec.Code)
		}

		// Test create booking with invalid data (missing therapist ID)
		invalidBookingData := map[string]interface{}{
			"clientId":   testData.ClientID,
			"timeSlotId": testData.TimeSlotID,
			"startTime":  "2024-12-15T11:00:00Z",
		}
		invalidBody, _ := json.Marshal(invalidBookingData)

		createReq := httptest.NewRequest("POST", "/api/v1/bookings", bytes.NewBuffer(invalidBody))
		createReq.Header.Set("Content-Type", "application/json")
		createRec := httptest.NewRecorder()

		mux.ServeHTTP(createRec, createReq)

		if createRec.Code != http.StatusBadRequest {
			t.Errorf("Expected status %d for invalid booking data, got %d", http.StatusBadRequest, createRec.Code)
		}

		// Test confirm non-existent booking
		confirmData := map[string]interface{}{
			"bookingId":  nonExistentID,
			"paidAmount": 9999, // $99.99 USD
			"language":   "english",
		}
		confirmBody, _ := json.Marshal(confirmData)
		confirmReq := httptest.NewRequest("PUT", "/api/v1/bookings/"+nonExistentID+"/confirm", bytes.NewBuffer(confirmBody))
		confirmReq.Header.Set("Content-Type", "application/json")
		confirmRec := httptest.NewRecorder()

		mux.ServeHTTP(confirmRec, confirmReq)

		if confirmRec.Code != http.StatusNotFound {
			t.Errorf("Expected status %d for confirming non-existent booking, got %d", http.StatusNotFound, confirmRec.Code)
		}

		// Test invalid state parameter
		invalidStateReq := httptest.NewRequest("GET", "/api/v1/therapists/"+string(testData.TherapistID)+"/bookings?state=invalid", nil)
		invalidStateRec := httptest.NewRecorder()

		mux.ServeHTTP(invalidStateRec, invalidStateReq)

		if invalidStateRec.Code != http.StatusBadRequest {
			t.Errorf("Expected status %d for invalid state parameter, got %d", http.StatusBadRequest, invalidStateRec.Code)
		}
	})
}

type BookingTestData struct {
	TherapistID      domain.TherapistID
	ClientID         domain.ClientID
	TimeSlotID       domain.TimeSlotID
	SpecializationID domain.SpecializationID
}

func insertBookingTestData(t *testing.T, database ports.SQLDatabase) *BookingTestData {
	now := time.Now().UTC()

	// Generate test IDs
	specializationID := domain.NewSpecializationID()
	therapistID := domain.NewTherapistID()
	clientID := domain.NewClientID()
	timeSlotID := domain.NewTimeSlotID()

	// Insert specialization
	_, err := database.Exec(`
		INSERT INTO specializations (id, name, created_at, updated_at)
		VALUES (?, ?, ?, ?)
	`, specializationID, "Test Specialization", now, now)
	if err != nil {
		t.Fatalf("Failed to insert test specialization: %v", err)
	}

	// Insert therapist
	_, err = database.Exec(`
		INSERT INTO therapists (id, name, email, phone_number, whatsapp_number, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`, therapistID, "Dr. Test Therapist", "test.therapist@example.com", "+1234567890", "+1234567890", now, now)
	if err != nil {
		t.Fatalf("Failed to insert test therapist: %v", err)
	}

	// Insert therapist specialization
	_, err = database.Exec(`
		INSERT INTO therapist_specializations (id, therapist_id, specialization_id, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?)
	`, "therapist_spec_test_1", therapistID, specializationID, now, now)
	if err != nil {
		t.Fatalf("Failed to insert therapist specialization: %v", err)
	}

	// Insert client
	_, err = database.Exec(`
		INSERT INTO clients (id, name, whatsapp_number, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?)
	`, clientID, "Test Client", "+1234567891", now, now)
	if err != nil {
		t.Fatalf("Failed to insert test client: %v", err)
	}

	// Insert time slot
	_, err = database.Exec(`
		INSERT INTO time_slots (id, therapist_id, day_of_week, start_time, end_time, pre_session_buffer, post_session_buffer, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, timeSlotID, therapistID, "Monday", "10:00:00", "11:00:00", 0, 0, now, now)
	if err != nil {
		t.Fatalf("Failed to insert test time slot: %v", err)
	}

	return &BookingTestData{
		TherapistID:      therapistID,
		ClientID:         clientID,
		TimeSlotID:       timeSlotID,
		SpecializationID: specializationID,
	}
}

func setupBookingTestDB(_ *testing.T) (ports.SQLDatabase, func()) {
	// Use in-memory database for testing
	dbFilename := ":memory:"

	database := db.NewDatabase(db.DatabaseConfig{
		DBFilename: dbFilename,
		SchemaFile: "../../../schema.sql",
	})

	// Return cleanup function
	cleanup := func() {
		database.Close()
	}

	return database, cleanup
}
