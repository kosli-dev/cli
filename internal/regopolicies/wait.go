package regopolicies

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/kosli-dev/cli/internal/poll"
)

// DefaultWaitOptions gives validation the same thirty seconds the platform
// states for an evaluation. A bundle at the size cap takes about ten seconds to
// compile, so this is ample without hiding a stuck version.
var DefaultWaitOptions = WaitOptions{
	Timeout: 30 * time.Second,
	Initial: 500 * time.Millisecond,
	Max:     5 * time.Second,
}

// WaitOptions bounds how long a version is asked about and how often, with the
// same meaning as poll.Options. A zero field takes its default.
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

// StillPendingError says the wait expired before the version was validated. It
// carries no version, so an unfinished validation is never read as an outcome.
type StillPendingError struct {
	Org        string
	Identifier string
	Version    int
	Waited     time.Duration
}

func (e *StillPendingError) Error() string {
	return fmt.Sprintf(
		"version %d of rego policy %s is still pending after %s; read it later with GET /api/v2/rego-policies/%s/%s/versions/%d",
		e.Version, e.Identifier, e.Waited, e.Org, e.Identifier, e.Version)
}

// WaitForTerminal polls until the version is validated, the caller gives up, or
// the budget expires. It returns a version only when that version is terminal.
func (c *Client) WaitForTerminal(ctx context.Context, org, identifier string, version int, options WaitOptions) (*Version, error) {
	read, err := poll.Until(ctx, poll.Options(options.withDefaults()),
		func() (*Version, bool, error) {
			read, err := c.GetVersion(org, identifier, version)
			if err != nil {
				return nil, false, err
			}
			return read, read.IsTerminal(), nil
		})

	var expired *poll.ExpiredError
	if errors.As(err, &expired) {
		return nil, &StillPendingError{
			Org: org, Identifier: identifier, Version: version, Waited: expired.Waited}
	}
	return read, err
}
