package main

import (
	"fmt"
	"io"
	"net/http"

	"github.com/kosli-dev/cli/internal/requests"
	"github.com/spf13/cobra"
)

const addEnvironmentShortDesc = `Add a physical environment to a logical environment.`

const addEnvironmentLongDesc = addEnvironmentShortDesc + `
Add a physical Kosli environment to a logical Kosli environment.
`

const addEnvironmentExample = `
# add a physical environment to a logical environment:
kosli add environment \
	--physical prod-k8 \
	--logical prod \
	--api-token yourAPIToken \
	--org yourOrgName 
`

type addEnvironmentOptions struct {
	payload AddEnvironmentPayload
	logical string
}

type AddEnvironmentPayload struct {
	Pysical string `json:"physical_env_name"`
}

func newAddEnvironmentCmd(out io.Writer) *cobra.Command {
	o := new(addEnvironmentOptions)
	cmd := &cobra.Command{
		Use:     "environment",
		Aliases: []string{"env"},
		Short:   addEnvironmentShortDesc,
		Long:    addEnvironmentLongDesc,
		Example: addEnvironmentExample,
		Args:    cobra.ExactArgs(0),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			err := RequireGlobalFlags(global, []string{"Org", "ApiToken"})
			if err != nil {
				return ErrorBeforePrintingUsage(cmd, err.Error())
			}

			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			url := fmt.Sprintf("%s/api/v2/environments/%s/%s/add-physical-env-to-logical", global.Host, global.Org, o.logical)

			reqParams := &requests.RequestParams{
				Method:   http.MethodPut,
				URL:      url,
				Payload:  o.payload,
				DryRun:   global.DryRun,
				Password: global.ApiToken,
			}
			_, err := kosliClient.Do(reqParams)
			if err == nil && !global.DryRun {
				logger.Info("environment '%s' was added to '%s'", o.payload.Pysical, o.logical)
			}
			return err
		},
	}

	cmd.Flags().StringVar(&o.payload.Pysical, "physical", "", physicalEnvFlag)
	cmd.Flags().StringVar(&o.logical, "logical", "", logicalEnvFlag)
	err := RequireFlags(cmd, []string{"physical", "logical"})
	if err != nil {
		logger.Error("failed to configure required flags: %v", err)
	}

	addDryRunFlag(cmd)
	return cmd
}
