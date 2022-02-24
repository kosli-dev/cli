package main

import (
	"fmt"
	"io"
	"net/http"

	"github.com/merkely-development/reporter/internal/requests"
	"github.com/spf13/cobra"
)

const enviromentDeclareDesc = `
Declare or update a Merkely environment.
`

const enviromentDeclareExample = `
# declare (or update) a Merkely environment:
merkely environment declare 
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
		Short:   "Declare or update a Merkely environment",
		Long:    enviromentDeclareDesc,
		Example: enviromentDeclareExample,
		Args:    NoArgs,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			err := RequireGlobalFlags(global, []string{"Owner", "ApiToken"})
			if err != nil {
				return ErrorAfterPrintingHelp(cmd, err.Error())
			}

			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if payload.Type != "ECS" && payload.Type != "K8S" && payload.Type != "server" && payload.Type != "S3" {
				return fmt.Errorf("%s is not a valid environment type", payload.Type)
			}
			payload.Owner = global.Owner
			url := fmt.Sprintf("%s/api/v1/environments/%s/", global.Host, global.Owner)

			_, err := requests.SendPayload(payload, url, "", global.ApiToken,
				global.MaxAPIRetries, global.DryRun, http.MethodPut, log)
			return err
		},
	}

	cmd.Flags().StringVarP(&payload.Name, "name", "n", "", "The name of environment.")
	cmd.Flags().StringVarP(&payload.Type, "environment-type", "t", "", "The type of environment. Valid options are: [K8S, ECS, server, S3]")
	cmd.Flags().StringVarP(&payload.Description, "description", "d", "", "[optional] The environment description.")

	err := RequireFlags(cmd, []string{"name", "environment-type"})
	if err != nil {
		log.Fatalf("failed to configure required flags: %v", err)
	}

	return cmd
}
