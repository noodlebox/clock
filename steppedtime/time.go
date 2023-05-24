package steppedtime

import (
	"time"
)

// time's Duration is sufficiently simple to reuse here
type Duration = time.Duration

const (
	Nanosecond  = time.Nanosecond
	Microsecond = time.Microsecond
	Millisecond = time.Millisecond
	Second      = time.Second
	Minute      = time.Minute
	Hour        = time.Hour
)

// Helpers for generating Duration values

func (*Clock) Nanoseconds(n int64) Duration {
	return Duration(n * int64(Nanosecond))
}

func (*Clock) Microseconds(n int64) Duration {
	return Duration(n * int64(Microsecond))
}

func (*Clock) Milliseconds(n int64) Duration {
	return Duration(n * int64(Millisecond))
}

func (*Clock) Seconds(n float64) Duration {
	return Duration(n * float64(Second))
}

func (*Clock) Minutes(n float64) Duration {
	return Duration(n * float64(Minute))
}

func (*Clock) Hours(n float64) Duration {
	return Duration(n * float64(Hour))
}

func (*Clock) ParseDuration(s string) (Duration, error) {
	return time.ParseDuration(s)
}

// The number of nanoseconds since the start of the clock
type Time int64

func (t Time) Add(d Duration) Time {
	return t + Time(d)
}

func (t Time) Sub(u Time) Duration {
	return Duration(t - u)
}

func (t Time) After(u Time) bool {
	return t > u
}

func (t Time) Before(u Time) bool {
	return t < u
}

func (t Time) Compare(u Time) int {
	switch {
	case t < u:
		return -1
	case t > u:
		return 1
	}
	return 0
}

func (t Time) Equal(u Time) bool {
	return t == u
}

func (t Time) IsZero() bool {
	return t == 0
}
