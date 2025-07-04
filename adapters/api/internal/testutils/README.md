# Test Utilities

This package provides shared utilities for E2E tests to reduce code duplication and standardize test setup across different handlers.

## Philosophy

These utilities extract only the **boring, infrastructure-related code** while keeping test logic explicit and readable in the test files themselves. This means:

- ✅ Database setup/cleanup
- ✅ Entity creation with sensible defaults  
- ✅ Repository wiring
- ✅ Basic HTTP assertions
- ❌ Request builders (makes tests harder to read)
- ❌ Complex test flows (business logic should be visible)
- ❌ Domain-specific assertions (keep in tests for clarity)

## Usage

### Database Setup

```go
// For most tests - uses temporary file
database, cleanup := testutils.SetupTestDB(t)
defer cleanup()

// For faster tests - uses in-memory DB
database, cleanup := testutils.SetupInMemoryTestDB(t)
defer cleanup()
```

### Entity Creation

```go
// Create basic entities with defaults
therapistID := testutils.CreateTestTherapist(t, database)
clientID := testutils.CreateTestClient(t, database)

// Create with custom values
therapistID := testutils.CreateTestTherapistWithName(t, database, "Dr. Custom Name")

// Create custom timeslots
timeslotID := testutils.CreateTestTimeSlotCustom(t, database, therapistID, "Monday", "10:00", 60, true)

// Create full test data set
testData := testutils.CreateFullTestData(t, database)
// testData now contains: TherapistID, ClientID, TimeSlotID, SpecializationID
```

### Repository Setup

```go
repos := testutils.SetupRepositories(database)
// repos contains: TherapistRepo, TimeSlotRepo, BookingRepo, ClientRepo, SessionRepo

// Use in usecases
createUsecase := create_booking.NewUsecase(
    repos.BookingRepo, 
    repos.TherapistRepo, 
    repos.ClientRepo, 
    repos.TimeSlotRepo,
)
```

### HTTP Assertions

```go
// Basic status check
testutils.AssertStatus(t, rec, http.StatusOK)

// Parse JSON and check status
var response map[string]interface{}
testutils.AssertJSONResponse(t, rec, http.StatusCreated, &response)

// Check specific fields
testutils.AssertStringField(t, response, "name", "Expected Name")
testutils.AssertBoolField(t, response, "isActive", true)
testutils.AssertFloatField(t, response, "amount", 99.99)

// Check error responses
testutils.AssertError(t, rec, http.StatusBadRequest)
```

## Example Test Structure

```go
func TestMyFeature(t *testing.T) {
    // Setup infrastructure (boring stuff -> utilities)
    database, cleanup := testutils.SetupTestDB(t)
    defer cleanup()
    
    therapistID := testutils.CreateTestTherapist(t, database)
    repos := testutils.SetupRepositories(database)
    
    // Setup usecases and handlers (test-specific logic -> explicit)
    usecase := my_usecase.NewUsecase(repos.TherapistRepo)
    handler := my_handler.NewHandler(usecase)
    
    mux := http.NewServeMux()
    handler.RegisterRoutes(mux)
    
    t.Run("Test scenario", func(t *testing.T) {
        // HTTP request construction (explicit -> readable)
        requestData := map[string]interface{}{
            "therapistId": therapistID,
            "action": "doSomething",
        }
        body, _ := json.Marshal(requestData)
        
        req := httptest.NewRequest("POST", "/api/v1/my-endpoint", bytes.NewBuffer(body))
        req.Header.Set("Content-Type", "application/json")
        rec := httptest.NewRecorder()
        
        mux.ServeHTTP(rec, req)
        
        // Basic assertions (utilities for common patterns)
        var response map[string]interface{}
        testutils.AssertJSONResponse(t, rec, http.StatusOK, &response)
        testutils.AssertStringField(t, response, "status", "success")
        
        // Domain-specific assertions (explicit -> meaningful)
        if response["therapistId"] != string(therapistID) {
            t.Errorf("Expected therapistId %s, got %s", therapistID, response["therapistId"])
        }
    })
}
```

## Benefits

1. **Eliminates duplicate infrastructure code** - database setup, entity creation, repository wiring
2. **Keeps tests readable** - business logic and assertions remain in test files
3. **Standardizes test patterns** - consistent setup across all handlers
4. **Reduces maintenance** - infrastructure changes only need updating in one place
5. **Fast compilation** - utilities build independently and are reused

## Files

- `database.go` - Database setup and cleanup utilities
- `entities.go` - Test entity creation with sensible defaults
- `repositories.go` - Test repository implementations and wiring
- `assertions.go` - Basic HTTP response assertions
- `handlers.go` - Common repository setup patterns 