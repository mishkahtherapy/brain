package booking_db

import (
	"database/sql"
	"errors"
	"log/slog"
	"time"

	"github.com/mishkahtherapy/brain/core/domain"
	"github.com/mishkahtherapy/brain/core/domain/booking"
	"github.com/mishkahtherapy/brain/core/ports"
)

type BookingRepository struct {
	db ports.SQLDatabase
}

var ErrBookingNotFound = errors.New("booking not found")
var ErrBookingAlreadyExists = errors.New("booking already exists")
var ErrBookingIDIsRequired = errors.New("booking id is required")
var ErrBookingTimeSlotIDIsRequired = errors.New("booking timeslot id is required")
var ErrBookingTherapistIDIsRequired = errors.New("booking therapist id is required")
var ErrBookingClientIDIsRequired = errors.New("booking client id is required")
var ErrBookingStateIsRequired = errors.New("booking state is required")
var ErrBookingStartTimeIsRequired = errors.New("booking start time is required")
var ErrBookingCreatedAtIsRequired = errors.New("booking created at is required")
var ErrBookingUpdatedAtIsRequired = errors.New("booking updated at is required")
var ErrFailedToGetBookings = errors.New("failed to get bookings")
var ErrFailedToCreateBooking = errors.New("failed to create booking")
var ErrFailedToUpdateBooking = errors.New("failed to update booking")
var ErrFailedToDeleteBooking = errors.New("failed to delete booking")
var ErrInvalidDateRange = errors.New("invalid date range")

func NewBookingRepository(db ports.SQLDatabase) *BookingRepository {
	return &BookingRepository{db: db}
}

func (r *BookingRepository) GetByID(id domain.BookingID) (*booking.Booking, error) {
	query := `
		SELECT id, timeslot_id, therapist_id, client_id, start_time, timezone_offset, state, created_at, updated_at
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
		&booking.TimezoneOffset,
		&booking.State,
		&booking.CreatedAt,
		&booking.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrBookingNotFound
		}
		slog.Error("error getting booking by id", "error", err)
		return nil, ErrFailedToGetBookings
	}
	return booking, nil
}

func (r *BookingRepository) Create(booking *booking.Booking) error {
	if booking.ID == "" {
		return ErrBookingIDIsRequired
	}

	if booking.TimeSlotID == "" {
		return ErrBookingTimeSlotIDIsRequired
	}

	if booking.TherapistID == "" {
		return ErrBookingTherapistIDIsRequired
	}

	if booking.ClientID == "" {
		return ErrBookingClientIDIsRequired
	}

	if booking.State == "" {
		return ErrBookingStateIsRequired
	}

	if booking.StartTime == (domain.UTCTimestamp{}) {
		return ErrBookingStartTimeIsRequired
	}

	if booking.CreatedAt == (domain.UTCTimestamp{}) {
		return ErrBookingCreatedAtIsRequired
	}

	if booking.UpdatedAt == (domain.UTCTimestamp{}) {
		return ErrBookingUpdatedAtIsRequired
	}

	query := `
		INSERT INTO bookings (id, timeslot_id, therapist_id, client_id, start_time, timezone_offset, state, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	_, err := r.db.Exec(
		query,
		booking.ID,
		booking.TimeSlotID,
		booking.TherapistID,
		booking.ClientID,
		booking.StartTime,
		booking.TimezoneOffset,
		booking.State,
		booking.CreatedAt,
		booking.UpdatedAt,
	)
	if err != nil {
		slog.Error("error creating booking", "error", err)
		return ErrFailedToCreateBooking
	}
	return nil
}

func (r *BookingRepository) Update(booking *booking.Booking) error {
	if booking.ID == "" {
		return ErrBookingIDIsRequired
	}

	if booking.TimeSlotID == "" {
		return ErrBookingTimeSlotIDIsRequired
	}

	if booking.TherapistID == "" {
		return ErrBookingTherapistIDIsRequired
	}

	if booking.ClientID == "" {
		return ErrBookingClientIDIsRequired
	}

	if booking.State == "" {
		return ErrBookingStateIsRequired
	}

	if booking.StartTime == (domain.UTCTimestamp{}) {
		return ErrBookingStartTimeIsRequired
	}

	if booking.UpdatedAt == (domain.UTCTimestamp{}) {
		return ErrBookingUpdatedAtIsRequired
	}

	query := `
		UPDATE bookings 
		SET timeslot_id = ?, therapist_id = ?, client_id = ?, start_time = ?, timezone_offset = ?, state = ?, updated_at = ?
		WHERE id = ?
	`
	result, err := r.db.Exec(
		query,
		booking.TimeSlotID,
		booking.TherapistID,
		booking.ClientID,
		booking.StartTime,
		booking.TimezoneOffset,
		booking.State,
		booking.UpdatedAt,
		booking.ID,
	)
	if err != nil {
		slog.Error("error updating booking", "error", err)
		return ErrFailedToUpdateBooking
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		slog.Error("error getting rows affected after update", "error", err)
		return ErrFailedToUpdateBooking
	}

	if rowsAffected == 0 {
		return ErrBookingNotFound
	}

	return nil
}

func (r *BookingRepository) Delete(id domain.BookingID) error {
	if id == "" {
		return ErrBookingIDIsRequired
	}

	query := `DELETE FROM bookings WHERE id = ?`
	result, err := r.db.Exec(query, id)
	if err != nil {
		slog.Error("error deleting booking", "error", err)
		return ErrFailedToDeleteBooking
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		slog.Error("error getting rows affected after delete", "error", err)
		return ErrFailedToDeleteBooking
	}

	if rowsAffected == 0 {
		return ErrBookingNotFound
	}

	return nil
}

func (r *BookingRepository) ListByTherapist(therapistID domain.TherapistID) ([]*booking.Booking, error) {
	if therapistID == "" {
		return nil, ErrBookingTherapistIDIsRequired
	}

	query := `
		SELECT id, timeslot_id, therapist_id, client_id, start_time, timezone_offset, state, created_at, updated_at
		FROM bookings
		WHERE therapist_id = ?
		ORDER BY start_time ASC
	`
	rows, err := r.db.Query(query, therapistID)
	if err != nil {
		slog.Error("error listing bookings by therapist", "error", err)
		return nil, ErrFailedToGetBookings
	}
	defer rows.Close()

	return r.scanBookings(rows)
}

func (r *BookingRepository) ListByClient(clientID domain.ClientID) ([]*booking.Booking, error) {
	if clientID == "" {
		return nil, ErrBookingClientIDIsRequired
	}

	query := `
		SELECT id, timeslot_id, therapist_id, client_id, start_time, timezone_offset, state, created_at, updated_at
		FROM bookings
		WHERE client_id = ?
		ORDER BY start_time ASC
	`
	rows, err := r.db.Query(query, clientID)
	if err != nil {
		slog.Error("error listing bookings by client", "error", err)
		return nil, ErrFailedToGetBookings
	}
	defer rows.Close()

	return r.scanBookings(rows)
}

func (r *BookingRepository) ListByState(state booking.BookingState) ([]*booking.Booking, error) {
	if state == "" {
		return nil, ErrBookingStateIsRequired
	}

	query := `
		SELECT id, timeslot_id, therapist_id, client_id, start_time, timezone_offset, state, created_at, updated_at
		FROM bookings
		WHERE state = ?
		ORDER BY start_time ASC
	`
	rows, err := r.db.Query(query, state)
	if err != nil {
		slog.Error("error listing bookings by state", "error", err)
		return nil, ErrFailedToGetBookings
	}
	defer rows.Close()

	return r.scanBookings(rows)
}

func (r *BookingRepository) ListByTherapistAndState(therapistID domain.TherapistID, state booking.BookingState) ([]*booking.Booking, error) {
	if therapistID == "" {
		return nil, ErrBookingTherapistIDIsRequired
	}

	if state == "" {
		return nil, ErrBookingStateIsRequired
	}

	query := `
		SELECT id, timeslot_id, therapist_id, client_id, start_time, timezone_offset, state, created_at, updated_at
		FROM bookings
		WHERE therapist_id = ? AND state = ?
		ORDER BY start_time ASC
	`
	rows, err := r.db.Query(query, therapistID, state)
	if err != nil {
		slog.Error("error listing bookings by therapist and state", "error", err)
		return nil, ErrFailedToGetBookings
	}
	defer rows.Close()

	return r.scanBookings(rows)
}

func (r *BookingRepository) ListByClientAndState(clientID domain.ClientID, state booking.BookingState) ([]*booking.Booking, error) {
	if clientID == "" {
		return nil, ErrBookingClientIDIsRequired
	}

	if state == "" {
		return nil, ErrBookingStateIsRequired
	}

	query := `
		SELECT id, timeslot_id, therapist_id, client_id, start_time, timezone_offset, state, created_at, updated_at
		FROM bookings
		WHERE client_id = ? AND state = ?
		ORDER BY start_time ASC
	`
	rows, err := r.db.Query(query, clientID, state)
	if err != nil {
		slog.Error("error listing bookings by client and state", "error", err)
		return nil, ErrFailedToGetBookings
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
			&booking.TimezoneOffset,
			&booking.State,
			&booking.CreatedAt,
			&booking.UpdatedAt,
		)
		if err != nil {
			slog.Error("error scanning booking", "error", err)
			return nil, ErrFailedToGetBookings
		}
		bookings = append(bookings, booking)
	}
	return bookings, nil
}

func (r *BookingRepository) ListConfirmedByTherapistForDateRange(
	therapistID domain.TherapistID, startDate, endDate time.Time) ([]*booking.Booking, error) {
	if therapistID == "" {
		return nil, ErrBookingTherapistIDIsRequired
	}

	// TODO: make this a parameter.
	// Calculate endTimeForBookingStartedBeforeRange (startDate - 1 hour)
	// This captures bookings that start before the range but extend into it
	oneHourBefore := startDate.Add(-time.Hour)

	query := `
	       SELECT id, timeslot_id, therapist_id, client_id, start_time, timezone_offset, state, created_at, updated_at
	       FROM bookings
	       WHERE therapist_id = ?
	       AND state = ?
	       AND (
		       -- Bookings that start within the range
		       (start_time >= ? AND start_time <= ?)
		       OR
		       -- Bookings that start before the range but end within or after it
		       -- (assuming fixed 1-hour duration for all bookings)
		       (start_time >= ? AND start_time < ?)
	       )
	       ORDER BY start_time ASC
	   `

	rows, err := r.db.Query(
		query,
		therapistID,
		booking.BookingStateConfirmed, // TODO: refactor this parameter
		startDate, endDate,            // For bookings starting within range
		oneHourBefore, startDate, // For bookings starting before range but extending into it
	)
	if err != nil {
		slog.Error("error listing confirmed bookings by therapist for date range",
			"error", err,
			"therapistID", therapistID,
			"startDate", startDate,
			"endDate", endDate,
		)
		return nil, ErrFailedToGetBookings
	}
	defer rows.Close()

	return r.scanBookings(rows)
}

// Search returns all bookings whose start_time is within the inclusive range
// [startDate, endDate]. When state is provided (non-nil), the results are
// further filtered by the given booking state.
func (r *BookingRepository) Search(startDate, endDate time.Time, state *booking.BookingState) ([]*booking.Booking, error) {
	query := `
		SELECT id, timeslot_id, therapist_id, client_id, start_time, timezone_offset, state, created_at, updated_at
		FROM bookings
		WHERE 1=1
	`
	params := []interface{}{}

	// Add start date filter if provided (not zero time)
	if !startDate.IsZero() {
		query += " AND start_time >= ?"
		params = append(params, startDate)
	}

	// Add end date filter if provided (not zero time)
	if !endDate.IsZero() {
		query += " AND start_time <= ?"
		params = append(params, endDate)
	}

	// Add state filter if provided
	if state != nil {
		query += " AND state = ?"
		params = append(params, *state)
	}

	query += " ORDER BY start_time ASC"

	rows, err := r.db.Query(query, params...)
	if err != nil {
		slog.Error("error searching bookings", "error", err)
		return nil, ErrFailedToGetBookings
	}
	defer rows.Close()

	return r.scanBookings(rows)
}
