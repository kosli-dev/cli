---
title: "kosli pipeline declare"
---

## kosli pipeline declare

Declare or update a Kosli pipeline

### Synopsis


Declare or update a Kosli pipeline by providing a JSON pipefile or by providing pipeline parameters in flags. 
The pipefile contains the pipeline metadata and compliance policy.


```shell
kosli pipeline declare [flags]
```

### Flags
| Flag | Description |
| :--- | :--- |
|        --description string  |  [optional] The Kosli pipeline description.  |
|    -h, --help  |  help for declare  |
|        --pipefile string  |  [deprecated] The path to the JSON pipefile.  |
|        --pipeline string  |  The name of the pipeline to be created or updated.  |
|    -t, --template strings  |  The comma-separated list of required compliance controls names. (default [artifact])  |
|        --visibility string  |  The visibility of the Kosli pipeline. Options are [public, private]. (default "private")  |


### Options inherited from parent commands
| Flag | Description |
| :--- | :--- |
|    -a, --api-token string  |  The Kosli API token.  |
|    -c, --config-file string  |  [optional] The Kosli config file path. (default "merkely")  |
|    -D, --dry-run  |  Whether to run in dry-run mode. When enabled, data is not sent to Kosli and the CLI exits with 0 exit code regardless of errors.  |
|    -H, --host string  |  The Kosli endpoint. (default "https://app.merkely.com")  |
|    -r, --max-api-retries int  |  How many times should API calls be retried when the API host is not reachable. (default 3)  |
|    -o, --owner string  |  The Kosli user or organization.  |
|    -v, --verbose  |  Print verbose logs to stdout.  |


### Examples

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
    "description": "yourPipelinedescription",
    "visibility": "public or private",
    "template": [
        "artifact",
        "evidence-type1",
        "evidence-type2"
    ]
}

```

