package cloudrun

// reportType is the literal sent in the top-level "type" field, mirroring the
// "K8S", "ECS", "azure-apps", … values used by the other env-type reports.
const reportType = "cloud-run"

// EnvRequest is the PUT body sent to the Kosli "report/cloud-run" endpoint.
// Field naming mirrors the conventions documented in the server's
// out-snapshot-examples.txt (top-level "type" + "artifacts", camelCase
// per-artifact fields).
type EnvRequest struct {
	Type      string          `json:"type"`
	Artifacts []*RevisionData `json:"artifacts"`
}

// RevisionData represents one Cloud Run revision in the snapshot payload.
// One artifact is emitted per revision in each service's traffic configuration.
type RevisionData struct {
	RevisionName string            `json:"revisionName"`
	ServiceName  string            `json:"serviceName,omitempty"`
	Digests      map[string]string `json:"digests"`
	CreatedAt    int64             `json:"creationTimestamp"`
}

// ToEnvRequest flattens services into a list of revision artifacts. Services
// with no revisions contribute nothing, mirroring the ECS behaviour of
// services with no running tasks.
func ToEnvRequest(services []Service) *EnvRequest {
	artifacts := []*RevisionData{}
	for _, svc := range services {
		for _, rev := range svc.Revisions {
			artifacts = append(artifacts, &RevisionData{
				RevisionName: rev.Name,
				ServiceName:  svc.Name,
				Digests:      rev.Digests,
				CreatedAt:    rev.CreatedAt.Unix(),
			})
		}
	}
	return &EnvRequest{Type: reportType, Artifacts: artifacts}
}
