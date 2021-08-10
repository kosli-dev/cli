package main

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/merkely-development/reporter/internal/aws"
	"github.com/merkely-development/reporter/internal/kube"
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
	family      string
	serviceName string
}

func newEcsEnvCmd(out io.Writer) *cobra.Command {
	o := new(ecsEnvOptions)
	cmd := &cobra.Command{
		Use:     "ecs env-name",
		Short:   "Report images data from AWS ECS cluster to Merkely.",
		Long:    ecsEnvDesc,
		Example: ecsEnvExample,
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) > 1 {
				return fmt.Errorf("only environment name argument is allowed")
			}
			if len(args) == 0 || args[0] == "" {
				return fmt.Errorf("environment name is required")
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {

			if o.serviceName != "" && (o.cluster != "" || o.family != "") {
				return fmt.Errorf("cannot specify --service-name with --family or --cluster")
			}

			envName := args[0]
			url := fmt.Sprintf("%s/api/v1/environments/%s/%s/data", global.host, global.owner, envName)
			client, err := aws.NewAWSClient()
			if err != nil {
				return err
			}
			tasksData, err := aws.ListEcsTasks(client, o.cluster, o.family, o.serviceName)
			if err != nil {
				return err
			}
			fmt.Println(*tasksData[0])

			requestBody := &requests.EnvRequest{
				Data: []*kube.PodData{},
			}
			js, _ := json.MarshalIndent(requestBody, "", "    ")

			if global.dryRun {
				fmt.Println("############### THIS IS A DRY-RUN  ###############")
				fmt.Println(string(js))
			} else {
				fmt.Println("****** Sending a Test to the API ******")
				fmt.Println(string(js))
				resp, err := requests.DoPut(js, url, global.apiToken, global.maxAPIRetries)
				if err != nil {
					return err
				}
				if resp.StatusCode != 201 && resp.StatusCode != 200 {
					return fmt.Errorf("failed to send scrape data: %v", resp.Body)
				}
			}
			return nil
		},
	}

	cmd.Flags().StringVarP(&o.cluster, "cluster", "C", "", "name of the ECS cluster")
	cmd.Flags().StringVarP(&o.family, "family", "f", "", "name of the ECS task definition family")
	cmd.Flags().StringVarP(&o.serviceName, "service-name", "s", "", "name of the ECS service")

	return cmd
}
