package domain

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"
)

type UTCTimestamp time.Time

func NewUTCTimestamp() UTCTimestamp {
	return UTCTimestamp(time.Now().UTC().Round(time.Second))
}

func (i UTCTimestamp) Time() time.Time {
	return time.Time(i)
}

func (i UTCTimestamp) Year() int {
	return time.Time(i).Year()
}

func (i UTCTimestamp) Month() time.Month {
	return time.Time(i).Month()
}

func (i UTCTimestamp) Day() int {
	return time.Time(i).Day()
}

func (i UTCTimestamp) Hour() int {
	return time.Time(i).Hour()
}

func (i UTCTimestamp) Minute() int {
	return time.Time(i).Minute()
}

func (i UTCTimestamp) Before(other UTCTimestamp) bool {
	return time.Time(i).Before(time.Time(other))
}

func (i UTCTimestamp) After(other UTCTimestamp) bool {
	return time.Time(i).After(time.Time(other))
}

func (i UTCTimestamp) Equal(other UTCTimestamp) bool {
	return time.Time(i).Equal(time.Time(other))
}

func (i UTCTimestamp) Add(d time.Duration) UTCTimestamp {
	return UTCTimestamp(time.Time(i).Add(d))
}

func (i UTCTimestamp) Sub(other UTCTimestamp) time.Duration {
	return time.Time(i).Sub(time.Time(other))
}

func (i UTCTimestamp) Format(layout string) string {
	return time.Time(i).UTC().Format(layout)
}

// MarshalJSON implements json.Marshaler interface
func (i UTCTimestamp) MarshalJSON() ([]byte, error) {
	t := time.Time(i).UTC()
	return json.Marshal(t.Format(time.RFC3339))
}

func (i UTCTimestamp) String() string {
	return time.Time(i).UTC().Format(time.RFC3339)
}

// UnmarshalJSON implements json.Unmarshaler interface
func (i *UTCTimestamp) UnmarshalJSON(data []byte) error {
	var timeStr string
	if err := json.Unmarshal(data, &timeStr); err != nil {
		return err
	}

	t, err := time.Parse(time.RFC3339, timeStr)
	if err != nil {
		return err
	}

	*i = UTCTimestamp(t.UTC())
	return nil
}

func (i UTCTimestamp) Value() (driver.Value, error) {
	converted := time.Time(i).UTC().Round(time.Second)
	return driver.Value(converted), nil
}

func (i *UTCTimestamp) Scan(value interface{}) error {
	if value == nil {
		return nil
	}

	var result time.Time
	switch v := value.(type) {
	case time.Time:
		result = v
	case string:
		var err error
		result, err = time.Parse(time.RFC3339, v)
		if err != nil {
			return fmt.Errorf("failed to parse time string: %w", err)
		}
	default:
		return fmt.Errorf("unsupported type for ExpiryTime: %T", value)
	}

	*i = UTCTimestamp(result)
	return nil
}
