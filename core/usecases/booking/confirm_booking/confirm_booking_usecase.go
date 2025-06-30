package confirm_booking

import (
	"errors"

	"github.com/mishkahtherapy/brain/core/domain"
	"github.com/mishkahtherapy/brain/core/ports"
)

var ErrFailedToConfirmBooking = errors.New("failed to confirm booking")
var ErrBookingIDIsRequired = errors.New("booking ID is required")
var ErrBookingNotFound = errors.New("booking not found")
var ErrInvalidBookingState = errors.New("booking must be in pending state to be confirmed")
var ErrPaidAmountIsRequired = errors.New("paid amount is required")
var ErrLanguageIsRequired = errors.New("language is required")
var ErrFailedToCreateSession = errors.New("failed to create session for confirmed booking")

type Input struct {
	BookingID  domain.BookingID       `json:"bookingId"`
	PaidAmount int                    `json:"paidAmount"` // USD cents
	Language   domain.SessionLanguage `json:"language"`
}

type Usecase struct {
	bookingRepo ports.BookingRepository
	sessionRepo ports.SessionRepository
}

func NewUsecase(bookingRepo ports.BookingRepository, sessionRepo ports.SessionRepository) *Usecase {
	return &Usecase{
		bookingRepo: bookingRepo,
		sessionRepo: sessionRepo,
	}
}

func (u *Usecase) Execute(input Input) (*domain.Booking, error) {
	// Validate required fields
	if input.BookingID == "" {
		return nil, ErrBookingIDIsRequired
	}
	if input.PaidAmount <= 0 {
		return nil, ErrPaidAmountIsRequired
	}
	if input.Language == "" {
		return nil, ErrLanguageIsRequired
	}

	// Get existing booking
	booking, err := u.bookingRepo.GetByID(input.BookingID)
	if err != nil || booking == nil {
		return nil, ErrBookingNotFound
	}

	// Validate booking is in Pending state
	if booking.State != domain.BookingStatePending {
		return nil, ErrInvalidBookingState
	}

	// Change state to Confirmed
	booking.State = domain.BookingStateConfirmed
	booking.UpdatedAt = domain.NewUTCTimestamp()

	err = u.bookingRepo.Update(booking)
	if err != nil {
		return nil, ErrFailedToConfirmBooking
	}

	// Create a new session for the confirmed booking
	now := domain.NewUTCTimestamp()
	session := &domain.Session{
		ID:          domain.NewSessionID(),
		BookingID:   booking.ID,
		TherapistID: booking.TherapistID,
		ClientID:    booking.ClientID,
		TimeSlotID:  booking.TimeSlotID,
		StartTime:   booking.StartTime,
		PaidAmount:  input.PaidAmount,
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

	return booking, nil
}
