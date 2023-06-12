---
title: "kosli report evidence artifact generic"
beta: false
---

# kosli report evidence artifact generic

## Synopsis

Report generic evidence to an artifact in a Kosli flow.
The artifact SHA256 fingerprint is calculated (based on --artifact-type flag) or alternatively it can be provided directly (with --fingerprint flag).

```shell
kosli report evidence artifact generic [IMAGE-NAME | FILE-PATH | DIR-PATH] [flags]
```

## Flags
| Flag | Description |
| :--- | :--- |
|    -t, --artifact-type string  |  [conditional] The type of the artifact to calculate its SHA256 fingerprint. One of: [docker, file, dir]. Only required if you don't specify '--fingerprint'.  |
|    -b, --build-url string  |  The url of CI pipeline that generated the evidence. (defaulted in some CIs: https://docs.kosli.com/ci-defaults ).  |
|    -C, --compliant  |  [defaulted] Whether the evidence is compliant or not. (default true)  |
|    -d, --description string  |  [optional] The evidence description.  |
|    -D, --dry-run  |  [optional] Run in dry-run mode. When enabled, no data is sent to Kosli and the CLI exits with 0 exit code regardless of any errors.  |
|        --evidence-fingerprint string  |  [optional] The SHA256 fingerprint of the evidence file or dir.  |
|    -e, --evidence-paths strings  |  [optional] The comma-separated list of paths containing supporting proof for the reported evidence. Paths can be for files or directories.  |
|        --evidence-url string  |  [optional] The external URL where the evidence file or dir is stored.  |
|    -F, --fingerprint string  |  [conditional] The SHA256 fingerprint of the artifact. Only required if you don't specify '--artifact-type'.  |
|    -f, --flow string  |  The Kosli flow name.  |
|    -h, --help  |  help for generic  |
|    -n, --name string  |  The name of the evidence.  |
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
|        --org string  |  The Kosli organization.  |


## Examples

```shell

# report a generic evidence about a pre-built docker image:
kosli report evidence artifact generic yourDockerImageName \
	--api-token yourAPIToken \
	--artifact-type docker \
	--build-url https://exampleci.com \
	--name yourEvidenceName \
	--org yourOrgName \
	--flow yourFlowName 

# report a generic evidence about a directory type artifact:
kosli report evidence artifact generic /path/to/your/dir \
	--api-token yourAPIToken \
	--artifact-type dir \
	--build-url https://exampleci.com \
	--name yourEvidenceName \
	--org yourOrgName	\
	--flow yourFlowName 

# report a generic evidence about an artifact with a provided fingerprint (sha256)
kosli report evidence artifact generic \
	--api-token yourAPIToken \
	--build-url https://exampleci.com \	
	--name yourEvidenceName \
	--org yourOrgName \
	--flow yourFlowName \
	--fingerprint yourArtifactFingerprint

# report a generic evidence about an artifact with evidence file upload
kosli report evidence artifact generic \
	--api-token yourAPIToken \
	--build-url https://exampleci.com \	
	--name yourEvidenceName \
	--org yourOrgName \
	--flow yourFlowName \
	--fingerprint yourArtifactFingerprint \
	--evidence-paths=yourEvidencePathName

# report a generic evidence about an artifact with evidence file upload via API
curl -X 'POST' \
	'https://app.kosli.com/api/v2/evidence/yourOrgName/artifact/yourFlowName/generic' \
	-H 'accept: application/json' \
	-H 'Content-Type: multipart/form-data' \
	-F 'evidence_json={
  	  "artifact_fingerprint": "yourArtifactFingerprint",
	  "name": "yourEvidenceName",
      "build_url": "https://exampleci.com",
      "is_compliant": true
    }' \
	-F 'evidence_file=@yourEvidencePathName'

```

