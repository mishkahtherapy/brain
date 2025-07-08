package timeslot

import "github.com/mishkahtherapy/brain/core/domain"

type DayOfWeek string

const (
	DayOfWeekMonday    DayOfWeek = "Monday"
	DayOfWeekTuesday   DayOfWeek = "Tuesday"
	DayOfWeekWednesday DayOfWeek = "Wednesday"
	DayOfWeekThursday  DayOfWeek = "Thursday"
	DayOfWeekFriday    DayOfWeek = "Friday"
	DayOfWeekSaturday  DayOfWeek = "Saturday"
	DayOfWeekSunday    DayOfWeek = "Sunday"
)

type TimeSlot struct {
	ID                domain.TimeSlotID     `json:"id"`
	TherapistID       domain.TherapistID    `json:"therapistId"`
	IsActive          bool                  `json:"isActive"`
	DayOfWeek         DayOfWeek             `json:"dayOfWeek"`         // UTC day
	StartTime         string                `json:"startTime"`         // UTC time e.g. "22:30"
	DurationMinutes   int                   `json:"durationMinutes"`   // Duration in minutes e.g. 60
	PreSessionBuffer  int                   `json:"preSessionBuffer"`  // minutes (advance notice), used only when preparing schedule.
	PostSessionBuffer int                   `json:"postSessionBuffer"` // minutes (break after session).
	TimezoneOffset    domain.TimezoneOffset `json:"timezoneOffset"`    // Client's timezone offset in minutes from UTC
	BookingIDs        []domain.BookingID    `json:"bookingIds"`
	CreatedAt         domain.UTCTimestamp   `json:"createdAt"`
	UpdatedAt         domain.UTCTimestamp   `json:"updatedAt"`
}

// Helper method to convert buffer to minutes for calculations
func (ts *TimeSlot) PreSessionBufferInMinutes() int {
	return ts.PreSessionBuffer
}

func (ts *TimeSlot) PostSessionBufferInMinutes() int {
	return ts.PostSessionBuffer
}
