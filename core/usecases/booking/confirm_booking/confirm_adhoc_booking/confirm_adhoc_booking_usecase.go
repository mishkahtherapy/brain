package confirm_adhoc_booking

import (
	"log/slog"

	"github.com/mishkahtherapy/brain/core/domain"
	"github.com/mishkahtherapy/brain/core/domain/booking"
	"github.com/mishkahtherapy/brain/core/ports"
	"github.com/mishkahtherapy/brain/core/usecases/booking/confirm_booking"
	"github.com/mishkahtherapy/brain/core/usecases/common"
	"github.com/mishkahtherapy/brain/core/usecases/notify_therapist_booking"
)

type Input struct {
	BookingID     domain.AdhocBookingID
	PaidAmountUSD int // USD cents
	Language      domain.SessionLanguage
}

type Usecase struct {
	adhocBookingRepo    ports.AdhocBookingRepository
	sessionRepo         ports.SessionRepository
	therapistRepo       ports.TherapistRepository
	notificationPort    ports.NotificationPort
	notificationRepo    ports.NotificationRepository
	therapistAppBaseURL string
	transactionPort     ports.TransactionPort

	cancelPendingBookings *confirm_booking.PendingBookingConflictResolver
	notifyTherapist       *notify_therapist_booking.Usecase
}

func NewUsecase(
	bookingRepo ports.BookingRepository,
	adhocBookingRepo ports.AdhocBookingRepository,
	sessionRepo ports.SessionRepository,
	therapistRepo ports.TherapistRepository,
	notificationPort ports.NotificationPort,
	notificationRepo ports.NotificationRepository,
	therapistAppBaseURL string,
	transactionPort ports.TransactionPort,
	notifyTherapist *notify_therapist_booking.Usecase,
) *Usecase {
	return &Usecase{
		adhocBookingRepo:    adhocBookingRepo,
		sessionRepo:         sessionRepo,
		therapistRepo:       therapistRepo,
		notificationPort:    notificationPort,
		notificationRepo:    notificationRepo,
		therapistAppBaseURL: therapistAppBaseURL,
		transactionPort:     transactionPort,
		cancelPendingBookings: confirm_booking.NewPendingBookingConflictResolver(
			bookingRepo,
			adhocBookingRepo,
		),
		notifyTherapist: notifyTherapist,
	}
}

func (u *Usecase) Execute(input Input) (*ports.BookingResponse, error) {
	err := u.validateInput(input)
	if err != nil {
		return nil, err
	}

	// Get pending booking
	toBeConfirmedBooking, err := u.getAdhocBooking(input.BookingID)
	if err != nil {
		return nil, err
	}

	// ------------------
	// Confirm booking (run in a transaction)
	// ------------------
	tx, err := u.transactionPort.Begin()
	if err != nil {
		return nil, err
	}

	err = u.cancelPendingBookings.CancelConflicts(tx,
		toBeConfirmedBooking.TherapistID,
		toBeConfirmedBooking.StartTime,
		toBeConfirmedBooking.Duration,
		toBeConfirmedBooking.ID,
		"", // No regular booking id for adhoc bookings
	)
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	session, err := u.confirmBooking(tx, toBeConfirmedBooking, input.PaidAmountUSD, input.Language)
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	err = u.transactionPort.Commit(tx)
	if err != nil {
		return nil, err
	}
	// ------------------

	u.notifyTherapist.Execute(session)
	return &ports.BookingResponse{
		AdhocBookingID:       toBeConfirmedBooking.ID,
		TherapistID:          toBeConfirmedBooking.TherapistID,
		ClientID:             toBeConfirmedBooking.ClientID,
		State:                toBeConfirmedBooking.State,
		StartTime:            toBeConfirmedBooking.StartTime,
		Duration:             toBeConfirmedBooking.Duration,
		ClientTimezoneOffset: toBeConfirmedBooking.ClientTimezoneOffset,
	}, nil
}

func (u *Usecase) validateInput(input Input) error {
	// Validate required fields
	if input.BookingID == "" {
		return common.ErrBookingIDIsRequired
	}
	if input.PaidAmountUSD <= 0 {
		return common.ErrPaidAmountIsRequired
	}
	if input.Language == "" {
		return common.ErrLanguageIsRequired
	}

	return nil
}

func (u *Usecase) getAdhocBooking(bookingID domain.AdhocBookingID) (*booking.AdhocBooking, error) {
	toBeConfirmedBooking, err := u.adhocBookingRepo.GetByID(bookingID)
	if err != nil || toBeConfirmedBooking == nil {
		return nil, common.ErrBookingNotFound
	}
	// Validate booking is in Pending state
	if toBeConfirmedBooking.State != booking.BookingStatePending {
		slog.Error("to be confirmed adhoc booking is not in Pending state",
			slog.Group(
				"booking",
				"id", toBeConfirmedBooking.ID,
				"state", toBeConfirmedBooking.State,
			),
		)
		return nil, common.ErrInvalidBookingState
	}

	return toBeConfirmedBooking, nil
}

func (u *Usecase) confirmBooking(
	tx ports.SQLTx,
	existingBooking *booking.AdhocBooking,
	paidAmountUSD int,
	language domain.SessionLanguage,
) (*domain.Session, error) {
	// Change state to Confirmed
	err := u.adhocBookingRepo.UpdateStateTx(
		tx,
		existingBooking.ID,
		booking.BookingStateConfirmed,
		domain.NewUTCTimestamp().Time(),
	)
	if err != nil {
		return nil, common.ErrFailedToConfirmBooking
	}

	// Create a new session for the confirmed booking
	now := domain.NewUTCTimestamp()
	session := &domain.Session{
		ID:                   domain.NewSessionID(),
		AdhocBookingID:       existingBooking.ID,
		TherapistID:          existingBooking.TherapistID,
		ClientID:             existingBooking.ClientID,
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
		return nil, common.ErrFailedToCreateSession
	}

	return session, nil
}
