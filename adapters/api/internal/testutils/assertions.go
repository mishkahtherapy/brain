package testutils

import (
	"encoding/json"
	"net/http/httptest"
	"testing"
)

// AssertStatus verifies the HTTP status code
func AssertStatus(t *testing.T, rec *httptest.ResponseRecorder, expectedStatus int) {
	t.Helper()
	if rec.Code != expectedStatus {
		t.Fatalf("Expected status %d, got %d. Body: %s", expectedStatus, rec.Code, rec.Body.String())
	}
}

// AssertJSONResponse verifies response status and parses JSON
func AssertJSONResponse(t *testing.T, rec *httptest.ResponseRecorder, expectedStatus int, target interface{}) {
	t.Helper()
	AssertStatus(t, rec, expectedStatus)

	if err := json.Unmarshal(rec.Body.Bytes(), target); err != nil {
		t.Fatalf("Failed to parse JSON response: %v. Body: %s", err, rec.Body.String())
	}
}

// AssertError verifies response contains an error message
func AssertError(t *testing.T, rec *httptest.ResponseRecorder, expectedStatus int) {
	t.Helper()
	AssertStatus(t, rec, expectedStatus)

	var errorResponse map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &errorResponse); err != nil {
		t.Fatalf("Failed to parse error response JSON: %v. Body: %s", err, rec.Body.String())
	}

	if _, hasError := errorResponse["error"]; !hasError {
		t.Errorf("Expected error response to contain 'error' field. Body: %s", rec.Body.String())
	}
}

// AssertStringField verifies a string field in JSON response
func AssertStringField(t *testing.T, data map[string]interface{}, fieldName string, expected string) {
	t.Helper()
	actual, ok := data[fieldName].(string)
	if !ok {
		t.Fatalf("Expected field %s to be a string, got %T", fieldName, data[fieldName])
	}
	if actual != expected {
		t.Errorf("Expected %s to be %s, got %s", fieldName, expected, actual)
	}
}

// AssertBoolField verifies a boolean field in JSON response
func AssertBoolField(t *testing.T, data map[string]interface{}, fieldName string, expected bool) {
	t.Helper()
	actual, ok := data[fieldName].(bool)
	if !ok {
		t.Fatalf("Expected field %s to be a bool, got %T", fieldName, data[fieldName])
	}
	if actual != expected {
		t.Errorf("Expected %s to be %t, got %t", fieldName, expected, actual)
	}
}

// AssertFloatField verifies a numeric field in JSON response
func AssertFloatField(t *testing.T, data map[string]interface{}, fieldName string, expected float64) {
	t.Helper()
	actual, ok := data[fieldName].(float64)
	if !ok {
		t.Fatalf("Expected field %s to be a number, got %T", fieldName, data[fieldName])
	}
	if actual != expected {
		t.Errorf("Expected %s to be %f, got %f", fieldName, expected, actual)
	}
}

// AssertFieldExists verifies a field exists in JSON response
func AssertFieldExists(t *testing.T, data map[string]interface{}, fieldName string) {
	t.Helper()
	if _, exists := data[fieldName]; !exists {
		t.Errorf("Expected field %s to exist in response", fieldName)
	}
}
