package cloudrun

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"sort"
	"strings"

	"github.com/kosli-dev/cli/internal/digest"
	"github.com/kosli-dev/cli/internal/logger"
	"github.com/kosli-dev/cli/internal/requests"
	"golang.org/x/oauth2"
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
// The HTTP client is built once in cloudrun.New() and shared across all
// Resolve calls in a single snapshot — keeping the connection pool warm
// so TCP/TLS handshakes amortize across the (potentially hundreds of)
// artifacts in a report.
type gcpRegistryResolver struct {
	tokens oauth2.TokenSource
	client *requests.Client
	log    *logger.Logger
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
	hex, err := digest.RemoteDockerImageSha256(r.client, path, tag, "https://"+host+"/v2", tok.AccessToken)
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

// tagResolver resolves a digest-pinned image reference (e.g.
// "europe-west1-docker.pkg.dev/proj/repo/img@sha256:<hex>") to a tag
// pointing at that digest, by querying Artifact Registry's
// dockerImages.list endpoint. Used by `kosli snapshot cloud-run
// --resolve-names` to convert the artifact's display name from the
// opaque digest to a human-readable tag (typically a commit SHA).
type tagResolver interface {
	// LatestTag returns one tag pointing at the digest. When the digest
	// has multiple tags, the longest one wins (commit SHAs beat moving
	// tags like ":released"), with lex order as tiebreak. Returns an
	// error if no tags exist for the digest or the API call fails.
	LatestTag(digestPinnedImage string) (tag string, err error)
}

// gcpArtifactRegistryTagResolver implements tagResolver against
// Artifact Registry (*-docker.pkg.dev). The legacy Container Registry
// hosts (gcr.io family) do not expose an equivalent reverse-lookup
// API, so they are rejected with errNonGCPRegistry. The HTTP client
// is shared with the forward (gcpRegistryResolver) path so a single
// connection pool serves every per-artifact API call in the snapshot.
type gcpArtifactRegistryTagResolver struct {
	tokens oauth2.TokenSource
	client *requests.Client
	log    *logger.Logger
}

func (r *gcpArtifactRegistryTagResolver) LatestTag(image string) (string, error) {
	parts, err := splitDigestPinnedAR(image)
	if err != nil {
		return "", err
	}
	tok, err := r.tokens.Token()
	if err != nil {
		return "", fmt.Errorf("getting GCP access token: %w", err)
	}

	// dockerImages.get returns a single DockerImage by exact resource name.
	// The resource name is "<image>@sha256:<hex>" within the repo; "@" and
	// ":" are allowed in URL path segments per RFC 3986 and Google's
	// gateway accepts them unescaped.
	u := fmt.Sprintf(
		"https://artifactregistry.googleapis.com/v1/projects/%s/locations/%s/repositories/%s/dockerImages/%s@%s",
		parts.project, parts.location, parts.repo, parts.image, parts.digest,
	)

	res, err := r.client.Do(&requests.RequestParams{
		Method: http.MethodGet,
		URL:    u,
		Token:  tok.AccessToken,
	})
	if err != nil {
		return "", err
	}

	var resp struct {
		Tags []string `json:"tags"`
	}
	if err := json.Unmarshal([]byte(res.Body), &resp); err != nil {
		return "", fmt.Errorf("parsing AR get response: %w", err)
	}
	return pickLatestTag(resp.Tags)
}

// pickLatestTag selects a tag from those pointing at a single digest.
// AR exposes no per-tag timestamp, so "latest" can't be time-based.
// Rule: longest tag wins (commit SHAs are 40 hex chars; moving aliases
// like ":released", ":latest", ":dev" are shorter). Lex order breaks
// ties for deterministic output.
func pickLatestTag(tags []string) (string, error) {
	if len(tags) == 0 {
		return "", errors.New("digest has no tags")
	}
	if len(tags) == 1 {
		return tags[0], nil
	}
	sorted := append([]string(nil), tags...)
	sort.Slice(sorted, func(i, j int) bool {
		if len(sorted[i]) != len(sorted[j]) {
			return len(sorted[i]) > len(sorted[j])
		}
		return sorted[i] < sorted[j]
	})
	return sorted[0], nil
}

// digestPinnedRefParts holds the AR-resource-name components extracted
// from a digest-pinned image reference, ready to interpolate into the
// dockerImages.get resource URL.
type digestPinnedRefParts struct {
	project  string
	location string
	repo     string
	image    string
	digest   string // "sha256:<hex>"
}

// splitDigestPinnedAR parses a digest-pinned Artifact Registry image
// reference into its constituent parts. Rejects non-AR hosts (including
// gcr.io, since the legacy GCR API does not expose tag reverse-lookup),
// non-digest-pinned references, and references with too few path
// segments.
func splitDigestPinnedAR(image string) (digestPinnedRefParts, error) {
	idx := strings.Index(image, sha256Marker)
	if idx < 0 {
		return digestPinnedRefParts{}, fmt.Errorf("image %q is not digest-pinned", image)
	}
	ref := image[:idx]
	digestPart := image[idx+1:] // strip "@" → "sha256:<hex>"

	parts := strings.SplitN(ref, "/", 4)
	if len(parts) < 4 {
		return digestPinnedRefParts{}, fmt.Errorf("image %q has too few path segments for AR", image)
	}
	host, project, repo, imagePath := parts[0], parts[1], parts[2], parts[3]

	location := strings.TrimSuffix(host, "-docker.pkg.dev")
	if location == host {
		return digestPinnedRefParts{}, fmt.Errorf("host %q is not Artifact Registry (tag resolution unsupported)", host)
	}
	return digestPinnedRefParts{
		project:  project,
		location: location,
		repo:     repo,
		image:    imagePath,
		digest:   digestPart,
	}, nil
}
