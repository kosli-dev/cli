package cloudrun

// EnvRequest is the PUT body sent to the Kosli "report/cloud-run" endpoint.
// It mirrors the shape of EcsEnvRequest in internal/aws.
type EnvRequest struct {
	Artifacts []*RevisionData `json:"artifacts"`
}

// RevisionData represents one Cloud Run revision in the snapshot payload.
// One artifact is emitted per revision in each service's traffic configuration.
type RevisionData struct {
	RevisionName string            `json:"revisionName"`
	Service      string            `json:"service_name,omitempty"`
	Project      string            `json:"project,omitempty"`
	Region       string            `json:"region,omitempty"`
	Digests      map[string]string `json:"digests"`
	CreatedAt    int64             `json:"creationTimestamp"`
}

// ToEnvRequest flattens services into a list of revision artifacts. Services
// with no revisions contribute nothing, mirroring the ECS behaviour of
// services with no running tasks.
func ToEnvRequest(services []Service, project, region string) *EnvRequest {
	artifacts := []*RevisionData{}
	for _, svc := range services {
		for _, rev := range svc.Revisions {
			artifacts = append(artifacts, &RevisionData{
				RevisionName: rev.Name,
				Service:      svc.Name,
				Project:      project,
				Region:       region,
				Digests:      rev.Digests,
				CreatedAt:    rev.CreatedAt.Unix(),
			})
		}
	}
	return &EnvRequest{Artifacts: artifacts}
}
