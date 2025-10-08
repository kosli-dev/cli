package main

import (
	"fmt"
	"io"
	"net/http"

	"github.com/kosli-dev/cli/internal/aws"
	"github.com/kosli-dev/cli/internal/filters"
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
	clustersFilter *filters.ResourceFilterOptions
	serviceFilter  *filters.ResourceFilterOptions
	serviceName    string
	awsStaticCreds *aws.AWSStaticCreds
}

func newSnapshotECSCmd(out io.Writer) *cobra.Command {
	o := new(snapshotECSOptions)
	o.awsStaticCreds = new(aws.AWSStaticCreds)
	o.clustersFilter = new(filters.ResourceFilterOptions)
	o.serviceFilter = new(filters.ResourceFilterOptions)
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
			// service filtering flags
			// Include flags vs exclude flags mutual exclusion
			err = MuXRequiredFlags(cmd, []string{"services", "exclude-services"}, false)
			if err != nil {
				return err
			}
			err = MuXRequiredFlags(cmd, []string{"services", "exclude-services-regex"}, false)
			if err != nil {
				return err
			}
			err = MuXRequiredFlags(cmd, []string{"services-regex", "exclude-services"}, false)
			if err != nil {
				return err
			}
			err = MuXRequiredFlags(cmd, []string{"services-regex", "exclude-services-regex"}, false)
			if err != nil {
				return err
			}

			// Deprecated flag vs new flags mutual exclusion
			err = MuXRequiredFlags(cmd, []string{"service-name", "services", "services-regex", "exclude-services", "exclude-services-regex"}, false)
			if err != nil {
				return err
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return o.run(args)
		},
	}

	cmd.Flags().StringSliceVar(&o.clustersFilter.IncludeNames, "clusters", []string{}, ecsClustersFlag)
	cmd.Flags().StringSliceVar(&o.clustersFilter.IncludeNamesRegex, "clusters-regex", []string{}, ecsClustersRegexFlag)
	cmd.Flags().StringSliceVar(&o.clustersFilter.ExcludeNames, "exclude", []string{}, ecsExcludeClustersFlag)
	cmd.Flags().StringSliceVar(&o.clustersFilter.ExcludeNamesRegex, "exclude-regex", []string{}, ecsExcludeClustersRegexFlag)

	cmd.Flags().StringSliceVar(&o.serviceFilter.IncludeNames, "services", []string{}, ecsServicesFlag)
	cmd.Flags().StringSliceVar(&o.serviceFilter.IncludeNamesRegex, "services-regex", []string{}, ecsServicesRegexFlag)
	cmd.Flags().StringSliceVar(&o.serviceFilter.ExcludeNames, "exclude-services", []string{}, ecsExcludeServicesFlag)
	cmd.Flags().StringSliceVar(&o.serviceFilter.ExcludeNamesRegex, "exclude-services-regex", []string{}, ecsExcludeServicesRegexFlag)

	cmd.Flags().StringSliceVarP(&o.clustersFilter.IncludeNames, "cluster", "C", []string{}, ecsClusterFlag)
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

	tasksData, err := o.awsStaticCreds.GetEcsTasksData(o.clustersFilter, o.serviceFilter)
	if err != nil {
		return err
	}

	payload := &aws.EcsEnvRequest{
		Artifacts: tasksData,
	}

	reqParams := &requests.RequestParams{
		Method:  http.MethodPut,
		URL:     url,
		Payload: payload,
		DryRun:  global.DryRun,
		Token:   global.ApiToken,
	}
	_, err = kosliClient.Do(reqParams)
	if err == nil && !global.DryRun {
		logger.Info("[%d] containers were reported to environment %s", len(payload.Artifacts), envName)
	}
	return err
}
