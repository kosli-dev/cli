---
title: "kosli get attestation"
beta: false
deprecated: false
summary: "Get an attestation using its name or id.  "
---

# kosli get attestation

## Synopsis

```shell
kosli get attestation [ATTESTATION-NAME] [flags]
```

Get an attestation using its name or id.  

You can get an attestation from a trail or artifact using its name. The attestation name should be given
WITHOUT dot-notation.  
To get an attestation from a trail, specify the trail name using the `--trail` flag.  
To get an attestation from an artifact, specify the artifact fingerprint using the `--fingerprint` flag.  
These flags cannot be used together. In both cases the flow must also be specified using the `--flow` flag.  
If there are multiple attestations with the same name on the trail or artifact, a list of all will be returned.

You can also get an attestation by its id using the `--attestation-id` flag. This cannot be used with the attestation name,
or any of the `--flow`, `--trail` or `--fingerprint` flags.


## Flags
| Flag | Description |
| :--- | :--- |
|        --attestation-id string  |  [conditional] The unique identifier of the attestation to retrieve. Cannot be used together with ATTESTATION-NAME.  |
|    -F, --fingerprint string  |  [conditional] The fingerprint of the artifact for the attestation. Cannot be used together with --trail or --attestation-id.  |
|    -f, --flow string  |  [conditional] The name of the Kosli flow for the attestation. Required if ATTESTATION-NAME provided. Cannot be used together with --attestation-id.  |
|    -h, --help  |  help for attestation  |
|    -o, --output string  |  [defaulted] The format of the output. Valid formats are: [table, json]. (default "table")  |
|    -t, --trail string  |  [conditional] The name of the Kosli trail for the attestation. Cannot be used together with --fingerprint or --attestation-id.  |


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


## Live Example

{{< raw-html >}}To view a live example of 'kosli get attestation' you can run the commands below (for the <a href="https://app.kosli.com/cyber-dojo/environments/aws-prod/snapshots/">cyber-dojo</a> demo organization).<br/><a href="https://app.kosli.com/api/v2/livedocs/cyber-dojo/cli?command=kosli+get+attestation+snyk-container-scan+--flow=differ-ci+--fingerprint=0cbbe3a6e73e733e8ca4b8813738d68e824badad0508ff20842832b5143b48c0+--output=json">Run the commands below and view the output.</a><pre>export KOSLI_ORG=cyber-dojo
export KOSLI_API_TOKEN=Pj_XT2deaVA6V1qrTlthuaWsmjVt4eaHQwqnwqjRO3A  # read-only
kosli get attestation snyk-container-scan --flow=differ-ci --fingerprint=0cbbe3a6e73e733e8ca4b8813738d68e824badad0508ff20842832b5143b48c0 --output=json</pre>{{< / raw-html >}}

## Examples Use Cases

These examples all assume that the flags  `--api-token`, `--org`, `--host`, (and `--flow`, `--trail` when required), are [set/provided](https://docs.kosli.com/getting_started/install/#assigning-flags-via-environment-variables). 

##### get an attestation by name from a trail (requires the --trail flag)

```shell
kosli get attestation attestationName 

```

##### get an attestation by name from an artifact

```shell
kosli get attestation attestationName 
	--fingerprint fingerprint

```

##### get an attestation by its id

```shell
kosli get attestation --attestation-id attestationID
```

