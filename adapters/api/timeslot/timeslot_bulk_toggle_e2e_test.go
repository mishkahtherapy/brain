package timeslot_handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mishkahtherapy/brain/adapters/api/internal/testutils"
	"github.com/mishkahtherapy/brain/core/domain"
	"github.com/mishkahtherapy/brain/core/usecases/timeslot/bulk_toggle_therapist_timeslots"
	"github.com/mishkahtherapy/brain/core/usecases/timeslot/create_therapist_timeslot"
	"github.com/mishkahtherapy/brain/core/usecases/timeslot/delete_therapist_timeslot"
	"github.com/mishkahtherapy/brain/core/usecases/timeslot/get_therapist_timeslot"
	"github.com/mishkahtherapy/brain/core/usecases/timeslot/list_therapist_timeslots"
	"github.com/mishkahtherapy/brain/core/usecases/timeslot/update_therapist_timeslot"
)

func TestBulkToggleTimeslots(t *testing.T) {
	// Setup test database using utilities
	database, cleanup := testutils.SetupTestDB(t)
	defer cleanup()

	// Insert test therapist using utilities
	testTherapistID := testutils.CreateTestTherapist(t, database)

	// Setup repositories using utilities
	repos := testutils.SetupRepositories(database)

	// Setup usecases
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

		// Assert response using utilities
		testutils.AssertStatus(t, rr, http.StatusOK)

		var response map[string]interface{}
		testutils.AssertJSONResponse(t, rr, http.StatusOK, &response)

		expectedMessage := "Bulk toggle completed successfully"
		testutils.AssertStringField(t, response, "message", expectedMessage)

		// Verify all timeslots are now inactive
		for _, timeslotID := range timeslotIDs {
			timeslot := getTimeslotByID(t, mux, testTherapistID, timeslotID)
			if isActive, ok := timeslot["isActive"].(bool); !ok || isActive {
				t.Errorf("Expected timeslot %s to be inactive after bulk deactivate", timeslotID)
			}
		}
	})

	t.Run("Bulk activate all timeslots", func(t *testing.T) {
		// Create a fresh therapist for this test to avoid conflicts
		freshTherapistID := testutils.CreateTestTherapistWithName(t, database, "Dr. Fresh Therapist")

		// Create multiple inactive timeslots by creating active ones and deactivating them
		timeslotIDs := createMultipleTimeslots(t, mux, freshTherapistID, 3)

		// First deactivate them all
		deactivateData := map[string]bool{"isActive": false}
		deactivateBody, _ := json.Marshal(deactivateData)
		deactivateReq := httptest.NewRequest(
			http.MethodPut,
			fmt.Sprintf("/api/v1/therapists/%s/timeslots/bulk-toggle", freshTherapistID),
			bytes.NewBuffer(deactivateBody),
		)
		deactivateReq.Header.Set("Content-Type", "application/json")
		deactivateRec := httptest.NewRecorder()
		mux.ServeHTTP(deactivateRec, deactivateReq)

		testutils.AssertStatus(t, deactivateRec, http.StatusOK)

		// Verify all timeslots are inactive
		for _, timeslotID := range timeslotIDs {
			timeslot := getTimeslotByID(t, mux, freshTherapistID, timeslotID)
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
			fmt.Sprintf("/api/v1/therapists/%s/timeslots/bulk-toggle", freshTherapistID),
			bytes.NewBuffer(requestBodyJSON),
		)
		req.Header.Set("Content-Type", "application/json")

		rr := httptest.NewRecorder()
		mux.ServeHTTP(rr, req)

		// Assert response using utilities
		var response map[string]interface{}
		testutils.AssertJSONResponse(t, rr, http.StatusOK, &response)

		expectedMessage := "Bulk toggle completed successfully"
		testutils.AssertStringField(t, response, "message", expectedMessage)

		// Verify all timeslots are now active
		for _, timeslotID := range timeslotIDs {
			timeslot := getTimeslotByID(t, mux, freshTherapistID, timeslotID)
			if isActive, ok := timeslot["isActive"].(bool); !ok || !isActive {
				t.Errorf("Expected timeslot %s to be active after bulk activate", timeslotID)
			}
		}
	})

	t.Run("Bulk toggle with no timeslots", func(t *testing.T) {
		// Create a new therapist with no timeslots using utilities with unique name
		newTherapistID := testutils.CreateTestTherapistWithName(t, database, "Dr. Empty Schedule")

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
		var response map[string]interface{}
		testutils.AssertJSONResponse(t, rr, http.StatusOK, &response)

		expectedMessage := "Bulk toggle completed successfully"
		testutils.AssertStringField(t, response, "message", expectedMessage)
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

		// Should return 500 (internal server error) when therapist not found
		testutils.AssertError(t, rr, http.StatusInternalServerError)
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
		testutils.AssertError(t, rr, http.StatusBadRequest)
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

		// Empty therapist ID in URL likely returns 301 (redirect) or 404, let's check what we actually get
		if rr.Code != http.StatusMovedPermanently && rr.Code != http.StatusNotFound && rr.Code != http.StatusBadRequest {
			t.Errorf("Expected status 301, 404, or 400 for missing therapist ID, got %d. Response: %s", rr.Code, rr.Body.String())
		}
	})
}

// Helper function to create multiple timeslots and return their IDs
func createMultipleTimeslots(t *testing.T, mux *http.ServeMux, therapistID domain.TherapistID, count int) []string {
	timeslotIDs := make([]string, count)
	days := []string{"Monday", "Tuesday", "Wednesday", "Thursday", "Friday"}

	for i := 0; i < count; i++ {
		timeslotData := map[string]interface{}{
			"therapistId":           string(therapistID),
			"dayOfWeek":             days[i%len(days)],
			"startTime":             fmt.Sprintf("%02d:00", 9+i*2), // Use different times: 09:00, 11:00, 13:00
			"durationMinutes":       60,
			"timezoneOffset":        0, // UTC
			"advanceNotice":         15,
			"afterSessionBreakTime": 30, // Fix: Must be at least 30 minutes
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
