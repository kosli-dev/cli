---
title: "merkely environment report s3"
---

## merkely environment report s3

Report artifact from AWS S3 bucket to Merkely.

### Synopsis


Report the artifact deployed in an AWS S3 bucket and its digest to Merkely. 


```shell
merkely environment report s3 env-name [flags]
```

### Examples

```shell

# report what is running in an AWS S3 bucket (AWS auth provided in env variables):
export AWS_REGION=yourAWSRegion
export AWS_ACCESS_KEY_ID=yourAWSAccessKeyID
export AWS_SECRET_ACCESS_KEY=yourAWSSecretAccessKey

merkely environment report s3 yourEnvironmentName \
	--bucket yourBucketName \
	--api-token yourAPIToken \
	--owner yourOrgName

# report what is running in an AWS S3 bucket (AWS auth provided in flags):
merkely environment report s3 yourEnvironmentName \
	--bucket yourBucketName \
	--aws-key-id yourAWSAccessKeyID \
	--aws-secret-key yourAWSSecretAccessKey \
	--aws-region yourAWSRegion \
	--api-token yourAPIToken \
	--owner yourOrgName	

```

### Options
| Flag | Description |
| :--- | :--- |
|        --aws-key-id string  |  The AWS access key ID  |
|        --aws-region string  |  The AWS region  |
|        --aws-secret-key string  |  The AWS secret key  |
|        --bucket string  |  The name of the S3 bucket.  |
|    -h, --help  |  help for s3  |


### Options inherited from parent commands
| Flag | Description |
| :--- | :--- |
|    -a, --api-token string  |  The merkely API token.  |
|    -c, --config-file string  |  [optional] The merkely config file path. (default "merkely")  |
|    -D, --dry-run  |  Whether to run in dry-run mode. When enabled, data is not sent to Merkely and the CLI exits with 0 exit code regardless of errors.  |
|    -H, --host string  |  The merkely endpoint. (default "https://app.merkely.com")  |
|    -r, --max-api-retries int  |  How many times should API calls be retried when the API host is not reachable. (default 3)  |
|    -o, --owner string  |  The merkely user or organization.  |
|    -v, --verbose  |  Print verbose logs to stdout.  |


