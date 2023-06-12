// Package relativetime provides a clock that can be set to track a reference
// clock with a specified offset and scaling factor, and may start, stop, or
// adjust tracking parameters while running. It uses a generic interface so
// that it may be used with clocks using various implementations of time or
// duration values.
package relativetime
