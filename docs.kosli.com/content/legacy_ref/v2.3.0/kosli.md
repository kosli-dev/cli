---
title: "kosli"
experimental: false
---

# kosli

## Synopsis

The Kosli evidence reporting CLI.

Environment variables:
You can set any flag from an environment variable by capitalizing it in snake case and adding the KOSLI_ prefix.
For example, to set --api-token from an environment variable, you can export KOSLI_API_TOKEN=YOUR_API_TOKEN.

Setting the API token to DRY_RUN sets the --dry-run flag.


## Flags
| Flag | Description |
| :--- | :--- |
|    -a, --api-token string  |  The Kosli API token.  |
|    -c, --config-file string  |  [optional] The Kosli config file path. (default "kosli")  |
|        --debug  |  [optional] Print debug logs to stdout.  |
|    -h, --help  |  help for kosli  |
|    -H, --host string  |  [defaulted] The Kosli endpoint. (default "https://app.kosli.com")  |
|    -r, --max-api-retries int  |  [defaulted] How many times should API calls be retried when the API host is not reachable. (default 3)  |
|        --org string  |  The Kosli organization.  |


