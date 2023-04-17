package main

import (
	"io"

	"github.com/spf13/cobra"
)

const assertPRDesc = `All Kosli pullrequests assertion commands. Return non-zero exit code if the assertion fails.`

func newAssertPRCmd(out io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "pullrequest",
		Aliases: []string{"pr", "mergerequest", "mr"},
		Short:   assertPRDesc,
		Long:    assertPRDesc,
	}

	// Add subcommands
	cmd.AddCommand(
		newAssertPullRequestBitbucketCmd(out),
		newAssertPullRequestGithubCmd(out),
		newAssertPullRequestGitlabCmd(out),
		newAssertPullRequestAzureCmd(out),
	)

	return cmd
}
