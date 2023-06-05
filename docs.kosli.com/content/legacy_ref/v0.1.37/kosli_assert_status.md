---
title: "kosli assert status"
---

## kosli assert status

Assert the status of Kosli server.

### Synopsis

Assert the status of Kosli server.
Exits with non-zero code if Kosli server down.

```shell
kosli assert status [flags]
```

### Flags
| Flag | Description |
| :--- | :--- |
|    -h, --help  |  help for status  |


### Options inherited from parent commands
| Flag | Description |
| :--- | :--- |
|    -a, --api-token string  |  The Kosli API token.  |
|    -c, --config-file string  |  [optional] The Kosli config file path. (default "kosli")  |
|        --debug  |  [optional] Print debug logs to stdout.  |
|    -H, --host string  |  [defaulted] The Kosli endpoint. (default "https://app.kosli.com")  |
|    -r, --max-api-retries int  |  [defaulted] How many times should API calls be retried when the API host is not reachable. (default 3)  |
|        --owner string  |  The Kosli user or organization.  |


