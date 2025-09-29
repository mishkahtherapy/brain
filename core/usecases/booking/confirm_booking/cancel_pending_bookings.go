package confirm_booking

import (
	"time"

	"github.com/mishkahtherapy/brain/core/domain"
	"github.com/mishkahtherapy/brain/core/domain/booking"
	"github.com/mishkahtherapy/brain/core/ports"
)

type PendingBookingConflictResolver struct {
	bookingRepo      ports.BookingRepository
	adhocBookingRepo ports.AdhocBookingRepository
}

func NewPendingBookingConflictResolver(
	bookingRepo ports.BookingRepository,
	adhocBookingRepo ports.AdhocBookingRepository,
) *PendingBookingConflictResolver {
	return &PendingBookingConflictResolver{
		bookingRepo:      bookingRepo,
		adhocBookingRepo: adhocBookingRepo,
	}
}

func (c *PendingBookingConflictResolver) CancelConflicts(tx ports.SQLTx,
	therapistID domain.TherapistID,
	bookingStartTime domain.UTCTimestamp,
	bookingDuration domain.DurationMinutes,
	adhocBookingID domain.AdhocBookingID,
	regularBookingID domain.BookingID,
) error {
	// Get other bookings at the same time
	startTime := time.Time(bookingStartTime)
	endTime := startTime.Add(time.Duration(bookingDuration) * time.Minute)

	therapistBookings, err := c.cancelRegularBookings(tx, regularBookingID, therapistID, startTime, endTime)
	if err != nil {
		return err
	}

	adhocBookings, err := c.cancelAdhocBookings(tx, adhocBookingID, therapistID, startTime, endTime)
	if err != nil {
		return err
	}

	// If there are no adhoc bookings, return nil. Don't need to notify operators in this case.
	if len(adhocBookings) == 0 && len(therapistBookings) == 0 {
		return nil
	}

	// TODO: notify operators with cancellations.

	return nil
}

func (c *PendingBookingConflictResolver) cancelRegularBookings(tx ports.SQLTx,
	toBeConfirmedBookingID domain.BookingID,
	therapistID domain.TherapistID,
	startTime time.Time,
	endTime time.Time,
) ([]*booking.Booking, error) {

	therapistBookings, err := c.bookingRepo.ListByTherapistForDateRange(
		therapistID,
		[]booking.BookingState{booking.BookingStatePending, booking.BookingStateConfirmed},
		startTime,
		endTime,
	)
	if err != nil {
		return nil, err
	}

	// If any booking is already confirmed, return an error
	for _, b := range therapistBookings {
		if b.State == booking.BookingStateConfirmed {
			return nil, booking.ErrBookingAlreadyConfirmed
		}
	}

	toBeCancelled := make([]domain.BookingID, 0)
	for _, b := range therapistBookings {
		if b.ID == toBeConfirmedBookingID {
			continue
		}
		toBeCancelled = append(toBeCancelled, b.ID)
	}

	if len(toBeCancelled) == 0 {
		return []*booking.Booking{}, nil
	}

	// Cancel the bookings
	err = c.bookingRepo.BulkCancel(tx, toBeCancelled)
	if err != nil {
		return nil, err
	}
	return therapistBookings, nil
}

func (c *PendingBookingConflictResolver) cancelAdhocBookings(
	tx ports.SQLTx,
	toBeConfirmedBookingID domain.AdhocBookingID,
	therapistID domain.TherapistID,
	startTime time.Time,
	endTime time.Time,
) ([]*booking.AdhocBooking, error) {

	adhocBookings, err := c.adhocBookingRepo.ListByTherapistForDateRange(
		therapistID,
		[]booking.BookingState{booking.BookingStatePending, booking.BookingStateConfirmed},
		startTime,
		endTime,
	)
	if err != nil {
		return nil, err
	}

	// If any booking is already confirmed, return an error
	for _, b := range adhocBookings {
		if b.State == booking.BookingStateConfirmed {
			return nil, booking.ErrBookingAlreadyConfirmed
		}
	}

	toBeCancelled := make([]domain.AdhocBookingID, 0)
	// Cancel all conflicting adhoc bookings
	for _, b := range adhocBookings {
		if b.ID == toBeConfirmedBookingID {
			continue
		}
		toBeCancelled = append(toBeCancelled, b.ID)
	}

	if len(toBeCancelled) == 0 {
		return []*booking.AdhocBooking{}, nil
	}

	// Cancel the bookings
	err = c.adhocBookingRepo.BulkCancel(tx, toBeCancelled)
	if err != nil {
		return nil, err
	}

	return adhocBookings, nil
}
