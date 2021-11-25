---
title: "merkely create pipeline"
---

## merkely create pipeline

Create a Merkely pipeline

### Synopsis


Create a Merkely pipeline by providing a JSON pipefile.
The pipefile contains the pipeline metadata and compliance template.


```
merkely create pipeline [flags]
```

### Examples

```

* create a Merkely pipeline with a pipefile:
merkely create pipeline --api-token 1234 /path/to/pipefile.json

* The pipefile format is:
{
    "owner": "organization-name",
    "name": "pipeline-name",
    "description": "pipeline short description",
    "visibility": "public or private",
    "template": [
        "artifact",
        "evidence-type1",
		"evidence-type2"
    ]
}

```

### Options

```
  -h, --help   help for pipeline
```

### Options inherited from parent commands

```
  -a, --api-token string      The merkely API token.
  -c, --config-file string    [optional] The merkely config file path. (default "merkely")
  -D, --dry-run               Whether to send the request to the endpoint or just log it in stdout.
  -H, --host string           The merkely endpoint. (default "https://app.merkely.com")
  -r, --max-api-retries int   How many times should API calls be retried when the API host is not reachable. (default 3)
  -o, --owner string          The merkely organization.
  -v, --verbose               Print verbose logs to stdout.
```

### SEE ALSO

* [merkely create](/client_reference/merkely_create/)	 - Create objects in Merkely.

