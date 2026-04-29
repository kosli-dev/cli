---
title: "kosli evaluate trail"
beta: false
deprecated: false
summary: "[BETA] Evaluate a trail against a policy."
---

# kosli evaluate trail

## Synopsis

```shell
kosli evaluate trail TRAIL-NAME [flags]
```

[BETA] Evaluate a trail against a policy.
Fetch a single trail from Kosli and evaluate it against a Rego policy.
The trail data is passed to the policy as `input.trail`.

Use `--attestations` to enrich the input with detailed attestation data
(e.g. pull request approvers, scan results). Use `--show-input` to inspect the
full data structure available to the policy. Use `--output json` for structured output.

## Flags
| Flag | Description |
| :--- | :--- |
|        --assert  |  [optional] Exit with a non-zero status when the policy denies. This is the current default; pass --assert to lock it in across future releases.  |
|        --attestations strings  |  [optional] Limit which attestations are included. Plain name for trail-level, dot-qualified (artifact.name) for artifact-level.  |
|    -f, --flow string  |  The Kosli flow name.  |
|    -h, --help  |  help for trail  |
|        --no-assert  |  [optional] Print the result and always exit 0, even when the policy denies. Use when this command feeds another tool as a policy decision point.  |
|    -o, --output string  |  [defaulted] The format of the output. Valid formats are: [table, json]. (default "table")  |
|        --params string  |  [optional] Policy parameters as inline JSON or @file.json. Available in policies as data.params.  |
|    -p, --policy string  |  Path to a Rego policy file to evaluate against the trail.  |
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


## Live Examples in different CI systems

{{< tabs "live-examples" "col-no-wrap" >}}{{< tab "GitHub" >}}View an example of the `kosli evaluate trail` command in GitHub.

In [this YAML file](https://app.kosli.com/api/v2/livedocs/cyber-dojo/yaml?ci=github&command=kosli%2Bevaluate%2Btrail){{< /tab >}}{{< /tabs >}}

## Examples Use Cases

These examples all assume that the flags  `--api-token`, `--org`, `--host`, (and `--flow`, `--trail` when required), are [set/provided](https://docs.kosli.com/getting_started/install/#assigning-flags-via-environment-variables). 

##### evaluate a trail against a policy

```shell
kosli evaluate trail yourTrailName 
	--policy yourPolicyFile.rego 

```

##### evaluate a trail with attestation enrichment

```shell
kosli evaluate trail yourTrailName 
	--policy yourPolicyFile.rego 
	--attestations pull-request 

```

##### evaluate a trail and show the policy input data

```shell
kosli evaluate trail yourTrailName 
	--policy yourPolicyFile.rego 
	--show-input 
	--output json 

```

##### evaluate a trail with policy parameters (inline JSON)

```shell
kosli evaluate trail yourTrailName 
	--policy yourPolicyFile.rego 
	--params '{"min_approvers": 2}' 

```

##### evaluate a trail with policy parameters from a file

```shell
kosli evaluate trail yourTrailName 
	--policy yourPolicyFile.rego 
	--params @params.json 

```

##### evaluate a trail as a decision point (print verdict, never fail the step)

```shell
kosli evaluate trail yourTrailName 
	--policy yourPolicyFile.rego 
	--no-assert 
```

