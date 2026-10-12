package evaluations

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/kosli-dev/cli/internal/poll"
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

// WaitOptions are poll.Options in which a field that is not set takes its default.
type WaitOptions = poll.Options

// StillPendingError says the wait expired with the evaluation unfinished. It
// is deliberately not a verdict and carries no result: an evaluation that has
// not answered must never be reported as a denial. The id is on it so the
// caller can point at the evaluation that is still running.
type StillPendingError struct {
	Host   string
	Org    string
	ID     string
	Waited time.Duration
}

func (e *StillPendingError) Error() string {
	pending := fmt.Sprintf("evaluation %s is still pending after %s", e.ID, e.Waited)
	endpoint, err := evaluationURL(e.Host, e.Org, e.ID)
	if err != nil {
		// Unreachable: the wait reads the evaluation from the same URL first.
		return pending
	}
	return pending + "; read it later with GET " + endpoint
}

// Get reads one evaluation. A pending one is a normal answer, not an error.
func (c *Client) Get(org, id string) (*Evaluation, error) {
	endpoint, err := evaluationURL(c.host, org, id)
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
func (c *Client) WaitForTerminal(ctx context.Context, org, id string, options WaitOptions) (*Evaluation, error) {
	evaluation, err := poll.Until(ctx, options.Or(DefaultWaitOptions),
		func() (*Evaluation, bool, error) {
			evaluation, err := c.Get(org, id)
			if err != nil {
				return nil, false, err
			}
			return evaluation, evaluation.IsTerminal(), nil
		})

	var expired *poll.ExpiredError
	if errors.As(err, &expired) {
		return nil, &StillPendingError{Host: c.host, Org: org, ID: id, Waited: expired.Waited}
	}
	return evaluation, err
}

func evaluationURL(host, org, id string) (string, error) {
	return url.JoinPath(host, "api/v2/evaluations", org, id)
}
