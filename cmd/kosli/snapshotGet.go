package main

import (
	"io"

	"github.com/spf13/cobra"
)

const snapshotGetDesc = `Get a specific environment snapshot.`

type snapshotGetOptions struct {
	json bool
}

func newSnapshotGetCmd(out io.Writer) *cobra.Command {
	o := new(snapshotGetOptions)
	cmd := &cobra.Command{
		Use:   "get ENVIRONMENT-NAME",
		Short: snapshotGetDesc,
		Long:  snapshotGetDesc,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			err := RequireGlobalFlags(global, []string{"Owner", "ApiToken"})
			if err != nil {
				return ErrorBeforePrintingUsage(cmd, err.Error())
			}
			if len(args) < 1 {
				return ErrorBeforePrintingUsage(cmd, "environment name argument is required")
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return o.run(out, args)
		},
	}

	cmd.Flags().BoolVarP(&o.json, "json", "j", false, environmentJsonFlag)

	return cmd
}

func (o *snapshotGetOptions) run(out io.Writer, args []string) error {
	return getSnapshot(out, o, args)
}
