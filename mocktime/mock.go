package mocktime

import (
	"time"

	"github.com/noodlebox/clock/realtime"
	"github.com/noodlebox/clock/relativetime"
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

// Timer is an alias for [relativetime.Timer] using the types [Time] and
// [Duration].
type Timer = relativetime.Timer[Time, Duration]

// Ticker is an alias for [relativetime.Ticker] using the types [Time] and
// [Duration].
type Ticker = relativetime.Ticker[Time, Duration]

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

// Wrap package-level functions around Clock methods

var clock Clock

func init() {
	clock = NewClockAt(realtime.Clock{}.Date(
		2009, November, 10, 23, 0, 0, 0, UTC,
	))
	clock.Start()
}

// Start starts or resumes the global Clock instance.
func Start() { clock.Start() }

// Stop pauses the global Clock instance.
func Stop() { clock.Stop() }

// Active returns true if the global Clock instance is currently running.
func Active() { clock.Active() }

// SetScale sets the scaling factor for the global Clock instance.
func SetScale(scale float64) { clock.SetScale(scale) }

// Scale returns the scaling factor of the global Clock instance.
func Scale() float64 { return clock.Scale() }

// Set changes the current time on the global Clock instance to now.
func Set(now Time) { clock.Set(now) }

// Step advances the current time on the global Clock instance by dt.
func Step(dt Duration) { clock.Step(dt) }

// NextAt returns the time of the next scheduled Timer or Ticker on the
// global Clock instance.
func NextAt() Time { return clock.NextAt() }

// Fastforward steps the global Clock instance forward to trigger timers
// until there are no timers left to trigger on it.
func Fastforward() { clock.Fastforward() }

// After waits for the duration to elapse and then sends the current time on
// the returned channel. It is equivalent to NewTimer(d).C(). The underlying
// Timer is not recovered by the garbage collector until the timer fires. If
// efficiency is a concern, use clock.NewTimer instead and call Timer.Stop if
// the timer is no longer needed.
func After(d Duration) <-chan Time { return clock.After(d) }

// Sleep pauses the current goroutine for at least the duration d. A negative
// or zero duration causes Sleep to return immediately.
func Sleep(d Duration) { clock.Sleep(d) }

// Tick is a convenience wrapper for NewTicker providing access to the
// ticking channel only. While Tick is useful for clients that have no need
// to shut down the Ticker, be aware that without a way to shut it down the
// underlying Ticker cannot be recovered by the garbage collector; it
// "leaks". Unlike NewTicker, Tick will return nil if d <= 0.
func Tick(d Duration) <-chan Time { return clock.Tick(d) }

// ParseDuration parses a duration string. A duration string is a possibly
// signed sequence of decimal numbers, each with optional fraction and a unit
// suffix, such as "300ms", "-1.5h" or "2h45m". Valid time units are "ns",
// "us" (or "µs"), "ms", "s", "m", "h".
func ParseDuration(s string) (Duration, error) { return clock.ParseDuration(s) }

// Since returns the time elapsed since t. It is shorthand for Now().Sub(t).
func Since(t Time) Duration { return clock.Since(t) }

// Until returns the duration until t. It is shorthand for t.Sub(Now()).
func Until(t Time) Duration { return clock.Until(t) }

// NewTicker returns a new Ticker containing a channel that will send the
// current time on the channel after each tick. The period of the ticks is
// specified by the duration argument. The ticker will adjust the time
// interval or drop ticks to make up for slow receivers. The duration d must
// be greater than zero; if not, NewTicker will panic. Stop the ticker to
// release associated resources.
func NewTicker(d Duration) *Ticker { return clock.NewTicker(d) }

// See [time.Date].
func Date(year int, month Month, day, hour, min, sec, nsec int, loc *Location) Time {
	return clock.Date(year, month, day, hour, min, sec, nsec, loc)
}

// Now returns the current time on the global Clock instance.
func Now() Time { return clock.Now() }

// See [time.Parse].
func Parse(layout, value string) (Time, error) { return clock.Parse(layout, value) }

// See [time.ParseInLocation].
func ParseInLocation(layout, value string, loc *Location) (Time, error) {
	return clock.ParseInLocation(layout, value, loc)
}

// See [time.Unix].
func Unix(sec int64, nsec int64) Time { return clock.Unix(sec, nsec) }

// See [time.UnixMicro].
func UnixMicro(usec int64) Time { return clock.UnixMicro(usec) }

// See [time.UnixMilli].
func UnixMilli(msec int64) Time { return clock.UnixMilli(msec) }

// AfterFunc waits for the duration to elapse and then calls f in its own
// goroutine. It returns a Timer that can be used to cancel the call using
// its Stop method.
func AfterFunc(d Duration, f func()) *Timer { return clock.AfterFunc(d, f) }

// NewTimer creates a new Timer that will send the current time on its
// channel after at least duration d.
func NewTimer(d Duration) *Timer { return clock.NewTimer(d) }

// See [time.FixedZone].
func FixedZone(name string, offset int) *Location { return clock.FixedZone(name, offset) }

// See [time.LoadLocation].
func LoadLocation(name string) (*Location, error) { return clock.LoadLocation(name) }

// See [time.LoadLocationFromTZData].
func LoadLocationFromTZData(name string, data []byte) (*Location, error) {
	return clock.LoadLocationFromTZData(name, data)
}
