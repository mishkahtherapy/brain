package ports

import (
	"github.com/mishkahtherapy/brain/core/domain/timeslot"
)

type TimeSlotRepository interface {
	GetByID(id string) (*timeslot.TimeSlot, error)
	Create(timeslot *timeslot.TimeSlot) error
	Update(timeslot *timeslot.TimeSlot) error
	Delete(id string) error
	ListByTherapist(therapistID string) ([]*timeslot.TimeSlot, error)
	ListByDay(therapistID string, day string) ([]*timeslot.TimeSlot, error)
	BulkToggleByTherapistID(therapistID string, isActive bool) error
}
