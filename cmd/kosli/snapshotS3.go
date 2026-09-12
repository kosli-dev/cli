package main

import (
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/kosli-dev/cli/internal/aws"
	"github.com/kosli-dev/cli/internal/requests"
	"github.com/spf13/cobra"
)

const snapshotS3ShortDesc = `Report a snapshot of the content of an AWS S3 bucket to Kosli.`

const snapshotS3LongDesc = snapshotS3ShortDesc + awsAuthDesc + `
You can report the entire bucket content, or filter some of the content using ^--include^ / ^--exclude^ (literal prefix match) or ^--include-regex^ / ^--exclude-regex^ (Go regular expressions matched against the full object key).
In all cases, the content is reported as one artifact. If you wish to report separate files/dirs within the same bucket as separate artifacts, you need to run the command twice.
Object keys are never used as local file names: each object is downloaded to a temporary file, hashed and removed, and the fingerprint is computed from the keys and the content digests, so any key S3 accepts can be fingerprinted on any operating system.
Keys that cannot form a directory tree are rejected and fail the snapshot, naming every key involved: a key containing a ^..^ segment, two keys that resolve to the same path (such as ^a//b^ and ^a/b^), or an object whose key is also a prefix of other objects (such as ^a^ beside ^a/b^). A legitimate key of that shape can be left out with ^--exclude-regex^ (anchor and escape it, since the pattern is a regular expression matched against the whole key); when ^--include^ or ^--include-regex^ is set, exclude filters are ignored, so narrow the include filter instead.

` + kosliIgnoreDescNoExclude

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

# report contents of an AWS S3 bucket, excluding all PNG files via a regex:
kosli snapshot s3 yourEnvironmentName \
	--bucket yourBucketName \
	--exclude-regex '.*\.png$' \
	--api-token yourAPIToken \
	--org yourOrgName
`

type snapshotS3Options struct {
	bucket              string
	includePaths        []string
	includeRegex        []string
	excludePaths        []string
	excludeRegex        []string
	downloadConcurrency int
	downloadBudget      string
	downloadLimits      aws.DownloadLimits
	awsStaticCreds      *aws.AWSStaticCreds
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

			// Include flags and exclude flags are mutually exclusive
			// in every combination — choose one direction at a time.
			for _, pair := range [][]string{
				{"include", "exclude"},
				{"include", "exclude-regex"},
				{"include-regex", "exclude"},
				{"include-regex", "exclude-regex"},
			} {
				if err = MuXRequiredFlags(cmd, pair, false); err != nil {
					return err
				}
			}

			return o.resolveDownloadLimits()
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return o.run(args)
		},
	}

	cmd.Flags().StringVar(&o.bucket, "bucket", "", bucketNameFlag)
	cmd.Flags().StringSliceVarP(&o.includePaths, "include", "i", []string{}, bucketPathsFlag)
	cmd.Flags().StringSliceVar(&o.includeRegex, "include-regex", []string{}, bucketPathsRegexFlag)
	cmd.Flags().StringSliceVarP(&o.excludePaths, "exclude", "x", []string{}, excludeBucketPathsFlag)
	cmd.Flags().StringSliceVar(&o.excludeRegex, "exclude-regex", []string{}, excludeBucketPathsRegexFlag)
	cmd.Flags().IntVar(&o.downloadConcurrency, "download-concurrency", aws.DefaultDownloadLimits.Concurrency, downloadConcurrencyFlag)
	cmd.Flags().StringVar(&o.downloadBudget, "download-budget", defaultDownloadBudget, downloadBudgetFlag)
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

	if err := ensureEnvironment(envName, "S3"); err != nil {
		return err
	}

	url, err := url.JoinPath(global.Host, "api/v2/environments", global.Org, envName, "report/S3")
	if err != nil {
		return err
	}

	s3Data, err := o.awsStaticCreds.GetS3Data(o.bucket, o.includePaths, o.includeRegex, o.excludePaths, o.excludeRegex, o.downloadLimits, logger)
	if err != nil {
		return err
	}
	payload := &aws.S3EnvRequest{
		Artifacts: s3Data,
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
		logger.Info("bucket %s was reported to environment %s", o.bucket, envName)
	}
	return err
}

// defaultDownloadBudget spells aws.DefaultDownloadLimits.BytesInFlight the way
// the flag reads it.
const defaultDownloadBudget = "512M"

// resolveDownloadLimits validates the download flags and turns them into the
// limits the aws package takes.
func (o *snapshotS3Options) resolveDownloadLimits() error {
	if o.downloadConcurrency < 1 {
		return fmt.Errorf("--download-concurrency must be at least 1, got %d", o.downloadConcurrency)
	}
	budget, err := parseByteSize(o.downloadBudget)
	if err != nil {
		return fmt.Errorf("invalid --download-budget: %v", err)
	}
	o.downloadLimits = aws.DownloadLimits{Concurrency: o.downloadConcurrency, BytesInFlight: budget}
	return nil
}
