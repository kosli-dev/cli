---
title: "kosli environment report lambda"
---

## kosli environment report lambda

Report artifact from AWS Lambda to Kosli.

### Synopsis


Report the artifact deployed in an AWS Lambda and its digest to Kosli. 


```shell
kosli environment report lambda env-name [flags]
```

### Flags
| Flag | Description |
| :--- | :--- |
|        --aws-key-id string  |  The AWS access key ID  |
|        --aws-region string  |  The AWS region  |
|        --aws-secret-key string  |  The AWS secret key  |
|        --function-name string  |  The name of the AWS Lambda function.  |
|        --function-version string  |  [optional] The version of the AWS Lambda function.  |
|    -h, --help  |  help for lambda  |


### Options inherited from parent commands
| Flag | Description |
| :--- | :--- |
|    -a, --api-token string  |  The Kosli API token.  |
|    -c, --config-file string  |  [optional] The Kosli config file path. (default "merkely")  |
|    -D, --dry-run  |  Whether to run in dry-run mode. When enabled, data is not sent to Kosli and the CLI exits with 0 exit code regardless of errors.  |
|    -H, --host string  |  The Kosli endpoint. (default "https://app.merkely.com")  |
|    -r, --max-api-retries int  |  How many times should API calls be retried when the API host is not reachable. (default 3)  |
|    -o, --owner string  |  The Kosli user or organization.  |
|    -v, --verbose  |  Print verbose logs to stdout.  |


### Examples

```shell

# report what is running in the latest version AWS Lambda function (AWS auth provided in env variables):
export AWS_REGION=yourAWSRegion
export AWS_ACCESS_KEY_ID=yourAWSAccessKeyID
export AWS_SECRET_ACCESS_KEY=yourAWSSecretAccessKey

kosli environment report lambda myEnvironment \
	--function-name yourFunctionName \
	--api-token yourAPIToken \
	--owner yourOrgName

# report what is running in a specific version of an AWS Lambda function (AWS auth provided in flags):
kosli environment report lambda myEnvironment \
	--function-name yourFunctionName \
	--function-version yourFunctionVersion \
	--aws-key-id yourAWSAccessKeyID \
	--aws-secret-key yourAWSSecretAccessKey \
	--aws-region yourAWSRegion \
	--api-token yourAPIToken \
	--owner yourOrgName

```

