---
title: "kosli report workflow"
beta: true
---

# kosli report workflow

{{< hint warning >}}**kosli report workflow** is a beta feature. 
Beta features provide early access to product functionality. These features may change between releases without warning, or can be removed from a future release.
You can enable beta features by using the `kosli enable beta` command.{{< /hint >}}
## Synopsis

Report a workflow creation to a Kosli audit-trail.

```shell
kosli report workflow [flags]
```

## Flags
| Flag | Description |
| :--- | :--- |
|        --audit-trail string  |  The Kosli audit trail name.  |
|    -D, --dry-run  |  [optional] Run in dry-run mode. When enabled, no data is sent to Kosli and the CLI exits with 0 exit code regardless of any errors.  |
|    -h, --help  |  help for workflow  |
|        --id string  |  The ID of the workflow.  |


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

# Report to a Kosli audit-trail that a workflow has been created
kosli report workflow \
	--audit-trail auditTrailName \
	--api-token yourAPIToken \
	--id yourID \
	--org yourOrgName

```

