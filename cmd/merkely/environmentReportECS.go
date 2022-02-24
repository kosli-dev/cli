package main

import (
	"fmt"
	"io"
	"net/http"

	"github.com/merkely-development/reporter/internal/aws"
	"github.com/merkely-development/reporter/internal/requests"
	"github.com/spf13/cobra"
)

const environmentReportECSDesc = `
List the artifacts deployed in an AWS ECS cluster and their digests 
and report them to Merkely. 
`

const environmentReportECSExample = `
# report what is running in an entire AWS ECS cluster:
export AWS_REGION=yourAWSRegion
export AWS_ACCESS_KEY_ID=yourAWSAccessKeyID
export AWS_SECRET_ACCESS_KEY=yourAWSSecretAccessKey

merkely environment report ecs yourEnvironmentName \
	--api-token yourAPIToken \
	--owner yourOrgName
`

type environmentReportECSOptions struct {
	cluster     string
	serviceName string
	id          string
}

func newEnvironmentReportECSCmd(out io.Writer) *cobra.Command {
	o := new(environmentReportECSOptions)
	cmd := &cobra.Command{
		Use:     "ecs env-name",
		Short:   "Report images data from AWS ECS cluster to Merkely.",
		Long:    environmentReportECSDesc,
		Example: environmentReportECSExample,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if len(args) > 1 {
				return ErrorAfterPrintingHelp(cmd, "only env-name argument is allowed")
			}
			if len(args) == 0 || args[0] == "" {
				return ErrorAfterPrintingHelp(cmd, "env-name argument is required")
			}

			err := RequireGlobalFlags(global, []string{"Owner", "ApiToken"})
			if err != nil {
				return ErrorAfterPrintingHelp(cmd, err.Error())
			}

			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return o.run(args)
		},
	}

	cmd.Flags().StringVarP(&o.cluster, "cluster", "C", "", "The name of the ECS cluster.")
	cmd.Flags().StringVarP(&o.serviceName, "service-name", "s", "", "The name of the ECS service.")
	return cmd
}

func (o *environmentReportECSOptions) run(args []string) error {
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

	requestBody := &aws.EcsEnvRequest{
		Artifacts: tasksData,
		Type:      "ECS",
		Id:        o.id,
	}

	_, err = requests.SendPayload(requestBody, url, "", global.ApiToken,
		global.MaxAPIRetries, global.DryRun, http.MethodPut, log)
	return err
}
