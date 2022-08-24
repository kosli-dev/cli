---
title: "kosli deployment get"
---

## kosli deployment get

Get a deployment from a specified pipeline

### Synopsis

Get a deployment from a specified pipeline

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
|    -D, --dry-run  |  [optional] Whether to run in dry-run mode. When enabled, data is not sent to Kosli and the CLI exits with 0 exit code regardless of errors.  |
|    -H, --host string  |  [defaulted] The Kosli endpoint. (default "https://app.kosli.com")  |
|    -r, --max-api-retries int  |  [defaulted] How many times should API calls be retried when the API host is not reachable. (default 3)  |
|        --owner string  |  The Kosli user or organization.  |
|    -v, --verbose  |  [optional] Print verbose logs to stdout.  |


### Examples

```shell

# get the latest deployment in a pipeline
kosli deployment get yourPipelineName \
	--api-token yourAPIToken \
	--owner yourOrgName

# get previous deployment in a pipeline
kosli deployment get yourPipelineName~1 \
	--api-token yourAPIToken \
	--owner yourOrgName

# get the 10th deployment in a pipeline
kosli deployment get yourPipelineName#10 \
	--api-token yourAPIToken \
	--owner yourOrgName

```

