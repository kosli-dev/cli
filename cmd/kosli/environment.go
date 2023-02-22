package main

import (
	"io"

	"github.com/spf13/cobra"
)

const environmentDesc = `All environment operations in Kosli.`

func newEnvironmentCmd(out io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "environment",
		Aliases: []string{"env"},
		Short:   environmentDesc,
		Long:    environmentDesc,
	}

	// Add subcommands
	cmd.AddCommand(
		newEnvironmentReportCmd(out),
	)

	return cmd
}
