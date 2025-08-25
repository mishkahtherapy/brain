package api

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/mishkahtherapy/brain/core/domain"
	"github.com/mishkahtherapy/brain/core/usecases/common"
	"github.com/mishkahtherapy/brain/core/usecases/session/get_session"
	"github.com/mishkahtherapy/brain/core/usecases/session/list_sessions_admin"
	"github.com/mishkahtherapy/brain/core/usecases/session/list_sessions_by_client"
	"github.com/mishkahtherapy/brain/core/usecases/session/list_sessions_by_therapist"
	"github.com/mishkahtherapy/brain/core/usecases/session/update_meeting_url"
	"github.com/mishkahtherapy/brain/core/usecases/session/update_session_notes"
	"github.com/mishkahtherapy/brain/core/usecases/session/update_session_state"
)

const defaultSessionDuration = 60

type SessionHandler struct {
	// createSessionUsecase           create_session.Usecase
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
	// createUsecase create_session.Usecase,
	getUsecase get_session.Usecase,
	updateStateUsecase update_session_state.Usecase,
	updateNotesUsecase update_session_notes.Usecase,
	updateMeetingURLUsecase update_meeting_url.Usecase,
	listByTherapistUsecase list_sessions_by_therapist.Usecase,
	listByClientUsecase list_sessions_by_client.Usecase,
	listAdminUsecase list_sessions_admin.Usecase,
) *SessionHandler {
	return &SessionHandler{
		// createSessionUsecase:           createUsecase,
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
	// createUsecase create_session.Usecase,
	getUsecase get_session.Usecase,
	updateStateUsecase update_session_state.Usecase,
	updateNotesUsecase update_session_notes.Usecase,
	updateMeetingURLUsecase update_meeting_url.Usecase,
	listByTherapistUsecase list_sessions_by_therapist.Usecase,
	listByClientUsecase list_sessions_by_client.Usecase,
	listAdminUsecase list_sessions_admin.Usecase,
) {
	// h.createSessionUsecase = createUsecase
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
	mux.HandleFunc("GET /api/v1/sessions/{id}", h.handleGetSession)
	mux.HandleFunc("PUT /api/v1/sessions/{id}/state", h.handleUpdateSessionState)
	mux.HandleFunc("PUT /api/v1/sessions/{id}/notes", h.handleUpdateSessionNotes)
	mux.HandleFunc("PUT /api/v1/sessions/{id}/meeting-url", h.handleUpdateMeetingURL)
	mux.HandleFunc("GET /api/v1/therapists/{id}/sessions", h.handleListSessionsByTherapist)
	mux.HandleFunc("GET /api/v1/clients/{id}/sessions", h.handleListSessionsByClient)
	mux.HandleFunc("GET /api/v1/admin/sessions", h.handleListSessionsAdmin)
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
		if err == common.ErrSessionNotFound {
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
		case common.ErrSessionIDIsRequired,
			common.ErrStateIsRequired:
			rw.WriteBadRequest(err.Error())
		case common.ErrSessionNotFound:
			rw.WriteNotFound(err.Error())
		case common.ErrInvalidStateTransition:
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
		case common.ErrSessionIDIsRequired,
			common.ErrNotesIsRequired:
			rw.WriteBadRequest(err.Error())
		case common.ErrSessionNotFound:
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
		case common.ErrSessionIDIsRequired,
			common.ErrMeetingURLIsRequired:
			rw.WriteBadRequest(err.Error())
		case common.ErrSessionNotFound:
			rw.WriteNotFound(err.Error())
		case common.ErrInvalidMeetingURL:
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
		case common.ErrTherapistIDIsRequired:
			rw.WriteBadRequest(err.Error())
		default:
			rw.WriteError(err, http.StatusInternalServerError)
		}
		return
	}

	// TODO: Remove this once we have a proper duration implementation
	for _, session := range sessions {
		session.Duration = defaultSessionDuration
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
		case common.ErrClientIDIsRequired:
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

	// Parse query parameters for date range
	var input list_sessions_admin.Input

	if startDateParam := r.URL.Query().Get("startDate"); startDateParam != "" {
		if startDate, err := time.Parse("2006-01-02", startDateParam); err != nil {
			rw.WriteBadRequest("Invalid startDate format. Use YYYY-MM-DD")
			return
		} else {
			input.StartDate = startDate
		}
	}

	if endDateParam := r.URL.Query().Get("endDate"); endDateParam != "" {
		if endDate, err := time.Parse("2006-01-02", endDateParam); err != nil {
			rw.WriteBadRequest("Invalid endDate format. Use YYYY-MM-DD")
			return
		} else {
			input.EndDate = endDate
		}
	}

	sessions, err := h.listSessionsAdminUsecase.Execute(input)
	if err != nil {
		switch err {
		case common.ErrInvalidDateRange:
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
