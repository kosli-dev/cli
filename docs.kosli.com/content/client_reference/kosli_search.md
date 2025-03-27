---
title: "kosli search"
beta: false
deprecated: false
summary: "Search for a git commit or an artifact fingerprint in Kosli.  "
---

# kosli search

## Synopsis

Search for a git commit or an artifact fingerprint in Kosli.   
You can use short git commit or artifact fingerprint shas, but you must provide at least 5 characters.

```shell
kosli search {GIT-COMMIT | FINGERPRINT} [flags]
```

## Flags
| Flag | Description |
| :--- | :--- |
|    -h, --help  |  help for search  |
|    -o, --output string  |  [defaulted] The format of the output. Valid formats are: [table, json]. (default "table")  |


## Flags inherited from parent commands
| Flag | Description |
| :--- | :--- |
|    -a, --api-token string  |  The Kosli API token.  |
|    -c, --config-file string  |  [optional] The Kosli config file path. (default "kosli")  |
|        --debug  |  [optional] Print debug logs to stdout. A boolean flag https://docs.kosli.com/faq/#boolean-flags (default false)  |
|    -H, --host string  |  [defaulted] The Kosli endpoint. (default "https://app.kosli.com")  |
|        --http-proxy string  |  [optional] The HTTP proxy URL including protocol and port number. e.g. 'http://proxy-server-ip:proxy-port'  |
|    -r, --max-api-retries int  |  [defaulted] How many times should API calls be retried when the API host is not reachable. (default 3)  |
|        --org string  |  The Kosli organization.  |


## Examples Use Cases

These examples all assume that the flags  `--api-token`, `--org`, `--host`, (and `--flow`, `--trail` when required), are set/provided. 

**Search for a git commit in Kosli**

```shell
kosli search YOUR_GIT_COMMIT 

```

**Search for an artifact fingerprint in Kosli**

```shell
kosli search YOUR_ARTIFACT_FINGERPRINT 
```

