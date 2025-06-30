package create_session

import (
	"errors"
	"time"

	"github.com/mishkahtherapy/brain/core/domain"
	"github.com/mishkahtherapy/brain/core/ports"
)

// Error definitions
var ErrFailedToCreateSession = errors.New("failed to create session")
var ErrBookingIDIsRequired = errors.New("booking ID is required")
var ErrTherapistIDIsRequired = errors.New("therapist ID is required")
var ErrClientIDIsRequired = errors.New("client ID is required")
var ErrTimeSlotIDIsRequired = errors.New("timeslot ID is required")
var ErrStartTimeIsRequired = errors.New("start time is required")
var ErrPaidAmountIsRequired = errors.New("paid amount is required")
var ErrLanguageIsRequired = errors.New("language is required")
var ErrBookingNotFound = errors.New("booking not found")
var ErrTherapistNotFound = errors.New("therapist not found")
var ErrClientNotFound = errors.New("client not found")
var ErrTimeSlotNotFound = errors.New("timeslot not found")

// Input struct defines all required parameters for creating a session
type Input struct {
	BookingID   domain.BookingID       `json:"bookingId"`
	TherapistID domain.TherapistID     `json:"therapistId"`
	ClientID    domain.ClientID        `json:"clientId"`
	TimeSlotID  domain.TimeSlotID      `json:"timeSlotId"`
	StartTime   domain.UTCTimestamp    `json:"startTime"`
	PaidAmount  int                    `json:"paidAmount"` // USD cents
	Language    domain.SessionLanguage `json:"language"`
}

// Usecase struct with required dependencies
type Usecase struct {
	sessionRepo   ports.SessionRepository
	bookingRepo   ports.BookingRepository
	therapistRepo ports.TherapistRepository
	clientRepo    ports.ClientRepository
	timeSlotRepo  ports.TimeSlotRepository
}

// NewUsecase creates a new instance of the create session usecase
func NewUsecase(
	sessionRepo ports.SessionRepository,
	bookingRepo ports.BookingRepository,
	therapistRepo ports.TherapistRepository,
	clientRepo ports.ClientRepository,
	timeSlotRepo ports.TimeSlotRepository,
) *Usecase {
	return &Usecase{
		sessionRepo:   sessionRepo,
		bookingRepo:   bookingRepo,
		therapistRepo: therapistRepo,
		clientRepo:    clientRepo,
		timeSlotRepo:  timeSlotRepo,
	}
}

// Execute creates a new session based on the provided input
func (u *Usecase) Execute(input Input) (*domain.Session, error) {
	// Validate required fields
	if err := validateInput(input); err != nil {
		return nil, err
	}

	// Verify the booking exists
	booking, err := u.bookingRepo.GetByID(input.BookingID)
	if err != nil || booking == nil {
		return nil, ErrBookingNotFound
	}

	// Verify the therapist exists
	therapist, err := u.therapistRepo.GetByID(input.TherapistID)
	if err != nil || therapist == nil {
		return nil, ErrTherapistNotFound
	}

	// Verify the client exists
	client, err := u.clientRepo.GetByID(input.ClientID)
	if err != nil || client == nil {
		return nil, ErrClientNotFound
	}

	// Verify the timeslot exists
	timeSlot, err := u.timeSlotRepo.GetByID(string(input.TimeSlotID))
	if err != nil || timeSlot == nil {
		return nil, ErrTimeSlotNotFound
	}

	// Create a new session with PLANNED state
	now := domain.NewUTCTimestamp()
	session := &domain.Session{
		ID:          domain.NewSessionID(),
		BookingID:   input.BookingID,
		TherapistID: input.TherapistID,
		ClientID:    input.ClientID,
		TimeSlotID:  input.TimeSlotID,
		StartTime:   input.StartTime,
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

	return session, nil
}

// validateInput ensures all required fields are provided
func validateInput(input Input) error {
	if input.BookingID == "" {
		return ErrBookingIDIsRequired
	}
	if input.TherapistID == "" {
		return ErrTherapistIDIsRequired
	}
	if input.ClientID == "" {
		return ErrClientIDIsRequired
	}
	if input.TimeSlotID == "" {
		return ErrTimeSlotIDIsRequired
	}
	if time.Time(input.StartTime).IsZero() {
		return ErrStartTimeIsRequired
	}
	if input.PaidAmount <= 0 {
		return ErrPaidAmountIsRequired
	}
	if input.Language == "" {
		return ErrLanguageIsRequired
	}
	return nil
}
