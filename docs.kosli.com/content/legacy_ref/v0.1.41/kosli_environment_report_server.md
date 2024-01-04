---
title: "kosli environment report server"
---

# kosli environment report server

## Synopsis

Report artifacts running in a server environment to Kosli.
You can report directory or file artifacts in one or more server paths.

```shell
kosli environment report server ENVIRONMENT-NAME [flags]
```

## Flags
| Flag | Description |
| :--- | :--- |
|    -D, --dry-run  |  [optional] Run in dry-run mode. When enabled, no data is sent to Kosli and the CLI exits with 0 exit code regardless of any errors.  |
|    -h, --help  |  help for server  |
|    -p, --paths strings  |  The comma separated list of artifact directories.  |


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

# report directory artifacts running in a server at a list of paths:
kosli environment report server yourEnvironmentName \
	--paths a/b/c,e/f/g \
	--api-token yourAPIToken \
	--owner yourOrgName  
```

