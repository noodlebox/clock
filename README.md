# clock
The clock package duplicates much of the functionality of the `time` package of the standard library, while allowing for greater control over the flow of time for use in testing and simulations. Each clock instance may track its own flow of time (using its own definitions of Time or Duration as necessary), while appropriately triggering Timers and Tickers created with them. These may be useful both for mocking the `time` package in test code as well as working with simulation timers with the familiar API of the `time` package.

In general, clocks provided by these packages should behave monotonically (unless explicitly set or stepped backwards) and should also provide a thread-safe public interface. There are several implementations supplied in subpackages below.

## clock/realtime
A thin wrapper around the `time` package. One important caveat is that Timers and Tickers provide access to their channel via a `C()` method rather than a field of the same name.

## clock/steppedtime
A basic clock implementation using a simpler time representation that starts at zero and counts upwards. It advances only when explicitly stepped.

## clock/relativetime
A clock that can be set to track another clock as a reference with a specified offset and scaling factor. It may start, stop, or adjust any tracking parameters at runtime, with timers created on it behaving appropriately. It is defined with a generic interface so that it may be used with clocks using various implementations of time or duration values.

## clock/mocktime
Uses relativetime and realtime above to provide a drop in replacement for a realtime clock with all the additional control of a relative clock. It also provides package-level functions to match API of the `time` standard library. Note that the caveats for Timers and Tickers mentioned for realtime clocks mentoned above apply here as well.
