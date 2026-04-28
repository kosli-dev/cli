package main

import (
	"fmt"
	"io"

	"github.com/spf13/cobra"
)

const snapshotCloudRunShortDesc = `Report a snapshot of running services in a Google Cloud Run project and region to Kosli.  `
const snapshotCloudRunLongDesc = snapshotCloudRunShortDesc + `
Currently a hidden, in-development command — it always runs in dry-run mode and does not yet talk to GCP or to Kosli.`

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
			return o.run(out, args)
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

func (o *snapshotCloudRunOptions) run(out io.Writer, args []string) error {
	_, err := fmt.Fprintln(out, "cloud-run snapshot: not yet implemented (forced dry-run)")
	return err
}
