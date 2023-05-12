---
title: "kosli list artifacts"
experimental: false
---

# kosli list artifacts

## Synopsis

List artifacts in a flow. The results are paginated and ordered from latest to oldest.
By default, the page limit is 15 artifacts per page.


```shell
kosli list artifacts [flags]
```

## Flags
| Flag | Description |
| :--- | :--- |
|    -f, --flow string  |  The Kosli flow name.  |
|    -h, --help  |  help for artifacts  |
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
|        --org string  |  The Kosli organization.  |


## Examples

```shell

# list the last 15 artifacts for a flow:
kosli list artifacts \
	--flow yourFlowName \
	--api-token yourAPIToken \
	--org yourOrgName

# list the last 30 artifacts for a flow:
kosli list artifacts \
	--flow yourFlowName \
	--page-limit 30 \
	--api-token yourAPIToken \
	--org yourOrgName

# list the last 30 artifacts for a flow (in JSON):
kosli list artifacts \
	--flow yourFlowName \	
	--page-limit 30 \
	--api-token yourAPIToken \
	--org yourOrgName \
	--output json

```

