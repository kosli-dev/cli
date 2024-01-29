---
title: "kosli list audit-trails"
beta: false
deprecated: true
---

# kosli list audit-trails

{{< hint danger >}}**kosli list audit-trails** is a deprecated. Audit trails are deprecated. Please use Flows and Trail instead.  Deprecated commands will be removed in a future release.{{< /hint >}}
## Synopsis

List audit trails for an org.

```shell
kosli list audit-trails [flags]
```

## Flags
| Flag | Description |
| :--- | :--- |
|    -h, --help  |  help for audit-trails  |
|    -o, --output string  |  [defaulted] The format of the output. Valid formats are: [table, json]. (default "table")  |


## Options inherited from parent commands
| Flag | Description |
| :--- | :--- |
|    -a, --api-token string  |  The Kosli API token.  |
|    -c, --config-file string  |  [optional] The Kosli config file path. (default "kosli")  |
|        --debug  |  [optional] Print debug logs to stdout. A boolean flag https://docs.kosli.com/faq/#boolean-flags (default false)  |
|    -H, --host string  |  [defaulted] The Kosli endpoint. (default "https://app.kosli.com")  |
|    -r, --max-api-retries int  |  [defaulted] How many times should API calls be retried when the API host is not reachable. (default 3)  |
|        --org string  |  The Kosli organization.  |


