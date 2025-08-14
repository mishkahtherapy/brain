package confirm_booking

import (
	"errors"

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
	bookingRepo ports.BookingRepository
	sessionRepo ports.SessionRepository
}

func NewUsecase(bookingRepo ports.BookingRepository, sessionRepo ports.SessionRepository) *Usecase {
	return &Usecase{
		bookingRepo: bookingRepo,
		sessionRepo: sessionRepo,
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

	return existingBooking, nil
}
