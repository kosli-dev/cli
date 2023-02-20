---
title: "kosli approval approval"
---

## kosli approval approval

Get an approval from a specified flow.

### Synopsis

Get an approval from a specified flow.
The expected argument is an expression to specify the approval to get.
It has the format <FLOW_NAME>[SEPARATOR][INTEGER_REFERENCE]

Specify SNAPPISH by:
	flowName~<N>  N'th behind the latest approval
	flowName#<N>  approval number N
	flowName      the latest approval

Examples of valid expressions are: flow (latest approval), flow#10 (approval number 10), flow~2 (the third latest approval)

```shell
kosli approval approval SNAPPISH [flags]
```

### Flags
| Flag | Description |
| :--- | :--- |
|    -h, --help  |  help for approval  |
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

# get second behind the latest approval from a flow
kosli get approval flowName~1 \
	--api-token yourAPIToken \
	--owner orgName

# get the 10th approval from a flow
kosli get approval flowName#10 \
	--api-token yourAPIToken \
	--owner orgName

# get the latest approval from a flow
kosli get approval flowName \
	--api-token yourAPIToken \
	--owner orgName
```

