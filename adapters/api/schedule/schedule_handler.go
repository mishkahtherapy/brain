package schedule_handler

import (
	"net/http"
	"time"

	"github.com/mishkahtherapy/brain/adapters/api"
	"github.com/mishkahtherapy/brain/core/usecases/schedule/get_schedule"
)

type ScheduleHandler struct {
	getScheduleUsecase get_schedule.Usecase
}

func NewScheduleHandler(getScheduleUsecase get_schedule.Usecase) *ScheduleHandler {
	return &ScheduleHandler{
		getScheduleUsecase: getScheduleUsecase,
	}
}

func (h *ScheduleHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/v1/schedule", h.handleGetSchedule)
}

func (h *ScheduleHandler) handleGetSchedule(w http.ResponseWriter, r *http.Request) {
	rw := api.NewResponseWriter(w)

	// Parse specialization parameter (required)
	specialization := r.URL.Query().Get("specialization")
	if specialization == "" {
		rw.WriteBadRequest("specialization tag is required")
		return
	}

	// Parse english parameter (optional)
	english := false
	englishParam := r.URL.Query().Get("requiresEnglish")
	if englishParam == "true" {
		english = true
	}

	// Parse startDate parameter (optional)
	var startDate time.Time
	startDateParam := r.URL.Query().Get("startDate")
	if startDateParam != "" {
		var err error
		startDate, err = time.Parse("2006-01-02", startDateParam)
		if err != nil {
			rw.WriteBadRequest("invalid startDate format: use YYYY-MM-DD")
			return
		}
	}

	// Parse endDate parameter (optional)
	var endDate time.Time
	endDateParam := r.URL.Query().Get("endDate")
	if endDateParam != "" {
		var err error
		endDate, err = time.Parse("2006-01-02", endDateParam)
		if err != nil {
			rw.WriteBadRequest("invalid endDate format: use YYYY-MM-DD")
			return
		}
	}

	// Validate date range if both are provided
	if !startDate.IsZero() && !endDate.IsZero() && endDate.Before(startDate) {
		rw.WriteBadRequest("endDate must be after startDate")
		return
	}

	// Create input for usecase
	input := get_schedule.Input{
		SpecializationTag: specialization,
		MustSpeakEnglish:  english,
		StartDate:         startDate,
		EndDate:           endDate,
	}

	// Execute usecase
	schedule, err := h.getScheduleUsecase.Execute(input)
	if err != nil {
		// Handle specific business logic errors
		switch err {
		case get_schedule.ErrSpecializationTagIsRequired:
			rw.WriteBadRequest(err.Error())
		case get_schedule.ErrInvalidDateRange:
			rw.WriteBadRequest(err.Error())
		default:
			rw.WriteError(err, http.StatusInternalServerError)
		}
		return
	}

	// Return response
	if err := rw.WriteJSON(schedule, http.StatusOK); err != nil {
		rw.WriteError(err, http.StatusInternalServerError)
	}
}
