---
title: "kosli list artifacts"
beta: false
deprecated: false
summary: "List artifacts in a flow. "
---

# kosli list artifacts

## Synopsis

```shell
kosli list artifacts [flags]
```

List artifacts in a flow. The results are paginated and ordered from latest to oldest.
By default, the page limit is 15 artifacts per page.


## Flags
| Flag | Description |
| :--- | :--- |
|    -f, --flow string  |  The Kosli flow name.  |
|    -h, --help  |  help for artifacts  |
|    -o, --output string  |  [defaulted] The format of the output. Valid formats are: [table, json]. (default "table")  |
|        --page int  |  [defaulted] The page number of a response. (default 1)  |
|    -n, --page-limit int  |  [defaulted] The number of elements per page. (default 15)  |


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

These examples all assume that the flags  `--api-token`, `--org`, `--host`, (and `--flow`, `--trail` when required), are [set/provided](https://docs.kosli.com/getting_started/install/#assigning-flags-via-environment-variables). 

##### list the last 15 artifacts for a flow

```shell
kosli list artifacts 

```

##### list the last 30 artifacts for a flow

```shell
kosli list artifacts 
	--page-limit 30 

```

##### list the last 30 artifacts for a flow (in JSON)

```shell
kosli list artifacts 
	--page-limit 30 
	--output json
```

