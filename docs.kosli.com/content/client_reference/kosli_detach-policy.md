---
title: "kosli detach-policy"
beta: false
deprecated: false
summary: "Detach a policy from one or more Kosli environments.  "
---

# kosli detach-policy

## Synopsis

If the environment has no more policies attached to it, then its snapshots' status will become "unknown".

```shell
kosli detach-policy POLICY-NAME [flags]
```

## Flags
| Flag | Description |
| :--- | :--- |
|    -D, --dry-run  |  [optional] Run in dry-run mode. When enabled, no data is sent to Kosli and the CLI exits with 0 exit code regardless of any errors.  |
|    -e, --environment strings  |  the list of environment names to detach the policy from  |
|    -h, --help  |  help for detach-policy  |


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


## Examples Use Cases

**detach policy from multiple environment**

```shell
kosli detach-policy yourPolicyName \
	--environment yourFirstEnvironmentName \
	--environment yourSecondEnvironmentName \
	--api-token yourAPIToken \
	--org yourOrgName
```

