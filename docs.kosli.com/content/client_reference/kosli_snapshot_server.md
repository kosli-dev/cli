---
title: "kosli snapshot server"
beta: false
---

# kosli snapshot server

## Synopsis

Report a snapshot of artifacts running in a server environment to Kosli.  
You can report directory or file artifacts in one or more server paths.

When fingerprinting a 'dir' artifact, you can exclude certain paths from fingerprint calculation 
using the '--exclude' flag.  
Excluded paths are relative to the artifact path(s) and can be literal paths or
glob patterns.  
The supported glob pattern syntax is what is documented here: https://pkg.go.dev/path/filepath#Match

```shell
kosli snapshot server ENVIRONMENT-NAME [flags]
```

## Flags
| Flag | Description |
| :--- | :--- |
|    -D, --dry-run  |  [optional] Run in dry-run mode. When enabled, no data is sent to Kosli and the CLI exits with 0 exit code regardless of any errors.  |
|    -e, --exclude strings  |  [optional] The comma separated list of directories and files to exclude from fingerprinting.  |
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
|        --org string  |  The Kosli organization.  |


## Examples

```shell

# report directory artifacts running in a server at a list of paths:
kosli snapshot server yourEnvironmentName \
	--paths a/b/c,e/f/g \
	--api-token yourAPIToken \
	--org yourOrgName  
	
# exclude certain paths when reporting directory artifacts: 
# the example below, any path matching [a/b/c/logs, a/b/c/*/logs, a/b/c/*/*/logs]
# will be skipped when calculating the fingerprint
kosli snapshot server yourEnvironmentName \
	--paths a/b/c \
	--exclude logs,"*/logs","*/*/logs"
	--api-token yourAPIToken \
	--org yourOrgName  

```

