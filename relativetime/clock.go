package relativetime

import (
	"container/heap"
)

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

// A clock implementation that tracks a reference clock with a configurable
// scaling factor.
// NOTE: composition with the reference clock would be such a nice feature
// here, to inherit all the methods of the reference clock. Maybe in a future
// version of Go... See `mocktime` package for an example of using embedding
// with instantiated generic types for a drop in replacement for a reference
// clock.
type Clock[T Time[T, D], D Duration, RT RTimer[T, D]] struct {
	ref       RClock[T, D, RT]
	scale     float64
	active    bool
	now, rNow T // last sync point

	queue queue[T, D]     // Upcoming events, in local time
	waker RTimer[T, D]    // Interface used here for a default value of nil
	sleep <-chan struct{} // Interrupts the waker, or signals its completion
}

func NewClock[T Time[T, D], D Duration, RT RTimer[T, D]](ref RClock[T, D, RT], at T, scale float64) (c *Clock[T, D, RT]) {
	c = &Clock[T, D, RT]{
		ref:    ref,
		active: false,
		scale:  scale,
		now:    at,
		rNow:   ref.Now(),
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

	next := c.queue.peek()
	if next == nil {
		// Nothing currently scheduled
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

	for t := c.queue.peek(); t != nil && t.when.Before(c.now); t = c.queue.peek() {
		if t.period.Seconds() <= 0 {
			c.unschedule(t)
		} else {
			t.when = c.now.Add(t.period)
			c.reschedule(t)
		}
		t.f(c.now)
	}

	c.resetWaker()
}

// stop the waker when modifying the queue then reset it afterwards

func (c *Clock[T, D, RT]) schedule(t *timer[T, D]) {
	heap.Push(&c.queue, t)
}

func (c *Clock[T, D, RT]) unschedule(t *timer[T, D]) {
	if t.index == -1 {
		return
	}
	heap.Remove(&c.queue, t.index)
}

func (c *Clock[T, D, RT]) reschedule(t *timer[T, D]) {
	if t.index == -1 {
		c.schedule(t)
		return
	}
	heap.Fix(&c.queue, t.index)
}

// This method is called whenever a reference timer triggers.
// This is the other way for the reference sync point to advance, aside from
// calling Now() on the reference timer.
func (c *Clock[T, D, RT]) wake(now T) {
	// Don't step backwards in case this callback ends up delayed
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
	c.stopWaker()
	c.resetWaker()
}

// Stop() stops tracking the reference clock, if currently running.
// It is fine to call Stop() on a clock that is not running.
func (c *Clock[T, D, RT]) Stop() {
	// Sync up first
	c.advanceRef(c.ref.Now())

	c.active = false
	c.stopWaker()
}

func (c *Clock[T, D, RT]) Active() bool {
	return c.active
}

func (c *Clock[T, D, RT]) SetScale(scale float64) {
	// Sync up first
	c.advanceRef(c.ref.Now())

	c.scale = scale
	c.stopWaker()
	c.resetWaker()
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

// Use reference clock to implement Seconds method, to allow a relative clock
// to satisfy the reference clock interface itself.
func (c *Clock[T, D, RT]) Seconds(n float64) D {
	return c.ref.Seconds(n)
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
	if d.Seconds() <= 0 {
		return
	}

	// Sync up
	c.advanceRef(c.ref.Now())

	ch := make(chan struct{})
	c.schedule(&timer[T, D]{
		f:    func(T) { close(ch) },
		when: c.now.Add(d),
	})
	<-ch
}

type scheduler[T Time[T, D], D Duration] interface {
	schedule(t *timer[T, D])
	unschedule(t *timer[T, D])
	reschedule(t *timer[T, D])
	stopWaker()
	resetWaker()
	Now() T
}

type Ticker[T Time[T, D], D Duration] struct {
	c <-chan T
	t *timer[T, D]
	s scheduler[T, D]
}

func (t *Ticker[T, D]) C() <-chan T {
	return t.c
}

func (t *Ticker[T, D]) Reset(d D) {
	if d.Seconds() <= 0 {
		panic("non-positive interval for relativetime.Ticker.Reset")
	}
	if t.t == nil {
		panic("Reset called on uninitialized relativetime.Ticker")
	}

	t.s.stopWaker()
	t.t.when = t.s.Now().Add(d)
	t.t.period = d
	t.s.reschedule(t.t)
	t.s.resetWaker()
}

func (t *Ticker[T, D]) Stop() {
	if t.t == nil {
		panic("Stop called on uninitialized relativetime.Ticker")
	}

	t.s.stopWaker()
	t.s.unschedule(t.t)
	t.s.resetWaker()
}

func (c *Clock[T, D, RT]) NewTicker(d D) *Ticker[T, D] {
	if d.Seconds() <= 0 {
		panic("non-positive interval for relativetime.Clock.NewTicker")
	}

	// Sync up
	c.advanceRef(c.ref.Now())

	ch := make(chan T, 1)
	tm := &timer[T, D]{
		f: func(when T) {
			select {
			case ch <- when:
			default:
			}
		},
		when:   c.now.Add(d),
		period: d,
	}
	c.stopWaker()
	c.schedule(tm)
	c.resetWaker()
	return &Ticker[T, D]{ch, tm, c}
}

func (c *Clock[T, D, RT]) Tick(d D) <-chan T {
	if d.Seconds() <= 0 {
		return nil
	}

	return c.NewTicker(d).c
}

type Timer[T Time[T, D], D Duration] struct {
	c <-chan T
	t *timer[T, D]
	s scheduler[T, D]
}

func (t *Timer[T, D]) C() <-chan T {
	return t.c
}

func (t *Timer[T, D]) Reset(d D) (active bool) {
	if t.t == nil {
		panic("Reset called on uninitialized relativetime.Timer")
	}

	t.s.stopWaker()
	t.t.when = t.s.Now().Add(d)
	active = (t.t.index != -1)
	t.s.reschedule(t.t)
	t.s.resetWaker()
	return
}

func (t *Timer[T, D]) Stop() (active bool) {
	if t.t == nil {
		panic("Stop called on uninitialized relativetime.Timer")
	}

	t.s.stopWaker()
	active = (t.t.index != -1)
	t.s.unschedule(t.t)
	t.s.resetWaker()
	return
}

func (c *Clock[T, D, RT]) NewTimer(d D) *Timer[T, D] {
	// Sync up
	c.advanceRef(c.ref.Now())

	ch := make(chan T, 1)
	tm := &timer[T, D]{
		f: func(when T) {
			select {
			case ch <- when:
			default:
			}
		},
		when: c.now.Add(d),
	}
	c.stopWaker()
	c.schedule(tm)
	c.resetWaker()
	return &Timer[T, D]{ch, tm, c}
}

func (c *Clock[T, D, RT]) After(d D) <-chan T {
	return c.NewTimer(d).c
}

func (c *Clock[T, D, RT]) AfterFunc(d D, f func()) *Timer[T, D] {
	// Sync up
	c.advanceRef(c.ref.Now())

	tm := &timer[T, D]{
		f:    func(T) { go f() },
		when: c.now.Add(d),
	}
	c.stopWaker()
	c.schedule(tm)
	c.resetWaker()
	return &Timer[T, D]{t: tm, s: c}
}
