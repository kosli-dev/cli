---
title: "kosli pipeline artifact report evidence generic"
---

# kosli pipeline artifact report evidence generic

## Synopsis

Report a generic evidence to an artifact in a Kosli pipeline.
The artifact SHA256 fingerprint is calculated (based on --artifact-type flag) or alternatively it can be provided directly (with --sha256 flag).

```shell
kosli pipeline artifact report evidence generic [IMAGE-NAME | FILE-PATH | DIR-PATH] [flags]
```

## Flags
| Flag | Description |
| :--- | :--- |
|    -t, --artifact-type string  |  [conditional] The type of the artifact to calculate its SHA256 fingerprint. One of: [docker, file, dir]. Only required if you don't specify '--sha256' or '--fingerprint'.  |
|    -b, --build-url string  |  The url of CI pipeline that generated the evidence. (defaulted in some CIs: https://docs.kosli.com/ci-defaults ).  |
|    -C, --compliant  |  [defaulted] Whether the evidence is compliant or not. (default true)  |
|    -d, --description string  |  [optional] The evidence description.  |
|    -D, --dry-run  |  [optional] Run in dry-run mode. When enabled, no data is sent to Kosli and the CLI exits with 0 exit code regardless of any errors.  |
|        --evidence-fingerprint string  |  The fingerprint of the evidence.  |
|        --evidence-url string  |  The URL to the evidence.  |
|    -f, --fingerprint string  |  [conditional] The SHA256 fingerprint for the artifact. Only required if you don't specify '--artifact-type'.  |
|    -h, --help  |  help for generic  |
|    -n, --name string  |  The name of the evidence.  |
|    -p, --pipeline string  |  The Kosli pipeline name.  |
|        --registry-password string  |  [conditional] The docker registry password or access token. Only required if you want to read docker image SHA256 digest from a remote docker registry.  |
|        --registry-provider string  |  [conditional] The docker registry provider or url. Only required if you want to read docker image SHA256 digest from a remote docker registry.  |
|        --registry-username string  |  [conditional] The docker registry username. Only required if you want to read docker image SHA256 digest from a remote docker registry.  |
|    -u, --user-data string  |  [optional] The path to a JSON file containing additional data you would like to attach to this evidence.  |


## Options inherited from parent commands
| Flag | Description |
| :--- | :--- |
|    -a, --api-token string  |  The Kosli API token.  |
|    -c, --config-file string  |  [optional] The Kosli config file path. (default "kosli")  |
|        --debug  |  [optional] Print debug logs to stdout.  |
|    -H, --host string  |  [defaulted] The Kosli endpoint. (default "https://app.kosli.com")  |
|    -r, --max-api-retries int  |  [defaulted] How many times should API calls be retried when the API host is not reachable. (default 3)  |
|        --owner string  |  The Kosli user or organization.  |


## Examples

```shell

# report a generic evidence about a pre-built docker image:
kosli pipeline artifact report evidence generic yourDockerImageName \
	--api-token yourAPIToken \
	--artifact-type docker \
	--build-url https://exampleci.com \
	--evidence-type yourEvidenceType \
	--owner yourOrgName \
	--pipeline yourPipelineName 

# report a generic evidence about a directory type artifact:
kosli pipeline artifact report evidence generic /path/to/your/dir \
	--api-token yourAPIToken \
	--artifact-type dir \
	--build-url https://exampleci.com \
	--evidence-type yourEvidenceType \
	--owner yourOrgName	\
	--pipeline yourPipelineName 

# report a generic evidence about an artifact with a provided fingerprint (sha256)
kosli pipeline artifact report evidence generic \
	--api-token yourAPIToken \
	--build-url https://exampleci.com \	
	--evidence-type yourEvidenceType \
	--owner yourOrgName \
	--pipeline yourPipelineName \
	--sha256 yourSha256

```

