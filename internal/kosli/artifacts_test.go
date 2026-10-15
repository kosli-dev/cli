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
	value, err := artifact.Value()
	require.NoError(t, err)
	require.Equal(t, "app", value.Filename)
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

func TestListArtifactsByCommitSHARequestsTheCommitsArtifactsAndDecodesThem(t *testing.T) {
	server, seen := newFakeServer(t, http.StatusOK, `[{"filename":"a"},{"filename":"b"}]`)
	client := newTestClient(t, server.URL)

	artifacts, err := client.ListArtifactsByCommitSHA(context.Background(), "flow-1", "0fc1ba9", "trail-1")

	require.NoError(t, err)
	require.Equal(t, "/api/v2/artifacts/test-org/flow-1/commit_sha/0fc1ba9", seen.rawPath)
	require.Equal(t, "trail=trail-1", seen.rawQuery)
	value, err := artifacts.Value()
	require.NoError(t, err)
	require.Len(t, value, 2)
	require.Equal(t, "b", value[1].Filename)
}

func TestArtifactDecodesStringTimestamps(t *testing.T) {
	// the artifacts API returns created_at as a numeric string, not a number
	server, _ := newFakeServer(t, http.StatusOK,
		`{"created_at":"1452902400.25","history":[{"event":"cli reported","timestamp":"1452902400"}]}`)
	client := newTestClient(t, server.URL)

	artifact, err := client.GetArtifactByFingerprint(context.Background(), "flow-1", "abc", "")

	require.NoError(t, err)
	value, err := artifact.Value()
	require.NoError(t, err)
	require.Equal(t, "1452902400.25", value.CreatedAt.String())
	require.Equal(t, "1452902400", value.History[0].Timestamp.String())
}

func TestArtifactTellsAPresentButEmptyTrailFromAnAbsentOne(t *testing.T) {
	server, _ := newFakeServer(t, http.StatusOK, `{"trail_name":""}`)
	client := newTestClient(t, server.URL)

	artifact, err := client.GetArtifactByFingerprint(context.Background(), "flow-1", "abc", "")

	require.NoError(t, err)
	value, err := artifact.Value()
	require.NoError(t, err)
	require.NotNil(t, value.TrailName)
	require.Nil(t, value.TemplateReferenceName)
}
