package relativetime

import (
	"container/heap"
	"sync"
)

// RClock is a generic interface for the minimal API needed to serve as a
// reference clock.
type RClock[T Time[T, D], D Duration, TM RTimer[D]] interface {
	Now() T
	Seconds(float64) D
	AfterFunc(D, func()) TM
}

// RTimer is a generic interface for the minimal API needed for a reference
// Timer implementation.
type RTimer[D Duration] interface {
	Reset(d D) bool
	Stop() bool
}

// Time is a generic interface for the minimal API needed for a Time
// implementation.
type Time[T any, D Duration] interface {
	Add(D) T
	Sub(T) D
	After(T) bool
	Before(T) bool
	Equal(T) bool
	IsZero() bool
}

// Duration is an interface for the minimal API needed for a Duration
// implementation.
type Duration interface {
	Seconds() float64
}

type waker[T Time[T, D], D Duration, RT RTimer[D]] struct {
	queue  queue[T, D] // Upcoming events, in local time
	waker  RTimer[D]   // Interface used here for a default value of nil
	wakeAt T           // Local time of next scheduled waking
	mu     sync.RWMutex
	*Clock[T, D, RT]
}

const nwakers = 4

// Clock is a clock that tracks a reference clock with a configurable scaling
// factor.
//
// NOTE: composition with the reference clock would be such a nice feature
// here, to inherit all the methods of the reference clock. Maybe in a future
// version of Go... See [github.com/noodlebox/clock/mocktime] package for an
// example of using embedding with instantiated generic types for a drop in
// replacement for a reference clock.
type Clock[T Time[T, D], D Duration, RT RTimer[D]] struct {
	ref       RClock[T, D, RT]
	scale     float64
	active    bool
	now, rNow T // last sync point

	mu sync.RWMutex // Protects now, rNow, scale, and active

	waker  chan *waker[T, D, RT]
	wakers [nwakers]waker[T, D, RT]
}

// NewClock returns a new Clock set to at synchronized to the current time on
// ref with a scale factor of scale.
func NewClock[T Time[T, D], D Duration, RT RTimer[D]](ref RClock[T, D, RT], at T, scale float64) (c *Clock[T, D, RT]) {
	c = &Clock[T, D, RT]{
		ref:    ref,
		active: false,
		scale:  scale,
		now:    at,
		rNow:   ref.Now(),
		waker:  make(chan *waker[T, D, RT], nwakers),
	}
	for i, _ := range c.wakers {
		c.wakers[i].Clock = c
		c.waker <- &c.wakers[i]
	}
	return
}

func (c *Clock[T, D, RT]) lock()   { c.mu.Lock() }
func (c *Clock[T, D, RT]) unlock() { c.mu.Unlock() }

func (c *Clock[T, D, RT]) rlock()   { c.mu.RLock() }
func (c *Clock[T, D, RT]) runlock() { c.mu.RUnlock() }

// Syncing with the reference clock is done lazily. This method updates the
// sync points based on difference between a new reference time and the last
// sync point. Fields that would affect how the reference is tracked should
// not change between resyncs. This should be called to ensure sync points
// are not stale before any change to one of these fields.
func (c *Clock[T, D, RT]) advanceRef(now T) {
	c.now = c.nowLocal(now)

	// Update ref sync time
	c.rNow = now
}

func (c *Clock[T, D, RT]) nowLocal(now T) T {
	then := c.rNow

	// No local change if stopped, scale is zero, or ref clock hasn't changed
	if !c.active || c.scale == 0.0 || now.Equal(then) {
		return c.now
	}
	dt := now.Sub(then)
	if c.scale != 1.0 {
		// Apply scale via conversion to float64 in seconds
		dt = c.ref.Seconds(dt.Seconds() * c.scale)
	}
	// We're at now now.
	return c.now.Add(dt)
}

// Should only be called after a proper sync, so c.now is valid.
func (c *Clock[T, D, RT]) resetWakers(dirty bool) {
	now := c.now
	for i, _ := range c.wakers {
		w := &c.wakers[i]
		go func() {
			w.lock()
			w.reset(now, dirty)
			w.unlock()
		}()
	}
}

// Should only be called after a proper sync, so c.now is valid.
func (c *Clock[T, D, RT]) checkWakers() {
	now := c.now
	for i, _ := range c.wakers {
		w := &c.wakers[i]
		go func() {
			w.lock()
			w.checkSchedule(now)
			w.unlock()
		}()
	}
}

func (w *waker[T, D, RT]) lock()   { w.mu.Lock() }
func (w *waker[T, D, RT]) unlock() { w.mu.Unlock() }

func (w *waker[T, D, RT]) stop() {
	if w.waker == nil {
		return
	}
	w.waker.Stop()
}

func (w *waker[T, D, RT]) reset(now T, dirty bool) {
	w.rlock()
	active, scale := w.active, w.scale
	w.runlock()
	if !active || scale == 0.0 {
		w.stop()
		return
	}

	next := w.queue.peek()
	if next == nil {
		// Nothing currently scheduled
		w.stop()
		return
	}

	if !dirty && w.waker != nil && next.when.Equal(w.wakeAt) {
		// Waker already set to the correct time, let it be
		return
	}

	w.wakeAt = next.when

	// Duration on reference clock until next timer should trigger
	dt := w.ref.Seconds(next.when.Sub(now).Seconds() / scale)

	if w.waker == nil {
		w.waker = w.ref.AfterFunc(dt, w.wake)
	} else {
		w.waker.Reset(dt)
	}
}

// Check schedule for pending events that should trigger now.
func (w *waker[T, D, RT]) checkSchedule(now T) {
	for t := w.queue.peek(); t != nil && !t.when.After(now); t = w.queue.peek() {
		if t.period.Seconds() <= 0 {
			w.unschedule(t)
		} else {
			t.when = now.Add(t.period)
			w.reschedule(t)
		}
		t.f(now)
	}

	w.reset(now, false)
}

// stop the waker when modifying the queue then reset it afterwards

func (w *waker[T, D, RT]) schedule(t *timer[T, D]) {
	heap.Push(&w.queue, t)
}

func (w *waker[T, D, RT]) unschedule(t *timer[T, D]) {
	if t.index == -1 {
		return
	}
	heap.Remove(&w.queue, t.index)
}

func (w *waker[T, D, RT]) reschedule(t *timer[T, D]) {
	if t.index == -1 {
		w.schedule(t)
		return
	}
	heap.Fix(&w.queue, t.index)
}

// This method is called whenever a reference timer triggers.
// This is the other way for the reference sync point to advance, aside from
// calling Now() on the reference timer.
func (w *waker[T, D, RT]) wake() {
	now := w.Now()
	w.lock()
	w.checkSchedule(now)
	w.unlock()
}

// Start begins tracking the reference clock, if not already running. It is
// fine to call Start() on a clock that is already running.
func (c *Clock[T, D, RT]) Start() {
	c.lock()
	// Sync up first
	c.advanceRef(c.ref.Now())

	dirty := !c.active // Did the setting change?
	c.active = true

	c.resetWakers(dirty)
	c.unlock()
}

// Stop stops tracking the reference clock, if currently running. It is fine
// to call Stop() on a clock that is not running.
func (c *Clock[T, D, RT]) Stop() {
	c.lock()
	// Sync up first
	c.advanceRef(c.ref.Now())

	dirty := c.active // Did the setting change?
	c.active = false

	c.resetWakers(dirty)
	c.unlock()
}

// Active returns true if currently tracking the reference clock.
func (c *Clock[T, D, RT]) Active() (active bool) {
	c.rlock()
	active = c.active
	c.runlock()
	return
}

// SetScale sets the scaling factor for tracking the reference clock.
func (c *Clock[T, D, RT]) SetScale(scale float64) {
	c.lock()
	// Sync up first
	c.advanceRef(c.ref.Now())

	dirty := c.scale != scale // Did the setting change?
	c.scale = scale

	c.resetWakers(dirty)
	c.unlock()
}

// Scale returns the scaling factor for tracking the reference clock.
func (c *Clock[T, D, RT]) Scale() (scale float64) {
	c.rlock()
	scale = c.scale
	c.runlock()
	return
}

// Set sets the local sync point with the current reference time to now. If
// any timers are active, a value of now earlier than the previous setting
// may lead to undefined behavior.
func (c *Clock[T, D, RT]) Set(now T) {
	c.lock()
	// Reset sync point to given time
	c.now, c.rNow = now, c.ref.Now()

	// Check whether we're due for any scheduled events
	c.checkWakers()
	c.unlock()
}

// Step advances the local time forward by dt. If any timers are active, a
// negative value for dt may lead to undefined behavior.
func (c *Clock[T, D, RT]) Step(dt D) {
	c.lock()
	// Sync up first
	c.advanceRef(c.ref.Now())

	c.now = c.now.Add(dt)

	// Check whether we're due for any scheduled events
	c.checkWakers()
	c.unlock()
}

// NextAt returns the time at which the next scheduled timer should trigger.
// If no timers are scheduled, returns a zero value.
func (c *Clock[T, D, RT]) NextAt() (when T) {
	// TODO: could check wakers concurrently, but this is enough for now
	for i, _ := range c.wakers {
		w := &c.wakers[i]
		w.rlock()
		next := w.queue.peek()
		if next != nil && when.IsZero() || when.After(next.when) {
			when = next.when
		}
		w.runlock()
	}
	return
}

// Seconds returns a Duration value representing n Seconds. This is provided
// to allow a relative clock itself to satisfy the reference clock interface.
func (c *Clock[T, D, RT]) Seconds(n float64) D {
	return c.ref.Seconds(n)
}

// Now returns the current time.
func (c *Clock[T, D, RT]) Now() (now T) {
	now = c.ref.Now()
	c.rlock()
	now = c.nowLocal(now)
	c.runlock()
	return
}

// Since returns the time elapsed since t. It is shorthand for
// clock.Now().Sub(t).
func (c *Clock[T, D, RT]) Since(t T) D {
	return c.Now().Sub(t)
}

// Until returns the duration until t. It is shorthand for t.Sub(clock.Now()).
func (c *Clock[T, D, RT]) Until(t T) D {
	return t.Sub(c.Now())
}

// Sleep pauses the current goroutine for at least the duration d. A negative
// or zero duration causes Sleep to return immediately.
func (c *Clock[T, D, RT]) Sleep(d D) {
	if d.Seconds() <= 0 {
		return
	}

	now := c.Now()
	ch := make(chan struct{})
	tm := &timer[T, D]{
		f:    func(T) { close(ch) },
		when: now.Add(d),
	}
	w := <-c.waker
	w.lock()
	w.schedule(tm)
	if tm.index == 0 {
		w.reset(now, false)
	}
	w.unlock()
	c.waker <- w
	<-ch
}

type scheduler[T Time[T, D], D Duration] interface {
	schedule(t *timer[T, D])
	unschedule(t *timer[T, D])
	reschedule(t *timer[T, D])
	reset(now T, dirty bool)
	lock()
	unlock()
	Now() T
}

// A Ticker provides a channel that delivers “ticks” of a clock at
// intervals.
type Ticker[T Time[T, D], D Duration] struct {
	c <-chan T
	t *timer[T, D]
	s scheduler[T, D]
}

// C returns the channel on which the ticks are delivered.
func (t *Ticker[T, D]) C() <-chan T {
	return t.c
}

// Reset stops a ticker and resets its period to the specified duration. The
// next tick will arrive after the new period elapses. The duration d must be
// greater than zero; if not, Reset will panic.
func (t *Ticker[T, D]) Reset(d D) {
	if d.Seconds() <= 0 {
		panic("non-positive interval for relativetime.Ticker.Reset")
	}
	if t.t == nil {
		panic("Reset called on uninitialized relativetime.Ticker")
	}

	now := t.s.Now()
	t.s.lock()
	t.t.when = now.Add(d)
	t.t.period = d
	isNext := t.t.index == 0
	t.s.reschedule(t.t)
	if isNext || t.t.index == 0 {
		t.s.reset(now, false)
	}
	t.s.unlock()
}

// Stop turns off a ticker. After Stop, no more ticks will be sent. Stop does
// not close the channel, to prevent a concurrent goroutine reading from the
// channel from seeing an erroneous "tick".
func (t *Ticker[T, D]) Stop() {
	if t.t == nil {
		panic("Stop called on uninitialized relativetime.Ticker")
	}

	now := t.s.Now()
	t.s.lock()
	isNext := t.t.index == 0
	t.s.unschedule(t.t)
	if isNext {
		t.s.reset(now, false)
	}
	t.s.unlock()
}

// NewTicker returns a new Ticker containing a channel that will send the
// current time on the channel after each tick. The period of the ticks is
// specified by the duration argument. The ticker will adjust the time
// interval or drop ticks to make up for slow receivers. The duration d must
// be greater than zero; if not, NewTicker will panic. Stop the ticker to
// release associated resources.
func (c *Clock[T, D, RT]) NewTicker(d D) *Ticker[T, D] {
	if d.Seconds() <= 0 {
		panic("non-positive interval for relativetime.Clock.NewTicker")
	}

	now := c.Now()
	ch := make(chan T, 1)
	tm := &timer[T, D]{
		f: func(when T) {
			select {
			case ch <- when:
			default:
			}
		},
		when:   now.Add(d),
		period: d,
	}
	w := <-c.waker
	w.lock()
	w.schedule(tm)
	if tm.index == 0 {
		w.reset(now, false)
	}
	w.unlock()
	c.waker <- w
	return &Ticker[T, D]{ch, tm, w}
}

// Tick is a convenience wrapper for NewTicker providing access to the
// ticking channel only. While Tick is useful for clients that have no need
// to shut down the Ticker, be aware that without a way to shut it down the
// underlying Ticker cannot be recovered by the garbage collector; it
// "leaks". Unlike NewTicker, Tick will return nil if d <= 0.
func (c *Clock[T, D, RT]) Tick(d D) <-chan T {
	if d.Seconds() <= 0 {
		return nil
	}

	return c.NewTicker(d).c
}

// The Timer type represents a single event. When the Timer expires, the
// current time will be sent on the channel returned by C(), unless the Timer
// was created by AfterFunc. A Timer must be created with NewTimer or
// AfterFunc.
type Timer[T Time[T, D], D Duration] struct {
	c <-chan T
	t *timer[T, D]
	s scheduler[T, D]
}

// C returns the channel on which the ticks are delivered.
func (t *Timer[T, D]) C() <-chan T {
	return t.c
}

// Reset changes the timer to expire after duration d. It returns true if the
// timer had been active, false if the timer had expired or been stopped.
func (t *Timer[T, D]) Reset(d D) (active bool) {
	if t.t == nil {
		panic("Reset called on uninitialized relativetime.Timer")
	}

	now := t.s.Now()
	t.s.lock()

	active = (t.t.index != -1)

	t.t.when = now.Add(d)
	isNext := t.t.index == 0
	t.s.reschedule(t.t)
	if isNext || t.t.index == 0 {
		t.s.reset(now, false)
	}
	t.s.unlock()

	return
}

// Stop prevents the Timer from firing. It returns true if the call stops the
// timer, false if the timer has already expired or been stopped. Stop does
// not close the channel, to prevent a read from the channel succeeding
// incorrectly.
func (t *Timer[T, D]) Stop() (active bool) {
	if t.t == nil {
		panic("Stop called on uninitialized relativetime.Timer")
	}

	now := t.s.Now()
	t.s.lock()

	active = (t.t.index != -1)

	isNext := t.t.index == 0
	t.s.unschedule(t.t)
	if isNext {
		t.s.reset(now, false)
	}
	t.s.unlock()

	return
}

// NewTimer creates a new Timer that will send the current time on its
// channel after at least duration d.
func (c *Clock[T, D, RT]) NewTimer(d D) *Timer[T, D] {
	now := c.Now()
	ch := make(chan T, 1)
	tm := &timer[T, D]{
		f: func(when T) {
			select {
			case ch <- when:
			default:
			}
		},
		when: now.Add(d),
	}
	w := <-c.waker
	w.lock()
	w.schedule(tm)
	if tm.index == 0 {
		w.reset(now, false)
	}
	w.unlock()
	c.waker <- w
	return &Timer[T, D]{ch, tm, w}
}

// After waits for the duration to elapse and then sends the current time on
// the returned channel. It is equivalent to clock.NewTimer(d).C(). The
// underlying Timer is not recovered by the garbage collector until the timer
// fires. If efficiency is a concern, use clock.NewTimer instead and call
// Timer.Stop if the timer is no longer needed.
func (c *Clock[T, D, RT]) After(d D) <-chan T {
	return c.NewTimer(d).c
}

// AfterFunc waits for the duration to elapse and then calls f in its own
// goroutine. It returns a Timer that can be used to cancel the call using
// its Stop method.
func (c *Clock[T, D, RT]) AfterFunc(d D, f func()) *Timer[T, D] {
	now := c.Now()
	tm := &timer[T, D]{
		f:    func(T) { go f() },
		when: now.Add(d),
	}
	w := <-c.waker
	w.lock()
	w.schedule(tm)
	if tm.index == 0 {
		w.reset(now, false)
	}
	w.unlock()
	c.waker <- w
	return &Timer[T, D]{t: tm, s: w}
}
