package main

import (
	"fmt"
	"io"
	"net/http"

	"github.com/kosli-dev/cli/internal/requests"
	"github.com/spf13/cobra"
)

const environmentDeclareDesc = `Declare a Kosli environment.`

const environmentDeclareExample = `
# declare a Kosli environment:
kosli environment declare 
	--name yourEnvironmentName \
	--environment-type K8S \
	--description "my new env" \
	--api-token yourAPIToken \
	--owner yourOrgName 
`

type CreateEnvironmentPayload struct {
	Owner       string `json:"owner"`
	Name        string `json:"name"`
	Type        string `json:"type"`
	Description string `json:"description"`
}

func newEnvironmentDeclareCmd(out io.Writer) *cobra.Command {
	payload := new(CreateEnvironmentPayload)
	cmd := &cobra.Command{
		Use:     "declare",
		Short:   environmentDeclareDesc,
		Long:    environmentDeclareDesc,
		Example: environmentDeclareExample,
		Args:    cobra.NoArgs,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			err := RequireGlobalFlags(global, []string{"Owner", "ApiToken"})
			if err != nil {
				return ErrorBeforePrintingUsage(cmd, err.Error())
			}

			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			payload.Owner = global.Owner
			url := fmt.Sprintf("%s/api/v1/environments/%s/", global.Host, global.Owner)

			reqParams := &requests.RequestParams{
				Method:   http.MethodPut,
				URL:      url,
				Payload:  payload,
				DryRun:   global.DryRun,
				Password: global.ApiToken,
			}
			_, err := kosliClient.Do(reqParams)
			if err == nil && !global.DryRun {
				logger.Info("environment %s was created", payload.Name)
			}
			return err
		},
	}

	cmd.Flags().StringVarP(&payload.Name, "name", "n", "", newEnvNameFlag)
	cmd.Flags().StringVarP(&payload.Type, "environment-type", "t", "", newEnvTypeFlag)
	cmd.Flags().StringVarP(&payload.Description, "description", "d", "", envDescriptionFlag)
	addDryRunFlag(cmd)

	err := RequireFlags(cmd, []string{"name", "environment-type"})
	if err != nil {
		logger.Error("failed to configure required flags: %v", err)
	}

	return cmd
}