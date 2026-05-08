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
Each Cloud Run service contributes one artifact per revision in its traffic
configuration. Each Cloud Run Job contributes one artifact, identified by the
image bound to the Job (Jobs do not have a revision/traffic-split model).
Idle Jobs (no currently-running Execution) are included.

GCP authentication uses Application Default Credentials. On a developer
machine, run ^gcloud auth application-default login^; in GCE/GKE/Cloud Run
the metadata server / Workload Identity is used automatically. The caller
needs at least ^roles/run.viewer^ on the target project.

Skip all filtering flags to report every service and every job in the given
project + region. Use ^--include^ and/or ^--include-regex^ to snapshot only a
subset, OR ^--exclude^ and/or ^--exclude-regex^ to omit a subset; include and
exclude are mutually exclusive. Filters apply uniformly to both service and
job names and are case-sensitive.

Currently a hidden, in-development command. Use --dry-run to inspect the payload without sending it to Kosli.`

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

var newCloudRunClient = func(ctx context.Context) (cloudRunLister, error) {
	return cloudrun.New(ctx)
}

type snapshotCloudRunOptions struct {
	project        string
	region         string
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
		Hidden:  true,
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
	client, err := newCloudRunClient(ctx)
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
