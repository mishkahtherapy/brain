package api

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/mishkahtherapy/brain/core/domain"
	"github.com/mishkahtherapy/brain/core/usecases/session/create_session"
	"github.com/mishkahtherapy/brain/core/usecases/session/get_session"
	"github.com/mishkahtherapy/brain/core/usecases/session/list_sessions_admin"
	"github.com/mishkahtherapy/brain/core/usecases/session/list_sessions_by_client"
	"github.com/mishkahtherapy/brain/core/usecases/session/list_sessions_by_therapist"
	"github.com/mishkahtherapy/brain/core/usecases/session/update_meeting_url"
	"github.com/mishkahtherapy/brain/core/usecases/session/update_session_notes"
	"github.com/mishkahtherapy/brain/core/usecases/session/update_session_state"
)

type SessionHandler struct {
	createSessionUsecase           create_session.Usecase
	getSessionUsecase              get_session.Usecase
	updateSessionStateUsecase      update_session_state.Usecase
	updateSessionNotesUsecase      update_session_notes.Usecase
	updateMeetingURLUsecase        update_meeting_url.Usecase
	listSessionsByTherapistUsecase list_sessions_by_therapist.Usecase
	listSessionsByClientUsecase    list_sessions_by_client.Usecase
	listSessionsAdminUsecase       list_sessions_admin.Usecase
}

// NewSessionHandler creates a new instance of the SessionHandler
func NewSessionHandler(
	createUsecase create_session.Usecase,
	getUsecase get_session.Usecase,
	updateStateUsecase update_session_state.Usecase,
	updateNotesUsecase update_session_notes.Usecase,
	updateMeetingURLUsecase update_meeting_url.Usecase,
	listByTherapistUsecase list_sessions_by_therapist.Usecase,
	listByClientUsecase list_sessions_by_client.Usecase,
	listAdminUsecase list_sessions_admin.Usecase,
) *SessionHandler {
	return &SessionHandler{
		createSessionUsecase:           createUsecase,
		getSessionUsecase:              getUsecase,
		updateSessionStateUsecase:      updateStateUsecase,
		updateSessionNotesUsecase:      updateNotesUsecase,
		updateMeetingURLUsecase:        updateMeetingURLUsecase,
		listSessionsByTherapistUsecase: listByTherapistUsecase,
		listSessionsByClientUsecase:    listByClientUsecase,
		listSessionsAdminUsecase:       listAdminUsecase,
	}
}

// SetUsecases sets the usecases for the handler (used for testing)
func (h *SessionHandler) SetUsecases(
	createUsecase create_session.Usecase,
	getUsecase get_session.Usecase,
	updateStateUsecase update_session_state.Usecase,
	updateNotesUsecase update_session_notes.Usecase,
	updateMeetingURLUsecase update_meeting_url.Usecase,
	listByTherapistUsecase list_sessions_by_therapist.Usecase,
	listByClientUsecase list_sessions_by_client.Usecase,
	listAdminUsecase list_sessions_admin.Usecase,
) {
	h.createSessionUsecase = createUsecase
	h.getSessionUsecase = getUsecase
	h.updateSessionStateUsecase = updateStateUsecase
	h.updateSessionNotesUsecase = updateNotesUsecase
	h.updateMeetingURLUsecase = updateMeetingURLUsecase
	h.listSessionsByTherapistUsecase = listByTherapistUsecase
	h.listSessionsByClientUsecase = listByClientUsecase
	h.listSessionsAdminUsecase = listAdminUsecase
}

// RegisterRoutes registers all the routes handled by the SessionHandler
func (h *SessionHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/v1/sessions", h.handleCreateSession)
	mux.HandleFunc("GET /api/v1/sessions/{id}", h.handleGetSession)
	mux.HandleFunc("PUT /api/v1/sessions/{id}/state", h.handleUpdateSessionState)
	mux.HandleFunc("PUT /api/v1/sessions/{id}/notes", h.handleUpdateSessionNotes)
	mux.HandleFunc("PUT /api/v1/sessions/{id}/meeting-url", h.handleUpdateMeetingURL)
	mux.HandleFunc("GET /api/v1/therapists/{id}/sessions", h.handleListSessionsByTherapist)
	mux.HandleFunc("GET /api/v1/clients/{id}/sessions", h.handleListSessionsByClient)
	mux.HandleFunc("GET /api/v1/admin/sessions", h.handleListSessionsAdmin)
}

// handleCreateSession handles POST /api/v1/sessions
func (h *SessionHandler) handleCreateSession(w http.ResponseWriter, r *http.Request) {
	rw := NewResponseWriter(w)

	var input create_session.Input
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		rw.WriteBadRequest(err.Error())
		return
	}

	session, err := h.createSessionUsecase.Execute(input)
	if err != nil {
		// Handle specific business logic errors
		switch err {
		case create_session.ErrBookingIDIsRequired,
			create_session.ErrTherapistIDIsRequired,
			create_session.ErrClientIDIsRequired,
			create_session.ErrTimeSlotIDIsRequired,
			create_session.ErrStartTimeIsRequired,
			create_session.ErrPaidAmountIsRequired,
			create_session.ErrLanguageIsRequired:
			rw.WriteBadRequest(err.Error())
		case create_session.ErrBookingNotFound,
			create_session.ErrTherapistNotFound,
			create_session.ErrClientNotFound,
			create_session.ErrTimeSlotNotFound:
			rw.WriteNotFound(err.Error())
		default:
			rw.WriteError(err, http.StatusInternalServerError)
		}
		return
	}

	if err := rw.WriteJSON(session, http.StatusCreated); err != nil {
		rw.WriteError(err, http.StatusInternalServerError)
	}
}

// handleGetSession handles GET /api/v1/sessions/{id}
func (h *SessionHandler) handleGetSession(w http.ResponseWriter, r *http.Request) {
	rw := NewResponseWriter(w)

	// Read session id from path
	id := domain.SessionID(r.PathValue("id"))
	if id == "" {
		rw.WriteBadRequest("Missing session ID")
		return
	}

	session, err := h.getSessionUsecase.Execute(id)
	if err != nil {
		if err == get_session.ErrSessionNotFound {
			rw.WriteNotFound(err.Error())
			return
		}
		rw.WriteError(err, http.StatusInternalServerError)
		return
	}

	if err := rw.WriteJSON(session, http.StatusOK); err != nil {
		rw.WriteError(err, http.StatusInternalServerError)
	}
}

// handleUpdateSessionState handles PUT /api/v1/sessions/{id}/state
func (h *SessionHandler) handleUpdateSessionState(w http.ResponseWriter, r *http.Request) {
	rw := NewResponseWriter(w)

	// Read session id from path
	id := domain.SessionID(r.PathValue("id"))
	if id == "" {
		rw.WriteBadRequest("Missing session ID")
		return
	}

	// Parse request body to get new state
	var requestBody struct {
		NewState domain.SessionState `json:"newState"`
	}

	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
		rw.WriteBadRequest(err.Error())
		return
	}

	input := update_session_state.Input{
		SessionID: id,
		NewState:  requestBody.NewState,
	}

	session, err := h.updateSessionStateUsecase.Execute(input)
	if err != nil {
		switch err {
		case update_session_state.ErrSessionIDIsRequired,
			update_session_state.ErrStateIsRequired:
			rw.WriteBadRequest(err.Error())
		case update_session_state.ErrSessionNotFound:
			rw.WriteNotFound(err.Error())
		case update_session_state.ErrInvalidStateTransition:
			rw.WriteBadRequest(err.Error())
		default:
			rw.WriteError(err, http.StatusInternalServerError)
		}
		return
	}

	if err := rw.WriteJSON(session, http.StatusOK); err != nil {
		rw.WriteError(err, http.StatusInternalServerError)
	}
}

// handleUpdateSessionNotes handles PUT /api/v1/sessions/{id}/notes
func (h *SessionHandler) handleUpdateSessionNotes(w http.ResponseWriter, r *http.Request) {
	rw := NewResponseWriter(w)

	// Read session id from path
	id := domain.SessionID(r.PathValue("id"))
	if id == "" {
		rw.WriteBadRequest("Missing session ID")
		return
	}

	// Parse request body to get notes
	var requestBody struct {
		Notes string `json:"notes"`
	}

	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
		rw.WriteBadRequest(err.Error())
		return
	}

	input := update_session_notes.Input{
		SessionID: id,
		Notes:     requestBody.Notes,
	}

	session, err := h.updateSessionNotesUsecase.Execute(input)
	if err != nil {
		switch err {
		case update_session_notes.ErrSessionIDIsRequired,
			update_session_notes.ErrNotesIsRequired:
			rw.WriteBadRequest(err.Error())
		case update_session_notes.ErrSessionNotFound:
			rw.WriteNotFound(err.Error())
		default:
			rw.WriteError(err, http.StatusInternalServerError)
		}
		return
	}

	if err := rw.WriteJSON(session, http.StatusOK); err != nil {
		rw.WriteError(err, http.StatusInternalServerError)
	}
}

// handleUpdateMeetingURL handles PUT /api/v1/sessions/{id}/meeting-url
func (h *SessionHandler) handleUpdateMeetingURL(w http.ResponseWriter, r *http.Request) {
	rw := NewResponseWriter(w)

	// Read session id from path
	id := domain.SessionID(r.PathValue("id"))
	if id == "" {
		rw.WriteBadRequest("Missing session ID")
		return
	}

	// Parse request body to get meeting URL
	var requestBody struct {
		MeetingURL string `json:"meetingUrl"`
	}

	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
		rw.WriteBadRequest(err.Error())
		return
	}

	input := update_meeting_url.Input{
		SessionID:  id,
		MeetingURL: requestBody.MeetingURL,
	}

	session, err := h.updateMeetingURLUsecase.Execute(input)
	if err != nil {
		switch err {
		case update_meeting_url.ErrSessionIDIsRequired,
			update_meeting_url.ErrMeetingURLIsRequired,
			update_meeting_url.ErrInvalidMeetingURL:
			rw.WriteBadRequest(err.Error())
		case update_meeting_url.ErrSessionNotFound:
			rw.WriteNotFound(err.Error())
		default:
			rw.WriteError(err, http.StatusInternalServerError)
		}
		return
	}

	if err := rw.WriteJSON(session, http.StatusOK); err != nil {
		rw.WriteError(err, http.StatusInternalServerError)
	}
}

// handleListSessionsByTherapist handles GET /api/v1/therapists/{id}/sessions
func (h *SessionHandler) handleListSessionsByTherapist(w http.ResponseWriter, r *http.Request) {
	rw := NewResponseWriter(w)

	// Read therapist id from path
	therapistID := domain.TherapistID(r.PathValue("id"))
	if therapistID == "" {
		rw.WriteBadRequest("Missing therapist ID")
		return
	}

	input := list_sessions_by_therapist.Input{
		TherapistID: therapistID,
	}

	sessions, err := h.listSessionsByTherapistUsecase.Execute(input)
	if err != nil {
		switch err {
		case list_sessions_by_therapist.ErrTherapistIDIsRequired:
			rw.WriteBadRequest(err.Error())
		default:
			rw.WriteError(err, http.StatusInternalServerError)
		}
		return
	}

	if err := rw.WriteJSON(sessions, http.StatusOK); err != nil {
		rw.WriteError(err, http.StatusInternalServerError)
	}
}

// handleListSessionsByClient handles GET /api/v1/clients/{id}/sessions
func (h *SessionHandler) handleListSessionsByClient(w http.ResponseWriter, r *http.Request) {
	rw := NewResponseWriter(w)

	// Read client id from path
	clientID := domain.ClientID(r.PathValue("id"))
	if clientID == "" {
		rw.WriteBadRequest("Missing client ID")
		return
	}

	input := list_sessions_by_client.Input{
		ClientID: clientID,
	}

	sessions, err := h.listSessionsByClientUsecase.Execute(input)
	if err != nil {
		switch err {
		case list_sessions_by_client.ErrClientIDIsRequired:
			rw.WriteBadRequest(err.Error())
		default:
			rw.WriteError(err, http.StatusInternalServerError)
		}
		return
	}

	if err := rw.WriteJSON(sessions, http.StatusOK); err != nil {
		rw.WriteError(err, http.StatusInternalServerError)
	}
}

// handleListSessionsAdmin handles GET /api/v1/admin/sessions
func (h *SessionHandler) handleListSessionsAdmin(w http.ResponseWriter, r *http.Request) {
	rw := NewResponseWriter(w)

	// Parse start and end date query parameters
	startDateParam := r.URL.Query().Get("startDate")
	endDateParam := r.URL.Query().Get("endDate")

	var startDate, endDate time.Time
	var err error

	// Parse dates if provided
	if startDateParam != "" {
		startDate, err = time.Parse(time.RFC3339, startDateParam)
		if err != nil {
			rw.WriteBadRequest("Invalid startDate format. Use RFC3339 format (e.g., 2025-06-30T00:00:00Z)")
			return
		}
	}

	if endDateParam != "" {
		endDate, err = time.Parse(time.RFC3339, endDateParam)
		if err != nil {
			rw.WriteBadRequest("Invalid endDate format. Use RFC3339 format (e.g., 2025-06-30T00:00:00Z)")
			return
		}
	}

	input := list_sessions_admin.Input{
		StartDate: startDate,
		EndDate:   endDate,
	}

	sessions, err := h.listSessionsAdminUsecase.Execute(input)
	if err != nil {
		switch err {
		case list_sessions_admin.ErrInvalidDateRange:
			rw.WriteBadRequest(err.Error())
		default:
			rw.WriteError(err, http.StatusInternalServerError)
		}
		return
	}

	if err := rw.WriteJSON(sessions, http.StatusOK); err != nil {
		rw.WriteError(err, http.StatusInternalServerError)
	}
}
