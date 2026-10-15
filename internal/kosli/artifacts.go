package kosli

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"

	"github.com/kosli-dev/cli/internal/requests"
)

// Artifact is an Artifact as a Flow records it.
type Artifact struct {
	Filename              string           `json:"filename"`
	FlowName              string           `json:"flow_name"`
	TrailName             *string          `json:"trail_name"`
	TemplateReferenceName *string          `json:"template_reference_name"`
	Fingerprint           string           `json:"fingerprint"`
	CreatedAt             json.Number      `json:"created_at"`
	GitCommit             string           `json:"git_commit"`
	CommitURL             string           `json:"commit_url"`
	BuildURL              string           `json:"build_url"`
	HTMLURL               string           `json:"html_url"`
	State                 string           `json:"state"`
	Running               []EnvironmentRef `json:"running"`
	Exited                []EnvironmentRef `json:"exited"`
	History               []ArtifactEvent  `json:"history"`
}

// EnvironmentRef names the Snapshot of an Environment an Artifact ran in.
type EnvironmentRef struct {
	EnvironmentName string  `json:"environment_name"`
	SnapshotIndex   float64 `json:"snapshot_index"`
}

// ArtifactEvent is one entry in an Artifact's history.
type ArtifactEvent struct {
	Event     string      `json:"event"`
	Timestamp json.Number `json:"timestamp"`
}

// GetArtifactByFingerprint calls get_artifact_by_fingerprint. It returns the
// Artifact with fingerprint in flow, the latest in trail when trail is not empty.
func (c *Client) GetArtifactByFingerprint(ctx context.Context, flow, fingerprint, trail string) (*Response[Artifact], error) {
	endpoint, err := c.endpoint("artifacts", c.org, flow, "fingerprint", fingerprint)
	if err != nil {
		return nil, err
	}
	result, err := c.send(ctx, &requests.RequestParams{Method: http.MethodGet, URL: withQuery(endpoint, trailQuery(trail))})
	if err != nil {
		return nil, err
	}
	return decode[Artifact](result)
}

// ListArtifactsByCommitSHA calls list_artifacts_by_commit_sha. It returns the
// Artifacts built from commitSHA in flow, only those in trail when trail is not
// empty.
func (c *Client) ListArtifactsByCommitSHA(ctx context.Context, flow, commitSHA, trail string) (*Response[[]Artifact], error) {
	endpoint, err := c.endpoint("artifacts", c.org, flow, "commit_sha", commitSHA)
	if err != nil {
		return nil, err
	}
	result, err := c.send(ctx, &requests.RequestParams{Method: http.MethodGet, URL: withQuery(endpoint, trailQuery(trail))})
	if err != nil {
		return nil, err
	}
	return decode[[]Artifact](result)
}

func trailQuery(trail string) url.Values {
	query := url.Values{}
	if trail != "" {
		query.Set("trail", trail)
	}
	return query
}
