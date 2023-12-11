---
title: "kosli report evidence artifact snyk"
beta: false
---

# kosli report evidence artifact snyk

## Synopsis

Report Snyk vulnerability scan evidence for an artifact in a Kosli flow.    
The --scan-results .json file is parsed and uploaded to Kosli's evidence vault.  
If no vulnerabilities are detected the evidence is reported as compliant. Otherwise the evidence is reported as non-compliant.  
The artifact SHA256 fingerprint is calculated (based on --artifact-type flag) or alternatively it can be provided directly (with --fingerprint flag).

```shell
kosli report evidence artifact snyk [IMAGE-NAME | FILE-PATH | DIR-PATH] [flags]
```

## Flags
| Flag | Description |
| :--- | :--- |
|    -t, --artifact-type string  |  [conditional] The type of the artifact to calculate its SHA256 fingerprint. One of: [docker, file, dir]. Only required if you don't specify '--fingerprint'.  |
|    -b, --build-url string  |  The url of CI pipeline that generated the evidence. (defaulted in some CIs: https://docs.kosli.com/ci-defaults ).  |
|    -D, --dry-run  |  [optional] Run in dry-run mode. When enabled, no data is sent to Kosli and the CLI exits with 0 exit code regardless of any errors.  |
|        --evidence-fingerprint string  |  [optional] The SHA256 fingerprint of the evidence file or dir.  |
|        --evidence-url string  |  [optional] The external URL where the evidence file or dir is stored.  |
|    -x, --exclude strings  |  [optional] The comma separated list of directories and files to exclude from fingerprinting. Only applicable for --artifact-type dir.  |
|    -F, --fingerprint string  |  [conditional] The SHA256 fingerprint of the artifact. Only required if you don't specify '--artifact-type'.  |
|    -f, --flow string  |  The Kosli flow name.  |
|    -h, --help  |  help for snyk  |
|    -n, --name string  |  The name of the evidence.  |
|        --registry-password string  |  [conditional] The docker registry password or access token. Only required if you want to read docker image SHA256 digest from a remote docker registry.  |
|        --registry-provider string  |  [conditional] The docker registry provider or url. Only required if you want to read docker image SHA256 digest from a remote docker registry.  |
|        --registry-username string  |  [conditional] The docker registry username. Only required if you want to read docker image SHA256 digest from a remote docker registry.  |
|    -R, --scan-results string  |  The path to Snyk scan results JSON file from 'snyk test' and 'snyk container test'. The Snyk results will be uploaded to Kosli's evidence vault.  |
|    -u, --user-data string  |  [optional] The path to a JSON file containing additional data you would like to attach to the evidence.  |


## Options inherited from parent commands
| Flag | Description |
| :--- | :--- |
|    -a, --api-token string  |  The Kosli API token.  |
|    -c, --config-file string  |  [optional] The Kosli config file path. (default "kosli")  |
|        --debug  |  [optional] Print debug logs to stdout. A boolean flag https://docs.kosli.com/faq/#boolean-flags (default false)  |
|    -H, --host string  |  [defaulted] The Kosli endpoint. (default "https://app.kosli.com")  |
|    -r, --max-api-retries int  |  [defaulted] How many times should API calls be retried when the API host is not reachable. (default 3)  |
|        --org string  |  The Kosli organization.  |


## Examples

```shell

# report Snyk vulnerability scan evidence about a file artifact:
kosli report evidence artifact snyk FILE.tgz \
	--artifact-type file \
	--name yourEvidenceName \
	--flow yourFlowName \
	--build-url https://exampleci.com \
	--api-token yourAPIToken \
	--org yourOrgName	\
	--scan-results yourSnykJSONScanResults

# report Snyk vulnerability scan evidence about an artifact using an available Sha256 digest:
kosli report evidence artifact snyk \
	--fingerprint yourSha256 \
	--name yourEvidenceName \
	--flow yourFlowName \
	--build-url https://exampleci.com \
	--api-token yourAPIToken \
	--org yourOrgName	\
	--scan-results yourSnykJSONScanResults

```

