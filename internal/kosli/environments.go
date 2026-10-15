package kosli

import (
	"context"
	"net/http"

	"github.com/kosli-dev/cli/internal/requests"
)

// Environment is an Environment's metadata, with the fields callers read so far.
type Environment struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

// GetEnvironment calls get_environment. It returns the Environment called name.
func (c *Client) GetEnvironment(ctx context.Context, name string) (*Response[Environment], error) {
	endpoint, err := c.endpoint("environments", c.org, name)
	if err != nil {
		return nil, err
	}
	result, err := c.send(ctx, &requests.RequestParams{Method: http.MethodGet, URL: endpoint})
	if err != nil {
		return nil, err
	}
	return &Response[Environment]{Result: *result}, nil
}
