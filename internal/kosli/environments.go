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

// CreateOrUpdateEnvironmentRequest describes the Environment that
// CreateOrUpdateEnvironment creates or updates.
type CreateOrUpdateEnvironmentRequest struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	Description string `json:"description"`
	// Deprecated: the server accepts but ignores it, since scaling events are
	// no longer recorded. It is sent only for the deprecated scaling flags.
	IncludeScaling *bool `json:"include_scaling,omitempty"`
	// Deprecated: the server accepts but ignores it; use policies instead. It is
	// sent only for the deprecated --require-provenance flag.
	RequireProvenance    *bool    `json:"require_provenance,omitempty"`
	IncludedEnvironments []string `json:"included_environments,omitempty"`
}

// CreateOrUpdateEnvironment calls create_or_update_environment. It creates the
// Environment, or overwrites the one that already has that name.
func (c *Client) CreateOrUpdateEnvironment(ctx context.Context, request CreateOrUpdateEnvironmentRequest) (*Result, error) {
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
