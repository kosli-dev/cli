package main

import (
	"io"

	"github.com/spf13/cobra"
)

const deploymentReadDesc = `All deployment read operations.`

func newDeploymentReadCmd(out io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "deployment",
		Short: deploymentReadDesc,
		Long:  deploymentReadDesc,
	}

	// Add subcommands
	cmd.AddCommand(
		newDeploymentLsCmd(out),
	)

	return cmd
}
