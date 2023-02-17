---
title: "kosli approval get"
---

## kosli approval get

Get an approval from a specified pipeline.

### Synopsis

Get an approval from a specified pipeline.
The expected argument is an expression to specify the approval to get.
It has the format <PIPELINE_NAME>[SEPARATOR][INTEGER_REFERENCE]

Specify SNAPPISH by:
	pipelineName~<N>  N'th behind the latest approval
	pipelineName#<N>  approval number N
	pipelineName      the latest approval

Examples of valid expressions are: pipe (latest approval), pipe#10 (approval number 10), pipe~2 (the third latest approval)

```shell
kosli approval get SNAPPISH [flags]
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

# get second behind the latest approval from a pipeline
kosli approval get pipelineName~1 \
	--api-token yourAPIToken \
	--owner orgName

# get the 10th approval from a pipeline
kosli approval get pipelineName#10 \
	--api-token yourAPIToken \
	--owner orgName

# get the latest approval from a pipeline
kosli approval get pipelineName \
	--api-token yourAPIToken \
	--owner orgName
```

