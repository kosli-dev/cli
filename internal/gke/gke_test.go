package gke

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"

	"cloud.google.com/go/asset/apiv1/assetpb"
	"github.com/kosli-dev/cli/internal/kube"
	"github.com/kosli-dev/cli/internal/logger"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/encoding/protojson"
	corev1 "k8s.io/api/core/v1"
)

type fakeAPI struct {
	assets []*assetpb.Asset
	err    error
	parent string
}

func (f *fakeAPI) listPodAssets(_ context.Context, parent string) ([]*assetpb.Asset, error) {
	f.parent = parent
	return f.assets, f.err
}

func fixtureAssets(t *testing.T) []*assetpb.Asset {
	t.Helper()
	data, err := os.ReadFile("testdata/list_pod_assets.json")
	require.NoError(t, err)
	var resp assetpb.ListAssetsResponse
	require.NoError(t, protojson.Unmarshal(data, &resp))
	return resp.GetAssets()
}

func TestParseAssetName(t *testing.T) {
	for _, tt := range []struct {
		name         string
		assetName    string
		wantLocation string
		wantCluster  string
		wantErr      bool
	}{
		{
			name:         "regional cluster",
			assetName:    "//container.googleapis.com/projects/p/locations/europe-west1/clusters/c1/k8s/namespaces/ns/pods/pod-1",
			wantLocation: "europe-west1",
			wantCluster:  "c1",
		},
		{
			name:         "zonal cluster",
			assetName:    "//container.googleapis.com/projects/p/locations/us-central1-a/clusters/c2/k8s/namespaces/ns/pods/pod-1",
			wantLocation: "us-central1-a",
			wantCluster:  "c2",
		},
		{
			name:      "another service",
			assetName: "//run.googleapis.com/projects/p/locations/europe-west1/services/svc",
			wantErr:   true,
		},
		{
			name:      "cluster rather than pod",
			assetName: "//container.googleapis.com/projects/p/locations/europe-west1/clusters/c1",
			wantErr:   true,
		},
		{
			name:      "namespace rather than pod",
			assetName: "//container.googleapis.com/projects/p/locations/europe-west1/clusters/c1/k8s/namespaces/ns",
			wantErr:   true,
		},
		{
			name:      "empty",
			assetName: "",
			wantErr:   true,
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			location, cluster, err := parseAssetName(tt.assetName)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tt.wantLocation, location)
			require.Equal(t, tt.wantCluster, cluster)
		})
	}
}

func TestListPods(t *testing.T) {
	t.Run("decodes the pod fields a snapshot reads", func(t *testing.T) {
		api := &fakeAPI{assets: fixtureAssets(t)}
		pods, err := (&Client{api: api}).ListPods(context.Background(), "projects/my-project")
		require.NoError(t, err)
		require.Equal(t, "projects/my-project", api.parent)
		require.Len(t, pods, 3)

		got := pods[0]
		require.Equal(t, "europe-west1", got.Location)
		require.Equal(t, "autopilot-cluster-1", got.Cluster)
		require.Equal(t, "api-7d9f8b6c4d-x2x9k", got.Pod.Name)
		require.Equal(t, "payments", got.Pod.Namespace)
		require.Equal(t, time.Date(2026, 9, 1, 10, 0, 0, 0, time.UTC), got.Pod.CreationTimestamp.UTC())
		require.Len(t, got.Pod.OwnerReferences, 1)
		require.Equal(t, "ReplicaSet", got.Pod.OwnerReferences[0].Kind)
		require.Equal(t, "api-7d9f8b6c4d", got.Pod.OwnerReferences[0].Name)
		require.Equal(t, "europe-west1-docker.pkg.dev/my-project/apps/api:1.4.2", got.Pod.Spec.Containers[0].Image)
		require.Equal(t, corev1.PodRunning, got.Pod.Status.Phase)
		require.Equal(t, []corev1.ContainerStatus{{
			Name:    "api",
			Image:   "europe-west1-docker.pkg.dev/my-project/apps/api:1.4.2",
			ImageID: "europe-west1-docker.pkg.dev/my-project/apps/api@sha256:3b2a8f2f5c1e4d6a7b8c9d0e1f2a3b4c5d6e7f8091a2b3c4d5e6f708192a3b4c",
			Ready:   true,
		}}, got.Pod.Status.ContainerStatuses)

		require.Equal(t, "us-central1-a", pods[1].Location)
		require.Equal(t, "prod-cluster", pods[1].Cluster)
	})

	t.Run("decoded pods produce the snapshot k8s payload", func(t *testing.T) {
		pods, err := (&Client{api: &fakeAPI{assets: fixtureAssets(t)}}).ListPods(context.Background(), "projects/my-project")
		require.NoError(t, err)

		k8sPods := make([]corev1.Pod, 0, len(pods))
		for _, p := range pods {
			k8sPods = append(k8sPods, p.Pod)
		}
		podsData, err := kube.ProcessPods(k8sPods, logger.NewStandardLogger())
		require.NoError(t, err)

		byName := map[string]*kube.PodData{}
		for _, d := range podsData {
			byName[d.PodName] = d
		}
		require.Len(t, byName, 2, "the Succeeded pod is not reported")
		api := byName["api-7d9f8b6c4d-x2x9k"]
		require.NotNil(t, api)
		require.Equal(t, map[string]string{
			"europe-west1-docker.pkg.dev/my-project/apps/api:1.4.2": "3b2a8f2f5c1e4d6a7b8c9d0e1f2a3b4c5d6e7f8091a2b3c4d5e6f708192a3b4c",
		}, api.Digests)
		require.Equal(t, time.Date(2026, 9, 1, 10, 0, 0, 0, time.UTC).Unix(), api.CreationTimestamp)
		require.NotNil(t, byName["kube-dns-5b8c7d9f6-abcde"])
	})

	t.Run("fails on an asset whose name is not a pod", func(t *testing.T) {
		asset := &assetpb.Asset{Name: "//container.googleapis.com/projects/p/locations/l/clusters/c", AssetType: podAssetType}
		_, err := (&Client{api: &fakeAPI{assets: []*assetpb.Asset{asset}}}).ListPods(context.Background(), "projects/p")
		require.ErrorContains(t, err, "//container.googleapis.com/projects/p/locations/l/clusters/c")
	})

	t.Run("passes list errors through", func(t *testing.T) {
		listErr := errors.New("boom")
		_, err := (&Client{api: &fakeAPI{err: listErr}}).ListPods(context.Background(), "projects/p")
		require.ErrorIs(t, err, listErr)
	})
}

func TestClose(t *testing.T) {
	require.NoError(t, (&Client{api: &fakeAPI{}}).Close())
}
