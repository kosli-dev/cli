package main

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/merkely-development/reporter/internal/aws"
	"github.com/merkely-development/reporter/internal/requests"
	"github.com/spf13/cobra"
)

const ecsEnvDesc = `
List the artifacts deployed in an AWS ECS cluster and their digests 
and report them to Merkely. 
`

const ecsEnvExample = `
* report what's running in an entire AWS ECS cluster:
merkely report env ecs prod --api-token 1234 --owner exampleOrg
`

type ecsEnvOptions struct {
	cluster     string
	serviceName string
	id          string
}

func newEcsEnvCmd(out io.Writer) *cobra.Command {
	o := new(ecsEnvOptions)
	cmd := &cobra.Command{
		Use:     "ecs env-name",
		Short:   "Report images data from AWS ECS cluster to Merkely.",
		Long:    ecsEnvDesc,
		Example: ecsEnvExample,
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
				if o.serviceName != "" {
					o.id = o.serviceName
				} else if o.cluster != "" {
					o.id = o.cluster
				} else {
					o.id = envName
				}
			}
			url := fmt.Sprintf("%s/api/v1/environments/%s/%s/data", global.Host, global.Owner, envName)
			client, err := aws.NewAWSClient()
			if err != nil {
				return err
			}
			tasksData, err := aws.GetEcsTasksData(client, o.cluster, o.serviceName)
			if err != nil {
				return err
			}

			requestBody := &requests.EcsEnvRequest{
				Artifacts: tasksData,
				Type:      "ECS",
				Id:        o.id,
			}
			js, _ := json.MarshalIndent(requestBody, "", "    ")

			return requests.SendPayload(js, url, global.ApiToken,
				global.MaxAPIRetries, global.DryRun)
		},
	}

	cmd.Flags().StringVarP(&o.cluster, "cluster", "C", "", "name of the ECS cluster")
	cmd.Flags().StringVarP(&o.serviceName, "service-name", "s", "", "name of the ECS service")
	cmd.Flags().StringVarP(&o.id, "id", "i", "", "the unique identifier of the source infrastructure of the report (e.g. the ECS cluster/service name)."+
		"If not set, it is defaulted based on the following order: --service-name, --cluster, environment name.")

	return cmd
}
