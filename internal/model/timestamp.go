package model

import (
	"fmt"
	"strings"
	"time"
)

// wireLayout renders an instant as UTC with exactly three fractional digits and
// an explicit "+00:00" offset, e.g. 2017-10-02T04:58:58.508+00:00.
const wireLayout = "2006-01-02T15:04:05.000"

// storageLayout is how instants are written back to the database.
const storageLayout = "2006-01-02 15:04:05.000-07:00"

// parseLayouts covers the forms the seed data and our own writes produce.
var parseLayouts = []string{
	"2006-01-02 15:04:05.000-07",
	"2006-01-02 15:04:05.000-07:00",
	"2006-01-02 15:04:05-07",
	"2006-01-02 15:04:05-07:00",
	time.RFC3339Nano,
}

// Timestamp is an instant that serializes in the application's wire format.
type Timestamp time.Time

// Time returns the underlying instant.
func (t Timestamp) Time() time.Time { return time.Time(t) }

// IsZero reports whether the instant is unset.
func (t Timestamp) IsZero() bool { return time.Time(t).IsZero() }

// MarshalJSON renders the instant in UTC with millisecond precision.
func (t Timestamp) MarshalJSON() ([]byte, error) {
	return []byte(`"` + time.Time(t).UTC().Format(wireLayout) + `+00:00"`), nil
}

// String renders the instant in the database storage format.
func (t Timestamp) String() string { return time.Time(t).Format(storageLayout) }

// ParseTimestamp reads any of the instant formats the application stores.
func ParseTimestamp(s string) (Timestamp, error) {
	s = strings.TrimSpace(s)
	for _, layout := range parseLayouts {
		if v, err := time.Parse(layout, s); err == nil {
			return Timestamp(v), nil
		}
	}
	return Timestamp{}, fmt.Errorf("model: cannot parse timestamp %q", s)
}
