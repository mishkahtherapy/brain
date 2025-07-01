package timeslot_db

import (
	"database/sql"
	"errors"
	"log/slog"

	"github.com/mishkahtherapy/brain/core/domain"
	"github.com/mishkahtherapy/brain/core/domain/timeslot"
	"github.com/mishkahtherapy/brain/core/ports"
)

type TimeSlotRepository struct {
	db ports.SQLDatabase
}

// Error definitions
var ErrTimeSlotNotFound = errors.New("timeslot not found")
var ErrTimeSlotAlreadyExists = errors.New("timeslot already exists")
var ErrTimeSlotIDIsRequired = errors.New("timeslot id is required")
var ErrTimeSlotTherapistIDIsRequired = errors.New("timeslot therapist id is required")
var ErrTimeSlotDayOfWeekIsRequired = errors.New("timeslot day of week is required")
var ErrTimeSlotStartTimeIsRequired = errors.New("timeslot start time is required")
var ErrTimeSlotEndTimeIsRequired = errors.New("timeslot end time is required")
var ErrTimeSlotCreatedAtIsRequired = errors.New("timeslot created at is required")
var ErrTimeSlotUpdatedAtIsRequired = errors.New("timeslot updated at is required")
var ErrFailedToGetTimeSlots = errors.New("failed to get timeslots")
var ErrFailedToCreateTimeSlot = errors.New("failed to create timeslot")
var ErrFailedToUpdateTimeSlot = errors.New("failed to update timeslot")
var ErrFailedToDeleteTimeSlot = errors.New("failed to delete timeslot")

func NewTimeSlotRepository(db ports.SQLDatabase) *TimeSlotRepository {
	return &TimeSlotRepository{db: db}
}

func (r *TimeSlotRepository) GetByID(id string) (*timeslot.TimeSlot, error) {
	query := `
		SELECT id, therapist_id, day_of_week, start_time, end_time,
		       pre_session_buffer, post_session_buffer, created_at, updated_at
		FROM time_slots
		WHERE id = ?
	`
	row := r.db.QueryRow(query, id)
	timeslot := &timeslot.TimeSlot{}
	err := row.Scan(
		&timeslot.ID,
		&timeslot.TherapistID,
		&timeslot.DayOfWeek,
		&timeslot.StartTime,
		&timeslot.EndTime,
		&timeslot.PreSessionBuffer,
		&timeslot.PostSessionBuffer,
		&timeslot.CreatedAt,
		&timeslot.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrTimeSlotNotFound
		}
		slog.Error("error getting timeslot by id", "error", err)
		return nil, ErrFailedToGetTimeSlots
	}

	// Get bookings associated with this timeslot
	bookingQuery := `
		SELECT id FROM bookings WHERE timeslot_id = ?
	`
	rows, err := r.db.Query(bookingQuery, id)
	if err != nil {
		slog.Error("error getting bookings for timeslot", "error", err)
		return nil, ErrFailedToGetTimeSlots
	}
	defer rows.Close()

	timeslot.BookingIDs = make([]domain.BookingID, 0)
	for rows.Next() {
		var bookingID domain.BookingID
		if err := rows.Scan(&bookingID); err != nil {
			slog.Error("error scanning booking id", "error", err)
			return nil, ErrFailedToGetTimeSlots
		}
		timeslot.BookingIDs = append(timeslot.BookingIDs, bookingID)
	}

	return timeslot, nil
}

func (r *TimeSlotRepository) Create(timeslot *timeslot.TimeSlot) error {
	// Validate required fields
	if timeslot.ID == "" {
		return ErrTimeSlotIDIsRequired
	}

	if timeslot.TherapistID == "" {
		return ErrTimeSlotTherapistIDIsRequired
	}

	if timeslot.DayOfWeek == "" {
		return ErrTimeSlotDayOfWeekIsRequired
	}

	if timeslot.StartTime == "" {
		return ErrTimeSlotStartTimeIsRequired
	}

	if timeslot.EndTime == "" {
		return ErrTimeSlotEndTimeIsRequired
	}

	if timeslot.CreatedAt == (domain.UTCTimestamp{}) {
		return ErrTimeSlotCreatedAtIsRequired
	}

	if timeslot.UpdatedAt == (domain.UTCTimestamp{}) {
		return ErrTimeSlotUpdatedAtIsRequired
	}

	query := `
		INSERT INTO time_slots (
			id, therapist_id, day_of_week, start_time, end_time,
			pre_session_buffer, post_session_buffer, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	_, err := r.db.Exec(
		query,
		timeslot.ID,
		timeslot.TherapistID,
		timeslot.DayOfWeek,
		timeslot.StartTime,
		timeslot.EndTime,
		timeslot.PreSessionBuffer,
		timeslot.PostSessionBuffer,
		timeslot.CreatedAt,
		timeslot.UpdatedAt,
	)
	if err != nil {
		slog.Error("error creating timeslot", "error", err)
		return ErrFailedToCreateTimeSlot
	}
	return nil
}

func (r *TimeSlotRepository) Update(timeslot *timeslot.TimeSlot) error {
	// Validate required fields
	if timeslot.ID == "" {
		return ErrTimeSlotIDIsRequired
	}

	if timeslot.TherapistID == "" {
		return ErrTimeSlotTherapistIDIsRequired
	}

	if timeslot.DayOfWeek == "" {
		return ErrTimeSlotDayOfWeekIsRequired
	}

	if timeslot.StartTime == "" {
		return ErrTimeSlotStartTimeIsRequired
	}

	if timeslot.EndTime == "" {
		return ErrTimeSlotEndTimeIsRequired
	}

	if timeslot.UpdatedAt == (domain.UTCTimestamp{}) {
		return ErrTimeSlotUpdatedAtIsRequired
	}

	query := `
		UPDATE time_slots
		SET therapist_id = ?, day_of_week = ?, start_time = ?, end_time = ?,
		    pre_session_buffer = ?, post_session_buffer = ?, updated_at = ?
		WHERE id = ?
	`
	result, err := r.db.Exec(
		query,
		timeslot.TherapistID,
		timeslot.DayOfWeek,
		timeslot.StartTime,
		timeslot.EndTime,
		timeslot.PreSessionBuffer,
		timeslot.PostSessionBuffer,
		timeslot.UpdatedAt,
		timeslot.ID,
	)
	if err != nil {
		slog.Error("error updating timeslot", "error", err)
		return ErrFailedToUpdateTimeSlot
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		slog.Error("error getting rows affected after update", "error", err)
		return ErrFailedToUpdateTimeSlot
	}

	if rowsAffected == 0 {
		return ErrTimeSlotNotFound
	}

	return nil
}

func (r *TimeSlotRepository) Delete(id string) error {
	if id == "" {
		return ErrTimeSlotIDIsRequired
	}

	query := `DELETE FROM time_slots WHERE id = ?`
	result, err := r.db.Exec(query, id)
	if err != nil {
		slog.Error("error deleting timeslot", "error", err)
		return ErrFailedToDeleteTimeSlot
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		slog.Error("error getting rows affected after delete", "error", err)
		return ErrFailedToDeleteTimeSlot
	}

	if rowsAffected == 0 {
		return ErrTimeSlotNotFound
	}

	return nil
}

func (r *TimeSlotRepository) ListByTherapist(therapistID string) ([]*timeslot.TimeSlot, error) {
	if therapistID == "" {
		return nil, ErrTimeSlotTherapistIDIsRequired
	}

	query := `
		SELECT id, therapist_id, day_of_week, start_time, end_time,
		       pre_session_buffer, post_session_buffer, created_at, updated_at
		FROM time_slots
		WHERE therapist_id = ?
		ORDER BY day_of_week, start_time
	`
	rows, err := r.db.Query(query, therapistID)
	if err != nil {
		slog.Error("error listing timeslots by therapist", "error", err)
		return nil, ErrFailedToGetTimeSlots
	}
	defer rows.Close()

	return r.scanTimeSlotsWithoutBookings(rows)
}

func (r *TimeSlotRepository) ListByDay(therapistID string, day string) ([]*timeslot.TimeSlot, error) {
	if therapistID == "" {
		return nil, ErrTimeSlotTherapistIDIsRequired
	}

	if day == "" {
		return nil, ErrTimeSlotDayOfWeekIsRequired
	}

	query := `
		SELECT id, therapist_id, day_of_week, start_time, end_time,
		       pre_session_buffer, post_session_buffer, created_at, updated_at
		FROM time_slots
		WHERE therapist_id = ? AND day_of_week = ?
		ORDER BY start_time
	`
	rows, err := r.db.Query(query, therapistID, day)
	if err != nil {
		slog.Error("error listing timeslots by day", "error", err)
		return nil, ErrFailedToGetTimeSlots
	}
	defer rows.Close()

	return r.scanTimeSlots(rows)
}

// Helper method to scan multiple timeslot rows
func (r *TimeSlotRepository) scanTimeSlots(rows *sql.Rows) ([]*timeslot.TimeSlot, error) {
	timeslots := make([]*timeslot.TimeSlot, 0)
	for rows.Next() {
		timeslot := &timeslot.TimeSlot{}
		err := rows.Scan(
			&timeslot.ID,
			&timeslot.TherapistID,
			&timeslot.DayOfWeek,
			&timeslot.StartTime,
			&timeslot.EndTime,
			&timeslot.PreSessionBuffer,
			&timeslot.PostSessionBuffer,
			&timeslot.CreatedAt,
			&timeslot.UpdatedAt,
		)
		if err != nil {
			slog.Error("error scanning timeslot", "error", err)
			return nil, ErrFailedToGetTimeSlots
		}

		// Get bookings for this timeslot
		bookingQuery := `SELECT id FROM bookings WHERE timeslot_id = ?`
		bookingRows, err := r.db.Query(bookingQuery, timeslot.ID)
		if err != nil {
			slog.Error("error getting bookings for timeslot", "error", err)
			return nil, ErrFailedToGetTimeSlots
		}

		timeslot.BookingIDs = make([]domain.BookingID, 0)
		for bookingRows.Next() {
			var bookingID domain.BookingID
			if err := bookingRows.Scan(&bookingID); err != nil {
				bookingRows.Close()
				slog.Error("error scanning booking id", "error", err)
				return nil, ErrFailedToGetTimeSlots
			}
			timeslot.BookingIDs = append(timeslot.BookingIDs, bookingID)
		}
		bookingRows.Close()

		timeslots = append(timeslots, timeslot)
	}
	return timeslots, nil
}

// scanTimeSlotsWithoutBookings scans timeslots without fetching booking IDs (for performance)
func (r *TimeSlotRepository) scanTimeSlotsWithoutBookings(rows *sql.Rows) ([]*timeslot.TimeSlot, error) {
	timeslots := make([]*timeslot.TimeSlot, 0)
	for rows.Next() {
		timeslot := &timeslot.TimeSlot{}
		err := rows.Scan(
			&timeslot.ID,
			&timeslot.TherapistID,
			&timeslot.DayOfWeek,
			&timeslot.StartTime,
			&timeslot.EndTime,
			&timeslot.PreSessionBuffer,
			&timeslot.PostSessionBuffer,
			&timeslot.CreatedAt,
			&timeslot.UpdatedAt,
		)
		if err != nil {
			slog.Error("error scanning timeslot", "error", err)
			return nil, ErrFailedToGetTimeSlots
		}

		// Initialize empty BookingIDs instead of querying
		timeslot.BookingIDs = make([]domain.BookingID, 0)
		timeslots = append(timeslots, timeslot)
	}
	return timeslots, nil
}
