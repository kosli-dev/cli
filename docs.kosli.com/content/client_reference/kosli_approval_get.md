---
title: "kosli approval get"
---

## kosli approval get

Get an approval from a specified pipeline

### Synopsis

Get an approval from a specified pipeline

```shell
kosli approval get APPROVAL-ID [flags]
```

### Flags
| Flag | Description |
| :--- | :--- |
|    -h, --help  |  help for get  |
|    -o, --output string  |  [defaulted] The format of the output. Valid formats are: [table, json]. (default "table")  |
|    -p, --pipeline string  |  The Kosli pipeline name.  |


### Options inherited from parent commands
| Flag | Description |
| :--- | :--- |
|    -a, --api-token string  |  The Kosli API token.  |
|    -c, --config-file string  |  [optional] The Kosli config file path. (default "kosli")  |
|    -D, --dry-run  |  [optional] Whether to run in dry-run mode. When enabled, data is not sent to Kosli and the CLI exits with 0 exit code regardless of errors.  |
|    -H, --host string  |  [defaulted] The Kosli endpoint. (default "https://app.kosli.com")  |
|    -r, --max-api-retries int  |  [defaulted] How many times should API calls be retried when the API host is not reachable. (default 3)  |
|        --owner string  |  The Kosli user or organization.  |
|    -v, --verbose  |  [optional] Print verbose logs to stdout.  |


