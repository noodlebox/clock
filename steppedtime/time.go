package steppedtime

import (
	"time"
)

// See [time.Duration].
type Duration = time.Duration

// Duration constants.
const (
	Nanosecond  = time.Nanosecond
	Microsecond = time.Microsecond
	Millisecond = time.Millisecond
	Second      = time.Second
	Minute      = time.Minute
	Hour        = time.Hour
)

// Helpers for generating Duration values

// Nanoseconds returns a Duration value representing n nanoseconds.
func (*Clock) Nanoseconds(n int64) Duration {
	return Duration(n * int64(Nanosecond))
}

// Microseconds returns a Duration value representing n microseconds.
func (*Clock) Microseconds(n int64) Duration {
	return Duration(n * int64(Microsecond))
}

// Milliseconds returns a Duration value representing n milliseconds.
func (*Clock) Milliseconds(n int64) Duration {
	return Duration(n * int64(Millisecond))
}

// Seconds returns a Duration value representing n Seconds.
func (*Clock) Seconds(n float64) Duration {
	return Duration(n * float64(Second))
}

// Minutes returns a Duration value representing n Minutes.
func (*Clock) Minutes(n float64) Duration {
	return Duration(n * float64(Minute))
}

// Hours returns a Duration value representing n Hours.
func (*Clock) Hours(n float64) Duration {
	return Duration(n * float64(Hour))
}

// ParseDuration parses a duration string. A duration string is a possibly
// signed sequence of decimal numbers, each with optional fraction and a unit
// suffix, such as "300ms", "-1.5h" or "2h45m". Valid time units are "ns",
// "us" (or "µs"), "ms", "s", "m", "h".
func (*Clock) ParseDuration(s string) (Duration, error) {
	return time.ParseDuration(s)
}

// Time represents the number of nanoseconds since the start of the clock.
type Time int64

// Add returns the time t+d.
func (t Time) Add(d Duration) Time {
	return t + Time(d)
}

// Sub returns the duration t-u.
func (t Time) Sub(u Time) Duration {
	return Duration(t - u)
}

// After reports whether the time instant t is after u.
func (t Time) After(u Time) bool {
	return t > u
}

// Before reports whether the time instant t is before u.
func (t Time) Before(u Time) bool {
	return t < u
}

// Compare compares the time instant t with u. If t is before u, it returns
// -1; if t is after u, it returns +1; if they're the same, it returns 0.
func (t Time) Compare(u Time) int {
	switch {
	case t < u:
		return -1
	case t > u:
		return 1
	}
	return 0
}

// Equal reports whether t and u represent the same time instant.
func (t Time) Equal(u Time) bool {
	return t == u
}

// IsZero reports whether t represents the zero time instant, the start of the clock.
func (t Time) IsZero() bool {
	return t == 0
}
