package booking_db

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

type BookingRepository struct {
	db ports.SQLDatabase
}

func NewBookingRepository(db ports.SQLDatabase) ports.BookingRepository {
	return &BookingRepository{db: db}
}

func (r *BookingRepository) GetByID(id domain.BookingID) (*booking.Booking, error) {
	query := `
		SELECT id, timeslot_id, therapist_id, client_id, start_time, duration_minutes, client_timezone_offset, state, created_at, updated_at
		FROM bookings
		WHERE id = ?
	`
	row := r.db.QueryRow(query, id)
	booking := &booking.Booking{}
	err := row.Scan(
		&booking.ID,
		&booking.TimeSlotID,
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

func (r *BookingRepository) Create(booking *booking.Booking) error {
	if booking.ID == "" {
		return ports.ErrBookingIDIsRequired
	}

	if booking.TimeSlotID == "" {
		return ports.ErrBookingTimeSlotIDIsRequired
	}

	if booking.TherapistID == "" {
		return ports.ErrBookingTherapistIDIsRequired
	}

	if booking.ClientID == "" {
		return ports.ErrBookingClientIDIsRequired
	}

	if booking.State == "" {
		return ports.ErrBookingStateIsRequired
	}

	if booking.StartTime == (domain.UTCTimestamp{}) {
		return ports.ErrBookingStartTimeIsRequired
	}

	if booking.CreatedAt == (domain.UTCTimestamp{}) {
		return ports.ErrBookingCreatedAtIsRequired
	}

	if booking.UpdatedAt == (domain.UTCTimestamp{}) {
		return ports.ErrBookingUpdatedAtIsRequired
	}

	if booking.Duration == 0 {
		return ports.ErrBookingDurationIsRequired
	}

	query := `
		INSERT INTO bookings (
			id, timeslot_id, therapist_id, client_id, start_time, duration_minutes, client_timezone_offset, state, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	_, err := r.db.Exec(
		query,
		booking.ID,
		booking.TimeSlotID,
		booking.TherapistID,
		booking.ClientID,
		booking.StartTime,
		booking.Duration,
		booking.ClientTimezoneOffset,
		booking.State,
		booking.CreatedAt,
		booking.UpdatedAt,
	)
	if err != nil {
		slog.Error("error creating booking", "error", err)
		return ports.ErrFailedToCreateBooking
	}
	return nil
}

func (r *BookingRepository) UpdateTx(sqlExec ports.SQLExec, booking *booking.Booking) error {
	if booking.ID == "" {
		return ports.ErrBookingIDIsRequired
	}

	if booking.TimeSlotID == "" {
		return ports.ErrBookingTimeSlotIDIsRequired
	}

	if booking.TherapistID == "" {
		return ports.ErrBookingTherapistIDIsRequired
	}

	if booking.ClientID == "" {
		return ports.ErrBookingClientIDIsRequired
	}

	if booking.State == "" {
		return ports.ErrBookingStateIsRequired
	}

	if booking.StartTime == (domain.UTCTimestamp{}) {
		return ports.ErrBookingStartTimeIsRequired
	}

	if booking.UpdatedAt == (domain.UTCTimestamp{}) {
		return ports.ErrBookingUpdatedAtIsRequired
	}

	query := `
		UPDATE bookings 
			SET timeslot_id = ?, therapist_id = ?, client_id = ?, start_time = ?, duration_minutes = ?, client_timezone_offset = ?, state = ?, updated_at = ?
		WHERE id = ?
	`
	result, err := sqlExec.Exec(
		query,
		booking.TimeSlotID,
		booking.TherapistID,
		booking.ClientID,
		booking.StartTime,
		booking.Duration,
		booking.ClientTimezoneOffset,
		booking.State,
		booking.UpdatedAt,
		booking.ID,
	)
	if err != nil {
		slog.Error("error updating booking", "error", err)
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

func (r *BookingRepository) Update(booking *booking.Booking) error {
	return r.UpdateTx(r.db, booking)
}

func (r *BookingRepository) Delete(id domain.BookingID) error {
	if id == "" {
		return ports.ErrBookingIDIsRequired
	}

	query := `DELETE FROM bookings WHERE id = ?`
	result, err := r.db.Exec(query, id)
	if err != nil {
		slog.Error("error deleting booking", "error", err)
		return ports.ErrFailedToDeleteBooking
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		slog.Error("error getting rows affected after delete", "error", err)
		return ports.ErrFailedToDeleteBooking
	}

	if rowsAffected == 0 {
		return ports.ErrBookingNotFound
	}

	return nil
}

func (r *BookingRepository) ListByTherapist(therapistID domain.TherapistID) ([]*booking.Booking, error) {
	if therapistID == "" {
		return nil, ports.ErrBookingTherapistIDIsRequired
	}

	query := `
		SELECT id, timeslot_id, therapist_id, client_id, start_time, duration_minutes, client_timezone_offset, state, created_at, updated_at
		FROM bookings
		WHERE therapist_id = ?
		ORDER BY start_time ASC
	`
	rows, err := r.db.Query(query, therapistID)
	if err != nil {
		slog.Error("error listing bookings by therapist", "error", err)
		return nil, ports.ErrFailedToGetBookings
	}
	defer rows.Close()

	return r.scanBookings(rows)
}

func (r *BookingRepository) ListByClient(clientID domain.ClientID) ([]*booking.Booking, error) {
	if clientID == "" {
		return nil, ports.ErrBookingClientIDIsRequired
	}

	query := `
		SELECT id, timeslot_id, therapist_id, client_id, start_time, duration_minutes, client_timezone_offset, state, created_at, updated_at
		FROM bookings
		WHERE client_id = ?
		ORDER BY start_time ASC
	`
	rows, err := r.db.Query(query, clientID)
	if err != nil {
		slog.Error("error listing bookings by client", "error", err)
		return nil, ports.ErrFailedToGetBookings
	}
	defer rows.Close()

	return r.scanBookings(rows)
}

func (r *BookingRepository) ListByState(state booking.BookingState) ([]*booking.Booking, error) {
	if state == "" {
		return nil, ports.ErrBookingStateIsRequired
	}

	query := `
		SELECT id, timeslot_id, therapist_id, client_id, start_time, duration_minutes, client_timezone_offset, state, created_at, updated_at
		FROM bookings
		WHERE state = ?
		ORDER BY start_time ASC
	`
	rows, err := r.db.Query(query, state)
	if err != nil {
		slog.Error("error listing bookings by state", "error", err)
		return nil, ports.ErrFailedToGetBookings
	}
	defer rows.Close()

	return r.scanBookings(rows)
}

func (r *BookingRepository) ListByTherapistAndState(therapistID domain.TherapistID, state booking.BookingState) ([]*booking.Booking, error) {
	if therapistID == "" {
		return nil, ports.ErrBookingTherapistIDIsRequired
	}

	if state == "" {
		return nil, ports.ErrBookingStateIsRequired
	}

	query := `
		SELECT id, timeslot_id, therapist_id, client_id, start_time, duration_minutes, client_timezone_offset, state, created_at, updated_at
		FROM bookings
		WHERE therapist_id = ? AND state = ?
		ORDER BY start_time ASC
	`
	rows, err := r.db.Query(query, therapistID, state)
	if err != nil {
		slog.Error("error listing bookings by therapist and state", "error", err)
		return nil, ports.ErrFailedToGetBookings
	}
	defer rows.Close()

	return r.scanBookings(rows)
}

func (r *BookingRepository) ListByClientAndState(clientID domain.ClientID, state booking.BookingState) ([]*booking.Booking, error) {
	if clientID == "" {
		return nil, ports.ErrBookingClientIDIsRequired
	}

	if state == "" {
		return nil, ports.ErrBookingStateIsRequired
	}

	query := `
		SELECT id, timeslot_id, therapist_id, client_id, start_time, duration_minutes, client_timezone_offset, state, created_at, updated_at
		FROM bookings
		WHERE client_id = ? AND state = ?
		ORDER BY start_time ASC
	`
	rows, err := r.db.Query(query, clientID, state)
	if err != nil {
		slog.Error("error listing bookings by client and state", "error", err)
		return nil, ports.ErrFailedToGetBookings
	}
	defer rows.Close()

	return r.scanBookings(rows)
}

// Helper method to scan multiple booking rows
func (r *BookingRepository) scanBookings(rows *sql.Rows) ([]*booking.Booking, error) {
	bookings := make([]*booking.Booking, 0)
	for rows.Next() {
		booking := &booking.Booking{}
		err := rows.Scan(
			&booking.ID,
			&booking.TimeSlotID,
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
			slog.Error("error scanning booking", "error", err)
			return nil, ports.ErrFailedToGetBookings
		}
		bookings = append(bookings, booking)
	}
	return bookings, nil
}

func (r *BookingRepository) BulkListByTherapistForDateRange(
	therapistIDs []domain.TherapistID,
	states []booking.BookingState,
	startDate time.Time,
	endDate time.Time,
) (map[domain.TherapistID][]*booking.Booking, error) {
	if len(therapistIDs) == 0 {
		return nil, ports.ErrBookingTherapistIDIsRequired
	}

	// Calculate endTimeForBookingStartedBeforeRange (startDate - 1 hour)
	// FIXME: I don't capture bookings that start before the range but extend into it
	// Example: a booking at 11.30PM that ends at 12.30AM next day is not captured.

	query := `
	       SELECT id, timeslot_id, therapist_id, client_id, start_time, duration_minutes, client_timezone_offset, state, created_at, updated_at
	       FROM bookings
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
		slog.Error("error listing confirmed bookings by therapist for date range",
			"error", err,
			"therapistIDs", therapistIDs,
			"startDate", startDate,
			"endDate", endDate,
		)
		return nil, ports.ErrFailedToGetBookings
	}
	defer rows.Close()

	bookings := make(map[domain.TherapistID][]*booking.Booking)
	for rows.Next() {
		booking := &booking.Booking{}
		err := rows.Scan(
			&booking.ID,
			&booking.TimeSlotID,
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
			slog.Error("error scanning booking", "error", err)
			return nil, ports.ErrFailedToGetBookings
		}
		bookings[booking.TherapistID] = append(bookings[booking.TherapistID], booking)
	}

	return bookings, nil
}

// Search returns all bookings whose start_time is within the inclusive range
// [startDate, endDate]. When state is provided (non-nil), the results are
// further filtered by the given booking state.
func (r *BookingRepository) Search(startDate, endDate time.Time, states []booking.BookingState) ([]*booking.Booking, error) {
	query := `
		SELECT id, timeslot_id, therapist_id, client_id, start_time, duration_minutes, client_timezone_offset, state, created_at, updated_at
		FROM bookings
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

	return r.scanBookings(rows)
}

func (r *BookingRepository) BulkCancel(tx ports.SQLTx, bookingIDs []domain.BookingID) error {
	query := `
		UPDATE bookings
		SET state = ?
		WHERE id IN (%s)
	`
	values := make([]any, 0)
	values = append(values, booking.BookingStateCancelled)

	placeholders := make([]string, len(bookingIDs))
	for i := range bookingIDs {
		placeholders[i] = "?"
		values = append(values, bookingIDs[i])
	}
	placeholdersStr := strings.Join(placeholders, ",")
	query = fmt.Sprintf(query, placeholdersStr)

	_, err := tx.Exec(query, values...)
	if err != nil {
		slog.Error("error bulk cancelling bookings", "error", err)
		return ports.ErrFailedToUpdateBooking
	}
	return nil
}
