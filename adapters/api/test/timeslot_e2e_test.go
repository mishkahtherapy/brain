package test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/mishkahtherapy/brain/adapters/api"
	"github.com/mishkahtherapy/brain/adapters/db"
	"github.com/mishkahtherapy/brain/adapters/db/therapist_db"
	"github.com/mishkahtherapy/brain/adapters/db/timeslot_db"
	"github.com/mishkahtherapy/brain/core/domain"
	"github.com/mishkahtherapy/brain/core/domain/timeslot"
	"github.com/mishkahtherapy/brain/core/ports"
	"github.com/mishkahtherapy/brain/core/usecases/timeslot/create_therapist_timeslot"
	"github.com/mishkahtherapy/brain/core/usecases/timeslot/delete_therapist_timeslot"
	"github.com/mishkahtherapy/brain/core/usecases/timeslot/get_therapist_timeslot"
	"github.com/mishkahtherapy/brain/core/usecases/timeslot/list_therapist_timeslots"
	"github.com/mishkahtherapy/brain/core/usecases/timeslot/update_therapist_timeslot"

	_ "github.com/glebarez/go-sqlite"
)

func TestTimeslotE2E(t *testing.T) {
	// Setup test database
	database, cleanup := setupTimeslotTestDB(t)
	defer cleanup()

	// Insert test therapist
	testTherapistID := insertTestTherapist(t, database)

	// Setup repositories
	therapistRepo := therapist_db.NewTherapistRepository(database)
	timeslotRepo := timeslot_db.NewTimeSlotRepository(database)

	// Setup usecases
	createUsecase := create_therapist_timeslot.NewUsecase(therapistRepo, timeslotRepo)
	getUsecase := get_therapist_timeslot.NewUsecase(therapistRepo, timeslotRepo)
	updateUsecase := update_therapist_timeslot.NewUsecase(therapistRepo, timeslotRepo)
	deleteUsecase := delete_therapist_timeslot.NewUsecase(therapistRepo, timeslotRepo)
	listUsecase := list_therapist_timeslots.NewUsecase(therapistRepo, timeslotRepo)

	// Setup handler
	timeslotHandler := api.NewTimeslotHandler(
		*createUsecase,
		*getUsecase,
		*updateUsecase,
		*deleteUsecase,
		*listUsecase,
	)

	// Setup router
	mux := http.NewServeMux()
	timeslotHandler.RegisterRoutes(mux)

	t.Run("Complete timeslot workflow", func(t *testing.T) {
		// Step 1: Create a timeslot
		timeslotData := map[string]interface{}{
			"dayOfWeek":         "Monday",
			"startTime":         "09:00",
			"endTime":           "17:00",
			"preSessionBuffer":  15,
			"postSessionBuffer": 30,
		}
		timeslotBody, _ := json.Marshal(timeslotData)

		createReq := httptest.NewRequest("POST", "/api/v1/therapists/"+string(testTherapistID)+"/timeslots", bytes.NewBuffer(timeslotBody))
		createReq.Header.Set("Content-Type", "application/json")
		createRec := httptest.NewRecorder()

		mux.ServeHTTP(createRec, createReq)

		// Verify creation response
		if createRec.Code != http.StatusCreated {
			t.Fatalf("Expected status %d, got %d. Body: %s", http.StatusCreated, createRec.Code, createRec.Body.String())
		}

		// Parse created timeslot
		var createdTimeslot timeslot.TimeSlot
		if err := json.Unmarshal(createRec.Body.Bytes(), &createdTimeslot); err != nil {
			t.Fatalf("Failed to parse created timeslot: %v", err)
		}

		// Verify created timeslot data
		if createdTimeslot.TherapistID != testTherapistID {
			t.Errorf("Expected therapist ID %s, got %s", testTherapistID, createdTimeslot.TherapistID)
		}
		if createdTimeslot.DayOfWeek != timeslot.DayOfWeekMonday {
			t.Errorf("Expected day %s, got %s", timeslot.DayOfWeekMonday, createdTimeslot.DayOfWeek)
		}
		if createdTimeslot.StartTime != "09:00" {
			t.Errorf("Expected start time %s, got %s", "09:00", createdTimeslot.StartTime)
		}
		if createdTimeslot.EndTime != "17:00" {
			t.Errorf("Expected end time %s, got %s", "17:00", createdTimeslot.EndTime)
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

		// Step 2: Get the timeslot by ID
		getReq := httptest.NewRequest("GET", "/api/v1/therapists/"+string(testTherapistID)+"/timeslots/"+string(createdTimeslot.ID), nil)
		getRec := httptest.NewRecorder()

		mux.ServeHTTP(getRec, getReq)

		// Verify get response
		if getRec.Code != http.StatusOK {
			t.Fatalf("Expected status %d, got %d. Body: %s", http.StatusOK, getRec.Code, getRec.Body.String())
		}

		// Parse retrieved timeslot
		var retrievedTimeslot timeslot.TimeSlot
		if err := json.Unmarshal(getRec.Body.Bytes(), &retrievedTimeslot); err != nil {
			t.Fatalf("Failed to parse retrieved timeslot: %v", err)
		}

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

		// Step 3: Update the timeslot
		updateData := map[string]interface{}{
			"dayOfWeek":         "Tuesday",
			"startTime":         "10:00",
			"endTime":           "16:00",
			"preSessionBuffer":  10,
			"postSessionBuffer": 45,
		}
		updateBody, _ := json.Marshal(updateData)

		updateReq := httptest.NewRequest("PUT", "/api/v1/therapists/"+string(testTherapistID)+"/timeslots/"+string(createdTimeslot.ID), bytes.NewBuffer(updateBody))
		updateReq.Header.Set("Content-Type", "application/json")
		updateRec := httptest.NewRecorder()

		mux.ServeHTTP(updateRec, updateReq)

		// Verify update response
		if updateRec.Code != http.StatusOK {
			t.Fatalf("Expected status %d, got %d. Body: %s", http.StatusOK, updateRec.Code, updateRec.Body.String())
		}

		// Parse updated timeslot
		var updatedTimeslot timeslot.TimeSlot
		if err := json.Unmarshal(updateRec.Body.Bytes(), &updatedTimeslot); err != nil {
			t.Fatalf("Failed to parse updated timeslot: %v", err)
		}

		// Verify updated timeslot data
		if updatedTimeslot.ID != createdTimeslot.ID {
			t.Errorf("Expected ID to remain %s, got %s", createdTimeslot.ID, updatedTimeslot.ID)
		}
		if updatedTimeslot.DayOfWeek != timeslot.DayOfWeekTuesday {
			t.Errorf("Expected updated day %s, got %s", timeslot.DayOfWeekTuesday, updatedTimeslot.DayOfWeek)
		}
		if updatedTimeslot.StartTime != "10:00" {
			t.Errorf("Expected updated start time %s, got %s", "10:00", updatedTimeslot.StartTime)
		}
		if updatedTimeslot.EndTime != "16:00" {
			t.Errorf("Expected updated end time %s, got %s", "16:00", updatedTimeslot.EndTime)
		}
		if updatedTimeslot.PostSessionBuffer != 45 {
			t.Errorf("Expected updated post-session buffer %d, got %d", 45, updatedTimeslot.PostSessionBuffer)
		}

		// Step 4: List all timeslots for therapist
		listAllReq := httptest.NewRequest("GET", "/api/v1/therapists/"+string(testTherapistID)+"/timeslots", nil)
		listAllRec := httptest.NewRecorder()

		mux.ServeHTTP(listAllRec, listAllReq)

		// Verify list response
		if listAllRec.Code != http.StatusOK {
			t.Fatalf("Expected status %d, got %d. Body: %s", http.StatusOK, listAllRec.Code, listAllRec.Body.String())
		}

		// Parse list response
		var listResponse struct {
			Timeslots []timeslot.TimeSlot `json:"timeslots"`
		}
		if err := json.Unmarshal(listAllRec.Body.Bytes(), &listResponse); err != nil {
			t.Fatalf("Failed to parse list response: %v", err)
		}

		// Verify our timeslot is in the list
		found := false
		for _, ts := range listResponse.Timeslots {
			if ts.ID == createdTimeslot.ID {
				found = true
				if ts.DayOfWeek != timeslot.DayOfWeekTuesday {
					t.Errorf("Expected listed timeslot to have updated day %s, got %s", timeslot.DayOfWeekTuesday, ts.DayOfWeek)
				}
				break
			}
		}
		if !found {
			t.Error("Created timeslot not found in list of all timeslots")
		}

		// Step 5: Create another timeslot for filtering test
		wednesdayData := map[string]interface{}{
			"dayOfWeek":         "Wednesday",
			"startTime":         "14:00",
			"endTime":           "18:00",
			"preSessionBuffer":  0,
			"postSessionBuffer": 30,
		}
		wednesdayBody, _ := json.Marshal(wednesdayData)

		createWedReq := httptest.NewRequest("POST", "/api/v1/therapists/"+string(testTherapistID)+"/timeslots", bytes.NewBuffer(wednesdayBody))
		createWedReq.Header.Set("Content-Type", "application/json")
		createWedRec := httptest.NewRecorder()

		mux.ServeHTTP(createWedRec, createWedReq)

		if createWedRec.Code != http.StatusCreated {
			t.Fatalf("Expected status %d for Wednesday timeslot, got %d. Body: %s", http.StatusCreated, createWedRec.Code, createWedRec.Body.String())
		}

		// Step 6: Test day filtering
		listTuesdayReq := httptest.NewRequest("GET", "/api/v1/therapists/"+string(testTherapistID)+"/timeslots?day=Tuesday", nil)
		listTuesdayRec := httptest.NewRecorder()

		mux.ServeHTTP(listTuesdayRec, listTuesdayReq)

		if listTuesdayRec.Code != http.StatusOK {
			t.Fatalf("Expected status %d for Tuesday filter, got %d. Body: %s", http.StatusOK, listTuesdayRec.Code, listTuesdayRec.Body.String())
		}

		var tuesdayResponse struct {
			Timeslots []timeslot.TimeSlot `json:"timeslots"`
		}
		if err := json.Unmarshal(listTuesdayRec.Body.Bytes(), &tuesdayResponse); err != nil {
			t.Fatalf("Failed to parse Tuesday list response: %v", err)
		}

		// Verify only Tuesday timeslots are returned
		for _, ts := range tuesdayResponse.Timeslots {
			if ts.DayOfWeek != timeslot.DayOfWeekTuesday {
				t.Errorf("Expected only Tuesday timeslots, found %s", ts.DayOfWeek)
			}
		}

		// Step 7: Delete the original timeslot
		deleteReq := httptest.NewRequest("DELETE", "/api/v1/therapists/"+string(testTherapistID)+"/timeslots/"+string(createdTimeslot.ID), nil)
		deleteRec := httptest.NewRecorder()

		mux.ServeHTTP(deleteRec, deleteReq)

		// Verify delete response
		if deleteRec.Code != http.StatusNoContent {
			t.Fatalf("Expected status %d, got %d. Body: %s", http.StatusNoContent, deleteRec.Code, deleteRec.Body.String())
		}

		// Step 8: Verify timeslot is deleted
		getDeletedReq := httptest.NewRequest("GET", "/api/v1/therapists/"+string(testTherapistID)+"/timeslots/"+string(createdTimeslot.ID), nil)
		getDeletedRec := httptest.NewRecorder()

		mux.ServeHTTP(getDeletedRec, getDeletedReq)

		if getDeletedRec.Code != http.StatusNotFound {
			t.Errorf("Expected status %d for deleted timeslot, got %d", http.StatusNotFound, getDeletedRec.Code)
		}
	})

	t.Run("Business rule validation", func(t *testing.T) {
		// Test 30-minute minimum post-session buffer rule
		invalidBufferData := map[string]interface{}{
			"dayOfWeek":         "Friday",
			"startTime":         "09:00",
			"endTime":           "17:00",
			"preSessionBuffer":  15,
			"postSessionBuffer": 20, // Invalid: less than 30 minutes
		}
		invalidBufferBody, _ := json.Marshal(invalidBufferData)

		invalidBufferReq := httptest.NewRequest("POST", "/api/v1/therapists/"+string(testTherapistID)+"/timeslots", bytes.NewBuffer(invalidBufferBody))
		invalidBufferReq.Header.Set("Content-Type", "application/json")
		invalidBufferRec := httptest.NewRecorder()

		mux.ServeHTTP(invalidBufferRec, invalidBufferReq)

		if invalidBufferRec.Code != http.StatusBadRequest {
			t.Errorf("Expected status %d for invalid post-session buffer, got %d", http.StatusBadRequest, invalidBufferRec.Code)
		}

		// Test negative pre-session buffer
		negativeBufferData := map[string]interface{}{
			"dayOfWeek":         "Friday",
			"startTime":         "09:00",
			"endTime":           "17:00",
			"preSessionBuffer":  -5, // Invalid: negative
			"postSessionBuffer": 30,
		}
		negativeBufferBody, _ := json.Marshal(negativeBufferData)

		negativeBufferReq := httptest.NewRequest("POST", "/api/v1/therapists/"+string(testTherapistID)+"/timeslots", bytes.NewBuffer(negativeBufferBody))
		negativeBufferReq.Header.Set("Content-Type", "application/json")
		negativeBufferRec := httptest.NewRecorder()

		mux.ServeHTTP(negativeBufferRec, negativeBufferReq)

		if negativeBufferRec.Code != http.StatusBadRequest {
			t.Errorf("Expected status %d for negative pre-session buffer, got %d", http.StatusBadRequest, negativeBufferRec.Code)
		}

		// Test overlapping timeslots
		// First create a valid timeslot
		firstTimeslotData := map[string]interface{}{
			"dayOfWeek":         "Saturday",
			"startTime":         "10:00",
			"endTime":           "14:00",
			"preSessionBuffer":  0,
			"postSessionBuffer": 30,
		}
		firstTimeslotBody, _ := json.Marshal(firstTimeslotData)

		firstTimeslotReq := httptest.NewRequest("POST", "/api/v1/therapists/"+string(testTherapistID)+"/timeslots", bytes.NewBuffer(firstTimeslotBody))
		firstTimeslotReq.Header.Set("Content-Type", "application/json")
		firstTimeslotRec := httptest.NewRecorder()

		mux.ServeHTTP(firstTimeslotRec, firstTimeslotReq)

		if firstTimeslotRec.Code != http.StatusCreated {
			t.Fatalf("Failed to create first timeslot for overlap test: %d", firstTimeslotRec.Code)
		}

		// Now try to create an overlapping timeslot
		overlappingData := map[string]interface{}{
			"dayOfWeek":         "Saturday",
			"startTime":         "12:00", // Overlaps with first timeslot (10:00-14:00)
			"endTime":           "16:00",
			"preSessionBuffer":  0,
			"postSessionBuffer": 30,
		}
		overlappingBody, _ := json.Marshal(overlappingData)

		overlappingReq := httptest.NewRequest("POST", "/api/v1/therapists/"+string(testTherapistID)+"/timeslots", bytes.NewBuffer(overlappingBody))
		overlappingReq.Header.Set("Content-Type", "application/json")
		overlappingRec := httptest.NewRecorder()

		mux.ServeHTTP(overlappingRec, overlappingReq)

		if overlappingRec.Code != http.StatusConflict {
			t.Errorf("Expected status %d for overlapping timeslot, got %d", http.StatusConflict, overlappingRec.Code)
		}
	})

	t.Run("Time format validation", func(t *testing.T) {
		// Test invalid time format (AM/PM)
		invalidTimeData := map[string]interface{}{
			"dayOfWeek":         "Sunday",
			"startTime":         "9:00 AM", // Invalid: AM/PM format
			"endTime":           "5:00 PM",
			"preSessionBuffer":  0,
			"postSessionBuffer": 30,
		}
		invalidTimeBody, _ := json.Marshal(invalidTimeData)

		invalidTimeReq := httptest.NewRequest("POST", "/api/v1/therapists/"+string(testTherapistID)+"/timeslots", bytes.NewBuffer(invalidTimeBody))
		invalidTimeReq.Header.Set("Content-Type", "application/json")
		invalidTimeRec := httptest.NewRecorder()

		mux.ServeHTTP(invalidTimeRec, invalidTimeReq)

		if invalidTimeRec.Code != http.StatusBadRequest {
			t.Errorf("Expected status %d for AM/PM time format, got %d", http.StatusBadRequest, invalidTimeRec.Code)
		}

		// Test single digit hour
		singleDigitData := map[string]interface{}{
			"dayOfWeek":         "Sunday",
			"startTime":         "9:00", // Invalid: single digit hour
			"endTime":           "17:00",
			"preSessionBuffer":  0,
			"postSessionBuffer": 30,
		}
		singleDigitBody, _ := json.Marshal(singleDigitData)

		singleDigitReq := httptest.NewRequest("POST", "/api/v1/therapists/"+string(testTherapistID)+"/timeslots", bytes.NewBuffer(singleDigitBody))
		singleDigitReq.Header.Set("Content-Type", "application/json")
		singleDigitRec := httptest.NewRecorder()

		mux.ServeHTTP(singleDigitRec, singleDigitReq)

		if singleDigitRec.Code != http.StatusBadRequest {
			t.Errorf("Expected status %d for single digit hour, got %d", http.StatusBadRequest, singleDigitRec.Code)
		}

		// Test invalid time range (start after end)
		invalidRangeData := map[string]interface{}{
			"dayOfWeek":         "Sunday",
			"startTime":         "17:00",
			"endTime":           "09:00", // Invalid: end before start
			"preSessionBuffer":  0,
			"postSessionBuffer": 30,
		}
		invalidRangeBody, _ := json.Marshal(invalidRangeData)

		invalidRangeReq := httptest.NewRequest("POST", "/api/v1/therapists/"+string(testTherapistID)+"/timeslots", bytes.NewBuffer(invalidRangeBody))
		invalidRangeReq.Header.Set("Content-Type", "application/json")
		invalidRangeRec := httptest.NewRecorder()

		mux.ServeHTTP(invalidRangeRec, invalidRangeReq)

		if invalidRangeRec.Code != http.StatusBadRequest {
			t.Errorf("Expected status %d for invalid time range, got %d", http.StatusBadRequest, invalidRangeRec.Code)
		}
	})

	t.Run("Error cases", func(t *testing.T) {
		// Test with non-existent therapist
		nonExistentTherapistID := "therapist_00000000-0000-0000-0000-000000000000"

		timeslotData := map[string]interface{}{
			"dayOfWeek":         "Monday",
			"startTime":         "09:00",
			"endTime":           "17:00",
			"preSessionBuffer":  0,
			"postSessionBuffer": 30,
		}
		timeslotBody, _ := json.Marshal(timeslotData)

		nonExistentReq := httptest.NewRequest("POST", "/api/v1/therapists/"+nonExistentTherapistID+"/timeslots", bytes.NewBuffer(timeslotBody))
		nonExistentReq.Header.Set("Content-Type", "application/json")
		nonExistentRec := httptest.NewRecorder()

		mux.ServeHTTP(nonExistentRec, nonExistentReq)

		if nonExistentRec.Code != http.StatusNotFound {
			t.Errorf("Expected status %d for non-existent therapist, got %d", http.StatusNotFound, nonExistentRec.Code)
		}

		// Test get non-existent timeslot
		nonExistentTimeslotID := "timeslot_00000000-0000-0000-0000-000000000000"
		getReq := httptest.NewRequest("GET", "/api/v1/therapists/"+string(testTherapistID)+"/timeslots/"+nonExistentTimeslotID, nil)
		getRec := httptest.NewRecorder()

		mux.ServeHTTP(getRec, getReq)

		if getRec.Code != http.StatusNotFound {
			t.Errorf("Expected status %d for non-existent timeslot, got %d", http.StatusNotFound, getRec.Code)
		}

		// Test missing required fields
		missingFieldsData := map[string]interface{}{
			"dayOfWeek": "Monday",
			// Missing startTime, endTime, buffers
		}
		missingFieldsBody, _ := json.Marshal(missingFieldsData)

		missingFieldsReq := httptest.NewRequest("POST", "/api/v1/therapists/"+string(testTherapistID)+"/timeslots", bytes.NewBuffer(missingFieldsBody))
		missingFieldsReq.Header.Set("Content-Type", "application/json")
		missingFieldsRec := httptest.NewRecorder()

		mux.ServeHTTP(missingFieldsRec, missingFieldsReq)

		if missingFieldsRec.Code != http.StatusBadRequest {
			t.Errorf("Expected status %d for missing required fields, got %d", http.StatusBadRequest, missingFieldsRec.Code)
		}

		// Test invalid day of week
		invalidDayData := map[string]interface{}{
			"dayOfWeek":         "InvalidDay",
			"startTime":         "09:00",
			"endTime":           "17:00",
			"preSessionBuffer":  0,
			"postSessionBuffer": 30,
		}
		invalidDayBody, _ := json.Marshal(invalidDayData)

		invalidDayReq := httptest.NewRequest("POST", "/api/v1/therapists/"+string(testTherapistID)+"/timeslots", bytes.NewBuffer(invalidDayBody))
		invalidDayReq.Header.Set("Content-Type", "application/json")
		invalidDayRec := httptest.NewRecorder()

		mux.ServeHTTP(invalidDayRec, invalidDayReq)

		if invalidDayRec.Code != http.StatusBadRequest {
			t.Errorf("Expected status %d for invalid day of week, got %d", http.StatusBadRequest, invalidDayRec.Code)
		}

		// Test invalid JSON
		invalidJSONReq := httptest.NewRequest("POST", "/api/v1/therapists/"+string(testTherapistID)+"/timeslots", bytes.NewBufferString("{invalid json"))
		invalidJSONReq.Header.Set("Content-Type", "application/json")
		invalidJSONRec := httptest.NewRecorder()

		mux.ServeHTTP(invalidJSONRec, invalidJSONReq)

		if invalidJSONRec.Code != http.StatusBadRequest {
			t.Errorf("Expected status %d for invalid JSON, got %d", http.StatusBadRequest, invalidJSONRec.Code)
		}
	})
}

func insertTestTherapist(t *testing.T, database ports.SQLDatabase) domain.TherapistID {
	now := time.Now().UTC()
	therapistID := domain.NewTherapistID()

	_, err := database.Exec(`
		INSERT INTO therapists (id, name, email, phone_number, whatsapp_number, speaks_english, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`, therapistID, "Dr. Test Therapist", "test@example.com", "+1234567890", "+1234567890", true, now, now)

	if err != nil {
		t.Fatalf("Failed to insert test therapist: %v", err)
	}

	return therapistID
}

func setupTimeslotTestDB(t *testing.T) (ports.SQLDatabase, func()) {
	// Create temporary database file
	tmpfile, err := os.CreateTemp("", "timeslot_test_*.db")
	if err != nil {
		t.Fatal(err)
	}
	tmpfile.Close()
	dbFilename := tmpfile.Name()

	database := db.NewDatabase(db.DatabaseConfig{
		DBFilename: dbFilename,
		SchemaFile: "../../../schema.sql",
	})

	// Return cleanup function
	cleanup := func() {
		database.Close()
		os.Remove(dbFilename)
	}

	return database, cleanup
}
