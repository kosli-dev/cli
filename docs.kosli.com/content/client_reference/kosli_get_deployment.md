---
title: "kosli get deployment"
experimental: false
---

# kosli get deployment

## Synopsis

Get a deployment from a specified flow.
Expression can be specified as follows:
- flowName~<N>  N'th behind the latest deployment
- flowName#<N>  deployment number N
- flowName      the latest deployment

```shell
kosli get deployment EXPRESSION [flags]
```

## Flags
| Flag | Description |
| :--- | :--- |
|    -h, --help  |  help for deployment  |
|    -o, --output string  |  [defaulted] The format of the output. Valid formats are: [table, json]. (default "table")  |


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

# get previous deployment in a flow
kosli get deployment flowName~1 \
	--api-token yourAPIToken \
	--org orgName

# get the 10th deployment in a flow
kosli get deployment flowName#10 \
	--api-token yourAPIToken \
	--org orgName

# get the latest deployment in a flow
kosli get deployment flowName \
	--api-token yourAPIToken \
	--org orgName
```

