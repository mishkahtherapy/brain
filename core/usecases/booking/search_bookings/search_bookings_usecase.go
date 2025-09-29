package search_bookings

import (
	"log/slog"
	"time"

	"github.com/mishkahtherapy/brain/core/domain"
	"github.com/mishkahtherapy/brain/core/domain/booking"
	"github.com/mishkahtherapy/brain/core/domain/client"
	"github.com/mishkahtherapy/brain/core/domain/therapist"
	"github.com/mishkahtherapy/brain/core/ports"
	"github.com/mishkahtherapy/brain/core/usecases/common"
)

// Input represents the parameters accepted by the Search Bookings use-case.
// Start and End define the inclusive UTC time range to search within.
// When State is nil no filtering by booking state is applied.
// If provided, State must be one of the valid booking.BookingState constants.
// Validation is performed inside Execute.

type Input struct {
	Start  time.Time
	End    time.Time
	States []booking.BookingState
}

type Output struct {
	RegularBookingID        domain.BookingID       `json:"regularBookingId,omitempty"`
	AdhocBookingID          domain.AdhocBookingID  `json:"adhocBookingId,omitempty"`
	TherapistID             domain.TherapistID     `json:"therapistId"`
	TherapistName           string                 `json:"therapistName"`
	ClientID                domain.ClientID        `json:"clientId"`
	ClientName              string                 `json:"clientName"`
	State                   booking.BookingState   `json:"state"`
	StartTime               domain.UTCTimestamp    `json:"startTime"` // ISO 8601 datetime, e.g. "2024-06-01T09:00:00Z"
	Duration                domain.DurationMinutes `json:"duration"`
	ClientTimezoneOffset    domain.TimezoneOffset  `json:"clientTimezoneOffset"`
	TherapistTimezoneOffset domain.TimezoneOffset  `json:"therapistTimezoneOffset"`
}

type Usecase struct {
	bookingRepo      ports.BookingRepository
	adhocBookingRepo ports.AdhocBookingRepository
	therapistRepo    ports.TherapistRepository
	clientRepo       ports.ClientRepository
}

func NewUsecase(
	bookingRepo ports.BookingRepository,
	adhocBookingRepo ports.AdhocBookingRepository,
	therapistRepo ports.TherapistRepository,
	clientRepo ports.ClientRepository,
) *Usecase {
	return &Usecase{
		bookingRepo:      bookingRepo,
		adhocBookingRepo: adhocBookingRepo,
		therapistRepo:    therapistRepo,
		clientRepo:       clientRepo,
	}
}

func (u *Usecase) Execute(input Input) ([]*Output, error) {
	// Validate date range only if both dates are provided
	if !input.Start.IsZero() && !input.End.IsZero() && input.End.Before(input.Start) {
		return nil, common.ErrInvalidDateRange
	}

	// Delegate to repository
	bookings, err := u.bookingRepo.Search(input.Start, input.End, input.States)
	if err != nil {
		return nil, common.ErrFailedToListBookings
	}

	adhocBookings, err := u.adhocBookingRepo.Search(input.Start, input.End, input.States)
	if err != nil {
		return nil, common.ErrFailedToListBookings
	}

	therapistIds, clientIds := getTherapistAndClientIds(bookings, adhocBookings)
	therapists, err := u.therapistRepo.FindByIDs(therapistIds)
	if err != nil {
		return nil, common.ErrFailedToListBookings
	}
	if len(therapistIds) != len(therapists) {
		slog.Error("therapistIds and therapists have different lengths", "therapistIds", therapistIds, "therapists", therapists, "bookings", bookings)
		return nil, common.ErrFailedToListBookings
	}

	therapistMap := make(map[domain.TherapistID]*therapist.Therapist)
	for _, therapist := range therapists {
		therapistMap[therapist.ID] = therapist
	}

	clients, err := u.clientRepo.FindByIDs(clientIds)
	if err != nil {
		return nil, common.ErrFailedToListBookings
	}
	if len(clientIds) != len(clients) {
		slog.Error("clientIds and clients have different lengths", "clientIds", clientIds, "clients", clients, "bookings", bookings)
		return nil, common.ErrFailedToListBookings
	}

	clientMap := make(map[domain.ClientID]*client.Client)
	for _, client := range clients {
		clientMap[client.ID] = client
	}

	outputs := make([]*Output, 0)

	for _, booking := range bookings {
		outputs = append(outputs, &Output{
			RegularBookingID:        booking.ID,
			TherapistID:             booking.TherapistID,
			TherapistName:           therapistMap[booking.TherapistID].Name,
			ClientID:                booking.ClientID,
			ClientName:              clientMap[booking.ClientID].Name,
			State:                   booking.State,
			StartTime:               booking.StartTime,
			Duration:                booking.Duration,
			ClientTimezoneOffset:    booking.ClientTimezoneOffset,
			TherapistTimezoneOffset: therapistMap[booking.TherapistID].TimezoneOffset,
		})
	}

	for _, adhocBooking := range adhocBookings {
		outputs = append(outputs, &Output{
			AdhocBookingID:          adhocBooking.ID,
			TherapistID:             adhocBooking.TherapistID,
			TherapistName:           therapistMap[adhocBooking.TherapistID].Name,
			ClientID:                adhocBooking.ClientID,
			ClientName:              clientMap[adhocBooking.ClientID].Name,
			State:                   adhocBooking.State,
			StartTime:               adhocBooking.StartTime,
			Duration:                adhocBooking.Duration,
			ClientTimezoneOffset:    adhocBooking.ClientTimezoneOffset,
			TherapistTimezoneOffset: therapistMap[adhocBooking.TherapistID].TimezoneOffset,
		})
	}

	return outputs, nil
}

func getTherapistAndClientIds(bookings []*booking.Booking, adhocBookings []*booking.AdhocBooking) ([]domain.TherapistID, []domain.ClientID) {
	therapistIds := make(map[domain.TherapistID]struct{})
	clientIds := make(map[domain.ClientID]struct{})

	for _, booking := range bookings {
		therapistIds[booking.TherapistID] = struct{}{}
		clientIds[booking.ClientID] = struct{}{}
	}

	for _, adhocBooking := range adhocBookings {
		therapistIds[adhocBooking.TherapistID] = struct{}{}
		clientIds[adhocBooking.ClientID] = struct{}{}
	}

	therapistIdsSlice := make([]domain.TherapistID, 0, len(therapistIds))
	clientIdsSlice := make([]domain.ClientID, 0, len(clientIds))

	for therapistId := range therapistIds {
		therapistIdsSlice = append(therapistIdsSlice, therapistId)
	}

	for clientId := range clientIds {
		clientIdsSlice = append(clientIdsSlice, clientId)
	}

	return therapistIdsSlice, clientIdsSlice
}
