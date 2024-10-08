package main

import (
	"fmt"
	"io"
	"net/http"

	"github.com/kosli-dev/cli/internal/aws"
	"github.com/kosli-dev/cli/internal/requests"
	"github.com/spf13/cobra"
)

const snapshotECSShortDesc = `Report a snapshot of running containers in an AWS ECS cluster or service to Kosli.  `
const snapshotECSLongDesc = snapshotECSShortDesc + `
The reported data includes container image digests and creation timestamps.` + awsAuthDesc

const snapshotECSExample = `
# report what is running in an entire AWS ECS cluster:
export AWS_REGION=yourAWSRegion
export AWS_ACCESS_KEY_ID=yourAWSAccessKeyID
export AWS_SECRET_ACCESS_KEY=yourAWSSecretAccessKey

kosli snapshot ecs yourEnvironmentName \
	--cluster yourECSClusterName \
	--api-token yourAPIToken \
	--org yourOrgName

# report what is running in all AWS ECS clusters to a single Kosli environment:
export AWS_REGION=yourAWSRegion
export AWS_ACCESS_KEY_ID=yourAWSAccessKeyID
export AWS_SECRET_ACCESS_KEY=yourAWSSecretAccessKey

kosli snapshot ecs yourEnvironmentName \
	--scan-all \
	--api-token yourAPIToken \
	--org yourOrgName

# report what is running in a specific AWS ECS service within a cluster:
export AWS_REGION=yourAWSRegion
export AWS_ACCESS_KEY_ID=yourAWSAccessKeyID
export AWS_SECRET_ACCESS_KEY=yourAWSSecretAccessKey

kosli snapshot ecs yourEnvironmentName \
	--cluster yourECSClusterName \
	--service-name yourECSServiceName \
	--api-token yourAPIToken \
	--org yourOrgName

# report what is running in in a specific AWS ECS service within a cluster (AWS auth provided in flags):
kosli snapshot ecs yourEnvironmentName \
	--cluster yourECSClusterName \
	--service-name yourECSServiceName \
	--aws-key-id yourAWSAccessKeyID \
	--aws-secret-key yourAWSSecretAccessKey \
	--aws-region yourAWSRegion \
	--api-token yourAPIToken \
	--org yourOrgName
`

type snapshotECSOptions struct {
	cluster        string
	serviceName    string
	awsStaticCreds *aws.AWSStaticCreds
	scanAll        bool
}

func newSnapshotECSCmd(out io.Writer) *cobra.Command {
	o := new(snapshotECSOptions)
	o.awsStaticCreds = new(aws.AWSStaticCreds)
	cmd := &cobra.Command{
		Use:     "ecs ENVIRONMENT-NAME",
		Short:   snapshotECSShortDesc,
		Long:    snapshotECSLongDesc,
		Example: snapshotECSExample,
		Args:    cobra.ExactArgs(1),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			err := RequireGlobalFlags(global, []string{"Org", "ApiToken"})
			if err != nil {
				return ErrorBeforePrintingUsage(cmd, err.Error())
			}

			err = MuXRequiredFlags(cmd, []string{"cluster", "scan-all"}, true)
			if err != nil {
				return err
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return o.run(args)
		},
	}

	cmd.Flags().StringVarP(&o.cluster, "cluster", "C", "", ecsClusterFlag)
	cmd.Flags().StringVarP(&o.serviceName, "service-name", "s", "", ecsServiceFlag)
	cmd.Flags().BoolVarP(&o.scanAll, "scan-all", "A", false, ecsScanAllFlag)
	addAWSAuthFlags(cmd, o.awsStaticCreds)
	addDryRunFlag(cmd)

	return cmd
}

func (o *snapshotECSOptions) run(args []string) error {
	envName := args[0]
	url := fmt.Sprintf("%s/api/v2/environments/%s/%s/report/ECS", global.Host, global.Org, envName)
	logger.Debug("ECS Snapshot parameters: scan-all: %t, cluster: %s", o.scanAll, o.cluster)

	tasksDataList := []*aws.EcsTaskData{}

	if o.scanAll {
		logger.Debug("Attempting to find ECS clusters")
		clusters, err := o.awsStaticCreds.GetECSClusters()
		if err != nil {
			logger.Error("Failed to get ECS clusters: %v", err)
			return err
		}

		logger.Debug("Found %d ECS clusters.", len(clusters))
		for _, cluster := range clusters {
			logger.Debug("Processing ECS cluster with ARN: %s", cluster)
			tasksData, err := o.awsStaticCreds.GetEcsTasksData(cluster, o.serviceName)
			if err != nil {
				return err
			}
			// append to EcsTaskDataList
			tasksDataList = append(tasksDataList, tasksData...)

		}
	} else {
		tasksData, err := o.awsStaticCreds.GetEcsTasksData(o.cluster, o.serviceName)
		if err != nil {
			return err
		}
		tasksDataList = append(tasksDataList, tasksData...)
	}

	payload := &aws.EcsEnvRequest{
		Artifacts: tasksDataList,
	}

	reqParams := &requests.RequestParams{
		Method:   http.MethodPut,
		URL:      url,
		Payload:  payload,
		DryRun:   global.DryRun,
		Password: global.ApiToken,
	}
	_, err := kosliClient.Do(reqParams)
	if err == nil && !global.DryRun {
		logger.Info("[%d] containers were reported to environment %s", len(payload.Artifacts), envName)
	}
	return err
}
