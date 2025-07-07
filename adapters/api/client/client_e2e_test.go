package client_handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/mishkahtherapy/brain/adapters/db"
	"github.com/mishkahtherapy/brain/adapters/db/client_db"
	"github.com/mishkahtherapy/brain/core/domain"
	"github.com/mishkahtherapy/brain/core/domain/client"
	"github.com/mishkahtherapy/brain/core/ports"
	"github.com/mishkahtherapy/brain/core/usecases/client/create_client"
	"github.com/mishkahtherapy/brain/core/usecases/client/get_all_clients"
	"github.com/mishkahtherapy/brain/core/usecases/client/get_client"

	_ "github.com/glebarez/go-sqlite"
)

func TestClientE2E(t *testing.T) {
	// Setup test database
	database, cleanup := setupClientTestDB(t)
	defer cleanup()

	// Setup repositories
	clientRepo := client_db.NewClientRepository(database)

	// Setup usecases
	createUsecase := create_client.NewUsecase(clientRepo)
	getAllUsecase := get_all_clients.NewUsecase(clientRepo)
	getUsecase := get_client.NewUsecase(clientRepo)

	// Setup handler
	clientHandler := NewClientHandler(*createUsecase, *getAllUsecase, *getUsecase)

	// Setup router
	mux := http.NewServeMux()
	clientHandler.RegisterRoutes(mux)

	t.Run("Complete client workflow", func(t *testing.T) {
		// Step 1: Create a new client
		clientData := map[string]interface{}{
			"name":           "John Doe",
			"whatsAppNumber": "+1234567890",
			"timezoneOffset": -300,
		}
		clientBody, _ := json.Marshal(clientData)

		createReq := httptest.NewRequest("POST", "/api/v1/clients", bytes.NewBuffer(clientBody))
		createReq.Header.Set("Content-Type", "application/json")
		createRec := httptest.NewRecorder()

		mux.ServeHTTP(createRec, createReq)

		// Verify creation response
		if createRec.Code != http.StatusCreated {
			t.Fatalf("Expected status %d, got %d. Body: %s", http.StatusCreated, createRec.Code, createRec.Body.String())
		}

		// Parse created client
		var createdClient client.Client
		if err := json.Unmarshal(createRec.Body.Bytes(), &createdClient); err != nil {
			t.Fatalf("Failed to parse created client: %v", err)
		}

		// Verify created client data
		if createdClient.Name != "John Doe" {
			t.Errorf("Expected name %s, got %s", "John Doe", createdClient.Name)
		}
		if createdClient.WhatsAppNumber != "+1234567890" {
			t.Errorf("Expected WhatsApp number %s, got %s", "+1234567890", createdClient.WhatsAppNumber)
		}
		if createdClient.TimezoneOffset != -300 {
			t.Errorf("Expected timezoneOffset %d, got %d", -300, createdClient.TimezoneOffset)
		}
		if createdClient.ID == "" {
			t.Error("Expected ID to be set")
		}
		if len(createdClient.BookingIDs) != 0 {
			t.Errorf("Expected empty booking IDs, got %v", createdClient.BookingIDs)
		}
		if createdClient.CreatedAt == (domain.UTCTimestamp{}) {
			t.Error("Expected CreatedAt to be set")
		}
		if createdClient.UpdatedAt == (domain.UTCTimestamp{}) {
			t.Error("Expected UpdatedAt to be set")
		}

		// Step 2: Get the client by ID
		getReq := httptest.NewRequest("GET", "/api/v1/clients/"+string(createdClient.ID), nil)
		getRec := httptest.NewRecorder()

		mux.ServeHTTP(getRec, getReq)

		// Verify get response
		if getRec.Code != http.StatusOK {
			t.Fatalf("Expected status %d, got %d. Body: %s", http.StatusOK, getRec.Code, getRec.Body.String())
		}

		// Parse retrieved client
		var retrievedClient client.Client
		if err := json.Unmarshal(getRec.Body.Bytes(), &retrievedClient); err != nil {
			t.Fatalf("Failed to parse retrieved client: %v", err)
		}

		// Verify retrieved client matches created one
		if retrievedClient.ID != createdClient.ID {
			t.Errorf("Expected ID %s, got %s", createdClient.ID, retrievedClient.ID)
		}
		if retrievedClient.Name != createdClient.Name {
			t.Errorf("Expected name %s, got %s", createdClient.Name, retrievedClient.Name)
		}
		if retrievedClient.WhatsAppNumber != createdClient.WhatsAppNumber {
			t.Errorf("Expected WhatsApp number %s, got %s", createdClient.WhatsAppNumber, retrievedClient.WhatsAppNumber)
		}
		if retrievedClient.TimezoneOffset != createdClient.TimezoneOffset {
			t.Errorf("Expected timezoneOffset %d, got %d", createdClient.TimezoneOffset, retrievedClient.TimezoneOffset)
		}

		// Step 3: Get all clients
		getAllReq := httptest.NewRequest("GET", "/api/v1/clients", nil)
		getAllRec := httptest.NewRecorder()

		mux.ServeHTTP(getAllRec, getAllReq)

		// Verify get all response
		if getAllRec.Code != http.StatusOK {
			t.Fatalf("Expected status %d, got %d. Body: %s", http.StatusOK, getAllRec.Code, getAllRec.Body.String())
		}

		// Parse all clients
		var allClients []*client.Client
		if err := json.Unmarshal(getAllRec.Body.Bytes(), &allClients); err != nil {
			t.Fatalf("Failed to parse all clients: %v", err)
		}

		// Verify our client is in the list
		found := false
		for _, client := range allClients {
			if client.ID == createdClient.ID {
				found = true
				break
			}
		}
		if !found {
			t.Error("Created client not found in list of all clients")
		}
	})

	t.Run("Timezone validation", func(t *testing.T) {
		// Test create client without timezone (should fail)
		clientWithoutTimezoneData := map[string]interface{}{
			"name":           "Test User",
			"whatsAppNumber": "+1234567893",
		}
		clientWithoutTimezoneBody, _ := json.Marshal(clientWithoutTimezoneData)

		createReq := httptest.NewRequest("POST", "/api/v1/clients", bytes.NewBuffer(clientWithoutTimezoneBody))
		createReq.Header.Set("Content-Type", "application/json")
		createRec := httptest.NewRecorder()

		mux.ServeHTTP(createRec, createReq)

		if createRec.Code != http.StatusCreated {
			t.Errorf("Expected status %d when timezoneOffset omitted (default 0), got %d. Body: %s", http.StatusCreated, createRec.Code, createRec.Body.String())
		}

		// Test create client with invalid timezone (should fail)
		clientWithInvalidTimezoneData := map[string]interface{}{
			"name":           "Test User",
			"whatsAppNumber": "+1234567894",
			"timezoneOffset": 9999,
		}
		clientWithInvalidTimezoneBody, _ := json.Marshal(clientWithInvalidTimezoneData)

		createInvalidReq := httptest.NewRequest("POST", "/api/v1/clients", bytes.NewBuffer(clientWithInvalidTimezoneBody))
		createInvalidReq.Header.Set("Content-Type", "application/json")
		createInvalidRec := httptest.NewRecorder()

		mux.ServeHTTP(createInvalidRec, createInvalidReq)

		if createInvalidRec.Code != http.StatusBadRequest {
			t.Errorf("Expected status %d for invalid timezone, got %d. Body: %s", http.StatusBadRequest, createInvalidRec.Code, createInvalidRec.Body.String())
		}

		// Test create client with valid timezone
		clientWithValidTimezoneData := map[string]interface{}{
			"name":           "Test User",
			"whatsAppNumber": "+1234567895",
			"timezoneOffset": 0,
		}
		clientWithValidTimezoneBody, _ := json.Marshal(clientWithValidTimezoneData)

		createValidReq := httptest.NewRequest("POST", "/api/v1/clients", bytes.NewBuffer(clientWithValidTimezoneBody))
		createValidReq.Header.Set("Content-Type", "application/json")
		createValidRec := httptest.NewRecorder()

		mux.ServeHTTP(createValidRec, createValidReq)

		if createValidRec.Code != http.StatusCreated {
			t.Errorf("Expected status %d for valid timezone, got %d. Body: %s", http.StatusCreated, createValidRec.Code, createValidRec.Body.String())
		}

		// Parse and verify the created client has correct timezone
		var createdClient client.Client
		if err := json.Unmarshal(createValidRec.Body.Bytes(), &createdClient); err != nil {
			t.Fatalf("Failed to parse created client: %v", err)
		}
		if createdClient.TimezoneOffset != 0 {
			t.Errorf("Expected timezoneOffset %d, got %d", 0, createdClient.TimezoneOffset)
		}
	})

	t.Run("Error cases", func(t *testing.T) {
		// Test get non-existent client
		nonExistentID := "client_00000000-0000-0000-0000-000000000000"
		getReq := httptest.NewRequest("GET", "/api/v1/clients/"+nonExistentID, nil)
		getRec := httptest.NewRecorder()

		mux.ServeHTTP(getRec, getReq)

		if getRec.Code != http.StatusNotFound {
			t.Errorf("Expected status %d for non-existent client, got %d", http.StatusNotFound, getRec.Code)
		}

		// Test create client with missing name (should work since name is optional)
		clientWithoutNameData := map[string]interface{}{
			"whatsAppNumber": "+1234567891", // Different number to avoid duplicate
			"timezoneOffset": 0,
		}
		clientWithoutNameBody, _ := json.Marshal(clientWithoutNameData)

		createWithoutNameReq := httptest.NewRequest("POST", "/api/v1/clients", bytes.NewBuffer(clientWithoutNameBody))
		createWithoutNameReq.Header.Set("Content-Type", "application/json")
		createWithoutNameRec := httptest.NewRecorder()

		mux.ServeHTTP(createWithoutNameRec, createWithoutNameReq)

		if createWithoutNameRec.Code != http.StatusCreated {
			t.Errorf("Expected status %d for client without name, got %d. Body: %s", http.StatusCreated, createWithoutNameRec.Code, createWithoutNameRec.Body.String())
		}

		// Test create client with invalid data (missing WhatsApp number)
		invalidClientData := map[string]interface{}{
			"name":           "Test User",
			"timezoneOffset": 0,
		}
		invalidBody, _ := json.Marshal(invalidClientData)

		createReq := httptest.NewRequest("POST", "/api/v1/clients", bytes.NewBuffer(invalidBody))
		createReq.Header.Set("Content-Type", "application/json")
		createRec := httptest.NewRecorder()

		mux.ServeHTTP(createRec, createReq)

		if createRec.Code != http.StatusBadRequest {
			t.Errorf("Expected status %d for missing WhatsApp number, got %d", http.StatusBadRequest, createRec.Code)
		}

		// Test create client with duplicate WhatsApp number
		// First create a client
		firstClientData := map[string]interface{}{
			"name":           "First Client",
			"whatsAppNumber": "+1234567892", // Unique number for this test
			"timezoneOffset": 0,
		}
		firstClientBody, _ := json.Marshal(firstClientData)

		firstReq := httptest.NewRequest("POST", "/api/v1/clients", bytes.NewBuffer(firstClientBody))
		firstReq.Header.Set("Content-Type", "application/json")
		firstRec := httptest.NewRecorder()

		mux.ServeHTTP(firstRec, firstReq)

		if firstRec.Code != http.StatusCreated {
			t.Fatalf("Failed to create first client for duplicate test: %d", firstRec.Code)
		}

		// Now try to create a duplicate
		duplicateClientData := map[string]interface{}{
			"name":           "Jane Doe",
			"whatsAppNumber": "+1234567892", // Same as first client
			"timezoneOffset": 0,
		}
		duplicateBody, _ := json.Marshal(duplicateClientData)

		duplicateReq := httptest.NewRequest("POST", "/api/v1/clients", bytes.NewBuffer(duplicateBody))
		duplicateReq.Header.Set("Content-Type", "application/json")
		duplicateRec := httptest.NewRecorder()

		mux.ServeHTTP(duplicateRec, duplicateReq)

		if duplicateRec.Code != http.StatusConflict {
			t.Errorf("Expected status %d for duplicate WhatsApp number, got %d", http.StatusConflict, duplicateRec.Code)
		}

		// Test create client with invalid WhatsApp number format
		invalidWhatsAppData := map[string]interface{}{
			"name":           "Invalid User",
			"whatsAppNumber": "invalid",
			"timezoneOffset": 0,
		}
		invalidWhatsAppBody, _ := json.Marshal(invalidWhatsAppData)

		invalidWhatsAppReq := httptest.NewRequest("POST", "/api/v1/clients", bytes.NewBuffer(invalidWhatsAppBody))
		invalidWhatsAppReq.Header.Set("Content-Type", "application/json")
		invalidWhatsAppRec := httptest.NewRecorder()

		mux.ServeHTTP(invalidWhatsAppRec, invalidWhatsAppReq)

		if invalidWhatsAppRec.Code != http.StatusBadRequest {
			t.Errorf("Expected status %d for invalid WhatsApp number format, got %d", http.StatusBadRequest, invalidWhatsAppRec.Code)
		}
	})
}

func setupClientTestDB(t *testing.T) (ports.SQLDatabase, func()) {
	// Use temporary file database for testing instead of :memory:
	tmpfile, err := os.CreateTemp("", "client_test_*.db")
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
