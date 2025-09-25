package adhoc_booking_db

import (
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/mishkahtherapy/brain/core/domain"
	"github.com/mishkahtherapy/brain/core/domain/booking"
	"github.com/mishkahtherapy/brain/core/ports"
)

type AdhocBookingRepository struct {
	db ports.SQLDatabase
}

func NewAdhocBookingRepository(db ports.SQLDatabase) ports.AdhocBookingRepository {
	return &AdhocBookingRepository{db: db}
}

func (r *AdhocBookingRepository) Create(adhocBooking *booking.AdhocBooking) error {
	query := `
		INSERT INTO adhoc_bookings (id, therapist_id, client_id, start_time, duration_minutes, client_timezone_offset, state, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`
	_, err := r.db.Exec(query, adhocBooking.ID, adhocBooking.TherapistID, adhocBooking.ClientID, adhocBooking.StartTime, adhocBooking.Duration, adhocBooking.ClientTimezoneOffset, adhocBooking.State, adhocBooking.CreatedAt, adhocBooking.UpdatedAt)
	if err != nil {
		slog.Error("error creating adhoc booking", "error", err)
		return ports.ErrFailedToCreateBooking
	}
	return nil
}

func (r *AdhocBookingRepository) BulkListByTherapistForDateRange(
	therapistIDs []domain.TherapistID,
	states []booking.BookingState,
	startDate time.Time,
	endDate time.Time,
) (map[domain.TherapistID][]*booking.AdhocBooking, error) {
	if len(therapistIDs) == 0 {
		return nil, ports.ErrBookingTherapistIDIsRequired
	}

	// Calculate endTimeForBookingStartedBeforeRange (startDate - 1 hour)
	// FIXME: I don't capture bookings that start before the range but extend into it
	// Example: a booking at 11.30PM that ends at 12.30AM next day is not captured.

	query := `
	       SELECT id, therapist_id, client_id, start_time, duration_minutes, client_timezone_offset, state, created_at, updated_at
	       FROM adhoc_bookings
	       WHERE state IN (%s)
	       AND (
		       -- Bookings that start within the range
		       (start_time >= ? AND start_time <= ?)
		       OR
		       -- Parse partially overlapping bookings. Add start_time + duration_minutes
		       (
			   	datetime(start_time, '+' || duration_minutes || ' minutes') > ? 
				AND 
			    datetime(start_time, '+' || duration_minutes || ' minutes') <= ?
			   )
	       )
	       AND therapist_id IN (%s)
	       ORDER BY start_time ASC
	   `

	statePlaceholders := make([]string, 0)
	stateValues := make([]interface{}, 0)
	for _, state := range states {
		statePlaceholders = append(statePlaceholders, "?")
		stateValues = append(stateValues, state)
	}
	statePlaceholdersStr := strings.Join(statePlaceholders, ",")

	therapistIDPlaceholders := make([]string, 0)
	therapistIds := make([]interface{}, 0)
	for _, id := range therapistIDs {
		therapistIDPlaceholders = append(therapistIDPlaceholders, "?")
		therapistIds = append(therapistIds, id)
	}
	therapistIDPlaceholdersStr := strings.Join(therapistIDPlaceholders, ",")

	query = fmt.Sprintf(query, statePlaceholdersStr, therapistIDPlaceholdersStr)

	values := []interface{}{}

	values = append(values, stateValues...)
	values = append(values, startDate)
	values = append(values, endDate)
	values = append(values, startDate)
	values = append(values, endDate)
	values = append(values, therapistIds...)

	rows, err := r.db.Query(query, values...)

	if err != nil {
		slog.Error("error listing confirmed adhoc bookings by therapist for date range",
			"error", err,
			"therapistIDs", therapistIDs,
			"startDate", startDate,
			"endDate", endDate,
		)
		return nil, ports.ErrFailedToGetBookings
	}
	defer rows.Close()

	adhocBookings := make(map[domain.TherapistID][]*booking.AdhocBooking)
	for rows.Next() {
		adhocBooking := &booking.AdhocBooking{}
		err := rows.Scan(
			&adhocBooking.ID,
			&adhocBooking.TherapistID,
			&adhocBooking.ClientID,
			&adhocBooking.StartTime,
			&adhocBooking.Duration,
			&adhocBooking.ClientTimezoneOffset,
			&adhocBooking.State,
			&adhocBooking.CreatedAt,
			&adhocBooking.UpdatedAt,
		)
		if err != nil {
			slog.Error("error scanning adhoc booking", "error", err)
			return nil, ports.ErrFailedToGetBookings
		}
		adhocBookings[adhocBooking.TherapistID] = append(adhocBookings[adhocBooking.TherapistID], adhocBooking)
	}

	return adhocBookings, nil
}
