package create_booking

import (
	"testing"
	"time"

	"github.com/mishkahtherapy/brain/core/domain"
	"github.com/mishkahtherapy/brain/core/domain/booking"
	"github.com/mishkahtherapy/brain/core/domain/client"
	"github.com/mishkahtherapy/brain/core/domain/therapist"
	"github.com/mishkahtherapy/brain/core/domain/timeslot"
	"github.com/mishkahtherapy/brain/core/usecases/common"
)

// -----------------------------
// In-memory fakes
// -----------------------------

type inMemoryBookingRepo struct{ bookings []*booking.Booking }

func (r *inMemoryBookingRepo) GetByID(id domain.BookingID) (*booking.Booking, error) { return nil, nil }
func (r *inMemoryBookingRepo) Create(b *booking.Booking) error {
	r.bookings = append(r.bookings, b)
	return nil
}
func (r *inMemoryBookingRepo) Update(*booking.Booking) error    { return nil }
func (r *inMemoryBookingRepo) Delete(id domain.BookingID) error { return nil }
func (r *inMemoryBookingRepo) ListByTherapist(id domain.TherapistID) ([]*booking.Booking, error) {
	return r.bookings, nil
}
func (r *inMemoryBookingRepo) ListByClient(domain.ClientID) ([]*booking.Booking, error) {
	return nil, nil
}
func (r *inMemoryBookingRepo) ListByState(booking.BookingState) ([]*booking.Booking, error) {
	return nil, nil
}
func (r *inMemoryBookingRepo) ListByTherapistAndState(domain.TherapistID, booking.BookingState) ([]*booking.Booking, error) {
	return nil, nil
}
func (r *inMemoryBookingRepo) ListByClientAndState(domain.ClientID, booking.BookingState) ([]*booking.Booking, error) {
	return nil, nil
}
func (r *inMemoryBookingRepo) BulkListByTherapistForDateRange([]domain.TherapistID, booking.BookingState, time.Time, time.Time) (map[domain.TherapistID][]*booking.Booking, error) {
	return nil, nil
}

// Search satisfies the new method in the BookingRepository interface for tests.
func (r *inMemoryBookingRepo) Search(startDate, endDate time.Time, state *booking.BookingState) ([]*booking.Booking, error) {
	// Return all in-memory bookings ignoring filters for simplicity in unit tests.
	return r.bookings, nil
}

type inMemoryTherapistRepo struct{}

func (r *inMemoryTherapistRepo) GetByID(id domain.TherapistID) (*therapist.Therapist, error) {
	// return non-nil dummy therapist
	return &therapist.Therapist{ID: id, Name: "Dr Test"}, nil
}

// other methods stubbed
func (r *inMemoryTherapistRepo) GetByEmail(domain.Email) (*therapist.Therapist, error) {
	return nil, nil
}
func (r *inMemoryTherapistRepo) GetByWhatsAppNumber(domain.WhatsAppNumber) (*therapist.Therapist, error) {
	return nil, nil
}
func (r *inMemoryTherapistRepo) Create(*therapist.Therapist) error { return nil }
func (r *inMemoryTherapistRepo) Update(*therapist.Therapist) error { return nil }
func (r *inMemoryTherapistRepo) UpdateSpecializations(domain.TherapistID, []domain.SpecializationID) error {
	return nil
}
func (r *inMemoryTherapistRepo) Delete(domain.TherapistID) error       { return nil }
func (r *inMemoryTherapistRepo) List() ([]*therapist.Therapist, error) { return nil, nil }
func (r *inMemoryTherapistRepo) FindBySpecializationAndLanguage(string, bool) ([]*therapist.Therapist, error) {
	return nil, nil
}

type inMemoryClientRepo struct{}

func (r *inMemoryClientRepo) GetByID(id domain.ClientID) (*client.Client, error) {
	return &client.Client{ID: id, WhatsAppNumber: "+111"}, nil
}
func (r *inMemoryClientRepo) GetByWhatsAppNumber(domain.WhatsAppNumber) (*client.Client, error) {
	return nil, nil
}
func (r *inMemoryClientRepo) UpdateTimezoneOffset(domain.ClientID, domain.TimezoneOffset) error {
	return nil
}
func (r *inMemoryClientRepo) Create(*client.Client) error     { return nil }
func (r *inMemoryClientRepo) Update(*client.Client) error     { return nil }
func (r *inMemoryClientRepo) Delete(domain.ClientID) error    { return nil }
func (r *inMemoryClientRepo) List() ([]*client.Client, error) { return nil, nil }

type inMemoryTimeSlotRepo struct {
	slots map[domain.TimeSlotID]*timeslot.TimeSlot
}

func (r *inMemoryTimeSlotRepo) GetByID(id domain.TimeSlotID) (*timeslot.TimeSlot, error) {
	return r.slots[id], nil
}
func (r *inMemoryTimeSlotRepo) Create(*timeslot.TimeSlot) error   { return nil }
func (r *inMemoryTimeSlotRepo) Update(*timeslot.TimeSlot) error   { return nil }
func (r *inMemoryTimeSlotRepo) Delete(id domain.TimeSlotID) error { return nil }
func (r *inMemoryTimeSlotRepo) ListByTherapist(therapistID domain.TherapistID) ([]*timeslot.TimeSlot, error) {
	out := make([]*timeslot.TimeSlot, 0, len(r.slots))
	for _, s := range r.slots {
		out = append(out, s)
	}
	return out, nil
}
func (r *inMemoryTimeSlotRepo) BulkToggleByTherapistID(domain.TherapistID, bool) error { return nil }
func (r *inMemoryTimeSlotRepo) BulkListByTherapist(therapistIDs []domain.TherapistID) (map[domain.TherapistID][]*timeslot.TimeSlot, error) {
	out := make(map[domain.TherapistID][]*timeslot.TimeSlot)
	for _, s := range r.slots {
		if _, ok := out[s.TherapistID]; !ok {
			out[s.TherapistID] = []*timeslot.TimeSlot{}
		}

		out[s.TherapistID] = append(out[s.TherapistID], s)
	}
	return out, nil
}

// -----------------------------
// Tests
// -----------------------------

func mustParse(t *testing.T, s string) domain.UTCTimestamp {
	parsed, err := time.Parse(time.RFC3339, s)
	if err != nil {
		t.Fatalf("parse time: %v", err)
	}
	return domain.UTCTimestamp(parsed)
}

func TestCreateBooking_ConflictDetection(t *testing.T) {
	// Common entities
	therapistID := domain.TherapistID("therapist_1")
	clientID := domain.ClientID("client_1")

	tsMorning := &timeslot.TimeSlot{
		ID:                "slot_morning",
		TherapistID:       therapistID,
		Duration:          60,
		PostSessionBuffer: 15,
	}

	tsLateMorning := &timeslot.TimeSlot{
		ID:                "slot_late",
		TherapistID:       therapistID,
		Duration:          60,
		PostSessionBuffer: 15,
	}

	slotRepo := &inMemoryTimeSlotRepo{slots: map[domain.TimeSlotID]*timeslot.TimeSlot{
		tsMorning.ID:     tsMorning,
		tsLateMorning.ID: tsLateMorning,
	}}

	therapistRepo := &inMemoryTherapistRepo{}
	clientRepo := &inMemoryClientRepo{}

	cases := []struct {
		name           string
		existing       []*booking.Booking
		newInput       Input
		expectConflict bool
	}{
		{
			name:     "no conflict - empty schedule",
			existing: nil,
			newInput: Input{
				TherapistID:    therapistID,
				ClientID:       clientID,
				TimeSlotID:     tsMorning.ID,
				StartTime:      mustParse(t, "2025-07-07T09:00:00Z"),
				TimezoneOffset: 0,
			},
			expectConflict: false,
		},
		{
			name: "conflict exact overlap same slot",
			existing: []*booking.Booking{
				{
					TherapistID: therapistID,
					ClientID:    clientID,
					TimeSlotID:  tsMorning.ID,
					StartTime:   mustParse(t, "2025-07-07T09:00:00Z"),
					State:       booking.BookingStateConfirmed,
				},
			},
			newInput: Input{
				TherapistID:    therapistID,
				ClientID:       clientID,
				TimeSlotID:     tsMorning.ID,
				StartTime:      mustParse(t, "2025-07-07T09:00:00Z"),
				TimezoneOffset: 0,
			},
			expectConflict: true,
		},
		{
			name: "conflict due to post buffer",
			existing: []*booking.Booking{
				{
					TherapistID: therapistID,
					ClientID:    clientID,
					TimeSlotID:  tsMorning.ID,
					StartTime:   mustParse(t, "2025-07-07T09:00:00Z"),
					State:       booking.BookingStateConfirmed,
				},
			},
			newInput: Input{
				TherapistID:    therapistID,
				ClientID:       clientID,
				TimeSlotID:     tsMorning.ID,
				StartTime:      mustParse(t, "2025-07-07T10:05:00Z"), // starts within 15-min buffer
				TimezoneOffset: 0,
			},
			expectConflict: true,
		},
		{
			name: "conflict overlap different slots",
			existing: []*booking.Booking{
				{
					TherapistID: therapistID,
					ClientID:    clientID,
					TimeSlotID:  tsMorning.ID,
					StartTime:   mustParse(t, "2025-07-07T09:30:00Z"),
					State:       booking.BookingStatePending,
				},
			},
			newInput: Input{
				TherapistID:    therapistID,
				ClientID:       clientID,
				TimeSlotID:     tsLateMorning.ID,
				StartTime:      mustParse(t, "2025-07-07T10:00:00Z"), // overlaps 10:00-10:30
				TimezoneOffset: 0,
			},
			expectConflict: true,
		},
		{
			name: "no conflict back-to-back respecting buffer",
			existing: []*booking.Booking{
				{
					TherapistID: therapistID,
					ClientID:    clientID,
					TimeSlotID:  tsMorning.ID,
					StartTime:   mustParse(t, "2025-07-07T09:00:00Z"),
					State:       booking.BookingStateConfirmed,
				},
			},
			newInput: Input{
				TherapistID:    therapistID,
				ClientID:       clientID,
				TimeSlotID:     tsMorning.ID,
				StartTime:      mustParse(t, "2025-07-07T10:15:00Z"), // exactly after 15-min buffer
				TimezoneOffset: 0,
			},
			expectConflict: false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			bookingRepo := &inMemoryBookingRepo{bookings: tc.existing}

			uc := NewUsecase(bookingRepo, therapistRepo, clientRepo, slotRepo)
			_, err := uc.Execute(tc.newInput)

			if tc.expectConflict {
				if err != common.ErrTimeSlotAlreadyBooked {
					t.Fatalf("expected conflict error, got %v", err)
				}
			} else {
				if err != nil {
					t.Fatalf("expected success, got %v", err)
				}
			}
		})
	}
}
