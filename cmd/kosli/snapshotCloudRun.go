package main

import (
	"context"
	"io"
	"net/http"
	"net/url"

	"github.com/kosli-dev/cli/internal/cloudrun"
	"github.com/kosli-dev/cli/internal/filters"
	"github.com/kosli-dev/cli/internal/requests"
	"github.com/spf13/cobra"
)

const snapshotCloudRunShortDesc = `Report a snapshot of Cloud Run services and jobs in a Google Cloud project and region to Kosli.  `
const snapshotCloudRunLongDesc = snapshotCloudRunShortDesc + `
Coverage:

| Deploy method | Container? | API | Reported | Notes |
|---|---|---|---|---|
| Cloud Run service (image-deployed) | Yes | ^run^ | ✓ Full | Baseline. |
| Cloud Run service (source / Buildpacks) | Yes (built for you) | ^run^ | ✓ Full | Same API surface as image-deployed; same behaviour. |
| Cloud Run Job | Yes | ^run^ | ✓ Full | Jobs surface by resource (not per-Execution). Visible whether running or idle. |
| Cloud Run function (Cloud Functions 2nd gen) | Yes (Buildpacks) | ^cloudfunctions^ + ^run^ | ✓ Full | Surfaces as the backing Cloud Run service. Image path uses ^gcf-artifacts/...^ encoding. |
| Cloud Functions 1st gen | No (Google packages the source) | ^cloudfunctions^ only | ✗ | Legacy. Separate API; out of scope for this command. |
| App Engine Standard | No (gVisor sandbox, not a container) | ^appengine^ | ✗ | Different API; intentionally out of scope. |
| App Engine Flexible | Yes (containers on managed VMs) | ^appengine^ | ✗ | Mostly superseded by Cloud Run; out of scope. |
| GKE (Standard / Autopilot) | Yes | ^container^ + Kubernetes API | ✗ | Use ^kosli snapshot k8s^. |
| Cloud Run for Anthos | Yes (knative on GKE) | knative on the GKE cluster | ✗ | Niche; managed Cloud Run replaced this for most users. |
| Compute Engine + Container-Optimized OS | Yes (Docker on a VM) | ^compute^ | ✗ | Containers on VMs; out of scope. |

Each Cloud Run service contributes one artifact per revision in its traffic
configuration. Each Cloud Run Job contributes one artifact, identified by the
image bound to the Job (Jobs do not have a revision/traffic-split model).
Idle Jobs (no currently-running Execution) are included.

GCP authentication uses Application Default Credentials. On a developer
machine, run ^gcloud auth application-default login^; in GCE/GKE/Cloud Run
the metadata server / Workload Identity is used automatically. The caller
needs ^roles/run.viewer^ on the target project, plus
^roles/artifactregistry.reader^ on the Artifact Registry repository (or the
project) for digest and tag resolution on tag-pinned images. Missing the AR
role is non-fatal — tag-pinned artifacts then surface with empty digests.

Digest and tag resolution is scoped to Artifact Registry (^*-docker.pkg.dev^)
and the legacy Container Registry (^*.gcr.io^). Images from other registries
(Docker Hub, Quay, ECR, etc.) are reported as-is.

Skip all filtering flags to report every service and every job in the given
project + region. Use ^--include^ and/or ^--include-regex^ to snapshot only a
subset, OR ^--exclude^ and/or ^--exclude-regex^ to omit a subset; include and
exclude are mutually exclusive. Filters apply uniformly to both service and
job names and are case-sensitive.

Pass ^--resolve-names^ to rewrite digest-pinned Service artifact names back
to their deploy-time tags (commit SHA / version) via an Artifact Registry
reverse-lookup. Only supported for Artifact Registry hosts.`

const snapshotCloudRunExample = `
# report every Cloud Run service and job in a project + region:
kosli snapshot cloud-run yourEnvironmentName \
	--project yourGCPProject \
	--region yourGCPRegion \
	--api-token yourAPIToken \
	--org yourOrgName

# report only the named services and jobs:
kosli snapshot cloud-run yourEnvironmentName \
	--project yourGCPProject \
	--region yourGCPRegion \
	--include hello-world,sandman-job \
	--api-token yourAPIToken \
	--org yourOrgName

# report everything except the kosli-reporter job (the Job that runs this command):
kosli snapshot cloud-run yourEnvironmentName \
	--project yourGCPProject \
	--region yourGCPRegion \
	--exclude kosli-reporter \
	--api-token yourAPIToken \
	--org yourOrgName
`

// cloudRunLister is the seam between the command and the GCP client. Tests
// override newCloudRunClient with a stub that returns canned services and
// jobs. Both lookups use the same project/region.
type cloudRunLister interface {
	ListServices(ctx context.Context, project, region string) ([]cloudrun.Service, error)
	ListJobs(ctx context.Context, project, region string) ([]cloudrun.Job, error)
}

var newCloudRunClient = func(ctx context.Context, resolveNames bool) (cloudRunLister, error) {
	return cloudrun.New(ctx, logger, resolveNames)
}

type snapshotCloudRunOptions struct {
	project        string
	region         string
	resolveNames   bool
	resourceFilter *filters.ResourceFilterOptions
}

func newSnapshotCloudRunCmd(out io.Writer) *cobra.Command {
	o := new(snapshotCloudRunOptions)
	o.resourceFilter = new(filters.ResourceFilterOptions)
	cmd := &cobra.Command{
		Use:     "cloud-run ENVIRONMENT-NAME",
		Short:   snapshotCloudRunShortDesc,
		Long:    snapshotCloudRunLongDesc,
		Example: snapshotCloudRunExample,
		Args:    cobra.ExactArgs(1),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if err := RequireGlobalFlags(global, []string{"Org", "ApiToken"}); err != nil {
				return ErrorBeforePrintingUsage(cmd, err.Error())
			}
			for _, pair := range [][]string{
				{"include", "exclude"},
				{"include", "exclude-regex"},
				{"include-regex", "exclude"},
				{"include-regex", "exclude-regex"},
			} {
				if err := MuXRequiredFlags(cmd, pair, false); err != nil {
					return err
				}
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return o.run(args)
		},
	}

	cmd.Flags().StringVar(&o.project, "project", "", "[required] GCP project ID.")
	cmd.Flags().StringVar(&o.region, "region", "", "[required] GCP region (e.g. europe-west1).")
	cmd.Flags().BoolVar(&o.resolveNames, "resolve-names", false, cloudRunResolveNamesFlag)
	cmd.Flags().StringSliceVar(&o.resourceFilter.IncludeNames, "include", []string{}, cloudRunIncludeFlag)
	cmd.Flags().StringSliceVar(&o.resourceFilter.IncludeNamesRegex, "include-regex", []string{}, cloudRunIncludeRegexFlag)
	cmd.Flags().StringSliceVar(&o.resourceFilter.ExcludeNames, "exclude", []string{}, cloudRunExcludeFlag)
	cmd.Flags().StringSliceVar(&o.resourceFilter.ExcludeNamesRegex, "exclude-regex", []string{}, cloudRunExcludeRegexFlag)
	addDryRunFlag(cmd)

	if err := RequireFlags(cmd, []string{"project", "region"}); err != nil {
		logger.Error("failed to configure required flags: %v", err)
	}

	return cmd
}

func (o *snapshotCloudRunOptions) run(args []string) error {
	envName := args[0]
	reportURL, err := url.JoinPath(global.Host, "api/v2/environments", global.Org, envName, "report/cloud-run")
	if err != nil {
		return err
	}

	ctx := context.Background()
	client, err := newCloudRunClient(ctx, o.resolveNames)
	if err != nil {
		return err
	}
	if closer, ok := client.(io.Closer); ok {
		defer func() { _ = closer.Close() }()
	}
	services, err := client.ListServices(ctx, o.project, o.region)
	if err != nil {
		return cloudrun.Classify(err, o.project, o.region)
	}
	jobs, err := client.ListJobs(ctx, o.project, o.region)
	if err != nil {
		return cloudrun.Classify(err, o.project, o.region)
	}

	filteredServices := make([]cloudrun.Service, 0, len(services))
	for _, svc := range services {
		include, err := o.resourceFilter.ShouldInclude(svc.Name)
		if err != nil {
			return err
		}
		if include {
			filteredServices = append(filteredServices, svc)
		}
	}
	filteredJobs := make([]cloudrun.Job, 0, len(jobs))
	for _, job := range jobs {
		include, err := o.resourceFilter.ShouldInclude(job.Name)
		if err != nil {
			return err
		}
		if include {
			filteredJobs = append(filteredJobs, job)
		}
	}

	payload := cloudrun.ToEnvRequest(filteredServices, filteredJobs)

	reqParams := &requests.RequestParams{
		Method:  http.MethodPut,
		URL:     reportURL,
		Payload: payload,
		DryRun:  global.DryRun,
		Token:   global.ApiToken,
	}
	_, err = kosliClient.Do(reqParams)
	if err == nil && !global.DryRun {
		logger.Info("[%d] artifacts were reported to environment %s", len(payload.Artifacts), envName)
	}
	return err
}
