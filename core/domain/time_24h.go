package domain

import "time"

// TimeFormat is the standard 24-hour time format (HH:MM) used throughout the application
// This constant can be reused by other packages that need to parse time strings
// Based on Go's reference time: Mon Jan 2 15:04:05 MST 2006
var Time24hLayout = "15:04"

type Time24h string

func NewTime24h(raw string) Time24h {
	parsedTime, err := time.Parse(Time24hLayout, raw)
	if err != nil {
		panic(err)
	}
	return Time24h(parsedTime.Format(Time24hLayout))
}

func ValidateTime24h(raw string) error {
	_, err := time.Parse(Time24hLayout, raw)
	if err != nil {
		return err
	}
	return nil
}

func (i Time24h) ParseTime() (time.Time, error) {
	parsedTime, err := time.Parse(Time24hLayout, string(i))
	if err != nil {
		return time.Time{}, err
	}
	return parsedTime.UTC(), nil
}
