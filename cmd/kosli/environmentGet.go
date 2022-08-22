package main

import (
	"io"

	"github.com/spf13/cobra"
)

const environmentGetDesc = `Get a specific environment snapshot.`

type environmentGetOptions struct {
	output string
}

const environmentGetExample = `
# get the latest snapshot of an environment:
kosli environment get yourEnvironmentName
	--api-token yourAPIToken \
	--owner yourOrgName 

# get the SECOND latest snapshot of an environment:
kosli environment get yourEnvironmentName~1
	--api-token yourAPIToken \
	--owner yourOrgName 

# get the snapshot number 23 of an environment:
kosli environment get yourEnvironmentName#23
	--api-token yourAPIToken \
	--owner yourOrgName 
`

func newEnvironmentGetCmd(out io.Writer) *cobra.Command {
	o := new(environmentGetOptions)
	cmd := &cobra.Command{
		Use:     "get ENVIRONMENT-NAME-OR-EXPRESSION",
		Short:   environmentGetDesc,
		Long:    environmentGetDesc,
		Example: environmentGetExample,
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

func (o *environmentGetOptions) run(out io.Writer, args []string) error {
	return getSnapshot(out, o, args)
}
