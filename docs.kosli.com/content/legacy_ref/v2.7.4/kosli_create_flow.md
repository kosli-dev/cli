---
title: "kosli create flow"
beta: false
---

# kosli create flow

## Synopsis

Create or update a Kosli flow.
You can specify flow parameters in flags.

```shell
kosli create flow FLOW-NAME [flags]
```

## Flags
| Flag | Description |
| :--- | :--- |
|        --description string  |  [optional] The Kosli flow description.  |
|    -D, --dry-run  |  [optional] Run in dry-run mode. When enabled, no data is sent to Kosli and the CLI exits with 0 exit code regardless of any errors.  |
|    -h, --help  |  help for flow  |
|    -t, --template strings  |  [defaulted] The comma-separated list of required compliance controls names.  |
|    -f, --template-file string  |  The path to a yaml template file.  |
|        --visibility string  |  [defaulted] The visibility of the Kosli flow. Valid visibilities are [public, private]. (default "private")  |


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

# create/update a Kosli flow (with legacy template):
kosli create flow yourFlowName \
	--description yourFlowDescription \
    --visibility private OR public \
	--template artifact,evidence-type1,evidence-type2 \
	--api-token yourAPIToken \
	--org yourOrgName

# create/update a Kosli flow (with template file):
kosli create flow yourFlowName \
	--description yourFlowDescription \
	--visibility private OR public \
	--template-file /path/to/your/template/file.yml \
	--api-token yourAPIToken \
	--org yourOrgName

```

