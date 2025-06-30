package ports

import "github.com/mishkahtherapy/brain/core/domain"

type BookingRepository interface {
	GetByID(id domain.BookingID) (*domain.Booking, error)
	Create(booking *domain.Booking) error
	Update(booking *domain.Booking) error
	Delete(id domain.BookingID) error
	ListByTherapist(therapistID domain.TherapistID) ([]*domain.Booking, error)
	ListByClient(clientID domain.ClientID) ([]*domain.Booking, error)
	ListByState(state domain.BookingState) ([]*domain.Booking, error)
	ListByTherapistAndState(therapistID domain.TherapistID, state domain.BookingState) ([]*domain.Booking, error)
	ListByClientAndState(clientID domain.ClientID, state domain.BookingState) ([]*domain.Booking, error)
}
