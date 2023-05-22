package clock

import (
	"time"
)

type Location = time.Location
type Month = time.Month
type Weekday = time.Weekday
type Duration = time.Duration

const (
	Nanosecond  = time.Nanosecond
	Microsecond = time.Microsecond
	Millisecond = time.Millisecond
	Second      = time.Second
	Minute      = time.Minute
	Hour        = time.Hour
)

// Clock[T] is a minimal generic API for a clock that uses a given `Time`
// implementation, T. The standard library's `time.Time` is valid for T here.
type Clock[T Time[T]] interface {
	// Generate `Time`s
	Now() T

	// Generate `Duration`s
	ParseDuration(string) (Duration, error)
	Since(T) Duration
	Until(T) Duration

	// Program flow control
	Sleep(d Duration)

	// Generate `Ticker`s
	NewTicker(d Duration) Ticker[T]
	Tick(d Duration) <-chan T

	// Generate `Timer`s
	NewTimer(Duration) Timer[T]
	After(Duration) <-chan T
	AfterFunc(Duration, func()) Timer[T]
}

// LocatedClock[T] is a generic API for a clock that uses a given
// `LocatedTime` implementation, T. The standard library's `time.Time` is
// valid for T here.
type LocatedClock[T LocatedTime[T]] interface {
	Clock[T]

	// Generate `LocatedTime`s
	Parse(layout, value string) (T, error)
	ParseInLocation(layout, value string, loc *Location) (T, error)
	Date(year int, month Month, day, hour, min, sec, nsec int, loc *Location) T
	Unix(sec int64, nsec int64) T
	UnixMicro(usec int64) T
	UnixMilli(msec int64) T
	UnixNano(nsec int64) T
}

/*
// A Duration represents the elapsed time between two Time values. The
// standard library's `time.Duration` implements `Duration[time.Duration]`.
type Duration[D any] interface {
	Abs() D // go1.19
	Round(D) D
	Truncate(D) D

	// Conversions to standard units
	Nanoseconds() int64
	Microseconds() int64
	Milliseconds() int64
	Seconds() float64
	Minutes() float64
	Hours() float64

	String() string
}
*/

// A Time represents an instant in time marked by the `Clock` that generated
// it. The standard library's `time.Time` implements `Time[time.Time]`.
type Time[T any] interface {
	Add(Duration) T
	Sub(T) Duration

	// Comparisons
	After(T) bool
	Before(T) bool
	//Compare(T) int // go1.20+
	Equal(T) bool
	IsZero() bool
}

// A LocatedTime is a `Time` that additionally has a Location associated with
// it, allowing it to be represented in terrestrial units of time. The standard
// library's `time.Time` implements `LocatedTime[time.Time]`.
type LocatedTime[T any] interface {
	Time[T]

	AppendFormat(b []byte, layout string) []byte
	Clock() (hour, min, sec int)
	Date() (year int, month Month, day int)
	Day() int
	Format(layout string) string
	Hour() int
	ISOWeek() (year, week int)
	In(loc *Location) T
	IsDST() bool
	Local() T
	Location() *Location
	Minute() int
	Month() Month
	Nanosecond() int
	Round(d Duration) T
	Second() int
	Truncate(d Duration) T
	UTC() T
	Unix() int64
	UnixMicro() int64
	UnixMilli() int64
	UnixNano() int64
	Weekday() Weekday
	Year() int
	YearDay() int
	Zone() (name string, offset int)
	ZoneBounds() (start, end T) // go1.19
}

// A Ticker holds a channel that delivers “ticks” of a clock at intervals.
type Ticker[T Time[T]] interface {
	C() <-chan T
	Reset(d Duration)
	Stop()
}

type Timer[T Time[T]] interface {
	C() <-chan T
	Reset(d Duration) bool
	Stop() bool
}
