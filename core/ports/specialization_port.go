package ports

import (
	"github.com/mishkahtherapy/brain/core/domain"
	"github.com/mishkahtherapy/brain/core/domain/specialization"
)

type SpecializationRepository interface {
	Create(specialization *specialization.Specialization) error
	GetByID(id domain.SpecializationID) (*specialization.Specialization, error)
	GetByName(name string) (*specialization.Specialization, error)
	BulkGetByIds(ids []domain.SpecializationID) (map[domain.SpecializationID]*specialization.Specialization, error)
	GetAll() ([]*specialization.Specialization, error)
}
