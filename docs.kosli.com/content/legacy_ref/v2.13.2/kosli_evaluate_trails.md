---
title: "kosli evaluate trails"
beta: false
deprecated: false
summary: "Evaluate multiple trails against a policy."
---

# kosli evaluate trails

## Synopsis

```shell
kosli evaluate trails TRAIL-NAME [TRAIL-NAME...] [flags]
```

Evaluate multiple trails against a policy.
Fetch multiple trails from Kosli and evaluate them together against a Rego policy using OPA.
The trail data is passed to the policy as `input.trails` (an array), unlike
`evaluate trail` which passes `input.trail` (a single object).

Use `--attestations` to enrich the input with detailed attestation data
(e.g. pull request approvers, scan results). Use `--show-input` to inspect the
full data structure available to the policy. Use `--output json` for structured output.

## Flags
| Flag | Description |
| :--- | :--- |
|        --attestations strings  |  [optional] Limit which attestations are included. Plain name for trail-level, dot-qualified (artifact.name) for artifact-level.  |
|    -f, --flow string  |  The Kosli flow name.  |
|    -h, --help  |  help for trails  |
|    -o, --output string  |  [defaulted] The format of the output. Valid formats are: [table, json]. (default "table")  |
|    -p, --policy string  |  Path to a Rego policy file to evaluate against the trails.  |
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

##### evaluate multiple trails against a policy

```shell
kosli evaluate trails yourTrailName1 yourTrailName2 
	--policy yourPolicyFile.rego 

```

##### evaluate trails with attestation enrichment

```shell
kosli evaluate trails yourTrailName1 yourTrailName2 
	--policy yourPolicyFile.rego 
	--attestations pull-request 

```

##### evaluate trails with JSON output and show the policy input

```shell
kosli evaluate trails yourTrailName1 yourTrailName2 
	--policy yourPolicyFile.rego 
	--show-input 
	--output json 
```

