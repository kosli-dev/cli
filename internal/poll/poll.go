// Package poll repeats a read, backing off, until it reports done.
package poll

import (
	"context"
	"fmt"
	"time"
)

// Options must have every field positive. Timeout bounds the polling, not a read
// in flight, so one slow read can overrun it.
type Options struct {
	Timeout time.Duration
	Initial time.Duration
	Max     time.Duration
}

// ExpiredError says the budget ran out before the read reported done.
type ExpiredError struct {
	Waited time.Duration
}

func (e *ExpiredError) Error() string {
	return fmt.Sprintf("still pending after %s", e.Waited)
}

// Until calls read until it reports done, read fails, ctx ends, or the budget
// expires. It returns a value only when read reported it done. Cancelling ctx
// cannot abort a read in flight.
func Until[T any](ctx context.Context, options Options, read func() (T, bool, error)) (T, error) {
	var zero T
	// Measured, because a slow read can overrun the budget.
	started := time.Now()
	expired := time.After(options.Timeout)
	interval := options.Initial

	for {
		// Checked first, so a cancelled caller sends no read.
		if err := ctx.Err(); err != nil {
			return zero, err
		}

		value, done, err := read()
		if err != nil {
			// The HTTP client has already retried transient errors.
			return zero, err
		}
		if done {
			return value, nil
		}

		select {
		case <-ctx.Done():
			return zero, ctx.Err()
		case <-expired:
			return zero, &ExpiredError{Waited: time.Since(started).Round(time.Millisecond)}
		case <-time.After(interval):
		}
		interval = nextInterval(interval, options.Max)
	}
}

func nextInterval(current, max time.Duration) time.Duration {
	if doubled := current * 2; doubled < max {
		return doubled
	}
	return max
}
