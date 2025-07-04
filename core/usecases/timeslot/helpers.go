package timeslot_usecase

import (
	"time"

	"github.com/mishkahtherapy/brain/core/domain/timeslot"
)

// TimeFormat is the standard 24-hour time format (HH:MM) used throughout the application
// This constant can be reused by other packages that need to parse time strings
// Based on Go's reference time: Mon Jan 2 15:04:05 MST 2006
const TimeFormat = "15:04"

// Helper function to validate day of week
func IsValidDayOfWeek(day timeslot.DayOfWeek) bool {
	validDays := []timeslot.DayOfWeek{
		timeslot.DayOfWeekMonday,
		timeslot.DayOfWeekTuesday,
		timeslot.DayOfWeekWednesday,
		timeslot.DayOfWeekThursday,
		timeslot.DayOfWeekFriday,
		timeslot.DayOfWeekSaturday,
		timeslot.DayOfWeekSunday,
	}

	for _, validDay := range validDays {
		if day == validDay {
			return true
		}
	}
	return false
}

// Helper function to check if two time ranges overlap
func TimesOverlap(start1, end1, start2, end2 time.Time) bool {
	return start1.Before(end2) && start2.Before(end1)
}

// Helper function to parse and validate time format
// This function ensures strict 24-hour HH:MM format and rejects AM/PM formats
func ParseTimeString(timeStr string) (time.Time, error) {
	// First check if the string has the exact length expected for HH:MM format
	if len(timeStr) != 5 {
		return time.Time{}, timeslot.ErrInvalidTimeFormat
	}

	// Use Go's time.Parse which automatically rejects AM/PM formats when using 15:04
	// It will return an error like "extra text: ' AM'" for formats like "9:30 AM"
	parsedTime, err := time.Parse(TimeFormat, timeStr)
	if err != nil {
		return time.Time{}, timeslot.ErrInvalidTimeFormat
	}

	return parsedTime, nil
}

// Helper function to validate time range (start must be before end)
func ValidateTimeRange(startTime, endTime string) error {
	start, err := ParseTimeString(startTime)
	if err != nil {
		return timeslot.ErrInvalidTimeFormat
	}

	end, err := ParseTimeString(endTime)
	if err != nil {
		return timeslot.ErrInvalidTimeFormat
	}

	if !end.After(start) {
		return timeslot.ErrInvalidTimeRange
	}

	return nil
}

// Helper function to validate buffer times
func ValidateBufferTimes(preSessionBuffer, postSessionBuffer int) error {
	if preSessionBuffer < 0 {
		return timeslot.ErrPreSessionBufferNegative
	}

	if postSessionBuffer < 30 {
		return timeslot.ErrPostSessionBufferTooLow
	}

	return nil
}

// ========== NEW TIMEZONE AND DURATION FUNCTIONS ==========

// Convert local day/time to UTC day/time
func ConvertLocalToUTC(localDay string, localStart string, timezoneOffset int) (utcDay string, utcStart string, err error) {
	// Get base date for the local day (using a reference week starting Sunday 2000-01-02)
	baseDate := getBaseDateForDay(localDay)

	// Parse local time
	localTime, err := time.Parse(TimeFormat, localStart)
	if err != nil {
		return "", "", timeslot.ErrInvalidTimeFormat
	}

	// Create local datetime
	localDateTime := time.Date(baseDate.Year(), baseDate.Month(), baseDate.Day(),
		localTime.Hour(), localTime.Minute(), 0, 0, time.UTC)

	// Convert to UTC by subtracting timezone offset
	utcDateTime := localDateTime.Add(-time.Duration(timezoneOffset) * time.Minute)

	return getDayOfWeek(utcDateTime), utcDateTime.Format(TimeFormat), nil
}

// Convert UTC day/time to local day/time
func ConvertUTCToLocal(utcDay string, utcStart string, timezoneOffset int) (localDay string, localStart string, err error) {
	// Get base date for the UTC day
	baseDate := getBaseDateForDay(utcDay)

	// Parse UTC time
	utcTime, err := time.Parse(TimeFormat, utcStart)
	if err != nil {
		return "", "", timeslot.ErrInvalidTimeFormat
	}

	// Create UTC datetime
	utcDateTime := time.Date(baseDate.Year(), baseDate.Month(), baseDate.Day(),
		utcTime.Hour(), utcTime.Minute(), 0, 0, time.UTC)

	// Convert to local by adding timezone offset
	localDateTime := utcDateTime.Add(time.Duration(timezoneOffset) * time.Minute)

	return getDayOfWeek(localDateTime), localDateTime.Format(TimeFormat), nil
}

// Get actual time range for a time slot (handles cross-day scenarios)
func GetActualTimeRange(slot timeslot.TimeSlot) (start, end time.Time) {
	baseDate := getBaseDateForDay(string(slot.DayOfWeek))
	startTime, _ := time.Parse(TimeFormat, slot.StartTime)
	start = time.Date(baseDate.Year(), baseDate.Month(), baseDate.Day(),
		startTime.Hour(), startTime.Minute(), 0, 0, time.UTC)
	end = start.Add(time.Duration(slot.DurationMinutes) * time.Minute)
	return start, end
}

// Check if two time slots have conflicting time ranges
func HasTimeSlotConflict(slot1, slot2 timeslot.TimeSlot) bool {
	start1, end1 := GetActualTimeRange(slot1)
	start2, end2 := GetActualTimeRange(slot2)
	return start1.Before(end2) && start2.Before(end1)
}

// Get effective time range for a time slot including buffers (handles cross-day scenarios)
func GetEffectiveTimeRange(slot timeslot.TimeSlot) (start, end time.Time) {
	baseDate := getBaseDateForDay(string(slot.DayOfWeek))
	startTime, _ := time.Parse(TimeFormat, slot.StartTime)
	sessionStart := time.Date(baseDate.Year(), baseDate.Month(), baseDate.Day(),
		startTime.Hour(), startTime.Minute(), 0, 0, time.UTC)
	sessionEnd := sessionStart.Add(time.Duration(slot.DurationMinutes) * time.Minute)

	// Apply buffers to get effective range
	effectiveStart := sessionStart.Add(-time.Duration(slot.PreSessionBuffer) * time.Minute)
	effectiveEnd := sessionEnd.Add(time.Duration(slot.PostSessionBuffer) * time.Minute)

	return effectiveStart, effectiveEnd
}

// Check if two time slots have sufficient gap between them (at least 30 minutes)
func HasSufficientGapBetweenSlots(slot1, slot2 timeslot.TimeSlot) bool {
	// Skip validation if slots are on different days
	if slot1.DayOfWeek != slot2.DayOfWeek {
		return true
	}

	start1, end1 := GetEffectiveTimeRange(slot1)
	start2, end2 := GetEffectiveTimeRange(slot2)

	// Check if there's at least 30 minutes between the slots
	minGap := 30 * time.Minute

	// If slot1 ends before slot2 starts, check the gap
	if end1.Before(start2) || end1.Equal(start2) {
		return start2.Sub(end1) >= minGap
	}

	// If slot2 ends before slot1 starts, check the gap
	if end2.Before(start1) || end2.Equal(start1) {
		return start1.Sub(end2) >= minGap
	}

	// If neither condition is met, the slots overlap - no sufficient gap
	return false
}

// Check if two time slots have conflicting effective time ranges (including buffers)
func HasEffectiveTimeSlotConflict(slot1, slot2 timeslot.TimeSlot) bool {
	// Skip validation if slots are on different days
	if slot1.DayOfWeek != slot2.DayOfWeek {
		return false
	}

	start1, end1 := GetEffectiveTimeRange(slot1)
	start2, end2 := GetEffectiveTimeRange(slot2)
	return start1.Before(end2) && start2.Before(end1)
}

// Validate duration
func ValidateDuration(durationMinutes int) error {
	if durationMinutes <= 0 || durationMinutes > 24*60 {
		return timeslot.ErrInvalidDuration
	}
	return nil
}

// Validate timezone offset (between -12 to +14 hours in minutes)
func ValidateTimezoneOffset(offsetMinutes int) error {
	if offsetMinutes < -720 || offsetMinutes > 840 {
		return timeslot.ErrInvalidTimezoneOffset
	}
	return nil
}

// Calculate end time from start time and duration
func CalculateEndTime(startTime string, durationMinutes int) (endTime string, crossesDay bool) {
	start, _ := time.Parse(TimeFormat, startTime)
	baseDate := time.Date(2000, 1, 1, start.Hour(), start.Minute(), 0, 0, time.UTC)
	end := baseDate.Add(time.Duration(durationMinutes) * time.Minute)

	crossesDay = end.Day() != baseDate.Day()
	return end.Format(TimeFormat), crossesDay
}

// ========== HELPER UTILITY FUNCTIONS ==========

// Get base date for a day of week (using reference week starting Sunday 2000-01-02)
func getBaseDateForDay(dayOfWeek string) time.Time {
	baseDate := time.Date(2000, 1, 2, 0, 0, 0, 0, time.UTC) // Sunday

	dayOffset := getDayOffset(dayOfWeek)
	return baseDate.AddDate(0, 0, dayOffset)
}

// Get day offset from Sunday (0=Sunday, 1=Monday, etc.)
func getDayOffset(dayOfWeek string) int {
	switch dayOfWeek {
	case "Sunday":
		return 0
	case "Monday":
		return 1
	case "Tuesday":
		return 2
	case "Wednesday":
		return 3
	case "Thursday":
		return 4
	case "Friday":
		return 5
	case "Saturday":
		return 6
	default:
		return 0
	}
}

// Get day of week string from time.Time
func getDayOfWeek(t time.Time) string {
	switch t.Weekday() {
	case time.Sunday:
		return "Sunday"
	case time.Monday:
		return "Monday"
	case time.Tuesday:
		return "Tuesday"
	case time.Wednesday:
		return "Wednesday"
	case time.Thursday:
		return "Thursday"
	case time.Friday:
		return "Friday"
	case time.Saturday:
		return "Saturday"
	default:
		return "Sunday"
	}
}
