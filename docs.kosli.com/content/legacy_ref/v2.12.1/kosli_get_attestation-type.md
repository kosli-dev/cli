---
title: "kosli get attestation-type"
beta: false
deprecated: false
summary: "Get a custom Kosli attestation type.  "
---

# kosli get attestation-type

## Synopsis

```shell
kosli get attestation-type TYPE-NAME [flags]
```

Get a custom Kosli attestation type.  
The TYPE-NAME can be specified as follows:
- customTypeName
	- Returns the unversioned custom attestation type, containing details of all versions of the type.
	- e.g. `custom-type`
- customTypeName@vN
	- Returns the Nth version of the custom attestation type.
	- If a non-integer version number is given, the unversioned custom attestation type is returned.
	- e.g. `custom-type@v4`


## Flags
| Flag | Description |
| :--- | :--- |
|    -h, --help  |  help for attestation-type  |
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

These examples all assume that the flags  `--api-token`, `--org`, `--host`, (and `--flow`, `--trail` when required), are [set/provided](https://docs.kosli.com/getting_started/install/#assigning-flags-via-environment-variables). 

##### get an unversioned custom attestation type

```shell
kosli get attestation-type customTypeName

```

##### get version 1 of a custom attestation type

```shell
kosli get attestation-type customTypeName@v1
```

