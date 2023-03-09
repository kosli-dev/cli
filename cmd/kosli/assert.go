package main

import (
	"io"

	"github.com/spf13/cobra"
)

const assertDesc = `All Kosli assertion commands. Return non-zero exit code if the assertion fails.`

func newAssertCmd(out io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "assert",
		Short: assertDesc,
		Long:  assertDesc,
	}

	// Add subcommands
	cmd.AddCommand(
		newAssertPRCmd(out),
		newAssertStatusCmd(out),
		newAssertArtifactCmd(out),
		newAssertSnapshotCmd(out),
		newAssertApprovalCmd(out),
	)

	return cmd
}
