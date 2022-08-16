---
title: "kosli environment report ecs"
---

## kosli environment report ecs

Report images data from AWS ECS cluster to Kosli.

### Synopsis


List the artifacts deployed in an AWS ECS cluster and their digests 
and report them to Kosli. 


```shell
kosli environment report ecs ENVIRONMENT-NAME [flags]
```

### Flags
| Flag | Description |
| :--- | :--- |
|    -C, --cluster string  |  The name of the ECS cluster.  |
|    -h, --help  |  help for ecs  |
|    -s, --service-name string  |  The name of the ECS service.  |


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

# report what is running in an entire AWS ECS cluster:
export AWS_REGION=yourAWSRegion
export AWS_ACCESS_KEY_ID=yourAWSAccessKeyID
export AWS_SECRET_ACCESS_KEY=yourAWSSecretAccessKey

kosli environment report ecs yourEnvironmentName \
	--api-token yourAPIToken \
	--owner yourOrgName

```

