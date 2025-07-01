package timeslot

import (
	"time"

	"github.com/mishkahtherapy/brain/core/domain"
)

// TimeFormat is the standard 24-hour time format (HH:MM) used throughout the application
// This constant can be reused by other packages that need to parse time strings
// Based on Go's reference time: Mon Jan 2 15:04:05 MST 2006
const TimeFormat = "15:04"

// Helper function to validate day of week
func IsValidDayOfWeek(day domain.DayOfWeek) bool {
	validDays := []domain.DayOfWeek{
		domain.DayOfWeekMonday,
		domain.DayOfWeekTuesday,
		domain.DayOfWeekWednesday,
		domain.DayOfWeekThursday,
		domain.DayOfWeekFriday,
		domain.DayOfWeekSaturday,
		domain.DayOfWeekSunday,
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
		return time.Time{}, ErrInvalidTimeFormat
	}

	// Use Go's time.Parse which automatically rejects AM/PM formats when using 15:04
	// It will return an error like "extra text: ' AM'" for formats like "9:30 AM"
	parsedTime, err := time.Parse(TimeFormat, timeStr)
	if err != nil {
		return time.Time{}, ErrInvalidTimeFormat
	}

	return parsedTime, nil
}

// Helper function to validate time range (start must be before end)
func ValidateTimeRange(startTime, endTime string) error {
	start, err := ParseTimeString(startTime)
	if err != nil {
		return ErrInvalidTimeFormat
	}

	end, err := ParseTimeString(endTime)
	if err != nil {
		return ErrInvalidTimeFormat
	}

	if !end.After(start) {
		return ErrInvalidTimeRange
	}

	return nil
}

// Helper function to validate buffer times
func ValidateBufferTimes(preSessionBuffer, postSessionBuffer int) error {
	if preSessionBuffer < 0 {
		return ErrPreSessionBufferNegative
	}

	if postSessionBuffer < 30 {
		return ErrPostSessionBufferTooLow
	}

	return nil
}
