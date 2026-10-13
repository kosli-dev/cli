package regopolicies

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/kosli-dev/cli/internal/poll"
)

// DefaultWaitOptions allows 30 s: a bundle at the size cap takes about 10 s to
// validate.
var DefaultWaitOptions = WaitOptions{
	Timeout: 30 * time.Second,
	Initial: 500 * time.Millisecond,
	Max:     5 * time.Second,
}

// WaitOptions are poll.Options in which a field that is not set takes its default.
type WaitOptions = poll.Options

// StillPendingError says the wait expired before the version was validated.
type StillPendingError struct {
	Identifier string
	Version    int
	URL        string
	Waited     time.Duration
}

func (e *StillPendingError) Error() string {
	return fmt.Sprintf("version %d of rego policy %s is still pending after %s; read it later with GET %s",
		e.Version, e.Identifier, e.Waited, e.URL)
}

// WaitForTerminal polls until the version is validated, the caller gives up, or
// the budget expires. It returns a version only when that version is terminal.
func (c *Client) WaitForTerminal(ctx context.Context, org, identifier string, version int, options WaitOptions) (*Version, error) {
	endpoint, err := VersionURL(c.host, org, identifier, version)
	if err != nil {
		return nil, err
	}
	read, err := poll.Until(ctx, options.Or(DefaultWaitOptions),
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
			Identifier: identifier, Version: version, URL: endpoint, Waited: expired.Waited}
	}
	return read, err
}
