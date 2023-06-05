---
title: "kosli environment report ecs"
---

# kosli environment report ecs

## Synopsis

Report running containers data from AWS ECS cluster or service to Kosli.
The reported data includes container image digests and creation timestamps.
To authenticate to AWS, you can either: 
	1) provide the AWS static credentials via flags or by exporting the equivalent KOSLI env vars (e.g. KOSLI_AWS_KEY_ID)
	2) export the AWS env vars (e.g. AWS_ACCESS_KEY_ID).
	3) Use a shared config/credentials file under the $HOME/.aws
Option 1 takes highest precedence, while option 3 is the lowest.
More details can be found here: https://aws.github.io/aws-sdk-go-v2/docs/configuring-sdk/#specifying-credentials
	

```shell
kosli environment report ecs ENVIRONMENT-NAME [flags]
```

## Flags
| Flag | Description |
| :--- | :--- |
|        --aws-key-id string  |  The AWS access key ID.  |
|        --aws-region string  |  The AWS region.  |
|        --aws-secret-key string  |  The AWS secret access key.  |
|    -C, --cluster string  |  The name of the ECS cluster.  |
|    -D, --dry-run  |  [optional] Run in dry-run mode. When enabled, no data is sent to Kosli and the CLI exits with 0 exit code regardless of any errors.  |
|    -h, --help  |  help for ecs  |
|    -s, --service-name string  |  The name of the ECS service.  |


## Options inherited from parent commands
| Flag | Description |
| :--- | :--- |
|    -a, --api-token string  |  The Kosli API token.  |
|    -c, --config-file string  |  [optional] The Kosli config file path. (default "kosli")  |
|        --debug  |  [optional] Print debug logs to stdout.  |
|    -H, --host string  |  [defaulted] The Kosli endpoint. (default "https://app.kosli.com")  |
|    -r, --max-api-retries int  |  [defaulted] How many times should API calls be retried when the API host is not reachable. (default 3)  |
|        --owner string  |  The Kosli user or organization.  |


## Examples

```shell

# report what is running in an entire AWS ECS cluster:
export AWS_REGION=yourAWSRegion
export AWS_ACCESS_KEY_ID=yourAWSAccessKeyID
export AWS_SECRET_ACCESS_KEY=yourAWSSecretAccessKey

kosli environment report ecs yourEnvironmentName \
	--cluster yourECSClusterName \
	--api-token yourAPIToken \
	--owner yourOrgName

# report what is running in a specific AWS ECS service:
export AWS_REGION=yourAWSRegion
export AWS_ACCESS_KEY_ID=yourAWSAccessKeyID
export AWS_SECRET_ACCESS_KEY=yourAWSSecretAccessKey

kosli environment report ecs yourEnvironmentName \
	--service-name yourECSServiceName \
	--api-token yourAPIToken \
	--owner yourOrgName

# report what is running in in a specific AWS ECS service (AWS auth provided in flags):
kosli environment report ecs yourEnvironmentName \
	--service-name yourECSServiceName \
	--aws-key-id yourAWSAccessKeyID \
	--aws-secret-key yourAWSSecretAccessKey \
	--aws-region yourAWSRegion \
	--api-token yourAPIToken \
	--owner yourOrgName

```

