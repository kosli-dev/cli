package main

import (
	"fmt"
	"io"
	"net/http"

	"github.com/merkely-development/reporter/internal/requests"
	"github.com/spf13/cobra"
)

const createEnvDesc = `
Create a Merkely environment.
`

const createEnvExample = `
* create a Merkely environment:
merkely create environment --api-token 1234 --owner test --name newEnv --type K8S --description "my new env"
`

type CreateEnvironmentPayload struct {
	Owner       string `json:"owner"`
	Name        string `json:"name"`
	Type        string `json:"type"`
	Description string `json:"description"`
}

func newEnvironmentCmd(out io.Writer) *cobra.Command {
	payload := new(CreateEnvironmentPayload)
	cmd := &cobra.Command{
		Use:               "environment",
		Short:             "Create a Merkely environment",
		Long:              createEnvDesc,
		Example:           createEnvExample,
		DisableAutoGenTag: true,
		Args:              NoArgs,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			err := RequireGlobalFlags(global, []string{"Owner", "ApiToken"})
			if err != nil {
				return err
			}

			if payload.Type != "ECS" && payload.Type != "K8S" && payload.Type != "server" {
				return fmt.Errorf("%s is not a valid environment type", payload.Type)
			}

			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			payload.Owner = global.Owner
			url := fmt.Sprintf("%s/api/v1/environments/%s/", global.Host, global.Owner)

			_, err := requests.SendPayload(payload, url, "", global.ApiToken,
				global.MaxAPIRetries, global.DryRun, http.MethodPut, log)
			return err
		},
	}

	cmd.Flags().StringVarP(&payload.Name, "name", "n", "", "The name of environment.")
	cmd.Flags().StringVarP(&payload.Type, "type", "t", "", "The type of environment. Valid options are: [K8S, ECS, server]")
	cmd.Flags().StringVarP(&payload.Description, "description", "d", "", "[optional] The environment description.")

	err := RequireFlags(cmd, []string{"name", "type"})
	if err != nil {
		log.Fatalf("failed to configure required flags: %v", err)
	}

	return cmd
}
