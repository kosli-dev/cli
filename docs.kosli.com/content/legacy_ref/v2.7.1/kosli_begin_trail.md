---
title: "kosli begin trail"
beta: false
---

# kosli begin trail

## Synopsis

Begin or update a Kosli flow trail.

```shell
kosli begin trail TRAIL-NAME [flags]
```

## Flags
| Flag | Description |
| :--- | :--- |
|        --description string  |  [optional] The Kosli trail description.  |
|    -D, --dry-run  |  [optional] Run in dry-run mode. When enabled, no data is sent to Kosli and the CLI exits with 0 exit code regardless of any errors.  |
|        --flow string  |  The Kosli flow name.  |
|    -h, --help  |  help for trail  |
|    -f, --template-file string  |  The path to a yaml template file.  |
|    -u, --user-data string  |  [optional] The path to a JSON file containing additional data you would like to attach to the flow trail.  |


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

# begin/update a Kosli flow trail:
kosli begin trail yourTrailName \
	--description yourTrailDescription \
	--template-file /path/to/your/template/file.yml \
	--user-data /path/to/your/user-data/file.json \
	--api-token yourAPIToken \
	--org yourOrgName

```

