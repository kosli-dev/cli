package cloudrun

// reportType is the literal sent in the top-level "type" field, mirroring the
// "K8S", "ECS", "azure-apps", … values used by the other env-type reports.
const reportType = "cloud-run"

// Kind values for ArtifactData.Kind. Both Cloud Run Services (one artifact per
// revision) and Cloud Run Jobs (one artifact per Job) flow through the same
// artifact shape, distinguished by this field.
const (
	KindService = "service"
	KindJob     = "job"
)

// EnvRequest is the PUT body sent to the Kosli "report/cloud-run" endpoint.
// Field naming mirrors the conventions documented in the server's
// out-snapshot-examples.txt (top-level "type" + "artifacts").
type EnvRequest struct {
	Type      string          `json:"type"`
	Artifacts []*ArtifactData `json:"artifacts"`
}

// ArtifactData is one entry in the snapshot's artifacts array. Service-revision
// artifacts and Job artifacts share this shape; Kind discriminates between
// them and the kind-specific identifying fields populate accordingly. Fields
// are flat (top-level) to match the convention used by ECS, K8S, and the
// other env-type reports. JSON keys are camelCase to mirror the GCP Cloud Run
// Admin API v2, which is the source-of-truth for these field names.
type ArtifactData struct {
	Digests      map[string]string `json:"digests"`
	CreatedAt    int64             `json:"creationTimestamp"`
	Kind         string            `json:"kind"`
	ServiceName  string            `json:"serviceName,omitempty"`
	RevisionName string            `json:"revisionName,omitempty"`
	JobName      string            `json:"jobName,omitempty"`
}

// ToEnvRequest flattens services into a list of artifacts, one per revision.
// Services with no revisions contribute nothing, mirroring the ECS behaviour
// of services with no running tasks.
func ToEnvRequest(services []Service) *EnvRequest {
	artifacts := []*ArtifactData{}
	for _, svc := range services {
		for _, rev := range svc.Revisions {
			artifacts = append(artifacts, &ArtifactData{
				Digests:      rev.Digests,
				CreatedAt:    rev.CreatedAt.Unix(),
				Kind:         KindService,
				ServiceName:  svc.Name,
				RevisionName: rev.Name,
			})
		}
	}
	return &EnvRequest{Type: reportType, Artifacts: artifacts}
}
