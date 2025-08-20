package test

import (
	"encoding/json"
	"net/http"

	"github.com/mishkahtherapy/brain/adapters/api"
	"github.com/mishkahtherapy/brain/core/domain"
	"github.com/mishkahtherapy/brain/core/ports"
)

type TestHandler struct {
	notificationPort ports.NotificationPort
	notificationRepo ports.NotificationRepository
}

func NewTestHandler(notificationPort ports.NotificationPort, notificationRepo ports.NotificationRepository) *TestHandler {
	return &TestHandler{notificationPort: notificationPort, notificationRepo: notificationRepo}
}

func (h *TestHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/v1/test/notification", h.handleTestNotification)
}

func (h *TestHandler) handleTestNotification(w http.ResponseWriter, r *http.Request) {
	rw := api.NewResponseWriter(w)

	var requestBody struct {
		DeviceID     domain.DeviceID    `json:"deviceId"`
		Notification ports.Notification `json:"notification"`
	}

	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
		rw.WriteBadRequest(err.Error())
		return
	}

	_, err := h.notificationPort.SendNotification(requestBody.DeviceID, requestBody.Notification)
	if err != nil {
		rw.WriteError(err, http.StatusInternalServerError)
		return
	}

	rw.WriteNoContent()
}
