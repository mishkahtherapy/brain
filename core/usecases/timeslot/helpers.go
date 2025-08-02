package timeslot_usecase

import (
	"time"

	"github.com/mishkahtherapy/brain/core/domain"
	"github.com/mishkahtherapy/brain/core/domain/timeslot"
)

const MIN_POST_SESSION_BUFFER_MINUTES = 15

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
func ParseTimeString(timeStr domain.Time24h) (time.Time, error) {
	// First check if the string has the exact length expected for HH:MM format
	if len(timeStr) != 5 {
		return time.Time{}, timeslot.ErrInvalidTimeFormat
	}

	// Use Go's time.Parse which automatically rejects AM/PM formats when using 15:04
	// It will return an error like "extra text: ' AM'" for formats like "9:30 AM"
	parsedTime, err := timeStr.ParseTime()
	if err != nil {
		return time.Time{}, timeslot.ErrInvalidTimeFormat
	}

	return parsedTime, nil
}

// Helper function to validate time range (start must be before end)
func ValidateTimeRange(startTime, endTime domain.Time24h) error {
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
func ValidateBufferTimes(preSessionBuffer, postSessionBuffer domain.DurationMinutes) error {
	if preSessionBuffer < 0 {
		return timeslot.ErrPreSessionBufferNegative
	}

	if postSessionBuffer < MIN_POST_SESSION_BUFFER_MINUTES {
		return timeslot.ErrPostSessionBufferTooLow
	}

	return nil
}

// Get actual time range for a time slot (handles cross-day scenarios)
func GetActualTimeRange(slot timeslot.TimeSlot) (start, end time.Time) {
	baseDate := getBaseDateForDay(string(slot.DayOfWeek))
	startTime, _ := slot.Start.ParseTime()
	start = time.Date(baseDate.Year(), baseDate.Month(), baseDate.Day(),
		startTime.Hour(), startTime.Minute(), 0, 0, time.UTC)
	end = start.Add(time.Duration(slot.Duration) * time.Minute)
	return start, end
}

// Check if two time slots have conflicting time ranges
func HasTimeSlotConflict(slot1, slot2 timeslot.TimeSlot) bool {
	// Skip validation if slots are on different days
	if slot1.DayOfWeek != slot2.DayOfWeek {
		return false
	}

	start1, end1 := GetActualTimeRange(slot1)
	start2, end2 := GetActualTimeRange(slot2)
	return start1.Before(end2) && start2.Before(end1)
}

// Get effective time range for a time slot including buffers (handles cross-day scenarios)
func ApplyTimesToReferenceDate(slot timeslot.TimeSlot) (start, end time.Time) {
	baseDate := getBaseDateForDay(string(slot.DayOfWeek))
	startTime, _ := slot.Start.ParseTime()
	sessionStart := time.Date(baseDate.Year(), baseDate.Month(), baseDate.Day(),
		startTime.Hour(), startTime.Minute(), 0, 0, time.UTC)
	sessionEnd := sessionStart.Add(time.Duration(slot.Duration) * time.Minute)

	return sessionStart, sessionEnd
}

// Check if two time slots have sufficient gap between them (at least 30 minutes)
func HasSufficientGapBetweenSlots(slot1, slot2 timeslot.TimeSlot) bool {
	// Skip validation if slots are on different days
	if slot1.DayOfWeek != slot2.DayOfWeek {
		return true
	}

	start1, end1 := ApplyTimesToReferenceDate(slot1)
	start2, end2 := ApplyTimesToReferenceDate(slot2)

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

	start1, end1 := ApplyTimesToReferenceDate(slot1)
	start2, end2 := ApplyTimesToReferenceDate(slot2)
	return start1.Before(end2) && start2.Before(end1)
}

// Validate duration
func ValidateDuration(durationMinutes domain.DurationMinutes) error {
	if durationMinutes <= 0 || durationMinutes > 24*60 {
		return timeslot.ErrInvalidDuration
	}
	return nil
}

// Validate timezone offset (between -12 to +14 hours in minutes)
func ValidateTimezoneOffset(offsetMinutes domain.TimezoneOffset) error {
	if offsetMinutes < -720 || offsetMinutes > 840 {
		return timeslot.ErrInvalidTimezoneOffset
	}
	return nil
}

// Helper function to get base date for a day of week
func getBaseDateForDay(dayOfWeek string) time.Time {
	// Use a reference week starting Sunday 2000-01-02
	baseDate := time.Date(2000, 1, 2, 0, 0, 0, 0, time.UTC)
	dayOffset := getDayOffset(dayOfWeek)
	return baseDate.AddDate(0, 0, dayOffset)
}

// Helper function to get day offset from Sunday
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
