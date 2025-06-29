package ports

import "github.com/mishkahtherapy/brain/core/domain"

type TherapistRepository interface {
	GetByID(id domain.TherapistID) (*domain.Therapist, error)
	GetByEmail(email domain.Email) (*domain.Therapist, error)
	Create(therapist *domain.Therapist) error
	UpdateSpecializations(therapistID domain.TherapistID, specializationIDs []domain.SpecializationID) error
	Delete(id domain.TherapistID) error
	List() ([]*domain.Therapist, error)
}
