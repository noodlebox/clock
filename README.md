# clock
The clock package duplicates much of the functionality of `time` from the standard library, while allowing for greater control over the flow of time for use in testing and simulations. Each clock instance may track its own flow of time (using its own definitions of Time or Duration as necessary), while appropriately triggering Timers and Tickers created with them. These may be useful for mocking the `time` package in test code as well as working with simulation timers while retaining the familiar API of the `time` package.

In general, clocks provided by these packages should behave monotonically (unless explicitly set or stepped backwards) and should also provide a thread-safe public interface. There are several implementations supplied in subpackages below.

The design of this package is mostly complete, though there may still be some bugs to work out or minor API changes before a stable release.

## clock/realtime
A thin wrapper around the `time` package. One important caveat is that Timers and Tickers provide access to their channel via a `C()` method rather than a field of the same name. This was decided to permit easier specification of interfaces.

## clock/steppedtime
A basic clock implementation using a simple time representation that starts at zero and counts upwards. It advances only when explicitly stepped.

## clock/relativetime
A clock that can be set to track another clock as a reference with a specified offset and scaling factor. It may start, stop, or adjust any tracking parameters at runtime, with timers created on it behaving appropriately. It is defined with a generic interface so that it may be used with clocks that use various implementations of time or duration values.

## clock/mocktime
Uses relativetime and realtime to implement a drop in replacement for a realtime clock with all the additional control of a relative clock. It also provides package-level functions to match the API of the standard library's `time` package, for mocking purposes. Note that the caveats for Timers and Tickers mentioned for realtime clocks above apply here as well.
