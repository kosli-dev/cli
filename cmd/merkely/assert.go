package main

import (
	"io"

	"github.com/spf13/cobra"
)

const assertDesc = `All Merkely assertion commands. Return non-zero exit code if the assertion fails.`

func newAssertCmd(out io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "assert",
		Short: assertDesc,
		Long:  assertDesc,
	}

	// Add subcommands
	cmd.AddCommand(
		newAssertPullRequestBitbucketCmd(out),
		newAssertPullRequestGithubCmd(out),
		newAssertStatusCmd(out),
	)

	return cmd
}
