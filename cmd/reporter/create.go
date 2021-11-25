package main

import (
	"io"

	"github.com/spf13/cobra"
)

const createDesc = `Create objects in Merkely.`

func newCreateCmd(out io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:               "create",
		Short:             createDesc,
		Long:              createDesc,
		DisableAutoGenTag: true,
	}

	// Add subcommands
	cmd.AddCommand(
		newPipelineCmd(out),
		newEnvironmentCmd(out),
	)

	return cmd
}
