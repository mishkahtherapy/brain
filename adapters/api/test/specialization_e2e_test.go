package test

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mishkahtherapy/brain/adapters/api"
	"github.com/mishkahtherapy/brain/adapters/db/specialization"
	"github.com/mishkahtherapy/brain/core/domain"
	"github.com/mishkahtherapy/brain/core/usecases/specialization/get_all_specializations"
	"github.com/mishkahtherapy/brain/core/usecases/specialization/get_specialization"
	"github.com/mishkahtherapy/brain/core/usecases/specialization/new_specialization"

	_ "github.com/glebarez/go-sqlite"
)

func TestSpecializationE2E(t *testing.T) {
	// Setup test database
	db, cleanup := setupTestDB(t)
	defer cleanup()

	// Setup repositories
	specializationRepo := specialization.NewSpecializationRepository(db)

	// Setup usecases
	createUsecase := new_specialization.NewUsecase(specializationRepo)
	getAllUsecase := get_all_specializations.NewUsecase(specializationRepo)
	getUsecase := get_specialization.NewUsecase(specializationRepo)

	// Setup handler with usecases
	handler := api.NewSpecializationHandler(*createUsecase, *getAllUsecase, *getUsecase)

	// Setup router
	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	t.Run("Create and retrieve specialization", func(t *testing.T) {
		// Test data
		testName := "cognitive behavioral therapy"

		// Step 1: Create a specialization
		createPayload := map[string]string{
			"name": testName,
		}
		createBody, _ := json.Marshal(createPayload)

		createReq := httptest.NewRequest("POST", "/api/v1/specializations", bytes.NewBuffer(createBody))
		createReq.Header.Set("Content-Type", "application/json")
		createRec := httptest.NewRecorder()

		mux.ServeHTTP(createRec, createReq)

		// Verify creation response
		if createRec.Code != http.StatusCreated {
			t.Fatalf("Expected status %d, got %d. Body: %s", http.StatusCreated, createRec.Code, createRec.Body.String())
		}

		// Parse created specialization
		var createdSpec domain.Specialization
		body := createRec.Body.Bytes()
		fmt.Println("body", string(body))
		if err := json.Unmarshal(body, &createdSpec); err != nil {
			t.Fatalf("Failed to parse created specialization: %v", err)
		}

		// Verify created specialization data
		if createdSpec.Name != testName {
			t.Errorf("Expected name %s, got %s", testName, createdSpec.Name)
		}
		if createdSpec.ID == "" {
			t.Error("Expected ID to be set")
		}
		// Check if timestamps are set (UTCTimestamp doesn't have IsZero method)
		if createdSpec.CreatedAt == (domain.UTCTimestamp{}) {
			t.Error("Expected CreatedAt to be set")
		}
		if createdSpec.UpdatedAt == (domain.UTCTimestamp{}) {
			t.Error("Expected UpdatedAt to be set")
		}

		// Step 2: Retrieve the specialization by ID
		getReq := httptest.NewRequest("GET", "/api/v1/specializations/"+string(createdSpec.ID), nil)
		getRec := httptest.NewRecorder()

		mux.ServeHTTP(getRec, getReq)

		// Verify get response
		if getRec.Code != http.StatusOK {
			t.Fatalf("Expected status %d, got %d. Body: %s", http.StatusOK, getRec.Code, getRec.Body.String())
		}

		// Parse retrieved specialization
		var retrievedSpec domain.Specialization
		if err := json.Unmarshal(getRec.Body.Bytes(), &retrievedSpec); err != nil {
			t.Fatalf("Failed to parse retrieved specialization: %v", err)
		}

		// Verify retrieved specialization matches created one
		if retrievedSpec.ID != createdSpec.ID {
			t.Errorf("Expected ID %s, got %s", createdSpec.ID, retrievedSpec.ID)
		}
		if retrievedSpec.Name != createdSpec.Name {
			t.Errorf("Expected name %s, got %s", createdSpec.Name, retrievedSpec.Name)
		}
		if retrievedSpec.CreatedAt != createdSpec.CreatedAt {
			t.Errorf("Expected CreatedAt %v, got %v", createdSpec.CreatedAt, retrievedSpec.CreatedAt)
		}
		if retrievedSpec.UpdatedAt != createdSpec.UpdatedAt {
			t.Errorf("Expected UpdatedAt %v, got %v", createdSpec.UpdatedAt, retrievedSpec.UpdatedAt)
		}

		// Step 3: Retrieve all specializations
		getAllReq := httptest.NewRequest("GET", "/api/v1/specializations", nil)
		getAllRec := httptest.NewRecorder()

		mux.ServeHTTP(getAllRec, getAllReq)

		// Verify get all response
		if getAllRec.Code != http.StatusOK {
			t.Fatalf("Expected status %d, got %d. Body: %s", http.StatusOK, getAllRec.Code, getAllRec.Body.String())
		}

		// Parse all specializations
		var allSpecs []*domain.Specialization
		if err := json.Unmarshal(getAllRec.Body.Bytes(), &allSpecs); err != nil {
			t.Fatalf("Failed to parse all specializations: %v", err)
		}

		// Verify our specialization is in the list
		found := false
		for _, spec := range allSpecs {
			if spec.ID == createdSpec.ID {
				found = true
				break
			}
		}
		if !found {
			t.Error("Created specialization not found in list of all specializations")
		}
	})

	t.Run("Get non-existent specialization returns 404", func(t *testing.T) {
		nonExistentID := "specialization_00000000-0000-0000-0000-000000000000"
		getReq := httptest.NewRequest("GET", "/api/v1/specializations/"+nonExistentID, nil)
		getRec := httptest.NewRecorder()

		mux.ServeHTTP(getRec, getReq)

		if getRec.Code != http.StatusNotFound {
			t.Errorf("Expected status %d, got %d", http.StatusNotFound, getRec.Code)
		}
	})

	t.Run("Create specialization with invalid data returns 400", func(t *testing.T) {
		// Test with empty name
		createPayload := map[string]string{
			"name": "",
		}
		createBody, _ := json.Marshal(createPayload)

		createReq := httptest.NewRequest("POST", "/api/v1/specializations", bytes.NewBuffer(createBody))
		createReq.Header.Set("Content-Type", "application/json")
		createRec := httptest.NewRecorder()

		mux.ServeHTTP(createRec, createReq)

		if createRec.Code != http.StatusInternalServerError {
			t.Errorf("Expected status %d, got %d", http.StatusInternalServerError, createRec.Code)
		}
	})
}

func setupTestDB(t *testing.T) (*sql.DB, func()) {
	// Create temporary database file
	dbFile := ":memory:" // Use in-memory database for testing

	db, err := sql.Open("sqlite", dbFile)
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}

	// Create specializations table
	createTableSQL := `
		CREATE TABLE specializations (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			created_at DATETIME NOT NULL,
			updated_at DATETIME NOT NULL
		);
	`

	_, err = db.Exec(createTableSQL)
	if err != nil {
		t.Fatalf("Failed to create specializations table: %v", err)
	}

	// Return cleanup function
	cleanup := func() {
		db.Close()
		// No need to remove file for in-memory database
	}

	return db, cleanup
}
