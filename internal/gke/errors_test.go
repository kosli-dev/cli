package gke

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func serviceDisabledError(t *testing.T, code codes.Code) error {
	t.Helper()
	s, err := status.New(code, "Cloud Asset API has not been used in project 123 before or it is disabled.").
		WithDetails(&errdetails.ErrorInfo{
			Reason:   "SERVICE_DISABLED",
			Domain:   "googleapis.com",
			Metadata: map[string]string{"service": "cloudasset.googleapis.com", "consumer": "projects/123"},
		})
	require.NoError(t, err)
	return s.Err()
}

func TestClassify(t *testing.T) {
	t.Run("nil stays nil", func(t *testing.T) {
		require.NoError(t, Classify(nil, "projects/p"))
	})

	t.Run("non-gRPC error passes through", func(t *testing.T) {
		original := errors.New("some plain go error")
		require.Same(t, original, Classify(original, "projects/p"))
	})

	t.Run("unrecognised gRPC code passes through", func(t *testing.T) {
		original := status.Error(codes.InvalidArgument, "bad parent")
		require.Same(t, original, Classify(original, "projects/p"))
	})

	t.Run("unauthenticated advises on Application Default Credentials", func(t *testing.T) {
		original := status.Error(codes.Unauthenticated, "token expired")
		got := Classify(original, "projects/p")
		require.ErrorContains(t, got, "GCP authentication failed")
		require.ErrorContains(t, got, "gcloud auth application-default login")
		require.ErrorContains(t, got, "Workload Identity")
		require.ErrorIs(t, got, original)
	})

	t.Run("permission denied names the parent and the permissions", func(t *testing.T) {
		original := status.Error(codes.PermissionDenied, "missing permission")
		got := Classify(original, "folders/123")
		require.ErrorContains(t, got, "GCP permission denied")
		require.ErrorContains(t, got, `"folders/123"`)
		require.ErrorContains(t, got, "cloudasset.assets.listContainerPod")
		require.ErrorContains(t, got, "serviceusage.services.use")
		require.ErrorIs(t, got, original)
	})

	for _, code := range []codes.Code{codes.PermissionDenied, codes.FailedPrecondition} {
		t.Run(fmt.Sprintf("disabled API reported as %s asks to enable it", code), func(t *testing.T) {
			original := serviceDisabledError(t, code)
			got := Classify(original, "projects/p")
			require.ErrorContains(t, got, "enable the Cloud Asset API (cloudasset.googleapis.com)")
			require.NotContains(t, got.Error(), "listContainerPod")
			require.ErrorIs(t, got, original)
		})
	}

	t.Run("not found names the parent", func(t *testing.T) {
		original := status.Error(codes.NotFound, "no such resource")
		got := Classify(original, "organizations/456")
		require.ErrorContains(t, got, `"organizations/456" not found or not accessible`)
		require.ErrorIs(t, got, original)
	})

	t.Run("classifies a wrapped gRPC error", func(t *testing.T) {
		original := status.Error(codes.PermissionDenied, "missing permission")
		got := Classify(fmt.Errorf("listing GKE pods in projects/p: %w", original), "projects/p")
		require.ErrorContains(t, got, "GCP permission denied")
	})
}
