package ports

import (
	"time"

	"github.com/mishkahtherapy/brain/core/domain"
	"github.com/mishkahtherapy/brain/core/domain/booking"
)

type AdhocBookingRepository interface {
	GetByID(id domain.AdhocBookingID) (*booking.AdhocBooking, error)
	Create(adhocBooking *booking.AdhocBooking) error
	UpdateState(adhocBookingID domain.AdhocBookingID, state booking.BookingState, updatedAt time.Time) error
	UpdateStateTx(sqlExec SQLExec, adhocBookingID domain.AdhocBookingID, state booking.BookingState, updatedAt time.Time) error
	ListByTherapistForDateRange(
		therapistID domain.TherapistID,
		states []booking.BookingState,
		startDate, endDate time.Time,
	) ([]*booking.AdhocBooking, error)
	BulkListByTherapistForDateRange(
		therapistIDs []domain.TherapistID,
		states []booking.BookingState,
		startDate, endDate time.Time,
	) (map[domain.TherapistID][]*booking.AdhocBooking, error)
	BulkCancel(tx SQLTx, adhocBookingIDs []domain.AdhocBookingID) error
	Search(startDate, endDate time.Time, states []booking.BookingState) ([]*booking.AdhocBooking, error)
	List(filters BookingFilters) ([]*booking.AdhocBooking, error)
}
