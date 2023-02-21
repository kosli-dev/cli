package main

import (
	"io"

	"github.com/spf13/cobra"
)

type environmentGetOptions struct {
	output string
}

const getSnapshotDescShort = `Get a specific environment snapshot.`

const getSnapshotDesc = getSnapshotDescShort + `
Specify SNAPPISH by:
	environmentName~<N>  N'th behind the latest snapshot
	environmentName#<N>  snapshot number N
	environmentName      the latest snapshot`

const getSnapshotExample = `
# get the latest snapshot of an environment:
kosli get snapshot yourEnvironmentName
	--api-token yourAPIToken \
	--owner yourOrgName 

# get the SECOND latest snapshot of an environment:
kosli get snapshot yourEnvironmentName~1
	--api-token yourAPIToken \
	--owner yourOrgName 

# get the snapshot number 23 of an environment:
kosli get snapshot yourEnvironmentName#23
	--api-token yourAPIToken \
	--owner yourOrgName `

func newGetSnapshotCmd(out io.Writer) *cobra.Command {
	o := new(environmentGetOptions)
	cmd := &cobra.Command{
		Use:     "snapshot ENVIRONMENT-NAME-OR-EXPRESSION",
		Short:   getSnapshotDescShort,
		Long:    getSnapshotDesc,
		Example: getSnapshotExample,
		Args:    cobra.ExactArgs(1),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			err := RequireGlobalFlags(global, []string{"Owner", "ApiToken"})
			if err != nil {
				return ErrorBeforePrintingUsage(cmd, err.Error())
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

func (o *environmentGetOptions) run(out io.Writer, args []string) error {
	return getSnapshot(out, o, args)
}
