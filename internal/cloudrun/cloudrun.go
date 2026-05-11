// Package cloudrun reads Cloud Run service and revision data from GCP for
// snapshot reporting. The package is designed around an unexported apiClient
// interface so production code uses the real Cloud Run Admin API v2 and tests
// can swap in a fake without touching GCP.
package cloudrun

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	run "cloud.google.com/go/run/apiv2"
	"cloud.google.com/go/run/apiv2/runpb"
	"github.com/kosli-dev/cli/internal/logger"
	"github.com/kosli-dev/cli/internal/requests"
	"golang.org/x/oauth2/google"
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

// Job is a Cloud Run Job and the digest-pinned images of the containers in its
// task template. Jobs do not have a revision/traffic-split model — there is one
// current image per Job, taken from Template.Template.Containers. Same digest
// semantics as Revision: "" means the image was not digest-pinned and the
// digest could not be parsed without a registry lookup.
type Job struct {
	Name      string
	Digests   map[string]string
	CreatedAt time.Time
}

// apiClient is the unexported seam that lets tests substitute a fake.
type apiClient interface {
	listServices(ctx context.Context, project, region string) ([]*runpb.Service, error)
	getRevision(ctx context.Context, name string) (*runpb.Revision, error)
	listJobs(ctx context.Context, project, region string) ([]*runpb.Job, error)
}

// Client fetches Cloud Run data from GCP. The optional resolver, when set,
// is consulted for any tag-pinned image whose digest cannot be parsed from
// the image string alone — it queries the OCI Distribution endpoint of the
// hosting registry to look up the current sha256. The optional tagResolver,
// when set (via the --resolve-names flag at construction time), is consulted
// for any digest-pinned image to recover a tag pointing at the same digest,
// so the artifact's display name surfaces the deploy-time tag (commit SHA
// or version) instead of the opaque sha256. Both resolvers fail open: the
// original key/value stays in place if the lookup fails.
type Client struct {
	api         apiClient
	resolver    digestResolver
	tagResolver tagResolver
	log         *logger.Logger
}

// New returns a Client backed by the real Cloud Run Admin API v2 using
// Application Default Credentials. Construction errors (typically rare in a
// cluster cron job, since the metadata server provides credentials) are
// wrapped with a generic "GCP client setup failed" prefix; the SDK's own
// message is preserved via %w for diagnosis. Callers should defer Close().
//
// The returned Client is wired with a registry-lookup resolver for tag-
// pinned images using ADC. When resolveNames is true, it is also wired with
// a reverse tag resolver against Artifact Registry, so digest-pinned
// artifact names (e.g. for Service revisions) are rewritten to their
// tag-pinned equivalents (commit SHA, version) for display. If ADC is
// unavailable both resolvers are left nil and the Client falls back to the
// pre-resolution behaviour: tag-pinned images report empty digests and
// digest-pinned names stay digest-pinned.
func New(ctx context.Context, log *logger.Logger, resolveNames bool) (*Client, error) {
	services, err := run.NewServicesClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("GCP client setup failed: %w", err)
	}
	revisions, err := run.NewRevisionsClient(ctx)
	if err != nil {
		_ = services.Close()
		return nil, fmt.Errorf("GCP client setup failed: %w", err)
	}
	jobs, err := run.NewJobsClient(ctx)
	if err != nil {
		_ = services.Close()
		_ = revisions.Close()
		return nil, fmt.Errorf("GCP client setup failed: %w", err)
	}

	var digestRes digestResolver
	var tagRes tagResolver
	src, terr := google.DefaultTokenSource(ctx, "https://www.googleapis.com/auth/cloud-platform")
	switch {
	case terr != nil:
		log.Debug("ADC token source unavailable; registry resolution disabled: %v", terr)
	default:
		// One HTTP client shared by both resolvers so that hundreds of
		// per-artifact registry calls reuse a single connection pool
		// (TCP + TLS handshakes amortise across the snapshot).
		kosliClient, cerr := requests.NewKosliClient("", 1, log.DebugEnabled, log)
		if cerr != nil {
			log.Debug("registry HTTP client unavailable; registry resolution disabled: %v", cerr)
			break
		}
		digestRes = &gcpRegistryResolver{tokens: src, client: kosliClient, log: log}
		if resolveNames {
			tagRes = &gcpArtifactRegistryTagResolver{tokens: src, client: kosliClient, log: log}
		}
	}

	return &Client{
		api:         &gcpAPI{services: services, revisions: revisions, jobs: jobs},
		resolver:    digestRes,
		tagResolver: tagRes,
		log:         log,
	}, nil
}

// Close releases the underlying gRPC connections. Safe to call on a Client
// constructed with a fake apiClient (returns nil). All clients are always
// closed; errors from each are joined so none are silently dropped.
func (c *Client) Close() error {
	g, ok := c.api.(*gcpAPI)
	if !ok {
		return nil
	}
	return errors.Join(g.services.Close(), g.revisions.Close(), g.jobs.Close())
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
		revision := toRevision(rev)
		c.resolveTagPinnedDigests(revision.Digests)
		c.resolveNamesForDigestPinned(revision.Digests)
		svc.Revisions = append(svc.Revisions, revision)
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

// ListJobs returns every Cloud Run Job in the given project+region. Each Job
// carries the digest-pinned images of the containers in its task template. The
// API returns the full Job resource (including its template) on the list call,
// so unlike services no per-resource follow-up is needed.
func (c *Client) ListJobs(ctx context.Context, project, region string) ([]Job, error) {
	rawJobs, err := c.api.listJobs(ctx, project, region)
	if err != nil {
		return nil, err
	}
	out := make([]Job, 0, len(rawJobs))
	for _, raw := range rawJobs {
		job := toJob(raw)
		c.resolveTagPinnedDigests(job.Digests)
		c.resolveNamesForDigestPinned(job.Digests)
		out = append(out, job)
	}
	return out, nil
}

// resolveTagPinnedDigests walks the digests map and, for each entry whose
// value is empty (i.e., the image string was not digest-pinned), asks the
// resolver to look up the digest in the hosting registry. On success the
// resolved hex is filled in as the value; the tag-pinned key is left in
// place so the artifact keeps its recognizable deploy-time name (e.g.
// "…/ghost-job:c85b06b0…" rather than "…/ghost-job@sha256:…"). Failures
// are logged at debug level and leave the entry untouched — never fatal.
//
// No-op when the Client was constructed without a resolver (test paths,
// or when ADC token-source construction failed).
func (c *Client) resolveTagPinnedDigests(digests map[string]string) {
	if c.resolver == nil {
		return
	}
	for image, hex := range digests {
		if hex != "" {
			continue
		}
		resolvedHex, err := c.resolver.Resolve(image)
		if err != nil {
			c.log.Debug("registry digest resolution failed for %q: %v", image, err)
			continue
		}
		digests[image] = resolvedHex
	}
}

// resolveNamesForDigestPinned walks the digests map and, for each entry
// whose image string is digest-pinned (contains "@sha256:"), asks the tag
// resolver for a tag pointing at the digest. On success the digest-pinned
// key is replaced with a tag-pinned key ("image:<tag>"); the hex value
// (the digest) is preserved. Failures are logged at debug level and leave
// the entry untouched — never fatal.
//
// No-op when the Client was constructed with resolveNames=false (the
// tagResolver is nil), or when ADC token-source construction failed.
func (c *Client) resolveNamesForDigestPinned(digests map[string]string) {
	if c.tagResolver == nil {
		return
	}
	for image, hex := range digests {
		idx := strings.Index(image, sha256Marker)
		if idx < 0 {
			continue // not digest-pinned
		}
		tag, err := c.tagResolver.LatestTag(image)
		if err != nil {
			c.log.Debug("tag resolution failed for %q: %v", image, err)
			continue
		}
		tagPinned := image[:idx] + ":" + tag
		delete(digests, image)
		digests[tagPinned] = hex
	}
}

func toJob(j *runpb.Job) Job {
	containers := j.GetTemplate().GetTemplate().GetContainers()
	digests := make(map[string]string, len(containers))
	for _, container := range containers {
		image := container.GetImage()
		digests[image] = parseDigest(image)
	}
	var createdAt time.Time
	if ts := j.GetCreateTime(); ts != nil {
		createdAt = ts.AsTime()
	}
	return Job{
		Name:      shortName(j.GetName()),
		Digests:   digests,
		CreatedAt: createdAt,
	}
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
	jobs      *run.JobsClient
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

func (g *gcpAPI) listJobs(ctx context.Context, project, region string) ([]*runpb.Job, error) {
	parent := fmt.Sprintf("projects/%s/locations/%s", project, region)
	it := g.jobs.ListJobs(ctx, &runpb.ListJobsRequest{Parent: parent})
	var out []*runpb.Job
	for {
		job, err := it.Next()
		if err == iterator.Done {
			return out, nil
		}
		if err != nil {
			return nil, fmt.Errorf("listing Cloud Run jobs in %s: %w", parent, err)
		}
		out = append(out, job)
	}
}
