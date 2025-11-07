---
title: "kosli report evidence artifact junit"
beta: false
deprecated: true
summary: "Report JUnit test evidence for an artifact in a Kosli flow.  "
---

# kosli report evidence artifact junit

{{% hint danger %}}
**kosli report evidence artifact junit** is deprecated. See **kosli attest** commands.  Deprecated commands will be removed in a future release.
{{% /hint %}}
## Synopsis

```shell
kosli report evidence artifact junit [IMAGE-NAME | FILE-PATH | DIR-PATH] [flags]
```

Report JUnit test evidence for an artifact in a Kosli flow.    
All .xml files from --results-dir are parsed and uploaded to Kosli's evidence vault.  
If there are no failing tests and no errors the evidence is reported as compliant. Otherwise the evidence is reported as non-compliant.  

The artifact fingerprint can be provided directly with the `--fingerprint` flag, or 
calculated based on `--artifact-type` flag.

Artifact type can be one of: "file" for files, "dir" for directories, "oci" for container
images in registries or "docker" for local docker images.



## Flags
| Flag | Description |
| :--- | :--- |
|    -t, --artifact-type string  |  The type of the artifact to calculate its SHA256 fingerprint. One of: [oci, docker, file, dir]. Only required if you want Kosli to calculate the fingerprint for you (i.e. when you don't specify '--fingerprint' on commands that allow it).  |
|    -b, --build-url string  |  The url of CI pipeline that generated the evidence. (defaulted in some CIs: https://docs.kosli.com/ci-defaults ).  |
|    -D, --dry-run  |  [optional] Run in dry-run mode. When enabled, no data is sent to Kosli and the CLI exits with 0 exit code regardless of any errors.  |
|        --evidence-fingerprint string  |  [optional] The SHA256 fingerprint of the evidence file or dir.  |
|        --evidence-url string  |  [optional] The external URL where the evidence file or dir is stored.  |
|    -x, --exclude strings  |  [optional] The comma separated list of directories and files to exclude from fingerprinting. Can take glob patterns. Only applicable for --artifact-type dir.  |
|    -F, --fingerprint string  |  [conditional] The SHA256 fingerprint of the artifact. Only required if you don't specify '--artifact-type'.  |
|    -f, --flow string  |  The Kosli flow name.  |
|    -h, --help  |  help for junit  |
|    -n, --name string  |  The name of the evidence.  |
|        --registry-password string  |  [conditional] The container registry password or access token. Only required if you want to read container image SHA256 digest from a remote container registry.  |
|        --registry-username string  |  [conditional] The container registry username. Only required if you want to read container image SHA256 digest from a remote container registry.  |
|    -R, --results-dir string  |  [defaulted] The path to a directory with JUnit test results. By default, the directory will be uploaded to Kosli's evidence vault. (default ".")  |
|    -u, --user-data string  |  [optional] The path to a JSON file containing additional data you would like to attach to the evidence.  |


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

**report JUnit test evidence about a file artifact**

```shell
kosli report evidence artifact junit FILE.tgz 
	--artifact-type file 
	--name yourEvidenceName 
	--build-url https://exampleci.com 
	--results-dir yourFolderWithJUnitResults

```

**report JUnit test evidence about an artifact using an available Sha256 digest**

```shell
kosli report evidence artifact junit 
	--fingerprint yourSha256 
	--name yourEvidenceName 
	--build-url https://exampleci.com 
	--results-dir yourFolderWithJUnitResults
```

