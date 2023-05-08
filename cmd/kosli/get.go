package main

import (
	"io"

	"github.com/spf13/cobra"
)

const getDesc = `All Kosli get commands.`

func newGetCmd(out io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get",
		Short: getDesc,
		Long:  getDesc,
	}

	// Add subcommands
	cmd.AddCommand(
		newGetApprovalCmd(out),
		newGetArtifactCmd(out),
		newGetDeploymentCmd(out),
		newGetEnvironmentCmd(out),
		newGetFlowCmd(out),
		newGetSnapshotCmd(out),
		newGetAuditTrailCmd(out),
		newGetWorkflowCmd(out),
	)
	return cmd
}
