package kosli

import (
	"context"
	"net/http"
	"net/url"

	"github.com/kosli-dev/cli/internal/requests"
)

// AssertArtifactFilter chooses what an Artifact is asserted against. Environment
// and Policies are mutually exclusive; with neither, the Artifact is asserted
// against the templates of the Flows it is in. An empty Flow means every Flow.
type AssertArtifactFilter struct {
	Flow        string
	Environment string
	Policies    []string
}

// AssertArtifact calls assert_artifact. It returns the assertion of the Artifact
// with fingerprint undecoded, so that a caller gating on compliance decodes only
// that field and is not failed by a change in the rest of the body.
func (c *Client) AssertArtifact(ctx context.Context, fingerprint string, filter AssertArtifactFilter) (*Result, error) {
	endpoint, err := c.endpoint("asserts", c.org, "fingerprint", fingerprint)
	if err != nil {
		return nil, err
	}
	query := url.Values{}
	if filter.Flow != "" {
		query.Set("flow_name", filter.Flow)
	}
	if filter.Environment != "" {
		query.Set("environment_name", filter.Environment)
	}
	for _, policy := range filter.Policies {
		query.Add("policy_name", policy)
	}
	return c.send(ctx, &requests.RequestParams{Method: http.MethodGet, URL: withQuery(endpoint, query)})
}
