package main

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/merkely-development/reporter/internal/requests"
	"github.com/merkely-development/reporter/internal/server"
	"github.com/spf13/cobra"
)

const serverEnvDesc = `
List the artifacts deployed in a server environment and their digests 
and report them to Merkely. 
`

const serverEnvExample = `
* report directory artifacts running in a server at a list of paths:
merkely report env server prod --api-token 1234 --owner exampleOrg --id prod-server --paths a/b/c, e/f/g
`

type serverEnvOptions struct {
	paths []string
	id    string
}

func newServerEnvCmd(out io.Writer) *cobra.Command {
	o := new(serverEnvOptions)
	cmd := &cobra.Command{
		Use:     "server [-p /path/of/artifacts/directory] [-i infrastructure-identifier] env-name",
		Short:   "Report directory artifacts data in the given list of paths to Merkely.",
		Long:    serverEnvDesc,
		Aliases: []string{"directories"},
		Example: serverEnvExample,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if len(args) > 1 {
				return fmt.Errorf("only environment name argument is allowed")
			}
			if len(args) == 0 || args[0] == "" {
				return fmt.Errorf("environment name is required")
			}

			err := RequireGlobalFlags(global, []string{"Owner", "ApiToken"})
			if err != nil {
				return err
			}

			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			envName := args[0]
			if o.id == "" {
				o.id = envName
			}

			url := fmt.Sprintf("%s/api/v1/environments/%s/%s/data", global.Host, global.Owner, envName)

			artifacts, err := server.CreateServerArtifactsData(o.paths, log)
			if err != nil {
				return err
			}
			requestBody := &requests.ServerEnvRequest{
				Artifacts: artifacts,
				Type:      "server",
				Id:        o.id,
			}
			js, _ := json.MarshalIndent(requestBody, "", "    ")

			return requests.SendPayload(js, url, global.ApiToken,
				global.MaxAPIRetries, global.DryRun, "PUT", log)
		},
	}

	cmd.Flags().StringSliceVarP(&o.paths, "paths", "p", []string{}, "The comma separated list of artifact directories.")
	cmd.Flags().StringVarP(&o.id, "id", "i", "", "The unique identifier of the source infrastructure of the report (e.g. the K8S cluster/namespace name). If not set, it is defaulted to environment name.")

	err := RequireFlags(cmd, []string{"paths"})
	if err != nil {
		log.Fatalf("failed to configure required flags: %v", err)
	}

	return cmd
}
