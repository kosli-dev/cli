package cloudrun

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestClassify_NilStaysNil(t *testing.T) {
	require.NoError(t, Classify(nil, "p", "r"))
}

func TestClassify_NonGRPCErrorPassesThrough(t *testing.T) {
	original := errors.New("some plain go error")
	got := Classify(original, "p", "r")
	require.Same(t, original, got)
}

func TestClassify_UnknownGRPCCodePassesThrough(t *testing.T) {
	original := status.Error(codes.ResourceExhausted, "rate limited")
	got := Classify(original, "p", "r")
	require.Same(t, original, got)
}

func TestClassify_UnauthenticatedReturnsADCAdvice(t *testing.T) {
	original := status.Error(codes.Unauthenticated, "token expired")
	got := Classify(original, "proj-1", "europe-west1")

	require.Error(t, got)
	require.Contains(t, got.Error(), "GCP authentication failed")
	require.Contains(t, got.Error(), "GOOGLE_APPLICATION_CREDENTIALS")
	require.Contains(t, got.Error(), "gcloud auth application-default login")
	require.Contains(t, got.Error(), "metadata server")
	require.Contains(t, got.Error(), "Workload Identity")
	require.ErrorIs(t, got, original, "underlying error must be preserved via %%w")
}

func TestClassify_PermissionDeniedNamesProjectAndRoleViewer(t *testing.T) {
	original := status.Error(codes.PermissionDenied, "missing iam role")
	got := Classify(original, "proj-1", "europe-west1")

	require.Error(t, got)
	require.Contains(t, got.Error(), "GCP permission denied")
	require.Contains(t, got.Error(), "roles/run.viewer")
	require.Contains(t, got.Error(), `"proj-1"`)
	require.ErrorIs(t, got, original)
}

func TestClassify_NotFoundNamesProjectAndRegion(t *testing.T) {
	original := status.Error(codes.NotFound, "no such resource")
	got := Classify(original, "bad-project", "europe-west1")

	require.Error(t, got)
	require.Contains(t, got.Error(), "not found or not accessible")
	require.Contains(t, got.Error(), `"bad-project"`)
	require.Contains(t, got.Error(), `"europe-west1"`)
	require.ErrorIs(t, got, original)
}
