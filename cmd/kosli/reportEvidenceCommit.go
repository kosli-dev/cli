package main

import (
	"io"

	"github.com/spf13/cobra"
)

const reportEvidenceCommitDesc = `All Kosli commit commands.`

func newReportEvidenceCommitCmd(out io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "commit",
		Short: reportEvidenceCommitDesc,
		Long:  reportEvidenceCommitDesc,
	}

	// Add subcommands
	cmd.AddCommand(
		newReportEvidenceCommitPRCmd(out),
		newReportEvidenceCommitGenericCmd(out),
		newReportEvidenceCommitJunitCmd(out),
		newReportEvidenceCommitSnykCmd(out),
		newReportEvidenceCommitJiraCmd(out),
	)

	return cmd
}
