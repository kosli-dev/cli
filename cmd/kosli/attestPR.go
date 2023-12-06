package main

import (
	"io"

	"github.com/spf13/cobra"
)

const attestPRDesc = `All Kosli commands to attest pull/merge request.`

func newAttestPRCmd(out io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "pullrequest",
		Aliases: []string{"pr", "mr", "mergerequest"},
		Short:   attestPRDesc,
		Long:    attestPRDesc,
		Hidden:  true,
	}

	// Add subcommands
	cmd.AddCommand(
		newAttestGitlabPRCmd(out),
		newAttestGithubPRCmd(out),
		newAttestBitbucketPRCmd(out),
		newAttestAzurePRCmd(out),
	)

	return cmd
}
