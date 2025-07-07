package timeslot_handler

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/mishkahtherapy/brain/adapters/api"
	"github.com/mishkahtherapy/brain/core/domain"
	"github.com/mishkahtherapy/brain/core/domain/timeslot"
	timeslot_usecase "github.com/mishkahtherapy/brain/core/usecases/timeslot"
	"github.com/mishkahtherapy/brain/core/usecases/timeslot/bulk_toggle_therapist_timeslots"
	"github.com/mishkahtherapy/brain/core/usecases/timeslot/create_therapist_timeslot"
	"github.com/mishkahtherapy/brain/core/usecases/timeslot/delete_therapist_timeslot"
	"github.com/mishkahtherapy/brain/core/usecases/timeslot/get_therapist_timeslot"
	"github.com/mishkahtherapy/brain/core/usecases/timeslot/list_therapist_timeslots"
	"github.com/mishkahtherapy/brain/core/usecases/timeslot/update_therapist_timeslot"
)

// Response structure for API clients (in local timezone)
type TimeslotResponse struct {
	ID                string `json:"id"`
	TherapistID       string `json:"therapistId"`
	IsActive          bool   `json:"isActive"`
	DayOfWeek         string `json:"dayOfWeek"`         // Local day
	StartTime         string `json:"startTime"`         // Local time
	EndTime           string `json:"endTime"`           // Calculated local end time
	DurationMinutes   int    `json:"durationMinutes"`   // Duration in minutes
	PreSessionBuffer  int    `json:"preSessionBuffer"`  // minutes
	PostSessionBuffer int    `json:"postSessionBuffer"` // minutes
	CreatedAt         string `json:"createdAt"`
	UpdatedAt         string `json:"updatedAt"`
}
type TimeslotHandler struct {
	bulkToggleUsecase     bulk_toggle_therapist_timeslots.Usecase
	createTimeslotUsecase create_therapist_timeslot.Usecase
	getTimeslotUsecase    get_therapist_timeslot.Usecase
	updateTimeslotUsecase update_therapist_timeslot.Usecase
	deleteTimeslotUsecase delete_therapist_timeslot.Usecase
	listTimeslotsUsecase  list_therapist_timeslots.Usecase
}

func NewTimeslotHandler(
	bulkToggleUsecase bulk_toggle_therapist_timeslots.Usecase,
	createUsecase create_therapist_timeslot.Usecase,
	getUsecase get_therapist_timeslot.Usecase,
	updateUsecase update_therapist_timeslot.Usecase,
	deleteUsecase delete_therapist_timeslot.Usecase,
	listUsecase list_therapist_timeslots.Usecase,
) *TimeslotHandler {
	return &TimeslotHandler{
		bulkToggleUsecase:     bulkToggleUsecase,
		createTimeslotUsecase: createUsecase,
		getTimeslotUsecase:    getUsecase,
		updateTimeslotUsecase: updateUsecase,
		deleteTimeslotUsecase: deleteUsecase,
		listTimeslotsUsecase:  listUsecase,
	}
}

func (h *TimeslotHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("PUT /api/v1/therapists/{therapistId}/timeslots/bulk-toggle", h.handleBulkToggleTimeslots)
	mux.HandleFunc("POST /api/v1/therapists/{therapistId}/timeslots", h.handleCreateTimeslot)
	mux.HandleFunc("GET /api/v1/therapists/{therapistId}/timeslots", h.handleListTimeslots)
	mux.HandleFunc("GET /api/v1/therapists/{therapistId}/timeslots/{timeslotId}", h.handleGetTimeslot)
	mux.HandleFunc("PUT /api/v1/therapists/{therapistId}/timeslots/{timeslotId}", h.handleUpdateTimeslot)
	mux.HandleFunc("DELETE /api/v1/therapists/{therapistId}/timeslots/{timeslotId}", h.handleDeleteTimeslot)
}

func (h *TimeslotHandler) handleBulkToggleTimeslots(w http.ResponseWriter, r *http.Request) {
	rw := api.NewResponseWriter(w)

	// Read therapist ID from path
	therapistID := domain.TherapistID(r.PathValue("therapistId"))
	if therapistID == "" {
		rw.WriteBadRequest("Missing therapist ID")
		return
	}

	// Parse request body
	var requestBody struct {
		IsActive bool `json:"isActive"`
	}

	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
		rw.WriteBadRequest("Invalid request body: " + err.Error())
		return
	}

	// Create input for usecase
	input := bulk_toggle_therapist_timeslots.Input{
		TherapistID: therapistID,
		IsActive:    requestBody.IsActive,
	}

	err := h.bulkToggleUsecase.Execute(input)
	if err != nil {
		// Handle specific business logic errors
		switch err {
		case timeslot.ErrTherapistIDRequired:
			rw.WriteBadRequest(err.Error())
		case timeslot.ErrTherapistNotFound:
			rw.WriteNotFound(err.Error())
		default:
			rw.WriteError(err, http.StatusInternalServerError)
		}
		return
	}

	// Return simple success response
	response := map[string]string{
		"message": "Bulk toggle completed successfully",
	}

	if err := rw.WriteJSON(response, http.StatusOK); err != nil {
		rw.WriteError(err, http.StatusInternalServerError)
	}
}

func (h *TimeslotHandler) handleCreateTimeslot(w http.ResponseWriter, r *http.Request) {
	rw := api.NewResponseWriter(w)

	// Read therapist ID from path
	therapistID := domain.TherapistID(r.PathValue("therapistId"))
	if therapistID == "" {
		rw.WriteBadRequest("Missing therapist ID")
		return
	}

	// Parse request body
	var requestBody struct {
		DayOfWeek         string                `json:"dayOfWeek"`         // Local day
		StartTime         string                `json:"startTime"`         // Local time
		DurationMinutes   int                   `json:"durationMinutes"`   // Duration in minutes
		TimezoneOffset    domain.TimezoneOffset `json:"timezoneOffset"`    // Minutes from UTC
		PreSessionBuffer  int                   `json:"preSessionBuffer"`  // minutes
		PostSessionBuffer int                   `json:"postSessionBuffer"` // minutes
	}

	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
		rw.WriteBadRequest(err.Error())
		return
	}

	// Create input for usecase
	input := create_therapist_timeslot.Input{
		TherapistID:       therapistID,
		LocalDayOfWeek:    requestBody.DayOfWeek,
		LocalStartTime:    requestBody.StartTime,
		DurationMinutes:   requestBody.DurationMinutes,
		TimezoneOffset:    requestBody.TimezoneOffset,
		PreSessionBuffer:  requestBody.PreSessionBuffer,
		PostSessionBuffer: requestBody.PostSessionBuffer,
	}

	newTimeslot, err := h.createTimeslotUsecase.Execute(input)
	if err != nil {
		// Handle specific business logic errors
		switch err {
		case timeslot.ErrTherapistIDRequired,
			timeslot.ErrDayOfWeekIsRequired,
			timeslot.ErrStartTimeIsRequired,
			timeslot.ErrDurationIsRequired,
			timeslot.ErrTimezoneOffsetRequired,
			timeslot.ErrInvalidTimeFormat,
			timeslot.ErrInvalidDuration,
			timeslot.ErrInvalidTimezoneOffset,
			timeslot.ErrInvalidDayOfWeek,
			timeslot.ErrPreSessionBufferNegative,
			timeslot.ErrPostSessionBufferTooLow:
			rw.WriteBadRequest(err.Error())
		case timeslot.ErrTherapistNotFound:
			rw.WriteNotFound(err.Error())
		case timeslot.ErrOverlappingTimeslot:
			rw.WriteError(err, http.StatusConflict)
		default:
			rw.WriteError(err, http.StatusInternalServerError)
		}
		return
	}

	if err := rw.WriteJSON(newTimeslot, http.StatusCreated); err != nil {
		rw.WriteError(err, http.StatusInternalServerError)
	}
}

func (h *TimeslotHandler) handleListTimeslots(w http.ResponseWriter, r *http.Request) {
	rw := api.NewResponseWriter(w)

	// Read therapist ID from path
	therapistID := domain.TherapistID(r.PathValue("therapistId"))
	if therapistID == "" {
		rw.WriteBadRequest("Missing therapist ID")
		return
	}

	// Create input for usecase
	input := list_therapist_timeslots.Input{
		TherapistID: therapistID,
	}

	output, err := h.listTimeslotsUsecase.Execute(input)
	if err != nil {
		// Handle specific business logic errors
		switch err {
		case timeslot.ErrTherapistIDRequired,
			timeslot.ErrInvalidDayOfWeek:
			rw.WriteBadRequest(err.Error())
		case timeslot.ErrTherapistNotFound:
			rw.WriteNotFound(err.Error())
		default:
			rw.WriteError(err, http.StatusInternalServerError)
		}
		return
	}

	if err := rw.WriteJSON(output, http.StatusOK); err != nil {
		rw.WriteError(err, http.StatusInternalServerError)
	}
}

func (h *TimeslotHandler) handleGetTimeslot(w http.ResponseWriter, r *http.Request) {
	rw := api.NewResponseWriter(w)

	// Read therapist ID from path
	therapistID := domain.TherapistID(r.PathValue("therapistId"))
	if therapistID == "" {
		rw.WriteBadRequest("Missing therapist ID")
		return
	}

	// Read timeslot ID from path
	timeslotID := domain.TimeSlotID(r.PathValue("timeslotId"))
	if timeslotID == "" {
		rw.WriteBadRequest("Missing timeslot ID")
		return
	}

	// Parse timezone offset from query parameter (required for response conversion)
	timezoneOffsetParam := r.URL.Query().Get("timezoneOffset")
	if timezoneOffsetParam == "" {
		rw.WriteBadRequest("Missing timezoneOffset query parameter")
		return
	}

	var timezoneOffset domain.TimezoneOffset
	if _, err := fmt.Sscanf(timezoneOffsetParam, "%d", &timezoneOffset); err != nil {
		rw.WriteBadRequest("Invalid timezoneOffset format")
		return
	}

	// Validate timezone offset
	if err := timeslot_usecase.ValidateTimezoneOffset(timezoneOffset); err != nil {
		rw.WriteBadRequest(err.Error())
		return
	}

	// Create input for usecase
	input := get_therapist_timeslot.Input{
		TherapistID: therapistID,
		TimeslotID:  timeslotID,
	}

	dbTimeslot, err := h.getTimeslotUsecase.Execute(input)
	if err != nil {
		// Handle specific business logic errors
		switch err {
		case timeslot.ErrTherapistIDRequired,
			timeslot.ErrTimeslotIDIsRequired:
			rw.WriteBadRequest(err.Error())
		case timeslot.ErrTherapistNotFound,
			timeslot.ErrTimeslotNotFound,
			timeslot.ErrTimeslotNotOwned:
			rw.WriteNotFound(err.Error())
		default:
			rw.WriteError(err, http.StatusInternalServerError)
		}
		return
	}

	if err := rw.WriteJSON(dbTimeslot, http.StatusOK); err != nil {
		rw.WriteError(err, http.StatusInternalServerError)
	}
}

func (h *TimeslotHandler) handleUpdateTimeslot(w http.ResponseWriter, r *http.Request) {
	rw := api.NewResponseWriter(w)

	// Read therapist ID from path
	therapistID := domain.TherapistID(r.PathValue("therapistId"))
	if therapistID == "" {
		rw.WriteBadRequest("Missing therapist ID")
		return
	}

	// Read timeslot ID from path
	timeslotID := domain.TimeSlotID(r.PathValue("timeslotId"))
	if timeslotID == "" {
		rw.WriteBadRequest("Missing timeslot ID")
		return
	}

	// Parse timezone offset from query parameter (required for input conversion)
	timezoneOffsetParam := r.URL.Query().Get("timezoneOffset")
	if timezoneOffsetParam == "" {
		rw.WriteBadRequest("Missing timezoneOffset query parameter")
		return
	}

	var timezoneOffset domain.TimezoneOffset
	if _, err := fmt.Sscanf(timezoneOffsetParam, "%d", &timezoneOffset); err != nil {
		rw.WriteBadRequest("Invalid timezoneOffset format")
		return
	}

	// Validate timezone offset
	if err := timeslot_usecase.ValidateTimezoneOffset(timezoneOffset); err != nil {
		rw.WriteBadRequest(err.Error())
		return
	}

	// Parse request body (contains local timezone data)
	var requestBody struct {
		DayOfWeek         timeslot.DayOfWeek `json:"dayOfWeek"`
		StartTime         string             `json:"startTime"` // Local time
		DurationMinutes   int                `json:"durationMinutes"`
		PreSessionBuffer  int                `json:"preSessionBuffer"`
		PostSessionBuffer int                `json:"postSessionBuffer"`
		IsActive          bool               `json:"isActive"`
	}

	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
		rw.WriteBadRequest(err.Error())
		return
	}

	// Convert local input to UTC for storage
	utcDayOfWeek, utcStartTime, err := timeslot_usecase.ConvertLocalToUTC(
		string(requestBody.DayOfWeek), requestBody.StartTime, timezoneOffset)
	if err != nil {
		rw.WriteBadRequest(err.Error())
		return
	}

	// Create input for usecase (with UTC data)
	input := update_therapist_timeslot.Input{
		TherapistID:       therapistID,
		TimeslotID:        timeslotID,
		DayOfWeek:         timeslot.DayOfWeek(utcDayOfWeek),
		StartTime:         utcStartTime,
		DurationMinutes:   requestBody.DurationMinutes,
		PreSessionBuffer:  requestBody.PreSessionBuffer,
		PostSessionBuffer: requestBody.PostSessionBuffer,
		IsActive:          requestBody.IsActive,
	}

	updatedTimeslot, err := h.updateTimeslotUsecase.Execute(input)
	if err != nil {
		// Handle specific business logic errors
		switch err {
		case timeslot.ErrTherapistIDRequired,
			timeslot.ErrTimeslotIDIsRequired,
			timeslot.ErrDayOfWeekIsRequired,
			timeslot.ErrStartTimeIsRequired,
			timeslot.ErrDurationIsRequired,
			timeslot.ErrInvalidDayOfWeek,
			timeslot.ErrInvalidTimeFormat,
			timeslot.ErrInvalidDuration,
			timeslot.ErrPreSessionBufferNegative,
			timeslot.ErrPostSessionBufferTooLow:
			rw.WriteBadRequest(err.Error())
		case timeslot.ErrTherapistNotFound,
			timeslot.ErrTimeslotNotFound,
			timeslot.ErrTimeslotNotOwned:
			rw.WriteNotFound(err.Error())
		case timeslot.ErrOverlappingTimeslot:
			rw.WriteError(err, http.StatusConflict)
		default:
			rw.WriteError(err, http.StatusInternalServerError)
		}
		return
	}

	if err := rw.WriteJSON(updatedTimeslot, http.StatusOK); err != nil {
		rw.WriteError(err, http.StatusInternalServerError)
	}
}

func (h *TimeslotHandler) handleDeleteTimeslot(w http.ResponseWriter, r *http.Request) {
	rw := api.NewResponseWriter(w)

	// Read therapist ID from path
	therapistID := domain.TherapistID(r.PathValue("therapistId"))
	if therapistID == "" {
		rw.WriteBadRequest("Missing therapist ID")
		return
	}

	// Read timeslot ID from path
	timeslotID := domain.TimeSlotID(r.PathValue("timeslotId"))
	if timeslotID == "" {
		rw.WriteBadRequest("Missing timeslot ID")
		return
	}

	// Create input for usecase
	input := delete_therapist_timeslot.Input{
		TherapistID: therapistID,
		TimeslotID:  timeslotID,
	}

	err := h.deleteTimeslotUsecase.Execute(input)
	if err != nil {
		// Handle specific business logic errors
		switch err {
		case timeslot.ErrTherapistIDRequired,
			timeslot.ErrTimeslotIDIsRequired:
			rw.WriteBadRequest(err.Error())
		case timeslot.ErrTherapistNotFound,
			timeslot.ErrTimeslotNotFound,
			timeslot.ErrTimeslotNotOwned:
			rw.WriteNotFound(err.Error())
		case timeslot.ErrTimeslotHasActiveBookings:
			rw.WriteError(err, http.StatusConflict)
		default:
			rw.WriteError(err, http.StatusInternalServerError)
		}
		return
	}

	// Return 204 No Content for successful deletion
	w.WriteHeader(http.StatusNoContent)
}
