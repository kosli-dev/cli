package main

import (
	"fmt"
	"io"
	"net/http"

	"github.com/kosli-dev/cli/internal/aws"
	"github.com/kosli-dev/cli/internal/requests"
	"github.com/spf13/cobra"
)

const environmentReportECSShortDesc = `Report running containers data from AWS ECS cluster or service to Kosli.`
const environmentReportECSLongDesc = environmentReportECSShortDesc + `
The reported data includes container image digests and creation timestamps.`

const environmentReportECSExample = `
# report what is running in an entire AWS ECS cluster:
export AWS_REGION=yourAWSRegion
export AWS_ACCESS_KEY_ID=yourAWSAccessKeyID
export AWS_SECRET_ACCESS_KEY=yourAWSSecretAccessKey

kosli environment report ecs yourEnvironmentName \
	--cluster yourECSClusterName \
	--api-token yourAPIToken \
	--owner yourOrgName

# report what is running in a specific AWS ECS service:
export AWS_REGION=yourAWSRegion
export AWS_ACCESS_KEY_ID=yourAWSAccessKeyID
export AWS_SECRET_ACCESS_KEY=yourAWSSecretAccessKey

kosli environment report ecs yourEnvironmentName \
	--service-name yourECSServiceName \
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
		Use:     "ecs ENVIRONMENT-NAME",
		Short:   environmentReportECSShortDesc,
		Long:    environmentReportECSLongDesc,
		Example: environmentReportECSExample,
		Args:    cobra.ExactArgs(1),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			err := RequireGlobalFlags(global, []string{"Owner", "ApiToken"})
			if err != nil {
				return ErrorBeforePrintingUsage(cmd, err.Error())
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return o.run(args)
		},
	}

	cmd.Flags().StringVarP(&o.cluster, "cluster", "C", "", ecsClusterFlag)
	cmd.Flags().StringVarP(&o.serviceName, "service-name", "s", "", ecsServiceFlag)
	addDryRunFlag(cmd)
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

	payload := &aws.EcsEnvRequest{
		Artifacts: tasksData,
		Type:      "ECS",
		Id:        o.id,
	}

	reqParams := &requests.RequestParams{
		Method:   http.MethodPut,
		URL:      url,
		Payload:  payload,
		DryRun:   global.DryRun,
		Password: global.ApiToken,
	}
	_, err = kosliClient.Do(reqParams)
	if err == nil && !global.DryRun {
		logger.Info("[%d] containers were reported to environment %s", len(payload.Artifacts), envName)
	}
	return err
}
