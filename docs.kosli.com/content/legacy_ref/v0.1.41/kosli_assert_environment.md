---
title: "kosli assert environment"
---

# kosli assert environment

## Synopsis

Assert the compliance status of an environment in Kosli.
Exits with non-zero code if the environment has a non-compliant status.
The expected argument is an expression to specify the specific environment snapshot to assert.
It has the format <ENVIRONMENT_NAME>[SEPARATOR][SNAPSHOT_REFERENCE] 

Separators can be:
- '#' to specify a specific snapshot number for the environment that is being asserted.
- '~' to get N-th behind the latest snapshot.

Examples of valid expressions are: 
- prod (latest snapshot of prod)
- prod#10 (snapshot number 10 of prod)
- prod~2 (third latest snapshot of prod)


```shell
kosli assert environment ENVIRONMENT-NAME-OR-EXPRESSION [flags]
```

## Flags
| Flag | Description |
| :--- | :--- |
|    -D, --dry-run  |  [optional] Run in dry-run mode. When enabled, no data is sent to Kosli and the CLI exits with 0 exit code regardless of any errors.  |
|    -h, --help  |  help for environment  |


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

kosli assert environment prod#5 \
	--api-token yourAPIToken \
	--owner yourOrgName

```

