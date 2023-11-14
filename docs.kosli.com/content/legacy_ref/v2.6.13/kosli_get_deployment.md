---
title: "kosli get deployment"
beta: false
---

# kosli get deployment

## Synopsis

Get a deployment from a specified flow.
EXPRESSION can be specified as follows:
- flowName
    - the latest deployment to flowName, at the time of the request
    - e.g., **dashboard**
- flowName#N
    - the Nth deployment, counting from 1
    - e.g., **dashboard#453**
- flowName~N
    - the Nth deployment behind the latest, at the time of the request
    - e.g., **dashboard~56**


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
|        --debug  |  [optional] Print debug logs to stdout. A boolean flag https://docs.kosli.com/faq/#boolean-flags (default false)  |
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

