package mocktime

import (
	"runtime"

	"github.com/noodlebox/clock/realtime"
	"github.com/noodlebox/clock/relativetime"
)

type baseClock struct {
	realtime.Clock
}

// Clock provides a drop in replacement for [realtime.Clock], but with
// additional methods to allow direct control over its behavior.
type Clock struct {
	*relativetime.Clock[Time, Duration, *realtime.Timer]
	baseClock // embed within a struct to ensure lower precedence
}

// NewClock returns a new Clock set to the current time.
func NewClock() Clock {
	rclock := realtime.NewClock()
	return Clock{
		relativetime.NewClock[Time, Duration, *realtime.Timer](rclock, rclock.Now(), 1.0),
		baseClock{rclock}, // zero value would work, but be explicit for clarity
	}
}

// NewClockAt returns a new Clock set to the the time, at.
func NewClockAt(at Time) Clock {
	rclock := realtime.NewClock()
	return Clock{
		relativetime.NewClock[Time, Duration, *realtime.Timer](rclock, at, 1.0),
		baseClock{rclock}, // zero value would work, but be explicit for clarity
	}
}

// Fastforward steps forward to trigger timers until there are no timers left
// to trigger.
func (c Clock) Fastforward() {
	active := c.Active()
	c.Stop()
	for when := c.NextAt(); !when.IsZero(); when = c.NextAt() {
		dt := c.Until(when)
		if dt < 0 {
			// Ensure we're never stepping backwards
			dt = 0
		}
		c.Step(dt)
		runtime.Gosched()
	}
	if active {
		c.Start()
	}
}
