---
title: "merkely environment report ecs"
---

## merkely environment report ecs

Report images data from AWS ECS cluster to Merkely.

### Synopsis


List the artifacts deployed in an AWS ECS cluster and their digests 
and report them to Merkely. 


```shell
merkely environment report ecs env-name [flags]
```

### Examples

```shell

# report what is running in an entire AWS ECS cluster:
export AWS_REGION=yourAWSRegion
export AWS_ACCESS_KEY_ID=yourAWSAccessKeyID
export AWS_SECRET_ACCESS_KEY=yourAWSSecretAccessKey

merkely environment report ecs yourEnvironmentName \
	--api-token yourAPIToken \
	--owner yourOrgName

```

### Options
| Flag | Description |
| :--- | :--- |
|    -C, --cluster string  |  The name of the ECS cluster.  |
|    -h, --help  |  help for ecs  |
|    -s, --service-name string  |  The name of the ECS service.  |


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


