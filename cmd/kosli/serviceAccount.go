package main

import (
	"io"

	"github.com/spf13/cobra"
)

const serviceAccountDesc = `All Kosli service account operations.`

func newServiceAccountCmd(out io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "service-account",
		Aliases: []string{"sa"},
		Short:   serviceAccountDesc,
		Long:    serviceAccountDesc,
	}

	// Add subcommands
	cmd.AddCommand(
		newServiceAccountApiKeysCmd(out),
	)

	return cmd
}
