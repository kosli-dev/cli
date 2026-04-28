package main

import (
	"context"
	"io"
	"net/http"
	"net/url"

	"github.com/kosli-dev/cli/internal/cloudrun"
	"github.com/kosli-dev/cli/internal/requests"
	"github.com/spf13/cobra"
)

const snapshotCloudRunShortDesc = `Report a snapshot of running services in a Google Cloud Run project and region to Kosli.  `
const snapshotCloudRunLongDesc = snapshotCloudRunShortDesc + `
Currently a hidden, in-development command — it always runs in dry-run mode regardless of the --dry-run flag.`

// cloudRunLister is the seam between the command and the GCP client. Tests
// override newCloudRunClient with a stub that returns canned services.
type cloudRunLister interface {
	ListServices(ctx context.Context, project, region string) ([]cloudrun.Service, error)
}

var newCloudRunClient = func(ctx context.Context) (cloudRunLister, error) {
	return cloudrun.New(ctx)
}

type snapshotCloudRunOptions struct {
	project string
	region  string
}

func newSnapshotCloudRunCmd(out io.Writer) *cobra.Command {
	o := new(snapshotCloudRunOptions)
	cmd := &cobra.Command{
		Use:    "cloud-run ENVIRONMENT-NAME",
		Short:  snapshotCloudRunShortDesc,
		Long:   snapshotCloudRunLongDesc,
		Hidden: true,
		Args:   cobra.ExactArgs(1),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if err := RequireGlobalFlags(global, []string{"Org", "ApiToken"}); err != nil {
				return ErrorBeforePrintingUsage(cmd, err.Error())
			}
			global.DryRun = true
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return o.run(args)
		},
	}

	cmd.Flags().StringVar(&o.project, "project", "", "[required] GCP project ID.")
	cmd.Flags().StringVar(&o.region, "region", "", "[required] GCP region (e.g. europe-west1).")
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
	services, err := client.ListServices(ctx, o.project, o.region)
	if err != nil {
		return err
	}

	payload := cloudrun.ToEnvRequest(services, o.project, o.region)

	reqParams := &requests.RequestParams{
		Method:  http.MethodPut,
		URL:     reportURL,
		Payload: payload,
		DryRun:  global.DryRun,
		Token:   global.ApiToken,
	}
	_, err = kosliClient.Do(reqParams)
	return err
}
