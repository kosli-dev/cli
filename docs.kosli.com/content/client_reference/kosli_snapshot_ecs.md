---
title: "kosli snapshot ecs"
beta: false
deprecated: false
summary: "Report a snapshot of running containers in one or more AWS ECS cluster(s) to Kosli.  "
---

# kosli snapshot ecs

## Synopsis

Report a snapshot of running containers in one or more AWS ECS cluster(s) to Kosli.  
Skip `--clusters` and `--clusters-regex` to report all clusters in a given AWS account. Or use `--exclude` and/or `--exclude-regex` to report all clusters excluding some.
The reported data includes container image digests and creation timestamps.

To authenticate to AWS, you can either:  
  1) provide the AWS static credentials via flags or by exporting the equivalent KOSLI env vars (e.g. KOSLI_AWS_KEY_ID)  
  2) export the AWS env vars (e.g. AWS_ACCESS_KEY_ID).  
  3) Use a shared config/credentials file under the $HOME/.aws  
  
Option 1 takes highest precedence, while option 3 is the lowest.  
More details can be found here: https://aws.github.io/aws-sdk-go-v2/docs/configuring-sdk/#specifying-credentials
	

```shell
kosli snapshot ecs ENVIRONMENT-NAME [flags]
```

## Flags
| Flag | Description |
| :--- | :--- |
|        --aws-key-id string  |  The AWS access key ID.  |
|        --aws-region string  |  The AWS region.  |
|        --aws-secret-key string  |  The AWS secret access key.  |
|        --clusters strings  |  [optional] The comma-separated list of ECS cluster names to snapshot. Can't be used together with --exclude or --exclude-regex.  |
|        --clusters-regex strings  |  [optional] The comma-separated list of ECS cluster name regex patterns to snapshot. Can't be used together with --exclude or --exclude-regex.  |
|    -D, --dry-run  |  [optional] Run in dry-run mode. When enabled, no data is sent to Kosli and the CLI exits with 0 exit code regardless of any errors.  |
|        --exclude strings  |  [optional] The comma-separated list of ECS cluster names to exclude. Can't be used together with --exclude or --exclude-regex.  |
|        --exclude-regex strings  |  [optional] The comma-separated list of ECS cluster name regex patterns to exclude. Can't be used together with --clusters or --clusters-regex.  |
|    -h, --help  |  help for ecs  |


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

**report what is running in an entire AWS ECS cluster**

```shell
export AWS_REGION=yourAWSRegion
export AWS_ACCESS_KEY_ID=yourAWSAccessKeyID
export AWS_SECRET_ACCESS_KEY=yourAWSSecretAccessKey

kosli snapshot ecs yourEnvironmentName \
	--clusters yourECSClusterName \
	--api-token yourAPIToken \
	--org yourOrgName

```

**report what is running in a specific AWS ECS service within a cluster**

```shell
export AWS_REGION=yourAWSRegion
export AWS_ACCESS_KEY_ID=yourAWSAccessKeyID
export AWS_SECRET_ACCESS_KEY=yourAWSSecretAccessKey

kosli snapshot ecs yourEnvironmentName \
	--clusters yourECSClusterName \
	--service-name yourECSServiceName \
	--api-token yourAPIToken \
	--org yourOrgName

```

**report what is running in all ECS clusters in an AWS account (AWS auth provided in flags)**

```shell
kosli snapshot ecs yourEnvironmentName \
	--aws-key-id yourAWSAccessKeyID \
	--aws-secret-key yourAWSSecretAccessKey \
	--aws-region yourAWSRegion \
	--api-token yourAPIToken \
	--org yourOrgName

```

**report what is running in all ECS clusters in an AWS account except for clusters with names matching given regex patterns**

```shell
kosli snapshot ecs yourEnvironmentName \
	--aws-key-id yourAWSAccessKeyID \
	--aws-secret-key yourAWSSecretAccessKey \
	--aws-region yourAWSRegion \
	--exclude-regex "those-names.*" \
	--api-token yourAPIToken \
	--org yourOrgName
```

