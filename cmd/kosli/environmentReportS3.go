package main

import (
	"fmt"
	"io"
	"net/http"

	"github.com/kosli-dev/cli/internal/aws"
	"github.com/kosli-dev/cli/internal/requests"
	"github.com/spf13/cobra"
)

const environmentReportS3ShortDesc = `Report an artifact deployed in AWS S3 bucket to Kosli. `

const environmentReportS3LongDesc = environmentReportS3ShortDesc + `
To authenticate to AWS, you can either export the AWS env vars or use the command flags to pass them.
See the examples below.
`

const environmentReportS3Example = `
# report what is running in an AWS S3 bucket (AWS auth provided in env variables):
export AWS_REGION=yourAWSRegion
export AWS_ACCESS_KEY_ID=yourAWSAccessKeyID
export AWS_SECRET_ACCESS_KEY=yourAWSSecretAccessKey

kosli environment report s3 yourEnvironmentName \
	--bucket yourBucketName \
	--api-token yourAPIToken \
	--owner yourOrgName

# report what is running in an AWS S3 bucket (AWS auth provided in flags):
kosli environment report s3 yourEnvironmentName \
	--bucket yourBucketName \
	--aws-key-id yourAWSAccessKeyID \
	--aws-secret-key yourAWSSecretAccessKey \
	--aws-region yourAWSRegion \
	--api-token yourAPIToken \
	--owner yourOrgName	
`

type environmentReportS3Options struct {
	bucket         string
	awsAuthOptions *awsAuthOptions
}

func newEnvironmentReportS3Cmd(out io.Writer) *cobra.Command {
	o := new(environmentReportS3Options)
	o.awsAuthOptions = new(awsAuthOptions)
	cmd := &cobra.Command{
		Use:     "s3 ENVIRONMENT-NAME",
		Aliases: []string{"S3"},
		Short:   environmentReportS3ShortDesc,
		Long:    environmentReportS3LongDesc,
		Example: environmentReportS3Example,
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

	cmd.Flags().StringVar(&o.bucket, "bucket", "", bucketNameFlag)
	addAWSAuthFlags(cmd, o.awsAuthOptions)

	err := RequireFlags(cmd, []string{"bucket"})
	if err != nil {
		logger.Error("failed to configure required flags: %v", err)
	}

	return cmd
}

func (o *environmentReportS3Options) run(args []string) error {
	envName := args[0]

	url := fmt.Sprintf("%s/api/v1/environments/%s/%s/data", global.Host, global.Owner, envName)
	creds := aws.AWSCredentials(o.awsAuthOptions.accessKey, o.awsAuthOptions.secretKey)
	s3Data, err := aws.GetS3Data(o.bucket, creds, o.awsAuthOptions.region, logger)
	if err != nil {
		return err
	}
	payload := &aws.S3EnvRequest{
		Artifacts: s3Data,
		Type:      "S3",
		Id:        envName,
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
		logger.Info("bucket %s was reported to environment %s", o.bucket, envName)
	}
	return err
}
