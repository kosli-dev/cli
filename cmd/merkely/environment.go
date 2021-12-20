package main

import (
	"io"

	"github.com/spf13/cobra"
)

const environmentDesc = `All environments operations in Merkely.`

func newEnvironmentCmd(out io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "environment",
		Short: environmentDesc,
		Long:  environmentDesc,
	}

	// Add subcommands
	cmd.AddCommand(
		newEnvironmentDeclareCmd(out),
		newEnvironmentReportCmd(out),
	)

	return cmd
}
