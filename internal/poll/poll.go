// Package poll asks the server about something that finishes asynchronously
// until it is done, backing off between reads.
package poll

import (
	"context"
	"fmt"
	"time"
)

// Options bounds how long a read is repeated and how often. Every field must be
// positive; callers own their defaults, since each resource has its own budget.
//
// Timeout governs the polling, not any single read. A read is bounded by the
// HTTP client's own retry policy instead, which has no overall deadline, so a
// server that answers very slowly can overrun this budget by one read. Holding
// to it exactly would mean giving the read a deadline of its own, which the
// shared client cannot take today.
type Options struct {
	Timeout time.Duration
	Initial time.Duration
	Max     time.Duration
}

// ExpiredError says the budget ran out before the read reported done. It
// carries no value, so an unfinished answer is never mistaken for a final one.
type ExpiredError struct {
	Waited time.Duration
}

func (e *ExpiredError) Error() string {
	return fmt.Sprintf("still pending after %s", e.Waited)
}

// Until calls read until it reports done, read fails, ctx ends, or the budget
// expires. It returns a value only when read reported it done.
//
// Cancelling ctx stops the polling but cannot abort a read already in flight,
// for the same reason the budget cannot: the read carries no context. Both take
// effect at the next turn of the loop.
func Until[T any](ctx context.Context, options Options, read func() (T, bool, error)) (T, error) {
	var zero T
	// Measured rather than assumed: a read carries no deadline of its own, so
	// the time actually spent can exceed the budget, and reporting the budget
	// would understate the wait to the one person asking whether the platform
	// met it.
	started := time.Now()
	expired := time.After(options.Timeout)
	interval := options.Initial

	for {
		// Checked before the read, so a caller who has already given up costs
		// the server nothing.
		if err := ctx.Err(); err != nil {
			return zero, err
		}

		value, done, err := read()
		if err != nil {
			// Transient trouble was already retried inside the HTTP client, so
			// what reaches here will not improve by being asked again.
			return zero, err
		}
		if done {
			return value, nil
		}

		select {
		case <-ctx.Done():
			return zero, ctx.Err()
		case <-expired:
			// Rounded: the point of the number is that the budget was spent, and
			// six digits of nanoseconds are noise in that sentence. An overrun
			// still shows, which is why it is measured at all.
			return zero, &ExpiredError{Waited: time.Since(started).Round(time.Millisecond)}
		case <-time.After(interval):
		}
		interval = nextInterval(interval, options.Max)
	}
}

// nextInterval doubles the wait between reads, never past max. Backing off
// keeps a slow resource from being asked about hundreds of times, while the
// first few reads stay close together so a fast one answers promptly.
func nextInterval(current, max time.Duration) time.Duration {
	if doubled := current * 2; doubled < max {
		return doubled
	}
	return max
}
