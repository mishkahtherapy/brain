package api

import (
	"encoding/json"
	"net/http"

	"github.com/mishkahtherapy/brain/core/domain"
	timeslotpkg "github.com/mishkahtherapy/brain/core/usecases/timeslot"
	"github.com/mishkahtherapy/brain/core/usecases/timeslot/create_therapist_timeslot"
	"github.com/mishkahtherapy/brain/core/usecases/timeslot/delete_therapist_timeslot"
	"github.com/mishkahtherapy/brain/core/usecases/timeslot/get_therapist_timeslot"
	"github.com/mishkahtherapy/brain/core/usecases/timeslot/list_therapist_timeslots"
	"github.com/mishkahtherapy/brain/core/usecases/timeslot/update_therapist_timeslot"
)

type TimeslotHandler struct {
	createTimeslotUsecase create_therapist_timeslot.Usecase
	getTimeslotUsecase    get_therapist_timeslot.Usecase
	updateTimeslotUsecase update_therapist_timeslot.Usecase
	deleteTimeslotUsecase delete_therapist_timeslot.Usecase
	listTimeslotsUsecase  list_therapist_timeslots.Usecase
}

func NewTimeslotHandler(
	createUsecase create_therapist_timeslot.Usecase,
	getUsecase get_therapist_timeslot.Usecase,
	updateUsecase update_therapist_timeslot.Usecase,
	deleteUsecase delete_therapist_timeslot.Usecase,
	listUsecase list_therapist_timeslots.Usecase,
) *TimeslotHandler {
	return &TimeslotHandler{
		createTimeslotUsecase: createUsecase,
		getTimeslotUsecase:    getUsecase,
		updateTimeslotUsecase: updateUsecase,
		deleteTimeslotUsecase: deleteUsecase,
		listTimeslotsUsecase:  listUsecase,
	}
}

func (h *TimeslotHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/v1/therapists/{therapistId}/timeslots", h.handleCreateTimeslot)
	mux.HandleFunc("GET /api/v1/therapists/{therapistId}/timeslots", h.handleListTimeslots)
	mux.HandleFunc("GET /api/v1/therapists/{therapistId}/timeslots/{timeslotId}", h.handleGetTimeslot)
	mux.HandleFunc("PUT /api/v1/therapists/{therapistId}/timeslots/{timeslotId}", h.handleUpdateTimeslot)
	mux.HandleFunc("DELETE /api/v1/therapists/{therapistId}/timeslots/{timeslotId}", h.handleDeleteTimeslot)
}

func (h *TimeslotHandler) handleCreateTimeslot(w http.ResponseWriter, r *http.Request) {
	rw := NewResponseWriter(w)

	// Read therapist ID from path
	therapistID := domain.TherapistID(r.PathValue("therapistId"))
	if therapistID == "" {
		rw.WriteBadRequest("Missing therapist ID")
		return
	}

	// Parse request body
	var requestBody struct {
		DayOfWeek         domain.DayOfWeek `json:"dayOfWeek"`
		StartTime         string           `json:"startTime"`
		EndTime           string           `json:"endTime"`
		PreSessionBuffer  int              `json:"preSessionBuffer"`
		PostSessionBuffer int              `json:"postSessionBuffer"`
	}

	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
		rw.WriteBadRequest(err.Error())
		return
	}

	// Create input for usecase
	input := create_therapist_timeslot.Input{
		TherapistID:       therapistID,
		DayOfWeek:         requestBody.DayOfWeek,
		StartTime:         requestBody.StartTime,
		EndTime:           requestBody.EndTime,
		PreSessionBuffer:  requestBody.PreSessionBuffer,
		PostSessionBuffer: requestBody.PostSessionBuffer,
	}

	timeslot, err := h.createTimeslotUsecase.Execute(input)
	if err != nil {
		// Handle specific business logic errors
		switch err {
		case timeslotpkg.ErrTherapistIDIsRequired,
			timeslotpkg.ErrDayOfWeekIsRequired,
			timeslotpkg.ErrStartTimeIsRequired,
			timeslotpkg.ErrEndTimeIsRequired,
			timeslotpkg.ErrInvalidDayOfWeek,
			timeslotpkg.ErrInvalidTimeFormat,
			timeslotpkg.ErrInvalidTimeRange,
			timeslotpkg.ErrPreSessionBufferNegative,
			timeslotpkg.ErrPostSessionBufferTooLow:
			rw.WriteBadRequest(err.Error())
		case timeslotpkg.ErrTherapistNotFound:
			rw.WriteNotFound(err.Error())
		case timeslotpkg.ErrOverlappingTimeslot:
			rw.WriteError(err, http.StatusConflict)
		default:
			rw.WriteError(err, http.StatusInternalServerError)
		}
		return
	}

	if err := rw.WriteJSON(timeslot, http.StatusCreated); err != nil {
		rw.WriteError(err, http.StatusInternalServerError)
	}
}

func (h *TimeslotHandler) handleListTimeslots(w http.ResponseWriter, r *http.Request) {
	rw := NewResponseWriter(w)

	// Read therapist ID from path
	therapistID := domain.TherapistID(r.PathValue("therapistId"))
	if therapistID == "" {
		rw.WriteBadRequest("Missing therapist ID")
		return
	}

	// Parse optional day query parameter
	dayParam := r.URL.Query().Get("day")
	var dayFilter *domain.DayOfWeek
	if dayParam != "" {
		day := domain.DayOfWeek(dayParam)
		dayFilter = &day
	}

	// Create input for usecase
	input := list_therapist_timeslots.Input{
		TherapistID: therapistID,
		DayOfWeek:   dayFilter,
	}

	output, err := h.listTimeslotsUsecase.Execute(input)
	if err != nil {
		// Handle specific business logic errors
		switch err {
		case timeslotpkg.ErrTherapistIDIsRequired,
			timeslotpkg.ErrInvalidDayOfWeek:
			rw.WriteBadRequest(err.Error())
		case timeslotpkg.ErrTherapistNotFound:
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
	rw := NewResponseWriter(w)

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
	input := get_therapist_timeslot.Input{
		TherapistID: therapistID,
		TimeslotID:  timeslotID,
	}

	timeslot, err := h.getTimeslotUsecase.Execute(input)
	if err != nil {
		// Handle specific business logic errors
		switch err {
		case timeslotpkg.ErrTherapistIDIsRequired,
			timeslotpkg.ErrTimeslotIDIsRequired:
			rw.WriteBadRequest(err.Error())
		case timeslotpkg.ErrTherapistNotFound,
			timeslotpkg.ErrTimeslotNotFound,
			timeslotpkg.ErrTimeslotNotOwned:
			rw.WriteNotFound(err.Error())
		default:
			rw.WriteError(err, http.StatusInternalServerError)
		}
		return
	}

	if err := rw.WriteJSON(timeslot, http.StatusOK); err != nil {
		rw.WriteError(err, http.StatusInternalServerError)
	}
}

func (h *TimeslotHandler) handleUpdateTimeslot(w http.ResponseWriter, r *http.Request) {
	rw := NewResponseWriter(w)

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

	// Parse request body
	var requestBody struct {
		DayOfWeek         domain.DayOfWeek `json:"dayOfWeek"`
		StartTime         string           `json:"startTime"`
		EndTime           string           `json:"endTime"`
		PreSessionBuffer  int              `json:"preSessionBuffer"`
		PostSessionBuffer int              `json:"postSessionBuffer"`
	}

	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
		rw.WriteBadRequest(err.Error())
		return
	}

	// Create input for usecase
	input := update_therapist_timeslot.Input{
		TherapistID:       therapistID,
		TimeslotID:        timeslotID,
		DayOfWeek:         requestBody.DayOfWeek,
		StartTime:         requestBody.StartTime,
		EndTime:           requestBody.EndTime,
		PreSessionBuffer:  requestBody.PreSessionBuffer,
		PostSessionBuffer: requestBody.PostSessionBuffer,
	}

	timeslot, err := h.updateTimeslotUsecase.Execute(input)
	if err != nil {
		// Handle specific business logic errors
		switch err {
		case timeslotpkg.ErrTherapistIDIsRequired,
			timeslotpkg.ErrTimeslotIDIsRequired,
			timeslotpkg.ErrDayOfWeekIsRequired,
			timeslotpkg.ErrStartTimeIsRequired,
			timeslotpkg.ErrEndTimeIsRequired,
			timeslotpkg.ErrInvalidDayOfWeek,
			timeslotpkg.ErrInvalidTimeFormat,
			timeslotpkg.ErrInvalidTimeRange,
			timeslotpkg.ErrPreSessionBufferNegative,
			timeslotpkg.ErrPostSessionBufferTooLow:
			rw.WriteBadRequest(err.Error())
		case timeslotpkg.ErrTherapistNotFound,
			timeslotpkg.ErrTimeslotNotFound,
			timeslotpkg.ErrTimeslotNotOwned:
			rw.WriteNotFound(err.Error())
		case timeslotpkg.ErrOverlappingTimeslot:
			rw.WriteError(err, http.StatusConflict)
		default:
			rw.WriteError(err, http.StatusInternalServerError)
		}
		return
	}

	if err := rw.WriteJSON(timeslot, http.StatusOK); err != nil {
		rw.WriteError(err, http.StatusInternalServerError)
	}
}

func (h *TimeslotHandler) handleDeleteTimeslot(w http.ResponseWriter, r *http.Request) {
	rw := NewResponseWriter(w)

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
		case timeslotpkg.ErrTherapistIDIsRequired,
			timeslotpkg.ErrTimeslotIDIsRequired:
			rw.WriteBadRequest(err.Error())
		case timeslotpkg.ErrTherapistNotFound,
			timeslotpkg.ErrTimeslotNotFound,
			timeslotpkg.ErrTimeslotNotOwned:
			rw.WriteNotFound(err.Error())
		case timeslotpkg.ErrTimeslotHasActiveBookings:
			rw.WriteError(err, http.StatusConflict)
		default:
			rw.WriteError(err, http.StatusInternalServerError)
		}
		return
	}

	// Return 204 No Content for successful deletion
	w.WriteHeader(http.StatusNoContent)
}
