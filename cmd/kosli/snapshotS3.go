package main

import (
	"fmt"
	"io"
	"net/http"

	"github.com/kosli-dev/cli/internal/aws"
	"github.com/kosli-dev/cli/internal/requests"
	"github.com/spf13/cobra"
)

const snapshotS3ShortDesc = `Report a snapshot of the content of an AWS S3 bucket to Kosli.`

const snapshotS3LongDesc = snapshotS3ShortDesc + awsAuthDesc + `
You can report the entire bucket content, or filter some of the content using ^--include^ and ^--exclude^.
In all cases, the content is reported as one artifact. If you wish to report separate files/dirs within the same bucket as separate artifacts, you need to run the command twice.

` + kosliIgnoreDesc

const snapshotS3Example = `
# report the contents of an entire AWS S3 bucket (AWS auth provided in env variables):
export AWS_REGION=yourAWSRegion
export AWS_ACCESS_KEY_ID=yourAWSAccessKeyID
export AWS_SECRET_ACCESS_KEY=yourAWSSecretAccessKey

kosli snapshot s3 yourEnvironmentName \
	--bucket yourBucketName \
	--api-token yourAPIToken \
	--org yourOrgName

# report what is running in an AWS S3 bucket (AWS auth provided in flags):
kosli snapshot s3 yourEnvironmentName \
	--bucket yourBucketName \
	--aws-key-id yourAWSAccessKeyID \
	--aws-secret-key yourAWSSecretAccessKey \
	--aws-region yourAWSRegion \
	--api-token yourAPIToken \
	--org yourOrgName	

# report a subset of contents of an AWS S3 bucket (AWS auth provided in env variables):
export AWS_REGION=yourAWSRegion
export AWS_ACCESS_KEY_ID=yourAWSAccessKeyID
export AWS_SECRET_ACCESS_KEY=yourAWSSecretAccessKey

kosli snapshot s3 yourEnvironmentName \
	--bucket yourBucketName \
	--include file.txt,path/within/bucket \
	--api-token yourAPIToken \
	--org yourOrgName

# report contents of an entire AWS S3 bucket, except for some paths (AWS auth provided in env variables):
export AWS_REGION=yourAWSRegion
export AWS_ACCESS_KEY_ID=yourAWSAccessKeyID
export AWS_SECRET_ACCESS_KEY=yourAWSSecretAccessKey

kosli snapshot s3 yourEnvironmentName \
	--bucket yourBucketName \
	--exclude file.txt,path/within/bucket \
	--api-token yourAPIToken \
	--org yourOrgName
`

type snapshotS3Options struct {
	bucket         string
	includePaths   []string
	excludePaths   []string
	awsStaticCreds *aws.AWSStaticCreds
}

func newSnapshotS3Cmd(out io.Writer) *cobra.Command {
	o := new(snapshotS3Options)
	o.awsStaticCreds = new(aws.AWSStaticCreds)
	cmd := &cobra.Command{
		Use:     "s3 ENVIRONMENT-NAME",
		Aliases: []string{"S3"},
		Short:   snapshotS3ShortDesc,
		Long:    snapshotS3LongDesc,
		Example: snapshotS3Example,
		Args:    cobra.ExactArgs(1),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			err := RequireGlobalFlags(global, []string{"Org", "ApiToken"})
			if err != nil {
				return ErrorBeforePrintingUsage(cmd, err.Error())
			}

			err = MuXRequiredFlags(cmd, []string{"include", "exclude"}, false)
			if err != nil {
				return err
			}

			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return o.run(args)
		},
	}

	cmd.Flags().StringVar(&o.bucket, "bucket", "", bucketNameFlag)
	cmd.Flags().StringSliceVarP(&o.includePaths, "include", "i", []string{}, bucketPathsFlag)
	cmd.Flags().StringSliceVarP(&o.excludePaths, "exclude", "x", []string{}, excludeBucketPathsFlag)
	addAWSAuthFlags(cmd, o.awsStaticCreds)
	addDryRunFlag(cmd)

	err := RequireFlags(cmd, []string{"bucket"})
	if err != nil {
		logger.Error("failed to configure required flags: %v", err)
	}

	return cmd
}

func (o *snapshotS3Options) run(args []string) error {
	envName := args[0]
	url := fmt.Sprintf("%s/api/v2/environments/%s/%s/report/S3", global.Host, global.Org, envName)

	s3Data, err := o.awsStaticCreds.GetS3Data(o.bucket, o.includePaths, o.excludePaths, logger)
	if err != nil {
		return err
	}
	payload := &aws.S3EnvRequest{
		Artifacts: s3Data,
	}

	reqParams := &requests.RequestParams{
		Method:   http.MethodPut,
		URL:      url,
		Payload:  payload,
		DryRun:   global.DryRun,
		Password: global.ApiToken,
	}
	_, err = kosliClient.Do(reqParams)
	if err == nil && global.DryRun == "false" {
		logger.Info("bucket %s was reported to environment %s", o.bucket, envName)
	}
	return err
}
