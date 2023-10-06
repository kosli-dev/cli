---
title: "kosli assert status"
beta: false
---

# kosli assert status

## Synopsis

Assert the status of a Kosli server.
Exits with non-zero code if the Kosli server down.

```shell
kosli assert status [flags]
```

## Flags
| Flag | Description |
| :--- | :--- |
|    -h, --help  |  help for status  |


## Options inherited from parent commands
| Flag | Description |
| :--- | :--- |
|    -a, --api-token string  |  The Kosli API token.  |
|    -c, --config-file string  |  [optional] The Kosli config file path. (default "kosli")  |
|        --debug  |  [optional] Print debug logs to stdout. A boolean flag https://docs.kosli.com/faq/#boolean-flags (default false)  |
|    -H, --host string  |  [defaulted] The Kosli endpoint. (default "https://app.kosli.com")  |
|    -r, --max-api-retries int  |  [defaulted] How many times should API calls be retried when the API host is not reachable. (default 3)  |
|        --org string  |  The Kosli organization.  |


