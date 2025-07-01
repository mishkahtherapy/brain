package api

import (
	"net/http"

	"github.com/mishkahtherapy/brain/core/domain"
	"github.com/mishkahtherapy/brain/core/usecases/common"
	"github.com/mishkahtherapy/brain/core/usecases/session/get_meeting_link"
)

// MeetingLinkProxyHandler handles redirecting to meeting URLs
type MeetingLinkProxyHandler struct {
	getMeetingLinkUsecase get_meeting_link.Usecase
}

// NewMeetingLinkProxyHandler creates a new instance of the MeetingLinkProxyHandler
func NewMeetingLinkProxyHandler(getMeetingLinkUsecase get_meeting_link.Usecase) *MeetingLinkProxyHandler {
	return &MeetingLinkProxyHandler{
		getMeetingLinkUsecase: getMeetingLinkUsecase,
	}
}

// SetUsecases sets the usecases for the handler (used for testing)
func (h *MeetingLinkProxyHandler) SetUsecases(getMeetingLinkUsecase get_meeting_link.Usecase) {
	h.getMeetingLinkUsecase = getMeetingLinkUsecase
}

// RegisterRoutes registers all the routes handled by the MeetingLinkProxyHandler
func (h *MeetingLinkProxyHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/v1/sessions/{id}/meeting", h.handleGetMeetingLink)
}

// handleGetMeetingLink redirects to the meeting URL for a session
// This handler supports safe redirection to video conferencing links
func (h *MeetingLinkProxyHandler) handleGetMeetingLink(w http.ResponseWriter, r *http.Request) {
	rw := NewResponseWriter(w)

	// Read session id from path
	id := domain.SessionID(r.PathValue("id"))
	if id == "" {
		rw.WriteBadRequest("Missing session ID")
		return
	}

	input := get_meeting_link.Input{
		SessionID: id,
	}

	output, err := h.getMeetingLinkUsecase.Execute(input)
	if err != nil {
		switch err {
		case common.ErrSessionIDIsRequired:
			rw.WriteBadRequest(err.Error())
		case common.ErrSessionNotFound:
			rw.WriteNotFound(err.Error())
		case common.ErrMeetingURLNotSet:
			rw.WriteNotFound(err.Error())
		default:
			rw.WriteError(err, http.StatusInternalServerError)
		}
		return
	}

	// Redirect to the meeting URL
	http.Redirect(w, r, output.MeetingURL, http.StatusTemporaryRedirect)
}
