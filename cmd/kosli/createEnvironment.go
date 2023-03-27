package main

import (
	"fmt"
	"io"
	"net/http"

	"github.com/kosli-dev/cli/internal/requests"
	"github.com/spf13/cobra"
)

const createEnvironmentDesc = `Create a Kosli environment.`

const createEnvironmentExample = `
# create a Kosli environment:
kosli create environment yourEnvironmentName
	--type K8S \
	--description "my new env" \
	--api-token yourAPIToken \
	--org yourOrgName 
`

type createEnvOptions struct {
	payload CreateEnvironmentPayload
}

type CreateEnvironmentPayload struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	Description string `json:"description"`
}

func newCreateEnvironmentCmd(out io.Writer) *cobra.Command {
	o := new(createEnvOptions)
	cmd := &cobra.Command{
		Use:     "environment ENVIRONMENT-NAME",
		Aliases: []string{"env"},
		Short:   createEnvironmentDesc,
		Long:    createEnvironmentDesc,
		Example: createEnvironmentExample,
		Args:    cobra.ExactArgs(1),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			err := RequireGlobalFlags(global, []string{"Org", "ApiToken"})
			if err != nil {
				return ErrorBeforePrintingUsage(cmd, err.Error())
			}

			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return o.run(args)
		},
	}

	cmd.Flags().StringVarP(&o.payload.Type, "type", "t", "", newEnvTypeFlag)
	cmd.Flags().StringVarP(&o.payload.Description, "description", "d", "", envDescriptionFlag)
	addDryRunFlag(cmd)

	err := RequireFlags(cmd, []string{"type"})
	if err != nil {
		logger.Error("failed to configure required flags: %v", err)
	}

	return cmd
}

func (o *createEnvOptions) run(args []string) error {
	o.payload.Name = args[0]
	url := fmt.Sprintf("%s/api/v2/environments/%s", global.Host, global.Org)

	reqParams := &requests.RequestParams{
		Method:   http.MethodPut,
		URL:      url,
		Payload:  o.payload,
		DryRun:   global.DryRun,
		Password: global.ApiToken,
	}
	_, err := kosliClient.Do(reqParams)
	if err == nil && !global.DryRun {
		logger.Info("environment %s was created", o.payload.Name)
	}
	return err
}
