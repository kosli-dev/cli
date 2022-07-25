package main

import (
	"fmt"
	"io"
	"net/http"

	"github.com/kosli-dev/cli/internal/aws"
	"github.com/kosli-dev/cli/internal/requests"
	"github.com/spf13/cobra"
)

const environmentReportS3Desc = `
Report the artifact deployed in an AWS S3 bucket and its digest to Kosli. 
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
	bucket    string
	accessKey string
	secretKey string
	region    string
}

func newEnvironmentReportS3Cmd(out io.Writer) *cobra.Command {
	o := new(environmentReportS3Options)
	cmd := &cobra.Command{
		Use:     "s3 ENVIRONMENT-NAME",
		Aliases: []string{"S3"},
		Short:   "Report artifact from AWS S3 bucket to Kosli.",
		Long:    environmentReportS3Desc,
		Example: environmentReportS3Example,
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

	cmd.Flags().StringVar(&o.bucket, "bucket", "", bucketNameFlag)
	cmd.Flags().StringVar(&o.accessKey, "aws-key-id", "", awsKeyIdFlag)
	cmd.Flags().StringVar(&o.secretKey, "aws-secret-key", "", awsSecretKeyFlag)
	cmd.Flags().StringVar(&o.region, "aws-region", "", awsRegionFlag)

	err := RequireFlags(cmd, []string{"bucket"})
	if err != nil {
		log.Fatalf("failed to configure required flags: %v", err)
	}

	return cmd
}

func (o *environmentReportS3Options) run(args []string) error {
	envName := args[0]

	url := fmt.Sprintf("%s/api/v1/environments/%s/%s/data", global.Host, global.Owner, envName)
	creds := aws.AWSCredentials(o.accessKey, o.secretKey)
	s3Data, err := aws.GetS3Data(o.bucket, creds, o.region)
	if err != nil {
		return err
	}
	requestBody := &aws.S3EnvRequest{
		Artifacts: s3Data,
		Type:      "S3",
		Id:        envName,
	}

	_, err = requests.SendPayload(requestBody, url, "", global.ApiToken,
		global.MaxAPIRetries, global.DryRun, http.MethodPut, log)
	return err
}
