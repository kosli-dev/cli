---
title: "kosli assert status"
beta: false
deprecated: false
summary: "Assert the status of a Kosli server."
---

# kosli assert status

## Synopsis

```shell
kosli assert status [flags]
```

Assert the status of a Kosli server.
Exits with non-zero code if the Kosli server down.

## Flags
| Flag | Description |
| :--- | :--- |
|    -h, --help  |  help for status  |


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


