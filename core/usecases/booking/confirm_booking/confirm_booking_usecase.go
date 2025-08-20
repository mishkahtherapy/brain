package confirm_booking

import (
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/mishkahtherapy/brain/core/domain"
	"github.com/mishkahtherapy/brain/core/domain/booking"
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
}

func NewUsecase(
	bookingRepo ports.BookingRepository,
	sessionRepo ports.SessionRepository,
	therapistRepo ports.TherapistRepository,
	notificationPort ports.NotificationPort,
	notificationRepo ports.NotificationRepository,
	therapistAppBaseURL string,
) *Usecase {
	return &Usecase{
		bookingRepo:         bookingRepo,
		sessionRepo:         sessionRepo,
		therapistRepo:       therapistRepo,
		notificationPort:    notificationPort,
		notificationRepo:    notificationRepo,
		therapistAppBaseURL: therapistAppBaseURL,
	}
}

func (u *Usecase) Execute(input Input) (*booking.Booking, error) {
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

	// Change state to Confirmed
	existingBooking.State = booking.BookingStateConfirmed
	existingBooking.UpdatedAt = domain.NewUTCTimestamp()

	err = u.bookingRepo.Update(existingBooking)
	if err != nil {
		return nil, common.ErrFailedToConfirmBooking
	}

	// Create a new session for the confirmed booking
	now := domain.NewUTCTimestamp()
	session := &domain.Session{
		ID:          domain.NewSessionID(),
		BookingID:   existingBooking.ID,
		TherapistID: existingBooking.TherapistID,
		ClientID:    existingBooking.ClientID,
		TimeSlotID:  existingBooking.TimeSlotID,
		StartTime:   existingBooking.StartTime,
		Duration:    existingBooking.Duration,
		PaidAmount:  input.PaidAmountUSD,
		Language:    input.Language,
		State:       domain.SessionStatePlanned,
		Notes:       "",
		MeetingURL:  "",
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	// Persist the session
	err = u.sessionRepo.CreateSession(session)
	if err != nil {
		return nil, ErrFailedToCreateSession
	}

	// Notify therapist
	therapist, err := u.therapistRepo.GetByID(existingBooking.TherapistID)
	if err != nil {
		return nil, err
	}

	if therapist.DeviceID == "" {
		slog.Info("therapist has no device id, skipping notification", "therapist_id", therapist.ID)
		return existingBooking, nil
	}

	therapistTimezoneOffset := int(therapist.TimezoneOffset / 60)
	timezoneLabel := fmt.Sprintf("UTC%+d", therapistTimezoneOffset)
	therapistTimezone := time.FixedZone(timezoneLabel, therapistTimezoneOffset)
	therapistTime := time.Date(existingBooking.StartTime.Year(), existingBooking.StartTime.Month(), existingBooking.StartTime.Day(), 0, 0, 0, 0, therapistTimezone)

	notification := ports.Notification{
		Title:    "Session Confirmed",
		Body:     fmt.Sprintf("Your next session is confirmed on %s", therapistTime.Format(time.DateOnly)),
		ImageURL: "https://therapist.mishkahtherapy.com/mishkah-logo.png",
		// TODO: add session id to the link
		Link: fmt.Sprintf("%s/sessions", u.therapistAppBaseURL),
	}

	firebaseNotificationId, err := u.notificationPort.SendNotification(therapist.DeviceID, notification)
	if err != nil {
		return nil, err
	}

	// Persist the notification
	err = u.notificationRepo.CreateNotification(therapist.ID, *firebaseNotificationId, notification)
	if err != nil {
		return nil, err
	}

	return existingBooking, nil
}
