---
title: "kosli pipeline declare"
---

# kosli pipeline declare

## Synopsis

Create or update a Kosli pipeline.
You can provide a JSON pipefile or specify pipeline parameters in flags. 
The pipefile contains the pipeline metadata and compliance policy (template).

```shell
kosli pipeline declare [flags]
```

## Flags
| Flag | Description |
| :--- | :--- |
|        --description string  |  [optional] The Kosli pipeline description.  |
|    -D, --dry-run  |  [optional] Run in dry-run mode. When enabled, no data is sent to Kosli and the CLI exits with 0 exit code regardless of any errors.  |
|    -h, --help  |  help for declare  |
|        --pipefile string  |  [deprecated] The path to the JSON pipefile.  |
|        --pipeline string  |  The name of the pipeline to be created or updated.  |
|    -t, --template strings  |  [defaulted] The comma-separated list of required compliance controls names. (default [artifact])  |
|        --visibility string  |  [defaulted] The visibility of the Kosli pipeline. Valid visibilities are [public, private]. (default "private")  |


## Options inherited from parent commands
| Flag | Description |
| :--- | :--- |
|    -a, --api-token string  |  The Kosli API token.  |
|    -c, --config-file string  |  [optional] The Kosli config file path. (default "kosli")  |
|        --debug  |  [optional] Print debug logs to stdout.  |
|    -H, --host string  |  [defaulted] The Kosli endpoint. (default "https://app.kosli.com")  |
|    -r, --max-api-retries int  |  [defaulted] How many times should API calls be retried when the API host is not reachable. (default 3)  |
|        --owner string  |  The Kosli user or organization.  |


## Examples

```shell

# create/update a Kosli pipeline without a pipefile:
kosli pipeline declare \
	--pipeline yourPipelineName \
	--description yourPipelineDescription \
    --visibility private OR public \
	--template artifact,evidence-type1,evidence-type2 \
	--api-token yourAPIToken \
	--owner yourOrgName

# create/update a Kosli pipeline with a pipefile (this is a legacy way which will be removed in the future):
kosli pipeline declare \
	--pipefile /path/to/pipefile.json \
	--api-token yourAPIToken \
	--owner yourOrgName

The pipefile format is:
{
    "name": "yourPipelineName",
    "description": "yourPipelineDescription",
    "visibility": "public or private",
    "template": [
        "artifact",
        "evidence-type1",
        "evidence-type2"
    ]
}

```

