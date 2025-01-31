package main

import (
	"io"

	"github.com/spf13/cobra"
)

const archiveDesc = `All Kosli archive commands.`

func newArchiveCmd(out io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "archive",
		Short: archiveDesc,
		Long:  archiveDesc,
	}

	// Add subcommands
	cmd.AddCommand(
		newArchiveFlowCmd(out),
		newArchiveEnvironmentCmd(out),
		newArchiveAttestationTypeCmd(out),
	)
	return cmd
}
