package steppedtime_test

import (
	"fmt"
	"runtime"
	"sync"
	"testing"

	truetime "time"

	. "github.com/noodlebox/clock/steppedtime"
)

var time = Clock{}

func init() {
	// Wow is this ugly, but it works for dropping in existing test code.
	// FIXME: Do better
	go func() {
		for i := 0 * Second; i < 5*Minute; i += Millisecond {
			time.Step(Millisecond)
			truetime.Sleep(Millisecond)
		}
	}()
}

func benchmark(b *testing.B, bench func(n int)) {

	// Create equal number of garbage timers on each P before starting
	// the benchmark.
	var wg sync.WaitGroup
	garbageAll := make([][]*Timer, runtime.GOMAXPROCS(0))
	for i := range garbageAll {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			garbage := make([]*Timer, 1<<15)
			for j := range garbage {
				garbage[j] = time.AfterFunc(Hour, nil)
			}
			garbageAll[i] = garbage
		}(i)
	}
	wg.Wait()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			bench(1000)
		}
	})
	b.StopTimer()

	for _, garbage := range garbageAll {
		for _, t := range garbage {
			t.Stop()
		}
	}
}

// These tests are mostly copied from src/time/tick_test.go

func TestTicker(t *testing.T) {
	// We want to test that a ticker takes as much time as expected.
	// Since we don't want the test to run for too long, we don't
	// want to use lengthy times. This makes the test inherently flaky.
	// Start with a short time, but try again with a long one if the
	// first test fails.

	baseCount := 10
	baseDelta := 20 * Millisecond

	// On Darwin ARM64 the tick frequency seems limited. Issue 35692.
	if (runtime.GOOS == "darwin" || runtime.GOOS == "ios") && runtime.GOARCH == "arm64" {
		// The following test will run ticker count/2 times then reset
		// the ticker to double the duration for the rest of count/2.
		// Since tick frequency is limited on Darwin ARM64, use even
		// number to give the ticks more time to let the test pass.
		// See CL 220638.
		baseCount = 6
		baseDelta = 100 * Millisecond
	}

	var errs []string
	logErrs := func() {
		for _, e := range errs {
			t.Log(e)
		}
	}

	for _, test := range []struct {
		count int
		delta Duration
	}{{
		count: baseCount,
		delta: baseDelta,
	}, {
		count: 8,
		delta: 1 * Second,
	}} {
		count, delta := test.count, test.delta
		ticker := time.NewTicker(delta)
		t0 := time.Now()
		for i := 0; i < count/2; i++ {
			<-ticker.C()
		}
		ticker.Reset(delta * 2)
		for i := count / 2; i < count; i++ {
			<-ticker.C()
		}
		ticker.Stop()
		t1 := time.Now()
		dt := t1.Sub(t0)
		target := 3 * delta * Duration(count/2)
		slop := target * 3 / 10
		if dt < target-slop || dt > target+slop {
			errs = append(errs, fmt.Sprintf("%d %s ticks then %d %s ticks took %s, expected [%s,%s]", count/2, delta, count/2, delta*2, dt, target-slop, target+slop))
			if dt > target+slop {
				// System may be overloaded; sleep a bit
				// in the hopes it will recover.
				time.Sleep(Second / 2)
			}
			continue
		}
		// time.Now test that the ticker stopped.
		time.Sleep(2 * delta)
		select {
		case <-ticker.C():
			errs = append(errs, "Ticker did not shut down")
			continue
		default:
			// ok
		}

		// Test passed, so all done.
		if len(errs) > 0 {
			t.Logf("saw %d errors, ignoring to avoid flakiness", len(errs))
			logErrs()
		}

		return
	}

	t.Errorf("saw %d errors", len(errs))
	logErrs()
}

// Test that a bug tearing down a ticker has been fixed. This routine should not deadlock.
func TestTeardown(t *testing.T) {
	Delta := 100 * Millisecond
	if testing.Short() {
		Delta = 20 * Millisecond
	}
	for i := 0; i < 3; i++ {
		ticker := time.NewTicker(Delta)
		<-ticker.C()
		ticker.Stop()
	}
}

// Test the time.Tick convenience wrapper.
func TestTick(t *testing.T) {
	// Test that giving a negative duration returns nil.
	if got := time.Tick(-1); got != nil {
		t.Errorf("time.Tick(-1) = %v; want nil", got)
	}
}

// Test that time.NewTicker panics when given a duration less than zero.
func TestNewTickerLtZeroDuration(t *testing.T) {
	defer func() {
		if err := recover(); err == nil {
			t.Errorf("time.NewTicker(-1) should have panicked")
		}
	}()
	time.NewTicker(-1)
}

// Test that Ticker.Reset panics when given a duration less than zero.
func TestTickerResetLtZeroDuration(t *testing.T) {
	defer func() {
		if err := recover(); err == nil {
			t.Errorf("Ticker.Reset(0) should have panicked")
		}
	}()
	tk := time.NewTicker(Second)
	tk.Reset(0)
}

func BenchmarkTicker(b *testing.B) {
	benchmark(b, func(n int) {
		ticker := time.NewTicker(Nanosecond)
		for i := 0; i < n; i++ {
			<-ticker.C()
		}
		ticker.Stop()
	})
}

func BenchmarkTickerReset(b *testing.B) {
	benchmark(b, func(n int) {
		ticker := time.NewTicker(Nanosecond)
		for i := 0; i < n; i++ {
			ticker.Reset(Nanosecond * 2)
		}
		ticker.Stop()
	})
}

func BenchmarkTickerResetNaive(b *testing.B) {
	benchmark(b, func(n int) {
		ticker := time.NewTicker(Nanosecond)
		for i := 0; i < n; i++ {
			ticker.Stop()
			ticker = time.NewTicker(Nanosecond * 2)
		}
		ticker.Stop()
	})
}
