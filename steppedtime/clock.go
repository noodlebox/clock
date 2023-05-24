package steppedtime

type Clock struct {
	now   Time
	queue queue
}

func NewClock() *Clock {
	return &Clock{}
}

// If any timers are active, a value of `now` earlier than the previous
// setting may lead to undefined behavior.
func (c *Clock) Set(now Time) {
	c.now = now

	// Check whether we're due for any scheduled events
	c.checkSchedule()
}

// If any timers are active, a negative value for dt may lead to undefined
// behavior.
func (c *Clock) Step(dt Duration) {
	c.now = c.now.Add(dt)

	// Check whether we're due for any scheduled events
	c.checkSchedule()
}

func (c *Clock) Now() Time {
	return c.now
}

func (c *Clock) Since(t Time) Duration {
	return c.now.Sub(t)
}

func (c *Clock) Until(t Time) Duration {
	return t.Sub(c.now)
}

func (c *Clock) Sleep(d Duration) {
	if d <= 0 {
		return
	}

	ch := make(chan struct{})
	c.schedule(&timer{
		f:    func(Time) { close(ch) },
		when: c.now.Add(d),
	})
	<-ch
}

type Ticker struct {
	c <-chan Time
	t *timer
	s *Clock
}

func (t *Ticker) C() <-chan Time {
	return t.c
}

func (t *Ticker) Reset(d Duration) {
	if d <= 0 {
		panic("non-positive interval for steppedtime.Ticker.Reset")
	}
	if t.t == nil {
		panic("Reset called on uninitialized steppedtime.Ticker")
	}

	t.t.when = t.s.now.Add(d)
	t.t.period = d
	t.s.reschedule(t.t)
}

func (t *Ticker) Stop() {
	if t.t == nil {
		panic("Stop called on uninitialized steppedtime.Ticker")
	}

	t.s.unschedule(t.t)
}

func (c *Clock) NewTicker(d Duration) *Ticker {
	if d <= 0 {
		panic("non-positive interval for steppedtime.Clock.NewTicker")
	}

	ch := make(chan Time, 1)
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
	return &Ticker{ch, tm, c}
}

func (c *Clock) Tick(d Duration) <-chan Time {
	if d <= 0 {
		return nil
	}

	return c.NewTicker(d).c
}

type Timer struct {
	c <-chan Time
	t *timer
	s *Clock
}

func (t *Timer) C() <-chan Time {
	return t.c
}

func (t *Timer) Reset(d Duration) (active bool) {
	if t.t == nil {
		panic("Reset called on uninitialized steppedtime.Timer")
	}

	t.t.when = t.s.now.Add(d)
	active = (t.t.index != -1)
	t.s.reschedule(t.t)
	return
}

func (t *Timer) Stop() (active bool) {
	if t.t == nil {
		panic("Stop called on uninitialized steppedtime.Timer")
	}

	active = (t.t.index != -1)
	t.s.unschedule(t.t)
	return
}

func (c *Clock) NewTimer(d Duration) *Timer {
	ch := make(chan Time, 1)
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
	return &Timer{ch, tm, c}
}

func (c *Clock) After(d Duration) <-chan Time {
	return c.NewTimer(d).c
}

func (c *Clock) AfterFunc(d Duration, f func()) *Timer {
	tm := &timer{
		f:    func(Time) { go f() },
		when: c.now.Add(d),
	}
	c.schedule(tm)
	return &Timer{t: tm, s: c}
}
