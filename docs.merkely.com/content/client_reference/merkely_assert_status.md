---
title: "merkely assert status"
---

## merkely assert status


Assert the status of Merkely server. Exits with non-zero code if Merkely server down.


### Synopsis


Assert the status of Merkely server. Exits with non-zero code if Merkely server down.


```shell
merkely assert status [flags]
```

### Flags
| Flag | Description |
| :--- | :--- |
|    -h, --help  |  help for status  |


### Options inherited from parent commands
| Flag | Description |
| :--- | :--- |
|    -a, --api-token string  |  The merkely API token.  |
|    -c, --config-file string  |  [optional] The merkely config file path. (default "merkely")  |
|    -D, --dry-run  |  Whether to run in dry-run mode. When enabled, data is not sent to Merkely and the CLI exits with 0 exit code regardless of errors.  |
|    -H, --host string  |  The merkely endpoint. (default "https://app.merkely.com")  |
|    -r, --max-api-retries int  |  How many times should API calls be retried when the API host is not reachable. (default 3)  |
|    -o, --owner string  |  The merkely user or organization.  |
|    -v, --verbose  |  Print verbose logs to stdout.  |


