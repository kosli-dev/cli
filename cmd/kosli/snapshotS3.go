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

By default the bucket content is fingerprinted by downloading every matching object and hashing it. ^--fingerprint-source metadata^ reads the SHA256 checksum S3 stores for each object instead, which avoids the download, the temporary disk space and the local hashing. Both sources produce the same fingerprint, so a snapshot matches the artifact you attested either way.

Fingerprinting from metadata comes with three conditions:
- Every matching object must carry a full-object SHA256 checksum. S3 only stores one when the upload asked for it, for example ^aws s3api put-object --checksum-algorithm SHA256^. Objects uploaded without one are reported, and cannot be fingerprinted this way.
- A multipart upload gets a composite SHA256, which hashes the checksums of the individual parts rather than the object content, so it cannot be used as the object's fingerprint. Such an object can be collapsed into a single part in place with ^aws s3api copy-object --checksum-algorithm SHA256 --copy-source yourBucket/yourKey --bucket yourBucket --key yourKey^.
- ^.kosli_ignore^ is not applied, because reading it would mean downloading it. A bucket with a ^.kosli_ignore^ at its root is reported rather than fingerprinted without its rules.

It does not reduce the permissions the command needs: AWS requires ^s3:GetObject^ to read an object's checksum, the same permission that downloading it needs. Reading the checksum of an SSE-KMS encrypted object additionally needs ^kms:GenerateDataKey^ and ^kms:Decrypt^.

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

# report contents of an AWS S3 bucket, excluding all PNG files via a regex:
kosli snapshot s3 yourEnvironmentName \
	--bucket yourBucketName \
	--exclude-regex '.*\.png$' \
	--api-token yourAPIToken \
	--org yourOrgName

# report contents of an AWS S3 bucket without downloading the objects,
# using the SHA256 checksums S3 stores for them:
kosli snapshot s3 yourEnvironmentName \
	--bucket yourBucketName \
	--fingerprint-source metadata \
	--api-token yourAPIToken \
	--org yourOrgName
`

// fingerprint sources accepted by --fingerprint-source
const (
	fingerprintSourceContent  = "content"
	fingerprintSourceMetadata = "metadata"
)

type snapshotS3Options struct {
	bucket            string
	includePaths      []string
	includeRegex      []string
	excludePaths      []string
	excludeRegex      []string
	fingerprintSource string
	awsStaticCreds    *aws.AWSStaticCreds
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

			if o.fingerprintSource != fingerprintSourceContent && o.fingerprintSource != fingerprintSourceMetadata {
				return ErrorBeforePrintingUsage(cmd, fmt.Sprintf(
					"%s is not a valid fingerprint source. Valid sources are: [%s]",
					o.fingerprintSource, validS3FingerprintSources))
			}

			return nil
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
	cmd.Flags().StringVar(&o.fingerprintSource, "fingerprint-source", fingerprintSourceContent, s3FingerprintSourceFlag)
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

	harvest := o.awsStaticCreds.GetS3Data
	if o.fingerprintSource == fingerprintSourceMetadata {
		harvest = o.awsStaticCreds.GetS3DataFromMetadata
	}

	s3Data, err := harvest(o.bucket, o.includePaths, o.includeRegex, o.excludePaths, o.excludeRegex, logger)
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
