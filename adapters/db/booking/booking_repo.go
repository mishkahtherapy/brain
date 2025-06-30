package booking

import (
	"database/sql"
	"errors"
	"log/slog"

	"github.com/mishkahtherapy/brain/adapters/db"
	"github.com/mishkahtherapy/brain/core/domain"
)

type BookingRepository struct {
	db db.SQLDatabase
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

func NewBookingRepository(db db.SQLDatabase) *BookingRepository {
	return &BookingRepository{db: db}
}

func (r *BookingRepository) GetByID(id domain.BookingID) (*domain.Booking, error) {
	query := `
		SELECT id, timeslot_id, therapist_id, client_id, start_time, state, created_at, updated_at
		FROM bookings
		WHERE id = ?
	`
	row := r.db.QueryRow(query, id)
	booking := &domain.Booking{}
	err := row.Scan(
		&booking.ID,
		&booking.TimeSlotID,
		&booking.TherapistID,
		&booking.ClientID,
		&booking.StartTime,
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

func (r *BookingRepository) Create(booking *domain.Booking) error {
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
		INSERT INTO bookings (id, timeslot_id, therapist_id, client_id, start_time, state, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`
	_, err := r.db.Exec(
		query,
		booking.ID,
		booking.TimeSlotID,
		booking.TherapistID,
		booking.ClientID,
		booking.StartTime,
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

func (r *BookingRepository) Update(booking *domain.Booking) error {
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
		SET timeslot_id = ?, therapist_id = ?, client_id = ?, start_time = ?, state = ?, updated_at = ?
		WHERE id = ?
	`
	result, err := r.db.Exec(
		query,
		booking.TimeSlotID,
		booking.TherapistID,
		booking.ClientID,
		booking.StartTime,
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

func (r *BookingRepository) ListByTherapist(therapistID domain.TherapistID) ([]*domain.Booking, error) {
	if therapistID == "" {
		return nil, ErrBookingTherapistIDIsRequired
	}

	query := `
		SELECT id, timeslot_id, therapist_id, client_id, start_time, state, created_at, updated_at
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

func (r *BookingRepository) ListByClient(clientID domain.ClientID) ([]*domain.Booking, error) {
	if clientID == "" {
		return nil, ErrBookingClientIDIsRequired
	}

	query := `
		SELECT id, timeslot_id, therapist_id, client_id, start_time, state, created_at, updated_at
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

func (r *BookingRepository) ListByState(state domain.BookingState) ([]*domain.Booking, error) {
	if state == "" {
		return nil, ErrBookingStateIsRequired
	}

	query := `
		SELECT id, timeslot_id, therapist_id, client_id, start_time, state, created_at, updated_at
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

func (r *BookingRepository) ListByTherapistAndState(therapistID domain.TherapistID, state domain.BookingState) ([]*domain.Booking, error) {
	if therapistID == "" {
		return nil, ErrBookingTherapistIDIsRequired
	}

	if state == "" {
		return nil, ErrBookingStateIsRequired
	}

	query := `
		SELECT id, timeslot_id, therapist_id, client_id, start_time, state, created_at, updated_at
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

func (r *BookingRepository) ListByClientAndState(clientID domain.ClientID, state domain.BookingState) ([]*domain.Booking, error) {
	if clientID == "" {
		return nil, ErrBookingClientIDIsRequired
	}

	if state == "" {
		return nil, ErrBookingStateIsRequired
	}

	query := `
		SELECT id, timeslot_id, therapist_id, client_id, start_time, state, created_at, updated_at
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
func (r *BookingRepository) scanBookings(rows *sql.Rows) ([]*domain.Booking, error) {
	bookings := make([]*domain.Booking, 0)
	for rows.Next() {
		booking := &domain.Booking{}
		err := rows.Scan(
			&booking.ID,
			&booking.TimeSlotID,
			&booking.TherapistID,
			&booking.ClientID,
			&booking.StartTime,
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
