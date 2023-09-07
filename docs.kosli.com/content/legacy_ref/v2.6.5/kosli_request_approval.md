---
title: "kosli request approval"
beta: false
---

# kosli request approval

## Synopsis

Request an approval of a deployment of an artifact in Kosli.  
The request should be reviewed in the Kosli UI.  
The artifact SHA256 fingerprint is calculated (based on --artifact-type flag) or alternatively it can be provided directly (with --fingerprint flag).

```shell
kosli request approval [IMAGE-NAME | FILE-PATH | DIR-PATH] [flags]
```

## Flags
| Flag | Description |
| :--- | :--- |
|    -t, --artifact-type string  |  [conditional] The type of the artifact to calculate its SHA256 fingerprint. One of: [docker, file, dir]. Only required if you don't specify '--fingerprint'.  |
|    -d, --description string  |  [optional] The approval description.  |
|    -D, --dry-run  |  [optional] Run in dry-run mode. When enabled, no data is sent to Kosli and the CLI exits with 0 exit code regardless of any errors.  |
|    -x, --exclude strings  |  [optional] The comma separated list of directories and files to exclude from fingerprinting. Only applicable for --artifact-type dir.  |
|    -F, --fingerprint string  |  [conditional] The SHA256 fingerprint of the artifact. Only required if you don't specify '--artifact-type'.  |
|    -f, --flow string  |  The Kosli flow name.  |
|    -h, --help  |  help for approval  |
|        --newest-commit string  |  [defaulted] The source commit sha for the newest change in the deployment. (default "HEAD")  |
|        --oldest-commit string  |  The source commit sha for the oldest change in the deployment.  |
|        --registry-password string  |  [conditional] The docker registry password or access token. Only required if you want to read docker image SHA256 digest from a remote docker registry.  |
|        --registry-provider string  |  [conditional] The docker registry provider or url. Only required if you want to read docker image SHA256 digest from a remote docker registry.  |
|        --registry-username string  |  [conditional] The docker registry username. Only required if you want to read docker image SHA256 digest from a remote docker registry.  |
|        --repo-root string  |  [defaulted] The directory where the source git repository is available. (default ".")  |
|    -u, --user-data string  |  [optional] The path to a JSON file containing additional data you would like to attach to this approval.  |


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

# Request that a file type artifact needs approval.
# The approval is for the last 5 git commits
kosli request approval FILE.tgz \
	--api-token yourAPIToken \
	--artifact-type file \
	--description "An optional description for the requested approval" \
	--newest-commit $(git rev-parse HEAD) \
	--oldest-commit $(git rev-parse HEAD~5) \
	--org yourOrgName \
	--flow yourFlowName 

# Request and approval for an artifact with a provided fingerprint (sha256).
# The approval is for the last 5 git commits
kosli request approval \
	--api-token yourAPIToken \
	--description "An optional description for the requested approval" \
	--newest-commit $(git rev-parse HEAD) \
	--oldest-commit $(git rev-parse HEAD~5)	\
	--org yourOrgName \
	--flow yourFlowName \
	--fingerprint yourArtifactFingerprint 

```

