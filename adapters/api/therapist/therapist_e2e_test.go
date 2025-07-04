package therapist_handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	specialization_handler "github.com/mishkahtherapy/brain/adapters/api/specialization"
	"github.com/mishkahtherapy/brain/adapters/db"
	"github.com/mishkahtherapy/brain/adapters/db/specialization_db"
	"github.com/mishkahtherapy/brain/adapters/db/therapist_db"
	"github.com/mishkahtherapy/brain/core/domain"
	"github.com/mishkahtherapy/brain/core/domain/specialization"
	"github.com/mishkahtherapy/brain/core/domain/therapist"
	"github.com/mishkahtherapy/brain/core/ports"
	"github.com/mishkahtherapy/brain/core/usecases/specialization/get_all_specializations"
	"github.com/mishkahtherapy/brain/core/usecases/specialization/get_specialization"
	"github.com/mishkahtherapy/brain/core/usecases/specialization/new_specialization"
	"github.com/mishkahtherapy/brain/core/usecases/therapist/get_all_therapists"
	"github.com/mishkahtherapy/brain/core/usecases/therapist/get_therapist"
	"github.com/mishkahtherapy/brain/core/usecases/therapist/new_therapist"
	"github.com/mishkahtherapy/brain/core/usecases/therapist/update_therapist_info"
	"github.com/mishkahtherapy/brain/core/usecases/therapist/update_therapist_specializations"

	_ "github.com/glebarez/go-sqlite"
)

func TestTherapistE2E(t *testing.T) {
	// Setup test database
	db, cleanup := setupTherapistTestDB(t)
	defer cleanup()

	// Setup repositories
	specializationRepo := specialization_db.NewSpecializationRepository(db)
	therapistRepo := therapist_db.NewTherapistRepository(db)

	// Setup specialization usecases (needed for therapist specialization management)
	newSpecializationUsecase := new_specialization.NewUsecase(specializationRepo)
	getAllSpecializationsUsecase := get_all_specializations.NewUsecase(specializationRepo)
	getSpecializationUsecase := get_specialization.NewUsecase(specializationRepo)

	// Setup therapist usecases
	newTherapistUsecase := new_therapist.NewUsecase(therapistRepo, specializationRepo)
	getAllTherapistsUsecase := get_all_therapists.NewUsecase(therapistRepo)
	getTherapistUsecase := get_therapist.NewUsecase(therapistRepo)
	updateTherapistInfoUsecase := update_therapist_info.NewUsecase(therapistRepo)
	updateTherapistSpecializationsUsecase := update_therapist_specializations.NewUsecase(therapistRepo, specializationRepo)

	// Setup handlers
	specializationHandler := specialization_handler.NewSpecializationHandler(*newSpecializationUsecase, *getAllSpecializationsUsecase, *getSpecializationUsecase)
	therapistHandler := NewTherapistHandler(*newTherapistUsecase, *getAllTherapistsUsecase, *getTherapistUsecase, *updateTherapistInfoUsecase, *updateTherapistSpecializationsUsecase)

	// Setup router
	mux := http.NewServeMux()
	specializationHandler.RegisterRoutes(mux)
	therapistHandler.RegisterRoutes(mux)

	t.Run("Complete therapist workflow", func(t *testing.T) {
		// Step 1: Create specializations first (needed for therapist)
		anxietySpecialization := createTestSpecialization(t, mux, "Anxiety Treatment")
		depressionSpecialization := createTestSpecialization(t, mux, "Depression Therapy")

		// Step 2: Create a new therapist
		therapistData := map[string]interface{}{
			"name":              "Dr. Sarah Johnson",
			"email":             "sarah.johnson@therapy.com",
			"phoneNumber":       "+1555001234",
			"whatsAppNumber":    "+1234567890",
			"specializationIds": []string{string(anxietySpecialization.ID)},
		}
		therapistBody, _ := json.Marshal(therapistData)

		createReq := httptest.NewRequest("POST", "/api/v1/therapists", bytes.NewBuffer(therapistBody))
		createReq.Header.Set("Content-Type", "application/json")
		createRec := httptest.NewRecorder()

		mux.ServeHTTP(createRec, createReq)

		// Verify creation response
		if createRec.Code != http.StatusCreated {
			t.Fatalf("Expected status %d, got %d. Body: %s", http.StatusCreated, createRec.Code, createRec.Body.String())
		}

		// Parse created therapist
		var createdTherapist therapist.Therapist
		if err := json.Unmarshal(createRec.Body.Bytes(), &createdTherapist); err != nil {
			t.Fatalf("Failed to parse created therapist: %v", err)
		}

		// Verify created therapist data
		if createdTherapist.Name != "Dr. Sarah Johnson" {
			t.Errorf("Expected name %s, got %s", "Dr. Sarah Johnson", createdTherapist.Name)
		}
		if createdTherapist.Email != "sarah.johnson@therapy.com" {
			t.Errorf("Expected email %s, got %s", "sarah.johnson@therapy.com", createdTherapist.Email)
		}
		if createdTherapist.PhoneNumber != "+1555001234" {
			t.Errorf("Expected phone number %s, got %s", "+1555001234", createdTherapist.PhoneNumber)
		}
		if createdTherapist.WhatsAppNumber != "+1234567890" {
			t.Errorf("Expected WhatsApp number %s, got %s", "+1234567890", createdTherapist.WhatsAppNumber)
		}
		if len(createdTherapist.Specializations) != 1 || createdTherapist.Specializations[0].ID != anxietySpecialization.ID {
			t.Errorf("Expected specialization IDs [%s], got %v", anxietySpecialization.ID, createdTherapist.Specializations)
		}
		if createdTherapist.ID == "" {
			t.Error("Expected ID to be set")
		}
		if createdTherapist.CreatedAt == (domain.UTCTimestamp{}) {
			t.Error("Expected CreatedAt to be set")
		}
		if createdTherapist.UpdatedAt == (domain.UTCTimestamp{}) {
			t.Error("Expected UpdatedAt to be set")
		}

		// Step 3: Get the therapist by ID
		getReq := httptest.NewRequest("GET", "/api/v1/therapists/"+string(createdTherapist.ID), nil)
		getRec := httptest.NewRecorder()

		mux.ServeHTTP(getRec, getReq)

		// Verify get response
		if getRec.Code != http.StatusOK {
			t.Fatalf("Expected status %d, got %d. Body: %s", http.StatusOK, getRec.Code, getRec.Body.String())
		}

		// Parse retrieved therapist
		var retrievedTherapist therapist.Therapist
		if err := json.Unmarshal(getRec.Body.Bytes(), &retrievedTherapist); err != nil {
			t.Fatalf("Failed to parse retrieved therapist: %v", err)
		}

		// Verify retrieved therapist matches created one
		if retrievedTherapist.ID != createdTherapist.ID {
			t.Errorf("Expected ID %s, got %s", createdTherapist.ID, retrievedTherapist.ID)
		}
		if retrievedTherapist.Name != createdTherapist.Name {
			t.Errorf("Expected name %s, got %s", createdTherapist.Name, retrievedTherapist.Name)
		}
		if retrievedTherapist.Email != createdTherapist.Email {
			t.Errorf("Expected email %s, got %s", createdTherapist.Email, retrievedTherapist.Email)
		}
		if len(retrievedTherapist.Specializations) != 1 || retrievedTherapist.Specializations[0].ID != anxietySpecialization.ID {
			t.Errorf("Expected specialization IDs [%s], got %v", anxietySpecialization.ID, retrievedTherapist.Specializations)
		}

		// Step 4: Update therapist specializations
		updateSpecsData := map[string]interface{}{
			"specializationIds": []string{string(anxietySpecialization.ID), string(depressionSpecialization.ID)},
		}
		updateSpecsBody, _ := json.Marshal(updateSpecsData)

		updateReq := httptest.NewRequest("PUT", "/api/v1/therapists/"+string(createdTherapist.ID)+"/specializations", bytes.NewBuffer(updateSpecsBody))
		updateReq.Header.Set("Content-Type", "application/json")
		updateRec := httptest.NewRecorder()

		mux.ServeHTTP(updateRec, updateReq)

		// Verify update response
		if updateRec.Code != http.StatusOK {
			t.Fatalf("Expected status %d, got %d. Body: %s", http.StatusOK, updateRec.Code, updateRec.Body.String())
		}

		// Parse updated therapist
		var updatedTherapist therapist.Therapist
		if err := json.Unmarshal(updateRec.Body.Bytes(), &updatedTherapist); err != nil {
			t.Fatalf("Failed to parse updated therapist: %v", err)
		}

		// Verify updated specializations
		if len(updatedTherapist.Specializations) != 2 {
			t.Errorf("Expected 2 specializations, got %d", len(updatedTherapist.Specializations))
		}
		expectedSpecs := map[domain.SpecializationID]bool{
			anxietySpecialization.ID:    true,
			depressionSpecialization.ID: true,
		}
		for _, specialization := range updatedTherapist.Specializations {
			if !expectedSpecs[specialization.ID] {
				t.Errorf("Unexpected specialization ID: %s", specialization.ID)
			}
		}

		// Step 5: Get therapist again to verify specializations persisted
		getAgainReq := httptest.NewRequest("GET", "/api/v1/therapists/"+string(createdTherapist.ID), nil)
		getAgainRec := httptest.NewRecorder()

		mux.ServeHTTP(getAgainRec, getAgainReq)

		// Verify get response
		if getAgainRec.Code != http.StatusOK {
			t.Fatalf("Expected status %d, got %d. Body: %s", http.StatusOK, getAgainRec.Code, getAgainRec.Body.String())
		}

		// Parse final therapist
		var finalTherapist therapist.Therapist
		if err := json.Unmarshal(getAgainRec.Body.Bytes(), &finalTherapist); err != nil {
			t.Fatalf("Failed to parse final therapist: %v", err)
		}

		// Verify final specializations match updated ones
		if len(finalTherapist.Specializations) != 2 {
			t.Errorf("Expected 2 specializations after update, got %d", len(finalTherapist.Specializations))
		}
		for _, spec := range finalTherapist.Specializations {
			if !expectedSpecs[spec.ID] {
				t.Errorf("Unexpected specialization ID in final therapist: %s", spec.ID)
			}
		}

		// Step 6: Test get all therapists
		getAllReq := httptest.NewRequest("GET", "/api/v1/therapists", nil)
		getAllRec := httptest.NewRecorder()

		mux.ServeHTTP(getAllRec, getAllReq)

		// Verify get all response
		if getAllRec.Code != http.StatusOK {
			t.Fatalf("Expected status %d, got %d. Body: %s", http.StatusOK, getAllRec.Code, getAllRec.Body.String())
		}

		// Parse all therapists
		var allTherapists []*therapist.Therapist
		if err := json.Unmarshal(getAllRec.Body.Bytes(), &allTherapists); err != nil {
			t.Fatalf("Failed to parse all therapists: %v", err)
		}

		// Verify our therapist is in the list
		found := false
		for _, therapist := range allTherapists {
			if therapist.ID == createdTherapist.ID {
				found = true
				break
			}
		}
		if !found {
			t.Error("Created therapist not found in list of all therapists")
		}
	})

	t.Run("Update therapist info", func(t *testing.T) {
		// Create a test therapist first
		anxietySpecialization := createTestSpecialization(t, mux, "Update Test Specialization")
		therapistData := map[string]interface{}{
			"name":              "Dr. Original Name",
			"email":             "original@therapy.com",
			"phoneNumber":       "+1111111111",
			"whatsAppNumber":    "+2222222222",
			"speaksEnglish":     false,
			"specializationIds": []string{string(anxietySpecialization.ID)},
		}
		therapistBody, _ := json.Marshal(therapistData)

		createReq := httptest.NewRequest("POST", "/api/v1/therapists", bytes.NewBuffer(therapistBody))
		createReq.Header.Set("Content-Type", "application/json")
		createRec := httptest.NewRecorder()
		mux.ServeHTTP(createRec, createReq)

		if createRec.Code != http.StatusCreated {
			t.Fatalf("Failed to create test therapist: %d, %s", createRec.Code, createRec.Body.String())
		}

		var createdTherapist therapist.Therapist
		json.Unmarshal(createRec.Body.Bytes(), &createdTherapist)

		// Test successful update
		t.Run("successful update", func(t *testing.T) {
			updateData := map[string]interface{}{
				"name":           "Dr. Updated Name",
				"email":          "updated@therapy.com",
				"phoneNumber":    "+3333333333",
				"whatsAppNumber": "+4444444444",
				"speaksEnglish":  true,
			}
			updateBody, _ := json.Marshal(updateData)

			updateReq := httptest.NewRequest("PUT", "/api/v1/therapists/"+string(createdTherapist.ID), bytes.NewBuffer(updateBody))
			updateReq.Header.Set("Content-Type", "application/json")
			updateRec := httptest.NewRecorder()
			mux.ServeHTTP(updateRec, updateReq)

			if updateRec.Code != http.StatusOK {
				t.Fatalf("Expected status %d, got %d. Body: %s", http.StatusOK, updateRec.Code, updateRec.Body.String())
			}

			var updatedTherapist therapist.Therapist
			json.Unmarshal(updateRec.Body.Bytes(), &updatedTherapist)

			// Verify all fields were updated
			if updatedTherapist.Name != "Dr. Updated Name" {
				t.Errorf("Expected name %s, got %s", "Dr. Updated Name", updatedTherapist.Name)
			}
			if updatedTherapist.Email != "updated@therapy.com" {
				t.Errorf("Expected email %s, got %s", "updated@therapy.com", updatedTherapist.Email)
			}
			if updatedTherapist.PhoneNumber != "+3333333333" {
				t.Errorf("Expected phone %s, got %s", "+3333333333", updatedTherapist.PhoneNumber)
			}
			if updatedTherapist.WhatsAppNumber != "+4444444444" {
				t.Errorf("Expected WhatsApp %s, got %s", "+4444444444", updatedTherapist.WhatsAppNumber)
			}
			if !updatedTherapist.SpeaksEnglish {
				t.Error("Expected SpeaksEnglish to be true")
			}

			// Verify immutable fields weren't changed
			if updatedTherapist.ID != createdTherapist.ID {
				t.Error("ID should not change")
			}
			if updatedTherapist.CreatedAt != createdTherapist.CreatedAt {
				t.Error("CreatedAt should not change")
			}
			if len(updatedTherapist.Specializations) != 1 || updatedTherapist.Specializations[0].ID != anxietySpecialization.ID {
				t.Error("Specializations should be preserved")
			}
		})

		// Test validation errors
		t.Run("validation errors", func(t *testing.T) {
			testCases := []struct {
				name         string
				updateData   map[string]interface{}
				expectedCode int
			}{
				{
					name: "missing name",
					updateData: map[string]interface{}{
						"email":          "test@therapy.com",
						"phoneNumber":    "+1111111111",
						"whatsAppNumber": "+2222222222",
						"speaksEnglish":  true,
					},
					expectedCode: http.StatusBadRequest,
				},
				{
					name: "missing email",
					updateData: map[string]interface{}{
						"name":           "Dr. Test",
						"phoneNumber":    "+1111111111",
						"whatsAppNumber": "+2222222222",
						"speaksEnglish":  true,
					},
					expectedCode: http.StatusBadRequest,
				},
				{
					name: "invalid phone number",
					updateData: map[string]interface{}{
						"name":           "Dr. Test",
						"email":          "test@therapy.com",
						"phoneNumber":    "invalid-phone",
						"whatsAppNumber": "+2222222222",
						"speaksEnglish":  true,
					},
					expectedCode: http.StatusBadRequest,
				},
				{
					name: "invalid WhatsApp number",
					updateData: map[string]interface{}{
						"name":           "Dr. Test",
						"email":          "test@therapy.com",
						"phoneNumber":    "+1111111111",
						"whatsAppNumber": "invalid-whatsapp",
						"speaksEnglish":  true,
					},
					expectedCode: http.StatusBadRequest,
				},
			}

			for _, tc := range testCases {
				t.Run(tc.name, func(t *testing.T) {
					updateBody, _ := json.Marshal(tc.updateData)
					updateReq := httptest.NewRequest("PUT", "/api/v1/therapists/"+string(createdTherapist.ID), bytes.NewBuffer(updateBody))
					updateReq.Header.Set("Content-Type", "application/json")
					updateRec := httptest.NewRecorder()
					mux.ServeHTTP(updateRec, updateReq)

					if updateRec.Code != tc.expectedCode {
						t.Errorf("Expected status %d, got %d. Body: %s", tc.expectedCode, updateRec.Code, updateRec.Body.String())
					}
				})
			}
		})

		// Test conflict scenarios
		t.Run("conflict scenarios", func(t *testing.T) {
			// Create another therapist to test conflicts
			otherTherapistData := map[string]interface{}{
				"name":              "Dr. Other",
				"email":             "other@therapy.com",
				"phoneNumber":       "+5555555555",
				"whatsAppNumber":    "+6666666666",
				"speaksEnglish":     true,
				"specializationIds": []string{string(anxietySpecialization.ID)},
			}
			otherBody, _ := json.Marshal(otherTherapistData)

			otherReq := httptest.NewRequest("POST", "/api/v1/therapists", bytes.NewBuffer(otherBody))
			otherReq.Header.Set("Content-Type", "application/json")
			otherRec := httptest.NewRecorder()
			mux.ServeHTTP(otherRec, otherReq)

			if otherRec.Code != http.StatusCreated {
				t.Fatalf("Failed to create other therapist: %d", otherRec.Code)
			}

			// Test email conflict
			t.Run("email already exists", func(t *testing.T) {
				updateData := map[string]interface{}{
					"name":           "Dr. Updated Name",
					"email":          "other@therapy.com", // This email already exists
					"phoneNumber":    "+3333333333",
					"whatsAppNumber": "+4444444444",
					"speaksEnglish":  true,
				}
				updateBody, _ := json.Marshal(updateData)

				updateReq := httptest.NewRequest("PUT", "/api/v1/therapists/"+string(createdTherapist.ID), bytes.NewBuffer(updateBody))
				updateReq.Header.Set("Content-Type", "application/json")
				updateRec := httptest.NewRecorder()
				mux.ServeHTTP(updateRec, updateReq)

				if updateRec.Code != http.StatusConflict {
					t.Errorf("Expected status %d for email conflict, got %d. Body: %s", http.StatusConflict, updateRec.Code, updateRec.Body.String())
				}
			})

			// Test WhatsApp conflict
			t.Run("whatsApp already exists", func(t *testing.T) {
				updateData := map[string]interface{}{
					"name":           "Dr. Updated Name",
					"email":          "updated2@therapy.com",
					"phoneNumber":    "+3333333333",
					"whatsAppNumber": "+6666666666", // This WhatsApp already exists
					"speaksEnglish":  true,
				}
				updateBody, _ := json.Marshal(updateData)

				updateReq := httptest.NewRequest("PUT", "/api/v1/therapists/"+string(createdTherapist.ID), bytes.NewBuffer(updateBody))
				updateReq.Header.Set("Content-Type", "application/json")
				updateRec := httptest.NewRecorder()
				mux.ServeHTTP(updateRec, updateReq)

				if updateRec.Code != http.StatusConflict {
					t.Errorf("Expected status %d for WhatsApp conflict, got %d. Body: %s", http.StatusConflict, updateRec.Code, updateRec.Body.String())
				}
			})
		})

		// Test non-existent therapist
		t.Run("therapist not found", func(t *testing.T) {
			updateData := map[string]interface{}{
				"name":           "Dr. Test",
				"email":          "test@therapy.com",
				"phoneNumber":    "+1111111111",
				"whatsAppNumber": "+2222222222",
				"speaksEnglish":  true,
			}
			updateBody, _ := json.Marshal(updateData)

			updateReq := httptest.NewRequest("PUT", "/api/v1/therapists/nonexistent", bytes.NewBuffer(updateBody))
			updateReq.Header.Set("Content-Type", "application/json")
			updateRec := httptest.NewRecorder()
			mux.ServeHTTP(updateRec, updateReq)

			if updateRec.Code != http.StatusNotFound {
				t.Errorf("Expected status %d for non-existent therapist, got %d. Body: %s", http.StatusNotFound, updateRec.Code, updateRec.Body.String())
			}
		})
	})

	t.Run("Error cases", func(t *testing.T) {
		// Test get non-existent therapist
		nonExistentID := "therapist_00000000-0000-0000-0000-000000000000"
		getReq := httptest.NewRequest("GET", "/api/v1/therapists/"+nonExistentID, nil)
		getRec := httptest.NewRecorder()

		mux.ServeHTTP(getRec, getReq)

		if getRec.Code != http.StatusNotFound {
			t.Errorf("Expected status %d for non-existent therapist, got %d", http.StatusNotFound, getRec.Code)
		}

		// Test create therapist with invalid data (missing email)
		invalidTherapistData := map[string]interface{}{
			"name":           "Dr. Invalid",
			"phoneNumber":    "+1555001234",
			"whatsAppNumber": "+1234567890",
		}
		invalidBody, _ := json.Marshal(invalidTherapistData)

		createReq := httptest.NewRequest("POST", "/api/v1/therapists", bytes.NewBuffer(invalidBody))
		createReq.Header.Set("Content-Type", "application/json")
		createRec := httptest.NewRecorder()

		mux.ServeHTTP(createRec, createReq)

		if createRec.Code != http.StatusBadRequest {
			t.Errorf("Expected status %d for invalid therapist data, got %d", http.StatusBadRequest, createRec.Code)
		}

		// Test update specializations for non-existent therapist
		updateData := map[string]interface{}{
			"specializationIds": []string{},
		}
		updateBody, _ := json.Marshal(updateData)

		updateReq := httptest.NewRequest("PUT", "/api/v1/therapists/"+nonExistentID+"/specializations", bytes.NewBuffer(updateBody))
		updateReq.Header.Set("Content-Type", "application/json")
		updateRec := httptest.NewRecorder()

		mux.ServeHTTP(updateRec, updateReq)

		if updateRec.Code != http.StatusNotFound {
			t.Errorf("Expected status %d for updating non-existent therapist, got %d", http.StatusNotFound, updateRec.Code)
		}
	})
}

// Helper function to create test specializations
func createTestSpecialization(t *testing.T, mux *http.ServeMux, name string) *specialization.Specialization {
	createPayload := map[string]string{
		"name": name,
	}
	createBody, _ := json.Marshal(createPayload)

	createReq := httptest.NewRequest("POST", "/api/v1/specializations", bytes.NewBuffer(createBody))
	createReq.Header.Set("Content-Type", "application/json")
	createRec := httptest.NewRecorder()

	mux.ServeHTTP(createRec, createReq)

	if createRec.Code != http.StatusCreated {
		t.Fatalf("Failed to create test specialization %s: status %d, body: %s", name, createRec.Code, createRec.Body.String())
	}

	var spec specialization.Specialization
	if err := json.Unmarshal(createRec.Body.Bytes(), &spec); err != nil {
		t.Fatalf("Failed to parse created specialization %s: %v", name, err)
	}

	return &spec
}

// Setup test database with all required tables for therapist testing
func setupTherapistTestDB(_ *testing.T) (ports.SQLDatabase, func()) {
	// Create temporary database file
	dbFilename := "therapist_test.db" // Use in-memory database for testing
	// Remove if exists
	if _, err := os.Stat(dbFilename); err == nil {
		os.Remove(dbFilename)
	}
	// Return cleanup function
	database := db.NewDatabase(db.DatabaseConfig{
		DBFilename: dbFilename,
		SchemaFile: "../../../schema.sql",
	})
	cleanup := func() {
		database.Close()
		os.Remove(dbFilename)
	}

	return database, cleanup
}
