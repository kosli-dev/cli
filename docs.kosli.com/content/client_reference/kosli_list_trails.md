---
title: "kosli list trails"
beta: false
deprecated: false
summary: "List Trails for a Flow in an org."
---

# kosli list trails

## Synopsis

```shell
kosli list trails [flags]
```

List Trails for a Flow in an org.The results are ordered from latest to oldest.  
If the `page-limit` flag is provided, the results will be paginated, otherwise all results will be 
returned.  
If `page-limit` is set to 0, all results will be returned.

## Flags
| Flag | Description |
| :--- | :--- |
|    -f, --flow string  |  The Kosli flow name.  |
|    -h, --help  |  help for trails  |
|    -o, --output string  |  [defaulted] The format of the output. Valid formats are: [table, json]. (default "table")  |
|        --page int  |  [defaulted] The page number of a response. (default 1)  |
|    -n, --page-limit int  |  [defaulted] The number of elements per page.  |


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

**list all trails for a flow**

```shell
kosli list trails 

```

**list the most recent 30 trails for a flow**

```shell
kosli list trails 
	--page-limit 30 

```

**show the second page of trails for a flow**

```shell
kosli list trails 
	--page-limit 30 
	--page 2 

```

**list all trails for a flow (in JSON)**

```shell
kosli list trails 
	--output json
```

