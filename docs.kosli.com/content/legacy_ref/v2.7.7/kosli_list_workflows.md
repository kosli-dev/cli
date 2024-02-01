---
title: "kosli list workflows"
beta: false
deprecated: true
---

# kosli list workflows

{{< hint danger >}}**kosli list workflows** is a deprecated. Audit trails are deprecated. Please use Flows and Trail instead.  Deprecated commands will be removed in a future release.{{< /hint >}}
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
|        --debug  |  [optional] Print debug logs to stdout. A boolean flag https://docs.kosli.com/faq/#boolean-flags (default false)  |
|    -H, --host string  |  [defaulted] The Kosli endpoint. (default "https://app.kosli.com")  |
|    -r, --max-api-retries int  |  [defaulted] How many times should API calls be retried when the API host is not reachable. (default 3)  |
|        --org string  |  The Kosli organization.  |


