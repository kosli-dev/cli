package cloudrun

import (
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/kosli-dev/cli/internal/requests"
	"github.com/stretchr/testify/require"
	"golang.org/x/oauth2"
)

// fakeRegistry answers HTTP requests by host + escaped path, recording each
// one, so resolver tests can assert the exact URL sent to Google.
type fakeRegistry struct {
	responses map[string]fakeResponse
	requested []string
}

type fakeResponse struct {
	status int
	body   string
	header http.Header
}

func (f *fakeRegistry) RoundTrip(req *http.Request) (*http.Response, error) {
	key := req.URL.Host + req.URL.EscapedPath()
	f.requested = append(f.requested, key)
	res, ok := f.responses[key]
	if !ok {
		res = fakeResponse{status: http.StatusNotFound, body: `{"error":{"code":404}}`}
	}
	header := res.header
	if header == nil {
		header = http.Header{}
	}
	return &http.Response{
		StatusCode: res.status,
		Header:     header,
		Body:       io.NopCloser(strings.NewReader(res.body)),
		Request:    req,
	}, nil
}

func newFakeRegistryClient(t *testing.T, registry *fakeRegistry) *requests.Client {
	return &requests.Client{
		Logger:     newDiscardLogger(t),
		HttpClient: &http.Client{Transport: registry},
	}
}

func staticToken() oauth2.TokenSource {
	return oauth2.StaticTokenSource(&oauth2.Token{AccessToken: "test-token"})
}

const testHex = "0b2fd168ee792a7b0c9d6f48a21f83a2ed6a4eaff9ce6cf49abec4a9e12ad6d7"

func TestLatestTag_EscapesNestedImagePath(t *testing.T) {
	registry := &fakeRegistry{responses: map[string]fakeResponse{
		"artifactregistry.googleapis.com/v1/projects/proj/locations/europe-west1/repositories/repo/dockerImages/parent%2Fchild%2Fimg@sha256:" + testHex: {
			status: http.StatusOK, body: `{"tags":["v1"]}`,
		},
	}}
	r := &gcpArtifactRegistryTagResolver{tokens: staticToken(), client: newFakeRegistryClient(t, registry), log: newDiscardLogger(t)}

	tag, err := r.LatestTag("europe-west1-docker.pkg.dev/proj/repo/parent/child/img@sha256:" + testHex)

	require.NoError(t, err)
	require.Equal(t, "v1", tag)
}
