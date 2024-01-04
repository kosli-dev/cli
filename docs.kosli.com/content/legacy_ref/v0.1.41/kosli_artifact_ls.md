---
title: "kosli artifact ls"
---

# kosli artifact ls

## Synopsis

List artifacts in a pipeline. The results are paginated and ordered from latests to oldest. 
By default, the page limit is 15 artifacts per page.


```shell
kosli artifact ls PIPELINE-NAME [flags]
```

## Flags
| Flag | Description |
| :--- | :--- |
|    -h, --help  |  help for ls  |
|    -o, --output string  |  [defaulted] The format of the output. Valid formats are: [table, json]. (default "table")  |
|        --page int  |  [defaulted] The page number of a response. (default 1)  |
|    -n, --page-limit int  |  [defaulted] The number of elements per page. (default 15)  |


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

# list the last 15 artifacts for a pipeline:
kosli artifact list yourPipelineName \
	--api-token yourAPIToken \
	--owner yourOrgName

# list the last 30 artifacts for a pipeline:
kosli artifact list yourPipelineName \
	--page-limit 30 \
	--api-token yourAPIToken \
	--owner yourOrgName

# list the last 30 artifacts for a pipeline (in JSON):
kosli artifact list yourPipelineName \
	--page-limit 30 \
	--api-token yourAPIToken \
	--owner yourOrgName \
	--output json

```

