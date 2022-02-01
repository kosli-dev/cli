package main

import (
	"io"

	"github.com/spf13/cobra"
)

const policyDesc = `All Merkely policies operations.`

func newPolicyCmd(out io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "policy",
		Short: policyDesc,
		Long:  policyDesc,
	}

	// Add subcommands
	cmd.AddCommand(
		newAllowedArtifactsCmd(out),
	)

	return cmd
}
