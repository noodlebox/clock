// Package realtime provides a thin wrapper around the [time] package. It
// works with [time.Time] and [time.Duration] values. [Timer] and [Ticker]
// override their corresponding C fields with a method, to work around the
// limitation of interfaces not being able to specify fields.
package realtime
