package cloudrun

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestToEnvRequest_EmptyInput(t *testing.T) {
	got := ToEnvRequest(nil, "p", "r")
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

	got := ToEnvRequest(services, "proj-1", "europe-west1")
	require.Len(t, got.Artifacts, 1)

	art := got.Artifacts[0]
	require.Equal(t, "svc-a-rev1", art.RevisionName)
	require.Equal(t, "svc-a", art.Service)
	require.Equal(t, "proj-1", art.Project)
	require.Equal(t, "europe-west1", art.Region)
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

	got := ToEnvRequest(services, "proj-1", "europe-west1")
	require.Len(t, got.Artifacts, 3)

	revisionNames := []string{got.Artifacts[0].RevisionName, got.Artifacts[1].RevisionName, got.Artifacts[2].RevisionName}
	require.Equal(t, []string{"a-rev1", "a-rev2", "b-rev1"}, revisionNames)

	require.Equal(t, "svc-a", got.Artifacts[0].Service)
	require.Equal(t, "svc-a", got.Artifacts[1].Service)
	require.Equal(t, "svc-b", got.Artifacts[2].Service)
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

	got := ToEnvRequest(services, "p", "r")
	require.Len(t, got.Artifacts, 1)
	require.Equal(t, "svc-a", got.Artifacts[0].Service)
}

func TestToEnvRequest_ProjectAndRegionOnEveryArtifact(t *testing.T) {
	services := []Service{
		{Name: "a", Revisions: []Revision{{Name: "a1"}}},
		{Name: "b", Revisions: []Revision{{Name: "b1"}}},
	}

	got := ToEnvRequest(services, "proj-x", "us-central1")

	for _, art := range got.Artifacts {
		require.Equal(t, "proj-x", art.Project)
		require.Equal(t, "us-central1", art.Region)
	}
}

func TestToEnvRequest_ZeroCreatedAtSerialisesAsZero(t *testing.T) {
	services := []Service{
		{Name: "svc", Revisions: []Revision{{Name: "rev"}}},
	}

	got := ToEnvRequest(services, "p", "r")
	require.Len(t, got.Artifacts, 1)
	require.Equal(t, time.Time{}.Unix(), got.Artifacts[0].CreatedAt)
}
