---
title: "kosli deployment get"
---

## kosli deployment get

Get a deployment from a specified pipeline.

### Synopsis

Get a deployment from a specified pipeline.
Specify SNAPPISH by:
	pipelineName~<N>  N'th behind the latest deployment
	pipelineName#<N>  deployment number N
	pipelineName      the latest deployment

```shell
kosli deployment get SNAPPISH [flags]
```

### Flags
| Flag | Description |
| :--- | :--- |
|    -h, --help  |  help for get  |
|    -o, --output string  |  [defaulted] The format of the output. Valid formats are: [table, json]. (default "table")  |


### Options inherited from parent commands
| Flag | Description |
| :--- | :--- |
|    -a, --api-token string  |  The Kosli API token.  |
|    -c, --config-file string  |  [optional] The Kosli config file path. (default "kosli")  |
|        --debug  |  [optional] Print debug logs to stdout.  |
|    -H, --host string  |  [defaulted] The Kosli endpoint. (default "https://app.kosli.com")  |
|    -r, --max-api-retries int  |  [defaulted] How many times should API calls be retried when the API host is not reachable. (default 3)  |
|        --owner string  |  The Kosli user or organization.  |


### Examples

```shell
# get previous deployment in a pipeline
kosli deployment get pipelineName~1 \
	--api-token yourAPIToken \
	--owner orgName

# get the 10th deployment in a pipeline
kosli deployment get pipelineName#10 \
	--api-token yourAPIToken \
	--owner orgName

# get the latest deployment in a pipeline
kosli deployment get pipelineName \
	--api-token yourAPIToken \
	--owner orgName
```

