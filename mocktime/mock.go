package mocktime

import (
	"github.com/noodlebox/clock/realtime"
	"github.com/noodlebox/clock/relativetime"
)

type Time = realtime.Time
type Duration = realtime.Duration
type Location = realtime.Location
type Month = realtime.Month
type Weekday = realtime.Weekday
type Timer = relativetime.Timer[Time, Duration]
type Ticker = relativetime.Ticker[Time, Duration]

// Duration constants
const (
	Nanosecond  = realtime.Nanosecond
	Microsecond = realtime.Microsecond
	Millisecond = realtime.Millisecond
	Second      = realtime.Second
	Minute      = realtime.Minute
	Hour        = realtime.Hour
)

// Location constants
var UTC = realtime.UTC
var Local = UTC

// Month constants
const (
	January   = realtime.January
	February  = realtime.February
	March     = realtime.March
	April     = realtime.April
	May       = realtime.May
	June      = realtime.June
	July      = realtime.July
	August    = realtime.August
	September = realtime.September
	October   = realtime.October
	November  = realtime.November
	December  = realtime.December
)

// Weekday constants
const (
	Sunday    = realtime.Sunday
	Monday    = realtime.Monday
	Tuesday   = realtime.Tuesday
	Wednesday = realtime.Wednesday
	Thursday  = realtime.Thursday
	Friday    = realtime.Friday
	Saturday  = realtime.Saturday
)

// Layouts
const (
	Layout      = "01/02 03:04:05PM '06 -0700" // The reference realtime, in numerical order.
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

// Wrap package-level functions around Clock methods

var clock Clock

func init() {
	clock = NewClockAt(realtime.Clock{}.Date(
		2009, November, 10, 23, 0, 0, 0, UTC,
	))
	clock.Start()
}

func After(d Duration) <-chan Time             { return clock.After(d) }
func Sleep(d Duration)                         { clock.Sleep(d) }
func Tick(d Duration) <-chan Time              { return clock.Tick(d) }
func ParseDuration(s string) (Duration, error) { return clock.ParseDuration(s) }
func Since(t Time) Duration                    { return clock.Since(t) }
func Until(t Time) Duration                    { return clock.Until(t) }
func NewTicker(d Duration) *Ticker             { return clock.NewTicker(d) }
func Date(year int, month Month, day, hour, min, sec, nsec int, loc *Location) Time {
	return clock.Date(year, month, day, hour, min, sec, nsec, loc)
}
func Now() Time                                { return clock.Now() }
func Parse(layout, value string) (Time, error) { return clock.Parse(layout, value) }
func ParseInLocation(layout, value string, loc *Location) (Time, error) {
	return clock.ParseInLocation(layout, value, loc)
}
func Unix(sec int64, nsec int64) Time       { return clock.Unix(sec, nsec) }
func UnixMicro(usec int64) Time             { return clock.UnixMicro(usec) }
func UnixMilli(msec int64) Time             { return clock.UnixMilli(msec) }
func AfterFunc(d Duration, f func()) *Timer { return clock.AfterFunc(d, f) }
func NewTimer(d Duration) *Timer            { return clock.NewTimer(d) }
