package main

import (
	"fmt"
	"io"
	"net/http"

	"github.com/kosli-dev/cli/internal/aws"
	"github.com/kosli-dev/cli/internal/requests"
	"github.com/spf13/cobra"
)

const snapshotLambdaShortDesc = `Report a snapshot of artifacts deployed as one or more AWS Lambda functions and their digests to Kosli.`

const snapshotLambdaLongDesc = snapshotLambdaShortDesc + `  
Skip --function-names to report all functions in a given AWS account.` + awsAuthDesc

const snapshotLambdaExample = `
# report all Lambda functions running in an AWS account (AWS auth provided in env variables):
export AWS_REGION=yourAWSRegion
export AWS_ACCESS_KEY_ID=yourAWSAccessKeyID
export AWS_SECRET_ACCESS_KEY=yourAWSSecretAccessKey

kosli snapshot lambda yourEnvironmentName \
	--api-token yourAPIToken \
	--org yourOrgName

# report what is running in the latest version of an AWS Lambda function (AWS auth provided in env variables):
export AWS_REGION=yourAWSRegion
export AWS_ACCESS_KEY_ID=yourAWSAccessKeyID
export AWS_SECRET_ACCESS_KEY=yourAWSSecretAccessKey

kosli snapshot lambda yourEnvironmentName \
	--function-names yourFunctionName \
	--api-token yourAPIToken \
	--org yourOrgName

# report what is running in the latest version of multiple AWS Lambda functions (AWS auth provided in env variables):
export AWS_REGION=yourAWSRegion
export AWS_ACCESS_KEY_ID=yourAWSAccessKeyID
export AWS_SECRET_ACCESS_KEY=yourAWSSecretAccessKey

kosli snapshot lambda yourEnvironmentName \
	--function-names yourFirstFunctionName,yourSecondFunctionName \
	--api-token yourAPIToken \
	--org yourOrgName

# report what is running in the latest version of an AWS Lambda function (AWS auth provided in flags):
kosli snapshot lambda yourEnvironmentName \
	--function-names yourFunctionName \
	--aws-key-id yourAWSAccessKeyID \
	--aws-secret-key yourAWSSecretAccessKey \
	--aws-region yourAWSRegion \
	--api-token yourAPIToken \
	--org yourOrgName
`

type snapshotLambdaOptions struct {
	functionNames   []string
	functionVersion string
	awsStaticCreds  *aws.AWSStaticCreds
}

func newSnapshotLambdaCmd(out io.Writer) *cobra.Command {
	o := new(snapshotLambdaOptions)
	o.awsStaticCreds = new(aws.AWSStaticCreds)
	cmd := &cobra.Command{
		Use:     "lambda ENVIRONMENT-NAME",
		Short:   snapshotLambdaShortDesc,
		Long:    snapshotLambdaLongDesc,
		Example: snapshotLambdaExample,
		Args:    cobra.ExactArgs(1),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			err := RequireGlobalFlags(global, []string{"Org", "ApiToken"})
			if err != nil {
				return ErrorBeforePrintingUsage(cmd, err.Error())
			}

			err = MuXRequiredFlags(cmd, []string{"function-name", "function-names"}, false)
			if err != nil {
				return err
			}

			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return o.run(args)
		},
	}

	cmd.Flags().StringSliceVar(&o.functionNames, "function-name", []string{}, functionNameFlag)
	cmd.Flags().StringSliceVar(&o.functionNames, "function-names", []string{}, functionNamesFlag)
	cmd.Flags().StringVar(&o.functionVersion, "function-version", "", functionVersionFlag)
	addAWSAuthFlags(cmd, o.awsStaticCreds)
	addDryRunFlag(cmd)

	err := DeprecateFlags(cmd, map[string]string{
		"function-name":    "use --function-names instead",
		"function-version": "--function-version is no longer supported. It will be removed in a future release.",
	})
	if err != nil {
		logger.Error("failed to configure deprecated flags: %v", err)
	}

	return cmd
}

func (o *snapshotLambdaOptions) run(args []string) error {
	envName := args[0]

	url := fmt.Sprintf("%s/api/v2/environments/%s/%s/report/lambda", global.Host, global.Org, envName)
	lambdaData, err := o.awsStaticCreds.GetLambdaPackageData(o.functionNames)
	if err != nil {
		return err
	}

	payload := &aws.LambdaEnvRequest{
		Artifacts: lambdaData,
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
		logger.Info("%d lambda functions were reported to environment %s", len(lambdaData), envName)
	}
	return err
}
