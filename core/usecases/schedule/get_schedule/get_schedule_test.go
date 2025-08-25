package get_schedule

import (
	"testing"
	"time"

	"github.com/mishkahtherapy/brain/core/domain"
	"github.com/mishkahtherapy/brain/core/domain/schedule"
)

func TestSplitTimeSlotWithBookings(t *testing.T) {
	nowTime, err := time.Parse(time.RFC3339, "2025-01-01T09:00:00Z")
	if err != nil {
		t.Fatalf("failed to parse time: %v", err)
	}
	now := domain.UTCTimestamp(nowTime)
	tests := []struct {
		name                            string
		slot                            timeRange
		bookings                        []timeRange
		afterSessionBreakTime           domain.AfterSessionBreakTimeMinutes
		timeRangeMinimumDurationMinutes domain.DurationMinutes
		expected                        []schedule.AvailableTimeRange
	}{
		{
			name: "no bookings",
			slot: timeRange{
				start: now,
				end:   now.Add(time.Hour),
			},
			afterSessionBreakTime:           10,
			timeRangeMinimumDurationMinutes: 0,
			expected: []schedule.AvailableTimeRange{
				{
					From: now,
					To:   now.Add(time.Hour),
				},
			},
		},
		{
			name: "full slot booking",
			slot: timeRange{
				start: now,
				end:   now.Add(time.Hour),
			},
			bookings: []timeRange{
				{
					start: now,
					end:   now.Add(time.Hour),
				},
			},
			afterSessionBreakTime:           10,
			timeRangeMinimumDurationMinutes: 0,
			expected:                        []schedule.AvailableTimeRange{},
		},
		{
			name: "partial slot booking exceeding time range minimum duration",
			slot: timeRange{
				start: now,
				end:   now.Add(time.Hour),
			},
			bookings: []timeRange{
				{
					start: now,
					end:   now.Add(time.Minute * 25),
				},
			},
			afterSessionBreakTime:           0,
			timeRangeMinimumDurationMinutes: 35,
			expected:                        []schedule.AvailableTimeRange{},
		},
		{
			name: "partial slot booking from slot start",
			slot: timeRange{
				start: now,
				end:   now.Add(time.Hour),
			},
			bookings: []timeRange{
				{
					start: now,
					end:   now.Add(time.Minute * 15),
				},
			},
			afterSessionBreakTime:           10,
			timeRangeMinimumDurationMinutes: 0,
			expected: []schedule.AvailableTimeRange{
				{
					From: now.Add(time.Minute * 25),
					To:   now.Add(time.Hour),
				},
			},
		},
		{
			name: "partial slot booking from slot end",
			slot: timeRange{
				start: now,
				end:   now.Add(time.Hour),
			},
			bookings: []timeRange{
				{
					start: now.Add(time.Hour).Add(-time.Minute * 15),
					end:   now.Add(time.Hour),
				},
			},
			afterSessionBreakTime:           10,
			timeRangeMinimumDurationMinutes: 0,
			expected: []schedule.AvailableTimeRange{
				{
					From: now,
					To:   now.Add(time.Hour).Add(-time.Minute * 25),
				},
			},
		},
		{
			name: "partial single slot booking in middle",
			slot: timeRange{
				start: now,
				end:   now.Add(time.Hour),
			},
			bookings: []timeRange{
				{
					start: now.Add(time.Minute * 15),
					end:   now.Add(time.Minute * 30),
				},
			},
			afterSessionBreakTime:           10,
			timeRangeMinimumDurationMinutes: 0,
			expected: []schedule.AvailableTimeRange{
				{
					From: now,
					To:   now.Add(time.Minute * 5),
				},
				{
					From: now.Add(time.Minute * 40),
					To:   now.Add(time.Hour),
				},
			},
		},
		{
			name: "partial multiple slot booking in middle",
			slot: timeRange{
				start: now,
				end:   now.Add(2 * time.Hour),
			},
			bookings: []timeRange{
				{
					start: now.Add(time.Minute * 30),
					end:   now.Add(time.Minute * 45),
				},
				{
					start: now.Add(time.Hour).Add(time.Minute * 15),
					end:   now.Add(time.Hour).Add(time.Minute * 25),
				},
			},
			afterSessionBreakTime:           10,
			timeRangeMinimumDurationMinutes: 0,
			expected: []schedule.AvailableTimeRange{
				{
					From: now,
					To:   now.Add(time.Minute * 20),
				},
				{
					From: now.Add(time.Minute * 55),
					To:   now.Add(time.Hour).Add(time.Minute * 5),
				},
				{
					From: now.Add(time.Hour).Add(time.Minute * 35),
					To:   now.Add(time.Hour * 2),
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			actual := findInterBookingAvailabilities(test.slot, test.afterSessionBreakTime, test.bookings, test.timeRangeMinimumDurationMinutes)
			if len(actual) != len(test.expected) {
				t.Errorf("expected %d available ranges, got %d", len(test.expected), len(actual))
			}
			// Check equality
			for i, expected := range test.expected {
				if expected.From != actual[i].From {
					t.Errorf("expected from %s, got %s", expected.From, actual[i].From)
				}
				if expected.To != actual[i].To {
					t.Errorf("expected to %s, got %s", expected.To, actual[i].To)
				}
			}
		})
	}
}
