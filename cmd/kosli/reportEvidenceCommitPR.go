package main

import (
	"io"

	"github.com/spf13/cobra"
)

const reportEvidenceCommitPRDesc = `All Kosli commands to report pull/merge request commands.`

func newReportEvidenceCommitPRCmd(out io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:        "pullrequest",
		Aliases:    []string{"pr", "mr", "mergerequest"},
		Short:      reportEvidenceCommitPRDesc,
		Long:       reportEvidenceCommitPRDesc,
		Deprecated: "see kosli attest commands",
	}

	// Add subcommands
	cmd.AddCommand(
		newReportEvidenceCommitPRBitbucketCmd(out),
		newReportEvidenceCommitPRGithubCmd(out),
		newReportEvidenceCommitPRGitlabCmd(out),
		newReportEvidenceCommitPRAzureCmd(out),
	)

	return cmd
}
