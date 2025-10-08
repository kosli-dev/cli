---
title: "kosli config"
beta: false
deprecated: false
summary: "Config global Kosli flags values and store them in $HOME/.kosli .  "
---

# kosli config

## Synopsis

Config global Kosli flags values and store them in $HOME/.kosli .  

Flag values are determined in the following order (highest precedence first):
- command line flags on each executed command.
- environment variables.
- custom config file provided with --config-file flag.
- default config file in $HOME/.kosli

You can configure global Kosli flags (the ones that apply to all/most commands) using their dedicated
convenience flags (e.g. --org). 

API tokens are stored in the suitable credentials manager on your machine. 

Other Kosli flags can be configured using the --set flag which takes a comma-separated list of key=value pairs.
Keys correspond to the specific flag name, capitalized. For instance: --flow would be set using --set FLOW=value


```shell
kosli config [flags]
```

## Flags
| Flag | Description |
| :--- | :--- |
|    -h, --help  |  help for config  |
|        --set stringToString  |  [optional] The key-value pairs to tag the resource with. The format is: key=value  |
|        --unset strings  |  [optional] The list of tag keys to remove from the resource.  |


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

**configure global flags in your default config file**

```shell
kosli config --org=yourOrg 
	--api-token=yourAPIToken 
	--host=https://app.kosli.com 
	--debug=false 
	--max-api-retries=3 
	--http-proxy=http://192.0.0.1:8080

```

**configure non-global flags in your default config file**

```shell
kosli config --set FLOW=yourFlowName

```

**remove a key from the default config file**

```shell
kosli config --unset FLOW
```

