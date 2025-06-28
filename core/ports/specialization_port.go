package ports

import "github.com/mishkahtherapy/brain/core/domain"

type SpecializationRepository interface {
	Create(specialization *domain.Specialization) error
	GetByID(id domain.SpecializationID) (*domain.Specialization, error)
	GetByName(name string) (*domain.Specialization, error)
	BulkGetByIds(ids []domain.SpecializationID) (map[domain.SpecializationID]*domain.Specialization, error)
	GetAll() ([]*domain.Specialization, error)
}
