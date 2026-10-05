package kosli

import (
	"context"
	"net/http"

	"github.com/kosli-dev/cli/internal/requests"
)

// GetTrail returns the Trail called name in flow, undecoded: its callers
// either print the body or reshape it as arbitrary JSON.
func (c *Client) GetTrail(ctx context.Context, flow, name string) (*Result, error) {
	endpoint, err := c.endpoint("trails", c.org, flow, name)
	if err != nil {
		return nil, err
	}
	return c.send(ctx, &requests.RequestParams{Method: http.MethodGet, URL: endpoint})
}
