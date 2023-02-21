package main

import (
	"io"

	"github.com/spf13/cobra"
)

const allowDesc = `All Kosli allow commands.`

func newAllowCmd(out io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "allow",
		Short: allowDesc,
		Long:  allowDesc,
	}

	// Add subcommands
	cmd.AddCommand(
		newAllowArtifactCmd(out),
	)

	return cmd
}
