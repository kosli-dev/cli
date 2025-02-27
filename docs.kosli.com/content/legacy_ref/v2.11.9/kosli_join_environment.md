---
title: "kosli join environment"
beta: false
deprecated: false
summary: "Join a physical environment to a logical environment."
---

# kosli join environment

## Synopsis

Join a physical environment to a logical environment.

```shell
kosli join environment [flags]
```

## Flags
| Flag | Description |
| :--- | :--- |
|    -D, --dry-run  |  [optional] Run in dry-run mode. When enabled, no data is sent to Kosli and the CLI exits with 0 exit code regardless of any errors.  |
|    -h, --help  |  help for environment  |
|        --logical string  |  [required] The logical environment.  |
|        --physical string  |  [required] The physical environment.  |


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

**join a physical environment to a logical environment**

```shell
kosli join environment \
	--physical prod-k8 \
	--logical prod \
	--api-token yourAPIToken \
	--org yourOrgName
```

