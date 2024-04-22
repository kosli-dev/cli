---
title: "kosli tag"
beta: false
deprecated: false
---

# kosli tag

## Synopsis

Tag a resource in Kosli with key-value pairs.  
use --set to add or update tags, and --unset to remove tags.


```shell
kosli tag RESOURCE-TYPE RESOURCE-ID [flags]
```

## Flags
| Flag | Description |
| :--- | :--- |
|    -D, --dry-run  |  [optional] Run in dry-run mode. When enabled, no data is sent to Kosli and the CLI exits with 0 exit code regardless of any errors.  |
|    -h, --help  |  help for tag  |
|        --set stringToString  |  [optional] The key-value pairs to tag the resource with. The format is: key=value  |
|        --unset strings  |  [optional] The list of tag keys to remove from the resource.  |


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

# add/update tags to a flow
kosli tag flow yourFlowName \
	--set key1=value1 \
	--set key2=value2 \
	--api-token yourApiToken \
	--org yourOrgName

# tag an environment
kosli tag env yourEnvironmentName \
	--set key1=value1 \
	--set key2=value2 \
	--api-token yourApiToken \
	--org yourOrgName

# add/update tags to an environment
kosli tag env yourEnvironmentName \
	--set key1=value1 \
	--set key2=value2 \
	--api-token yourApiToken \
	--org yourOrgName

# remove tags from an environment
kosli tag env yourEnvironmentName \
	--unset key1=value1 \
	--api-token yourApiToken \
	--org yourOrgName

```

