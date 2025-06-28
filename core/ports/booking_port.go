package ports

import "github.com/mishkahtherapy/brain/core/domain"

type BookingRepository interface {
	GetByID(id string) (*domain.Booking, error)
	Create(booking *domain.Booking) error
	Update(booking *domain.Booking) error
	Delete(id string) error
	ListByTherapist(therapistID string) ([]*domain.Booking, error)
}
