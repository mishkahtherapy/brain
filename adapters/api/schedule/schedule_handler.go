package schedule_handler

import (
	"net/http"
	"strings"
	"time"

	"github.com/mishkahtherapy/brain/adapters/api"
	"github.com/mishkahtherapy/brain/core/domain"
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

	// Parse specializationsParam parameter (required)
	specializationsParam := r.URL.Query().Get("specialization")
	therapistIdsParam := r.URL.Query().Get("therapistIds")

	if specializationsParam == "" && therapistIdsParam == "" {
		rw.WriteBadRequest("specialization or therapistIds is required")
		return
	}

	if specializationsParam != "" && therapistIdsParam != "" {
		rw.WriteBadRequest("specialization and therapistIds cannot be used together")
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

	therapistIds := []domain.TherapistID{}
	if therapistIdsParam != "" {
		therapistIdStrings := strings.Split(strings.TrimSpace(therapistIdsParam), ",")
		for _, id := range therapistIdStrings {
			therapistIds = append(therapistIds, domain.TherapistID(id))
		}
	}

	specializations := []string{}
	if specializationsParam != "" {
		specializationStrings := strings.Split(strings.TrimSpace(specializationsParam), ",")
		specializations = append(specializations, specializationStrings...)
	}

	// Create input for usecase
	input := get_schedule.Input{
		MustSpeakEnglish: english,
		StartDate:        startDate,
		EndDate:          endDate,
	}

	if len(specializations) > 0 {
		input.SpecializationTag = specializations[0]
	}

	if len(therapistIds) > 0 {
		input.TherapistIDs = therapistIds
	}

	// Execute usecase
	schedule, err := h.getScheduleUsecase.Execute(input)
	if err != nil {
		// Handle specific business logic errors
		switch err {
		case get_schedule.ErrSpecializationTagOrTherapistIDsIsRequired:
			rw.WriteBadRequest(err.Error())
		case get_schedule.ErrSpecializationTagAndTherapistIDsCannotBeUsedTogether:
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
