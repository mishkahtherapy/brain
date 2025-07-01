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
	ID                domain.TimeSlotID   `json:"id"`
	TherapistID       domain.TherapistID  `json:"therapistId"`
	DayOfWeek         DayOfWeek           `json:"dayOfWeek"`
	StartTime         string              `json:"startTime"`         // e.g. "15:00"
	EndTime           string              `json:"endTime"`           // e.g. "16:00"
	PreSessionBuffer  int                 `json:"preSessionBuffer"`  // minutes
	PostSessionBuffer int                 `json:"postSessionBuffer"` // minutes
	BookingIDs        []domain.BookingID  `json:"bookingIds"`
	CreatedAt         domain.UTCTimestamp `json:"createdAt"`
	UpdatedAt         domain.UTCTimestamp `json:"updatedAt"`
}

// Helper method to convert buffer to minutes for calculations
func (ts *TimeSlot) PreSessionBufferInMinutes() int {
	return ts.PreSessionBuffer
}

func (ts *TimeSlot) PostSessionBufferInMinutes() int {
	return ts.PostSessionBuffer
}
