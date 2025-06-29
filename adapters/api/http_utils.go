package api

import (
	"encoding/json"
	"net/http"
)

// ResponseWriter wraps common HTTP response writing operations
type ResponseWriter struct {
	w http.ResponseWriter
}

// NewResponseWriter creates a new ResponseWriter
func NewResponseWriter(w http.ResponseWriter) *ResponseWriter {
	return &ResponseWriter{w: w}
}

// WriteJSON writes a JSON response with the specified status code
func (rw *ResponseWriter) WriteJSON(data any, statusCode int) error {
	rw.w.Header().Set("Content-Type", "application/json")
	rw.w.WriteHeader(statusCode)
	return json.NewEncoder(rw.w).Encode(data)
}

// WriteError writes an error response with the specified status code
func (rw *ResponseWriter) WriteError(err error, statusCode int) {
	http.Error(rw.w, err.Error(), statusCode)
}

// WriteErrorMessage writes an error message with the specified status code
func (rw *ResponseWriter) WriteErrorMessage(message string, statusCode int) {
	http.Error(rw.w, message, statusCode)
}

// WriteCreated writes a 201 Created response
func (rw *ResponseWriter) WriteCreated() {
	rw.w.WriteHeader(http.StatusCreated)
}

// WriteNoContent writes a 204 No Content response
func (rw *ResponseWriter) WriteNoContent() {
	rw.w.WriteHeader(http.StatusNoContent)
}

// WriteMethodNotAllowed writes a 405 Method Not Allowed response
func (rw *ResponseWriter) WriteMethodNotAllowed() {
	http.Error(rw.w, "Method not allowed", http.StatusMethodNotAllowed)
}

type ErrorResponse struct {
	Error string `json:"error"`
}

// WriteBadRequest writes a 400 Bad Request response
func (rw *ResponseWriter) WriteBadRequest(message string) {
	rw.w.Header().Set("Content-Type", "application/json")
	rw.w.WriteHeader(http.StatusBadRequest)
	json.NewEncoder(rw.w).Encode(ErrorResponse{Error: message})
}

// WriteNotFound writes a 404 Not Found response
func (rw *ResponseWriter) WriteNotFound(message string) {
	rw.w.Header().Set("Content-Type", "application/json")
	rw.w.WriteHeader(http.StatusNotFound)
	json.NewEncoder(rw.w).Encode(ErrorResponse{Error: message})
}
