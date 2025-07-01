package confirm_booking

import (
	"errors"

	"github.com/mishkahtherapy/brain/core/domain"
	"github.com/mishkahtherapy/brain/core/ports"
	"github.com/mishkahtherapy/brain/core/usecases/common"
)

// Domain-specific error (not in common since it's specific to booking confirmation)
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
		return nil, common.ErrBookingIDIsRequired
	}
	if input.PaidAmount <= 0 {
		return nil, common.ErrPaidAmountIsRequired
	}
	if input.Language == "" {
		return nil, common.ErrLanguageIsRequired
	}

	// Get existing booking
	booking, err := u.bookingRepo.GetByID(input.BookingID)
	if err != nil || booking == nil {
		return nil, common.ErrBookingNotFound
	}

	// Validate booking is in Pending state
	if booking.State != domain.BookingStatePending {
		return nil, common.ErrInvalidBookingState
	}

	// Change state to Confirmed
	booking.State = domain.BookingStateConfirmed
	booking.UpdatedAt = domain.NewUTCTimestamp()

	err = u.bookingRepo.Update(booking)
	if err != nil {
		return nil, common.ErrFailedToConfirmBooking
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
