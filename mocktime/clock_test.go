package mocktime_test

import (
	"testing"

	. "github.com/noodlebox/clock/mocktime"
)

func BenchmarkNow(b *testing.B) {
	benchmark(b, func(n int) {
		for i := 0; i < n; i++ {
			_ = Now()
		}
	})
}

func BenchmarkClockNextAt(b *testing.B) {
	benchmark(b, func(n int) {
		for i := 0; i < n; i++ {
			_ = NextAt()
		}
	})
}

func BenchmarkClockScale(b *testing.B) {
	benchmark(b, func(n int) {
		for i := 0; i < n; i++ {
			SetScale(float64(n) / float64(n+1))
		}
	})
	SetScale(1.0)
}

func BenchmarkClockStopStart(b *testing.B) {
	benchmark(b, func(n int) {
		for i := 0; i < n; i++ {
			Stop()
			Start()
		}
	})
}

func BenchmarkClockSet(b *testing.B) {
	start := Now()
	benchmark(b, func(n int) {
		for i := 0; i < n; i++ {
			Set(start.Add(Duration(i) * Second))
		}
	})
	Set(start)
}

func BenchmarkClockStopSet(b *testing.B) {
	Stop()
	start := Now()
	benchmark(b, func(n int) {
		for i := 0; i < n; i++ {
			Set(start.Add(Duration(i) * Second))
		}
	})
	Set(start)
	Start()
}

func BenchmarkClockStep(b *testing.B) {
	start := Now()
	benchmark(b, func(n int) {
		for i := 0; i < n; i++ {
			Step(Millisecond)
		}
	})
	Set(start)
}

func BenchmarkClockStopStep(b *testing.B) {
	Stop()
	start := Now()
	benchmark(b, func(n int) {
		for i := 0; i < n; i++ {
			Step(Millisecond)
		}
	})
	Set(start)
	Start()
}
