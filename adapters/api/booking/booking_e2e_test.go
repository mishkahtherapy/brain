package booking_handler

import (
	"net/http"
	"testing"
	"time"

	"github.com/mishkahtherapy/brain/adapters/api/internal/testutils"
	"github.com/mishkahtherapy/brain/adapters/db"
	"github.com/mishkahtherapy/brain/adapters/db/booking_db"
	"github.com/mishkahtherapy/brain/adapters/db/therapist_db"
	"github.com/mishkahtherapy/brain/core/domain"
	"github.com/mishkahtherapy/brain/core/domain/booking"
	"github.com/mishkahtherapy/brain/core/domain/client"
	"github.com/mishkahtherapy/brain/core/domain/timeslot"
	"github.com/mishkahtherapy/brain/core/ports"
	"github.com/mishkahtherapy/brain/core/usecases/booking/cancel_booking"
	"github.com/mishkahtherapy/brain/core/usecases/booking/confirm_booking"
	"github.com/mishkahtherapy/brain/core/usecases/booking/create_booking"
	"github.com/mishkahtherapy/brain/core/usecases/booking/get_booking"
	"github.com/mishkahtherapy/brain/core/usecases/booking/list_bookings_by_client"
	"github.com/mishkahtherapy/brain/core/usecases/booking/list_bookings_by_therapist"
	"github.com/mishkahtherapy/brain/core/usecases/booking/search_bookings"

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

type TestClientRepository struct {
	db ports.SQLDatabase
}

func (r *TestClientRepository) BulkGetByID(ids []domain.ClientID) ([]*client.Client, error) {
	query := `SELECT id, name, whatsapp_number, timezone_offset, created_at, updated_at FROM clients WHERE id IN (?)`
	rows, err := r.db.Query(query, ids)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var clients []*client.Client
	for rows.Next() {
		var c client.Client
		err := rows.Scan(&c.ID, &c.Name, &c.WhatsAppNumber, &c.TimezoneOffset, &c.CreatedAt, &c.UpdatedAt)
		if err != nil {
			return nil, err
		}
		clients = append(clients, &c)
	}
	return clients, nil
}

func (r *TestClientRepository) Create(client *client.Client) error { return nil }
func (r *TestClientRepository) Update(client *client.Client) error { return nil }
func (r *TestClientRepository) Delete(id domain.ClientID) error    { return nil }
func (r *TestClientRepository) GetByWhatsAppNumber(whatsappNumber domain.WhatsAppNumber) (*client.Client, error) {
	return nil, nil
}
func (r *TestClientRepository) List() ([]*client.Client, error) { return nil, nil }
func (r *TestClientRepository) UpdateTimezoneOffset(id domain.ClientID, offsetMinutes domain.TimezoneOffset) error {
	return nil
}

type TestTimeSlotRepository struct {
	db ports.SQLDatabase
}

func (r *TestTimeSlotRepository) GetByID(id domain.TimeSlotID) (*timeslot.TimeSlot, error) {
	query := `SELECT id, therapist_id, day_of_week, start_time, duration_minutes, advance_notice, after_session_break_time, created_at, updated_at FROM time_slots WHERE id = ?`
	row := r.db.QueryRow(query, id)

	var timeSlot timeslot.TimeSlot
	err := row.Scan(&timeSlot.ID, &timeSlot.TherapistID, &timeSlot.DayOfWeek, &timeSlot.Start, &timeSlot.Duration, &timeSlot.AdvanceNotice, &timeSlot.AfterSessionBreakTime, &timeSlot.CreatedAt, &timeSlot.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &timeSlot, nil
}

func (r *TestTimeSlotRepository) Create(timeslot *timeslot.TimeSlot) error { return nil }
func (r *TestTimeSlotRepository) Update(timeslot *timeslot.TimeSlot) error { return nil }
func (r *TestTimeSlotRepository) Delete(id domain.TimeSlotID) error        { return nil }
func (r *TestTimeSlotRepository) ListByTherapist(therapistID domain.TherapistID) ([]*timeslot.TimeSlot, error) {
	return nil, nil
}
func (r *TestTimeSlotRepository) BulkToggleByTherapistID(therapistID domain.TherapistID, isActive bool) error {
	return nil
}
func (r *TestTimeSlotRepository) BulkListByTherapist(therapistIDs []domain.TherapistID) (map[domain.TherapistID][]*timeslot.TimeSlot, error) {
	return nil, nil
}

type TestNotificationPort struct {
	db ports.SQLDatabase
}

func (r *TestNotificationPort) SendNotification(deviceID domain.DeviceID, notification ports.Notification) (*ports.NotificationID, error) {
	return nil, nil
}

type TestNotificationRepository struct {
	db ports.SQLDatabase
}

func (r *TestNotificationRepository) CreateNotification(therapistID domain.TherapistID, firebaseNotificationID ports.NotificationID) error {
	return nil
}

func TestBookingE2E(t *testing.T) {
	// Setup test database
	database, cleanup := setupBookingTestDB(t)
	defer cleanup()

	// Setup repositories
	bookingRepo := booking_db.NewBookingRepository(database)
	therapistRepo := therapist_db.NewTherapistRepository(database)
	clientRepo := &TestClientRepository{db: database}
	timeSlotRepo := &TestTimeSlotRepository{db: database}
	sessionRepo := &TestSessionRepository{db: database}
	notificationPort := &TestNotificationPort{db: database}
	notificationRepo := &TestNotificationRepository{db: database}

	// Setup usecases
	createBookingUsecase := create_booking.NewUsecase(bookingRepo, therapistRepo, clientRepo, timeSlotRepo, getScheduleUsecase)
	getBookingUsecase := get_booking.NewUsecase(bookingRepo)
	confirmBookingUsecase := confirm_booking.NewUsecase(bookingRepo, sessionRepo, therapistRepo, notificationPort, notificationRepo)
	cancelBookingUsecase := cancel_booking.NewUsecase(bookingRepo)
	listByTherapistUsecase := list_bookings_by_therapist.NewUsecase(bookingRepo)
	listByClientUsecase := list_bookings_by_client.NewUsecase(bookingRepo)
	searchBookingsUsecase := search_bookings.NewUsecase(bookingRepo)

	// Setup handler
	bookingHandler := NewBookingHandler(
		*createBookingUsecase,
		*getBookingUsecase,
		*confirmBookingUsecase,
		*cancelBookingUsecase,
		*listByTherapistUsecase,
		*listByClientUsecase,
		*searchBookingsUsecase,
	)

	// Setup router
	mux := http.NewServeMux()
	bookingHandler.RegisterRoutes(mux)

	// Setup test utilities
	testUtils := testutils.NewBookingTestUtils(mux, database)

	t.Run("Complete booking workflow", func(t *testing.T) {
		// Create test data
		testData := testUtils.Database.CreateBookingTestData(t)

		// Step 1: Create a booking
		bookingRequest := testUtils.CreateBookingRequest(
			testData.TherapistID,
			testData.ClientID,
			testData.TimeSlotID,
			"2024-12-15T10:00:00Z",
			-300,
		)

		rec, createdBooking := testUtils.CreateBooking(t, bookingRequest)

		// Verify creation response
		testUtils.AssertBookingCreated(t, rec, testData.TherapistID, testData.ClientID, testData.TimeSlotID)

		// Verify timezone is stored correctly
		if createdBooking.ClientTimezoneOffset != -300 {
			t.Errorf("Expected timezone %d, got %d", -300, createdBooking.ClientTimezoneOffset)
		}

		// Step 2: Get the booking
		getRec := testUtils.HTTP.MakeRequest("GET", "/api/v1/bookings/"+string(createdBooking.ID), nil)
		testUtils.HTTP.AssertStatus(t, getRec, http.StatusOK)

		var retrievedBooking booking.Booking
		testUtils.HTTP.ParseResponse(t, getRec, &retrievedBooking)

		if retrievedBooking.ID != createdBooking.ID {
			t.Errorf("Expected booking ID %s, got %s", createdBooking.ID, retrievedBooking.ID)
		}

		// Step 3: Confirm the booking
		confirmData := map[string]interface{}{
			"bookingId":  createdBooking.ID,
			"paidAmount": 9999, // $99.99 USD
			"language":   "english",
		}

		confirmRec := testUtils.HTTP.MakeRequest("PUT", "/api/v1/bookings/"+string(createdBooking.ID)+"/confirm", confirmData)
		testUtils.HTTP.AssertStatus(t, confirmRec, http.StatusOK)

		// Step 4: List bookings by therapist
		listRec := testUtils.HTTP.MakeRequest("GET", "/api/v1/therapists/"+string(testData.TherapistID)+"/bookings", nil)
		testUtils.HTTP.AssertStatus(t, listRec, http.StatusOK)

		var bookings []booking.Booking
		testUtils.HTTP.ParseResponse(t, listRec, &bookings)

		if len(bookings) == 0 {
			t.Error("Expected at least one booking in the list")
		}

		// Step 5: Cancel the booking
		cancelRec := testUtils.HTTP.MakeRequest("PUT", "/api/v1/bookings/"+string(createdBooking.ID)+"/cancel", nil)
		testUtils.HTTP.AssertStatus(t, cancelRec, http.StatusOK)
	})

	t.Run("Timezone validation", func(t *testing.T) {
		// Create test data
		testData := testUtils.Database.CreateBookingTestData(t)

		// Test create booking without timezone (should fail)
		bookingWithoutTimezone := testUtils.CreateBookingRequest(
			testData.TherapistID,
			testData.ClientID,
			testData.TimeSlotID,
			"2024-12-15T12:00:00Z",
			-60, // Missing timezone
		)
		rec, _ := testUtils.CreateBooking(t, bookingWithoutTimezone)
		testUtils.AssertBookingError(t, rec, http.StatusBadRequest)

		// Test create booking with invalid timezone (should fail)
		bookingWithInvalidTimezone := testUtils.CreateBookingRequest(
			testData.TherapistID,
			testData.ClientID,
			testData.TimeSlotID,
			"2024-12-15T12:00:00Z",
			-60,
		)
		rec, _ = testUtils.CreateBooking(t, bookingWithInvalidTimezone)
		testUtils.AssertBookingError(t, rec, http.StatusBadRequest)

		// Test create booking with valid timezone (using isolated data to avoid conflicts)
		isolatedData := testUtils.CreateIsolatedBookingData(t, "Europe/London")
		bookingWithValidTimezone := testUtils.CreateBookingRequest(
			isolatedData.TherapistID,
			isolatedData.ClientID,
			isolatedData.TimeSlotID,
			"2024-12-21T15:00:00Z",
			-60,
		)
		rec, createdBooking := testUtils.CreateBooking(t, bookingWithValidTimezone)
		testUtils.AssertBookingCreated(t, rec, isolatedData.TherapistID, isolatedData.ClientID, isolatedData.TimeSlotID)

		// Verify the created booking has correct timezone
		if createdBooking.ClientTimezoneOffset != -60 {
			t.Errorf("Expected timezone %d, got %d", -60, createdBooking.ClientTimezoneOffset)
		}
	})

	t.Run("Error cases", func(t *testing.T) {
		// Test get non-existent booking
		nonExistentID := "booking_00000000-0000-0000-0000-000000000000"
		getRec := testUtils.HTTP.MakeRequest("GET", "/api/v1/bookings/"+nonExistentID, nil)
		testUtils.HTTP.AssertStatus(t, getRec, http.StatusNotFound)

		// Test create booking with invalid data (missing therapist ID)
		testData := testUtils.Database.CreateBookingTestData(t)
		invalidBookingData := map[string]interface{}{
			"clientId":   testData.ClientID,
			"timeSlotId": testData.TimeSlotID,
			"startTime":  "2024-12-15T11:00:00Z",
			"timezone":   "UTC",
		}
		rec := testUtils.HTTP.MakeRequest("POST", "/api/v1/bookings", invalidBookingData)
		testUtils.HTTP.AssertStatus(t, rec, http.StatusBadRequest)

		// Test confirm non-existent booking
		confirmData := map[string]interface{}{
			"bookingId":  nonExistentID,
			"paidAmount": 9999, // $99.99 USD
			"language":   "english",
		}
		confirmRec := testUtils.HTTP.MakeRequest("PUT", "/api/v1/bookings/"+nonExistentID+"/confirm", confirmData)
		testUtils.HTTP.AssertStatus(t, confirmRec, http.StatusNotFound)

		// Test invalid state parameter
		invalidStateRec := testUtils.HTTP.MakeRequest("GET", "/api/v1/therapists/"+string(testData.TherapistID)+"/bookings?state=invalid", nil)
		testUtils.HTTP.AssertStatus(t, invalidStateRec, http.StatusBadRequest)
	})
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
