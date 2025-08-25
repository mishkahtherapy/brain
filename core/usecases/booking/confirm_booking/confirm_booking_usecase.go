package confirm_booking

import (
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/mishkahtherapy/brain/core/domain"
	"github.com/mishkahtherapy/brain/core/domain/booking"
	"github.com/mishkahtherapy/brain/core/domain/therapist"
	"github.com/mishkahtherapy/brain/core/ports"
	"github.com/mishkahtherapy/brain/core/usecases/common"
)

// Domain-specific error (not in common since it's specific to booking confirmation)
var ErrFailedToCreateSession = errors.New("failed to create session for confirmed booking")

type Input struct {
	BookingID     domain.BookingID
	PaidAmountUSD int // USD cents
	Language      domain.SessionLanguage
}

type Usecase struct {
	bookingRepo         ports.BookingRepository
	sessionRepo         ports.SessionRepository
	therapistRepo       ports.TherapistRepository
	notificationPort    ports.NotificationPort
	notificationRepo    ports.NotificationRepository
	therapistAppBaseURL string
	transactionPort     ports.TransactionPort
}

func NewUsecase(
	bookingRepo ports.BookingRepository,
	sessionRepo ports.SessionRepository,
	therapistRepo ports.TherapistRepository,
	notificationPort ports.NotificationPort,
	notificationRepo ports.NotificationRepository,
	therapistAppBaseURL string,
	transactionPort ports.TransactionPort,
) *Usecase {
	return &Usecase{
		bookingRepo:         bookingRepo,
		sessionRepo:         sessionRepo,
		therapistRepo:       therapistRepo,
		notificationPort:    notificationPort,
		notificationRepo:    notificationRepo,
		therapistAppBaseURL: therapistAppBaseURL,
		transactionPort:     transactionPort,
	}
}

func (u *Usecase) Execute(input Input) (*booking.Booking, error) {
	therapistBookings, err := u.validateInput(input)
	if err != nil {
		return nil, err
	}

	existingBooking, err := u.bookingRepo.GetByID(input.BookingID)
	if err != nil {
		return nil, err
	}

	if existingBooking == nil {
		return nil, common.ErrBookingNotFound
	}

	// ------------------
	// Confirm booking (run in a transaction)
	// ------------------
	tx, err := u.transactionPort.Begin()
	if err != nil {
		return nil, err
	}

	session, err := u.confirmBooking(tx, existingBooking, input.PaidAmountUSD, input.Language)
	if err != nil {
		return nil, err
	}

	err = u.cancelConflictingBookings(tx, existingBooking.ID, therapistBookings)
	if err != nil {
		return nil, err
	}

	err = u.transactionPort.Commit(tx)
	if err != nil {
		return nil, err
	}
	// ------------------

	// Notify therapist
	therapist, err := u.therapistRepo.GetByID(existingBooking.TherapistID)
	if err != nil {
		return nil, err
	}

	if therapist.DeviceID == "" {
		slog.Info("therapist has no device id, skipping notification", "therapist_id", therapist.ID)
		return existingBooking, nil
	}

	err = u.notifyTherapist(session, *therapist)
	if err != nil {
		slog.Warn("failed to notify therapist", "therapist_id", therapist.ID, "error", err)
	}

	return existingBooking, nil
}

func (u *Usecase) notifyTherapist(session *domain.Session, therapist therapist.Therapist) error {
	therapistTimezoneOffset := int(therapist.TimezoneOffset / 60)
	timezoneLabel := fmt.Sprintf("UTC%+d", therapistTimezoneOffset)
	therapistTimezone := time.FixedZone(timezoneLabel, therapistTimezoneOffset)
	therapistTime := time.Date(session.StartTime.Year(), session.StartTime.Month(), session.StartTime.Day(), 0, 0, 0, 0, therapistTimezone)

	notification := ports.Notification{
		Title:    "Session Confirmed",
		Body:     fmt.Sprintf("Your next session is confirmed on %s", therapistTime.Format(time.DateOnly)),
		ImageURL: "https://therapist.mishkahtherapy.com/mishkah-logo.png",
		// TODO: add session id to the link
		Link: fmt.Sprintf("%s/sessions", u.therapistAppBaseURL),
	}

	firebaseNotificationId, err := u.notificationPort.SendNotification(therapist.DeviceID, notification)
	if err != nil {
		return err
	}

	// Persist the notification
	err = u.notificationRepo.CreateNotification(therapist.ID, *firebaseNotificationId, notification)
	if err != nil {
		return err
	}

	return nil
}

func (u *Usecase) validateInput(input Input) ([]*booking.Booking, error) {
	// Validate required fields
	if input.BookingID == "" {
		return nil, common.ErrBookingIDIsRequired
	}
	if input.PaidAmountUSD <= 0 {
		return nil, common.ErrPaidAmountIsRequired
	}
	if input.Language == "" {
		return nil, common.ErrLanguageIsRequired
	}

	// Get existing booking
	existingBooking, err := u.bookingRepo.GetByID(input.BookingID)
	if err != nil || existingBooking == nil {
		return nil, common.ErrBookingNotFound
	}

	// Validate booking is in Pending state
	if existingBooking.State != booking.BookingStatePending {
		return nil, common.ErrInvalidBookingState
	}

	// Get other bookings at the same time
	startTime := time.Time(existingBooking.StartTime)
	endTime := startTime.Add(time.Duration(existingBooking.Duration) * time.Minute)
	otherBookingMap, err := u.bookingRepo.BulkListByTherapistForDateRange(
		[]domain.TherapistID{existingBooking.TherapistID},
		[]booking.BookingState{booking.BookingStatePending, booking.BookingStateConfirmed},
		startTime,
		endTime,
	)

	if err != nil {
		return nil, err
	}

	therapistBookings := otherBookingMap[existingBooking.TherapistID]
	for _, b := range therapistBookings {
		if b.State == booking.BookingStateConfirmed {
			slog.Error("other booking found at the same time", "booking", b, "startTime", startTime, "endTime", endTime)
			// TODO: cleanup invalid states and force cancel other bookings
			tx, err := u.transactionPort.Begin()
			if err != nil {
				return nil, err
			}
			err = u.cancelConflictingBookings(tx, existingBooking.ID, therapistBookings)
			if err != nil {
				return nil, err
			}
			err = u.transactionPort.Commit(tx)
			if err != nil {
				return nil, err
			}
			return nil, common.ErrTimeSlotAlreadyBooked
		}
	}

	return therapistBookings, nil
}

func (u *Usecase) confirmBooking(tx ports.SQLTx, existingBooking *booking.Booking, paidAmountUSD int, language domain.SessionLanguage) (*domain.Session, error) {
	// Change state to Confirmed
	existingBooking.State = booking.BookingStateConfirmed
	existingBooking.UpdatedAt = domain.NewUTCTimestamp()

	err := u.bookingRepo.UpdateTx(tx, existingBooking)
	if err != nil {
		return nil, common.ErrFailedToConfirmBooking
	}

	// Create a new session for the confirmed booking
	now := domain.NewUTCTimestamp()
	session := &domain.Session{
		ID:                   domain.NewSessionID(),
		BookingID:            existingBooking.ID,
		TherapistID:          existingBooking.TherapistID,
		ClientID:             existingBooking.ClientID,
		TimeSlotID:           existingBooking.TimeSlotID,
		StartTime:            existingBooking.StartTime,
		Duration:             existingBooking.Duration,
		PaidAmount:           paidAmountUSD,
		Language:             language,
		State:                domain.SessionStatePlanned,
		Notes:                "",
		MeetingURL:           "",
		ClientTimezoneOffset: existingBooking.ClientTimezoneOffset,
		CreatedAt:            now,
		UpdatedAt:            now,
	}

	// Persist the session
	err = u.sessionRepo.CreateSession(tx, session)
	if err != nil {
		return nil, ErrFailedToCreateSession
	}

	return session, nil
}

func (u *Usecase) cancelConflictingBookings(tx ports.SQLTx, bookingID domain.BookingID, therapistBookings []*booking.Booking) error {
	toBeCancelled := make([]domain.BookingID, 0)
	for _, b := range therapistBookings {
		if b.ID == bookingID {
			continue
		}
		toBeCancelled = append(toBeCancelled, b.ID)
	}

	if len(toBeCancelled) == 0 {
		return nil
	}

	// Cancel the bookings
	err := u.bookingRepo.BulkCancel(tx, toBeCancelled)
	if err != nil {
		return err
	}

	return nil
}
