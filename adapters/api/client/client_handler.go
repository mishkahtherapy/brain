package client_handler

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/mishkahtherapy/brain/adapters/api"
	"github.com/mishkahtherapy/brain/core/domain"
	"github.com/mishkahtherapy/brain/core/usecases/client/create_client"
	"github.com/mishkahtherapy/brain/core/usecases/client/get_all_clients"
	"github.com/mishkahtherapy/brain/core/usecases/client/get_client"
	"github.com/mishkahtherapy/brain/core/usecases/common"
)

type ClientHandler struct {
	createClientUsecase  create_client.Usecase
	getClientUsecase     get_client.Usecase
	getAllClientsUsecase get_all_clients.Usecase
}

func NewClientHandler(
	createUsecase create_client.Usecase,
	getAllUsecase get_all_clients.Usecase,
	getUsecase get_client.Usecase,
) *ClientHandler {
	return &ClientHandler{
		createClientUsecase:  createUsecase,
		getClientUsecase:     getUsecase,
		getAllClientsUsecase: getAllUsecase,
	}
}

// SetUsecases sets the usecases for the handler (used for testing)
func (h *ClientHandler) SetUsecases(
	createUsecase create_client.Usecase,
	getAllUsecase get_all_clients.Usecase,
	getUsecase get_client.Usecase,
) {
	h.createClientUsecase = createUsecase
	h.getClientUsecase = getUsecase
	h.getAllClientsUsecase = getAllUsecase
}

func (h *ClientHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/v1/clients", h.handleCreateClient)
	mux.HandleFunc("GET /api/v1/clients", h.handleGetAllClients)
	mux.HandleFunc("GET /api/v1/clients/{id}", h.handleGetClient)
}

func (h *ClientHandler) handleCreateClient(w http.ResponseWriter, r *http.Request) {
	rw := api.NewResponseWriter(w)

	var input create_client.Input
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		rw.WriteBadRequest(err.Error())
		return
	}

	client, err := h.createClientUsecase.Execute(input)
	if err != nil {
		// Handle specific business logic errors
		switch err {
		case create_client.ErrWhatsAppNumberIsRequired,
			create_client.ErrInvalidWhatsAppNumber,
			create_client.ErrInvalidTimezoneOffset:
			rw.WriteBadRequest(err.Error())
		case create_client.ErrClientAlreadyExists:
			rw.WriteError(err, http.StatusConflict)
		default:
			rw.WriteError(err, http.StatusInternalServerError)
		}
		return
	}

	if err := rw.WriteJSON(client, http.StatusCreated); err != nil {
		rw.WriteError(err, http.StatusInternalServerError)
	}
}

func (h *ClientHandler) handleGetAllClients(w http.ResponseWriter, r *http.Request) {
	rw := api.NewResponseWriter(w)

	clients, err := h.getAllClientsUsecase.Execute()
	if err != nil {
		rw.WriteError(err, http.StatusInternalServerError)
		return
	}

	if err := rw.WriteJSON(clients, http.StatusOK); err != nil {
		rw.WriteError(err, http.StatusInternalServerError)
	}
}

func (h *ClientHandler) handleGetClient(w http.ResponseWriter, r *http.Request) {
	rw := api.NewResponseWriter(w)

	idQuery := strings.Split(r.URL.Query().Get("ids"), ",")
	if len(idQuery) == 0 {
		rw.WriteBadRequest("Missing client ID")
		return
	}

	ids := make([]domain.ClientID, len(idQuery))
	for i, id := range idQuery {
		ids[i] = domain.ClientID(id)
	}

	client, err := h.getClientUsecase.Execute(ids)
	if err != nil {
		if err == common.ErrClientNotFound {
			rw.WriteNotFound(err.Error())
			return
		}
		rw.WriteError(err, http.StatusInternalServerError)
		return
	}

	if err := rw.WriteJSON(client, http.StatusOK); err != nil {
		rw.WriteError(err, http.StatusInternalServerError)
	}
}
