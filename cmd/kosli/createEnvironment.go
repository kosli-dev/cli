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
kosli create environment 
	--name yourEnvironmentName \
	--environment-type K8S \
	--description "my new env" \
	--api-token yourAPIToken \
	--owner yourOrgName 
`

type createEnvOptions struct {
	payload CreateEnvironmentPayload
}

type CreateEnvironmentPayload struct {
	Owner       string `json:"owner"`
	Name        string `json:"name"`
	Type        string `json:"type"`
	Description string `json:"description"`
}

func newCreateEnvironmentCmd(out io.Writer) *cobra.Command {
	o := new(createEnvOptions)
	cmd := &cobra.Command{
		Use:     "environment",
		Aliases: []string{"env"},
		Short:   createEnvironmentDesc,
		Long:    createEnvironmentDesc,
		Example: createEnvironmentExample,
		Args:    cobra.NoArgs,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			err := RequireGlobalFlags(global, []string{"Owner", "ApiToken"})
			if err != nil {
				return ErrorBeforePrintingUsage(cmd, err.Error())
			}

			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return o.run()
		},
	}

	cmd.Flags().StringVarP(&o.payload.Name, "name", "n", "", newEnvNameFlag)
	cmd.Flags().StringVarP(&o.payload.Type, "environment-type", "t", "", newEnvTypeFlag)
	cmd.Flags().StringVarP(&o.payload.Description, "description", "d", "", envDescriptionFlag)
	addDryRunFlag(cmd)

	err := RequireFlags(cmd, []string{"name", "environment-type"})
	if err != nil {
		logger.Error("failed to configure required flags: %v", err)
	}

	return cmd
}

func (o *createEnvOptions) run() error {
	o.payload.Owner = global.Owner
	url := fmt.Sprintf("%s/api/v1/environments/%s/", global.Host, global.Owner)

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
