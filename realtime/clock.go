package realtime

import (
	"time"
)

// See [time.Time].
type Time = time.Time

// See [time.Duration].
type Duration = time.Duration

// See [time.Location].
type Location = time.Location

// See [time.Month].
type Month = time.Month

// See [time.Weekday].
type Weekday = time.Weekday

// See [time.ParseError].
type ParseError = time.ParseError

// Duration constants.
const (
	Nanosecond  = time.Nanosecond
	Microsecond = time.Microsecond
	Millisecond = time.Millisecond
	Second      = time.Second
	Minute      = time.Minute
	Hour        = time.Hour
)

// See [time.UTC].
var UTC = time.UTC

// See [time.Local].
var Local = time.Local

// Month constants.
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

// Weekday constants.
const (
	Sunday    = time.Sunday
	Monday    = time.Monday
	Tuesday   = time.Tuesday
	Wednesday = time.Wednesday
	Thursday  = time.Thursday
	Friday    = time.Friday
	Saturday  = time.Saturday
)

// Layouts (See [time.Layout]).
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

// Clock wraps package-level functions from [time]. Its methods are
// thread-safe and Clock objects may be copied freely. The zero-value of a
// Clock is perfectly valid.
type Clock struct{}

// NewClock returns a new Clock.
func NewClock() Clock {
	return Clock{}
}

// Helpers for generating Duration values

// Nanoseconds returns a Duration value representing n nanoseconds.
func (Clock) Nanoseconds(n int64) Duration {
	return Duration(n * int64(Nanosecond))
}

// Microseconds returns a Duration value representing n microseconds.
func (Clock) Microseconds(n int64) Duration {
	return Duration(n * int64(Microsecond))
}

// Milliseconds returns a Duration value representing n milliseconds.
func (Clock) Milliseconds(n int64) Duration {
	return Duration(n * int64(Millisecond))
}

// Seconds returns a Duration value representing n Seconds.
func (Clock) Seconds(n float64) Duration {
	return Duration(n * float64(Second))
}

// Minutes returns a Duration value representing n Minutes.
func (Clock) Minutes(n float64) Duration {
	return Duration(n * float64(Minute))
}

// Hours returns a Duration value representing n Hours.
func (Clock) Hours(n float64) Duration {
	return Duration(n * float64(Hour))
}

// Wrappers for `time` package functions

// Now returns the current local time.
func (Clock) Now() Time {
	return time.Now()
}

// ParseDuration parses a duration string. A duration string is a possibly
// signed sequence of decimal numbers, each with optional fraction and a unit
// suffix, such as "300ms", "-1.5h" or "2h45m". Valid time units are "ns",
// "us" (or "µs"), "ms", "s", "m", "h".
func (Clock) ParseDuration(s string) (Duration, error) {
	return time.ParseDuration(s)
}

// Since returns the time elapsed since t. It is shorthand for
// clock.Now().Sub(t).
func (Clock) Since(t Time) Duration {
	return time.Since(t)
}

// Until returns the duration until t. It is shorthand for t.Sub(clock.Now()).
func (Clock) Until(t Time) Duration {
	return time.Until(t)
}

// Sleep pauses the current goroutine for at least the duration d. A negative
// or zero duration causes Sleep to return immediately.
func (Clock) Sleep(d Duration) {
	time.Sleep(d)
}

// Ticker wraps [time.Ticker] to provide an interfaceable implementation.
type Ticker struct {
	*time.Ticker
}

// C returns the channel on which the ticks are delivered.
func (t *Ticker) C() <-chan Time {
	return t.Ticker.C
}

// NewTicker returns a new Ticker containing a channel that will send the
// current time on the channel after each tick. The period of the ticks is
// specified by the duration argument. The ticker will adjust the time
// interval or drop ticks to make up for slow receivers. The duration d must
// be greater than zero; if not, NewTicker will panic. Stop the ticker to
// release associated resources.
func (Clock) NewTicker(d Duration) *Ticker {
	return &Ticker{time.NewTicker(d)}
}

// Tick is a convenience wrapper for NewTicker providing access to the
// ticking channel only. While Tick is useful for clients that have no need
// to shut down the Ticker, be aware that without a way to shut it down the
// underlying Ticker cannot be recovered by the garbage collector; it
// "leaks". Unlike NewTicker, Tick will return nil if d <= 0.
func (Clock) Tick(d Duration) <-chan Time {
	return time.Tick(d)
}

// Timer wraps [time.Timer] to provide an interfaceable implementation.
type Timer struct {
	*time.Timer
}

// C returns the channel on which the ticks are delivered.
func (t *Timer) C() <-chan Time {
	return t.Timer.C
}

// NewTimer creates a new Timer that will send the current time on its
// channel after at least duration d.
func (Clock) NewTimer(d Duration) *Timer {
	return &Timer{time.NewTimer(d)}
}

// After waits for the duration to elapse and then sends the current time on
// the returned channel. It is equivalent to clock.NewTimer(d).C(). The
// underlying Timer is not recovered by the garbage collector until the timer
// fires. If efficiency is a concern, use clock.NewTimer instead and call
// Timer.Stop if the timer is no longer needed.
func (Clock) After(d Duration) <-chan Time {
	return time.After(d)
}

// AfterFunc waits for the duration to elapse and then calls f in its own
// goroutine. It returns a Timer that can be used to cancel the call using
// its Stop method.
func (Clock) AfterFunc(d Duration, f func()) *Timer {
	return &Timer{time.AfterFunc(d, f)}
}

// Wall clock (Location dependent) implementation

// See [time.Parse].
func (Clock) Parse(layout, value string) (Time, error) {
	return time.Parse(layout, value)
}

// See [time.ParseInLocation].
func (Clock) ParseInLocation(layout, value string, loc *Location) (Time, error) {
	return time.ParseInLocation(layout, value, loc)
}

// See [time.Date].
func (Clock) Date(year int, month Month, day, hour, min, sec, nsec int, loc *Location) Time {
	return time.Date(year, month, day, hour, min, sec, nsec, loc)
}

// See [time.Unix].
func (Clock) Unix(sec int64, nsec int64) Time {
	return time.Unix(sec, nsec)
}

// See [time.UnixMicro].
func (Clock) UnixMicro(usec int64) Time {
	return time.UnixMicro(usec)
}

// See [time.UnixMilli].
func (Clock) UnixMilli(msec int64) Time {
	return time.UnixMilli(msec)
}

// UnixNano is equivalent to clock.Unix(0, nsec).
func (Clock) UnixNano(nsec int64) Time {
	return time.Unix(0, nsec)
}

// Location functions

// See [time.FixedZone].
func (Clock) FixedZone(name string, offset int) *Location {
	return time.FixedZone(name, offset)
}

// See [time.LoadLocation].
func (Clock) LoadLocation(name string) (*Location, error) {
	return time.LoadLocation(name)
}

// See [time.LoadLocationFromTZData].
func (Clock) LoadLocationFromTZData(name string, data []byte) (*Location, error) {
	return time.LoadLocationFromTZData(name, data)
}
