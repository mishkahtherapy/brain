package ports

import "github.com/mishkahtherapy/brain/core/domain"

type TherapistRepository interface {
	GetByID(id string) (*domain.Therapist, error)
	Create(therapist *domain.Therapist) error
	Update(therapist *domain.Therapist) error
	Delete(id string) error
	List() ([]*domain.Therapist, error)
}
