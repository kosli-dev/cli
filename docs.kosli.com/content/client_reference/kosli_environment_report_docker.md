---
title: "kosli environment report docker"
---

## kosli environment report docker

Report running containers data from docker host to Kosli.

### Synopsis


List the artifacts running as containers and their digests 
and report them to Kosli. 


```shell
kosli environment report docker ENVIRONMENT-NAME [flags]
```

### Flags
| Flag | Description |
| :--- | :--- |
|    -h, --help  |  help for docker  |


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

# report what is running in a docker host:
kosli environment report docker yourEnvironmentName \
	--api-token yourAPIToken \
	--owner yourOrgName

```

