---
title: "kosli environment declare"
---

## kosli environment declare

Declare or update a Kosli environment

### Synopsis


Declare or update a Kosli environment.


```shell
kosli environment declare [flags]
```

### Flags
| Flag | Description |
| :--- | :--- |
|    -d, --description string  |  [optional] The environment description.  |
|    -t, --environment-type string  |  The type of environment. Valid options are: [K8S, ECS, server, S3]  |
|    -h, --help  |  help for declare  |
|    -n, --name string  |  The name of environment to be created.  |


### Options inherited from parent commands
| Flag | Description |
| :--- | :--- |
|    -a, --api-token string  |  The Kosli API token.  |
|    -c, --config-file string  |  [optional] The Kosli config file path. (default "merkely")  |
|    -D, --dry-run  |  Whether to run in dry-run mode. When enabled, data is not sent to Kosli and the CLI exits with 0 exit code regardless of errors.  |
|    -H, --host string  |  The Kosli endpoint. (default "https://app.merkely.com")  |
|    -r, --max-api-retries int  |  How many times should API calls be retried when the API host is not reachable. (default 3)  |
|    -o, --owner string  |  The Kosli user or organization.  |
|    -v, --verbose  |  Print verbose logs to stdout.  |


### Examples

```shell

# declare (or update) a Kosli environment:
kosli environment declare 
	--name yourEnvironmentName \
	--environment-type K8S \
	--description "my new env" \
	--api-token yourAPIToken \
	--owner yourOrgName 

```

