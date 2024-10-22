---
title: "kosli create environment"
beta: false
deprecated: false
---

# kosli create environment

## Synopsis

Create or update a Kosli environment.

``--type`` must match the type of environment you wish to record snapshots from.
The following types are supported:
  - k8s        - Kubernetes
  - ecs        - Amazon Elastic Container Service
  - s3         - Amazon S3 object storage
  - lambda     - AWS Lambda serverless
  - docker     - Docker images
  - azure-apps - Azure app services
  - server     - Generic type
  - logical    - Logical grouping of real environments

By default, the environment does not require artifacts provenance (i.e. environment snapshots will not 
become non-compliant because of artifacts that do not have provenance). You can require provenance for all artifacts
by setting --require-provenance=true

Also, by default, kosli will not make new snapshots for scaling events (change in number of instances running).
For large clusters the scaling events will often outnumber the actual change of SW.

It is possible to enable new snapshots for scaling events with the --include-scaling flag, or turn
it off again with the --exclude-scaling.

Logical environments are used for grouping of physical environments. For instance **prod-aws** and **prod-s3** can
be grouped into logical environment **prod**. Logical environments are view-only, you can not report snapshots
to them.


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
|        --included-environments strings  |  [optional] Comma separated list of environments to include in logical environment  |
|        --require-provenance  |  [defaulted] Require provenance for all artifacts running in environment snapshots.  |
|    -t, --type string  |  The type of environment. Valid types are: [K8S, ECS, server, S3, lambda, docker, azure-apps, logical].  |


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

**create a Kosli environment**

```shell
kosli create environment yourEnvironmentName
	--type K8S \
	--description "my new env" \
	--api-token yourAPIToken \
	--org yourOrgName 


kosli create environment yourLogicalEnvironmentName
	--type logical \
	--included-environments realEnv1,realEnv2,realEnv3
	--description "my full prod" \	
	--api-token yourAPIToken \
	--org yourOrgName
```

