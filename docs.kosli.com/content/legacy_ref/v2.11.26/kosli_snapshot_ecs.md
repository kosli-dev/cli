---
title: "kosli snapshot ecs"
beta: false
deprecated: false
summary: "Report a snapshot of running containers in one or more AWS ECS cluster(s) to Kosli.  "
---

# kosli snapshot ecs

## Synopsis

```shell
kosli snapshot ecs ENVIRONMENT-NAME [flags]
```

Report a snapshot of running containers in one or more AWS ECS cluster(s) to Kosli.  
Skip all filtering flags to report everything running in all clusters in a given AWS account. 

Use `--clusters` and/or `--clusters-regex` OR `--exclude` and/or `--exclude-regex` to filter the clusters to snapshot.
You can also filter the services within a cluster using `--services` and/or `--services-regex`. Or use `--exclude-services` and/or `--exclude-services-regex` to exclude some services. 
Note that service filtering is applied to all clusters being snapshot.

All filtering options are case-sensitive.

The reported data includes cluster and service names, container image digests and creation timestamps.

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
|        --clusters strings  |  [optional] The comma-separated list of ECS cluster names to snapshot. Can't be used together with --exclude or --exclude-regex.  |
|        --clusters-regex strings  |  [optional] The comma-separated list of ECS cluster name regex patterns to snapshot. Can't be used together with --exclude or --exclude-regex.  |
|    -D, --dry-run  |  [optional] Run in dry-run mode. When enabled, no data is sent to Kosli and the CLI exits with 0 exit code regardless of any errors.  |
|        --exclude strings  |  [optional] The comma-separated list of ECS cluster names to exclude. Can't be used together with --clusters or --clusters-regex.  |
|        --exclude-regex strings  |  [optional] The comma-separated list of ECS cluster name regex patterns to exclude. Can't be used together with --clusters or --clusters-regex.  |
|        --exclude-services strings  |  [optional] The comma-separated list of ECS service names to exclude. Can't be used together with --services or --services-regex.  |
|        --exclude-services-regex strings  |  [optional] The comma-separated list of ECS service name regex patterns to exclude. Can't be used together with --services or --services-regex.  |
|    -h, --help  |  help for ecs  |
|        --services strings  |  [optional] The comma-separated list of ECS service names to snapshot. Can't be used together with --exclude-services or --exclude-services-regex.  |
|        --services-regex strings  |  [optional] The comma-separated list of ECS service name regex patterns to snapshot. Can't be used together with --exclude-services or --exclude-services-regex.  |


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

**##### Authentication to AWS ######**

```shell

```

**authentication to AWS using flags**

```shell

kosli snapshot ecs yourEnvironmentName 
	--aws-key-id yourAWSAccessKeyID 
	--aws-secret-key yourAWSSecretAccessKey 
	--aws-region yourAWSRegion 

```

**authentication to AWS using env variables**

```shell

export AWS_ACCESS_KEY_ID=yourAWSAccessKeyID
export AWS_SECRET_ACCESS_KEY=yourAWSSecretAccessKey
export AWS_REGION=yourAWSRegion

kosli snapshot ecs yourEnvironmentName 

```

**##### reporting everything running in all clusters in a given AWS account ######**

```shell

kosli snapshot ecs my-env 

```

**##### filtering which ECS clusters to snapshot ######**

```shell

```

**########## including clusters**

```shell

```

**include clusters matching a name in the AWS account**

```shell
kosli snapshot ecs my-env --clusters my-cluster ...

```

**include clusters matching a pattern in the AWS account**

```shell
kosli snapshot ecs my-env --clusters-regex "my-cluster-*" ...

```

**include clusters matching a list of names in the AWS account**

```shell
kosli snapshot ecs my-env --clusters my-cluster1,my-cluster2 ...

```

**########## excluding clusters**

```shell

```

**exclude clusters matching a name in the AWS account**

```shell
kosli snapshot ecs my-env --exclude my-cluster ...

```

**exclude clusters matching a pattern in the AWS account**

```shell
kosli snapshot ecs my-env --exclude-regex "my-cluster-*" ...

```

**exclude clusters matching a list of names in the AWS account**

```shell
kosli snapshot ecs my-env --exclude my-cluster1,my-cluster2 ...



```

**##### filtering which ECS services to snapshot ######**

```shell

```

**########## including services**

```shell

```

**include Services matching a name in one cluster**

```shell
kosli snapshot ecs my-env --clusters my-cluster --services backend-app ...

```

**include Services matching a pattern in one cluster**

```shell
kosli snapshot ecs my-env --clusters my-cluster --services-regex "backend-*" ...

```

**include production Services only (by naming convention) in all clusters in the AWS account**

```shell
kosli snapshot ecs my-env --services-regex "*-prod-*" ...

```

**include Services matching a name in all clusters in the AWS account**

```shell
kosli snapshot ecs my-env --services backend-app ...

```

**include Services matching a list of names in all clusters in the AWS account**

```shell
kosli snapshot ecs my-env --services backend-app,frontend-app ...

```

**########## excluding services**

```shell

```

**exclude Services matching a pattern in one cluster**

```shell
kosli snapshot ecs my-env --clusters my-cluster --exclude-services-regex "backend-*" ...

```

**exclude Production services only (by naming convention)  in all clusters in the AWS account**

```shell
kosli snapshot ecs my-env --exclude-services-regex "*-prod-*" ...

```

**exclude Services matching a name in one cluster**

```shell
kosli snapshot ecs my-env --clusters my-cluster --exclude-services backend-app ...

```

**exclude Services matching a name in all clusters in the AWS account**

```shell
kosli snapshot ecs my-env --exclude-services backend-app ...

```

**exclude Services matching a list of names in all clusters in the AWS account**

```shell
kosli snapshot ecs my-env --exclude-services backend-app,frontend-app ...
```

