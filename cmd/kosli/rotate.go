package main

import (
	"io"

	"github.com/spf13/cobra"
)

const rotateDesc = `All Kosli rotate commands.`

func newRotateCmd(out io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "rotate",
		Aliases: []string{"ro"},
		Short:   rotateDesc,
		Long:    rotateDesc,
	}

	// Add subcommands
	cmd.AddCommand(
		newRotateApiKeyCmd(out),
	)

	return cmd
}
