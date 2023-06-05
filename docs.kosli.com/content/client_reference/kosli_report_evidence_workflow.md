---
title: "kosli report evidence workflow"
beta: true
---

# kosli report evidence workflow

{{< hint warning >}}**kosli report evidence workflow** is an beta feature. 
Beta features provide early access to product functionality. These features may change between releases without warning, or can be removed from a future release.
You can enable beta features by using the `kosli enable beta` command.{{< /hint >}}
## Synopsis

Report evidence for a workflow in Kosli.

```shell
kosli report evidence workflow [flags]
```

## Flags
| Flag | Description |
| :--- | :--- |
|        --audit-trail string  |  The Kosli audit trail name.  |
|    -D, --dry-run  |  [optional] Run in dry-run mode. When enabled, no data is sent to Kosli and the CLI exits with 0 exit code regardless of any errors.  |
|        --evidence-fingerprint string  |  [optional] The SHA256 fingerprint of the evidence file or dir.  |
|    -e, --evidence-paths strings  |  [optional] The comma-separated list of paths containing supporting proof for the reported evidence. Paths can be for files or directories.  |
|        --evidence-url string  |  [optional] The external URL where the evidence file or dir is stored.  |
|    -h, --help  |  help for workflow  |
|        --id string  |  The ID of the workflow.  |
|        --step string  |  The name of the step as defined in the audit trail's steps.  |
|    -u, --user-data string  |  [optional] The path to a JSON file containing additional data you would like to attach to this evidence.  |


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

# report evidence for a workflow:
kosli report evidence workflow \
	--audit-trail auditTrailName \
	--api-token yourAPIToken \
	--id yourID \
	--step step1 \
	--org yourOrgName

# report evidence with a file for a workflow:
kosli report evidence workflow \
	--audit-trail auditTrailName \
	--api-token yourAPIToken \
	--id yourID \
	--step step1 \
	--org yourOrgName \
	--evidence-paths /path/to/your/file

```

