package main

import (
	"fmt"
	"io"
	"net/http"

	"github.com/kosli-dev/cli/internal/aws"
	"github.com/kosli-dev/cli/internal/requests"
	"github.com/spf13/cobra"
)

const snapshotECSShortDesc = `Report a snapshot of running containers in one or more AWS ECS cluster(s) to Kosli.  `
const snapshotECSLongDesc = snapshotECSShortDesc + `
Skip ^--clusters^ and ^--clusters-regex^ to report all clusters in a given AWS account. Or use ^--exclude^ and/or ^--exclude-regex^ to report all clusters excluding some.
The reported data includes container image digests and creation timestamps.` + awsAuthDesc

const snapshotECSExample = `
# report what is running in an entire AWS ECS cluster:
export AWS_REGION=yourAWSRegion
export AWS_ACCESS_KEY_ID=yourAWSAccessKeyID
export AWS_SECRET_ACCESS_KEY=yourAWSSecretAccessKey

kosli snapshot ecs yourEnvironmentName \
	--clusters yourECSClusterName \
	--api-token yourAPIToken \
	--org yourOrgName

# report what is running in a specific AWS ECS service within a cluster:
export AWS_REGION=yourAWSRegion
export AWS_ACCESS_KEY_ID=yourAWSAccessKeyID
export AWS_SECRET_ACCESS_KEY=yourAWSSecretAccessKey

kosli snapshot ecs yourEnvironmentName \
	--clusters yourECSClusterName \
	--service-name yourECSServiceName \
	--api-token yourAPIToken \
	--org yourOrgName

# report what is running in all ECS clusters in an AWS account (AWS auth provided in flags):
kosli snapshot ecs yourEnvironmentName \
	--aws-key-id yourAWSAccessKeyID \
	--aws-secret-key yourAWSSecretAccessKey \
	--aws-region yourAWSRegion \
	--api-token yourAPIToken \
	--org yourOrgName

# report what is running in all ECS clusters in an AWS account except for clusters with names matching given regex patterns:
kosli snapshot ecs yourEnvironmentName \
	--aws-key-id yourAWSAccessKeyID \
	--aws-secret-key yourAWSSecretAccessKey \
	--aws-region yourAWSRegion \
	--exclude-regex "those-names.*" \
	--api-token yourAPIToken \
	--org yourOrgName
`

type snapshotECSOptions struct {
	clusters       []string
	clustersRegex  []string
	exclude        []string
	excludeRegex   []string
	serviceName    string
	cluster        string
	awsStaticCreds *aws.AWSStaticCreds
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

			err = MuXRequiredFlags(cmd, []string{"cluster", "clusters", "exclude"}, false)
			if err != nil {
				return err
			}
			err = MuXRequiredFlags(cmd, []string{"clusters-regex", "exclude"}, false)
			if err != nil {
				return err
			}
			err = MuXRequiredFlags(cmd, []string{"cluster", "clusters", "exclude-regex"}, false)
			if err != nil {
				return err
			}
			err = MuXRequiredFlags(cmd, []string{"clusters-regex", "exclude-regex"}, false)
			if err != nil {
				return err
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return o.run(args)
		},
	}

	cmd.Flags().StringSliceVar(&o.clusters, "clusters", []string{}, ecsClustersFlag)
	cmd.Flags().StringSliceVar(&o.clustersRegex, "clusters-regex", []string{}, ecsClustersRegexFlag)
	cmd.Flags().StringSliceVar(&o.exclude, "exclude", []string{}, ecsExcludeClustersFlag)
	cmd.Flags().StringSliceVar(&o.excludeRegex, "exclude-regex", []string{}, ecsExcludeClustersRegexFlag)

	cmd.Flags().StringVarP(&o.cluster, "cluster", "C", "", ecsClusterFlag)
	cmd.Flags().StringVarP(&o.serviceName, "service-name", "s", "", ecsServiceFlag)
	addAWSAuthFlags(cmd, o.awsStaticCreds)
	addDryRunFlag(cmd)

	err := DeprecateFlags(cmd, map[string]string{
		"cluster":      "use --clusters instead",
		"service-name": "it will be removed in a future release",
	})
	if err != nil {
		logger.Error("failed to configure deprecated flags: %v", err)
	}

	return cmd
}

func (o *snapshotECSOptions) run(args []string) error {
	envName := args[0]
	url := fmt.Sprintf("%s/api/v2/environments/%s/%s/report/ECS", global.Host, global.Org, envName)

	if o.cluster != "" {
		o.clusters = append(o.clusters, o.cluster)
	}

	tasksData, err := o.awsStaticCreds.GetEcsTasksData(o.clusters, o.clustersRegex, o.exclude, o.excludeRegex)
	if err != nil {
		return err
	}

	payload := &aws.EcsEnvRequest{
		Artifacts: tasksData,
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
