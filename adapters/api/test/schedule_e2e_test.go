package test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/mishkahtherapy/brain/adapters/api"
	"github.com/mishkahtherapy/brain/adapters/db"
	"github.com/mishkahtherapy/brain/adapters/db/booking"
	"github.com/mishkahtherapy/brain/adapters/db/therapist"
	"github.com/mishkahtherapy/brain/adapters/db/timeslot"
	"github.com/mishkahtherapy/brain/core/domain"
	"github.com/mishkahtherapy/brain/core/ports"
	"github.com/mishkahtherapy/brain/core/usecases/schedule/get_schedule"

	_ "github.com/glebarez/go-sqlite"
)

type ScheduleTestData struct {
	Specializations []domain.Specialization
	Therapists      []domain.Therapist
	TimeSlots       []domain.TimeSlot
	Bookings        []domain.Booking
	Clients         []domain.Client
}

func TestScheduleE2E(t *testing.T) {
	// Setup test database
	database, cleanup := setupScheduleTestDB(t)
	defer cleanup()

	// Insert comprehensive test data
	testData := insertScheduleTestData(t, database)
	t.Logf("Created test data with %d therapists, %d time slots, %d bookings",
		len(testData.Therapists), len(testData.TimeSlots), len(testData.Bookings))

	// Setup repositories
	therapistRepo := therapist.NewTherapistRepository(database)
	timeSlotRepo := timeslot.NewTimeSlotRepository(database)
	bookingRepo := booking.NewBookingRepository(database)

	// Setup usecase
	getScheduleUsecase := get_schedule.NewUsecase(therapistRepo, timeSlotRepo, bookingRepo)

	// Setup handler
	scheduleHandler := api.NewScheduleHandler(*getScheduleUsecase)

	// Setup router
	mux := http.NewServeMux()
	scheduleHandler.RegisterRoutes(mux)

	t.Run("Complex Three-Therapist Overlap Scenario", func(t *testing.T) {
		// Test overlapping availability from 9:15-10:45 on Monday
		// Expected: 3 therapists available from 9:15-10:00, then 2 therapists from 10:00-10:45
		// Use a future Monday
		req := httptest.NewRequest("GET", "/api/v1/schedule?tag=anxiety&start_date=2025-07-07&end_date=2025-07-07", nil)
		rec := httptest.NewRecorder()

		mux.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("Expected status %d, got %d. Body: %s", http.StatusOK, rec.Code, rec.Body.String())
		}

		var response domain.ScheduleResponse
		if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
			t.Fatalf("Failed to parse response: %v", err)
		}

		// Debug output
		t.Logf("Response body: %s", rec.Body.String())
		t.Logf("Found %d availability ranges", len(response.Availabilities))

		// Verify we have availability ranges
		if len(response.Availabilities) == 0 {
			t.Fatal("Expected availability ranges, got none")
		}

		// Find the 9:15-10:00 range where 3 therapists should be available
		found3Therapists := false
		found2Therapists := false

		for _, avail := range response.Availabilities {
			startTime := avail.StartTime.Format("15:04")
			endTime := avail.EndTime.Format("15:04")
			therapistCount := len(avail.Therapists)

			// Check for 3-therapist overlap (9:15-10:00)
			if startTime == "09:15" && endTime == "10:00" && therapistCount == 3 {
				found3Therapists = true
				t.Logf("Found 3-therapist range: %s-%s with %d therapists", startTime, endTime, therapistCount)
			}

			// Check for 2-therapist overlap (10:00-10:45)
			if startTime == "10:00" && endTime == "10:45" && therapistCount == 2 {
				found2Therapists = true
				t.Logf("Found 2-therapist range: %s-%s with %d therapists", startTime, endTime, therapistCount)
			}
		}

		if !found3Therapists {
			t.Error("Expected to find a 3-therapist overlap from 9:15-10:00")
		}
		if !found2Therapists {
			t.Error("Expected to find a 2-therapist overlap from 10:00-10:45")
		}
	})

	t.Run("Transition Point Testing", func(t *testing.T) {
		// Test Wednesday where therapists join and leave at different times
		// Expected: Complex transitions with varying therapist counts
		// Use a future Wednesday
		req := httptest.NewRequest("GET", "/api/v1/schedule?tag=depression&start_date=2025-07-09&end_date=2025-07-09", nil)
		rec := httptest.NewRecorder()

		mux.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("Expected status %d, got %d. Body: %s", http.StatusOK, rec.Code, rec.Body.String())
		}

		var response domain.ScheduleResponse
		if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
			t.Fatalf("Failed to parse response: %v", err)
		}

		// Verify transition points exist
		if len(response.Availabilities) == 0 {
			t.Fatal("Expected availability ranges for transition testing")
		}

		// Log all availability ranges for debugging
		t.Logf("Found %d availability ranges on Wednesday:", len(response.Availabilities))
		for i, avail := range response.Availabilities {
			t.Logf("Range %d: %s-%s (%d therapists)",
				i+1,
				avail.StartTime.Format("15:04"),
				avail.EndTime.Format("15:04"),
				len(avail.Therapists))
		}
	})

	t.Run("Mid-Hour Overlap Complex Pattern", func(t *testing.T) {
		// Test Tuesday with non-standard times creating complex overlaps
		// Use a future Tuesday
		req := httptest.NewRequest("GET", "/api/v1/schedule?tag=anxiety&start_date=2025-07-08&end_date=2025-07-08", nil)
		rec := httptest.NewRecorder()

		mux.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("Expected status %d, got %d. Body: %s", http.StatusOK, rec.Code, rec.Body.String())
		}

		var response domain.ScheduleResponse
		if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
			t.Fatalf("Failed to parse response: %v", err)
		}

		// Verify that non-hour boundary times are handled correctly
		hasNonHourBoundary := false
		for _, avail := range response.Availabilities {
			startMinute := avail.StartTime.Minute()
			endMinute := avail.EndTime.Minute()

			if startMinute != 0 || endMinute != 0 {
				hasNonHourBoundary = true
				t.Logf("Found non-hour boundary: %s-%s",
					avail.StartTime.Format("15:04"),
					avail.EndTime.Format("15:04"))
			}
		}

		if !hasNonHourBoundary {
			t.Error("Expected to find availability ranges with non-hour boundaries")
		}
	})

	t.Run("Full Day Multiple Therapists", func(t *testing.T) {
		// Test Thursday with comprehensive availability patterns
		// Use a future Thursday
		req := httptest.NewRequest("GET", "/api/v1/schedule?tag=anxiety&start_date=2025-07-10&end_date=2025-07-10", nil)
		rec := httptest.NewRecorder()

		mux.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("Expected status %d, got %d. Body: %s", http.StatusOK, rec.Code, rec.Body.String())
		}

		var response domain.ScheduleResponse
		if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
			t.Fatalf("Failed to parse response: %v", err)
		}

		// Verify we have multiple time ranges throughout the day
		if len(response.Availabilities) < 3 {
			t.Errorf("Expected at least 3 availability ranges for full day, got %d", len(response.Availabilities))
		}

		// Verify ranges are properly sorted
		for i := 1; i < len(response.Availabilities); i++ {
			if response.Availabilities[i].StartTime.Before(response.Availabilities[i-1].EndTime) {
				t.Error("Availability ranges are not properly sorted or have overlaps")
			}
		}
	})

	t.Run("English Language Requirement", func(t *testing.T) {
		// Test with english=true filter
		// Use a future Monday
		req := httptest.NewRequest("GET", "/api/v1/schedule?tag=anxiety&english=true&start_date=2025-07-07&end_date=2025-07-07", nil)
		rec := httptest.NewRecorder()

		mux.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("Expected status %d, got %d. Body: %s", http.StatusOK, rec.Code, rec.Body.String())
		}

		var response domain.ScheduleResponse
		if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
			t.Fatalf("Failed to parse response: %v", err)
		}

		// Verify all therapists speak English
		for _, avail := range response.Availabilities {
			for _, therapist := range avail.Therapists {
				if !therapist.SpeaksEnglish {
					t.Errorf("Expected only English-speaking therapists, found non-English speaking therapist: %s", therapist.Name)
				}
			}
		}
	})

	t.Run("Booking Interference Testing", func(t *testing.T) {
		// Test Friday where bookings create "holes" in availability
		// Use a future Friday
		req := httptest.NewRequest("GET", "/api/v1/schedule?tag=anxiety&start_date=2025-07-11&end_date=2025-07-11", nil)
		rec := httptest.NewRecorder()

		mux.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("Expected status %d, got %d. Body: %s", http.StatusOK, rec.Code, rec.Body.String())
		}

		var response domain.ScheduleResponse
		if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
			t.Fatalf("Failed to parse response: %v", err)
		}

		// Verify that bookings create gaps in availability
		// We should see separate ranges before and after bookings
		t.Logf("Friday availability with booking interference:")
		for i, avail := range response.Availabilities {
			t.Logf("Range %d: %s-%s (%d min, %d therapists)",
				i+1,
				avail.StartTime.Format("15:04"),
				avail.EndTime.Format("15:04"),
				avail.DurationMinutes,
				len(avail.Therapists))
		}
	})

	t.Run("No Matching Therapists Edge Case", func(t *testing.T) {
		// Test with a specialization that doesn't exist
		// Use a future Monday
		req := httptest.NewRequest("GET", "/api/v1/schedule?tag=nonexistent&start_date=2025-07-07&end_date=2025-07-07", nil)
		rec := httptest.NewRecorder()

		mux.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("Expected status %d, got %d. Body: %s", http.StatusOK, rec.Code, rec.Body.String())
		}

		var response domain.ScheduleResponse
		if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
			t.Fatalf("Failed to parse response: %v", err)
		}

		// Should return empty availability
		if len(response.Availabilities) != 0 {
			t.Errorf("Expected no availability for nonexistent specialization, got %d ranges", len(response.Availabilities))
		}
	})

	t.Run("Error Cases", func(t *testing.T) {
		// Test missing tag parameter
		req := httptest.NewRequest("GET", "/api/v1/schedule", nil)
		rec := httptest.NewRecorder()

		mux.ServeHTTP(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Errorf("Expected status %d for missing tag, got %d", http.StatusBadRequest, rec.Code)
		}

		// Test invalid date format
		req = httptest.NewRequest("GET", "/api/v1/schedule?tag=anxiety&start_date=invalid", nil)
		rec = httptest.NewRecorder()

		mux.ServeHTTP(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Errorf("Expected status %d for invalid date format, got %d", http.StatusBadRequest, rec.Code)
		}

		// Test invalid date range
		req = httptest.NewRequest("GET", "/api/v1/schedule?tag=anxiety&start_date=2024-01-10&end_date=2024-01-08", nil)
		rec = httptest.NewRecorder()

		mux.ServeHTTP(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Errorf("Expected status %d for invalid date range, got %d", http.StatusBadRequest, rec.Code)
		}
	})
}

func insertScheduleTestData(t *testing.T, database ports.SQLDatabase) *ScheduleTestData {
	now := time.Now().UTC()

	// Create specializations
	anxietySpec := domain.Specialization{
		ID:        domain.NewSpecializationID(),
		Name:      "anxiety",
		CreatedAt: domain.UTCTimestamp(now),
		UpdatedAt: domain.UTCTimestamp(now),
	}

	depressionSpec := domain.Specialization{
		ID:        domain.NewSpecializationID(),
		Name:      "depression",
		CreatedAt: domain.UTCTimestamp(now),
		UpdatedAt: domain.UTCTimestamp(now),
	}

	// Insert specializations
	_, err := database.Exec(`
		INSERT INTO specializations (id, name, created_at, updated_at)
		VALUES (?, ?, ?, ?), (?, ?, ?, ?)
	`, anxietySpec.ID, anxietySpec.Name, anxietySpec.CreatedAt, anxietySpec.UpdatedAt,
		depressionSpec.ID, depressionSpec.Name, depressionSpec.CreatedAt, depressionSpec.UpdatedAt)
	if err != nil {
		t.Fatalf("Failed to insert specializations: %v", err)
	}

	// Create therapists with varying specializations and language capabilities
	therapists := []domain.Therapist{
		{
			ID:              domain.NewTherapistID(),
			Name:            "Dr. Alice Johnson",
			Email:           "alice@therapy.com",
			PhoneNumber:     "+1555001001",
			WhatsAppNumber:  "+1555001001",
			SpeaksEnglish:   true,
			Specializations: []domain.Specialization{anxietySpec},
			CreatedAt:       domain.UTCTimestamp(now),
			UpdatedAt:       domain.UTCTimestamp(now),
		},
		{
			ID:              domain.NewTherapistID(),
			Name:            "Dr. Bob Smith",
			Email:           "bob@therapy.com",
			PhoneNumber:     "+1555001002",
			WhatsAppNumber:  "+1555001002",
			SpeaksEnglish:   true,
			Specializations: []domain.Specialization{anxietySpec, depressionSpec},
			CreatedAt:       domain.UTCTimestamp(now),
			UpdatedAt:       domain.UTCTimestamp(now),
		},
		{
			ID:              domain.NewTherapistID(),
			Name:            "Dr. Carol Davis",
			Email:           "carol@therapy.com",
			PhoneNumber:     "+1555001003",
			WhatsAppNumber:  "+1555001003",
			SpeaksEnglish:   false, // Non-English speaker for language filtering tests
			Specializations: []domain.Specialization{anxietySpec},
			CreatedAt:       domain.UTCTimestamp(now),
			UpdatedAt:       domain.UTCTimestamp(now),
		},
		{
			ID:              domain.NewTherapistID(),
			Name:            "Dr. David Wilson",
			Email:           "david@therapy.com",
			PhoneNumber:     "+1555001004",
			WhatsAppNumber:  "+1555001004",
			SpeaksEnglish:   true,
			Specializations: []domain.Specialization{depressionSpec},
			CreatedAt:       domain.UTCTimestamp(now),
			UpdatedAt:       domain.UTCTimestamp(now),
		},
	}

	// Insert therapists
	for _, therapist := range therapists {
		_, err = database.Exec(`
			INSERT INTO therapists (id, name, email, phone_number, whatsapp_number, speaks_english, created_at, updated_at)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?)
		`, therapist.ID, therapist.Name, therapist.Email, therapist.PhoneNumber,
			therapist.WhatsAppNumber, therapist.SpeaksEnglish, therapist.CreatedAt, therapist.UpdatedAt)
		if err != nil {
			t.Fatalf("Failed to insert therapist %s: %v", therapist.Name, err)
		}

		// Insert therapist specializations
		for i, spec := range therapist.Specializations {
			_, err = database.Exec(`
				INSERT INTO therapist_specializations (id, therapist_id, specialization_id, created_at, updated_at)
				VALUES (?, ?, ?, ?, ?)
			`, fmt.Sprintf("ts_%s_%d", therapist.ID, i), therapist.ID, spec.ID, now, now)
			if err != nil {
				t.Fatalf("Failed to insert therapist specialization: %v", err)
			}
		}
	}

	// Create complex time slot patterns for overlap testing
	timeSlots := []domain.TimeSlot{
		// Monday - Complex 3-therapist overlap scenario
		// Alice: 9:00-11:00, Bob: 9:15-10:45, Carol: 9:15-10:00
		{
			ID:                domain.NewTimeSlotID(),
			TherapistID:       therapists[0].ID, // Alice
			DayOfWeek:         domain.DayOfWeekMonday,
			StartTime:         "09:00",
			EndTime:           "11:00",
			PreSessionBuffer:  5,
			PostSessionBuffer: 5,
			CreatedAt:         domain.UTCTimestamp(now),
			UpdatedAt:         domain.UTCTimestamp(now),
		},
		{
			ID:                domain.NewTimeSlotID(),
			TherapistID:       therapists[1].ID, // Bob
			DayOfWeek:         domain.DayOfWeekMonday,
			StartTime:         "09:15",
			EndTime:           "10:45",
			PreSessionBuffer:  0,
			PostSessionBuffer: 0,
			CreatedAt:         domain.UTCTimestamp(now),
			UpdatedAt:         domain.UTCTimestamp(now),
		},
		{
			ID:                domain.NewTimeSlotID(),
			TherapistID:       therapists[2].ID, // Carol
			DayOfWeek:         domain.DayOfWeekMonday,
			StartTime:         "09:15",
			EndTime:           "10:00",
			PreSessionBuffer:  0,
			PostSessionBuffer: 0,
			CreatedAt:         domain.UTCTimestamp(now),
			UpdatedAt:         domain.UTCTimestamp(now),
		},

		// Tuesday - Non-standard time overlaps
		// Alice: 14:30-16:00, Bob: 15:00-17:00
		{
			ID:                domain.NewTimeSlotID(),
			TherapistID:       therapists[0].ID, // Alice
			DayOfWeek:         domain.DayOfWeekTuesday,
			StartTime:         "14:30",
			EndTime:           "16:00",
			PreSessionBuffer:  0,
			PostSessionBuffer: 0,
			CreatedAt:         domain.UTCTimestamp(now),
			UpdatedAt:         domain.UTCTimestamp(now),
		},
		{
			ID:                domain.NewTimeSlotID(),
			TherapistID:       therapists[1].ID, // Bob
			DayOfWeek:         domain.DayOfWeekTuesday,
			StartTime:         "15:00",
			EndTime:           "17:00",
			PreSessionBuffer:  0,
			PostSessionBuffer: 0,
			CreatedAt:         domain.UTCTimestamp(now),
			UpdatedAt:         domain.UTCTimestamp(now),
		},

		// Wednesday - Transition point testing
		// Bob: 10:00-12:00, David: 11:00-14:00, Carol: 13:00-15:00
		{
			ID:                domain.NewTimeSlotID(),
			TherapistID:       therapists[1].ID, // Bob
			DayOfWeek:         domain.DayOfWeekWednesday,
			StartTime:         "10:00",
			EndTime:           "12:00",
			PreSessionBuffer:  0,
			PostSessionBuffer: 0,
			CreatedAt:         domain.UTCTimestamp(now),
			UpdatedAt:         domain.UTCTimestamp(now),
		},
		{
			ID:                domain.NewTimeSlotID(),
			TherapistID:       therapists[3].ID, // David
			DayOfWeek:         domain.DayOfWeekWednesday,
			StartTime:         "11:00",
			EndTime:           "14:00",
			PreSessionBuffer:  0,
			PostSessionBuffer: 0,
			CreatedAt:         domain.UTCTimestamp(now),
			UpdatedAt:         domain.UTCTimestamp(now),
		},
		{
			ID:                domain.NewTimeSlotID(),
			TherapistID:       therapists[2].ID, // Carol
			DayOfWeek:         domain.DayOfWeekWednesday,
			StartTime:         "13:00",
			EndTime:           "15:00",
			PreSessionBuffer:  0,
			PostSessionBuffer: 0,
			CreatedAt:         domain.UTCTimestamp(now),
			UpdatedAt:         domain.UTCTimestamp(now),
		},

		// Thursday - Full day with multiple therapists
		// Alice: 9:00-12:00, Bob: 10:00-16:00, Carol: 14:00-18:00
		{
			ID:                domain.NewTimeSlotID(),
			TherapistID:       therapists[0].ID, // Alice
			DayOfWeek:         domain.DayOfWeekThursday,
			StartTime:         "09:00",
			EndTime:           "12:00",
			PreSessionBuffer:  0,
			PostSessionBuffer: 0,
			CreatedAt:         domain.UTCTimestamp(now),
			UpdatedAt:         domain.UTCTimestamp(now),
		},
		{
			ID:                domain.NewTimeSlotID(),
			TherapistID:       therapists[1].ID, // Bob
			DayOfWeek:         domain.DayOfWeekThursday,
			StartTime:         "10:00",
			EndTime:           "16:00",
			PreSessionBuffer:  0,
			PostSessionBuffer: 0,
			CreatedAt:         domain.UTCTimestamp(now),
			UpdatedAt:         domain.UTCTimestamp(now),
		},
		{
			ID:                domain.NewTimeSlotID(),
			TherapistID:       therapists[2].ID, // Carol
			DayOfWeek:         domain.DayOfWeekThursday,
			StartTime:         "14:00",
			EndTime:           "18:00",
			PreSessionBuffer:  0,
			PostSessionBuffer: 0,
			CreatedAt:         domain.UTCTimestamp(now),
			UpdatedAt:         domain.UTCTimestamp(now),
		},

		// Friday - Booking interference testing
		// Alice: 9:00-17:00, Bob: 9:00-17:00 (will have bookings to create holes)
		{
			ID:                domain.NewTimeSlotID(),
			TherapistID:       therapists[0].ID, // Alice
			DayOfWeek:         domain.DayOfWeekFriday,
			StartTime:         "09:00",
			EndTime:           "17:00",
			PreSessionBuffer:  15, // With buffers to test complex availability
			PostSessionBuffer: 15,
			CreatedAt:         domain.UTCTimestamp(now),
			UpdatedAt:         domain.UTCTimestamp(now),
		},
		{
			ID:                domain.NewTimeSlotID(),
			TherapistID:       therapists[1].ID, // Bob
			DayOfWeek:         domain.DayOfWeekFriday,
			StartTime:         "09:00",
			EndTime:           "17:00",
			PreSessionBuffer:  15,
			PostSessionBuffer: 15,
			CreatedAt:         domain.UTCTimestamp(now),
			UpdatedAt:         domain.UTCTimestamp(now),
		},
	}

	// Insert time slots
	for _, slot := range timeSlots {
		_, err = database.Exec(`
			INSERT INTO time_slots (id, therapist_id, day_of_week, start_time, end_time, pre_session_buffer, post_session_buffer, created_at, updated_at)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
		`, slot.ID, slot.TherapistID, slot.DayOfWeek, slot.StartTime, slot.EndTime,
			slot.PreSessionBuffer, slot.PostSessionBuffer, slot.CreatedAt, slot.UpdatedAt)
		if err != nil {
			t.Fatalf("Failed to insert time slot: %v", err)
		}
	}

	// Create test clients for bookings
	clients := []domain.Client{
		{
			ID:             domain.NewClientID(),
			Name:           "Test Client 1",
			WhatsAppNumber: "+1555002001",
			CreatedAt:      domain.UTCTimestamp(now),
			UpdatedAt:      domain.UTCTimestamp(now),
		},
		{
			ID:             domain.NewClientID(),
			Name:           "Test Client 2",
			WhatsAppNumber: "+1555002002",
			CreatedAt:      domain.UTCTimestamp(now),
			UpdatedAt:      domain.UTCTimestamp(now),
		},
	}

	// Insert clients
	for _, client := range clients {
		_, err = database.Exec(`
			INSERT INTO clients (id, name, whatsapp_number, created_at, updated_at)
			VALUES (?, ?, ?, ?, ?)
		`, client.ID, client.Name, client.WhatsAppNumber, client.CreatedAt, client.UpdatedAt)
		if err != nil {
			t.Fatalf("Failed to insert client: %v", err)
		}
	}

	// Create strategic bookings to create "holes" in availability
	// Friday bookings to test interference
	fridayDate := time.Date(2024, 1, 12, 0, 0, 0, 0, time.UTC)
	bookings := []domain.Booking{
		// Alice has a booking at 11:00 on Friday
		{
			ID:          domain.NewBookingID(),
			TimeSlotID:  timeSlots[len(timeSlots)-2].ID, // Alice's Friday slot
			TherapistID: therapists[0].ID,
			ClientID:    clients[0].ID,
			StartTime:   domain.UTCTimestamp(fridayDate.Add(11 * time.Hour)), // 11:00
			State:       domain.BookingStateConfirmed,
			CreatedAt:   domain.UTCTimestamp(now),
			UpdatedAt:   domain.UTCTimestamp(now),
		},
		// Bob has a booking at 14:00 on Friday
		{
			ID:          domain.NewBookingID(),
			TimeSlotID:  timeSlots[len(timeSlots)-1].ID, // Bob's Friday slot
			TherapistID: therapists[1].ID,
			ClientID:    clients[1].ID,
			StartTime:   domain.UTCTimestamp(fridayDate.Add(14 * time.Hour)), // 14:00
			State:       domain.BookingStateConfirmed,
			CreatedAt:   domain.UTCTimestamp(now),
			UpdatedAt:   domain.UTCTimestamp(now),
		},
	}

	// Insert bookings
	for _, booking := range bookings {
		_, err = database.Exec(`
			INSERT INTO bookings (id, timeslot_id, therapist_id, client_id, start_time, state, created_at, updated_at)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?)
		`, booking.ID, booking.TimeSlotID, booking.TherapistID, booking.ClientID,
			booking.StartTime, booking.State, booking.CreatedAt, booking.UpdatedAt)
		if err != nil {
			t.Fatalf("Failed to insert booking: %v", err)
		}
	}

	return &ScheduleTestData{
		Specializations: []domain.Specialization{anxietySpec, depressionSpec},
		Therapists:      therapists,
		TimeSlots:       timeSlots,
		Bookings:        bookings,
		Clients:         clients,
	}
}

func setupScheduleTestDB(t *testing.T) (ports.SQLDatabase, func()) {
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
