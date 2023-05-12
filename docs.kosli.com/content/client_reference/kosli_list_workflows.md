---
title: "kosli list workflows"
experimental: true
---

# kosli list workflows

{{< hint warning >}}**kosli list workflows** is an experimental feature. 
Experimental features provide early access to product functionality. These features may change between releases without warning, or can be removed from a future release.
You can enable/disable experimental features by using the **kosli config-experimental** command.{{< /hint >}}
## Synopsis

List workflows for an audit trail.

```shell
kosli list workflows [flags]
```

## Flags
| Flag | Description |
| :--- | :--- |
|        --audit-trail string  |  The Kosli audit trail name.  |
|    -h, --help  |  help for workflows  |
|    -o, --output string  |  [defaulted] The format of the output. Valid formats are: [table, json]. (default "table")  |
|        --page int  |  [defaulted] The page number of a response. (default 1)  |
|    -n, --page-limit int  |  [defaulted] The number of elements per page. (default 15)  |


## Options inherited from parent commands
| Flag | Description |
| :--- | :--- |
|    -a, --api-token string  |  The Kosli API token.  |
|    -c, --config-file string  |  [optional] The Kosli config file path. (default "kosli")  |
|        --debug  |  [optional] Print debug logs to stdout.  |
|    -H, --host string  |  [defaulted] The Kosli endpoint. (default "https://app.kosli.com")  |
|    -r, --max-api-retries int  |  [defaulted] How many times should API calls be retried when the API host is not reachable. (default 3)  |
|        --org string  |  The Kosli organization.  |


