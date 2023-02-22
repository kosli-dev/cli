---
title: "kosli pipeline approval request"
---

# kosli pipeline approval request

## Synopsis

Request an approval of a deployment of an artifact in Kosli.
The request should be reviewed in Kosli UI.The artifact SHA256 fingerprint is calculated (based on --artifact-type flag) or alternatively it can be provided directly (with --sha256 flag).

```shell
kosli pipeline approval request [IMAGE-NAME | FILE-PATH | DIR-PATH] [flags]
```

## Flags
| Flag | Description |
| :--- | :--- |
|    -t, --artifact-type string  |  [conditional] The type of the artifact to calculate its SHA256 fingerprint. One of: [docker, file, dir]. Only required if you don't specify '--sha256' or '--fingerprint'.  |
|    -d, --description string  |  [optional] The approval description.  |
|    -D, --dry-run  |  [optional] Run in dry-run mode. When enabled, no data is sent to Kosli and the CLI exits with 0 exit code regardless of any errors.  |
|    -h, --help  |  help for request  |
|        --newest-commit string  |  [defaulted] The source commit sha for the newest change in the deployment. (default "HEAD")  |
|        --oldest-commit string  |  The source commit sha for the oldest change in the deployment.  |
|    -p, --pipeline string  |  The Kosli pipeline name.  |
|        --registry-password string  |  [conditional] The docker registry password or access token. Only required if you want to read docker image SHA256 digest from a remote docker registry.  |
|        --registry-provider string  |  [conditional] The docker registry provider or url. Only required if you want to read docker image SHA256 digest from a remote docker registry.  |
|        --registry-username string  |  [conditional] The docker registry username. Only required if you want to read docker image SHA256 digest from a remote docker registry.  |
|        --repo-root string  |  [defaulted] The directory where the source git repository is volume-mounted. (default ".")  |
|    -s, --sha256 string  |  [conditional] The SHA256 fingerprint for the artifact. Only required if you don't specify '--artifact-type'.  |
|    -u, --user-data string  |  [optional] The path to a JSON file containing additional data you would like to attach to this approval.  |


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

# Request that a file type artifact needs approval.
# The approval is for the last 5 git commits
kosli pipeline approval request FILE.tgz \
	--api-token yourAPIToken \
	--artifact-type file \
	--description "An optional description for the requested approval" \
	--newest-commit $(git rev-parse HEAD) \
	--oldest-commit $(git rev-parse HEAD~5) \
	--owner yourOrgName \
	--pipeline yourPipelineName 

# Request and approval for an artifact with a provided fingerprint (sha256).
# The approval is for the last 5 git commits
kosli pipeline approval request \
	--api-token yourAPIToken \
	--description "An optional description for the requested approval" \
	--newest-commit $(git rev-parse HEAD) \
	--oldest-commit $(git rev-parse HEAD~5)	\
	--owner yourOrgName \
	--pipeline yourPipelineName \
	--sha256 yourSha256 

```

