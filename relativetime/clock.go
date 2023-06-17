package relativetime

import (
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
	waker  chan *clock[T, D, RT]
	wakers [nwakers]*clock[T, D, RT]
	keeper *clock[T, D, RT]

	mu sync.Mutex // Protects collecting all wakers
}

// NewClock returns a new Clock set to at synchronized to the current time on
// ref with a scale factor of scale.
func NewClock[T Time[T, D], D Duration, RT RTimer[D]](ref RClock[T, D, RT], at T, scale float64) (c *Clock[T, D, RT]) {
	rNow := ref.Now()
	c = &Clock[T, D, RT]{
		waker: make(chan *clock[T, D, RT], nwakers),
		keeper: &clock[T, D, RT]{
			ref:    ref,
			active: false,
			scale:  scale,
			now:    at,
			rNow:   rNow,
		},
	}
	for i, _ := range c.wakers {
		w := &clock[T, D, RT]{
			ref:    ref,
			active: false,
			scale:  scale,
			now:    at,
			rNow:   rNow,
			waking: make(chan struct{}, 1),
		}
		c.waker <- w
		c.wakers[i] = w
	}
	return
}

type clock[T Time[T, D], D Duration, RT RTimer[D]] struct {
	ref       RClock[T, D, RT]
	scale     float64
	active    bool
	now, rNow T // last sync point

	queue  queue[T, D] // Upcoming events, in local time
	waker  RTimer[D]   // Interface used here for a default value of nil
	wakeAt T           // Local time of next scheduled waking
	waking chan struct{}

	sync.RWMutex

	//*Clock[T, D, RT]
}

// Syncing with the reference clock is done lazily. This method updates the
// sync points based on difference between a new reference time and the last
// sync point. Fields that would affect how the reference is tracked should
// not change between resyncs. This should be called to ensure sync points
// are not stale before any change to one of these fields.
// Callers must hold a write lock.
func (c *clock[T, D, RT]) advanceRef(rNow T) {
	c.now = c.toLocal(rNow)
	c.rNow = rNow
}

func (c *clock[T, D, RT]) sync() T {
	c.advanceRef(c.ref.Now())
	return c.now
}

// Given a reference time, extrapolate to the local time. Times before the
// last sync point (c.rNow) are not guaranteed to be extrapolated correctly.
// Callers must hold at least a read lock.
func (c *clock[T, D, RT]) toLocal(when T) T {
	then := c.rNow

	// No local change if stopped, scale is zero, or ref clock hasn't changed
	if !c.active || c.scale == 0.0 || when.Equal(then) {
		return c.now
	}
	dt := when.Sub(then)
	if c.scale != 1.0 {
		// Apply scale via conversion to float64 in seconds
		dt = c.ref.Seconds(dt.Seconds() * c.scale)
	}
	// We're at now now.
	return c.now.Add(dt)
}

func (c *clock[T, D, RT]) stopWaker() {
	if c.waker == nil {
		return
	}
	c.waker.Stop()
	var zero T
	c.wakeAt = zero
}

func (c *clock[T, D, RT]) resetWaker() {
	if !c.active || c.scale == 0.0 {
		// Local time isn't changing
		c.stopWaker()
		return
	}

	next := c.queue.peek()
	if next == nil {
		// Nothing currently scheduled
		c.stopWaker()
		return
	}

	if c.waker != nil && next.when.Equal(c.wakeAt) {
		// Waker already set to the correct time, let it be
		return
	}
	select {
	case c.waking <- struct{}{}:
		<-c.waking
	default:
		return
	}

	c.wakeAt = next.when

	// Duration on reference clock until next timer should trigger
	dt := c.ref.Seconds(next.when.Sub(c.now).Seconds() / c.scale)

	if c.waker == nil {
		c.waker = c.ref.AfterFunc(dt, c.wake)
	} else {
		c.waker.Reset(dt)
	}
}

// Check schedule for pending events that should trigger now.
func (c *clock[T, D, RT]) checkSchedule() {
	for t := c.queue.peek(); t != nil && !t.when.After(c.now); t = c.queue.peek() {
		if t.period.Seconds() <= 0 {
			c.unschedule(t)
		} else {
			t.when = c.now.Add(t.period)
			c.reschedule(t)
		}
		t.f(c.now)
	}
}

func (c *clock[T, D, RT]) schedule(t *timer[T, D]) {
	c.queue.insert(t)
}

func (c *clock[T, D, RT]) unschedule(t *timer[T, D]) {
	if t.index < 0 {
		return
	}
	c.queue.remove(t)
}

func (c *clock[T, D, RT]) reschedule(t *timer[T, D]) {
	if t.index < 0 {
		c.queue.insert(t)
		return
	}
	c.queue.fix(t)
}

// This method is called whenever a reference timer triggers.
func (c *clock[T, D, RT]) wake() {
	select {
	case c.waking <- struct{}{}:
	default:
		return
	}
	c.Lock()
	<-c.waking
	c.sync()
	c.checkSchedule()
	c.resetWaker()
	c.Unlock()
}

// Call f (with read access) on a clock.
//	w := <-c.waker
//	w.RLock()
//	c.waker <- w
//	f(w)
//	w.RUnlock()

// Call f (with write access) on a clock.
//	w := <-c.waker
//	w.Lock()
//	f(w)
//	w.Unlock()
//	c.waker <- w

// Call f (with write access) on all clocks. This method blocks at least
// until locks have been acquired on each clock, with each clock unlocking
// when finished. This ensures that any following calls will get a synced
// clock. Other threads may race to acquire read locks on clocks, but once
// this thread has acquired a lock, further calls will block until a clock
// has finished.
func (c *Clock[T, D, RT]) sync(f func(*clock[T, D, RT])) {
	c.mu.Lock()
	var wg sync.WaitGroup
	wg.Add(len(c.wakers))
	for _, w := range c.wakers {
		go func(w *clock[T, D, RT]) {
			w.Lock()
			wg.Done()
			f(w)
			w.Unlock()
		}(w)
	}
	c.keeper.Lock()
	f(c.keeper)
	c.keeper.Unlock()
	wg.Wait()
	c.mu.Unlock()
}

// Start begins tracking the reference clock, if not already running. It is
// fine to call Start() on a clock that is already running.
func (c *Clock[T, D, RT]) Start() {
	rNow := c.keeper.ref.Now()
	c.sync(func(w *clock[T, D, RT]) {
		// Sync up before changing setting
		w.advanceRef(rNow)
		w.active = true

		w.resetWaker()
	})
}

// Stop stops tracking the reference clock, if currently running. It is fine
// to call Stop() on a clock that is not running.
func (c *Clock[T, D, RT]) Stop() {
	rNow := c.keeper.ref.Now()
	c.sync(func(w *clock[T, D, RT]) {
		// Sync up before changing setting
		w.advanceRef(rNow)
		w.active = false

		w.resetWaker()
	})
}

// Active returns true if currently tracking the reference clock.
func (c *Clock[T, D, RT]) Active() (active bool) {
	c.keeper.RLock()
	active = c.keeper.active
	c.keeper.RUnlock()
	return
}

// SetScale sets the scaling factor for tracking the reference clock.
func (c *Clock[T, D, RT]) SetScale(scale float64) {
	rNow := c.keeper.ref.Now()
	c.sync(func(w *clock[T, D, RT]) {
		// Sync up before changing setting
		w.advanceRef(rNow)
		w.scale = scale

		w.resetWaker()
	})
}

// Scale returns the scaling factor for tracking the reference clock.
func (c *Clock[T, D, RT]) Scale() (scale float64) {
	c.keeper.RLock()
	scale = c.keeper.scale
	c.keeper.RUnlock()
	return
}

// Set sets the local sync point with the current reference time to now. If
// any timers are active, a value of now earlier than the previous setting
// may lead to undefined behavior.
func (c *Clock[T, D, RT]) Set(now T) {
	rNow := c.keeper.ref.Now()
	c.sync(func(w *clock[T, D, RT]) {
		// Reset sync point to given time
		w.now, w.rNow = now, rNow

		w.checkSchedule()
		w.resetWaker()
	})
}

// Step advances the local time forward by dt. If any timers are active, a
// negative value for dt may lead to undefined behavior.
func (c *Clock[T, D, RT]) Step(dt D) {
	rNow := c.keeper.ref.Now()
	c.sync(func(w *clock[T, D, RT]) {
		// Sync up before changing setting
		w.advanceRef(rNow)
		w.now = w.now.Add(dt)

		w.checkSchedule()
		w.resetWaker()
	})
}

// NextAt returns the time at which the next scheduled timer should trigger.
// If no timers are scheduled, returns a zero value.
func (c *Clock[T, D, RT]) NextAt() (when T) {
	c.mu.Lock()
	var wg sync.WaitGroup
	wg.Add(len(c.wakers))
	ch := make(chan T, 1)
	var zero T
	ch <- zero
	for _, w := range c.wakers {
		go func(w *clock[T, D, RT]) {
			w.RLock()
			next := w.queue.peek()
			if next != nil {
				when := <-ch
				if when.IsZero() || when.After(next.when) {
					ch <- next.when
				} else {
					ch <- when
				}
			}
			wg.Done()
			w.RUnlock()
		}(w)
	}
	wg.Wait()
	c.mu.Unlock()
	return <-ch
}

// Seconds returns a Duration value representing n Seconds. This is provided
// to allow a relative clock itself to satisfy the reference clock interface.
func (c *Clock[T, D, RT]) Seconds(n float64) D {
	return c.keeper.ref.Seconds(n)
}

// Now returns the current time.
func (c *Clock[T, D, RT]) Now() (now T) {
	c.keeper.RLock()
	now = c.keeper.toLocal(c.keeper.ref.Now())
	c.keeper.RUnlock()
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

	w := <-c.waker
	w.Lock()
	ch := make(chan struct{})
	tm := &timer[T, D]{
		f:    func(T) { close(ch) },
		when: w.sync().Add(d),
	}
	w.schedule(tm)
	if tm.index == 0 {
		w.resetWaker()
	}
	w.Unlock()
	c.waker <- w
	<-ch
}

type scheduler[T Time[T, D], D Duration] interface {
	schedule(t *timer[T, D])
	unschedule(t *timer[T, D])
	reschedule(t *timer[T, D])
	resetWaker()
	Lock()
	Unlock()
	sync() T
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

	t.s.Lock()
	t.t.when = t.s.sync().Add(d)
	t.t.period = d
	isNext := t.t.index == 0
	t.s.reschedule(t.t)
	if isNext || t.t.index == 0 {
		t.s.resetWaker()
	}
	t.s.Unlock()
}

// Stop turns off a ticker. After Stop, no more ticks will be sent. Stop does
// not close the channel, to prevent a concurrent goroutine reading from the
// channel from seeing an erroneous "tick".
func (t *Ticker[T, D]) Stop() {
	if t.t == nil {
		panic("Stop called on uninitialized relativetime.Ticker")
	}

	t.s.Lock()
	isNext := t.t.index == 0
	t.s.unschedule(t.t)
	if isNext {
		t.s.sync()
		t.s.resetWaker()
	}
	t.s.Unlock()
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

	w := <-c.waker
	w.Lock()
	ch := make(chan T)
	tm := &timer[T, D]{
		when:   w.sync().Add(d),
		period: d,
	}
	wait := make(chan struct{}, 1)
	tm.f = func(when T) {
		select {
		case ch <- when:
		default:
			w.unschedule(tm)
			tm.index = -2
			select {
			case wait <- struct{}{}:
			default:
				// Already waiting with a value
				return
			}
			go func() {
				ch <- when
				w.Lock()
				<-wait
				if tm.index > -2 {
					// Reset() or Stop() was called while waiting
					w.Unlock()
					return
				}
				tm.when = w.sync().Add(tm.period)
				w.schedule(tm)
				if tm.index == 0 {
					w.resetWaker()
				}
				w.Unlock()
			}()
		}
	}
	w.schedule(tm)
	if tm.index == 0 {
		w.resetWaker()
	}
	w.Unlock()
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

	t.s.Lock()

	t.t.when = t.s.sync().Add(d)
	active = t.t.index >= 0
	isNext := t.t.index == 0
	t.s.reschedule(t.t)
	if isNext || t.t.index == 0 {
		t.s.resetWaker()
	}
	t.s.Unlock()

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

	t.s.Lock()

	active = t.t.index >= 0
	isNext := t.t.index == 0
	t.s.unschedule(t.t)
	if isNext {
		t.s.sync()
		t.s.resetWaker()
	}
	t.s.Unlock()

	return
}

// NewTimer creates a new Timer that will send the current time on its
// channel after at least duration d.
func (c *Clock[T, D, RT]) NewTimer(d D) *Timer[T, D] {
	w := <-c.waker
	w.Lock()
	ch := make(chan T, 1)
	tm := &timer[T, D]{
		f: func(when T) {
			select {
			case ch <- when:
			default:
			}
		},
		when: w.sync().Add(d),
	}
	w.schedule(tm)
	if tm.index == 0 {
		w.resetWaker()
	}
	w.Unlock()
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
	w := <-c.waker
	w.Lock()
	tm := &timer[T, D]{
		f:    func(T) { go f() },
		when: w.sync().Add(d),
	}
	w.schedule(tm)
	if tm.index == 0 {
		w.resetWaker()
	}
	w.Unlock()
	c.waker <- w
	return &Timer[T, D]{t: tm, s: w}
}
