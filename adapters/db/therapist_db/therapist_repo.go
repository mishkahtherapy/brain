package therapist_db

import (
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"github.com/mishkahtherapy/brain/core/domain"
	"github.com/mishkahtherapy/brain/core/domain/specialization"
	"github.com/mishkahtherapy/brain/core/domain/therapist"
	"github.com/mishkahtherapy/brain/core/ports"
)

type TherapistRepository struct {
	db ports.SQLDatabase
}

var ErrTherapistNotFound = errors.New("therapist not found")
var ErrTherapistAlreadyExists = errors.New("therapist already exists")
var ErrTherapistNameIsRequired = errors.New("therapist name is required")
var ErrTherapistEmailIsRequired = errors.New("therapist email is required")
var ErrTherapistCreatedAtIsRequired = errors.New("therapist created at is required")
var ErrTherapistUpdatedAtIsRequired = errors.New("therapist updated at is required")
var ErrTherapistIDIsRequired = errors.New("therapist id is required")
var ErrTherapistTimezoneOffsetIsRequired = errors.New("therapist timezone offset is required")
var ErrFailedToGetTherapists = errors.New("failed to get therapists")
var ErrFailedToCreateTherapist = errors.New("failed to create therapist")
var ErrFailedToUpdateTherapist = errors.New("failed to update therapist")
var ErrFailedToUpdateTherapistSpecializations = errors.New("failed to update therapist specializations")
var ErrDeviceIDIsRequired = errors.New("device id is required")

func NewTherapistRepository(db ports.SQLDatabase) ports.TherapistRepository {
	return &TherapistRepository{db: db}
}

func (r *TherapistRepository) Create(therapist *therapist.Therapist) error {
	if therapist.ID == "" {
		return ErrTherapistIDIsRequired
	}

	if therapist.Name == "" {
		return ErrTherapistNameIsRequired
	}

	if therapist.Email == "" {
		return ErrTherapistEmailIsRequired
	}

	if therapist.CreatedAt == (domain.UTCTimestamp{}) {
		return ErrTherapistCreatedAtIsRequired
	}

	if therapist.UpdatedAt == (domain.UTCTimestamp{}) {
		return ErrTherapistUpdatedAtIsRequired
	}

	tx, err := r.db.Begin()
	if err != nil {
		slog.Error("error beginning create therapist transaction", "error", err)
		return ErrFailedToCreateTherapist
	}

	// Insert therapist
	query := `
		INSERT INTO therapists (id, name, email, phone_number, whatsapp_number, speaks_english, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`
	_, err = tx.Exec(
		query,
		therapist.ID,
		therapist.Name,
		therapist.Email,
		therapist.PhoneNumber,
		therapist.WhatsAppNumber,
		therapist.SpeaksEnglish,
		therapist.CreatedAt,
		therapist.UpdatedAt,
	)
	if err != nil {
		tx.Rollback()
		slog.Error("error inserting therapist", "error", err)
		return ErrFailedToCreateTherapist
	}

	specializationIDs := make([]domain.SpecializationID, 0)
	for _, specialization := range therapist.Specializations {
		specializationIDs = append(specializationIDs, specialization.ID)
	}

	// Insert specializations
	err = r.insertTherapistSpecializations(tx, therapist.ID, specializationIDs)
	if err != nil {
		tx.Rollback()
		slog.Error("error inserting therapist specializations", "error", err)
		return ErrFailedToCreateTherapist
	}

	if err := tx.Commit(); err != nil {
		slog.Error("error committing create therapist transaction", "error", err)
		return ErrFailedToCreateTherapist
	}

	return nil
}

func (r *TherapistRepository) Update(therapist *therapist.Therapist) error {
	if therapist.ID == "" {
		return ErrTherapistIDIsRequired
	}

	if therapist.Name == "" {
		return ErrTherapistNameIsRequired
	}

	if therapist.Email == "" {
		return ErrTherapistEmailIsRequired
	}

	if therapist.UpdatedAt == (domain.UTCTimestamp{}) {
		return ErrTherapistUpdatedAtIsRequired
	}

	query := `
		UPDATE therapists 
		SET name = ?, email = ?, phone_number = ?, whatsapp_number = ?, speaks_english = ?, updated_at = ?
		WHERE id = ?
	`
	result, err := r.db.Exec(
		query,
		therapist.Name,
		therapist.Email,
		therapist.PhoneNumber,
		therapist.WhatsAppNumber,
		therapist.SpeaksEnglish,
		therapist.UpdatedAt,
		therapist.ID,
	)
	if err != nil {
		slog.Error("error updating therapist", "error", err)
		return ErrFailedToUpdateTherapist
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		slog.Error("error getting rows affected after update", "error", err)
		return ErrFailedToUpdateTherapist
	}

	if rowsAffected == 0 {
		return ErrTherapistNotFound
	}

	return nil
}

func (r *TherapistRepository) UpdateSpecializations(therapistID domain.TherapistID, specializationIDs []domain.SpecializationID) error {

	if therapistID == "" {
		return ErrTherapistIDIsRequired
	}

	tx, err := r.db.Begin()
	if err != nil {
		slog.Error("error beginning update therapist specializations transaction", "error", err)
		return ErrFailedToUpdateTherapistSpecializations
	}

	// Delete existing specializations
	query := `DELETE FROM therapist_specializations WHERE therapist_id = ?`
	_, err = tx.Exec(query, therapistID)
	if err != nil {
		tx.Rollback()
		slog.Error("error deleting existing therapist specializations", "error", err)
		return ErrFailedToUpdateTherapistSpecializations
	}

	// Insert new specializations
	err = r.insertTherapistSpecializations(tx, therapistID, specializationIDs)
	if err != nil {
		tx.Rollback()
		slog.Error("error inserting new therapist specializations", "error", err)
		return ErrFailedToUpdateTherapistSpecializations
	}

	if err := tx.Commit(); err != nil {
		slog.Error("error committing transaction", "error", err)
		return ErrFailedToUpdateTherapistSpecializations
	}

	return nil
}

func (r *TherapistRepository) UpdateDevice(therapistID domain.TherapistID, deviceID domain.DeviceID, deviceIDUpdatedAt domain.UTCTimestamp) error {
	if therapistID == "" {
		return ErrTherapistIDIsRequired
	}

	if deviceID == "" {
		return ErrDeviceIDIsRequired
	}

	query := `UPDATE therapists SET device_id = ?, device_id_updated_at = ? WHERE id = ?`
	_, err := r.db.Exec(query, deviceID, deviceIDUpdatedAt, therapistID)
	if err != nil {
		slog.Error("error updating therapist device", "error", err)
		return ErrFailedToUpdateTherapist
	}

	return nil
}

func (r *TherapistRepository) GetDevice(therapistID domain.TherapistID) (domain.DeviceID, error) {
	query := `SELECT device_id FROM therapists WHERE id = ? LIMIT 1`
	row := r.db.QueryRow(query, therapistID)
	var deviceID domain.DeviceID
	err := row.Scan(&deviceID)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", ErrTherapistNotFound
		}
		slog.Error("error getting therapist device", "error", err)
		return "", ErrFailedToGetTherapists
	}
	return deviceID, err
}

func (r *TherapistRepository) BulkGetDevices(therapistIDs []domain.TherapistID) (map[domain.TherapistID]domain.DeviceID, error) {
	query := `SELECT id, device_id FROM therapists WHERE id IN (%s)`

	placeholders := make([]string, 0)
	values := make([]interface{}, 0)

	for _, therapistID := range therapistIDs {
		placeholders = append(placeholders, "?")
		values = append(values, therapistID)
	}

	query = fmt.Sprintf(query, strings.Join(placeholders, ", "))
	rows, err := r.db.Query(query, values...)
	if err != nil {
		slog.Error("error getting therapist devices", "error", err)
		return nil, ErrFailedToGetTherapists
	}

	devices := make(map[domain.TherapistID]domain.DeviceID)
	for rows.Next() {
		var therapistID domain.TherapistID
		var deviceID sql.NullString
		err := rows.Scan(&therapistID, &deviceID)
		if err != nil {
			slog.Error("error scanning therapist device", "error", err)
			return nil, ErrFailedToGetTherapists
		}
		if deviceID.Valid {
			devices[therapistID] = domain.DeviceID(deviceID.String)
		}
	}
	return devices, nil
}

func (r *TherapistRepository) UpdateTimezoneOffset(therapistID domain.TherapistID, timezoneOffset domain.TimezoneOffset) error {
	if therapistID == "" {
		return ErrTherapistIDIsRequired
	}

	query := `UPDATE therapists SET timezone_offset = ? WHERE id = ?`
	_, err := r.db.Exec(query, timezoneOffset, therapistID)
	if err != nil {
		slog.Error("error updating therapist timezone offset", "error", err)
		return ErrFailedToUpdateTherapist
	}

	return nil
}

func (r *TherapistRepository) GetByID(id domain.TherapistID) (*therapist.Therapist, error) {
	query := `
		SELECT id, name, email, phone_number, whatsapp_number, speaks_english, device_id, timezone_offset, created_at, updated_at
		FROM therapists
		WHERE id = ?
	`
	row := r.db.QueryRow(query, id)
	therapist := &therapist.Therapist{}
	var deviceID sql.NullString
	err := row.Scan(
		&therapist.ID,
		&therapist.Name,
		&therapist.Email,
		&therapist.PhoneNumber,
		&therapist.WhatsAppNumber,
		&therapist.SpeaksEnglish,
		&deviceID,
		&therapist.TimezoneOffset,
		&therapist.CreatedAt,
		&therapist.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrTherapistNotFound
		}
		slog.Error("error getting therapist by id", "error", err)
		return nil, ErrFailedToGetTherapists
	}

	if deviceID.Valid {
		therapist.DeviceID = domain.DeviceID(deviceID.String)
	}

	// Load specializations
	specializations, err := r.bulkGetTherapistSpecializations([]domain.TherapistID{id})
	if err != nil {
		return nil, err
	}

	therapist.Specializations = specializations[id]
	return therapist, nil
}

func (r *TherapistRepository) GetByEmail(email domain.Email) (*therapist.Therapist, error) {
	query := `
		SELECT id, name, email, phone_number, whatsapp_number, speaks_english, timezone_offset, created_at, updated_at
		FROM therapists
		WHERE email = ?
	`
	row := r.db.QueryRow(query, email)
	therapist := &therapist.Therapist{}
	err := row.Scan(
		&therapist.ID,
		&therapist.Name,
		&therapist.Email,
		&therapist.PhoneNumber,
		&therapist.WhatsAppNumber,
		&therapist.SpeaksEnglish,
		&therapist.DeviceID,
		&therapist.TimezoneOffset,
		&therapist.CreatedAt,
		&therapist.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		slog.Error("error getting therapist by email", "error", err)
		return nil, ErrFailedToGetTherapists
	}

	// Load specializations
	specializations, err := r.bulkGetTherapistSpecializations([]domain.TherapistID{therapist.ID})
	if err != nil {
		return nil, ErrFailedToGetTherapists
	}

	therapist.Specializations = specializations[therapist.ID]
	return therapist, nil
}

func (r *TherapistRepository) GetByWhatsAppNumber(whatsappNumber domain.WhatsAppNumber) (*therapist.Therapist, error) {
	query := `
		SELECT id, name, email, phone_number, whatsapp_number, speaks_english, device_id, timezone_offset, created_at, updated_at
		FROM therapists
		WHERE whatsapp_number = ?
	`
	row := r.db.QueryRow(query, whatsappNumber)
	therapist := &therapist.Therapist{}
	err := row.Scan(
		&therapist.ID,
		&therapist.Name,
		&therapist.Email,
		&therapist.PhoneNumber,
		&therapist.WhatsAppNumber,
		&therapist.SpeaksEnglish,
		&therapist.DeviceID,
		&therapist.TimezoneOffset,
		&therapist.CreatedAt,
		&therapist.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		slog.Error("error getting therapist by whatsapp number", "error", err)
		return nil, ErrFailedToGetTherapists
	}

	// Load specializations
	specializations, err := r.bulkGetTherapistSpecializations([]domain.TherapistID{therapist.ID})
	if err != nil {
		return nil, ErrFailedToGetTherapists
	}

	therapist.Specializations = specializations[therapist.ID]
	return therapist, nil
}

func (r *TherapistRepository) Delete(id domain.TherapistID) error {
	// Due to foreign key constraints, this will cascade delete specializations, time slots, and bookings
	query := `DELETE FROM therapists WHERE id = ?`
	_, err := r.db.Exec(query, id)
	return err
}

func (r *TherapistRepository) List() ([]*therapist.Therapist, error) {
	query := `
		SELECT id, name, email, phone_number, whatsapp_number, speaks_english, device_id, timezone_offset, created_at, updated_at
		FROM therapists
		ORDER BY name ASC
	`
	rows, err := r.db.Query(query)
	if err != nil {
		slog.Error("error getting all therapists", "error", err)
		return nil, ErrFailedToGetTherapists
	}
	defer rows.Close()

	therapists := make([]*therapist.Therapist, 0)
	var deviceID sql.NullString
	for rows.Next() {
		therapist := &therapist.Therapist{}
		err := rows.Scan(
			&therapist.ID,
			&therapist.Name,
			&therapist.Email,
			&therapist.PhoneNumber,
			&therapist.WhatsAppNumber,
			&therapist.SpeaksEnglish,
			&deviceID,
			&therapist.TimezoneOffset,
			&therapist.CreatedAt,
			&therapist.UpdatedAt,
		)
		if err != nil {
			slog.Error("error scanning therapist", "error", err)
			return nil, ErrFailedToGetTherapists
		}

		if deviceID.Valid {
			therapist.DeviceID = domain.DeviceID(deviceID.String)
		}

		// Load specializations for each therapist
		specializations, err := r.bulkGetTherapistSpecializations([]domain.TherapistID{therapist.ID})
		if err != nil {
			return nil, ErrFailedToGetTherapists
		}

		therapist.Specializations = specializations[therapist.ID]
		therapists = append(therapists, therapist)
	}

	return therapists, nil
}

func (r *TherapistRepository) FindBySpecializationAndLanguage(specializationName string, mustSpeakEnglish bool) ([]*therapist.Therapist, error) {
	query := `
	       SELECT DISTINCT t.id, t.name, t.email, t.phone_number, t.whatsapp_number, t.speaks_english, t.device_id, t.timezone_offset, t.created_at, t.updated_at
	       FROM therapists t
	       JOIN therapist_specializations ts ON t.id = ts.therapist_id
	       JOIN specializations s ON ts.specialization_id = s.id
	       WHERE s.name = ?
	   `

	args := []interface{}{specializationName}

	if mustSpeakEnglish {
		query += " AND t.speaks_english = TRUE"
	}

	query += " ORDER BY t.name ASC"

	rows, err := r.db.Query(query, args...)
	if err != nil {
		slog.Error("error finding therapists by specialization and language", "error", err)
		return nil, ErrFailedToGetTherapists
	}
	defer rows.Close()

	therapists := make([]*therapist.Therapist, 0)
	therapistIDs := make([]domain.TherapistID, 0)
	var deviceID sql.NullString

	for rows.Next() {
		therapist := &therapist.Therapist{}
		err := rows.Scan(
			&therapist.ID,
			&therapist.Name,
			&therapist.Email,
			&therapist.PhoneNumber,
			&therapist.WhatsAppNumber,
			&therapist.SpeaksEnglish,
			&deviceID,
			&therapist.TimezoneOffset,
			&therapist.CreatedAt,
			&therapist.UpdatedAt,
		)
		if err != nil {
			slog.Error("error scanning therapist", "error", err)
			return nil, ErrFailedToGetTherapists
		}

		if deviceID.Valid {
			therapist.DeviceID = domain.DeviceID(deviceID.String)
		}

		therapists = append(therapists, therapist)
		therapistIDs = append(therapistIDs, therapist.ID)
	}

	// Load specializations for each therapist
	specializations, err := r.bulkGetTherapistSpecializations(therapistIDs)
	if err != nil {
		return nil, ErrFailedToGetTherapists
	}

	for _, therapist := range therapists {
		therapist.Specializations = specializations[therapist.ID]
	}

	return therapists, nil
}

func (r *TherapistRepository) FindByIDs(therapistIDs []domain.TherapistID) ([]*therapist.Therapist, error) {
	if len(therapistIDs) == 0 {
		return nil, nil
	}

	query := `
		SELECT id, name, email, phone_number, whatsapp_number, speaks_english, device_id, timezone_offset, created_at, updated_at
		FROM therapists
		WHERE id IN (%s)
	`

	placeholders := make([]string, 0)
	values := make([]interface{}, 0)

	for _, therapistID := range therapistIDs {
		placeholders = append(placeholders, "?")
		values = append(values, therapistID)
	}

	query = fmt.Sprintf(query, strings.Join(placeholders, ", "))
	rows, err := r.db.Query(query, values...)
	if err != nil {
		slog.Error("error finding therapists by ids", "error", err)
		return nil, ErrFailedToGetTherapists
	}
	defer rows.Close()

	therapists := make([]*therapist.Therapist, 0)
	var deviceID sql.NullString
	for rows.Next() {
		therapist := &therapist.Therapist{}
		err := rows.Scan(
			&therapist.ID,
			&therapist.Name,
			&therapist.Email,
			&therapist.PhoneNumber,
			&therapist.WhatsAppNumber,
			&therapist.SpeaksEnglish,
			&deviceID,
			&therapist.TimezoneOffset,
			&therapist.CreatedAt,
			&therapist.UpdatedAt,
		)
		if err != nil {
			slog.Error("error scanning therapist", "error", err)
			return nil, ErrFailedToGetTherapists
		}

		if deviceID.Valid {
			therapist.DeviceID = domain.DeviceID(deviceID.String)
		}

		therapists = append(therapists, therapist)
	}

	// Load specializations for each therapist
	specializations, err := r.bulkGetTherapistSpecializations(therapistIDs)
	if err != nil {
		return nil, ErrFailedToGetTherapists
	}

	for _, therapist := range therapists {
		therapist.Specializations = specializations[therapist.ID]
	}

	return therapists, nil
}

// Helper methods for managing therapist specializations

func (r *TherapistRepository) insertTherapistSpecializations(tx ports.SQLTx, therapistID domain.TherapistID, specializationIDs []domain.SpecializationID) error {
	if len(specializationIDs) == 0 {
		return nil
	}

	// Build bulk insert query
	placeholders := make([]string, 0)
	values := make([]interface{}, 0)

	timestamp := domain.NewUTCTimestamp()

	for i, specID := range specializationIDs {
		placeholders = append(placeholders, "(?, ?, ?, ?, ?)")
		values = append(values, fmt.Sprintf("therapist_spec_%s_%d", therapistID, i))
		values = append(values, therapistID)
		values = append(values, specID)
		values = append(values, timestamp)
		values = append(values, timestamp)
	}

	query := fmt.Sprintf(`
		INSERT INTO therapist_specializations (id, therapist_id, specialization_id, created_at, updated_at)
		VALUES %s
	`, strings.Join(placeholders, ", "))

	_, err := tx.Exec(query, values...)
	return err
}

func (r *TherapistRepository) bulkGetTherapistSpecializations(therapistIDs []domain.TherapistID) (map[domain.TherapistID][]specialization.Specialization, error) {
	query := `
		SELECT 
			therapist_specializations.therapist_id,
			specializations.id,
			specializations.name,
			specializations.created_at,
			specializations.updated_at
		FROM therapist_specializations
		JOIN specializations ON therapist_specializations.specialization_id = specializations.id
		WHERE therapist_id IN (%s)
		ORDER BY specializations.name ASC
	`

	placeholders := make([]string, 0)
	values := make([]interface{}, 0)

	for _, therapistID := range therapistIDs {
		placeholders = append(placeholders, "?")
		values = append(values, therapistID)
	}

	query = fmt.Sprintf(query, strings.Join(placeholders, ", "))
	rows, err := r.db.Query(query, values...)
	if err != nil {
		slog.Error("error getting therapist specializations", "error", err)
		return nil, ErrFailedToGetTherapists
	}
	defer rows.Close()

	specializations := make(map[domain.TherapistID][]specialization.Specialization)
	for rows.Next() {
		var therapistID domain.TherapistID
		var specID specialization.Specialization
		err := rows.Scan(&therapistID, &specID.ID, &specID.Name, &specID.CreatedAt, &specID.UpdatedAt)
		if err != nil {
			slog.Error("error scanning specialization id", "error", err)
			return nil, ErrFailedToGetTherapists
		}
		specializations[therapistID] = append(specializations[therapistID], specID)
	}
	return specializations, nil
}
