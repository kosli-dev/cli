---
title: "kosli create environment"
beta: false
deprecated: false
---

# kosli create environment

## Synopsis

Create or update a Kosli environment.

The **--type** must match the type of environment you wish to record snapshots from.
The following types are supported:
  - k8s        - Kubernetes
  - ecs        - Amazon Elastic Container Service
  - s3         - Amazon S3 object storage
  - lambda     - AWS Lambda serverless
  - docker     - Docker images
  - azure-apps - Azure app services
  - server     - Generic type

By default kosli will not make new snapshots for scaling events (change in number of instances running).
For large clusters the scaling events will often outnumber the actual change of SW.

It is possible to enable new snapshots for scaling events with the --include-scaling flag, or turn
it off again with the --exclude-scaling.


```shell
kosli create environment ENVIRONMENT-NAME [flags]
```

## Flags
| Flag | Description |
| :--- | :--- |
|    -d, --description string  |  [optional] The environment description.  |
|    -D, --dry-run  |  [optional] Run in dry-run mode. When enabled, no data is sent to Kosli and the CLI exits with 0 exit code regardless of any errors.  |
|        --exclude-scaling  |  [optional] Exclude scaling events for snapshots. Snapshots with scaling changes will not result in new environment records.  |
|    -h, --help  |  help for environment  |
|        --include-scaling  |  [optional] Include scaling events for snapshots. Snapshots with scaling changes will result in new environment records.  |
|    -t, --type string  |  The type of environment. Valid types are: [K8S, ECS, server, S3, lambda, docker, azure-apps].  |


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

# create a Kosli environment:
kosli create environment yourEnvironmentName
	--type K8S \
	--description "my new env" \
	--api-token yourAPIToken \
	--org yourOrgName 

```

