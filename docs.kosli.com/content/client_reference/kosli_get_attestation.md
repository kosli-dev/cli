---
title: "kosli get attestation"
beta: false
deprecated: false
summary: "Get attestation by name from a specified trail or artifact.  "
---

# kosli get attestation

## Synopsis

Get attestation by name from a specified trail or artifact.  
You can get an attestation from a trail or artifact using its name. The attestation name should be given
WITHOUT dot-notation.

To get an attestation from a trail, specify the trail name using the --trail flag.  
To get an attestation from an artifact, specify the artifact fingerprint using the --fingerprint flag.

In both cases the flow must also be specified using the --flow flag.

If there are multiple attestations with the same name on the trail or artifact, a list of all will be returned.


```shell
kosli get attestation ATTESTATION-NAME [flags]
```

## Flags
| Flag | Description |
| :--- | :--- |
|    -F, --fingerprint string  |  [conditional] The fingerprint of the artifact for the attestation. Cannot be used together with --trail.  |
|    -f, --flow string  |  The Kosli flow name.  |
|    -h, --help  |  help for attestation  |
|    -o, --output string  |  [defaulted] The format of the output. Valid formats are: [table, json]. (default "table")  |
|    -t, --trail string  |  [conditional] The name of the Kosli trailfor the attestation. Cannot be used together with --fingerprint.  |


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

**get an attestation from a trail (requires the --trail flag)**

```shell
kosli get attestation attestationName 

```

**get an attestation from an artifact**

```shell
kosli get attestation attestationName 
	--fingerprint fingerprint
```

