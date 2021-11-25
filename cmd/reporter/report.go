package main

import (
	"io"

	"github.com/spf13/cobra"
)

const reportDesc = `
Report compliance events back to Merkely.
`

func newReportCmd(out io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:               "report",
		Short:             "Report compliance events to Merkely.",
		Long:              reportDesc,
		DisableAutoGenTag: true,
		Aliases:           []string{"log"},
		//SuggestFor: []string{"reportenv", "env report", "envreport"},
	}

	// Add subcommands
	cmd.AddCommand(
		newEnvCmd(out),
		newArtifactCmd(out),
		newDeploymentCmd(out),
		newEvidenceCmd(out),
		newTestEvidenceCmd(out),
		newApproveDeploymentCmd(out),
		newRequestApprovalCmd(out),
	)

	return cmd
}
