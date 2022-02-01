package main

import (
	"io"

	"github.com/spf13/cobra"
)

const allowedArtifactsDesc = `All Merkely allowedartifacts policy operations.`

func newAllowedArtifactsCmd(out io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "allowedartifacts",
		Aliases: []string{"allowed-artifacts"},
		Short:   allowedArtifactsDesc,
		Long:    allowedArtifactsDesc,
	}

	// Add subcommands
	cmd.AddCommand(
		newAllowedArtifactsCreateCmd(out),
	)

	return cmd
}
