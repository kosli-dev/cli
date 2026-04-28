package cloudrun

import (
	"context"
	"testing"
	"time"

	"cloud.google.com/go/run/apiv2/runpb"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// fakeAPI is the in-memory test double for apiClient.
type fakeAPI struct {
	services  []*runpb.Service
	revisions map[string]*runpb.Revision
	listErr   error
	getErr    error
}

func (f *fakeAPI) listServices(_ context.Context, _, _ string) ([]*runpb.Service, error) {
	if f.listErr != nil {
		return nil, f.listErr
	}
	return f.services, nil
}

func (f *fakeAPI) getRevision(_ context.Context, name string) (*runpb.Revision, error) {
	if f.getErr != nil {
		return nil, f.getErr
	}
	rev, ok := f.revisions[name]
	if !ok {
		return nil, errNotFound{name: name}
	}
	return rev, nil
}

type errNotFound struct{ name string }

func (e errNotFound) Error() string { return "revision not found: " + e.name }

const (
	testProject = "hello-world-cli-demo"
	testRegion  = "europe-west1"
)

func svcResource(name string) string {
	return "projects/" + testProject + "/locations/" + testRegion + "/services/" + name
}

func revResource(svc, rev string) string {
	return svcResource(svc) + "/revisions/" + rev
}

func newClient(fake *fakeAPI) *Client {
	return &Client{api: fake}
}

func TestListServices_SingleRevisionDigestPinned(t *testing.T) {
	created := time.Date(2026, 4, 28, 12, 0, 0, 0, time.UTC)
	fake := &fakeAPI{
		services: []*runpb.Service{
			{
				Name: svcResource("svc1"),
				Uri:  "https://svc1.run.app",
				Traffic: []*runpb.TrafficTarget{
					{Revision: "svc1-rev1", Percent: 100, Type: runpb.TrafficTargetAllocationType_TRAFFIC_TARGET_ALLOCATION_TYPE_REVISION},
				},
			},
		},
		revisions: map[string]*runpb.Revision{
			revResource("svc1", "svc1-rev1"): {
				Name:       revResource("svc1", "svc1-rev1"),
				CreateTime: timestamppb.New(created),
				Containers: []*runpb.Container{
					{Image: "gcr.io/foo/bar@sha256:abc123"},
				},
			},
		},
	}

	got, err := newClient(fake).ListServices(context.Background(), testProject, testRegion)
	require.NoError(t, err)
	require.Len(t, got, 1)

	svc := got[0]
	require.Equal(t, "svc1", svc.Name)
	require.Equal(t, "https://svc1.run.app", svc.URI)
	require.Len(t, svc.Revisions, 1)

	rev := svc.Revisions[0]
	require.Equal(t, "svc1-rev1", rev.Name)
	require.True(t, rev.CreatedAt.Equal(created), "CreatedAt = %v, want %v", rev.CreatedAt, created)
	require.Equal(t, map[string]string{"gcr.io/foo/bar@sha256:abc123": "abc123"}, rev.Digests)
}

func TestListServices_TagPinnedImageYieldsEmptyDigest(t *testing.T) {
	fake := &fakeAPI{
		services: []*runpb.Service{
			{
				Name: svcResource("svc1"),
				Traffic: []*runpb.TrafficTarget{
					{Revision: "svc1-rev1", Percent: 100, Type: runpb.TrafficTargetAllocationType_TRAFFIC_TARGET_ALLOCATION_TYPE_REVISION},
				},
			},
		},
		revisions: map[string]*runpb.Revision{
			revResource("svc1", "svc1-rev1"): {
				Name: revResource("svc1", "svc1-rev1"),
				Containers: []*runpb.Container{
					{Image: "gcr.io/foo/bar:v1"},
				},
			},
		},
	}

	got, err := newClient(fake).ListServices(context.Background(), testProject, testRegion)
	require.NoError(t, err)
	require.Equal(t, map[string]string{"gcr.io/foo/bar:v1": ""}, got[0].Revisions[0].Digests)
}

func TestListServices_TrafficSplitReturnsBothRevisions(t *testing.T) {
	fake := &fakeAPI{
		services: []*runpb.Service{
			{
				Name: svcResource("svc1"),
				Traffic: []*runpb.TrafficTarget{
					{Revision: "svc1-rev1", Percent: 90, Type: runpb.TrafficTargetAllocationType_TRAFFIC_TARGET_ALLOCATION_TYPE_REVISION},
					{Revision: "svc1-rev2", Percent: 10, Type: runpb.TrafficTargetAllocationType_TRAFFIC_TARGET_ALLOCATION_TYPE_REVISION},
				},
			},
		},
		revisions: map[string]*runpb.Revision{
			revResource("svc1", "svc1-rev1"): {
				Name:       revResource("svc1", "svc1-rev1"),
				Containers: []*runpb.Container{{Image: "gcr.io/foo/bar@sha256:rev1"}},
			},
			revResource("svc1", "svc1-rev2"): {
				Name:       revResource("svc1", "svc1-rev2"),
				Containers: []*runpb.Container{{Image: "gcr.io/foo/bar@sha256:rev2"}},
			},
		},
	}

	got, err := newClient(fake).ListServices(context.Background(), testProject, testRegion)
	require.NoError(t, err)
	require.Len(t, got, 1)
	require.Len(t, got[0].Revisions, 2)

	names := []string{got[0].Revisions[0].Name, got[0].Revisions[1].Name}
	require.ElementsMatch(t, []string{"svc1-rev1", "svc1-rev2"}, names)
}

func TestListServices_ZeroPercentRevisionStillIncluded(t *testing.T) {
	fake := &fakeAPI{
		services: []*runpb.Service{
			{
				Name: svcResource("svc1"),
				Traffic: []*runpb.TrafficTarget{
					{Revision: "svc1-rev1", Percent: 100, Type: runpb.TrafficTargetAllocationType_TRAFFIC_TARGET_ALLOCATION_TYPE_REVISION},
					{Revision: "svc1-rev2", Percent: 0, Type: runpb.TrafficTargetAllocationType_TRAFFIC_TARGET_ALLOCATION_TYPE_REVISION},
				},
			},
		},
		revisions: map[string]*runpb.Revision{
			revResource("svc1", "svc1-rev1"): {Name: revResource("svc1", "svc1-rev1"), Containers: []*runpb.Container{{Image: "img@sha256:rev1"}}},
			revResource("svc1", "svc1-rev2"): {Name: revResource("svc1", "svc1-rev2"), Containers: []*runpb.Container{{Image: "img@sha256:rev2"}}},
		},
	}

	got, err := newClient(fake).ListServices(context.Background(), testProject, testRegion)
	require.NoError(t, err)
	require.Len(t, got[0].Revisions, 2)
}

func TestListServices_TrafficLatestResolvesToLatestReadyRevision(t *testing.T) {
	fake := &fakeAPI{
		services: []*runpb.Service{
			{
				Name:                svcResource("svc1"),
				LatestReadyRevision: revResource("svc1", "svc1-latest"),
				Traffic: []*runpb.TrafficTarget{
					{Type: runpb.TrafficTargetAllocationType_TRAFFIC_TARGET_ALLOCATION_TYPE_LATEST, Percent: 100},
				},
			},
		},
		revisions: map[string]*runpb.Revision{
			revResource("svc1", "svc1-latest"): {
				Name:       revResource("svc1", "svc1-latest"),
				Containers: []*runpb.Container{{Image: "img@sha256:latest-digest"}},
			},
		},
	}

	got, err := newClient(fake).ListServices(context.Background(), testProject, testRegion)
	require.NoError(t, err)
	require.Len(t, got[0].Revisions, 1)
	require.Equal(t, "svc1-latest", got[0].Revisions[0].Name)
}

func TestListServices_DedupesRevisionReferencedTwice(t *testing.T) {
	fake := &fakeAPI{
		services: []*runpb.Service{
			{
				Name:                svcResource("svc1"),
				LatestReadyRevision: revResource("svc1", "svc1-rev1"),
				Traffic: []*runpb.TrafficTarget{
					{Type: runpb.TrafficTargetAllocationType_TRAFFIC_TARGET_ALLOCATION_TYPE_LATEST, Percent: 50},
					{Revision: "svc1-rev1", Percent: 50, Type: runpb.TrafficTargetAllocationType_TRAFFIC_TARGET_ALLOCATION_TYPE_REVISION},
				},
			},
		},
		revisions: map[string]*runpb.Revision{
			revResource("svc1", "svc1-rev1"): {
				Name:       revResource("svc1", "svc1-rev1"),
				Containers: []*runpb.Container{{Image: "img@sha256:rev1"}},
			},
		},
	}

	got, err := newClient(fake).ListServices(context.Background(), testProject, testRegion)
	require.NoError(t, err)
	require.Len(t, got[0].Revisions, 1, "the same revision must not appear twice")
}

func TestListServices_MultipleContainersAllAppearInDigests(t *testing.T) {
	fake := &fakeAPI{
		services: []*runpb.Service{
			{
				Name: svcResource("svc1"),
				Traffic: []*runpb.TrafficTarget{
					{Revision: "svc1-rev1", Percent: 100, Type: runpb.TrafficTargetAllocationType_TRAFFIC_TARGET_ALLOCATION_TYPE_REVISION},
				},
			},
		},
		revisions: map[string]*runpb.Revision{
			revResource("svc1", "svc1-rev1"): {
				Name: revResource("svc1", "svc1-rev1"),
				Containers: []*runpb.Container{
					{Image: "gcr.io/foo/main@sha256:main"},
					{Image: "gcr.io/foo/sidecar@sha256:side"},
				},
			},
		},
	}

	got, err := newClient(fake).ListServices(context.Background(), testProject, testRegion)
	require.NoError(t, err)
	require.Equal(t, map[string]string{
		"gcr.io/foo/main@sha256:main":    "main",
		"gcr.io/foo/sidecar@sha256:side": "side",
	}, got[0].Revisions[0].Digests)
}

func TestListServices_ServiceWithNoTrafficTargetsHasEmptyRevisions(t *testing.T) {
	fake := &fakeAPI{
		services: []*runpb.Service{
			{Name: svcResource("svc1")},
		},
	}

	got, err := newClient(fake).ListServices(context.Background(), testProject, testRegion)
	require.NoError(t, err)
	require.Len(t, got, 1)
	require.Empty(t, got[0].Revisions)
}

func TestListServices_MultipleServices(t *testing.T) {
	fake := &fakeAPI{
		services: []*runpb.Service{
			{
				Name: svcResource("svc-a"),
				Traffic: []*runpb.TrafficTarget{
					{Revision: "svc-a-rev1", Percent: 100, Type: runpb.TrafficTargetAllocationType_TRAFFIC_TARGET_ALLOCATION_TYPE_REVISION},
				},
			},
			{
				Name: svcResource("svc-b"),
				Traffic: []*runpb.TrafficTarget{
					{Revision: "svc-b-rev1", Percent: 100, Type: runpb.TrafficTargetAllocationType_TRAFFIC_TARGET_ALLOCATION_TYPE_REVISION},
				},
			},
		},
		revisions: map[string]*runpb.Revision{
			revResource("svc-a", "svc-a-rev1"): {Name: revResource("svc-a", "svc-a-rev1"), Containers: []*runpb.Container{{Image: "img@sha256:a"}}},
			revResource("svc-b", "svc-b-rev1"): {Name: revResource("svc-b", "svc-b-rev1"), Containers: []*runpb.Container{{Image: "img@sha256:b"}}},
		},
	}

	got, err := newClient(fake).ListServices(context.Background(), testProject, testRegion)
	require.NoError(t, err)
	require.Len(t, got, 2)

	names := []string{got[0].Name, got[1].Name}
	require.ElementsMatch(t, []string{"svc-a", "svc-b"}, names)
}
