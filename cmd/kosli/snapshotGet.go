package main

import (
	"io"

	"github.com/spf13/cobra"
)

const snapshotGetDesc = `Get a specific environment snapshot.`

type snapshotGetOptions struct {
	output string
}

const snapshotGetExample = `
# get the latest snapshot of an environment:
kosli snapshot get yourEnvironmentName
	--api-token yourAPIToken \
	--owner yourOrgName 

# get the SECOND latest snapshot of an environment:
kosli snapshot get yourEnvironmentName~1
	--api-token yourAPIToken \
	--owner yourOrgName 

# get the snapshot number 23 of an environment:
kosli snapshot get yourEnvironmentName#23
	--api-token yourAPIToken \
	--owner yourOrgName 
`

func newSnapshotGetCmd(out io.Writer) *cobra.Command {
	o := new(snapshotGetOptions)
	cmd := &cobra.Command{
		Use:     "get ENVIRONMENT-NAME-OR-EXPRESSION",
		Short:   snapshotGetDesc,
		Long:    snapshotGetDesc,
		Example: snapshotGetExample,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			err := RequireGlobalFlags(global, []string{"Owner", "ApiToken"})
			if err != nil {
				return ErrorBeforePrintingUsage(cmd, err.Error())
			}
			if len(args) < 1 {
				return ErrorBeforePrintingUsage(cmd, "environment name/expression argument is required")
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return o.run(out, args)
		},
	}

	cmd.Flags().StringVarP(&o.output, "output", "o", "table", outputFlag)

	return cmd
}

func (o *snapshotGetOptions) run(out io.Writer, args []string) error {
	return getSnapshot(out, o, args)
}
