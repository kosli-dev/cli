package cloudrun

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestToEnvRequest_TypeIsCloudRun(t *testing.T) {
	got := ToEnvRequest(nil, nil)
	require.Equal(t, "cloud-run", got.Type)
}

func TestToEnvRequest_EmptyInput(t *testing.T) {
	got := ToEnvRequest(nil, nil)
	require.NotNil(t, got)
	require.Empty(t, got.Artifacts)
}

func TestToEnvRequest_SingleServiceSingleRevision(t *testing.T) {
	created := time.Date(2026, 4, 28, 12, 0, 0, 0, time.UTC)
	services := []Service{
		{
			Name: "svc-a",
			URI:  "https://svc-a.run.app",
			Revisions: []Revision{
				{
					Name:      "svc-a-rev1",
					Digests:   map[string]string{"img@sha256:aaa": "aaa"},
					CreatedAt: created,
				},
			},
		},
	}

	got := ToEnvRequest(services, nil)
	require.Len(t, got.Artifacts, 1)

	art := got.Artifacts[0]
	require.Equal(t, KindService, art.Kind)
	require.Equal(t, "svc-a", art.ServiceName)
	require.Equal(t, "svc-a-rev1", art.RevisionName)
	require.Empty(t, art.JobName)
	require.Equal(t, map[string]string{"img@sha256:aaa": "aaa"}, art.Digests)
	require.Equal(t, created.Unix(), art.CreatedAt)
}

func TestToEnvRequest_MultipleServicesMultipleRevisions(t *testing.T) {
	services := []Service{
		{
			Name: "svc-a",
			Revisions: []Revision{
				{Name: "a-rev1", Digests: map[string]string{"img@sha256:a1": "a1"}},
				{Name: "a-rev2", Digests: map[string]string{"img@sha256:a2": "a2"}},
			},
		},
		{
			Name: "svc-b",
			Revisions: []Revision{
				{Name: "b-rev1", Digests: map[string]string{"img@sha256:b1": "b1"}},
			},
		},
	}

	got := ToEnvRequest(services, nil)
	require.Len(t, got.Artifacts, 3)

	revisionNames := []string{
		got.Artifacts[0].RevisionName,
		got.Artifacts[1].RevisionName,
		got.Artifacts[2].RevisionName,
	}
	require.Equal(t, []string{"a-rev1", "a-rev2", "b-rev1"}, revisionNames)

	require.Equal(t, "svc-a", got.Artifacts[0].ServiceName)
	require.Equal(t, "svc-a", got.Artifacts[1].ServiceName)
	require.Equal(t, "svc-b", got.Artifacts[2].ServiceName)
}

func TestToEnvRequest_ServiceWithNoRevisionsContributesNothing(t *testing.T) {
	services := []Service{
		{Name: "empty-svc"},
		{
			Name: "svc-a",
			Revisions: []Revision{
				{Name: "rev", Digests: map[string]string{"img@sha256:x": "x"}},
			},
		},
	}

	got := ToEnvRequest(services, nil)
	require.Len(t, got.Artifacts, 1)
	require.Equal(t, "svc-a", got.Artifacts[0].ServiceName)
}

func TestToEnvRequest_ZeroCreatedAtSerialisesAsZero(t *testing.T) {
	services := []Service{
		{Name: "svc", Revisions: []Revision{{Name: "rev"}}},
	}

	got := ToEnvRequest(services, nil)
	require.Len(t, got.Artifacts, 1)
	require.Equal(t, time.Time{}.Unix(), got.Artifacts[0].CreatedAt)
}

// TestToEnvRequest_SerializesFlatFields_JSON locks the wire format: each
// artifact must serialise its kind/serviceName/revisionName fields as flat
// top-level keys (no nested context object), matching the convention used by
// ECS, K8S, and other env-type reports. Identifying fields are camelCase to
// mirror the GCP Cloud Run Admin API v2.
func TestToEnvRequest_SerializesFlatFields_JSON(t *testing.T) {
	services := []Service{
		{
			Name: "svc-a",
			Revisions: []Revision{
				{Name: "svc-a-rev1", Digests: map[string]string{"img@sha256:aaa": "aaa"}},
			},
		},
	}

	raw, err := json.Marshal(ToEnvRequest(services, nil))
	require.NoError(t, err)

	var decoded struct {
		Artifacts []map[string]any `json:"artifacts"`
	}
	require.NoError(t, json.Unmarshal(raw, &decoded))
	require.Len(t, decoded.Artifacts, 1)

	art := decoded.Artifacts[0]
	require.Equal(t, "service", art["kind"])
	require.Equal(t, "svc-a", art["serviceName"])
	require.Equal(t, "svc-a-rev1", art["revisionName"])
	require.NotContains(t, art, "cloud_run_context", "fields must be flat, not nested under cloud_run_context")
	require.NotContains(t, art, "service_name", "JSON keys must be camelCase to mirror GCP API")
	require.NotContains(t, art, "revision_name", "JSON keys must be camelCase to mirror GCP API")
}

func TestToEnvRequest_SingleJob(t *testing.T) {
	created := time.Date(2026, 4, 28, 12, 0, 0, 0, time.UTC)
	jobs := []Job{
		{
			Name:      "sandman-job",
			Digests:   map[string]string{"img@sha256:jjj": "jjj"},
			CreatedAt: created,
		},
	}

	got := ToEnvRequest(nil, jobs)
	require.Len(t, got.Artifacts, 1)

	art := got.Artifacts[0]
	require.Equal(t, KindJob, art.Kind)
	require.Equal(t, "sandman-job", art.JobName)
	require.Empty(t, art.ServiceName)
	require.Empty(t, art.RevisionName)
	require.Equal(t, map[string]string{"img@sha256:jjj": "jjj"}, art.Digests)
	require.Equal(t, created.Unix(), art.CreatedAt)
}

func TestToEnvRequest_ServicesAndJobsMixed(t *testing.T) {
	services := []Service{
		{Name: "svc-a", Revisions: []Revision{{Name: "svc-a-rev1", Digests: map[string]string{"img@sha256:s": "s"}}}},
	}
	jobs := []Job{
		{Name: "job-a", Digests: map[string]string{"img@sha256:j": "j"}},
	}

	got := ToEnvRequest(services, jobs)
	require.Len(t, got.Artifacts, 2)

	require.Equal(t, KindService, got.Artifacts[0].Kind, "service artifacts come first")
	require.Equal(t, "svc-a", got.Artifacts[0].ServiceName)
	require.Equal(t, "svc-a-rev1", got.Artifacts[0].RevisionName)
	require.Empty(t, got.Artifacts[0].JobName)

	require.Equal(t, KindJob, got.Artifacts[1].Kind)
	require.Equal(t, "job-a", got.Artifacts[1].JobName)
	require.Empty(t, got.Artifacts[1].ServiceName)
	require.Empty(t, got.Artifacts[1].RevisionName)
}

func TestToEnvRequest_JobSerializesAsFlatJobName(t *testing.T) {
	jobs := []Job{
		{Name: "sandman-job", Digests: map[string]string{"img@sha256:jjj": "jjj"}},
	}

	raw, err := json.Marshal(ToEnvRequest(nil, jobs))
	require.NoError(t, err)

	var decoded struct {
		Artifacts []map[string]any `json:"artifacts"`
	}
	require.NoError(t, json.Unmarshal(raw, &decoded))
	require.Len(t, decoded.Artifacts, 1)

	art := decoded.Artifacts[0]
	require.Equal(t, "job", art["kind"])
	require.Equal(t, "sandman-job", art["jobName"])
	require.NotContains(t, art, "serviceName", "job artifacts must not include serviceName")
	require.NotContains(t, art, "revisionName", "job artifacts must not include revisionName")
	require.NotContains(t, art, "job_name", "JSON keys must be camelCase to mirror GCP API")
}
