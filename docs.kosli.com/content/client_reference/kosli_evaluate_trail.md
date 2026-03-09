---
title: "kosli evaluate trail"
beta: false
deprecated: false
summary: "Evaluate a trail against a policy."
---

# kosli evaluate trail

## Synopsis

```shell
kosli evaluate trail TRAIL-NAME [flags]
```

Evaluate a trail against a policy.

## Flags
| Flag | Description |
| :--- | :--- |
|        --attestations strings  |  [optional] Limit which attestations are included. Plain name for trail-level, dot-qualified (artifact.name) for artifact-level.  |
|    -f, --flow string  |  The Kosli flow name.  |
|    -h, --help  |  help for trail  |
|    -o, --output string  |  [defaulted] The format of the output. Valid formats are: [table, json]. (default "table")  |
|    -p, --policy string  |  Path to a Rego policy file to evaluate against the trail.  |
|        --show-input  |  [optional] Include the policy input data in the output.  |


## Flags inherited from parent commands
| Flag | Description |
| :--- | :--- |
|    -a, --api-token string  |  The Kosli API token.  |
|    -c, --config-file string  |  [optional] The Kosli config file path. (default "kosli")  |
|        --debug  |  [optional] Print debug logs to stdout. A boolean flag https://docs.kosli.com/faq/#boolean-flags (default false)  |
|    -H, --host string  |  [defaulted] The Kosli endpoint. (default "https://app.kosli.com")  |
|        --http-proxy string  |  [optional] The HTTP proxy URL including protocol and port number. e.g. 'http://proxy-server-ip:proxy-port'  |
|    -r, --max-api-retries int  |  [defaulted] How many times should API calls be retried when the API host is not reachable. (default 3)  |
|        --org string  |  The Kosli organization.  |


