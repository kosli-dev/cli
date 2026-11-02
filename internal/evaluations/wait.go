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
	ID     string
	Waited time.Duration
}

func (e *StillPendingError) Error() string {
	return fmt.Sprintf("evaluation %s is still pending after %s; read it later with 'kosli get evaluation %s'",
		e.ID, e.Waited, e.ID)
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
		return nil, &StillPendingError{ID: id, Waited: expired.Waited}
	}
	return evaluation, err
}

// evaluationURL escapes the id into one path segment, so an id holding a slash
// cannot name another route. It refuses "." and "..", which a URL path
// resolves however they are escaped.
func evaluationURL(host, org, id string) (string, error) {
	switch id {
	case "", ".", "..":
		return "", fmt.Errorf("%q is not a valid evaluation id", id)
	}
	return url.JoinPath(host, "api/v2/evaluations/"+url.PathEscape(org)+"/"+url.PathEscape(id))
}

// Input reads the document the evaluation's policy reads as input, as the
// server sent it. Its 409 says the input cannot be built, which no retry
// changes.
func (c *Client) Input(org, id string) (string, error) {
	endpoint, err := evaluationURL(c.host, org, id)
	if err != nil {
		return "", err
	}
	response, err := c.http.Do(&requests.RequestParams{
		Method:               http.MethodGet,
		URL:                  endpoint + "/input",
		Token:                c.token,
		DisableConflictRetry: true,
	})
	if err != nil {
		return "", err
	}
	return response.Body, nil
}
