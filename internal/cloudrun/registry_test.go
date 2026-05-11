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

func (f *fakeResolver) Resolve(image string) (string, string, error) {
	if f.err != nil {
		return "", "", f.err
	}
	hex, ok := f.digests[image]
	if !ok {
		return "", "", errors.New("not found in fake")
	}
	host, path, _, err := splitTagPinnedImage(image)
	if err != nil {
		return "", "", err
	}
	return host + "/" + path + "@sha256:" + hex, hex, nil
}

func TestResolveTagPinnedDigests_ReplacesEmptyDigest(t *testing.T) {
	tagPinned := "europe-west1-docker.pkg.dev/proj/repo/ghost-job:c85b06b0"
	digests := map[string]string{tagPinned: ""}

	c := &Client{resolver: &fakeResolver{
		digests: map[string]string{tagPinned: "abc123"},
	}}
	c.resolveTagPinnedDigests(digests)

	require.Len(t, digests, 1)
	require.Contains(t, digests, "europe-west1-docker.pkg.dev/proj/repo/ghost-job@sha256:abc123")
	require.Equal(t, "abc123", digests["europe-west1-docker.pkg.dev/proj/repo/ghost-job@sha256:abc123"])
	require.NotContains(t, digests, tagPinned, "tag-pinned key must be removed once resolved")
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

	require.Len(t, digests, 2)
	require.Equal(t, "already", digests[already])
	require.Equal(t, "newhex", digests["europe-west1-docker.pkg.dev/proj/repo/sidecar@sha256:newhex"])
}
