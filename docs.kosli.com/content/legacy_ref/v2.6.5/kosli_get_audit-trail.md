---
title: "kosli get audit-trail"
beta: true
---

# kosli get audit-trail

{{< hint warning >}}**kosli get audit-trail** is a beta feature. 
Beta features provide early access to product functionality. These features may change between releases without warning, or can be removed from a future release.
You can enable beta features by using the `kosli enable beta` command.{{< /hint >}}
## Synopsis

Get the metadata of a specific audit trail.

```shell
kosli get audit-trail AUDIT-TRAIL-NAME [flags]
```

## Flags
| Flag | Description |
| :--- | :--- |
|    -h, --help  |  help for audit-trail  |
|    -o, --output string  |  [defaulted] The format of the output. Valid formats are: [table, json]. (default "table")  |


## Options inherited from parent commands
| Flag | Description |
| :--- | :--- |
|    -a, --api-token string  |  The Kosli API token.  |
|    -c, --config-file string  |  [optional] The Kosli config file path. (default "kosli")  |
|        --debug  |  [optional] Print debug logs to stdout.  |
|    -H, --host string  |  [defaulted] The Kosli endpoint. (default "https://app.kosli.com")  |
|    -r, --max-api-retries int  |  [defaulted] How many times should API calls be retried when the API host is not reachable. (default 3)  |
|        --org string  |  The Kosli organization.  |


