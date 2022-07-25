---
title: "kosli pipeline artifact report evidence generic"
---

## kosli pipeline artifact report evidence generic

Report a generic evidence to an artifact in a Kosli pipeline. 

### Synopsis


   Report a generic evidence to an artifact to a Kosli pipeline. 
   The artifact SHA256 fingerprint is calculated (based on --artifact-type flag) or alternatively it can be provided directly (with --sha256 flag).

```shell
kosli pipeline artifact report evidence generic [ARTIFACT-NAME-OR-PATH] [flags]
```

### Flags
| Flag | Description |
| :--- | :--- |
|    -t, --artifact-type string  |  The type of the artifact to calculate its SHA256 fingerprint. One of: [docker, file, dir]. Only required if you don't specify 'sha256'  |
|    -b, --build-url string  |  The url of CI pipeline that generated the evidence.  |
|    -C, --compliant  |  Whether the evidence is compliant or not. (default true)  |
|    -d, --description string  |  [optional] The evidence description.  |
|    -e, --evidence-type string  |  The type of evidence being reported.  |
|    -h, --help  |  help for generic  |
|    -p, --pipeline string  |  The Kosli pipeline name.  |
|        --registry-password string  |  The docker registry password or access token.  |
|        --registry-provider string  |  The docker registry provider or url.  |
|        --registry-username string  |  The docker registry username.  |
|    -s, --sha256 string  |  The SHA256 fingerprint for the artifact. Only required if you don't specify 'artifact-type'.  |
|    -u, --user-data string  |  [optional] The path to a JSON file containing additional data you would like to attach to this evidence.  |


### Options inherited from parent commands
| Flag | Description |
| :--- | :--- |
|    -a, --api-token string  |  The Kosli API token.  |
|    -c, --config-file string  |  [optional] The Kosli config file path. (default "merkely")  |
|    -D, --dry-run  |  Whether to run in dry-run mode. When enabled, data is not sent to Kosli and the CLI exits with 0 exit code regardless of errors.  |
|    -H, --host string  |  The Kosli endpoint. (default "https://app.merkely.com")  |
|    -r, --max-api-retries int  |  How many times should API calls be retried when the API host is not reachable. (default 3)  |
|    -o, --owner string  |  The Kosli user or organization.  |
|    -v, --verbose  |  Print verbose logs to stdout.  |


### Examples

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

