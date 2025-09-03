---
title: "kosli snapshot s3"
beta: false
deprecated: false
summary: "Report a snapshot of the content of an AWS S3 bucket to Kosli."
---

# kosli snapshot s3

## Synopsis

Report a snapshot of the content of an AWS S3 bucket to Kosli.

To authenticate to AWS, you can either:  
  1) provide the AWS static credentials via flags or by exporting the equivalent KOSLI env vars (e.g. KOSLI_AWS_KEY_ID)  
  2) export the AWS env vars (e.g. AWS_ACCESS_KEY_ID).  
  3) Use a shared config/credentials file under the $HOME/.aws  
  
Option 1 takes highest precedence, while option 3 is the lowest.  
More details can be found here: https://aws.github.io/aws-sdk-go-v2/docs/configuring-sdk/#specifying-credentials
	
You can report the entire bucket content, or filter some of the content using `--include` and `--exclude`.
In all cases, the content is reported as one artifact. If you wish to report separate files/dirs within the same bucket as separate artifacts, you need to run the command twice.

To specify paths in a directory artifact that should always be excluded from the SHA256 calculation, you can add a `.kosli_ignore` file to the root of the artifact.
Each line should specify a relative path or path glob to be ignored. You can include comments in this file, using `#`.
The `.kosli_ignore` will be treated as part of the artifact like any other file, unless it is explicitly ignored itself.

```shell
kosli snapshot s3 ENVIRONMENT-NAME [flags]
```

## Flags
| Flag | Description |
| :--- | :--- |
|        --aws-key-id string  |  The AWS access key ID.  |
|        --aws-region string  |  The AWS region.  |
|        --aws-secret-key string  |  The AWS secret access key.  |
|        --bucket string  |  The name of the S3 bucket.  |
|    -D, --dry-run  |  [optional] Run in dry-run mode. When enabled, no data is sent to Kosli and the CLI exits with 0 exit code regardless of any errors.  |
|    -x, --exclude strings  |  [optional] The comma separated list of file and/or directory paths in the S3 bucket to exclude when fingerprinting. Cannot be used together with --include.  |
|    -h, --help  |  help for s3  |
|    -i, --include strings  |  [optional] The comma separated list of file and/or directory paths in the S3 bucket to include when fingerprinting. Cannot be used together with --exclude.  |


## Flags inherited from parent commands
| Flag | Description |
| :--- | :--- |
|    -a, --api-token string  |  The Kosli API token.  |
|    -c, --config-file string  |  [optional] The Kosli config file path. (default "kosli")  |
|        --debug  |  [optional] Print debug logs to stdout. A boolean flag https://docs.kosli.com/faq/#boolean-flags (default false)  |
|    -H, --host string  |  [defaulted] The Kosli endpoint. (default "https://app.kosli.com")  |
|        --http-proxy string  |  [optional] The HTTP proxy URL including protocol and port number. e.g. 'http://proxy-server-ip:proxy-port'  |
|    -r, --max-api-retries int  |  [defaulted] How many times should API calls be retried when the API host is not reachable. (default 3)  |
|        --org string  |  The Kosli organization.  |


## Examples Use Cases

These examples all assume that the flags  `--api-token`, `--org`, `--host`, (and `--flow`, `--trail` when required), are set/provided. 

**report the contents of an entire AWS S3 bucket (AWS auth provided in env variables)**

```shell
export AWS_REGION=yourAWSRegion
export AWS_ACCESS_KEY_ID=yourAWSAccessKeyID
export AWS_SECRET_ACCESS_KEY=yourAWSSecretAccessKey

kosli snapshot s3 yourEnvironmentName 
	--bucket yourBucketName 

```

**report what is running in an AWS S3 bucket (AWS auth provided in flags)**

```shell
kosli snapshot s3 yourEnvironmentName 
	--bucket yourBucketName 
	--aws-key-id yourAWSAccessKeyID 
	--aws-secret-key yourAWSSecretAccessKey 
	--aws-region yourAWSRegion 

```

**report a subset of contents of an AWS S3 bucket (AWS auth provided in env variables)**

```shell
export AWS_REGION=yourAWSRegion
export AWS_ACCESS_KEY_ID=yourAWSAccessKeyID
export AWS_SECRET_ACCESS_KEY=yourAWSSecretAccessKey

kosli snapshot s3 yourEnvironmentName 
	--bucket yourBucketName 
	--include file.txt,path/within/bucket 

```

**report contents of an entire AWS S3 bucket, except for some paths (AWS auth provided in env variables)**

```shell
export AWS_REGION=yourAWSRegion
export AWS_ACCESS_KEY_ID=yourAWSAccessKeyID
export AWS_SECRET_ACCESS_KEY=yourAWSSecretAccessKey

kosli snapshot s3 yourEnvironmentName 
	--bucket yourBucketName 
	--exclude file.txt,path/within/bucket 
```

