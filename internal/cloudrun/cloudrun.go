// Package cloudrun reads Cloud Run service and revision data from GCP for
// snapshot reporting. The package is designed around an unexported apiClient
// interface so production code uses the real Cloud Run Admin API v2 and tests
// can swap in a fake without touching GCP.
package cloudrun

import (
	"context"
	"fmt"
	"strings"
	"time"

	run "cloud.google.com/go/run/apiv2"
	"cloud.google.com/go/run/apiv2/runpb"
	"google.golang.org/api/iterator"
)

const sha256Marker = "@sha256:"

// Service is a Cloud Run service together with the revisions referenced in
// its current traffic configuration (any percent, including 0%).
type Service struct {
	Name      string
	URI       string
	Revisions []Revision
}

// Revision is a single Cloud Run revision and the digest-pinned images of its
// containers. A digest value of "" means the image string was not digest-pinned
// and the digest could not be parsed without a registry lookup.
type Revision struct {
	Name      string
	Digests   map[string]string
	CreatedAt time.Time
}

// apiClient is the unexported seam that lets tests substitute a fake.
type apiClient interface {
	listServices(ctx context.Context, project, region string) ([]*runpb.Service, error)
	getRevision(ctx context.Context, name string) (*runpb.Revision, error)
}

// Client fetches Cloud Run data from GCP.
type Client struct {
	api apiClient
}

// New returns a Client backed by the real Cloud Run Admin API v2 using
// Application Default Credentials. Construction errors (typically rare in a
// cluster cron job, since the metadata server provides credentials) are
// wrapped with a generic "GCP client setup failed" prefix; the SDK's own
// message is preserved via %w for diagnosis.
func New(ctx context.Context) (*Client, error) {
	services, err := run.NewServicesClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("GCP client setup failed: %w", err)
	}
	revisions, err := run.NewRevisionsClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("GCP client setup failed: %w", err)
	}
	return &Client{api: &gcpAPI{services: services, revisions: revisions}}, nil
}

// ListServices returns every Cloud Run service in the given project+region,
// each populated with the revisions referenced in its traffic configuration.
// TrafficTarget entries of type LATEST are resolved via the service's
// LatestReadyRevision, and revisions referenced more than once are deduped.
func (c *Client) ListServices(ctx context.Context, project, region string) ([]Service, error) {
	rawServices, err := c.api.listServices(ctx, project, region)
	if err != nil {
		return nil, err
	}
	out := make([]Service, 0, len(rawServices))
	for _, raw := range rawServices {
		svc, err := c.toService(ctx, raw)
		if err != nil {
			return nil, err
		}
		out = append(out, svc)
	}
	return out, nil
}

func (c *Client) toService(ctx context.Context, raw *runpb.Service) (Service, error) {
	svc := Service{
		Name: shortName(raw.GetName()),
		URI:  raw.GetUri(),
	}
	revNames := trafficRevisionNames(raw)
	for _, revShort := range revNames {
		fullName := raw.GetName() + "/revisions/" + revShort
		rev, err := c.api.getRevision(ctx, fullName)
		if err != nil {
			return Service{}, fmt.Errorf("getting revision %s: %w", fullName, err)
		}
		svc.Revisions = append(svc.Revisions, toRevision(rev))
	}
	return svc, nil
}

// trafficRevisionNames returns the deduped short names of revisions referenced
// in the service's traffic configuration. TrafficTarget entries of type LATEST
// are resolved to LatestReadyRevision; entries with an empty resolved name are
// skipped.
func trafficRevisionNames(svc *runpb.Service) []string {
	seen := map[string]struct{}{}
	out := []string{}
	for _, t := range svc.GetTraffic() {
		name := t.GetRevision()
		if name == "" {
			name = shortName(svc.GetLatestReadyRevision())
		}
		if name == "" {
			continue
		}
		if _, dup := seen[name]; dup {
			continue
		}
		seen[name] = struct{}{}
		out = append(out, name)
	}
	return out
}

func toRevision(rev *runpb.Revision) Revision {
	digests := make(map[string]string, len(rev.GetContainers()))
	for _, container := range rev.GetContainers() {
		image := container.GetImage()
		digests[image] = parseDigest(image)
	}
	var createdAt time.Time
	if ts := rev.GetCreateTime(); ts != nil {
		createdAt = ts.AsTime()
	}
	return Revision{
		Name:      shortName(rev.GetName()),
		Digests:   digests,
		CreatedAt: createdAt,
	}
}

// parseDigest extracts the sha256 hex out of a digest-pinned image reference
// like "gcr.io/foo/bar@sha256:<hex>". Tag-pinned references and inputs without
// the marker yield an empty string, mirroring the ECS snapshot fallback.
func parseDigest(image string) string {
	idx := strings.Index(image, sha256Marker)
	if idx < 0 {
		return ""
	}
	return image[idx+len(sha256Marker):]
}

// shortName returns the last path component of a fully-qualified GCP resource
// name like "projects/p/locations/r/services/svc" -> "svc". Non-qualified
// inputs are returned unchanged.
func shortName(fullName string) string {
	if i := strings.LastIndex(fullName, "/"); i >= 0 {
		return fullName[i+1:]
	}
	return fullName
}

// gcpAPI is the production apiClient backed by the Cloud Run Admin API v2.
type gcpAPI struct {
	services  *run.ServicesClient
	revisions *run.RevisionsClient
}

func (g *gcpAPI) listServices(ctx context.Context, project, region string) ([]*runpb.Service, error) {
	parent := fmt.Sprintf("projects/%s/locations/%s", project, region)
	it := g.services.ListServices(ctx, &runpb.ListServicesRequest{Parent: parent})
	var out []*runpb.Service
	for {
		svc, err := it.Next()
		if err == iterator.Done {
			return out, nil
		}
		if err != nil {
			return nil, fmt.Errorf("listing Cloud Run services in %s: %w", parent, err)
		}
		out = append(out, svc)
	}
}

func (g *gcpAPI) getRevision(ctx context.Context, name string) (*runpb.Revision, error) {
	return g.revisions.GetRevision(ctx, &runpb.GetRevisionRequest{Name: name})
}
