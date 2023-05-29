// The clock package duplicates much of the functionality of `time` from the
// standard library, while allowing for greater control over the flow of time
// for use in testing or simulations. Each `Clock` instance may track its own
// flow of time, appropriately triggering `Timer`s and `Ticker`s created with
// them. There are several implementations supplied in subpackages:
//
//   realtime provides a thin wrapper around the time package.
//
//   steppedtime provides a basic clock implementation using a simpler time
//   representation starting at zero and counting upwards. It advances only
//   when explicitly stepped.
//
//   relativetime provides a clock that can be set to track a reference clock
//   with a specified offset and scaling factor, and may start, stop, or
//   adjust tracking parameters while running. It uses a generic interface so
//   that it may be used with clocks using various implementations of time or
//   duration values.
//
//   mocktime uses the relativetime and realtime packages to provide a drop
//   in replacement for a realtime clock that may be controlled as a relative
//   clock. It also provides package-level functions matching builtin `time`.
//
// `Clock`s created by this package should generally behave monotonically,
// unless explicitly set or stepped backwards. Their public interface should
// also be safe to use from multiple threads.
package clock
