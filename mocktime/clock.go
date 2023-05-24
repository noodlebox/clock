package mocktime

import (
	"github.com/noodlebox/clock/realtime"
	"github.com/noodlebox/clock/relativetime"
)

type baseClock struct {
	realtime.Clock
}

// Inherits methods from relative clock, falling back to real clock for the rest
type Clock struct {
	*relativetime.Clock[Time, Duration, *realtime.Timer]
	baseClock // embed within a struct to ensure lower precedence
}

func NewClock() Clock {
	rclock := realtime.NewClock()
	return Clock{
		relativetime.NewClock[Time, Duration, *realtime.Timer](rclock, rclock.Now(), 1.0),
		baseClock{rclock}, // zero value would work, but be explicit for clarity
	}
}
func NewClockAt(at Time) Clock {
	rclock := realtime.NewClock()
	return Clock{
		relativetime.NewClock[Time, Duration, *realtime.Timer](rclock, at, 1.0),
		baseClock{rclock}, // zero value would work, but be explicit for clarity
	}
}
