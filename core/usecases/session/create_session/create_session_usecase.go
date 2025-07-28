package create_session

import (
	"time"

	"github.com/mishkahtherapy/brain/core/domain"
	"github.com/mishkahtherapy/brain/core/ports"
	"github.com/mishkahtherapy/brain/core/usecases/common"
)

// Input struct defines all required parameters for creating a session
type Input struct {
	BookingID   domain.BookingID       `json:"bookingId"`
	TherapistID domain.TherapistID     `json:"therapistId"`
	ClientID    domain.ClientID        `json:"clientId"`
	TimeSlotID  domain.TimeSlotID      `json:"timeSlotId"`
	StartTime   domain.UTCTimestamp    `json:"startTime"`
	Duration    domain.DurationMinutes `json:"duration"`
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
		return nil, common.ErrBookingNotFound
	}

	// Verify the therapist exists
	therapist, err := u.therapistRepo.GetByID(input.TherapistID)
	if err != nil || therapist == nil {
		return nil, common.ErrTherapistNotFound
	}

	// Verify the client exists
	client, err := u.clientRepo.BulkGetByID([]domain.ClientID{input.ClientID})
	if err != nil || client == nil {
		return nil, common.ErrClientNotFound
	}

	// Verify the timeslot exists
	timeSlot, err := u.timeSlotRepo.GetByID(input.TimeSlotID)
	if err != nil || timeSlot == nil {
		return nil, common.ErrTimeSlotNotFound
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
		Duration:    input.Duration,
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
		return nil, common.ErrFailedToCreateSession
	}

	return session, nil
}

// validateInput ensures all required fields are provided
func validateInput(input Input) error {
	if input.BookingID == "" {
		return common.ErrBookingIDIsRequired
	}
	if input.TherapistID == "" {
		return common.ErrTherapistIDIsRequired
	}
	if input.ClientID == "" {
		return common.ErrClientIDIsRequired
	}
	if input.TimeSlotID == "" {
		return common.ErrTimeSlotIDIsRequired
	}
	if time.Time(input.StartTime).IsZero() {
		return common.ErrStartTimeIsRequired
	}
	if input.PaidAmount <= 0 {
		return common.ErrPaidAmountIsRequired
	}
	if input.Language == "" {
		return common.ErrLanguageIsRequired
	}
	return nil
}
