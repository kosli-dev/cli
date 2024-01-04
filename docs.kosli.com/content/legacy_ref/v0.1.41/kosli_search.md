---
title: "kosli search"
---

# kosli search

## Synopsis

Search for a git commit or an artifact fingerprint in Kosli. 
You can use short git commit or artifact fingerprint shas, but you must provide at least 5 characters.

```shell
kosli search GIT-COMMIT|FINGERPRINT [flags]
```

## Flags
| Flag | Description |
| :--- | :--- |
|    -h, --help  |  help for search  |
|    -o, --output string  |  [defaulted] The format of the output. Valid formats are: [table, json]. (default "table")  |


## Options inherited from parent commands
| Flag | Description |
| :--- | :--- |
|    -a, --api-token string  |  The Kosli API token.  |
|    -c, --config-file string  |  [optional] The Kosli config file path. (default "kosli")  |
|        --debug  |  [optional] Print debug logs to stdout.  |
|    -H, --host string  |  [defaulted] The Kosli endpoint. (default "https://app.kosli.com")  |
|    -r, --max-api-retries int  |  [defaulted] How many times should API calls be retried when the API host is not reachable. (default 3)  |
|        --owner string  |  The Kosli user or organization.  |


## Examples

```shell

# Search for a git commit in Kosli
kosli search YOUR_GIT_COMMIT \
	--api-token yourApiToken \
	--owner yourOrgName

# Search for an artifact fingerprint in Kosli
kosli search YOUR_FINGERPRINT \
	--api-token yourApiToken \
	--owner yourOrgName

```

