package timeslot_handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mishkahtherapy/brain/adapters/api/internal/testutils"
	"github.com/mishkahtherapy/brain/core/domain/timeslot"
	"github.com/mishkahtherapy/brain/core/usecases/timeslot/bulk_toggle_therapist_timeslots"
	"github.com/mishkahtherapy/brain/core/usecases/timeslot/create_therapist_timeslot"
	"github.com/mishkahtherapy/brain/core/usecases/timeslot/delete_therapist_timeslot"
	"github.com/mishkahtherapy/brain/core/usecases/timeslot/get_therapist_timeslot"
	"github.com/mishkahtherapy/brain/core/usecases/timeslot/list_therapist_timeslots"
	"github.com/mishkahtherapy/brain/core/usecases/timeslot/update_therapist_timeslot"

	_ "github.com/glebarez/go-sqlite"
)

func TestTimeslotE2E(t *testing.T) {
	// Setup test database using utilities
	database, cleanup := testutils.SetupTestDB(t)
	defer cleanup()

	// Insert test therapist using utilities
	testTherapistID := testutils.CreateTestTherapist(t, database)

	// Setup repositories using utilities
	repos := testutils.SetupRepositories(database)

	// Setup usecases (test-specific logic remains explicit)
	createUsecase := create_therapist_timeslot.NewUsecase(repos.TherapistRepo, repos.TimeSlotRepo)
	getUsecase := get_therapist_timeslot.NewUsecase(repos.TherapistRepo, repos.TimeSlotRepo)
	updateUsecase := update_therapist_timeslot.NewUsecase(repos.TherapistRepo, repos.TimeSlotRepo)
	deleteUsecase := delete_therapist_timeslot.NewUsecase(repos.TherapistRepo, repos.TimeSlotRepo)
	listUsecase := list_therapist_timeslots.NewUsecase(repos.TherapistRepo, repos.TimeSlotRepo)
	bulkToggleUsecase := bulk_toggle_therapist_timeslots.NewUsecase(repos.TherapistRepo, repos.TimeSlotRepo)

	// Setup handler
	timeslotHandler := NewTimeslotHandler(
		bulkToggleUsecase,
		*createUsecase,
		*getUsecase,
		*updateUsecase,
		*deleteUsecase,
		*listUsecase,
	)

	// Setup router
	mux := http.NewServeMux()
	timeslotHandler.RegisterRoutes(mux)

	// Test timezone offset (GMT+3 Cairo)
	const testTimezoneOffset = 180

	t.Run("Complete timeslot workflow", func(t *testing.T) {
		// Step 1: Create a timeslot (local time: 14:00-17:00 becomes 11:00-14:00 UTC)
		timeslotData := map[string]interface{}{
			"therapistId":       string(testTherapistID),
			"dayOfWeek":         "Monday",
			"startTime":         "14:00", // Local time (Cairo: GMT+3)
			"durationMinutes":   180,     // 3 hours duration
			"timezoneOffset":    testTimezoneOffset,
			"preSessionBuffer":  15,
			"postSessionBuffer": 30,
		}
		timeslotBody, _ := json.Marshal(timeslotData)

		createReq := httptest.NewRequest("POST", "/api/v1/therapists/"+string(testTherapistID)+"/timeslots?timezoneOffset="+string(rune(testTimezoneOffset)), bytes.NewBuffer(timeslotBody))
		createReq.Header.Set("Content-Type", "application/json")
		createRec := httptest.NewRecorder()

		mux.ServeHTTP(createRec, createReq)

		// Verify creation response using utilities
		var createdTimeslot timeslot.TimeSlot
		testutils.AssertJSONResponse(t, createRec, http.StatusCreated, &createdTimeslot)

		// Verify created timeslot data (should be converted back to local timezone for response)
		if createdTimeslot.TherapistID != testTherapistID {
			t.Errorf("Expected therapist ID %s, got %s", testTherapistID, createdTimeslot.TherapistID)
		}
		if createdTimeslot.DayOfWeek != timeslot.DayOfWeekMonday {
			t.Errorf("Expected day %s, got %s", timeslot.DayOfWeekMonday, createdTimeslot.DayOfWeek)
		}
		if createdTimeslot.StartTime != "14:00" {
			t.Errorf("Expected start time %s, got %s", "14:00", createdTimeslot.StartTime)
		}
		if createdTimeslot.DurationMinutes != 180 {
			t.Errorf("Expected duration %d, got %d", 180, createdTimeslot.DurationMinutes)
		}
		if createdTimeslot.PreSessionBuffer != 15 {
			t.Errorf("Expected pre-session buffer %d, got %d", 15, createdTimeslot.PreSessionBuffer)
		}
		if createdTimeslot.PostSessionBuffer != 30 {
			t.Errorf("Expected post-session buffer %d, got %d", 30, createdTimeslot.PostSessionBuffer)
		}
		if createdTimeslot.ID == "" {
			t.Error("Expected ID to be set")
		}
		if len(createdTimeslot.BookingIDs) != 0 {
			t.Errorf("Expected empty booking IDs, got %v", createdTimeslot.BookingIDs)
		}
		if !createdTimeslot.IsActive {
			t.Error("Expected new timeslot to be active by default")
		}

		// Step 2: Get the timeslot by ID (with timezone conversion)
		getReq := httptest.NewRequest("GET", "/api/v1/therapists/"+string(testTherapistID)+"/timeslots/"+string(createdTimeslot.ID)+"?timezoneOffset=180", nil)
		getRec := httptest.NewRecorder()

		mux.ServeHTTP(getRec, getReq)

		// Verify get response using utilities
		var retrievedTimeslot timeslot.TimeSlot
		testutils.AssertJSONResponse(t, getRec, http.StatusOK, &retrievedTimeslot)

		// Verify retrieved timeslot matches created one
		if retrievedTimeslot.ID != createdTimeslot.ID {
			t.Errorf("Expected ID %s, got %s", createdTimeslot.ID, retrievedTimeslot.ID)
		}
		if retrievedTimeslot.DayOfWeek != createdTimeslot.DayOfWeek {
			t.Errorf("Expected day %s, got %s", createdTimeslot.DayOfWeek, retrievedTimeslot.DayOfWeek)
		}
		if retrievedTimeslot.StartTime != createdTimeslot.StartTime {
			t.Errorf("Expected start time %s, got %s", createdTimeslot.StartTime, retrievedTimeslot.StartTime)
		}
		if retrievedTimeslot.DurationMinutes != createdTimeslot.DurationMinutes {
			t.Errorf("Expected duration %d, got %d", createdTimeslot.DurationMinutes, retrievedTimeslot.DurationMinutes)
		}

		// Step 5: Update the timeslot
		updateData := map[string]interface{}{
			"dayOfWeek":         "Wednesday",
			"startTime":         "15:00",
			"durationMinutes":   180, // 3 hours
			"preSessionBuffer":  10,
			"postSessionBuffer": 30,
			"isActive":          true,
		}
		updateBody, _ := json.Marshal(updateData)

		updateReq := httptest.NewRequest("PUT", "/api/v1/therapists/"+string(testTherapistID)+"/timeslots/"+string(createdTimeslot.ID)+"?timezoneOffset=180", bytes.NewBuffer(updateBody))
		updateReq.Header.Set("Content-Type", "application/json")
		updateRec := httptest.NewRecorder()

		mux.ServeHTTP(updateRec, updateReq)

		if updateRec.Code != http.StatusOK {
			t.Fatalf("Expected status %d for update, got %d. Body: %s", http.StatusOK, updateRec.Code, updateRec.Body.String())
		}

		// Verify update response using utilities
		var updatedTimeslot timeslot.TimeSlot
		testutils.AssertJSONResponse(t, updateRec, http.StatusOK, &updatedTimeslot)

		// Verify updated timeslot data
		if updatedTimeslot.ID != createdTimeslot.ID {
			t.Errorf("Expected ID to remain %s, got %s", createdTimeslot.ID, updatedTimeslot.ID)
		}
		if updatedTimeslot.DayOfWeek != timeslot.DayOfWeekWednesday {
			t.Errorf("Expected updated day %s, got %s", timeslot.DayOfWeekWednesday, updatedTimeslot.DayOfWeek)
		}
		if updatedTimeslot.StartTime != "15:00" {
			t.Errorf("Expected updated start time %s, got %s", "15:00", updatedTimeslot.StartTime)
		}
		if updatedTimeslot.DurationMinutes != 180 {
			t.Errorf("Expected updated duration %d, got %d", 180, updatedTimeslot.DurationMinutes)
		}
		if updatedTimeslot.PostSessionBuffer != 30 {
			t.Errorf("Expected updated post-session buffer %d, got %d", 30, updatedTimeslot.PostSessionBuffer)
		}
		if !updatedTimeslot.IsActive {
			t.Error("Expected timeslot to be active after update")
		}

		// Step 3: List all timeslots for therapist (with timezone conversion)
		listAllReq := httptest.NewRequest("GET", "/api/v1/therapists/"+string(testTherapistID)+"/timeslots?timezoneOffset=180", nil)
		listAllRec := httptest.NewRecorder()

		mux.ServeHTTP(listAllRec, listAllReq)

		// Verify list response using utilities
		var listResponse []timeslot.TimeSlot
		testutils.AssertJSONResponse(t, listAllRec, http.StatusOK, &listResponse)

		// Verify our timeslot is in the list
		found := false
		for _, ts := range listResponse {
			if ts.ID == createdTimeslot.ID {
				found = true
				if ts.DayOfWeek != timeslot.DayOfWeekWednesday {
					t.Errorf("Expected listed timeslot to have updated day %s, got %s", timeslot.DayOfWeekWednesday, ts.DayOfWeek)
				}
				break
			}
		}
		if !found {
			t.Error("Created timeslot not found in list of all timeslots")
		}

		// Step 5: Delete the original timeslot
		deleteReq := httptest.NewRequest("DELETE", "/api/v1/therapists/"+string(testTherapistID)+"/timeslots/"+string(createdTimeslot.ID), nil)
		deleteRec := httptest.NewRecorder()

		mux.ServeHTTP(deleteRec, deleteReq)

		// Verify delete response
		if deleteRec.Code != http.StatusNoContent {
			t.Fatalf("Expected status %d, got %d. Body: %s", http.StatusNoContent, deleteRec.Code, deleteRec.Body.String())
		}

		// Step 6: Verify timeslot is deleted
		getDeletedReq := httptest.NewRequest("GET", "/api/v1/therapists/"+string(testTherapistID)+"/timeslots/"+string(createdTimeslot.ID)+"?timezoneOffset=180", nil)
		getDeletedRec := httptest.NewRecorder()

		mux.ServeHTTP(getDeletedRec, getDeletedReq)

		if getDeletedRec.Code != http.StatusNotFound {
			t.Errorf("Expected status %d for deleted timeslot, got %d", http.StatusNotFound, getDeletedRec.Code)
		}
	})

	t.Run("Business rule validation", func(t *testing.T) {
		// Test 30-minute minimum post-session buffer rule
		invalidBufferData := map[string]interface{}{
			"therapistId":       string(testTherapistID),
			"dayOfWeek":         "Friday",
			"startTime":         "09:00",
			"durationMinutes":   480, // 8 hours
			"timezoneOffset":    testTimezoneOffset,
			"preSessionBuffer":  15,
			"postSessionBuffer": 20, // Invalid: less than 30 minutes
		}
		invalidBufferBody, _ := json.Marshal(invalidBufferData)

		invalidBufferReq := httptest.NewRequest("POST", "/api/v1/therapists/"+string(testTherapistID)+"/timeslots?timezoneOffset=180", bytes.NewBuffer(invalidBufferBody))
		invalidBufferReq.Header.Set("Content-Type", "application/json")
		invalidBufferRec := httptest.NewRecorder()

		mux.ServeHTTP(invalidBufferRec, invalidBufferReq)

		if invalidBufferRec.Code != http.StatusBadRequest {
			t.Errorf("Expected status %d for invalid post-session buffer, got %d", http.StatusBadRequest, invalidBufferRec.Code)
		}

		// Test negative pre-session buffer
		negativeBufferData := map[string]interface{}{
			"therapistId":       string(testTherapistID),
			"dayOfWeek":         "Friday",
			"startTime":         "09:00",
			"durationMinutes":   480,
			"timezoneOffset":    testTimezoneOffset,
			"preSessionBuffer":  -5, // Invalid: negative
			"postSessionBuffer": 30,
		}
		negativeBufferBody, _ := json.Marshal(negativeBufferData)

		negativeBufferReq := httptest.NewRequest("POST", "/api/v1/therapists/"+string(testTherapistID)+"/timeslots?timezoneOffset=180", bytes.NewBuffer(negativeBufferBody))
		negativeBufferReq.Header.Set("Content-Type", "application/json")
		negativeBufferRec := httptest.NewRecorder()

		mux.ServeHTTP(negativeBufferRec, negativeBufferReq)

		if negativeBufferRec.Code != http.StatusBadRequest {
			t.Errorf("Expected status %d for negative pre-session buffer, got %d", http.StatusBadRequest, negativeBufferRec.Code)
		}

		// Test overlapping timeslots
		// First create a valid timeslot
		firstTimeslotData := map[string]interface{}{
			"therapistId":       string(testTherapistID),
			"dayOfWeek":         "Saturday",
			"startTime":         "10:00",
			"durationMinutes":   240, // 4 hours (10:00-14:00)
			"timezoneOffset":    testTimezoneOffset,
			"preSessionBuffer":  0,
			"postSessionBuffer": 30,
		}
		firstTimeslotBody, _ := json.Marshal(firstTimeslotData)

		firstTimeslotReq := httptest.NewRequest("POST", "/api/v1/therapists/"+string(testTherapistID)+"/timeslots?timezoneOffset=180", bytes.NewBuffer(firstTimeslotBody))
		firstTimeslotReq.Header.Set("Content-Type", "application/json")
		firstTimeslotRec := httptest.NewRecorder()

		mux.ServeHTTP(firstTimeslotRec, firstTimeslotReq)

		if firstTimeslotRec.Code != http.StatusCreated {
			t.Fatalf("Failed to create first timeslot for overlap test: %d", firstTimeslotRec.Code)
		}

		// Now try to create an overlapping timeslot
		overlappingData := map[string]interface{}{
			"therapistId":       string(testTherapistID),
			"dayOfWeek":         "Saturday",
			"startTime":         "12:00", // Overlaps with existing Saturday 10:00-14:00
			"durationMinutes":   240,     // 4 hours (12:00-16:00)
			"timezoneOffset":    testTimezoneOffset,
			"preSessionBuffer":  0,
			"postSessionBuffer": 30,
		}
		overlappingBody, _ := json.Marshal(overlappingData)

		overlappingReq := httptest.NewRequest("POST", "/api/v1/therapists/"+string(testTherapistID)+"/timeslots?timezoneOffset=180", bytes.NewBuffer(overlappingBody))
		overlappingReq.Header.Set("Content-Type", "application/json")
		overlappingRec := httptest.NewRecorder()

		mux.ServeHTTP(overlappingRec, overlappingReq)

		if overlappingRec.Code != http.StatusConflict {
			t.Errorf("Expected status %d for overlapping timeslot, got %d", http.StatusConflict, overlappingRec.Code)
		}
	})

	t.Run("Duration validation", func(t *testing.T) {
		// Test zero duration
		zeroDurationData := map[string]interface{}{
			"therapistId":       string(testTherapistID),
			"dayOfWeek":         "Sunday",
			"startTime":         "09:00",
			"durationMinutes":   0, // Invalid: zero duration
			"timezoneOffset":    testTimezoneOffset,
			"preSessionBuffer":  0,
			"postSessionBuffer": 30,
		}
		zeroDurationBody, _ := json.Marshal(zeroDurationData)

		zeroDurationReq := httptest.NewRequest("POST", "/api/v1/therapists/"+string(testTherapistID)+"/timeslots?timezoneOffset=180", bytes.NewBuffer(zeroDurationBody))
		zeroDurationReq.Header.Set("Content-Type", "application/json")
		zeroDurationRec := httptest.NewRecorder()

		mux.ServeHTTP(zeroDurationRec, zeroDurationReq)

		if zeroDurationRec.Code != http.StatusBadRequest {
			t.Errorf("Expected status %d for zero duration, got %d", http.StatusBadRequest, zeroDurationRec.Code)
		}

		// Test negative duration
		negativeDurationData := map[string]interface{}{
			"therapistId":       string(testTherapistID),
			"dayOfWeek":         "Sunday",
			"startTime":         "09:00",
			"durationMinutes":   -60, // Invalid: negative duration
			"timezoneOffset":    testTimezoneOffset,
			"preSessionBuffer":  0,
			"postSessionBuffer": 30,
		}
		negativeDurationBody, _ := json.Marshal(negativeDurationData)

		negativeDurationReq := httptest.NewRequest("POST", "/api/v1/therapists/"+string(testTherapistID)+"/timeslots?timezoneOffset=180", bytes.NewBuffer(negativeDurationBody))
		negativeDurationReq.Header.Set("Content-Type", "application/json")
		negativeDurationRec := httptest.NewRecorder()

		mux.ServeHTTP(negativeDurationRec, negativeDurationReq)

		if negativeDurationRec.Code != http.StatusBadRequest {
			t.Errorf("Expected status %d for negative duration, got %d", http.StatusBadRequest, negativeDurationRec.Code)
		}

		// Test duration over 24 hours
		overDayDurationData := map[string]interface{}{
			"therapistId":       string(testTherapistID),
			"dayOfWeek":         "Sunday",
			"startTime":         "09:00",
			"durationMinutes":   1500, // Invalid: over 24 hours (1440 minutes)
			"timezoneOffset":    testTimezoneOffset,
			"preSessionBuffer":  0,
			"postSessionBuffer": 30,
		}
		overDayDurationBody, _ := json.Marshal(overDayDurationData)

		overDayDurationReq := httptest.NewRequest("POST", "/api/v1/therapists/"+string(testTherapistID)+"/timeslots?timezoneOffset=180", bytes.NewBuffer(overDayDurationBody))
		overDayDurationReq.Header.Set("Content-Type", "application/json")
		overDayDurationRec := httptest.NewRecorder()

		mux.ServeHTTP(overDayDurationRec, overDayDurationReq)

		if overDayDurationRec.Code != http.StatusBadRequest {
			t.Errorf("Expected status %d for over-day duration, got %d", http.StatusBadRequest, overDayDurationRec.Code)
		}
	})

	t.Run("Timezone validation", func(t *testing.T) {
		// Test invalid timezone offset (too positive)
		invalidPositiveOffsetData := map[string]interface{}{
			"therapistId":       string(testTherapistID),
			"dayOfWeek":         "Monday",
			"startTime":         "14:00",
			"durationMinutes":   60,
			"timezoneOffset":    900, // Invalid: +15 hours (max is +14)
			"preSessionBuffer":  0,
			"postSessionBuffer": 30,
		}
		invalidPositiveOffsetBody, _ := json.Marshal(invalidPositiveOffsetData)

		invalidPositiveOffsetReq := httptest.NewRequest("POST", "/api/v1/therapists/"+string(testTherapistID)+"/timeslots?timezoneOffset=180", bytes.NewBuffer(invalidPositiveOffsetBody))
		invalidPositiveOffsetReq.Header.Set("Content-Type", "application/json")
		invalidPositiveOffsetRec := httptest.NewRecorder()

		mux.ServeHTTP(invalidPositiveOffsetRec, invalidPositiveOffsetReq)

		if invalidPositiveOffsetRec.Code != http.StatusBadRequest {
			t.Errorf("Expected status %d for invalid positive timezone offset, got %d", http.StatusBadRequest, invalidPositiveOffsetRec.Code)
		}

		// Test invalid timezone offset (too negative)
		invalidNegativeOffsetData := map[string]interface{}{
			"therapistId":       string(testTherapistID),
			"dayOfWeek":         "Monday",
			"startTime":         "14:00",
			"durationMinutes":   60,
			"timezoneOffset":    -780, // Invalid: -13 hours (min is -12)
			"preSessionBuffer":  0,
			"postSessionBuffer": 30,
		}
		invalidNegativeOffsetBody, _ := json.Marshal(invalidNegativeOffsetData)

		invalidNegativeOffsetReq := httptest.NewRequest("POST", "/api/v1/therapists/"+string(testTherapistID)+"/timeslots?timezoneOffset=180", bytes.NewBuffer(invalidNegativeOffsetBody))
		invalidNegativeOffsetReq.Header.Set("Content-Type", "application/json")
		invalidNegativeOffsetRec := httptest.NewRecorder()

		mux.ServeHTTP(invalidNegativeOffsetRec, invalidNegativeOffsetReq)

		if invalidNegativeOffsetRec.Code != http.StatusBadRequest {
			t.Errorf("Expected status %d for invalid negative timezone offset, got %d", http.StatusBadRequest, invalidNegativeOffsetRec.Code)
		}

		// Test missing timezone offset in query parameter for GET request
		missingTimezoneReq := httptest.NewRequest("GET", "/api/v1/therapists/"+string(testTherapistID)+"/timeslots", nil)
		missingTimezoneRec := httptest.NewRecorder()

		mux.ServeHTTP(missingTimezoneRec, missingTimezoneReq)

		if missingTimezoneRec.Code != http.StatusBadRequest {
			t.Errorf("Expected status %d for missing timezone offset in query, got %d", http.StatusBadRequest, missingTimezoneRec.Code)
		}
	})

	t.Run("Error cases", func(t *testing.T) {
		// Test with non-existent therapist
		nonExistentTherapistID := "therapist_00000000-0000-0000-0000-000000000000"

		timeslotData := map[string]interface{}{
			"therapistId":       nonExistentTherapistID,
			"dayOfWeek":         "Monday",
			"startTime":         "09:00",
			"durationMinutes":   480,
			"timezoneOffset":    testTimezoneOffset,
			"preSessionBuffer":  0,
			"postSessionBuffer": 30,
		}
		timeslotBody, _ := json.Marshal(timeslotData)

		nonExistentReq := httptest.NewRequest("POST", "/api/v1/therapists/"+nonExistentTherapistID+"/timeslots?timezoneOffset=180", bytes.NewBuffer(timeslotBody))
		nonExistentReq.Header.Set("Content-Type", "application/json")
		nonExistentRec := httptest.NewRecorder()

		mux.ServeHTTP(nonExistentRec, nonExistentReq)

		if nonExistentRec.Code != http.StatusNotFound {
			t.Errorf("Expected status %d for non-existent therapist, got %d", http.StatusNotFound, nonExistentRec.Code)
		}

		// Test get non-existent timeslot
		nonExistentTimeslotID := "timeslot_00000000-0000-0000-0000-000000000000"
		getReq := httptest.NewRequest("GET", "/api/v1/therapists/"+string(testTherapistID)+"/timeslots/"+nonExistentTimeslotID+"?timezoneOffset=180", nil)
		getRec := httptest.NewRecorder()

		mux.ServeHTTP(getRec, getReq)

		// Use utility for error checking
		testutils.AssertStatus(t, getRec, http.StatusNotFound)

		// Test missing required fields
		missingFieldsData := map[string]interface{}{
			"dayOfWeek": "Monday",
			// Missing startTime, durationMinutes, timezone info, buffers
		}
		missingFieldsBody, _ := json.Marshal(missingFieldsData)

		missingFieldsReq := httptest.NewRequest("POST", "/api/v1/therapists/"+string(testTherapistID)+"/timeslots?timezoneOffset=180", bytes.NewBuffer(missingFieldsBody))
		missingFieldsReq.Header.Set("Content-Type", "application/json")
		missingFieldsRec := httptest.NewRecorder()

		mux.ServeHTTP(missingFieldsRec, missingFieldsReq)

		// Use utility for error checking
		testutils.AssertStatus(t, missingFieldsRec, http.StatusBadRequest)

		// Test invalid day of week
		invalidDayData := map[string]interface{}{
			"therapistId":       string(testTherapistID),
			"dayOfWeek":         "InvalidDay",
			"startTime":         "09:00",
			"durationMinutes":   480,
			"timezoneOffset":    testTimezoneOffset,
			"preSessionBuffer":  0,
			"postSessionBuffer": 30,
		}
		invalidDayBody, _ := json.Marshal(invalidDayData)

		invalidDayReq := httptest.NewRequest("POST", "/api/v1/therapists/"+string(testTherapistID)+"/timeslots?timezoneOffset=180", bytes.NewBuffer(invalidDayBody))
		invalidDayReq.Header.Set("Content-Type", "application/json")
		invalidDayRec := httptest.NewRecorder()

		mux.ServeHTTP(invalidDayRec, invalidDayReq)

		// Use utility for error checking
		testutils.AssertStatus(t, invalidDayRec, http.StatusBadRequest)

		// Test invalid JSON
		invalidJSONReq := httptest.NewRequest("POST", "/api/v1/therapists/"+string(testTherapistID)+"/timeslots?timezoneOffset=180", bytes.NewBufferString("{invalid json"))
		invalidJSONReq.Header.Set("Content-Type", "application/json")
		invalidJSONRec := httptest.NewRecorder()

		mux.ServeHTTP(invalidJSONRec, invalidJSONReq)

		// Use utility for error checking
		testutils.AssertStatus(t, invalidJSONRec, http.StatusBadRequest)
	})

	t.Run("IsActive field toggle", func(t *testing.T) {
		// Create a timeslot (should be active by default)
		timeslotData := map[string]interface{}{
			"therapistId":       string(testTherapistID),
			"dayOfWeek":         "Friday",
			"startTime":         "14:00",
			"durationMinutes":   60,
			"timezoneOffset":    testTimezoneOffset,
			"preSessionBuffer":  0,
			"postSessionBuffer": 30,
		}
		timeslotBody, _ := json.Marshal(timeslotData)

		createReq := httptest.NewRequest("POST", "/api/v1/therapists/"+string(testTherapistID)+"/timeslots?timezoneOffset=180", bytes.NewBuffer(timeslotBody))
		createReq.Header.Set("Content-Type", "application/json")
		createRec := httptest.NewRecorder()

		mux.ServeHTTP(createRec, createReq)

		// Use utility for JSON parsing
		var createdTimeslot timeslot.TimeSlot
		testutils.AssertJSONResponse(t, createRec, http.StatusCreated, &createdTimeslot)

		// Verify it's active by default
		if !createdTimeslot.IsActive {
			t.Error("Expected new timeslot to be active by default")
		}

		// Update to inactive
		updateToInactive := map[string]interface{}{
			"dayOfWeek":         "Friday",
			"startTime":         "14:00",
			"durationMinutes":   60,
			"preSessionBuffer":  0,
			"postSessionBuffer": 30,
			"isActive":          false,
		}
		updateBody, _ := json.Marshal(updateToInactive)

		updateReq := httptest.NewRequest("PUT", "/api/v1/therapists/"+string(testTherapistID)+"/timeslots/"+string(createdTimeslot.ID)+"?timezoneOffset=180", bytes.NewBuffer(updateBody))
		updateReq.Header.Set("Content-Type", "application/json")
		updateRec := httptest.NewRecorder()

		mux.ServeHTTP(updateRec, updateReq)

		// Use utility for JSON parsing
		var inactiveTimeslot timeslot.TimeSlot
		testutils.AssertJSONResponse(t, updateRec, http.StatusOK, &inactiveTimeslot)

		// Verify it's now inactive
		if inactiveTimeslot.IsActive {
			t.Error("Expected timeslot to be inactive after update")
		}

		// Update back to active
		updateToActive := map[string]interface{}{
			"dayOfWeek":         "Friday",
			"startTime":         "14:00",
			"durationMinutes":   60,
			"preSessionBuffer":  0,
			"postSessionBuffer": 30,
			"isActive":          true,
		}
		updateActiveBody, _ := json.Marshal(updateToActive)

		updateActiveReq := httptest.NewRequest("PUT", "/api/v1/therapists/"+string(testTherapistID)+"/timeslots/"+string(createdTimeslot.ID)+"?timezoneOffset=180", bytes.NewBuffer(updateActiveBody))
		updateActiveReq.Header.Set("Content-Type", "application/json")
		updateActiveRec := httptest.NewRecorder()

		mux.ServeHTTP(updateActiveRec, updateActiveReq)

		// Use utility for JSON parsing
		var activeTimeslot timeslot.TimeSlot
		testutils.AssertJSONResponse(t, updateActiveRec, http.StatusOK, &activeTimeslot)

		// Verify it's active again
		if !activeTimeslot.IsActive {
			t.Error("Expected timeslot to be active after re-activation")
		}
	})

	t.Run("Cross-day timezone conversion", func(t *testing.T) {
		// Test cross-day scenario: Monday 1:30 AM local → Sunday 22:30 UTC
		crossDayData := map[string]interface{}{
			"therapistId":       string(testTherapistID),
			"dayOfWeek":         "Monday",
			"startTime":         "01:30", // Local time (Cairo: GMT+3)
			"durationMinutes":   120,     // 2 hours
			"timezoneOffset":    testTimezoneOffset,
			"preSessionBuffer":  0,
			"postSessionBuffer": 30,
		}
		crossDayBody, _ := json.Marshal(crossDayData)

		crossDayReq := httptest.NewRequest("POST", "/api/v1/therapists/"+string(testTherapistID)+"/timeslots?timezoneOffset=180", bytes.NewBuffer(crossDayBody))
		crossDayReq.Header.Set("Content-Type", "application/json")
		crossDayRec := httptest.NewRecorder()

		mux.ServeHTTP(crossDayRec, crossDayReq)

		// Use utility for JSON parsing
		var crossDayTimeslot timeslot.TimeSlot
		testutils.AssertJSONResponse(t, crossDayRec, http.StatusCreated, &crossDayTimeslot)

		// Verify the response is still in local timezone (shows as Monday)
		if crossDayTimeslot.DayOfWeek != timeslot.DayOfWeekMonday {
			t.Errorf("Expected local day Monday, got %s", crossDayTimeslot.DayOfWeek)
		}
		if crossDayTimeslot.StartTime != "01:30" {
			t.Errorf("Expected local start time 01:30, got %s", crossDayTimeslot.StartTime)
		}
		if crossDayTimeslot.DurationMinutes != 120 {
			t.Errorf("Expected duration 120 minutes, got %d", crossDayTimeslot.DurationMinutes)
		}
	})
}
