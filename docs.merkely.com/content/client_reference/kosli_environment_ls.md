---
title: "kosli environment ls"
---

## kosli environment ls

List environments.

### Synopsis

List environments.

```shell
kosli environment ls [ENVIRONMENT-NAME] [flags]
```

### Flags
| Flag | Description |
| :--- | :--- |
|    -h, --help  |  help for ls  |
|    -j, --json  |  Print environment info as json.  |
|    -l, --long  |  Print long environment info.  |


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


