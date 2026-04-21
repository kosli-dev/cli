---
title: "kosli evaluate input"
beta: false
deprecated: false
summary: "Evaluate a local JSON input against a Rego policy."
---

# kosli evaluate input

## Synopsis

```shell
kosli evaluate input [flags]
```

Evaluate a local JSON input against a Rego policy.
Read JSON from a file or stdin and evaluate it against a Rego policy.
The input file should contain the raw JSON object your policy expects —
not the wrapper produced by `--show-input`. Use `jq '.input'` to extract
the policy input from a `--show-input --output json` capture.

The policy must use `package policy` and define an `allow` rule.
An optional `violations` rule (a set of strings) can provide human-readable denial reasons.
The command exits with code 0 when allowed and code 1 when denied.

When `--input-file` is omitted, JSON is read from stdin.

Use `--params` to pass configuration data to the policy as `data.params`.
This accepts inline JSON or a file reference (`@file.json`).

## Flags
| Flag | Description |
| :--- | :--- |
|    -h, --help  |  help for input  |
|    -i, --input-file string  |  [optional] Path to a JSON input file. Reads from stdin if omitted.  |
|    -o, --output string  |  [defaulted] The format of the output. Valid formats are: [table, json]. (default "table")  |
|        --params string  |  [optional] Policy parameters as inline JSON or @file.json. Available in policies as data.params.  |
|    -p, --policy string  |  Path to a Rego policy file to evaluate against the input.  |
|        --show-input  |  [optional] Include the policy input data in the output.  |


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

##### capture trail data for local policy iteration

```shell
kosli evaluate trail TRAIL --flow FLOW 
	--policy allow-all.rego 
	--show-input --output json | jq '.input' > trail-data.json

```

##### then iterate on your policy locally

```shell
kosli evaluate input 
	--input-file trail-data.json 
	--policy policy.rego

```

##### evaluate and show the data passed to the policy

```shell
kosli evaluate input 
	--input-file trail-data.json 
	--policy policy.rego 
	--show-input 
	--output json

```

##### read input from stdin

```shell
cat trail-data.json | kosli evaluate input 
	--policy policy.rego

```

##### evaluate with policy parameters (inline JSON)

```shell
kosli evaluate input 
	--input-file trail-data.json 
	--policy policy.rego 
	--params '{"threshold": 3}'

```

##### evaluate with policy parameters from a file

```shell
kosli evaluate input 
	--input-file trail-data.json 
	--policy policy.rego 
	--params @params.json
```

