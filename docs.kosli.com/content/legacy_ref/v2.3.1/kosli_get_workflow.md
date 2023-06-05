---
title: "kosli get workflow"
experimental: true
---

# kosli get workflow

{{< hint warning >}}**kosli get workflow** is an experimental feature. 
Experimental features provide early access to product functionality. These features may change between releases without warning, or can be removed from a future release.
You can enable experimental features by using the **kosli enable experimental** command.{{< /hint >}}
## Synopsis

Get a specific workflow for an organization

```shell
kosli get workflow ID [flags]
```

## Flags
| Flag | Description |
| :--- | :--- |
|        --audit-trail string  |  The Kosli audit trail name.  |
|    -h, --help  |  help for workflow  |
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


## Examples

```shell

# get workflow for an ID
kosli get workflow yourID \
	--audit-trail auditTrailName \
	--api-token yourAPIToken \
	--org orgName

```

