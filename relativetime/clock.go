package relativetime

// The minimal API needed to serve as a reference clock
// For example, `realtime.Clock` satisfies:
//   `RClock[time.Time, time.Duration, *realtime.Timer]`
type RClock[T Time[T, D], D Duration, TM RTimer[T, D]] interface {
	Now() T
	Seconds(float64) D
	NewTimer(D) TM
}

// Minimal API needed for a reference Timer implementation
// For example, `*realtime.Timer` satisfies:
//   `RTimer[time.Time, time.Duration]`
type RTimer[T Time[T, D], D Duration] interface {
	C() <-chan T
	Reset(d D) bool
	Stop() bool
}

// Minimal API needed for a Time implementation
// For example, `time.Time` satisfies:
//   `Time[time.Time, time.Duration]`
type Time[T any, D Duration] interface {
	Add(D) T
	Sub(T) D
	After(T) bool
	Before(T) bool
	Equal(T) bool
}

// Minimal API needed for a Duration implementation
// For example, `time.Duration` satisfies this interface
type Duration interface {
	Seconds() float64
}

type Clock[T Time[T, D], D Duration, RT RTimer[T, D]] struct {
	ref       RClock[T, D, RT]
	scale     float64
	active    bool
	now, rNow T // last sync point

	sched queue[T, D] // Upcoming events, in local time
	waker RTimer[T, D]
	sleep chan struct{}
}

func NewClock[T Time[T, D], D Duration, RT RTimer[T, D]](ref RClock[T, D, RT], at T, scale float64) (c *Clock[T, D, RT]) {
	c = &Clock[T, D, RT]{
		ref:    ref,
		active: false,
		scale:  scale,
		now:    ref.Now(),
		rNow:   at,
	}
	return
}

// Syncing with the reference clock is done lazily. This method updates the
// sync points based on difference between a new reference time and the last
// sync point. Fields that would affect how the reference is tracked should
// not change between resyncs. This should be called to ensure sync points
// are not stale before any change to one of these fields.
func (c *Clock[T, D, RT]) advanceRef(now T) {
	then := c.rNow

	// Update ref sync time
	c.rNow = now

	// No local change if stopped, scale is zero, or ref clock hasn't changed
	if !c.active || c.scale == 0.0 || now.Equal(then) {
		return
	}
	dt := now.Sub(then)
	if c.scale != 1.0 {
		// Apply scale via conversion to float64 in seconds
		dt = c.ref.Seconds(dt.Seconds() * c.scale)
	}
	// We're at now now.
	c.now = c.now.Add(dt)
}

func (c *Clock[T, D, RT]) stopWaker() {
	if c.waker == nil {
		return
	}

	// Interrupt waker routine if still running
	select {
	case _, ok := <-c.sleep:
		if !ok {
			// Already ended (c.sleep closed)
			return
		}
		// Did not consume from timer channel
		if !c.waker.Stop() {
			// Clear channel if timer has triggered but waker routine hadn't
			// consumed it before being interrupted
			<-c.waker.C()
		}
	}
}

func (c *Clock[T, D, RT]) resetWaker() {
	if !c.active || c.scale == 0.0 {
		return
	}

	next := c.sched.peek()
	if next == nil {
		// Nothing current scheduled
		return
	}
	// Duration on reference clock until next timer should trigger
	dt := c.ref.Seconds(next.when.Sub(c.now).Seconds() / c.scale)

	if c.waker == nil {
		c.waker = c.ref.NewTimer(dt)
	} else {
		c.waker.Reset(dt)
	}

	sleep := make(chan struct{})
	go func() {
		select {
		case t := <-c.waker.C():
			// Advance clock to t and process any timers that should trigger
			c.wake(t)
		case sleep <- struct{}{}:
			// Interrupted
		}
		// Signal that we've finished
		close(sleep)
	}()
	c.sleep = sleep
}

// Check schedule for pending events that should trigger now.
func (c *Clock[T, D, RT]) checkSchedule() {
	c.stopWaker()

	for t := c.sched.peek(); t != nil && t.when.Before(c.now); t = c.sched.peek() {
		if t.period.Seconds() <= 0 {
			c.sched.unschedule(t)
		} else {
			t.when = c.now.Add(t.period)
			c.sched.reschedule(t)
		}
		t.f(c.now)
	}

	c.resetWaker()
}

// This method is called whenever a reference timer triggers.
func (c *Clock[T, D, RT]) wake(now T) {
	// Don't step backwards if this callback ends up delayed
	if now.After(c.rNow) {
		c.advanceRef(now)
	}

	c.checkSchedule()
}

// Start() begins tracking the reference clock, if not already running.
// It is fine to call Start() on a clock that is already running.
func (c *Clock[T, D, RT]) Start() {
	// Sync up first
	c.advanceRef(c.ref.Now())

	c.active = true
}

// Stop() stops tracking the reference clock, if currently running.
// It is fine to call Stop() on a clock that is not running.
func (c *Clock[T, D, RT]) Stop() {
	// Sync up first
	c.advanceRef(c.ref.Now())

	c.active = false
}

func (c *Clock[T, D, RT]) Active() bool {
	return c.active
}

func (c *Clock[T, D, RT]) SetScale(scale float64) {
	// Sync up first
	c.advanceRef(c.ref.Now())

	c.scale = scale
}

func (c *Clock[T, D, RT]) Scale() float64 {
	return c.scale
}

// Set the local sync point with the current reference time to `now`
// If any timers are active, a value of `now` earlier than the previous
// setting may lead to undefined behavior.
func (c *Clock[T, D, RT]) Set(now T) {
	// Reset sync point to given time
	c.now, c.rNow = now, c.ref.Now()

	// Check whether we're due for any scheduled events
	c.checkSchedule()
}

// Advance the local time forward by `dt`.
// If any timers are active, a negative value for dt may lead to undefined
// behavior.
func (c *Clock[T, D, RT]) Step(dt D) {
	// Sync up first
	c.advanceRef(c.ref.Now())

	c.now = c.now.Add(dt)

	// Check whether we're due for any scheduled events
	c.checkSchedule()
}

func (c *Clock[T, D, RT]) Now() T {
	// Sync up
	c.advanceRef(c.ref.Now())

	return c.now
}

func (c *Clock[T, D, RT]) Since(t T) D {
	return c.Now().Sub(t)
}

func (c *Clock[T, D, RT]) Until(t T) D {
	return t.Sub(c.Now())
}

func (c *Clock[T, D, RT]) Sleep(d D) {
	// TODO
}

type Ticker[T Time[T, D], D Duration] struct {
	c chan T
}

func (t *Ticker[T, D]) C() <-chan T {
	return t.c
}

func (c *Clock[T, D, RT]) NewTicker(d D) *Ticker[T, D] {
	// TODO
	return nil
}

func (c *Clock[T, D, RT]) Tick(d D) <-chan T {
	return c.NewTicker(d).c
}

type Timer[T Time[T, D], D Duration] struct {
	c chan T
}

func (t *Timer[T, D]) C() <-chan T {
	return t.c
}

func (c *Clock[T, D, RT]) NewTimer(d D) *Timer[T, D] {
	// TODO
	return nil
}

func (c *Clock[T, D, RT]) After(d D) <-chan T {
	return c.NewTimer(d).c
}

func (c *Clock[T, D, RT]) AfterFunc(d D, f func()) *Timer[T, D] {
	// TODO
	return nil
}
