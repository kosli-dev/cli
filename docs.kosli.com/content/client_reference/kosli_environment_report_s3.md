---
title: "kosli environment report s3"
---

## kosli environment report s3

Report artifact from AWS S3 bucket to Kosli.

### Synopsis


Report the artifact deployed in an AWS S3 bucket and its digest to Kosli. 


```shell
kosli environment report s3 ENVIRONMENT-NAME [flags]
```

### Flags
| Flag | Description |
| :--- | :--- |
|        --aws-key-id string  |  The AWS access key ID.  |
|        --aws-region string  |  The AWS region.  |
|        --aws-secret-key string  |  The AWS secret key.  |
|        --bucket string  |  The name of the S3 bucket.  |
|    -h, --help  |  help for s3  |


### Options inherited from parent commands
| Flag | Description |
| :--- | :--- |
|    -a, --api-token string  |  The Kosli API token.  |
|    -c, --config-file string  |  [optional] The Kosli config file path. (default "kosli")  |
|    -D, --dry-run  |  [optional] Whether to run in dry-run mode. When enabled, data is not sent to Kosli and the CLI exits with 0 exit code regardless of errors.  |
|    -H, --host string  |  [defaulted] The Kosli endpoint. (default "https://app.kosli.com")  |
|    -r, --max-api-retries int  |  [defaulted] How many times should API calls be retried when the API host is not reachable. (default 3)  |
|        --owner string  |  The Kosli user or organization.  |
|    -v, --verbose  |  [optional] Print verbose logs to stdout.  |


### Examples

```shell

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

```

