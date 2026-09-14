package evaluations

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/kosli-dev/cli/internal/requests"
)

// DefaultWaitOptions is the budget the platform states for an evaluation:
// thirty seconds from asking to a terminal status. Waiting longer would
// outlast the guarantee; giving up sooner would call a healthy evaluation
// unfinished.
var DefaultWaitOptions = WaitOptions{
	Timeout: 30 * time.Second,
	Initial: 500 * time.Millisecond,
	Max:     5 * time.Second,
}

// WaitOptions bounds how long an evaluation is asked about and how often. A
// zero field takes its default.
//
// Timeout governs the polling, not any single read. A read is bounded by the
// HTTP client's own retry policy instead, which has no overall deadline, so a
// server that answers very slowly can overrun this budget by one read. Holding
// to it exactly would mean giving the read a deadline of its own, which the
// shared client cannot take today.
type WaitOptions struct {
	Timeout time.Duration
	Initial time.Duration
	Max     time.Duration
}

func (o WaitOptions) withDefaults() WaitOptions {
	if o.Timeout <= 0 {
		o.Timeout = DefaultWaitOptions.Timeout
	}
	if o.Initial <= 0 {
		o.Initial = DefaultWaitOptions.Initial
	}
	if o.Max <= 0 {
		o.Max = DefaultWaitOptions.Max
	}
	return o
}

// StillPendingError says the wait expired with the evaluation unfinished. It
// is deliberately not a verdict and carries no result: an evaluation that has
// not answered must never be reported as a denial. The id is on it so the
// caller can point at the evaluation that is still running.
type StillPendingError struct {
	Org    string
	ID     string
	Waited time.Duration
}

func (e *StillPendingError) Error() string {
	return fmt.Sprintf(
		"evaluation %s is still pending after %s; read it later with GET /api/v2/evaluations/%s/%s",
		e.ID, e.Waited, e.Org, e.ID)
}

// Get reads one evaluation. A pending one is a normal answer, not an error.
func (c *Client) Get(org, id string) (*Evaluation, error) {
	endpoint, err := url.JoinPath(c.host, "api/v2/evaluations", org, id)
	if err != nil {
		return nil, err
	}

	// DryRun is deliberately not set: it means do not write, and a read writes
	// nothing. That matches how every other read in this CLI is made.
	response, err := c.http.Do(&requests.RequestParams{
		Method: http.MethodGet,
		URL:    endpoint,
		Token:  c.token,
	})
	if err != nil {
		return nil, err
	}
	if response == nil {
		// Unreachable while the line above asks for a real request, and guarded
		// anyway: the shared client answers a suppressed request with nothing at
		// all, so anyone who later adds one here meets this instead of a panic.
		return nil, errors.New("no evaluation was read back")
	}
	return decodeEvaluation(response.Body)
}

// WaitForTerminal polls until the evaluation finishes, the caller gives up, or
// the budget expires. It returns an evaluation only when that evaluation is
// terminal, so a caller never has to ask whether the answer it holds is final.
//
// Cancelling ctx stops the polling but cannot abort a read already in flight,
// for the same reason the budget cannot: the read carries no context. Both take
// effect at the next turn of the loop.
func (c *Client) WaitForTerminal(ctx context.Context, org, id string, options WaitOptions) (*Evaluation, error) {
	options = options.withDefaults()
	expired := time.After(options.Timeout)
	interval := options.Initial

	for {
		// Checked before the read, so a caller who has already given up costs
		// the server nothing.
		if err := ctx.Err(); err != nil {
			return nil, err
		}

		evaluation, err := c.Get(org, id)
		if err != nil {
			// Transient trouble was already retried inside the HTTP client, so
			// what reaches here will not improve by being asked again.
			return nil, err
		}
		if evaluation.IsTerminal() {
			return evaluation, nil
		}

		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-expired:
			return nil, &StillPendingError{Org: org, ID: id, Waited: options.Timeout}
		case <-time.After(interval):
		}
		interval = nextInterval(interval, options.Max)
	}
}

// nextInterval doubles the wait between reads, never past max. Backing off
// keeps a slow evaluation from being asked about hundreds of times, while the
// first few reads stay close together so a fast one answers promptly.
func nextInterval(current, max time.Duration) time.Duration {
	if doubled := current * 2; doubled < max {
		return doubled
	}
	return max
}
