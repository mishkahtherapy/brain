package therapist

import (
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"github.com/mishkahtherapy/brain/adapters/db"
	"github.com/mishkahtherapy/brain/core/domain"
)

type TherapistRepository struct {
	db db.SQLDatabase
}

var ErrTherapistNotFound = errors.New("therapist not found")
var ErrTherapistAlreadyExists = errors.New("therapist already exists")
var ErrTherapistNameIsRequired = errors.New("therapist name is required")
var ErrTherapistEmailIsRequired = errors.New("therapist email is required")
var ErrTherapistCreatedAtIsRequired = errors.New("therapist created at is required")
var ErrTherapistUpdatedAtIsRequired = errors.New("therapist updated at is required")
var ErrTherapistIDIsRequired = errors.New("therapist id is required")
var ErrFailedToGetTherapists = errors.New("failed to get therapists")
var ErrFailedToCreateTherapist = errors.New("failed to create therapist")
var ErrFailedToUpdateTherapistSpecializations = errors.New("failed to update therapist specializations")

func NewTherapistRepository(db db.SQLDatabase) *TherapistRepository {
	return &TherapistRepository{db: db}
}

func (r *TherapistRepository) Create(therapist *domain.Therapist) error {
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
		INSERT INTO therapists (id, name, email, phone_number, whatsapp_number, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`
	_, err = tx.Exec(
		query,
		therapist.ID,
		therapist.Name,
		therapist.Email,
		therapist.PhoneNumber,
		therapist.WhatsAppNumber,
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

func (r *TherapistRepository) GetByID(id domain.TherapistID) (*domain.Therapist, error) {
	query := `
		SELECT id, name, email, phone_number, whatsapp_number, created_at, updated_at
		FROM therapists
		WHERE id = ?
	`
	row := r.db.QueryRow(query, id)
	therapist := &domain.Therapist{}
	err := row.Scan(
		&therapist.ID,
		&therapist.Name,
		&therapist.Email,
		&therapist.PhoneNumber,
		&therapist.WhatsAppNumber,
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

	// Load specializations
	specializations, err := r.bulkGetTherapistSpecializations([]domain.TherapistID{id})
	if err != nil {
		return nil, err
	}
	therapist.Specializations = specializations[id]

	// Initialize empty slices for TimeSlotIDs and BookingIDs
	therapist.TimeSlots = []domain.TimeSlot{}

	return therapist, nil
}

func (r *TherapistRepository) GetByEmail(email domain.Email) (*domain.Therapist, error) {
	query := `
		SELECT id, name, email, phone_number, whatsapp_number, created_at, updated_at
		FROM therapists
		WHERE email = ?
	`
	row := r.db.QueryRow(query, email)
	therapist := &domain.Therapist{}
	err := row.Scan(
		&therapist.ID,
		&therapist.Name,
		&therapist.Email,
		&therapist.PhoneNumber,
		&therapist.WhatsAppNumber,
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

	// Initialize empty slices for TimeSlotIDs and BookingIDs
	therapist.TimeSlots = []domain.TimeSlot{}

	return therapist, nil
}

func (r *TherapistRepository) Delete(id domain.TherapistID) error {
	// Due to foreign key constraints, this will cascade delete specializations, time slots, and bookings
	query := `DELETE FROM therapists WHERE id = ?`
	_, err := r.db.Exec(query, id)
	return err
}

func (r *TherapistRepository) List() ([]*domain.Therapist, error) {
	query := `
		SELECT id, name, email, phone_number, whatsapp_number, created_at, updated_at
		FROM therapists
		ORDER BY name ASC
	`
	rows, err := r.db.Query(query)
	if err != nil {
		slog.Error("error getting all therapists", "error", err)
		return nil, ErrFailedToGetTherapists
	}
	defer rows.Close()

	therapists := make([]*domain.Therapist, 0)
	for rows.Next() {
		therapist := &domain.Therapist{}
		err := rows.Scan(
			&therapist.ID,
			&therapist.Name,
			&therapist.Email,
			&therapist.PhoneNumber,
			&therapist.WhatsAppNumber,
			&therapist.CreatedAt,
			&therapist.UpdatedAt,
		)
		if err != nil {
			slog.Error("error scanning therapist", "error", err)
			return nil, ErrFailedToGetTherapists
		}

		// Load specializations for each therapist
		specializations, err := r.bulkGetTherapistSpecializations([]domain.TherapistID{therapist.ID})
		if err != nil {
			return nil, ErrFailedToGetTherapists
		}
		therapist.Specializations = specializations[therapist.ID]

		// Initialize empty slices for TimeSlotIDs and BookingIDs
		therapist.TimeSlots = []domain.TimeSlot{}

		therapists = append(therapists, therapist)
	}
	return therapists, nil
}

// Helper methods for managing therapist specializations

func (r *TherapistRepository) insertTherapistSpecializations(tx db.SQLTx, therapistID domain.TherapistID, specializationIDs []domain.SpecializationID) error {
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

func (r *TherapistRepository) bulkGetTherapistSpecializations(therapistIDs []domain.TherapistID) (map[domain.TherapistID][]domain.Specialization, error) {
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

	specializations := make(map[domain.TherapistID][]domain.Specialization)
	for rows.Next() {
		var therapistID domain.TherapistID
		var specID domain.Specialization
		err := rows.Scan(&therapistID, &specID.ID, &specID.Name, &specID.CreatedAt, &specID.UpdatedAt)
		if err != nil {
			slog.Error("error scanning specialization id", "error", err)
			return nil, ErrFailedToGetTherapists
		}
		specializations[therapistID] = append(specializations[therapistID], specID)
	}
	return specializations, nil
}
