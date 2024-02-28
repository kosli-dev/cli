---
title: "kosli snapshot lambda"
beta: false
deprecated: false
---

# kosli snapshot lambda

## Synopsis

Report a snapshot of artifacts deployed as one or more AWS Lambda functions and their digests to Kosli.  
Skip --function-names to report all functions in a given AWS account.

To authenticate to AWS, you can either:  
  1) provide the AWS static credentials via flags or by exporting the equivalent KOSLI env vars (e.g. KOSLI_AWS_KEY_ID)  
  2) export the AWS env vars (e.g. AWS_ACCESS_KEY_ID).  
  3) Use a shared config/credentials file under the $HOME/.aws  
  
Option 1 takes highest precedence, while option 3 is the lowest.  
More details can be found here: https://aws.github.io/aws-sdk-go-v2/docs/configuring-sdk/#specifying-credentials
	

```shell
kosli snapshot lambda ENVIRONMENT-NAME [flags]
```

## Flags
| Flag | Description |
| :--- | :--- |
|        --aws-key-id string  |  The AWS access key ID.  |
|        --aws-region string  |  The AWS region.  |
|        --aws-secret-key string  |  The AWS secret access key.  |
|    -D, --dry-run  |  [optional] Run in dry-run mode. When enabled, no data is sent to Kosli and the CLI exits with 0 exit code regardless of any errors.  |
|        --function-names strings  |  [optional] The comma-separated list of AWS Lambda function names to be reported.  |
|    -h, --help  |  help for lambda  |


## Options inherited from parent commands
| Flag | Description |
| :--- | :--- |
|    -a, --api-token string  |  The Kosli API token.  |
|    -c, --config-file string  |  [optional] The Kosli config file path. (default "kosli")  |
|        --debug  |  [optional] Print debug logs to stdout. A boolean flag https://docs.kosli.com/faq/#boolean-flags (default false)  |
|    -H, --host string  |  [defaulted] The Kosli endpoint. (default "https://app.kosli.com")  |
|    -r, --max-api-retries int  |  [defaulted] How many times should API calls be retried when the API host is not reachable. (default 3)  |
|        --org string  |  The Kosli organization.  |


## Examples

```shell

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

```
