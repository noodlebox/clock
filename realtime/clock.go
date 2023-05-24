package realtime

import (
	"time"
)

type Time = time.Time
type Duration = time.Duration
type Location = time.Location
type Month = time.Month
type Weekday = time.Weekday

// Duration constants
const (
	Nanosecond  = time.Nanosecond
	Microsecond = time.Microsecond
	Millisecond = time.Millisecond
	Second      = time.Second
	Minute      = time.Minute
	Hour        = time.Hour
)

// Month constants
const (
	January   = time.January
	February  = time.February
	March     = time.March
	April     = time.April
	May       = time.May
	June      = time.June
	July      = time.July
	August    = time.August
	September = time.September
	October   = time.October
	November  = time.November
	December  = time.December
)

// Weekday constants
const (
	Sunday    = time.Sunday
	Monday    = time.Monday
	Tuesday   = time.Tuesday
	Wednesday = time.Wednesday
	Thursday  = time.Thursday
	Friday    = time.Friday
	Saturday  = time.Saturday
)

// Layouts
const (
	Layout      = "01/02 03:04:05PM '06 -0700" // The reference time, in numerical order.
	ANSIC       = "Mon Jan _2 15:04:05 2006"
	UnixDate    = "Mon Jan _2 15:04:05 MST 2006"
	RubyDate    = "Mon Jan 02 15:04:05 -0700 2006"
	RFC822      = "02 Jan 06 15:04 MST"
	RFC822Z     = "02 Jan 06 15:04 -0700" // RFC822 with numeric zone
	RFC850      = "Monday, 02-Jan-06 15:04:05 MST"
	RFC1123     = "Mon, 02 Jan 2006 15:04:05 MST"
	RFC1123Z    = "Mon, 02 Jan 2006 15:04:05 -0700" // RFC1123 with numeric zone
	RFC3339     = "2006-01-02T15:04:05Z07:00"
	RFC3339Nano = "2006-01-02T15:04:05.999999999Z07:00"
	Kitchen     = "3:04PM"
	// Handy time stamps.
	Stamp      = "Jan _2 15:04:05"
	StampMilli = "Jan _2 15:04:05.000"
	StampMicro = "Jan _2 15:04:05.000000"
	StampNano  = "Jan _2 15:04:05.000000000"
	DateTime   = "2006-01-02 15:04:05"
	DateOnly   = "2006-01-02"
	TimeOnly   = "15:04:05"
)

// Wraps package level functions from `time` to implement
// `clock.LocatedClock[time.Time]`
type Clock struct{}

func NewClock() Clock {
	return Clock{}
}

// Helpers for generating Duration values

func (Clock) Nanoseconds(n int64) Duration {
	return Duration(n * int64(Nanosecond))
}

func (Clock) Microseconds(n int64) Duration {
	return Duration(n * int64(Microsecond))
}

func (Clock) Milliseconds(n int64) Duration {
	return Duration(n * int64(Millisecond))
}

func (Clock) Seconds(n float64) Duration {
	return Duration(n * float64(Second))
}

func (Clock) Minutes(n float64) Duration {
	return Duration(n * float64(Minute))
}

func (Clock) Hours(n float64) Duration {
	return Duration(n * float64(Hour))
}

// Wrappers for `time` package functions

func (Clock) Now() Time {
	return time.Now()
}

func (Clock) ParseDuration(s string) (Duration, error) {
	return time.ParseDuration(s)
}

func (Clock) Since(t Time) Duration {
	return time.Since(t)
}

func (Clock) Until(t Time) Duration {
	return time.Until(t)
}

func (Clock) Sleep(d Duration) {
	time.Sleep(d)
}

// Wraps time.Ticker to complete interfaceable implementation
type Ticker struct {
	*time.Ticker
}

func (t *Ticker) C() <-chan Time {
	return t.Ticker.C
}

func (Clock) NewTicker(d Duration) *Ticker {
	return &Ticker{time.NewTicker(d)}
}

func (Clock) Tick(d Duration) <-chan Time {
	return time.Tick(d)
}

// Wraps time.Timer to complete interfaceable implementation
type Timer struct {
	*time.Timer
}

func (t *Timer) C() <-chan Time {
	return t.Timer.C
}

func (Clock) NewTimer(d Duration) *Timer {
	return &Timer{time.NewTimer(d)}
}

func (Clock) After(d Duration) <-chan Time {
	return time.After(d)
}

func (Clock) AfterFunc(d Duration, f func()) *Timer {
	return &Timer{time.AfterFunc(d, f)}
}

// Wall clock (Location dependent) implementation

func (Clock) Parse(layout, value string) (Time, error) {
	return time.Parse(layout, value)
}

func (Clock) ParseInLocation(layout, value string, loc *Location) (Time, error) {
	return time.ParseInLocation(layout, value, loc)
}

func (Clock) Date(year int, month Month, day, hour, min, sec, nsec int, loc *Location) Time {
	return time.Date(year, month, day, hour, min, sec, nsec, loc)
}

func (Clock) Unix(sec int64, nsec int64) Time {
	return time.Unix(sec, nsec)
}

func (Clock) UnixMicro(usec int64) Time {
	return time.UnixMicro(usec)
}

func (Clock) UnixMilli(msec int64) Time {
	return time.UnixMilli(msec)
}

func (Clock) UnixNano(nsec int64) Time {
	return time.Unix(0, nsec)
}
