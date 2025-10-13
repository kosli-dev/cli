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
Skip all filtering flags to report everything running in all clusters in a given AWS account. 

Use ^--clusters^ and/or ^--clusters-regex^ OR ^--exclude^ and/or ^--exclude-regex^ to filter the clusters to snapshot.
You can also filter the services within a cluster using ^--services^ and/or ^--services-regex^. Or use ^--exclude-services^ and/or ^--exclude-services-regex^ to exclude some services. 
Note that service filtering is applied to all clusters being snapshot.

All filtering options are case-sensitive.

The reported data includes cluster and service names, container image digests and creation timestamps.` + awsAuthDesc

const snapshotECSExample = `
# authentication to AWS using flags
kosli snapshot ecs yourEnvironmentName \
	--aws-key-id yourAWSAccessKeyID \
	--aws-secret-key yourAWSSecretAccessKey \
	--aws-region yourAWSRegion \
	--api-token yourAPIToken \
	--org yourOrgName

# authentication to AWS using env variables
export AWS_ACCESS_KEY_ID=yourAWSAccessKeyID
export AWS_SECRET_ACCESS_KEY=yourAWSSecretAccessKey
export AWS_REGION=yourAWSRegion

kosli snapshot ecs yourEnvironmentName \
	--api-token yourAPIToken \
	--org yourOrgName

# reporting everything running in all clusters in a given AWS account 
kosli snapshot ecs my-env \
	--api-token yourAPIToken \
	--org yourOrgName

# include clusters matching a name in the AWS account
kosli snapshot ecs my-env --clusters my-cluster ...

# include clusters matching a pattern in the AWS account
kosli snapshot ecs my-env --clusters-regex "my-cluster-*" ...

# include clusters matching a list of names in the AWS account
kosli snapshot ecs my-env --clusters my-cluster1,my-cluster2 ...

# exclude clusters matching a name in the AWS account
kosli snapshot ecs my-env --exclude my-cluster ...

# exclude clusters matching a pattern in the AWS account
kosli snapshot ecs my-env --exclude-regex "my-cluster-*" ...

# exclude clusters matching a list of names in the AWS account
kosli snapshot ecs my-env --exclude my-cluster1,my-cluster2 ...

# include Services matching a name in one cluster
kosli snapshot ecs my-env --clusters my-cluster --services backend-app ...

# include Services matching a pattern in one cluster
kosli snapshot ecs my-env --clusters my-cluster --services-regex "backend-*" ...

# include production Services only (by naming convention) in all clusters in the AWS account
kosli snapshot ecs my-env --services-regex "*-prod-*" ...

# include Services matching a name in all clusters in the AWS account
kosli snapshot ecs my-env --services backend-app ...

# include Services matching a list of names in all clusters in the AWS account
kosli snapshot ecs my-env --services backend-app,frontend-app ...

# exclude Services matching a pattern in one cluster
kosli snapshot ecs my-env --clusters my-cluster --exclude-services-regex "backend-*" ...

# exclude Production services only (by naming convention)  in all clusters in the AWS account
kosli snapshot ecs my-env --exclude-services-regex "*-prod-*" ...

# exclude Services matching a name in one cluster
kosli snapshot ecs my-env --clusters my-cluster --exclude-services backend-app ...

# exclude Services matching a name in all clusters in the AWS account
kosli snapshot ecs my-env --exclude-services backend-app ...

# exclude Services matching a list of names in all clusters in the AWS account
kosli snapshot ecs my-env --exclude-services backend-app,frontend-app ...
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

	tasksData, err := o.awsStaticCreds.GetEcsTasksData(o.clustersFilter, o.serviceFilter, logger)
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
