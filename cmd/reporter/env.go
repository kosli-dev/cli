package main

import (
	"io"

	"github.com/spf13/cobra"
)

const envDesc = `
Report actual deployments in an environment back to Merkely.
This allows Merkely to determine Runtime compliance status of the environment.
`

func newEnvCmd(out io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:               "env",
		Short:             "Report running artifacts in an environment to Merkely.",
		Long:              envDesc,
		DisableAutoGenTag: true,
		Aliases:           []string{"environment"},
	}

	// Add subcommands
	cmd.AddCommand(
		newK8sEnvCmd(out),
		newEcsEnvCmd(out),
		newServerEnvCmd(out),
	)

	return cmd
}
