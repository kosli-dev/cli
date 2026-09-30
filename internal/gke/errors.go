package gke

import (
	"fmt"

	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Classify adds an actionable message to an Asset Inventory error based on its
// gRPC status; other errors pass through unchanged. parent is the listed scope,
// e.g. "projects/p".
func Classify(err error, parent string) error {
	if err == nil {
		return nil
	}
	s, ok := status.FromError(err)
	if !ok {
		return err
	}
	if serviceDisabled(s) {
		return fmt.Errorf(
			"GCP API disabled: enable the Cloud Asset API (cloudasset.googleapis.com) in the "+
				"quota project of the caller's credentials (underlying error: %w)",
			err,
		)
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
			"GCP permission denied: the caller needs 'cloudasset.assets.listContainerPod' and "+
				"'serviceusage.services.use' on %q, e.g. through a custom role or 'roles/cloudasset.viewer' "+
				"(underlying error: %w)",
			parent, err,
		)
	case codes.NotFound:
		return fmt.Errorf("GCP scope %q not found or not accessible (underlying error: %w)", parent, err)
	default:
		return err
	}
}

func serviceDisabled(s *status.Status) bool {
	for _, d := range s.Details() {
		if info, ok := d.(*errdetails.ErrorInfo); ok && info.GetReason() == "SERVICE_DISABLED" {
			return true
		}
	}
	return false
}
