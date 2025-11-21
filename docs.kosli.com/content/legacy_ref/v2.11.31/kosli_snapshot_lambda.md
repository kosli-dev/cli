---
title: "kosli snapshot lambda"
beta: false
deprecated: false
summary: "Report a snapshot of artifacts deployed as one or more AWS Lambda functions and their digests to Kosli."
---

# kosli snapshot lambda

## Synopsis

```shell
kosli snapshot lambda ENVIRONMENT-NAME [flags]
```

Report a snapshot of artifacts deployed as one or more AWS Lambda functions and their digests to Kosli.  
Skip `--function-names` and `--function-names-regex` to report all functions in a given AWS account. Or use `--exclude` and/or `--exclude-regex` to report all functions excluding some.

To authenticate to AWS, you can either:  
  1) provide the AWS static credentials via flags or by exporting the equivalent KOSLI env vars (e.g. KOSLI_AWS_KEY_ID)  
  2) export the AWS env vars (e.g. AWS_ACCESS_KEY_ID).  
  3) Use a shared config/credentials file under the $HOME/.aws  
  
Option 1 takes highest precedence, while option 3 is the lowest.  
More details can be found here: https://aws.github.io/aws-sdk-go-v2/docs/configuring-sdk/#specifying-credentials
	

## Flags
| Flag | Description |
| :--- | :--- |
|        --aws-key-id string  |  The AWS access key ID.  |
|        --aws-region string  |  The AWS region.  |
|        --aws-secret-key string  |  The AWS secret access key.  |
|    -D, --dry-run  |  [optional] Run in dry-run mode. When enabled, no data is sent to Kosli and the CLI exits with 0 exit code regardless of any errors.  |
|        --exclude strings  |  [optional] The comma-separated list of AWS Lambda function names to be excluded. Cannot be used together with --function-names  |
|        --exclude-regex strings  |  [optional] The comma-separated list of name regex patterns for AWS Lambda functions to be excluded. Cannot be used together with --function-names. Allowed regex patterns are described in https://github.com/google/re2/wiki/Syntax  |
|        --function-names strings  |  [optional] The comma-separated list of AWS Lambda function names to be reported. Cannot be used together with --exclude or --exclude-regex.  |
|        --function-names-regex strings  |  [optional] The comma-separated list of AWS Lambda function names regex patterns to be reported. Cannot be used together with --exclude or --exclude-regex.  |
|    -h, --help  |  help for lambda  |


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

These examples all assume that the flags  `--api-token`, `--org`, `--host`, (and `--flow`, `--trail` when required), are [set/provided](https://docs.kosli.com/getting_started/install/#assigning-flags-via-environment-variables). 

**report all Lambda functions running in an AWS account (AWS auth provided in env variables)**

```shell
export AWS_REGION=yourAWSRegion
export AWS_ACCESS_KEY_ID=yourAWSAccessKeyID
export AWS_SECRET_ACCESS_KEY=yourAWSSecretAccessKey

kosli snapshot lambda yourEnvironmentName 

```

**report all (excluding some) Lambda functions running in an AWS account (AWS auth provided in env variables)**

```shell
export AWS_REGION=yourAWSRegion
export AWS_ACCESS_KEY_ID=yourAWSAccessKeyID
export AWS_SECRET_ACCESS_KEY=yourAWSSecretAccessKey

kosli snapshot lambda yourEnvironmentName 
    --exclude function1,function2 
	--exclude-regex "^not-wanted.*" 

```

**report what is running in the latest version of an AWS Lambda function (AWS auth provided in env variables)**

```shell
export AWS_REGION=yourAWSRegion
export AWS_ACCESS_KEY_ID=yourAWSAccessKeyID
export AWS_SECRET_ACCESS_KEY=yourAWSSecretAccessKey

kosli snapshot lambda yourEnvironmentName 
	--function-names yourFunctionName 

```

**report what is running in the latest version of AWS Lambda functions that match a name regex**

```shell
export AWS_REGION=yourAWSRegion
export AWS_ACCESS_KEY_ID=yourAWSAccessKeyID
export AWS_SECRET_ACCESS_KEY=yourAWSSecretAccessKey

kosli snapshot lambda yourEnvironmentName 
	--function-names-regex yourFunctionNameRegexPattern 

```

**report what is running in the latest version of multiple AWS Lambda functions (AWS auth provided in env variables)**

```shell
export AWS_REGION=yourAWSRegion
export AWS_ACCESS_KEY_ID=yourAWSAccessKeyID
export AWS_SECRET_ACCESS_KEY=yourAWSSecretAccessKey

kosli snapshot lambda yourEnvironmentName 
	--function-names yourFirstFunctionName,yourSecondFunctionName 

```

**report what is running in the latest version of an AWS Lambda function (AWS auth provided in flags)**

```shell
kosli snapshot lambda yourEnvironmentName 
	--function-names yourFunctionName 
	--aws-key-id yourAWSAccessKeyID 
	--aws-secret-key yourAWSSecretAccessKey 
	--aws-region yourAWSRegion 
```

