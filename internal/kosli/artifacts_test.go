package kosli

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetArtifactByFingerprintRequestsTheArtifactAndDecodesIt(t *testing.T) {
	server, seen := newFakeServer(t, http.StatusOK, `{"filename":"app","fingerprint":"abc"}`)
	client := newTestClient(t, server.URL)

	artifact, err := client.GetArtifactByFingerprint(context.Background(), "flow-1", "abc", "")

	require.NoError(t, err)
	require.Equal(t, http.MethodGet, seen.method)
	require.Equal(t, "/api/v2/artifacts/test-org/flow-1/fingerprint/abc", seen.rawPath)
	require.Equal(t, "app", artifact.Value.Filename)
	require.JSONEq(t, `{"filename":"app","fingerprint":"abc"}`, string(artifact.Raw))
}

func TestGetArtifactByFingerprintFiltersByTrailOnlyWhenOneIsGiven(t *testing.T) {
	server, seen := newFakeServer(t, http.StatusOK, `{}`)
	client := newTestClient(t, server.URL)

	_, err := client.GetArtifactByFingerprint(context.Background(), "flow-1", "abc", "trail 1")
	require.NoError(t, err)
	require.Equal(t, "trail=trail+1", seen.rawQuery)

	_, err = client.GetArtifactByFingerprint(context.Background(), "flow-1", "abc", "")
	require.NoError(t, err)
	require.Empty(t, seen.rawQuery)
}
