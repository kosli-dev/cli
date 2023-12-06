package main

import (
	"io"

	"github.com/spf13/cobra"
)

const attestDesc = `All Kosli attest commands.`

func newAttestCmd(out io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:    "attest",
		Short:  attestDesc,
		Long:   attestDesc,
		Hidden: true,
	}

	// Add subcommands
	cmd.AddCommand(
		newAttestArtifactCmd(out),
		newAttestGenericCmd(out),
		newAttestSnykCmd(out),
		newAttestJunitCmd(out),
		newAttestJiraCmd(out),
		newAttestPRCmd(out),
	)
	return cmd
}
