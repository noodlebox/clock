package relativetime

// The minimal API needed to serve as a reference for a RelativeClock
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
}

// Minimal API needed for a Duration implementation
// For example, `time.Duration` satisfies this interface
type Duration interface {
	Seconds() float64
}

type Clock[T Time[T, D], D Duration, RT RTimer[T, D]] struct {
	Ref RClock[T, D, RT]
}

func NewClock[T Time[T, D], D Duration, RT RTimer[T, D]](ref RClock[T, D, RT]) *Clock[T, D, RT] {
	return &Clock[T, D, RT]{ref}
}

func (c *Clock[T, D, RT]) Now() T {
	return c.Ref.Now()
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
