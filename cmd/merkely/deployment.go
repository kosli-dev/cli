package main

import (
	"io"

	"github.com/spf13/cobra"
)

const deploymentDesc = `All deployment operations in a Merkely pipeline.`

func newDeploymentCmd(out io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "deployment",
		Short: deploymentDesc,
		Long:  deploymentDesc,
	}

	// Add subcommands
	cmd.AddCommand(
		newDeploymentReportCmd(out),
	)

	return cmd
}
