// Package gke reads GKE Pods from Cloud Asset Inventory for snapshot
// reporting. It needs only Google Cloud IAM permissions: no kubeconfig, no
// Kubernetes RBAC and no network path to a cluster control plane. Production
// code uses the real Asset Inventory API behind the unexported apiClient
// interface, which tests replace with a fake.
package gke

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	asset "cloud.google.com/go/asset/apiv1"
	"cloud.google.com/go/asset/apiv1/assetpb"
	gax "github.com/googleapis/gax-go/v2"
	"google.golang.org/api/iterator"
	"google.golang.org/grpc/codes"
	"google.golang.org/protobuf/encoding/protojson"
	corev1 "k8s.io/api/core/v1"
)

const (
	podAssetType = "k8s.io/Pod"
	// pageSize is the ListAssets maximum; the default of 100 multiplies the
	// calls counted against the Asset Inventory quota.
	pageSize = 1000
	// podAssetPrefix precedes projects/P/locations/L/clusters/C/k8s/namespaces/NS/pods/NAME
	podAssetPrefix = "//container.googleapis.com/"
)

// Pod is a Kubernetes Pod together with the GKE cluster it runs in.
type Pod struct {
	Location string
	Cluster  string
	Pod      corev1.Pod
}

// apiClient is the unexported seam that lets tests substitute a fake.
type apiClient interface {
	listPodAssets(ctx context.Context, parent string) ([]*assetpb.Asset, error)
}

// Client lists GKE Pods from Cloud Asset Inventory.
type Client struct {
	api apiClient
}

// New returns a Client backed by the Cloud Asset Inventory API using
// Application Default Credentials. Callers should defer Close().
func New(ctx context.Context) (*Client, error) {
	client, err := asset.NewClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("GCP client setup failed: %w", err)
	}
	return &Client{api: &gcpAPI{client: client}}, nil
}

// Close releases the underlying gRPC connection. Safe to call on a Client
// constructed with a fake apiClient (returns nil).
func (c *Client) Close() error {
	g, ok := c.api.(*gcpAPI)
	if !ok {
		return nil
	}
	return g.client.Close()
}

// ListPods returns every GKE Pod under parent, which is "projects/P",
// "folders/F" or "organizations/O".
func (c *Client) ListPods(ctx context.Context, parent string) ([]Pod, error) {
	assets, err := c.api.listPodAssets(ctx, parent)
	if err != nil {
		return nil, err
	}
	pods := make([]Pod, 0, len(assets))
	for _, a := range assets {
		pod, err := toPod(a)
		if err != nil {
			return nil, err
		}
		pods = append(pods, pod)
	}
	return pods, nil
}

// toPod decodes the Pod object that Asset Inventory stores as a JSON-shaped
// struct, so it is read with the Kubernetes API types' own JSON tags.
func toPod(a *assetpb.Asset) (Pod, error) {
	location, cluster, err := parseAssetName(a.GetName())
	if err != nil {
		return Pod{}, err
	}
	data, err := protojson.Marshal(a.GetResource().GetData())
	if err != nil {
		return Pod{}, fmt.Errorf("encoding pod asset %s: %w", a.GetName(), err)
	}
	var pod corev1.Pod
	if err := json.Unmarshal(data, &pod); err != nil {
		return Pod{}, fmt.Errorf("decoding pod asset %s: %w", a.GetName(), err)
	}
	return Pod{Location: location, Cluster: cluster, Pod: pod}, nil
}

// parseAssetName returns the location and cluster of a Pod asset name such as
// //container.googleapis.com/projects/P/locations/L/clusters/C/k8s/namespaces/NS/pods/NAME
func parseAssetName(name string) (location, cluster string, err error) {
	parts := strings.Split(strings.TrimPrefix(name, podAssetPrefix), "/")
	if !strings.HasPrefix(name, podAssetPrefix) || len(parts) != 11 ||
		parts[0] != "projects" || parts[2] != "locations" || parts[4] != "clusters" ||
		parts[6] != "k8s" || parts[7] != "namespaces" || parts[9] != "pods" {
		return "", "", fmt.Errorf("unexpected GKE pod asset name %q", name)
	}
	return parts[3], parts[5], nil
}

// gcpAPI is the production apiClient backed by the Cloud Asset Inventory API.
type gcpAPI struct {
	client *asset.Client
}

// listRetry adds ResourceExhausted to the SDK's default retry codes, so an
// organization-wide scan backs off on the Asset Inventory quota rather than failing.
var listRetry = gax.WithRetry(func() gax.Retryer {
	return gax.OnCodes([]codes.Code{
		codes.DeadlineExceeded,
		codes.Unavailable,
		codes.ResourceExhausted,
	}, gax.Backoff{
		Initial:    time.Second,
		Max:        time.Minute,
		Multiplier: 2,
	})
})

func (g *gcpAPI) listPodAssets(ctx context.Context, parent string) ([]*assetpb.Asset, error) {
	it := g.client.ListAssets(ctx, &assetpb.ListAssetsRequest{
		Parent:      parent,
		AssetTypes:  []string{podAssetType},
		ContentType: assetpb.ContentType_RESOURCE,
		PageSize:    pageSize,
	}, listRetry)
	var out []*assetpb.Asset
	for {
		a, err := it.Next()
		if err == iterator.Done {
			return out, nil
		}
		if err != nil {
			return nil, fmt.Errorf("listing GKE pods in %s: %w", parent, err)
		}
		out = append(out, a)
	}
}
