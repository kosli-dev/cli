---
title: "kosli begin trail"
beta: true
deprecated: false
---

# kosli begin trail

{{< hint warning >}}**kosli begin trail** is a beta feature. Beta features provide early access to product functionality.  These features may change between releases without warning, or can be removed in a future release.
Please contact us to enable this feature for your organization.{{< /hint >}}
## Synopsis

Begin or update a Kosli flow trail.

```shell
kosli begin trail TRAIL-NAME [flags]
```

## Flags
| Flag | Description |
| :--- | :--- |
|    -g, --commit string  |  [defaulted] The git commit from which the trail is begun. (defaulted in some CIs: https://docs.kosli.com/ci-defaults, otherwise defaults to HEAD ).  |
|        --description string  |  [optional] The Kosli trail description.  |
|    -D, --dry-run  |  [optional] Run in dry-run mode. When enabled, no data is sent to Kosli and the CLI exits with 0 exit code regardless of any errors.  |
|        --external-fingerprint stringToString  |  [optional] A SHA256 fingerprint of an external attachment represented by --external-url. The format is label=fingerprint. This flag can be set multiple times. There must be an external url with a matching label for each external fingerprint.  |
|        --external-url stringToString  |  [optional] Add labeled reference URL for an external resource. The format is label=url. This flag can be set multiple times. If the resource is a file or dir, you can optionally add its fingerprint via --external-fingerprint  |
|        --flow string  |  The Kosli flow name.  |
|    -h, --help  |  help for trail  |
|    -o, --origin-url string  |  [optional] The url pointing to where the attestation came from or is related. (defaulted to the CI url in some CIs: https://docs.kosli.com/ci-defaults ).  |
|        --repo-root string  |  [defaulted] The directory where the source git repository is available. Only used if --commit is used. (default ".")  |
|    -f, --template-file string  |  The path to a yaml template file.  |
|    -u, --user-data string  |  [optional] The path to a JSON file containing additional data you would like to attach to the flow trail.  |


## Options inherited from parent commands
| Flag | Description |
| :--- | :--- |
|    -a, --api-token string  |  The Kosli API token.  |
|    -c, --config-file string  |  [optional] The Kosli config file path. (default "kosli")  |
|        --debug  |  [optional] Print debug logs to stdout. A boolean flag https://docs.kosli.com/faq/#boolean-flags (default false)  |
|    -H, --host string  |  [defaulted] The Kosli endpoint. (default "https://app.kosli.com")  |
|    -r, --max-api-retries int  |  [defaulted] How many times should API calls be retried when the API host is not reachable. (default 3)  |
|        --org string  |  The Kosli organization.  |


## Examples

```shell

# begin/update a Kosli flow trail:
kosli begin trail yourTrailName \
	--description yourTrailDescription \
	--template-file /path/to/your/template/file.yml \
	--user-data /path/to/your/user-data/file.json \
	--api-token yourAPIToken \
	--org yourOrgName

```
