package ports

import "github.com/mishkahtherapy/brain/core/domain"

type TimeSlotRepository interface {
	GetByID(id string) (*domain.TimeSlot, error)
	Create(timeslot *domain.TimeSlot) error
	Update(timeslot *domain.TimeSlot) error
	Delete(id string) error
	ListByTherapist(therapistID string) ([]*domain.TimeSlot, error)
	ListByDay(therapistID string, day string) ([]*domain.TimeSlot, error)
}
