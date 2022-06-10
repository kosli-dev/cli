---
title: "kosli"
---

## kosli

The Kosli CLI.

### Synopsis

The Kosli evidence reporting CLI.

Environment variables:

| Name                               | Description                                                                       |
|------------------------------------|-----------------------------------------------------------------------------------|
| $MERKELY_API_TOKEN                 | set the Kosli API token.                                                        |
| $MERKELY_OWNER                     | set the Kosli Pipeline Owner.                                                   |
| $MERKELY_HOST                      | set the Kosli host.                                                             |
| $MERKELY_DRY_RUN                   | indicate whether or not Kosli CLI is running in Dry Run mode.                   |
| $MERKELY_MAX_API_RETRIES           | set the maximum number of API calling retries when the API host is not reachable. |
| $MERKELY_CONFIG_FILE               | set the path to Kosli config file where you can set your options.               |         


### Flags
| Flag | Description |
| :--- | :--- |
|    -a, --api-token string  |  The Kosli API token.  |
|    -c, --config-file string  |  [optional] The Kosli config file path. (default "merkely")  |
|    -D, --dry-run  |  Whether to run in dry-run mode. When enabled, data is not sent to Kosli and the CLI exits with 0 exit code regardless of errors.  |
|    -h, --help  |  help for kosli  |
|    -H, --host string  |  The Kosli endpoint. (default "https://app.merkely.com")  |
|    -r, --max-api-retries int  |  How many times should API calls be retried when the API host is not reachable. (default 3)  |
|    -o, --owner string  |  The Kosli user or organization.  |
|    -v, --verbose  |  Print verbose logs to stdout.  |


