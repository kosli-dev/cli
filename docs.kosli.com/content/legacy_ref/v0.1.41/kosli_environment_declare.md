---
title: "kosli environment declare"
---

# kosli environment declare

## Synopsis

Declare a Kosli environment.

```shell
kosli environment declare [flags]
```

## Flags
| Flag | Description |
| :--- | :--- |
|    -d, --description string  |  [optional] The environment description.  |
|    -D, --dry-run  |  [optional] Run in dry-run mode. When enabled, no data is sent to Kosli and the CLI exits with 0 exit code regardless of any errors.  |
|    -t, --environment-type string  |  The type of environment. Valid types are: [K8S, ECS, server, S3, lambda, docker].  |
|    -h, --help  |  help for declare  |
|    -n, --name string  |  The name of environment to be created.  |


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

# declare a Kosli environment:
kosli environment declare 
	--name yourEnvironmentName \
	--environment-type K8S \
	--description "my new env" \
	--api-token yourAPIToken \
	--owner yourOrgName 

```

