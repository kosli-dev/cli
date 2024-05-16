package main

import (
	"io"

	"github.com/spf13/cobra"
)

const beginDesc = `All Kosli begin commands.`

func newBeginCmd(out io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "begin",
		Aliases: []string{"start", "init"},
		Short:   beginDesc,
		Long:    beginDesc,
	}

	// Add subcommands
	cmd.AddCommand(
		newBeginTrailCmd(out),
	)
	return cmd
}
