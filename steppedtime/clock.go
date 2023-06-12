package steppedtime

import (
	"sync"
)

// Clock represents a simulation clock that only advances when explicitly
// stepped. Its methods are thread-safe. The zero-value of a Clock is
// perfectly valid.
type Clock struct {
	now   Time
	queue queue

	mu sync.Mutex
}

// NewClock returns a new Clock.
func NewClock() *Clock {
	return &Clock{}
}

func (c *Clock) lock()   { c.mu.Lock() }
func (c *Clock) unlock() { c.mu.Unlock() }

// Set sets the current time to now. If any timers are active, a value of now
// earlier than the previous setting may lead to undefined behavior.
func (c *Clock) Set(now Time) {
	c.lock()
	c.now = now

	// Check whether we're due for any scheduled events
	c.checkSchedule()
	c.unlock()
}

// Step advances the current time by dt. If any timers are active, a negative
// value for dt may lead to undefined behavior.
func (c *Clock) Step(dt Duration) {
	c.lock()
	c.now = c.now.Add(dt)

	// Check whether we're due for any scheduled events
	c.checkSchedule()
	c.unlock()
}

// Now returns the current time.
func (c *Clock) Now() (now Time) {
	c.lock()
	now = c.now
	c.unlock()
	return
}

// Since returns the time elapsed since t. It is shorthand for
// clock.Now().Sub(t).
func (c *Clock) Since(t Time) Duration {
	return c.Now().Sub(t)
}

// Until returns the duration until t. It is shorthand for t.Sub(clock.Now()).
func (c *Clock) Until(t Time) Duration {
	return t.Sub(c.Now())
}

// Sleep pauses the current goroutine for at least the duration d. A negative
// or zero duration causes Sleep to return immediately.
func (c *Clock) Sleep(d Duration) {
	if d <= 0 {
		return
	}

	ch := make(chan struct{})
	c.lock()
	c.schedule(&timer{
		f:    func(Time) { close(ch) },
		when: c.now.Add(d),
	})
	c.unlock()
	<-ch
}

// A Ticker provides a channel that delivers “ticks” of a clock at
// intervals.
type Ticker struct {
	c <-chan Time
	t *timer
	s *Clock
}

// C returns the channel on which the ticks are delivered.
func (t *Ticker) C() <-chan Time {
	return t.c
}

// Reset stops a ticker and resets its period to the specified duration. The
// next tick will arrive after the new period elapses. The duration d must be
// greater than zero; if not, Reset will panic.
func (t *Ticker) Reset(d Duration) {
	if d <= 0 {
		panic("non-positive interval for steppedtime.Ticker.Reset")
	}
	if t.t == nil {
		panic("Reset called on uninitialized steppedtime.Ticker")
	}

	t.s.lock()
	t.t.when = t.s.now.Add(d)
	t.t.period = d
	t.s.reschedule(t.t)
	t.s.unlock()
}

// Stop turns off a ticker. After Stop, no more ticks will be sent. Stop does
// not close the channel, to prevent a concurrent goroutine reading from the
// channel from seeing an erroneous "tick".
func (t *Ticker) Stop() {
	if t.t == nil {
		panic("Stop called on uninitialized steppedtime.Ticker")
	}

	t.s.lock()
	t.s.unschedule(t.t)
	t.s.unlock()
}

// NewTicker returns a new Ticker containing a channel that will send the
// current time on the channel after each tick. The period of the ticks is
// specified by the duration argument. The ticker will adjust the time
// interval or drop ticks to make up for slow receivers. The duration d must
// be greater than zero; if not, NewTicker will panic. Stop the ticker to
// release associated resources.
func (c *Clock) NewTicker(d Duration) *Ticker {
	if d <= 0 {
		panic("non-positive interval for steppedtime.Clock.NewTicker")
	}

	ch := make(chan Time, 1)
	c.lock()
	tm := &timer{
		f: func(when Time) {
			select {
			case ch <- when:
			default:
			}
		},
		when:   c.now.Add(d),
		period: d,
	}
	c.schedule(tm)
	c.unlock()
	return &Ticker{ch, tm, c}
}

// Tick is a convenience wrapper for NewTicker providing access to the
// ticking channel only. While Tick is useful for clients that have no need
// to shut down the Ticker, be aware that without a way to shut it down the
// underlying Ticker cannot be recovered by the garbage collector; it
// "leaks". Unlike NewTicker, Tick will return nil if d <= 0.
func (c *Clock) Tick(d Duration) <-chan Time {
	if d <= 0 {
		return nil
	}

	return c.NewTicker(d).c
}

// The Timer type represents a single event. When the Timer expires, the
// current time will be sent on the channel returned by C(), unless the Timer
// was created by AfterFunc. A Timer must be created with NewTimer or
// AfterFunc.
type Timer struct {
	c <-chan Time
	t *timer
	s *Clock
}

// C returns the channel on which the ticks are delivered.
func (t *Timer) C() <-chan Time {
	return t.c
}

// Reset changes the timer to expire after duration d. It returns true if the
// timer had been active, false if the timer had expired or been stopped.
func (t *Timer) Reset(d Duration) (active bool) {
	if t.t == nil {
		panic("Reset called on uninitialized steppedtime.Timer")
	}

	t.s.lock()
	t.t.when = t.s.now.Add(d)
	active = (t.t.index != -1)
	t.s.reschedule(t.t)
	t.s.unlock()
	return
}

// Stop prevents the Timer from firing. It returns true if the call stops the
// timer, false if the timer has already expired or been stopped. Stop does
// not close the channel, to prevent a read from the channel succeeding
// incorrectly.
func (t *Timer) Stop() (active bool) {
	if t.t == nil {
		panic("Stop called on uninitialized steppedtime.Timer")
	}

	t.s.lock()
	active = (t.t.index != -1)
	t.s.unschedule(t.t)
	t.s.unlock()
	return
}

// NewTimer creates a new Timer that will send the current time on its
// channel after at least duration d.
func (c *Clock) NewTimer(d Duration) *Timer {
	ch := make(chan Time, 1)
	c.lock()
	tm := &timer{
		f: func(when Time) {
			select {
			case ch <- when:
			default:
			}
		},
		when: c.now.Add(d),
	}
	c.schedule(tm)
	c.unlock()
	return &Timer{ch, tm, c}
}

// After waits for the duration to elapse and then sends the current time on
// the returned channel. It is equivalent to clock.NewTimer(d).C(). The
// underlying Timer is not recovered by the garbage collector until the timer
// fires. If efficiency is a concern, use clock.NewTimer instead and call
// Timer.Stop if the timer is no longer needed.
func (c *Clock) After(d Duration) <-chan Time {
	return c.NewTimer(d).c
}

// AfterFunc waits for the duration to elapse and then calls f in its own
// goroutine. It returns a Timer that can be used to cancel the call using
// its Stop method.
func (c *Clock) AfterFunc(d Duration, f func()) *Timer {
	c.lock()
	tm := &timer{
		f:    func(Time) { go f() },
		when: c.now.Add(d),
	}
	c.schedule(tm)
	c.unlock()
	return &Timer{t: tm, s: c}
}
