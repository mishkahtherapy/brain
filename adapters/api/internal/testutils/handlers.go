package testutils

import (
	"github.com/mishkahtherapy/brain/adapters/db/booking_db"
	"github.com/mishkahtherapy/brain/adapters/db/therapist_db"
	"github.com/mishkahtherapy/brain/adapters/db/timeslot_db"
	"github.com/mishkahtherapy/brain/core/ports"
)

// RepositorySet contains commonly used repositories
type RepositorySet struct {
	TherapistRepo ports.TherapistRepository
	TimeSlotRepo  ports.TimeSlotRepository
	BookingRepo   ports.BookingRepository
	ClientRepo    ports.ClientRepository
	SessionRepo   ports.SessionRepository
}

// SetupRepositories creates standard repositories plus test repositories for missing ones
func SetupRepositories(database ports.SQLDatabase) *RepositorySet {
	return &RepositorySet{
		TherapistRepo: therapist_db.NewTherapistRepository(database),
		TimeSlotRepo:  timeslot_db.NewTimeSlotRepository(database),
		BookingRepo:   booking_db.NewBookingRepository(database),
		ClientRepo:    NewTestClientRepository(database),
		SessionRepo:   NewTestSessionRepository(database),
	}
}
