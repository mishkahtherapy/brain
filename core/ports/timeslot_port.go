package ports

import (
	"github.com/mishkahtherapy/brain/core/domain"
	"github.com/mishkahtherapy/brain/core/domain/timeslot"
)

type TimeSlotRepository interface {
	GetByID(id domain.TimeSlotID) (*timeslot.TimeSlot, error)
	Create(timeslot *timeslot.TimeSlot) error
	Update(timeslot *timeslot.TimeSlot) error
	Delete(id domain.TimeSlotID) error
	ListByTherapist(therapistID domain.TherapistID) ([]*timeslot.TimeSlot, error)
	BulkListByTherapist(therapistIDs []domain.TherapistID) (map[domain.TherapistID][]*timeslot.TimeSlot, error)
	BulkToggleByTherapistID(therapistID domain.TherapistID, isActive bool) error
}
