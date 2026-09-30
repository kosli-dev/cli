package gke

import (
	"testing"

	"github.com/kosli-dev/cli/internal/filters"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func testPod(location, cluster, namespace, name string) Pod {
	return Pod{
		Location: location,
		Cluster:  cluster,
		Pod:      corev1.Pod{ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: namespace}},
	}
}

func TestFilterSelect(t *testing.T) {
	pods := []Pod{
		testPod("europe-west1", "prod-eu", "payments", "a"),
		testPod("europe-west1", "prod-eu", "kube-system", "b"),
		testPod("us-central1", "prod-us", "payments", "c"),
		testPod("us-central1", "staging", "payments", "d"),
	}

	for _, tt := range []struct {
		name    string
		filter  Filter
		want    []string
		wantErr string
	}{
		{
			name:   "zero filter selects every pod",
			filter: Filter{},
			want:   []string{"a", "b", "c", "d"},
		},
		{
			name:   "cluster names",
			filter: Filter{Clusters: filters.ResourceFilterOptions{IncludeNames: []string{"prod-us", "staging"}}},
			want:   []string{"c", "d"},
		},
		{
			name:   "cluster regex",
			filter: Filter{Clusters: filters.ResourceFilterOptions{IncludeNamesRegex: []string{"^prod-"}}},
			want:   []string{"a", "b", "c"},
		},
		{
			name:   "cluster names and regex are alternatives",
			filter: Filter{Clusters: filters.ResourceFilterOptions{IncludeNames: []string{"staging"}, IncludeNamesRegex: []string{"-eu$"}}},
			want:   []string{"a", "b", "d"},
		},
		{
			name:   "locations",
			filter: Filter{Locations: []string{"us-central1"}},
			want:   []string{"c", "d"},
		},
		{
			name:   "namespaces",
			filter: Filter{Namespaces: filters.ResourceFilterOptions{IncludeNames: []string{"kube-system"}}},
			want:   []string{"b"},
		},
		{
			name:   "excluded namespaces regex",
			filter: Filter{Namespaces: filters.ResourceFilterOptions{ExcludeNamesRegex: []string{"^kube-"}}},
			want:   []string{"a", "c", "d"},
		},
		{
			name: "every part must match",
			filter: Filter{
				Clusters:   filters.ResourceFilterOptions{IncludeNamesRegex: []string{"^prod-"}},
				Locations:  []string{"europe-west1"},
				Namespaces: filters.ResourceFilterOptions{ExcludeNames: []string{"kube-system"}},
			},
			want: []string{"a"},
		},
		{
			name:    "invalid cluster regex",
			filter:  Filter{Clusters: filters.ResourceFilterOptions{IncludeNamesRegex: []string{"("}}},
			wantErr: "invalid include name regex pattern (",
		},
		{
			name:    "invalid namespace regex",
			filter:  Filter{Namespaces: filters.ResourceFilterOptions{ExcludeNamesRegex: []string{"("}}},
			wantErr: "invalid exclude name regex pattern (",
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.filter.Select(pods)
			if tt.wantErr != "" {
				require.ErrorContains(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
			names := []string{}
			for _, p := range got {
				names = append(names, p.Name)
			}
			require.Equal(t, tt.want, names)
		})
	}
}

func TestFilterSelectLocations(t *testing.T) {
	pods := []Pod{
		testPod("europe-west1", "regional", "ns", "a"),
		testPod("europe-west1-b", "zonal-b", "ns", "b"),
		testPod("europe-west1-c", "zonal-c", "ns", "c"),
		testPod("europe-west10", "other-region", "ns", "d"),
		testPod("europe-west10-a", "other-region-zonal", "ns", "e"),
	}

	for _, tt := range []struct {
		name      string
		locations []string
		want      []string
	}{
		{name: "region matches its regional and zonal clusters", locations: []string{"europe-west1"}, want: []string{"a", "b", "c"}},
		{name: "zone matches only that zone", locations: []string{"europe-west1-b"}, want: []string{"b"}},
		{name: "region with a numeric suffix is another region", locations: []string{"europe-west10"}, want: []string{"d", "e"}},
		{name: "several locations", locations: []string{"europe-west1-c", "europe-west10"}, want: []string{"c", "d", "e"}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Filter{Locations: tt.locations}.Select(pods)
			require.NoError(t, err)
			names := []string{}
			for _, p := range got {
				names = append(names, p.Name)
			}
			require.Equal(t, tt.want, names)
		})
	}
}
