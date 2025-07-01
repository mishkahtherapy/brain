package specialization_db

import (
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"github.com/mishkahtherapy/brain/core/domain"
	"github.com/mishkahtherapy/brain/core/domain/specialization"
	"github.com/mishkahtherapy/brain/core/ports"
)

type SpecializationRepository struct {
	db ports.SQLDatabase
}

var ErrSpecializationNotFound = errors.New("specialization not found")
var ErrSpecializationAlreadyExists = errors.New("specialization already exists")
var ErrSpecializationNameIsRequired = errors.New("specialization name is required")
var ErrSpecializationCreatedAtIsRequired = errors.New("specialization created at is required")
var ErrSpecializationUpdatedAtIsRequired = errors.New("specialization updated at is required")
var ErrSpecializationIDIsRequired = errors.New("specialization id is required")
var ErrFailedToGetSpecializations = errors.New("failed to get specializations")

func NewSpecializationRepository(db ports.SQLDatabase) *SpecializationRepository {
	return &SpecializationRepository{db: db}
}

func (r *SpecializationRepository) Create(specialization *specialization.Specialization) error {
	if specialization.ID == "" {
		return ErrSpecializationIDIsRequired
	}

	if specialization.Name == "" {
		return ErrSpecializationNameIsRequired
	}

	if specialization.CreatedAt == (domain.UTCTimestamp{}) {
		return ErrSpecializationCreatedAtIsRequired
	}

	if specialization.UpdatedAt == (domain.UTCTimestamp{}) {
		return ErrSpecializationUpdatedAtIsRequired
	}

	query := `
		INSERT INTO specializations (id, name, created_at, updated_at)
		VALUES (?, ?, ?, ?)
	`
	_, err := r.db.Exec(
		query,
		specialization.ID,
		specialization.Name,
		specialization.CreatedAt,
		specialization.UpdatedAt,
	)
	return err
}

func (r *SpecializationRepository) BulkGetByIds(ids []domain.SpecializationID) (map[domain.SpecializationID]*specialization.Specialization, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	query := `
		SELECT id, name, created_at, updated_at
		FROM specializations
		WHERE id IN (%s)
	`
	placeholders := make([]string, len(ids))
	values := make([]interface{}, len(ids))
	for i := range ids {
		placeholders[i] = "?"
		values[i] = ids[i]
	}

	query = fmt.Sprintf(query, strings.Join(placeholders, ","))
	rows, err := r.db.Query(query, values...)
	if err != nil {
		slog.Error("error getting specializations by ids", "error", err)
		return nil, ErrFailedToGetSpecializations
	}
	defer rows.Close()

	specializations := make(map[domain.SpecializationID]*specialization.Specialization)
	for rows.Next() {
		specialization := &specialization.Specialization{}
		err := rows.Scan(
			&specialization.ID,
			&specialization.Name,
			&specialization.CreatedAt,
			&specialization.UpdatedAt,
		)
		if err != nil {
			slog.Error("error scanning specialization", "error", err)
			return nil, ErrFailedToGetSpecializations
		}
		specializations[specialization.ID] = specialization
	}
	return specializations, nil
}

func (r *SpecializationRepository) GetByID(id domain.SpecializationID) (*specialization.Specialization, error) {
	query := `
		SELECT id, name, created_at, updated_at
		FROM specializations
		WHERE id = ?
	`
	row := r.db.QueryRow(query, id)
	specialization := &specialization.Specialization{}
	err := row.Scan(
		&specialization.ID,
		&specialization.Name,
		&specialization.CreatedAt,
		&specialization.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		slog.Error("error getting specialization by id", "error", err)
		return nil, ErrFailedToGetSpecializations
	}
	return specialization, nil
}

func (r *SpecializationRepository) GetByName(name string) (*specialization.Specialization, error) {
	query := `
		SELECT id, name, created_at, updated_at
		FROM specializations
		WHERE name = ?
	`
	row := r.db.QueryRow(query, name)
	specialization := &specialization.Specialization{}
	err := row.Scan(
		&specialization.ID,
		&specialization.Name,
		&specialization.CreatedAt,
		&specialization.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		slog.Error("error getting specialization by name", "error", err)
		return nil, ErrFailedToGetSpecializations
	}
	return specialization, nil
}

func (r *SpecializationRepository) GetAll() ([]*specialization.Specialization, error) {
	query := `
		SELECT id, name, created_at, updated_at
		FROM specializations
		ORDER BY name ASC
	`
	rows, err := r.db.Query(query)
	if err != nil {
		slog.Error("error getting all specializations", "error", err)
		return nil, ErrFailedToGetSpecializations
	}
	defer rows.Close()

	specializations := make([]*specialization.Specialization, 0)
	for rows.Next() {
		specialization := &specialization.Specialization{}
		err := rows.Scan(
			&specialization.ID,
			&specialization.Name,
			&specialization.CreatedAt,
			&specialization.UpdatedAt,
		)
		if err != nil {
			slog.Error("error scanning specialization", "error", err)
			return nil, ErrFailedToGetSpecializations
		}
		specializations = append(specializations, specialization)
	}
	return specializations, nil
}
