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

	// Parse tag parameter (required)
	tag := r.URL.Query().Get("tag")
	if tag == "" {
		rw.WriteBadRequest("specialization tag is required")
		return
	}

	// Parse english parameter (optional)
	english := false
	englishParam := r.URL.Query().Get("english")
	if englishParam == "true" {
		english = true
	}

	// Parse start_date parameter (optional)
	var startDate time.Time
	startDateParam := r.URL.Query().Get("start_date")
	if startDateParam != "" {
		var err error
		startDate, err = time.Parse("2006-01-02", startDateParam)
		if err != nil {
			rw.WriteBadRequest("invalid start_date format: use YYYY-MM-DD")
			return
		}
	}

	// Parse end_date parameter (optional)
	var endDate time.Time
	endDateParam := r.URL.Query().Get("end_date")
	if endDateParam != "" {
		var err error
		endDate, err = time.Parse("2006-01-02", endDateParam)
		if err != nil {
			rw.WriteBadRequest("invalid end_date format: use YYYY-MM-DD")
			return
		}
	}

	// Validate date range if both are provided
	if !startDate.IsZero() && !endDate.IsZero() && endDate.Before(startDate) {
		rw.WriteBadRequest("end_date must be after start_date")
		return
	}

	// Create input for usecase
	input := get_schedule.Input{
		SpecializationTag: tag,
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
