package ports

import (
	"time"

	"github.com/mishkahtherapy/brain/core/domain"
	"github.com/mishkahtherapy/brain/core/domain/booking"
)

type AdhocBookingRepository interface {
	Create(adhocBooking *booking.AdhocBooking) error
	BulkListByTherapistForDateRange(
		therapistIDs []domain.TherapistID,
		states []booking.BookingState,
		startDate, endDate time.Time,
	) (map[domain.TherapistID][]*booking.AdhocBooking, error)
}
