---
title: "kosli approval ls"
---

## kosli approval ls

List a number of approvals in a pipeline.

### Synopsis

List a number of approvals in a pipeline.

```shell
kosli approval ls PIPELINE-NAME [flags]
```

### Flags
| Flag | Description |
| :--- | :--- |
|    -h, --help  |  help for ls  |
|    -o, --output string  |  [defaulted] The format of the output. Valid formats are: [table, json]. (default "table")  |
|        --page int  |  [defaulted] The page number of a response. (default 1)  |
|    -n, --page-limit int  |  [defaulted] The number of elements per page. (default 15)  |


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


