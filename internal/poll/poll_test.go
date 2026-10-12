package poll

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// testOptions has a Timeout that expires only when a test shortens it.
func testOptions() Options {
	return Options{
		Timeout: 2 * time.Second,
		Initial: time.Millisecond,
		Max:     2 * time.Millisecond,
	}
}

// sequence repeats the last state once the states run out.
func sequence(states ...string) (func() (string, bool, error), *int) {
	reads := 0
	return func() (string, bool, error) {
		state := states[min(reads, len(states)-1)]
		reads++
		return state, state != "pending", nil
	}, &reads
}

func TestUntilReturnsTheFirstValueThatIsDone(t *testing.T) {
	read, reads := sequence("pending", "pending", "valid")

	value, err := Until(context.Background(), testOptions(), read)

	require.NoError(t, err)
	require.Equal(t, "valid", value)
	require.Equal(t, 3, *reads)
}

func TestUntilStopsOnAFailedRead(t *testing.T) {
	failure := errors.New("read failed")
	reads := 0
	read := func() (string, bool, error) {
		reads++
		if reads == 2 {
			return "", false, failure
		}
		return "pending", false, nil
	}

	value, err := Until(context.Background(), testOptions(), read)

	require.ErrorIs(t, err, failure)
	require.Empty(t, value)
	require.Equal(t, 2, reads, "a failed read ends the wait rather than being polled through")
}

func TestUntilStopsWhenTheCallerGivesUp(t *testing.T) {
	read, reads := sequence("pending")
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	value, err := Until(ctx, testOptions(), read)

	require.ErrorIs(t, err, context.Canceled)
	require.Empty(t, value)
	require.Equal(t, 0, *reads, "a cancelled wait reads nothing")
}

func TestUntilGivesUpWithTheTimeItWaitedAndNoValue(t *testing.T) {
	read, _ := sequence("pending")
	options := testOptions()
	options.Timeout = 20 * time.Millisecond

	value, err := Until(context.Background(), options, read)

	require.Empty(t, value, "a wait that expired must never carry a value")
	var expired *ExpiredError
	require.True(t, errors.As(err, &expired), "expected an expired error, got %T", err)
	// Rounded to the millisecond before it is reported, so allow that much.
	require.GreaterOrEqual(t, expired.Waited+time.Millisecond, options.Timeout)
	require.Equal(t, expired.Waited.Round(time.Millisecond), expired.Waited)
}

func TestNextIntervalDoublesUpToTheCeiling(t *testing.T) {
	const max = 5 * time.Second

	for _, test := range []struct {
		current time.Duration
		want    time.Duration
	}{
		{500 * time.Millisecond, time.Second},
		{time.Second, 2 * time.Second},
		{2 * time.Second, 4 * time.Second},
		{4 * time.Second, max},
		{max, max},
	} {
		require.Equal(t, test.want, nextInterval(test.current, max),
			"doubling %s", test.current)
	}
}
