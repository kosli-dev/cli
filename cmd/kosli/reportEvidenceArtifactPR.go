package main

import (
	"io"

	"github.com/spf13/cobra"
)

const reportEvidenceArtifactPRDesc = `All Kosli commands to report pull/merge request.`

func newReportEvidenceArtifactPRCmd(out io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "pullrequest",
		Aliases: []string{"pr", "mr", "mergerequest"},
		Short:   reportEvidenceArtifactPRDesc,
		Long:    reportEvidenceArtifactPRDesc,
	}

	// Add subcommands
	cmd.AddCommand(
		newReportEvidenceArtifactPRBitbucketCmd(out),
		newReportEvidenceArtifactPRGithubCmd(out),
		newReportEvidenceArtifactPRGitlabCmd(out),
		newReportEvidenceArtifactPRAzureCmd(out),
	)

	return cmd
}
