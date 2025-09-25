package overlap_detector

import "time"

type OverlapDetector struct {
	baseStart time.Time
	baseEnd   time.Time
}

func New(baseStart time.Time, baseEnd time.Time) *OverlapDetector {
	return &OverlapDetector{
		baseStart: baseStart,
		baseEnd:   baseEnd,
	}
}

func (c *OverlapDetector) HasOverlap(
	otherStart time.Time,
	otherEnd time.Time,
) bool {
	if between(otherStart, c.baseStart, c.baseEnd) {
		return true
	}
	if between(otherEnd, c.baseStart, c.baseEnd) {
		return true
	}

	if between(c.baseStart, otherStart, otherEnd) {
		return true
	}

	if between(c.baseEnd, otherStart, otherEnd) {
		return true
	}

	return false
}

func between(x time.Time, start time.Time, end time.Time) bool {
	return x.Before(end) && x.After(start)
}

// func (c *ConflictDetector) HasConflict(
// 	otherStart time.Time,
// 	otherEnd time.Time,
// ) bool {
// 	firstEnd := firstStart.Add(time.Duration(firstDuration) * time.Minute)
// 	otherEnd := otherStart.Add(time.Duration(otherDuration) * time.Minute)

// 	// Back to back
// 	if firstEnd.Equal(otherStart) || firstStart.Equal(otherEnd) {
// 		return false
// 	}

// 	if between(firstStart, otherStart, otherEnd) {
// 		return true
// 	}
// 	if between(firstEnd, otherStart, otherEnd) {
// 		return true
// 	}

// 	return otherEnd.Before(firstStart) && firstEnd.Before(otherStart)

// }

// func between(x time.Time, start time.Time, end time.Time) bool {
// 	return x.Before(end) && x.After(start)
// }
