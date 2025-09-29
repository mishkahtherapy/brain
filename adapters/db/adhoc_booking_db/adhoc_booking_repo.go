package adhoc_booking_db

import (
	"database/sql"
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

// GetByID implements ports.AdhocBookingRepository.
func (r *AdhocBookingRepository) GetByID(id domain.AdhocBookingID) (*booking.AdhocBooking, error) {
	query := `
		SELECT 
			id, therapist_id,
			client_id, start_time,
			duration_minutes, client_timezone_offset,
			state, created_at, updated_at
		FROM adhoc_bookings
		WHERE id = ?
	`
	row := r.db.QueryRow(query, id)
	booking := &booking.AdhocBooking{}
	err := row.Scan(
		&booking.ID,
		&booking.TherapistID,
		&booking.ClientID,
		&booking.StartTime,
		&booking.Duration,
		&booking.ClientTimezoneOffset,
		&booking.State,
		&booking.CreatedAt,
		&booking.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ports.ErrBookingNotFound
		}
		slog.Error("error getting booking by id", "error", err)
		return nil, ports.ErrFailedToGetBookings
	}
	return booking, nil
}

// UpdateStateTx implements ports.AdhocBookingRepository.
func (r *AdhocBookingRepository) UpdateStateTx(
	sqlExec ports.SQLExec,
	adhocBookingID domain.AdhocBookingID,
	state booking.BookingState,
	updatedAt time.Time,
) error {
	if adhocBookingID == "" {
		return ports.ErrBookingIDIsRequired
	}

	if state == "" {
		return ports.ErrBookingStateIsRequired
	}

	query := `
		UPDATE adhoc_bookings 
			SET state = ?, updated_at = ?
		WHERE id = ?
	`
	result, err := sqlExec.Exec(
		query,
		state,
		updatedAt,
		adhocBookingID,
	)

	if err != nil {
		slog.Error("error updating adhoc booking", "error", err)
		return ports.ErrFailedToUpdateBooking
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		slog.Error("error getting rows affected after update", "error", err)
		return ports.ErrFailedToUpdateBooking
	}

	if rowsAffected == 0 {
		return ports.ErrBookingNotFound
	}

	return nil
}

func (r *AdhocBookingRepository) UpdateState(
	adhocBookingID domain.AdhocBookingID,
	state booking.BookingState,
	updatedAt time.Time,
) error {
	return r.UpdateStateTx(r.db, adhocBookingID, state, updatedAt)
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

func (r *AdhocBookingRepository) ListByTherapistForDateRange(
	therapistID domain.TherapistID,
	states []booking.BookingState,
	startDate time.Time,
	endDate time.Time,
) ([]*booking.AdhocBooking, error) {
	adhocBookings, err := r.BulkListByTherapistForDateRange(
		[]domain.TherapistID{therapistID},
		states,
		startDate,
		endDate,
	)
	if err != nil {
		return nil, err
	}
	return adhocBookings[therapistID], nil
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

func (r *AdhocBookingRepository) BulkCancel(tx ports.SQLTx, adhocBookingIDs []domain.AdhocBookingID) error {
	query := `
		UPDATE adhoc_bookings
		SET state = ?
		WHERE id IN (%s)
	`
	values := make([]any, 0)
	values = append(values, booking.BookingStateCancelled)

	placeholders := make([]string, len(adhocBookingIDs))
	for i := range adhocBookingIDs {
		placeholders[i] = "?"
		values = append(values, adhocBookingIDs[i])
	}
	placeholdersStr := strings.Join(placeholders, ",")
	query = fmt.Sprintf(query, placeholdersStr)

	_, err := tx.Exec(query, values...)
	if err != nil {
		slog.Error("error bulk cancelling adhoc bookings", "error", err)
		return ports.ErrFailedToUpdateBooking
	}
	return nil
}

func (r *AdhocBookingRepository) Search(startDate, endDate time.Time, states []booking.BookingState) ([]*booking.AdhocBooking, error) {
	query := `
		SELECT id, therapist_id, client_id, start_time, duration_minutes, client_timezone_offset, state, created_at, updated_at
		FROM adhoc_bookings
		WHERE 1=1
	`
	params := []interface{}{}

	// Add start date filter if provided (not zero time)
	if !startDate.IsZero() {
		query += " AND start_time >= ?"
		params = append(params, startDate)
	}

	// Add states filter if provided
	if len(states) > 0 {
		placeholders := make([]string, len(states))
		for i := range states {
			placeholders[i] = "?"
		}
		placeholdersStr := strings.Join(placeholders, ",")
		query += fmt.Sprintf(" AND state IN (%s)", placeholdersStr)
		for _, state := range states {
			params = append(params, state)
		}
	}

	// Add end date filter if provided (not zero time)
	if !endDate.IsZero() {
		query += " AND start_time <= ?"
		params = append(params, endDate)
	}

	query += " ORDER BY start_time ASC"

	rows, err := r.db.Query(query, params...)
	if err != nil {
		slog.Error("error searching bookings", "error", err)
		return nil, ports.ErrFailedToGetBookings
	}
	defer rows.Close()

	return r.scanAdhocBookings(rows)
}

func (r *AdhocBookingRepository) List(filters ports.BookingFilters) ([]*booking.AdhocBooking, error) {
	if !filters.IsValid() {
		return nil, ports.ErrInvalidBookingFilters
	}

	query := `
		SELECT id, therapist_id, client_id, start_time, duration_minutes, client_timezone_offset, state, created_at, updated_at
		FROM adhoc_bookings
		WHERE 1=1
	`

	params := []interface{}{}
	if filters.TherapistID != "" {
		query += ` AND therapist_id = ?`
		params = append(params, filters.TherapistID)
	}
	if filters.ClientID != "" {
		query += ` AND client_id = ?`
		params = append(params, filters.ClientID)
	}

	if filters.State != "" {
		query += ` AND state = ?`
		params = append(params, filters.State)
	}

	query += ` ORDER BY start_time ASC`

	rows, err := r.db.Query(query, params...)
	if err != nil {
		slog.Error("error listing bookings", "error", err)
		return nil, ports.ErrFailedToGetBookings
	}
	defer rows.Close()

	return r.scanAdhocBookings(rows)
}

// Helper method to scan multiple booking rows
func (r *AdhocBookingRepository) scanAdhocBookings(rows *sql.Rows) ([]*booking.AdhocBooking, error) {
	adhocBookings := make([]*booking.AdhocBooking, 0)
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
		adhocBookings = append(adhocBookings, adhocBooking)
	}
	return adhocBookings, nil
}
