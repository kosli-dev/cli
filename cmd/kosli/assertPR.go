package main

import (
	"io"

	"github.com/spf13/cobra"
)

const assertPRDesc = `All Kosli pullrequests assertion commands. Return non-zero exit code if the assertion fails.`

const assertPRNoRepoDesc = `No local git repository is needed: ^--commit^ is looked up through the git provider's API.
Outside a supported CI, set ^--commit^ to a full commit SHA: its ^HEAD^ default is not read from your local checkout.`

func newAssertPRCmd(out io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "pullrequest",
		Aliases: []string{"pr", "pull_request", "mr", "mergerequest", "merge_request"},
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
