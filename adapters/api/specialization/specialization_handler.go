package specialization_handler

import (
	"encoding/json"
	"net/http"

	"github.com/mishkahtherapy/brain/adapters/api"
	"github.com/mishkahtherapy/brain/core/domain"
	"github.com/mishkahtherapy/brain/core/usecases/specialization/get_all_specializations"
	"github.com/mishkahtherapy/brain/core/usecases/specialization/get_specialization"
	"github.com/mishkahtherapy/brain/core/usecases/specialization/new_specialization"
)

type SpecializationHandler struct {
	createSpecializationUsecase  new_specialization.Usecase
	getAllSpecializationsUsecase get_all_specializations.Usecase
	getSpecializationUsecase     get_specialization.Usecase
}

func NewSpecializationHandler(
	createUsecase new_specialization.Usecase,
	getAllSpecializationsUsecase get_all_specializations.Usecase,
	getSpecializationUsecase get_specialization.Usecase,
) *SpecializationHandler {
	return &SpecializationHandler{
		createSpecializationUsecase:  createUsecase,
		getAllSpecializationsUsecase: getAllSpecializationsUsecase,
		getSpecializationUsecase:     getSpecializationUsecase,
	}
}

// SetUsecases sets the usecases for the handler (used for testing)
func (h *SpecializationHandler) SetUsecases(
	createUsecase new_specialization.Usecase,
	getAllUsecase get_all_specializations.Usecase,
	getUsecase get_specialization.Usecase,
) {
	h.createSpecializationUsecase = createUsecase
	h.getAllSpecializationsUsecase = getAllUsecase
	h.getSpecializationUsecase = getUsecase
}

func (h *SpecializationHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/v1/specializations", h.handleCreateSpecialization)
	mux.HandleFunc("GET /api/v1/specializations", h.handleGetAllSpecializations)
	mux.HandleFunc("GET /api/v1/specializations/{id}", h.handleGetSpecialization)
}

func (h *SpecializationHandler) handleCreateSpecialization(w http.ResponseWriter, r *http.Request) {
	rw := api.NewResponseWriter(w)

	var input new_specialization.Input
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		rw.WriteBadRequest(err.Error())
		return
	}

	specialization, err := h.createSpecializationUsecase.Execute(input)
	if err != nil {
		rw.WriteError(err, http.StatusInternalServerError)
		return
	}

	if err := rw.WriteJSON(specialization, http.StatusCreated); err != nil {
		rw.WriteError(err, http.StatusInternalServerError)
	}
}

func (h *SpecializationHandler) handleGetAllSpecializations(w http.ResponseWriter, r *http.Request) {
	rw := api.NewResponseWriter(w)

	specializations, err := h.getAllSpecializationsUsecase.Execute()
	if err != nil {
		rw.WriteError(err, http.StatusInternalServerError)
		return
	}

	if err := rw.WriteJSON(specializations, http.StatusOK); err != nil {
		rw.WriteError(err, http.StatusInternalServerError)
	}
}

func (h *SpecializationHandler) handleGetSpecialization(w http.ResponseWriter, r *http.Request) {
	rw := api.NewResponseWriter(w)

	// read id from path
	id := domain.SpecializationID(r.PathValue("id"))
	if id == "" {
		rw.WriteBadRequest("Missing specialization ID")
		return
	}

	specialization, err := h.getSpecializationUsecase.Execute(id)
	if err != nil {
		rw.WriteError(err, http.StatusInternalServerError)
		return
	}
	if specialization == nil {
		rw.WriteNotFound("Specialization not found")
		return
	}

	if err := rw.WriteJSON(specialization, http.StatusOK); err != nil {
		rw.WriteError(err, http.StatusInternalServerError)
	}
}
