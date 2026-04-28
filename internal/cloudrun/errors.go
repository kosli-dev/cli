package cloudrun

import (
	"fmt"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Classify wraps a Cloud Run SDK error with a user-actionable message based on
// the gRPC status code. Errors without a recognised code (or non-gRPC errors)
// pass through unchanged so callers can still inspect them.
//
// project and region are interpolated into messages where they help the user
// localise the failure (e.g. NotFound on a misspelled project).
func Classify(err error, project, region string) error {
	if err == nil {
		return nil
	}
	s, ok := status.FromError(err)
	if !ok {
		return err
	}
	switch s.Code() {
	case codes.Unauthenticated:
		return fmt.Errorf(
			"GCP authentication failed: ensure Application Default Credentials are available "+
				"(GOOGLE_APPLICATION_CREDENTIALS, 'gcloud auth application-default login', "+
				"or GCE/GKE metadata server / Workload Identity) (underlying error: %w)",
			err,
		)
	case codes.PermissionDenied:
		return fmt.Errorf(
			"GCP permission denied: the caller needs 'roles/run.viewer' on project %q (underlying error: %w)",
			project, err,
		)
	case codes.NotFound:
		return fmt.Errorf(
			"GCP project %q or region %q not found or not accessible (underlying error: %w)",
			project, region, err,
		)
	default:
		return err
	}
}
