package timeslot

import (
	"time"

	"github.com/mishkahtherapy/brain/core/domain"
)

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

func MapToDayOfWeek(dayOfWeek time.Weekday) DayOfWeek {
	days := map[time.Weekday]DayOfWeek{
		time.Monday:    DayOfWeekMonday,
		time.Tuesday:   DayOfWeekTuesday,
		time.Wednesday: DayOfWeekWednesday,
		time.Thursday:  DayOfWeekThursday,
		time.Friday:    DayOfWeekFriday,
		time.Saturday:  DayOfWeekSaturday,
		time.Sunday:    DayOfWeekSunday,
	}

	return days[dayOfWeek]
}

type TimeSlot struct {
	ID                    domain.TimeSlotID                   `json:"id"`
	TherapistID           domain.TherapistID                  `json:"therapistId"`
	IsActive              bool                                `json:"isActive"`
	DayOfWeek             DayOfWeek                           `json:"dayOfWeek"`             // UTC day
	Start                 domain.Time24h                      `json:"start"`                 // UTC time e.g. "22:30"
	Duration              domain.DurationMinutes              `json:"duration"`              // Duration in minutes e.g. 60
	AdvanceNotice         domain.AdvanceNoticeMinutes         `json:"advanceNotice"`         // minutes (advance notice), used only when preparing schedule.
	AfterSessionBreakTime domain.AfterSessionBreakTimeMinutes `json:"afterSessionBreakTime"` // minutes (break after session).
	BookingIDs            []domain.BookingID                  `json:"bookingIds"`
	CreatedAt             domain.UTCTimestamp                 `json:"createdAt"`
	UpdatedAt             domain.UTCTimestamp                 `json:"updatedAt"`
}

func (ts *TimeSlot) ApplyToDate(date time.Time) (domain.UTCTimestamp, domain.UTCTimestamp) {
	slotStartTime, err := ts.Start.ParseTime()
	if err != nil {
		panic(err)
	}
	start := time.Date(date.Year(), date.Month(), date.Day(), slotStartTime.Hour(), slotStartTime.Minute(), 0, 0, time.UTC)
	end := start.Add(time.Duration(ts.Duration) * time.Minute)
	return domain.UTCTimestamp(start), domain.UTCTimestamp(end)
}
