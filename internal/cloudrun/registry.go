package cloudrun

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/kosli-dev/cli/internal/digest"
	"github.com/kosli-dev/cli/internal/logger"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

// digestResolver resolves a tag-pinned image reference (e.g.
// "europe-west1-docker.pkg.dev/proj/repo/img:v1") to its sha256 digest by
// querying the OCI Distribution endpoint of the hosting registry. Errors
// are non-fatal at the call site: a failed resolution leaves the original
// tag-pinned reference in place with an empty digest, mirroring the
// existing fallback for tag-pinned ECS / Service container images.
type digestResolver interface {
	// Resolve returns the bare hex digest for the given tag-pinned image
	// reference, or an error. Callers keep the original image string as
	// the artifact's identifier and only use the returned hex as the
	// digest value.
	Resolve(image string) (hexDigest string, err error)
}

// errNonGCPRegistry is returned by gcpRegistryResolver when the image's
// host is not a known GCP registry. We don't attempt to authenticate
// against unknown hosts because the GCP OAuth token only works against
// Artifact Registry and Container Registry.
var errNonGCPRegistry = errors.New("not a GCP registry host — skipping resolution")

// gcpRegistryResolver implements digestResolver for Artifact Registry
// (*-docker.pkg.dev) and the legacy Container Registry (gcr.io family),
// using a GCP OAuth access token from Application Default Credentials.
type gcpRegistryResolver struct {
	tokens oauth2.TokenSource
	log    *logger.Logger
}

func newGCPRegistryResolver(ctx context.Context, log *logger.Logger) (digestResolver, error) {
	src, err := google.DefaultTokenSource(ctx, "https://www.googleapis.com/auth/cloud-platform")
	if err != nil {
		return nil, fmt.Errorf("getting GCP token source: %w", err)
	}
	return &gcpRegistryResolver{tokens: src, log: log}, nil
}

func (r *gcpRegistryResolver) Resolve(image string) (string, error) {
	host, path, tag, err := splitTagPinnedImage(image)
	if err != nil {
		return "", err
	}
	if !isGCPRegistryHost(host) {
		return "", errNonGCPRegistry
	}
	tok, err := r.tokens.Token()
	if err != nil {
		return "", fmt.Errorf("getting GCP access token: %w", err)
	}
	hex, err := digest.RemoteDockerImageSha256(path, tag, "https://"+host+"/v2", tok.AccessToken, r.log)
	if err != nil {
		return "", err
	}
	if hex == "" {
		return "", fmt.Errorf("registry returned empty digest for %q", image)
	}
	return hex, nil
}

// splitTagPinnedImage parses a tag-pinned image reference of the form
//
//	host/path/segments...:tag
//
// returning (host, path, tag). It rejects digest-pinned references
// (containing "@sha256:") and references without a host segment.
func splitTagPinnedImage(image string) (host, path, tag string, err error) {
	if strings.Contains(image, sha256Marker) {
		return "", "", "", fmt.Errorf("image %q is already digest-pinned", image)
	}
	// The tag is everything after the last ":", which must come after the
	// last "/" (otherwise the colon is a port separator like in
	// "localhost:5000/foo").
	lastSlash := strings.LastIndex(image, "/")
	lastColon := strings.LastIndex(image, ":")
	if lastColon < 0 || lastColon < lastSlash {
		return "", "", "", fmt.Errorf("image %q is not tag-pinned", image)
	}
	ref := image[:lastColon]
	tag = image[lastColon+1:]
	firstSlash := strings.Index(ref, "/")
	if firstSlash < 0 {
		return "", "", "", fmt.Errorf("image %q has no host segment", image)
	}
	host = ref[:firstSlash]
	path = ref[firstSlash+1:]
	if tag == "" {
		return "", "", "", fmt.Errorf("image %q has empty tag", image)
	}
	return host, path, tag, nil
}

// isGCPRegistryHost reports whether the host is a known GCP-hosted
// registry that accepts a cloud-platform OAuth bearer token. Both
// Artifact Registry (*-docker.pkg.dev) and the legacy Container
// Registry (gcr.io family) are accepted.
func isGCPRegistryHost(host string) bool {
	if host == "docker.pkg.dev" || strings.HasSuffix(host, "-docker.pkg.dev") {
		return true
	}
	if host == "gcr.io" || strings.HasSuffix(host, ".gcr.io") {
		return true
	}
	return false
}
