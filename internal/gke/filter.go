package gke

import (
	"slices"

	"github.com/kosli-dev/cli/internal/filters"
	corev1 "k8s.io/api/core/v1"
)

// Filter selects pods by cluster, location and namespace; an unset part selects
// every value. Selection runs client side because ListAssets has no content filter.
type Filter struct {
	Clusters   filters.ResourceFilterOptions
	Locations  []string
	Namespaces filters.ResourceFilterOptions
}

// Select returns the Kubernetes Pods among pods that match every part of f.
func (f Filter) Select(pods []Pod) ([]corev1.Pod, error) {
	clusters := f.Clusters.Compile()
	namespaces := f.Namespaces.Compile()
	selected := []corev1.Pod{}
	for _, p := range pods {
		if len(f.Locations) > 0 && !slices.Contains(f.Locations, p.Location) {
			continue
		}
		included, err := clusters.ShouldInclude(p.Cluster)
		if err != nil {
			return nil, err
		}
		if !included {
			continue
		}
		included, err = namespaces.ShouldInclude(p.Pod.Namespace)
		if err != nil {
			return nil, err
		}
		if included {
			selected = append(selected, p.Pod)
		}
	}
	return selected, nil
}
