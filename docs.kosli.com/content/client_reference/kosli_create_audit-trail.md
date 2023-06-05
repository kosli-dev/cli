---
title: "kosli create audit-trail"
beta: true
---

# kosli create audit-trail

{{< hint warning >}}**kosli create audit-trail** is an beta feature. 
Beta features provide early access to product functionality. These features may change between releases without warning, or can be removed from a future release.
You can enable beta features by using the `kosli enable beta` command.{{< /hint >}}
## Synopsis

Create or update a Kosli audit trail.
You can specify audit trail parameters in flags.

```shell
kosli create audit-trail AUDIT-TRAIL-NAME [flags]
```

## Flags
| Flag | Description |
| :--- | :--- |
|        --description string  |  [optional] The Kosli flow description.  |
|    -D, --dry-run  |  [optional] Run in dry-run mode. When enabled, no data is sent to Kosli and the CLI exits with 0 exit code regardless of any errors.  |
|    -h, --help  |  help for audit-trail  |
|    -s, --steps strings  |  [defaulted] The comma-separated list of required audit trail steps names.  |


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

# create/update a Kosli audit trail:
kosli create audit-trail yourAuditTrailName \
	--description yourAuditTrailDescription \
	--steps step1,step2 \
	--api-token yourAPIToken \
	--org yourOrgName

```

