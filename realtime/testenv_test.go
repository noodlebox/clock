package realtime_test

import (
	"flag"
	"testing"

	. "github.com/noodlebox/clock/realtime"
)

// A clock instance for use in other tests
var time Clock

func init() {
	time = NewClock()
}

// Some helper functions from src/internal/testenv/testenv.go

var flaky = flag.Bool("flaky", false, "run known-flaky tests too")

func SkipFlaky(t testing.TB, issue int) {
	t.Helper()
	if !*flaky {
		t.Skipf("skipping known flaky test without the -flaky flag; see golang.org/issue/%d", issue)
	}
}

// FIXME: Not sure of the best way to handle this specifically. Interrupt
// normally has platform-specific behavior, defined as interrupt() in
// src/time/sys_*.go and aliased in src/time/internal_test.go. Several
// platform-specific implementations are a noop, so this should be fine as
// is for now, though do check on whether there's some important platform-
// specific behavior worth testing here.
func Interrupt() {}
