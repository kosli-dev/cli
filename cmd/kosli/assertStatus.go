package main

import (
	"io"

	"github.com/spf13/cobra"
)

const assertStatusShortDesc = `Assert the status of a Kosli server.`

const assertStatusLongDesc = assertStatusShortDesc + `
Exits with non-zero code if the Kosli server down.`

func newAssertStatusCmd(out io.Writer) *cobra.Command {
	o := &statusOptions{assert: true}
	cmd := &cobra.Command{
		Use:   "status",
		Short: assertStatusShortDesc,
		Long:  assertStatusLongDesc,
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return o.run(out)
		},
	}
	return cmd
}
