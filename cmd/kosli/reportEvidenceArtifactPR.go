package main

import (
	"io"

	"github.com/spf13/cobra"
)

const reportEvidenceArtifactPRDesc = `All Kosli evidence commands.`

func newReportEvidenceArtifactPRCmd(out io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "pullrequest",
		Aliases: []string{"pr", "mr", "mergerequest"},
		Short:   reportEvidenceArtifactDesc,
		Long:    reportEvidenceArtifactDesc,
	}

	// Add subcommands
	cmd.AddCommand(
		newPullRequestEvidenceBitbucketCmd(out),
		newPullRequestEvidenceGithubCmd(out),
		newPullRequestEvidenceGitlabCmd(out),
	)

	return cmd
}
