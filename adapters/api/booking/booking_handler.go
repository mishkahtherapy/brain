package booking_handler

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/mishkahtherapy/brain/adapters/api"
	"github.com/mishkahtherapy/brain/core/domain"
	"github.com/mishkahtherapy/brain/core/domain/booking"
	"github.com/mishkahtherapy/brain/core/usecases/booking/cancel_booking"
	"github.com/mishkahtherapy/brain/core/usecases/booking/confirm_booking"
	"github.com/mishkahtherapy/brain/core/usecases/booking/create_booking"
	"github.com/mishkahtherapy/brain/core/usecases/booking/get_booking"
	"github.com/mishkahtherapy/brain/core/usecases/booking/list_bookings_by_client"
	"github.com/mishkahtherapy/brain/core/usecases/booking/list_bookings_by_therapist"
	"github.com/mishkahtherapy/brain/core/usecases/booking/search_bookings"
	"github.com/mishkahtherapy/brain/core/usecases/common"
)

type BookingHandler struct {
	createBookingUsecase           create_booking.Usecase
	getBookingUsecase              get_booking.Usecase
	confirmBookingUsecase          confirm_booking.Usecase
	cancelBookingUsecase           cancel_booking.Usecase
	listBookingsByTherapistUsecase list_bookings_by_therapist.Usecase
	listBookingsByClientUsecase    list_bookings_by_client.Usecase
	searchBookingsUsecase          search_bookings.Usecase
}

func NewBookingHandler(
	createUsecase create_booking.Usecase,
	getUsecase get_booking.Usecase,
	confirmUsecase confirm_booking.Usecase,
	cancelUsecase cancel_booking.Usecase,
	listByTherapistUsecase list_bookings_by_therapist.Usecase,
	listByClientUsecase list_bookings_by_client.Usecase,
	searchUsecase search_bookings.Usecase,
) *BookingHandler {
	return &BookingHandler{
		createBookingUsecase:           createUsecase,
		getBookingUsecase:              getUsecase,
		confirmBookingUsecase:          confirmUsecase,
		cancelBookingUsecase:           cancelUsecase,
		listBookingsByTherapistUsecase: listByTherapistUsecase,
		listBookingsByClientUsecase:    listByClientUsecase,
		searchBookingsUsecase:          searchUsecase,
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
	searchUsecase search_bookings.Usecase,
) {
	h.createBookingUsecase = createUsecase
	h.getBookingUsecase = getUsecase
	h.confirmBookingUsecase = confirmUsecase
	h.cancelBookingUsecase = cancelUsecase
	h.listBookingsByTherapistUsecase = listByTherapistUsecase
	h.listBookingsByClientUsecase = listByClientUsecase
	h.searchBookingsUsecase = searchUsecase
}

func (h *BookingHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/v1/bookings", h.handleCreateBooking)
	mux.HandleFunc("GET /api/v1/bookings/search", h.handleSearchBookings)
	mux.HandleFunc("GET /api/v1/bookings/{id}", h.handleGetBooking)
	mux.HandleFunc("PUT /api/v1/bookings/{id}/confirm", h.handleConfirmBooking)
	mux.HandleFunc("PUT /api/v1/bookings/{id}/cancel", h.handleCancelBooking)
	mux.HandleFunc("GET /api/v1/therapists/{id}/bookings", h.handleListBookingsByTherapist)
	mux.HandleFunc("GET /api/v1/clients/{id}/bookings", h.handleListBookingsByClient)
}

func (h *BookingHandler) handleCreateBooking(w http.ResponseWriter, r *http.Request) {
	rw := api.NewResponseWriter(w)

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
			common.ErrStartTimeIsRequired,
			create_booking.ErrTimezoneIsRequired,
			create_booking.ErrInvalidTimezone,
			domain.ErrTimezoneIsRequired,
			common.ErrTherapistNotFound,
			common.ErrClientNotFound,
			common.ErrTimeSlotNotFound,
			domain.ErrInvalidTimezone:
			rw.WriteBadRequest(err.Error())
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

func (h *BookingHandler) handleSearchBookings(w http.ResponseWriter, r *http.Request) {
	rw := api.NewResponseWriter(w)

	// Parse optional start & end query params (YYYY-MM-DD expected)
	startParam := r.URL.Query().Get("start")
	endParam := r.URL.Query().Get("end")
	stateParam := r.URL.Query().Get("state")

	var startTime, endTime time.Time
	var err error

	// Parse start date if provided
	if startParam != "" {
		startTime, err = time.Parse(time.DateOnly, startParam)
		if err != nil {
			rw.WriteBadRequest("invalid start parameter. Expected YYYY-MM-DD format")
			return
		}
		startTime = startTime.UTC()
	}

	// Parse end date if provided
	if endParam != "" {
		endTime, err = time.Parse(time.DateOnly, endParam)
		if err != nil {
			rw.WriteBadRequest("invalid end parameter. Expected YYYY-MM-DD format")
			return
		}
		endTime = endTime.AddDate(0, 0, 1).Add(-time.Nanosecond).UTC() // End of day
	}

	// Validate date range only if both dates are provided
	if !startTime.IsZero() && !endTime.IsZero() && endTime.Before(startTime) {
		rw.WriteBadRequest("end must be after start")
		return
	}

	// Optional state filter
	var states []booking.BookingState
	if stateParam != "" {
		bookingStates := []booking.BookingState{}
		for _, state := range strings.Split(stateParam, ",") {
			bookingState := booking.BookingState(state)
			bookingStates = append(bookingStates, bookingState)
			if bookingState != booking.BookingStatePending &&
				bookingState != booking.BookingStateConfirmed &&
				bookingState != booking.BookingStateCancelled {
				rw.WriteBadRequest("Invalid state parameter. Must be one of: pending, confirmed, cancelled")
				return
			}
		}
		states = bookingStates
	}

	input := search_bookings.Input{
		Start:  startTime,
		End:    endTime,
		States: states,
	}

	bookings, err := h.searchBookingsUsecase.Execute(input)
	if err != nil {
		switch err {
		case common.ErrInvalidDateRange:
			rw.WriteBadRequest(err.Error())
		case common.ErrFailedToListBookings:
			rw.WriteError(err, http.StatusInternalServerError)
		default:
			rw.WriteError(err, http.StatusInternalServerError)
		}
		return
	}

	if err := rw.WriteJSON(bookings, http.StatusOK); err != nil {
		rw.WriteError(err, http.StatusInternalServerError)
	}
}

func (h *BookingHandler) handleGetBooking(w http.ResponseWriter, r *http.Request) {
	rw := api.NewResponseWriter(w)

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
	rw := api.NewResponseWriter(w)

	// Read id from path
	id := domain.BookingID(r.PathValue("id"))
	if id == "" {
		rw.WriteBadRequest("Missing booking ID")
		return
	}

	// Parse request body to get paid amount and language
	var requestBody struct {
		PaidAmountUSD int                    `json:"paidAmount"` // WhatsApp currency (smallest unit integer)
		Language      domain.SessionLanguage `json:"language"`
	}

	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
		rw.WriteBadRequest(err.Error())
		return
	}

	input := confirm_booking.Input{
		BookingID:     id,
		PaidAmountUSD: requestBody.PaidAmountUSD,
		Language:      requestBody.Language,
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
	rw := api.NewResponseWriter(w)

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
	rw := api.NewResponseWriter(w)

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
	rw := api.NewResponseWriter(w)

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
