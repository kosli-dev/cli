---
title: "kosli assert status"
---

## kosli assert status


Assert the status of Kosli server. Exits with non-zero code if Kosli server down.


### Synopsis


Assert the status of Kosli server. Exits with non-zero code if Kosli server down.


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
|    -c, --config-file string  |  [optional] The Kosli config file path. (default "merkely")  |
|    -D, --dry-run  |  Whether to run in dry-run mode. When enabled, data is not sent to Kosli and the CLI exits with 0 exit code regardless of errors.  |
|    -H, --host string  |  The Kosli endpoint. (default "https://app.merkely.com")  |
|    -r, --max-api-retries int  |  How many times should API calls be retried when the API host is not reachable. (default 3)  |
|    -o, --owner string  |  The Kosli user or organization.  |
|    -v, --verbose  |  Print verbose logs to stdout.  |


