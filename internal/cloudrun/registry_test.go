package cloudrun

import (
	"errors"
	"io"
	"testing"

	"github.com/kosli-dev/cli/internal/logger"
	"github.com/stretchr/testify/require"
)

// newDiscardLogger returns a logger that drops all output. Used by tests
// that exercise the resolver's debug-level failure path without polluting
// the test output stream.
func newDiscardLogger(_ *testing.T) *logger.Logger {
	return logger.NewLogger(io.Discard, io.Discard, false)
}

func TestSplitTagPinnedImage_ArtifactRegistry(t *testing.T) {
	host, path, tag, err := splitTagPinnedImage(
		"europe-west1-docker.pkg.dev/proj/repo/img:c85b06b09190b24f8d14aae380ca0d79ab008fc5",
	)
	require.NoError(t, err)
	require.Equal(t, "europe-west1-docker.pkg.dev", host)
	require.Equal(t, "proj/repo/img", path)
	require.Equal(t, "c85b06b09190b24f8d14aae380ca0d79ab008fc5", tag)
}

func TestSplitTagPinnedImage_GCR(t *testing.T) {
	host, path, tag, err := splitTagPinnedImage("gcr.io/proj/img:v1")
	require.NoError(t, err)
	require.Equal(t, "gcr.io", host)
	require.Equal(t, "proj/img", path)
	require.Equal(t, "v1", tag)
}

func TestSplitTagPinnedImage_DigestPinnedRejected(t *testing.T) {
	_, _, _, err := splitTagPinnedImage("gcr.io/proj/img@sha256:abc123")
	require.Error(t, err)
	require.Contains(t, err.Error(), "already digest-pinned")
}

func TestSplitTagPinnedImage_NoTagRejected(t *testing.T) {
	_, _, _, err := splitTagPinnedImage("gcr.io/proj/img")
	require.Error(t, err)
	require.Contains(t, err.Error(), "not tag-pinned")
}

func TestSplitTagPinnedImage_PortInHostNotMistakenForTag(t *testing.T) {
	// localhost:5000/foo has a colon that is a port separator, not a tag
	// separator. The function should refuse rather than treat "5000/foo"
	// as a tag.
	_, _, _, err := splitTagPinnedImage("localhost:5000/foo")
	require.Error(t, err)
	require.Contains(t, err.Error(), "not tag-pinned")
}

func TestSplitTagPinnedImage_PortAndTagBothPresent(t *testing.T) {
	host, path, tag, err := splitTagPinnedImage("localhost:5000/foo:v1")
	require.NoError(t, err)
	require.Equal(t, "localhost:5000", host)
	require.Equal(t, "foo", path)
	require.Equal(t, "v1", tag)
}

func TestSplitTagPinnedImage_EmptyTagRejected(t *testing.T) {
	_, _, _, err := splitTagPinnedImage("gcr.io/proj/img:")
	require.Error(t, err)
	require.Contains(t, err.Error(), "empty tag")
}

func TestSplitTagPinnedImage_NoHostRejected(t *testing.T) {
	_, _, _, err := splitTagPinnedImage("img:v1")
	require.Error(t, err)
	require.Contains(t, err.Error(), "no host segment")
}

func TestIsGCPRegistryHost(t *testing.T) {
	cases := []struct {
		host string
		want bool
	}{
		{"europe-west1-docker.pkg.dev", true},
		{"us-central1-docker.pkg.dev", true},
		{"docker.pkg.dev", true},
		{"gcr.io", true},
		{"eu.gcr.io", true},
		{"us.gcr.io", true},
		{"asia.gcr.io", true},
		{"docker.io", false},
		{"index.docker.io", false},
		{"quay.io", false},
		{"localhost:5000", false},
		{"123456789.dkr.ecr.us-east-1.amazonaws.com", false},
	}
	for _, c := range cases {
		t.Run(c.host, func(t *testing.T) {
			require.Equal(t, c.want, isGCPRegistryHost(c.host))
		})
	}
}

// fakeResolver is a digestResolver test double. It looks up the resolved
// digest hex by exact-match on the input image string.
type fakeResolver struct {
	digests map[string]string // image:tag -> resolvedHex
	err     error
}

func (f *fakeResolver) Resolve(image string) (string, error) {
	if f.err != nil {
		return "", f.err
	}
	hex, ok := f.digests[image]
	if !ok {
		return "", errors.New("not found in fake")
	}
	return hex, nil
}

func TestResolveTagPinnedDigests_FillsEmptyDigestKeepingTagPinnedKey(t *testing.T) {
	tagPinned := "europe-west1-docker.pkg.dev/proj/repo/ghost-job:c85b06b0"
	digests := map[string]string{tagPinned: ""}

	c := &Client{resolver: &fakeResolver{
		digests: map[string]string{tagPinned: "abc123"},
	}}
	c.resolveTagPinnedDigests(digests)

	require.Equal(t, map[string]string{tagPinned: "abc123"}, digests,
		"tag-pinned key must be preserved; only the digest value is filled in")
}

func TestResolveTagPinnedDigests_LeavesAlreadyResolvedAlone(t *testing.T) {
	digestPinned := "europe-west1-docker.pkg.dev/proj/repo/img@sha256:abc123"
	digests := map[string]string{digestPinned: "abc123"}

	c := &Client{resolver: &fakeResolver{}} // empty fake; would error on any call
	c.resolveTagPinnedDigests(digests)

	require.Equal(t, map[string]string{digestPinned: "abc123"}, digests)
}

func TestResolveTagPinnedDigests_FailureLeavesEntryInPlace(t *testing.T) {
	tagPinned := "europe-west1-docker.pkg.dev/proj/repo/img:v1"
	digests := map[string]string{tagPinned: ""}

	c := &Client{
		resolver: &fakeResolver{err: errors.New("registry unreachable")},
		log:      newDiscardLogger(t),
	}
	c.resolveTagPinnedDigests(digests)

	require.Equal(t, map[string]string{tagPinned: ""}, digests,
		"failed resolution must leave the original tag-pinned entry untouched")
}

func TestResolveTagPinnedDigests_NilResolverIsNoOp(t *testing.T) {
	tagPinned := "europe-west1-docker.pkg.dev/proj/repo/img:v1"
	digests := map[string]string{tagPinned: ""}

	c := &Client{} // no resolver
	c.resolveTagPinnedDigests(digests)

	require.Equal(t, map[string]string{tagPinned: ""}, digests)
}

// --- splitDigestPinnedAR ---

func TestSplitDigestPinnedAR_HappyPath(t *testing.T) {
	parts, err := splitDigestPinnedAR(
		"europe-west1-docker.pkg.dev/hello-world-cli-demo/containers/hello-world@sha256:0b2fd168ee792a7b0c9d6f48a21f83a2ed6a4eaff9ce6cf49abec4a9e12ad6d7",
	)
	require.NoError(t, err)
	require.Equal(t, "hello-world-cli-demo", parts.project)
	require.Equal(t, "europe-west1", parts.location)
	require.Equal(t, "containers", parts.repo)
	require.Equal(t, "hello-world", parts.image)
	require.Equal(t, "sha256:0b2fd168ee792a7b0c9d6f48a21f83a2ed6a4eaff9ce6cf49abec4a9e12ad6d7", parts.digest)
}

func TestSplitDigestPinnedAR_NestedImagePath(t *testing.T) {
	parts, err := splitDigestPinnedAR(
		"europe-west1-docker.pkg.dev/proj/repo/parent/child/img@sha256:abc",
	)
	require.NoError(t, err)
	require.Equal(t, "parent/child/img", parts.image)
}

func TestSplitDigestPinnedAR_TagPinnedRejected(t *testing.T) {
	_, err := splitDigestPinnedAR("europe-west1-docker.pkg.dev/proj/repo/img:v1")
	require.Error(t, err)
	require.Contains(t, err.Error(), "not digest-pinned")
}

func TestSplitDigestPinnedAR_NonARHostRejected(t *testing.T) {
	// 4 path segments so the parser passes the segment-count check and
	// reaches the host check. Refs from GCR or other registries do not
	// share AR's region-prefixed pkg.dev host pattern.
	_, err := splitDigestPinnedAR("eu.gcr.io/proj/repo/img@sha256:abc")
	require.Error(t, err)
	require.Contains(t, err.Error(), "not Artifact Registry")
}

func TestSplitDigestPinnedAR_TooFewSegmentsRejected(t *testing.T) {
	_, err := splitDigestPinnedAR("europe-west1-docker.pkg.dev/proj@sha256:abc")
	require.Error(t, err)
	require.Contains(t, err.Error(), "too few path segments")
}

// --- pickLatestTag ---

func TestPickLatestTag_SingleTag(t *testing.T) {
	got, err := pickLatestTag([]string{"d37cf45b250a17eb71ed885785d1fc3f05200e71"})
	require.NoError(t, err)
	require.Equal(t, "d37cf45b250a17eb71ed885785d1fc3f05200e71", got)
}

func TestPickLatestTag_LongestWins(t *testing.T) {
	// commit-SHA + moving tag: the longer (commit SHA) wins, since it's
	// more useful as an artifact identifier than ":released" or ":latest".
	got, err := pickLatestTag([]string{"released", "d37cf45b250a17eb71ed885785d1fc3f05200e71"})
	require.NoError(t, err)
	require.Equal(t, "d37cf45b250a17eb71ed885785d1fc3f05200e71", got)
}

func TestPickLatestTag_LexBreaksTie(t *testing.T) {
	// Two tags of equal length — fall back to lex order for determinism.
	got, err := pickLatestTag([]string{"v1.0.0-beta", "v1.0.0-alpha"})
	require.NoError(t, err)
	require.Equal(t, "v1.0.0-alpha", got)
}

func TestPickLatestTag_EmptyErrors(t *testing.T) {
	_, err := pickLatestTag(nil)
	require.Error(t, err)
}

// --- resolveNamesForDigestPinned ---

type fakeTagResolver struct {
	tags map[string]string // digest-pinned image -> resolved tag
	err  error
}

func (f *fakeTagResolver) LatestTag(image string) (string, error) {
	if f.err != nil {
		return "", f.err
	}
	tag, ok := f.tags[image]
	if !ok {
		return "", errors.New("not found in fake")
	}
	return tag, nil
}

func TestResolveNamesForDigestPinned_RewritesDigestToTag(t *testing.T) {
	digestPinned := "europe-west1-docker.pkg.dev/proj/repo/img@sha256:abc"
	digests := map[string]string{digestPinned: "abc"}

	c := &Client{tagResolver: &fakeTagResolver{
		tags: map[string]string{digestPinned: "v1.0.0"},
	}}
	c.resolveNamesForDigestPinned(digests)

	require.Equal(t, map[string]string{
		"europe-west1-docker.pkg.dev/proj/repo/img:v1.0.0": "abc",
	}, digests, "digest-pinned key must be replaced by tag-pinned key; hex value preserved")
}

func TestResolveNamesForDigestPinned_LeavesTagPinnedAlone(t *testing.T) {
	tagPinned := "europe-west1-docker.pkg.dev/proj/repo/img:v1"
	digests := map[string]string{tagPinned: "abc"}

	c := &Client{tagResolver: &fakeTagResolver{}} // empty fake; would error on any call
	c.resolveNamesForDigestPinned(digests)

	require.Equal(t, map[string]string{tagPinned: "abc"}, digests)
}

func TestResolveNamesForDigestPinned_FailureLeavesEntryInPlace(t *testing.T) {
	digestPinned := "europe-west1-docker.pkg.dev/proj/repo/img@sha256:abc"
	digests := map[string]string{digestPinned: "abc"}

	c := &Client{
		tagResolver: &fakeTagResolver{err: errors.New("api down")},
		log:         newDiscardLogger(t),
	}
	c.resolveNamesForDigestPinned(digests)

	require.Equal(t, map[string]string{digestPinned: "abc"}, digests)
}

func TestResolveNamesForDigestPinned_NilResolverIsNoOp(t *testing.T) {
	digestPinned := "europe-west1-docker.pkg.dev/proj/repo/img@sha256:abc"
	digests := map[string]string{digestPinned: "abc"}

	c := &Client{} // no tagResolver
	c.resolveNamesForDigestPinned(digests)

	require.Equal(t, map[string]string{digestPinned: "abc"}, digests)
}

func TestResolveTagPinnedDigests_MixedMap(t *testing.T) {
	already := "europe-west1-docker.pkg.dev/proj/repo/main@sha256:already"
	tagPinned := "europe-west1-docker.pkg.dev/proj/repo/sidecar:v1"
	digests := map[string]string{
		already:   "already",
		tagPinned: "",
	}

	c := &Client{resolver: &fakeResolver{
		digests: map[string]string{tagPinned: "newhex"},
	}}
	c.resolveTagPinnedDigests(digests)

	require.Equal(t, map[string]string{
		already:   "already",
		tagPinned: "newhex",
	}, digests, "both keys must be preserved; only the tag-pinned value is filled in")
}
