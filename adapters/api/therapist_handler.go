package api

import (
	"encoding/json"
	"net/http"

	"github.com/mishkahtherapy/brain/core/domain"
	"github.com/mishkahtherapy/brain/core/usecases/therapist/get_all_therapists"
	"github.com/mishkahtherapy/brain/core/usecases/therapist/get_therapist"
	"github.com/mishkahtherapy/brain/core/usecases/therapist/new_therapist"
	"github.com/mishkahtherapy/brain/core/usecases/therapist/update_therapist_info"
	"github.com/mishkahtherapy/brain/core/usecases/therapist/update_therapist_specializations"
)

type TherapistHandler struct {
	createTherapistUsecase                new_therapist.Usecase
	getAllTherapistsUsecase               get_all_therapists.Usecase
	getTherapistUsecase                   get_therapist.Usecase
	updateTherapistInfoUsecase            update_therapist_info.Usecase
	updateTherapistSpecializationsUsecase update_therapist_specializations.Usecase
}

func NewTherapistHandler(
	createUsecase new_therapist.Usecase,
	getAllUsecase get_all_therapists.Usecase,
	getUsecase get_therapist.Usecase,
	updateInfoUsecase update_therapist_info.Usecase,
	updateSpecializationsUsecase update_therapist_specializations.Usecase,
) *TherapistHandler {
	return &TherapistHandler{
		createTherapistUsecase:                createUsecase,
		getAllTherapistsUsecase:               getAllUsecase,
		getTherapistUsecase:                   getUsecase,
		updateTherapistInfoUsecase:            updateInfoUsecase,
		updateTherapistSpecializationsUsecase: updateSpecializationsUsecase,
	}
}

// SetUsecases sets the usecases for the handler (used for testing)
func (h *TherapistHandler) SetUsecases(
	createUsecase new_therapist.Usecase,
	getAllUsecase get_all_therapists.Usecase,
	getUsecase get_therapist.Usecase,
	updateInfoUsecase update_therapist_info.Usecase,
	updateSpecializationsUsecase update_therapist_specializations.Usecase,
) {
	h.createTherapistUsecase = createUsecase
	h.getAllTherapistsUsecase = getAllUsecase
	h.getTherapistUsecase = getUsecase
	h.updateTherapistInfoUsecase = updateInfoUsecase
	h.updateTherapistSpecializationsUsecase = updateSpecializationsUsecase
}

func (h *TherapistHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/v1/therapists", h.handleCreateTherapist)
	mux.HandleFunc("GET /api/v1/therapists", h.handleGetAllTherapists)
	mux.HandleFunc("GET /api/v1/therapists/{id}", h.handleGetTherapist)
	mux.HandleFunc("PUT /api/v1/therapists/{id}", h.handleUpdateTherapistInfo)
	mux.HandleFunc("PUT /api/v1/therapists/{id}/specializations", h.handleUpdateTherapistSpecializations)
}

func (h *TherapistHandler) handleCreateTherapist(w http.ResponseWriter, r *http.Request) {
	rw := NewResponseWriter(w)

	var input new_therapist.Input
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		rw.WriteBadRequest(err.Error())
		return
	}

	therapist, err := h.createTherapistUsecase.Execute(input)
	if err != nil {
		// Handle specific business logic errors
		switch err {
		case domain.ErrTherapistAlreadyExists:
			rw.WriteError(err, http.StatusConflict)
		case domain.ErrTherapistEmailRequired,
			domain.ErrTherapistNameRequired,
			domain.ErrTherapistPhoneRequired,
			domain.ErrTherapistWhatsAppRequired,
			domain.ErrTherapistInvalidPhone,
			domain.ErrTherapistInvalidWhatsApp:
			rw.WriteBadRequest(err.Error())
		default:
			rw.WriteError(err, http.StatusInternalServerError)
		}
		return
	}

	if err := rw.WriteJSON(therapist, http.StatusCreated); err != nil {
		rw.WriteError(err, http.StatusInternalServerError)
	}
}

func (h *TherapistHandler) handleGetAllTherapists(w http.ResponseWriter, r *http.Request) {
	rw := NewResponseWriter(w)

	therapists, err := h.getAllTherapistsUsecase.Execute()
	if err != nil {
		rw.WriteError(err, http.StatusInternalServerError)
		return
	}

	if err := rw.WriteJSON(therapists, http.StatusOK); err != nil {
		rw.WriteError(err, http.StatusInternalServerError)
	}
}

func (h *TherapistHandler) handleGetTherapist(w http.ResponseWriter, r *http.Request) {
	rw := NewResponseWriter(w)

	// read id from path
	id := domain.TherapistID(r.PathValue("id"))
	if id == "" {
		rw.WriteBadRequest("Missing therapist ID")
		return
	}

	therapist, err := h.getTherapistUsecase.Execute(id)
	if err != nil {
		if err == get_therapist.ErrTherapistNotFound {
			rw.WriteNotFound(err.Error())
			return
		}
		rw.WriteError(err, http.StatusInternalServerError)
		return
	}
	if therapist == nil {
		rw.WriteNotFound("Therapist not found")
		return
	}

	if err := rw.WriteJSON(therapist, http.StatusOK); err != nil {
		rw.WriteError(err, http.StatusInternalServerError)
	}
}

func (h *TherapistHandler) handleUpdateTherapistSpecializations(w http.ResponseWriter, r *http.Request) {
	rw := NewResponseWriter(w)

	// read id from path
	therapistID := domain.TherapistID(r.PathValue("id"))
	if therapistID == "" {
		rw.WriteBadRequest("Missing therapist ID")
		return
	}

	// Parse request body for specialization IDs
	var requestBody struct {
		SpecializationIDs []domain.SpecializationID `json:"specializationIds"`
	}
	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
		rw.WriteBadRequest(err.Error())
		return
	}

	// Create input for usecase
	input := update_therapist_specializations.Input{
		TherapistID:       therapistID,
		SpecializationIDs: requestBody.SpecializationIDs,
	}

	therapist, err := h.updateTherapistSpecializationsUsecase.Execute(input)
	if err != nil {
		// Handle specific business logic errors
		switch err {
		case update_therapist_specializations.ErrTherapistNotFound:
			rw.WriteNotFound(err.Error())
		case update_therapist_specializations.ErrSpecializationNotFound:
			rw.WriteBadRequest(err.Error())
		default:
			rw.WriteError(err, http.StatusInternalServerError)
		}
		return
	}

	if err := rw.WriteJSON(therapist, http.StatusOK); err != nil {
		rw.WriteError(err, http.StatusInternalServerError)
	}
}

func (h *TherapistHandler) handleUpdateTherapistInfo(w http.ResponseWriter, r *http.Request) {
	rw := NewResponseWriter(w)

	// Read therapist ID from path
	therapistID := domain.TherapistID(r.PathValue("id"))
	if therapistID == "" {
		rw.WriteBadRequest("Missing therapist ID")
		return
	}

	// Parse request body
	var requestBody struct {
		Name           string                `json:"name"`
		Email          domain.Email          `json:"email"`
		PhoneNumber    domain.PhoneNumber    `json:"phoneNumber"`
		WhatsAppNumber domain.WhatsAppNumber `json:"whatsAppNumber"`
		SpeaksEnglish  bool                  `json:"speaksEnglish"`
	}

	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
		rw.WriteBadRequest(err.Error())
		return
	}

	// Create input for usecase
	input := update_therapist_info.Input{
		TherapistID:    therapistID,
		Name:           requestBody.Name,
		Email:          requestBody.Email,
		PhoneNumber:    requestBody.PhoneNumber,
		WhatsAppNumber: requestBody.WhatsAppNumber,
		SpeaksEnglish:  requestBody.SpeaksEnglish,
	}

	therapist, err := h.updateTherapistInfoUsecase.Execute(input)
	if err != nil {
		// Handle specific business logic errors
		switch err {
		case domain.ErrTherapistIDRequired,
			domain.ErrTherapistNameRequired,
			domain.ErrTherapistEmailRequired,
			domain.ErrTherapistPhoneRequired,
			domain.ErrTherapistWhatsAppRequired,
			domain.ErrTherapistInvalidPhone,
			domain.ErrTherapistInvalidWhatsApp:
			rw.WriteBadRequest(err.Error())
		case domain.ErrTherapistNotFound:
			rw.WriteNotFound(err.Error())
		case domain.ErrTherapistEmailExists,
			domain.ErrTherapistWhatsAppExists:
			rw.WriteError(err, http.StatusConflict)
		default:
			rw.WriteError(err, http.StatusInternalServerError)
		}
		return
	}

	if err := rw.WriteJSON(therapist, http.StatusOK); err != nil {
		rw.WriteError(err, http.StatusInternalServerError)
	}
}
