---
title: "kosli create attestation-type"
beta: false
deprecated: false
summary: "Create or update a Kosli custom attestation type."
---

# kosli create attestation-type

## Synopsis

Create or update a Kosli custom attestation type.
You can specify attestation type parameters in flags.

`TYPE-NAME` must start with a letter or number, and only contain letters, numbers, `.`, `-`, `_`, and `~`.

`--schema` is a path to a file containing a JSON schema which will be used to validate attestations made using this type.  
The schema is used to specify the structure of the attestation data, e.g. any fields that are required or 
the expected type of the data.
See an example schema file 
[here](https://github.com/cyber-dojo/kosli-attestation-types/blob/f9130c58d3a8151b0b0e7c5db284e4380eb2d2cf/metrics-coverage.schema.json).

`--jq` defines an evaluation rule, given in jq-format, for this attestation type. The flag can be repeated in order to add additional rules.  
These rules specify acceptable values for attestation data, e.g. `.age >= 21` or `.failing_tests == 0`.  
When a custom attestation is reported, the provided data is evaluated according to the rules defined in its attestation-type. 
All rules must return `true` for the evaluation to pass and the attestation to be determined compliant.


```shell
kosli create attestation-type TYPE-NAME [flags]
```

## Flags
| Flag | Description |
| :--- | :--- |
|    -d, --description string  |  [optional] The attestation type description.  |
|    -D, --dry-run  |  [optional] Run in dry-run mode. When enabled, no data is sent to Kosli and the CLI exits with 0 exit code regardless of any errors.  |
|    -h, --help  |  help for attestation-type  |
|        --jq stringArray  |  [optional] The attestation type evaluation JQ rules.  |
|    -s, --schema string  |  [optional] Path to the attestation type schema in JSON Schema format.  |


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

{{< tabs "live-examples" "col-no-wrap" >}}{{< tab "GitHub" >}}View an example of the `kosli create attestation-type` command in GitHub.

In [this YAML file](https://app.kosli.com/api/v2/livedocs/cyber-dojo/yaml?ci=github&command=kosli+create+attestation-type){{< /tab >}}{{< /tabs >}}

## Examples Use Cases

These examples all assume that the flags  `--api-token`, `--org`, `--host`, (and `--flow`, `--trail` when required), are set/provided. 

**create/update a custom attestation type with no schema no evaluation rules**

```shell
kosli create attestation-type customTypeName

```

**create/update a custom attestation type with schema and jq evaluation rules**

```shell
kosli create attestation-type customTypeName 
    --description "Attest that a person meets the age requirements." 
    --schema person-schema.json 
    --jq ".age >= 18"
    --jq ".age < 65"
```

