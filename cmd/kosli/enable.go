package main

import (
	"io"

	"github.com/spf13/cobra"
)

const enableDesc = `Kosli enable commands.`

func newEnableCmd(out io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "enable",
		Short: enableDesc,
		Long:  enableDesc,
	}

	// Add subcommands
	cmd.AddCommand(
		newEnableBetaCmd(out),
	)

	return cmd
}
