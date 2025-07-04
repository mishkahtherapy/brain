package timeslot_handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/mishkahtherapy/brain/adapters/db"
	"github.com/mishkahtherapy/brain/adapters/db/therapist_db"
	"github.com/mishkahtherapy/brain/adapters/db/timeslot_db"
	"github.com/mishkahtherapy/brain/core/domain"
	"github.com/mishkahtherapy/brain/core/ports"
	"github.com/mishkahtherapy/brain/core/usecases/timeslot/bulk_toggle_therapist_timeslots"
	"github.com/mishkahtherapy/brain/core/usecases/timeslot/create_therapist_timeslot"
	"github.com/mishkahtherapy/brain/core/usecases/timeslot/delete_therapist_timeslot"
	"github.com/mishkahtherapy/brain/core/usecases/timeslot/get_therapist_timeslot"
	"github.com/mishkahtherapy/brain/core/usecases/timeslot/list_therapist_timeslots"
	"github.com/mishkahtherapy/brain/core/usecases/timeslot/update_therapist_timeslot"
)

func TestBulkToggleTimeslots(t *testing.T) {
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
	bulkToggleUsecase := bulk_toggle_therapist_timeslots.NewUsecase(therapistRepo, timeslotRepo)

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

	const testTimezoneOffset = 0 // UTC for simplicity

	t.Run("Bulk deactivate all timeslots", func(t *testing.T) {
		// Create multiple active timeslots
		timeslotIDs := createMultipleTimeslots(t, mux, testTherapistID, 3)

		// Verify all timeslots are active initially
		for _, timeslotID := range timeslotIDs {
			timeslot := getTimeslotByID(t, mux, testTherapistID, timeslotID)
			if isActive, ok := timeslot["isActive"].(bool); !ok || !isActive {
				t.Errorf("Expected timeslot %s to be active initially", timeslotID)
			}
		}

		// Bulk deactivate all timeslots
		requestBody := map[string]bool{
			"isActive": false,
		}
		requestBodyJSON, _ := json.Marshal(requestBody)

		req := httptest.NewRequest(
			http.MethodPut,
			fmt.Sprintf("/api/v1/therapists/%s/timeslots/bulk-toggle", testTherapistID),
			bytes.NewBuffer(requestBodyJSON),
		)
		req.Header.Set("Content-Type", "application/json")

		rr := httptest.NewRecorder()
		mux.ServeHTTP(rr, req)

		// Assert response
		if rr.Code != http.StatusOK {
			t.Fatalf("Expected status %d, got %d. Response: %s", http.StatusOK, rr.Code, rr.Body.String())
		}

		var response map[string]string
		if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}

		expectedMessage := "Bulk toggle completed successfully"
		if response["message"] != expectedMessage {
			t.Errorf("Expected message %s, got %s", expectedMessage, response["message"])
		}

		// Verify all timeslots are now inactive
		for _, timeslotID := range timeslotIDs {
			timeslot := getTimeslotByID(t, mux, testTherapistID, timeslotID)
			if isActive, ok := timeslot["isActive"].(bool); !ok || isActive {
				t.Errorf("Expected timeslot %s to be inactive after bulk deactivate", timeslotID)
			}
		}
	})

	t.Run("Bulk activate all timeslots", func(t *testing.T) {
		// Create multiple inactive timeslots by creating active ones and deactivating them
		timeslotIDs := createMultipleTimeslots(t, mux, testTherapistID, 3)

		// First deactivate them all
		deactivateData := map[string]bool{"isActive": false}
		deactivateBody, _ := json.Marshal(deactivateData)
		deactivateReq := httptest.NewRequest(
			http.MethodPut,
			fmt.Sprintf("/api/v1/therapists/%s/timeslots/bulk-toggle", testTherapistID),
			bytes.NewBuffer(deactivateBody),
		)
		deactivateReq.Header.Set("Content-Type", "application/json")
		deactivateRec := httptest.NewRecorder()
		mux.ServeHTTP(deactivateRec, deactivateReq)

		if deactivateRec.Code != http.StatusOK {
			t.Fatalf("Failed to deactivate timeslots: %s", deactivateRec.Body.String())
		}

		// Verify all timeslots are inactive
		for _, timeslotID := range timeslotIDs {
			timeslot := getTimeslotByID(t, mux, testTherapistID, timeslotID)
			if isActive, ok := timeslot["isActive"].(bool); !ok || isActive {
				t.Errorf("Expected timeslot %s to be inactive before bulk activate", timeslotID)
			}
		}

		// Now bulk activate all timeslots
		requestBody := map[string]bool{
			"isActive": true,
		}
		requestBodyJSON, _ := json.Marshal(requestBody)

		req := httptest.NewRequest(
			http.MethodPut,
			fmt.Sprintf("/api/v1/therapists/%s/timeslots/bulk-toggle", testTherapistID),
			bytes.NewBuffer(requestBodyJSON),
		)
		req.Header.Set("Content-Type", "application/json")

		rr := httptest.NewRecorder()
		mux.ServeHTTP(rr, req)

		// Assert response
		if rr.Code != http.StatusOK {
			t.Fatalf("Expected status %d, got %d. Response: %s", http.StatusOK, rr.Code, rr.Body.String())
		}

		var response map[string]string
		if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}

		expectedMessage := "Bulk toggle completed successfully"
		if response["message"] != expectedMessage {
			t.Errorf("Expected message %s, got %s", expectedMessage, response["message"])
		}

		// Verify all timeslots are now active
		for _, timeslotID := range timeslotIDs {
			timeslot := getTimeslotByID(t, mux, testTherapistID, timeslotID)
			if isActive, ok := timeslot["isActive"].(bool); !ok || !isActive {
				t.Errorf("Expected timeslot %s to be active after bulk activate", timeslotID)
			}
		}
	})

	t.Run("Bulk toggle with no timeslots", func(t *testing.T) {
		// Create a new therapist with no timeslots
		newTherapistID := insertTestTherapist(t, database)

		// Try to bulk toggle with no timeslots
		requestBody := map[string]bool{
			"isActive": false,
		}
		requestBodyJSON, _ := json.Marshal(requestBody)

		req := httptest.NewRequest(
			http.MethodPut,
			fmt.Sprintf("/api/v1/therapists/%s/timeslots/bulk-toggle", newTherapistID),
			bytes.NewBuffer(requestBodyJSON),
		)
		req.Header.Set("Content-Type", "application/json")

		rr := httptest.NewRecorder()
		mux.ServeHTTP(rr, req)

		// Should still return success
		if rr.Code != http.StatusOK {
			t.Errorf("Expected status %d, got %d. Response: %s", http.StatusOK, rr.Code, rr.Body.String())
		}

		var response map[string]string
		if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
			t.Errorf("Failed to unmarshal response: %v", err)
		}

		expectedMessage := "Bulk toggle completed successfully"
		if response["message"] != expectedMessage {
			t.Errorf("Expected message %s, got %s", expectedMessage, response["message"])
		}
	})

	t.Run("Bulk toggle with non-existent therapist", func(t *testing.T) {
		nonExistentTherapistID := "non-existent-therapist"

		requestBody := map[string]bool{
			"isActive": false,
		}
		requestBodyJSON, _ := json.Marshal(requestBody)

		req := httptest.NewRequest(
			http.MethodPut,
			fmt.Sprintf("/api/v1/therapists/%s/timeslots/bulk-toggle", nonExistentTherapistID),
			bytes.NewBuffer(requestBodyJSON),
		)
		req.Header.Set("Content-Type", "application/json")

		rr := httptest.NewRecorder()
		mux.ServeHTTP(rr, req)

		// Should return 404
		if rr.Code != http.StatusNotFound {
			t.Errorf("Expected status %d, got %d. Response: %s", http.StatusNotFound, rr.Code, rr.Body.String())
		}
	})

	t.Run("Bulk toggle with invalid JSON", func(t *testing.T) {
		req := httptest.NewRequest(
			http.MethodPut,
			fmt.Sprintf("/api/v1/therapists/%s/timeslots/bulk-toggle", testTherapistID),
			bytes.NewBufferString("{invalid json"),
		)
		req.Header.Set("Content-Type", "application/json")

		rr := httptest.NewRecorder()
		mux.ServeHTTP(rr, req)

		// Should return 400
		if rr.Code != http.StatusBadRequest {
			t.Errorf("Expected status %d, got %d. Response: %s", http.StatusBadRequest, rr.Code, rr.Body.String())
		}
	})

	t.Run("Bulk toggle with missing therapist ID", func(t *testing.T) {
		requestBody := map[string]bool{
			"isActive": false,
		}
		requestBodyJSON, _ := json.Marshal(requestBody)

		req := httptest.NewRequest(
			http.MethodPut,
			"/api/v1/therapists//timeslots/bulk-toggle",
			bytes.NewBuffer(requestBodyJSON),
		)
		req.Header.Set("Content-Type", "application/json")

		rr := httptest.NewRecorder()
		mux.ServeHTTP(rr, req)

		// Should return 400
		if rr.Code != http.StatusBadRequest {
			t.Errorf("Expected status %d, got %d. Response: %s", http.StatusBadRequest, rr.Code, rr.Body.String())
		}
	})
}

// Helper function to create multiple timeslots and return their IDs
func createMultipleTimeslots(t *testing.T, mux *http.ServeMux, therapistID domain.TherapistID, count int) []string {
	timeslotIDs := make([]string, count)
	days := []string{"Monday", "Tuesday", "Wednesday", "Thursday", "Friday"}

	for i := 0; i < count; i++ {
		timeslotData := map[string]interface{}{
			"therapistId":       string(therapistID),
			"dayOfWeek":         days[i%len(days)],
			"startTime":         fmt.Sprintf("%02d:00", 9+i),
			"durationMinutes":   60,
			"timezoneOffset":    0, // UTC
			"preSessionBuffer":  5,
			"postSessionBuffer": 5,
		}
		timeslotBody, _ := json.Marshal(timeslotData)

		createReq := httptest.NewRequest(
			"POST",
			fmt.Sprintf("/api/v1/therapists/%s/timeslots", therapistID),
			bytes.NewBuffer(timeslotBody),
		)
		createReq.Header.Set("Content-Type", "application/json")
		createRec := httptest.NewRecorder()

		mux.ServeHTTP(createRec, createReq)

		if createRec.Code != http.StatusCreated {
			t.Fatalf("Failed to create timeslot %d: %s", i, createRec.Body.String())
		}

		var response map[string]interface{}
		if err := json.Unmarshal(createRec.Body.Bytes(), &response); err != nil {
			t.Fatalf("Failed to unmarshal create response: %v", err)
		}

		timeslotID, ok := response["id"].(string)
		if !ok {
			t.Fatalf("Expected ID to be string, got %T", response["id"])
		}

		timeslotIDs[i] = timeslotID
	}

	return timeslotIDs
}

// Helper function to get a timeslot by ID
func getTimeslotByID(t *testing.T, mux *http.ServeMux, therapistID domain.TherapistID, timeslotID string) map[string]interface{} {
	getReq := httptest.NewRequest(
		"GET",
		fmt.Sprintf("/api/v1/therapists/%s/timeslots/%s?timezoneOffset=0", therapistID, timeslotID),
		nil,
	)
	getRec := httptest.NewRecorder()

	mux.ServeHTTP(getRec, getReq)

	if getRec.Code != http.StatusOK {
		t.Fatalf("Failed to get timeslot %s: %s", timeslotID, getRec.Body.String())
	}

	var timeslot map[string]interface{}
	if err := json.Unmarshal(getRec.Body.Bytes(), &timeslot); err != nil {
		t.Fatalf("Failed to unmarshal timeslot response: %v", err)
	}

	return timeslot
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
