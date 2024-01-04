---
title: "kosli environment log"
---

# kosli environment log

## Synopsis

List environment events.
The results are paginated and ordered from latests to oldest. 
By default, the page limit is 15 events per page.

You can optionally specify an INTERVAL between two snapshot expressions with <expression>..<expression>.
Expressions can be:
	~N   N'th behind the latest snapshot
	N    snapshot number N
	NOW  the latest snapshot
Either expression can be omitted to default to NOW.


```shell
kosli environment log ENV_NAME [INTERVAL] [flags]
```

## Flags
| Flag | Description |
| :--- | :--- |
|    -h, --help  |  help for log  |
|    -l, --long  |  [optional] Print detailed output.  |
|    -o, --output string  |  [defaulted] The format of the output. Valid formats are: [table, json]. (default "table")  |
|        --page int  |  [defaulted] The page number of a response. (default 1)  |
|    -n, --page-limit int  |  [defaulted] The number of elements per page. (default 15)  |
|        --reverse  |  [defaulted] Reverse the order of output list.  |


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

# list the last 15 events for an environment:
kosli environment log yourEnvironmentName \
	--api-token yourAPIToken \
	--owner yourOrgName

# list the last 30 events for an environment:
kosli environment log yourEnvironmentName \
	--page-limit 30 \
	--api-token yourAPIToken \
	--owner yourOrgName

# list the last 30 events for an environment (in JSON):
kosli environment log yourEnvironmentName \
	--page-limit 30 \
	--api-token yourAPIToken \
	--owner yourOrgName \
	--output json

```

