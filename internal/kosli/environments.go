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

// CreateEnvironmentRequest describes the Environment CreateEnvironment creates.
type CreateEnvironmentRequest struct {
	Name                 string   `json:"name"`
	Type                 string   `json:"type"`
	Description          string   `json:"description"`
	IncludeScaling       *bool    `json:"include_scaling,omitempty"`
	RequireProvenance    *bool    `json:"require_provenance,omitempty"`
	IncludedEnvironments []string `json:"included_environments,omitempty"`
}

// CreateEnvironment calls create_or_update_environment. It creates the
// Environment, or updates it when one with that name already exists.
func (c *Client) CreateEnvironment(ctx context.Context, request CreateEnvironmentRequest) (*Result, error) {
	endpoint, err := c.endpoint("environments", c.org)
	if err != nil {
		return nil, err
	}
	return c.send(ctx, &requests.RequestParams{Method: http.MethodPut, URL: endpoint, Payload: request})
}

// ReportSnapshot calls the report operation for envType, such as
// report_k8s_environment or report_ecs_environment. It reports payload as
// the current Snapshot of the Environment called name.
func (c *Client) ReportSnapshot(ctx context.Context, name, envType string, payload any) (*Result, error) {
	endpoint, err := c.endpoint("environments", c.org, name, "report", envType)
	if err != nil {
		return nil, err
	}
	return c.send(ctx, &requests.RequestParams{Method: http.MethodPut, URL: endpoint, Payload: payload})
}
