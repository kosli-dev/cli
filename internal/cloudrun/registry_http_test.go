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

func newTestDigestResolver(t *testing.T, registry *fakeRegistry) *gcpRegistryResolver {
	return &gcpRegistryResolver{tokens: staticToken(), client: newFakeRegistryClient(t, registry), log: newDiscardLogger(t)}
}

func tagResponse(tagName string) fakeResponse {
	versions := strings.Replace(tagName[:strings.LastIndex(tagName, "/tags/")], "artifactregistry.googleapis.com/v1/", "", 1)
	return fakeResponse{
		status: http.StatusOK,
		body:   `{"version":"` + versions + `/versions/sha256:` + testHex + `"}`,
	}
}

func manifestResponse() fakeResponse {
	return fakeResponse{
		status: http.StatusOK,
		body:   `{}`,
		header: http.Header{"Docker-Content-Digest": []string{"sha256:" + testHex}},
	}
}

func TestResolve_ReadsDigestFromArtifactRegistryTag(t *testing.T) {
	cases := []struct {
		name  string
		image string
		tag   string
	}{
		{"pkg.dev", "europe-west1-docker.pkg.dev/proj/repo/img:v1",
			"artifactregistry.googleapis.com/v1/projects/proj/locations/europe-west1/repositories/repo/packages/img/tags/v1"},
		{"pkg.dev nested image", "europe-west1-docker.pkg.dev/proj/repo/parent/child/img:v1",
			"artifactregistry.googleapis.com/v1/projects/proj/locations/europe-west1/repositories/repo/packages/parent%2Fchild%2Fimg/tags/v1"},
		{"gcr.io", "eu.gcr.io/proj/team/img:v1",
			"artifactregistry.googleapis.com/v1/projects/proj/locations/europe/repositories/eu.gcr.io/packages/team%2Fimg/tags/v1"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			registry := &fakeRegistry{responses: map[string]fakeResponse{c.tag: tagResponse(c.tag)}}

			hex, err := newTestDigestResolver(t, registry).Resolve(c.image)

			require.NoError(t, err)
			require.Equal(t, testHex, hex)
			require.Equal(t, []string{c.tag}, registry.requested)
		})
	}
}

func TestResolve_FallsBackToManifestRead(t *testing.T) {
	const image = "europe-west1-docker.pkg.dev/proj/repo/img:v1"
	const tag = "artifactregistry.googleapis.com/v1/projects/proj/locations/europe-west1/repositories/repo/packages/img/tags/v1"
	const manifest = "europe-west1-docker.pkg.dev/v2/proj/repo/img/manifests/v1"

	cases := []struct {
		name        string
		tagResponse fakeResponse
	}{
		{"permission denied", fakeResponse{status: http.StatusForbidden, body: `{"error":{"code":403}}`}},
		{"not a sha256 version", fakeResponse{status: http.StatusOK, body: `{"version":"projects/proj/locations/europe-west1/repositories/repo/packages/img/versions/1.0"}`}},
		{"unparseable body", fakeResponse{status: http.StatusOK, body: `not json`}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			registry := &fakeRegistry{responses: map[string]fakeResponse{
				tag:      c.tagResponse,
				manifest: manifestResponse(),
			}}

			hex, err := newTestDigestResolver(t, registry).Resolve(image)

			require.NoError(t, err)
			require.Equal(t, testHex, hex)
			require.Equal(t, []string{tag, manifest}, registry.requested)
		})
	}
}

func TestResolve_HostOutsideArtifactRegistryAPIReadsManifest(t *testing.T) {
	const manifest = "docker.pkg.dev/v2/proj/repo/img/manifests/v1"
	registry := &fakeRegistry{responses: map[string]fakeResponse{manifest: manifestResponse()}}

	hex, err := newTestDigestResolver(t, registry).Resolve("docker.pkg.dev/proj/repo/img:v1")

	require.NoError(t, err)
	require.Equal(t, testHex, hex)
	require.Equal(t, []string{manifest}, registry.requested)
}

func TestResolve_BothLookupsFail(t *testing.T) {
	registry := &fakeRegistry{}

	_, err := newTestDigestResolver(t, registry).Resolve("europe-west1-docker.pkg.dev/proj/repo/img:v1")

	require.Error(t, err)
	require.Len(t, registry.requested, 2)
}

func TestResolve_NonGCPRegistrySendsNoRequest(t *testing.T) {
	registry := &fakeRegistry{}

	_, err := newTestDigestResolver(t, registry).Resolve("docker.io/library/nginx:1.27")

	require.ErrorIs(t, err, errNonGCPRegistry)
	require.Empty(t, registry.requested)
}
