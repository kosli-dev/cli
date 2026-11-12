// Package registrytest runs an in-memory OCI registry for tests, and attaches
// attestations to images in it the way the tools that make them do.
package registrytest

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/go-containerregistry/pkg/registry"
	godigest "github.com/opencontainers/go-digest"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/stretchr/testify/require"
)

// Registry is an OCI registry served over plain HTTP on a loopback address.
type Registry struct {
	// Host is the registry's host and port, as an image reference names it.
	Host string

	t            testing.TB
	server       *httptest.Server
	referrersAPI bool
	username     string
	password     string
	corrupted    map[string]bool
}

// Option changes how a Registry behaves.
type Option func(*Registry)

// WithoutReferrersAPI makes the registry answer 404 to the Referrers API, as
// registries older than OCI 1.1 do, so that clients list an image's referrers
// under its fallback tag instead.
func WithoutReferrersAPI() Option {
	return func(r *Registry) { r.referrersAPI = false }
}

// RequiringCredentials makes the registry refuse any request that does not
// present this username and password.
func RequiringCredentials(username, password string) Option {
	return func(r *Registry) { r.username, r.password = username, password }
}

// New starts a registry that stops when the test ends.
func New(t testing.TB, opts ...Option) *Registry {
	t.Helper()
	r := &Registry{t: t, referrersAPI: true, corrupted: map[string]bool{}}
	for _, opt := range opts {
		opt(r)
	}
	handler := registry.New(registry.WithReferrersSupport(r.referrersAPI), registry.Logger(log.New(io.Discard, "", 0)))
	r.server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if r.username != "" {
			username, password, ok := req.BasicAuth()
			if !ok || username != r.username || password != r.password {
				w.Header().Set("WWW-Authenticate", `Basic realm="registrytest"`)
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
		}
		if req.Method == http.MethodGet && r.corrupted[req.URL.Path] {
			_, _ = w.Write([]byte(`{"corrupted": true}`))
			return
		}
		handler.ServeHTTP(w, req)
	}))
	t.Cleanup(r.server.Close)
	r.Host = strings.TrimPrefix(r.server.URL, "http://")
	return r
}

// PushBlob stores content in repo and returns its descriptor.
func (r *Registry) PushBlob(repo, mediaType string, content []byte) ocispec.Descriptor {
	r.t.Helper()
	descriptor := Descriptor(mediaType, content)
	r.do(http.MethodPost, fmt.Sprintf("/v2/%s/blobs/uploads/?digest=%s", repo, descriptor.Digest), "application/octet-stream", content, http.StatusCreated)
	return descriptor
}

// PushManifest stores manifest in repo under reference, a tag or the
// manifest's own digest, and returns its descriptor.
func (r *Registry) PushManifest(repo, reference, mediaType string, manifest []byte) ocispec.Descriptor {
	r.t.Helper()
	r.do(http.MethodPut, fmt.Sprintf("/v2/%s/manifests/%s", repo, reference), mediaType, manifest, http.StatusCreated)
	return Descriptor(mediaType, manifest)
}

// CorruptManifest makes the registry answer requests for reference in repo
// with content that is not the manifest stored there.
func (r *Registry) CorruptManifest(repo, reference string) {
	r.corrupted[fmt.Sprintf("/v2/%s/manifests/%s", repo, reference)] = true
}

// CorruptBlob makes the registry answer requests for the blob with digest in
// repo with content that does not hash to it.
func (r *Registry) CorruptBlob(repo, digest string) {
	r.corrupted[fmt.Sprintf("/v2/%s/blobs/%s", repo, digest)] = true
}

// Descriptor describes content as a manifest refers to it.
func Descriptor(mediaType string, content []byte) ocispec.Descriptor {
	return ocispec.Descriptor{
		MediaType: mediaType,
		Digest:    Digest(content),
		Size:      int64(len(content)),
	}
}

// Digest is the digest a registry addresses content by.
func Digest(content []byte) godigest.Digest {
	return godigest.FromBytes(content)
}

func (r *Registry) manifest(repo, reference string) ([]byte, bool) {
	r.t.Helper()
	response := r.request(http.MethodGet, fmt.Sprintf("/v2/%s/manifests/%s", repo, reference), "", nil)
	defer func() { _ = response.Body.Close() }()
	if response.StatusCode == http.StatusNotFound {
		return nil, false
	}
	require.Equal(r.t, http.StatusOK, response.StatusCode)
	content, err := io.ReadAll(response.Body)
	require.NoError(r.t, err)
	return content, true
}

func (r *Registry) do(method, path, contentType string, body []byte, wantStatus int) {
	r.t.Helper()
	response := r.request(method, path, contentType, body)
	defer func() { _ = response.Body.Close() }()
	answer, _ := io.ReadAll(response.Body)
	require.Equal(r.t, wantStatus, response.StatusCode, "%s %s: %s", method, path, answer)
}

func (r *Registry) request(method, path, contentType string, body []byte) *http.Response {
	r.t.Helper()
	request, err := http.NewRequest(method, r.server.URL+path, bytes.NewReader(body))
	require.NoError(r.t, err)
	if contentType != "" {
		request.Header.Set("Content-Type", contentType)
	}
	if r.username != "" {
		request.SetBasicAuth(r.username, r.password)
	}
	response, err := http.DefaultClient.Do(request)
	require.NoError(r.t, err)
	return response
}

func marshal(t testing.TB, value any) []byte {
	t.Helper()
	content, err := json.Marshal(value)
	require.NoError(t, err)
	return content
}
