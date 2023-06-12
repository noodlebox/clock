// Package clock offers a more object-oriented interface to the [time]
// package from the standard library in the form of "Clocks", allowing
// greater control over the flow of time. Clocks also enable isolation
// between individually configurable instances, analogous to [rand]'s
// alternative interface via instances of [rand.Rand]. Each Clock instance
// may track its own flow of time, appropriately triggering Timers and
// Tickers created with them. There are several implementations supplied by
// subpackages.
package clock
