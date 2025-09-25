package overlap_detector

import (
	"testing"
	"time"
)

func TestTimeslotOverlappsBooking(t *testing.T) {
	cases := []struct {
		name           string
		firstStart     time.Time
		firstEnd       time.Time
		otherStart     time.Time
		otherEnd       time.Time
		expectConflict bool
	}{
		{
			name:           "no conflict - first back to other start",
			firstStart:     mustParse(t, "10:00:00"),
			firstEnd:       mustParse(t, "11:00:00"),
			otherStart:     mustParse(t, "11:00:00"),
			otherEnd:       mustParse(t, "12:00:00"),
			expectConflict: false,
		},
		{
			name:           "no conflict - other back to first start",
			firstStart:     mustParse(t, "10:00:00"),
			firstEnd:       mustParse(t, "11:00:00"),
			otherStart:     mustParse(t, "09:00:00"),
			otherEnd:       mustParse(t, "10:00:00"),
			expectConflict: false,
		},
		{
			name:           "separated - first ends before other starts",
			firstStart:     mustParse(t, "10:00:00"),
			firstEnd:       mustParse(t, "11:00:00"),
			otherStart:     mustParse(t, "12:00:00"),
			otherEnd:       mustParse(t, "13:00:00"),
			expectConflict: false,
		},

		{
			name:           "separated - other ends before first starts",
			firstStart:     mustParse(t, "10:00:00"),
			firstEnd:       mustParse(t, "11:00:00"),
			otherStart:     mustParse(t, "08:00:00"),
			otherEnd:       mustParse(t, "09:00:00"),
			expectConflict: false,
		},

		{
			name:           "first start overlap",
			firstStart:     mustParse(t, "10:00:00"),
			firstEnd:       mustParse(t, "10:30:00"),
			otherStart:     mustParse(t, "10:00:00"),
			otherEnd:       mustParse(t, "11:00:00"),
			expectConflict: true,
		},
		{
			name:           "first end overlap",
			firstStart:     mustParse(t, "09:00:00"),
			firstEnd:       mustParse(t, "10:00:00"),
			otherStart:     mustParse(t, "09:30:00"),
			otherEnd:       mustParse(t, "10:30:00"),
			expectConflict: true,
		},
		{
			name:           "other start overlap",
			firstStart:     mustParse(t, "10:00:00"),
			firstEnd:       mustParse(t, "12:00:00"),
			otherStart:     mustParse(t, "11:00:00"),
			otherEnd:       mustParse(t, "13:00:00"),
			expectConflict: true,
		},
		{
			name:           "other end overlap",
			firstStart:     mustParse(t, "10:00:00"),
			firstEnd:       mustParse(t, "12:00:00"),
			otherStart:     mustParse(t, "09:00:00"),
			otherEnd:       mustParse(t, "11:00:00"),
			expectConflict: true,
		},
		{
			name:           "conflict - otherswallows",
			firstStart:     mustParse(t, "10:00:00"),
			firstEnd:       mustParse(t, "11:00:00"),
			otherStart:     mustParse(t, "09:00:00"),
			otherEnd:       mustParse(t, "12:00:00"),
			expectConflict: true,
		},
		{
			name:           "conflict - firstswallows",
			firstStart:     mustParse(t, "09:00:00"),
			firstEnd:       mustParse(t, "12:00:00"),
			otherStart:     mustParse(t, "10:00:00"),
			otherEnd:       mustParse(t, "11:00:00"),
			expectConflict: true,
		},
		{
			name:           "conflict - full overlap",
			firstStart:     mustParse(t, "09:00:00"),
			firstEnd:       mustParse(t, "12:00:00"),
			otherStart:     mustParse(t, "09:00:00"),
			otherEnd:       mustParse(t, "12:00:00"),
			expectConflict: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			conflictDetector := New(tc.firstStart, tc.firstEnd)
			conflict := conflictDetector.HasOverlap(tc.otherStart, tc.otherEnd)
			if conflict != tc.expectConflict {
				t.Errorf(
					"expected conflict %v, got %v: startDate=%v, endDate=%v, otherStartDate=%v, otherEndDate=%v",
					tc.expectConflict,
					conflict,
					tc.firstStart.Format(time.RFC3339),
					tc.firstEnd.Format(time.RFC3339),
					tc.otherStart.Format(time.RFC3339),
					tc.otherEnd.Format(time.RFC3339),
				)
			}
		})
	}
}

func mustParse(t *testing.T, s string) time.Time {
	parsed, err := time.Parse(time.TimeOnly, s)
	if err != nil {
		t.Fatalf("parse time: %v", err)
	}
	return parsed
}
