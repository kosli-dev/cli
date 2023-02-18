package main

import (
	"io"

	"github.com/spf13/cobra"
)

const reportEvidenceArtifactPRDesc = `All Kosli commands to report pull/merge request commands.`

func newReportEvidenceArtifactPRCmd(out io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "pullrequest",
		Aliases: []string{"pr", "mr", "mergerequest"},
		Short:   reportEvidenceArtifactPRDesc,
		Long:    reportEvidenceArtifactPRDesc,
	}

	// Add subcommands
	cmd.AddCommand(
		newPullRequestEvidenceBitbucketCmd(out),
		newPullRequestEvidenceGithubCmd(out),
		newPullRequestEvidenceGitlabCmd(out),
	)

	return cmd
}
