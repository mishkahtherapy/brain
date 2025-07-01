package api

import (
	"encoding/json"
	"net/http"

	"github.com/mishkahtherapy/brain/core/domain"
	"github.com/mishkahtherapy/brain/core/domain/booking"
	"github.com/mishkahtherapy/brain/core/usecases/booking/cancel_booking"
	"github.com/mishkahtherapy/brain/core/usecases/booking/confirm_booking"
	"github.com/mishkahtherapy/brain/core/usecases/booking/create_booking"
	"github.com/mishkahtherapy/brain/core/usecases/booking/get_booking"
	"github.com/mishkahtherapy/brain/core/usecases/booking/list_bookings_by_client"
	"github.com/mishkahtherapy/brain/core/usecases/booking/list_bookings_by_therapist"
	"github.com/mishkahtherapy/brain/core/usecases/common"
)

type BookingHandler struct {
	createBookingUsecase           create_booking.Usecase
	getBookingUsecase              get_booking.Usecase
	confirmBookingUsecase          confirm_booking.Usecase
	cancelBookingUsecase           cancel_booking.Usecase
	listBookingsByTherapistUsecase list_bookings_by_therapist.Usecase
	listBookingsByClientUsecase    list_bookings_by_client.Usecase
}

func NewBookingHandler(
	createUsecase create_booking.Usecase,
	getUsecase get_booking.Usecase,
	confirmUsecase confirm_booking.Usecase,
	cancelUsecase cancel_booking.Usecase,
	listByTherapistUsecase list_bookings_by_therapist.Usecase,
	listByClientUsecase list_bookings_by_client.Usecase,
) *BookingHandler {
	return &BookingHandler{
		createBookingUsecase:           createUsecase,
		getBookingUsecase:              getUsecase,
		confirmBookingUsecase:          confirmUsecase,
		cancelBookingUsecase:           cancelUsecase,
		listBookingsByTherapistUsecase: listByTherapistUsecase,
		listBookingsByClientUsecase:    listByClientUsecase,
	}
}

// SetUsecases sets the usecases for the handler (used for testing)
func (h *BookingHandler) SetUsecases(
	createUsecase create_booking.Usecase,
	getUsecase get_booking.Usecase,
	confirmUsecase confirm_booking.Usecase,
	cancelUsecase cancel_booking.Usecase,
	listByTherapistUsecase list_bookings_by_therapist.Usecase,
	listByClientUsecase list_bookings_by_client.Usecase,
) {
	h.createBookingUsecase = createUsecase
	h.getBookingUsecase = getUsecase
	h.confirmBookingUsecase = confirmUsecase
	h.cancelBookingUsecase = cancelUsecase
	h.listBookingsByTherapistUsecase = listByTherapistUsecase
	h.listBookingsByClientUsecase = listByClientUsecase
}

func (h *BookingHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/v1/bookings", h.handleCreateBooking)
	mux.HandleFunc("GET /api/v1/bookings/{id}", h.handleGetBooking)
	mux.HandleFunc("PUT /api/v1/bookings/{id}/confirm", h.handleConfirmBooking)
	mux.HandleFunc("PUT /api/v1/bookings/{id}/cancel", h.handleCancelBooking)
	mux.HandleFunc("GET /api/v1/therapists/{id}/bookings", h.handleListBookingsByTherapist)
	mux.HandleFunc("GET /api/v1/clients/{id}/bookings", h.handleListBookingsByClient)
}

func (h *BookingHandler) handleCreateBooking(w http.ResponseWriter, r *http.Request) {
	rw := NewResponseWriter(w)

	var input create_booking.Input
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		rw.WriteBadRequest(err.Error())
		return
	}

	booking, err := h.createBookingUsecase.Execute(input)
	if err != nil {
		// Handle specific business logic errors
		switch err {
		case common.ErrTherapistIDIsRequired,
			common.ErrClientIDIsRequired,
			common.ErrTimeSlotIDIsRequired,
			common.ErrStartTimeIsRequired:
			rw.WriteBadRequest(err.Error())
		case common.ErrTherapistNotFound,
			common.ErrClientNotFound,
			common.ErrTimeSlotNotFound:
			rw.WriteNotFound(err.Error())
		case common.ErrTimeSlotAlreadyBooked:
			rw.WriteError(err, http.StatusConflict)
		default:
			rw.WriteError(err, http.StatusInternalServerError)
		}
		return
	}

	if err := rw.WriteJSON(booking, http.StatusCreated); err != nil {
		rw.WriteError(err, http.StatusInternalServerError)
	}
}

func (h *BookingHandler) handleGetBooking(w http.ResponseWriter, r *http.Request) {
	rw := NewResponseWriter(w)

	// Read id from path
	id := domain.BookingID(r.PathValue("id"))
	if id == "" {
		rw.WriteBadRequest("Missing booking ID")
		return
	}

	booking, err := h.getBookingUsecase.Execute(id)
	if err != nil {
		if err == common.ErrBookingNotFound {
			rw.WriteNotFound(err.Error())
			return
		}
		rw.WriteError(err, http.StatusInternalServerError)
		return
	}

	if err := rw.WriteJSON(booking, http.StatusOK); err != nil {
		rw.WriteError(err, http.StatusInternalServerError)
	}
}

func (h *BookingHandler) handleConfirmBooking(w http.ResponseWriter, r *http.Request) {
	rw := NewResponseWriter(w)

	// Read id from path
	id := domain.BookingID(r.PathValue("id"))
	if id == "" {
		rw.WriteBadRequest("Missing booking ID")
		return
	}

	// Parse request body to get paid amount and language
	var requestBody struct {
		PaidAmount int                    `json:"paidAmount"`
		Language   domain.SessionLanguage `json:"language"`
	}

	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
		rw.WriteBadRequest(err.Error())
		return
	}

	input := confirm_booking.Input{
		BookingID:  id,
		PaidAmount: requestBody.PaidAmount,
		Language:   requestBody.Language,
	}

	booking, err := h.confirmBookingUsecase.Execute(input)
	if err != nil {
		// Handle specific business logic errors
		switch err {
		case common.ErrBookingIDIsRequired,
			common.ErrPaidAmountIsRequired,
			common.ErrLanguageIsRequired:
			rw.WriteBadRequest(err.Error())
		case common.ErrBookingNotFound:
			rw.WriteNotFound(err.Error())
		case common.ErrInvalidBookingState:
			rw.WriteBadRequest(err.Error())
		case confirm_booking.ErrFailedToCreateSession:
			rw.WriteError(err, http.StatusInternalServerError)
		default:
			rw.WriteError(err, http.StatusInternalServerError)
		}
		return
	}

	if err := rw.WriteJSON(booking, http.StatusOK); err != nil {
		rw.WriteError(err, http.StatusInternalServerError)
	}
}

func (h *BookingHandler) handleCancelBooking(w http.ResponseWriter, r *http.Request) {
	rw := NewResponseWriter(w)

	// Read id from path
	id := domain.BookingID(r.PathValue("id"))
	if id == "" {
		rw.WriteBadRequest("Missing booking ID")
		return
	}

	input := cancel_booking.Input{
		BookingID: id,
	}

	booking, err := h.cancelBookingUsecase.Execute(input)
	if err != nil {
		// Handle specific business logic errors
		switch err {
		case common.ErrBookingIDIsRequired:
			rw.WriteBadRequest(err.Error())
		case common.ErrBookingNotFound:
			rw.WriteNotFound(err.Error())
		case common.ErrInvalidStateTransition:
			rw.WriteBadRequest(err.Error())
		default:
			rw.WriteError(err, http.StatusInternalServerError)
		}
		return
	}

	if err := rw.WriteJSON(booking, http.StatusOK); err != nil {
		rw.WriteError(err, http.StatusInternalServerError)
	}
}

func (h *BookingHandler) handleListBookingsByTherapist(w http.ResponseWriter, r *http.Request) {
	rw := NewResponseWriter(w)

	// Read therapist id from path
	therapistID := domain.TherapistID(r.PathValue("id"))
	if therapistID == "" {
		rw.WriteBadRequest("Missing therapist ID")
		return
	}

	// Parse optional state query parameter
	var state *booking.BookingState
	if stateParam := r.URL.Query().Get("state"); stateParam != "" {
		bookingState := booking.BookingState(stateParam)
		// Validate state value
		if bookingState != booking.BookingStatePending &&
			bookingState != booking.BookingStateConfirmed &&
			bookingState != booking.BookingStateCancelled {
			rw.WriteBadRequest("Invalid state parameter. Must be one of: pending, confirmed, cancelled")
			return
		}
		state = &bookingState
	}

	input := list_bookings_by_therapist.Input{
		TherapistID: therapistID,
		State:       state,
	}

	bookings, err := h.listBookingsByTherapistUsecase.Execute(input)
	if err != nil {
		switch err {
		case common.ErrTherapistIDIsRequired:
			rw.WriteBadRequest(err.Error())
		default:
			rw.WriteError(err, http.StatusInternalServerError)
		}
		return
	}

	if err := rw.WriteJSON(bookings, http.StatusOK); err != nil {
		rw.WriteError(err, http.StatusInternalServerError)
	}
}

func (h *BookingHandler) handleListBookingsByClient(w http.ResponseWriter, r *http.Request) {
	rw := NewResponseWriter(w)

	// Read client id from path
	clientID := domain.ClientID(r.PathValue("id"))
	if clientID == "" {
		rw.WriteBadRequest("Missing client ID")
		return
	}

	// Parse optional state query parameter
	var state *booking.BookingState
	if stateParam := r.URL.Query().Get("state"); stateParam != "" {
		bookingState := booking.BookingState(stateParam)
		// Validate state value
		if bookingState != booking.BookingStatePending &&
			bookingState != booking.BookingStateConfirmed &&
			bookingState != booking.BookingStateCancelled {
			rw.WriteBadRequest("Invalid state parameter. Must be one of: pending, confirmed, cancelled")
			return
		}
		state = &bookingState
	}

	input := list_bookings_by_client.Input{
		ClientID: clientID,
		State:    state,
	}

	bookings, err := h.listBookingsByClientUsecase.Execute(input)
	if err != nil {
		switch err {
		case common.ErrClientIDIsRequired:
			rw.WriteBadRequest(err.Error())
		default:
			rw.WriteError(err, http.StatusInternalServerError)
		}
		return
	}

	if err := rw.WriteJSON(bookings, http.StatusOK); err != nil {
		rw.WriteError(err, http.StatusInternalServerError)
	}
}
