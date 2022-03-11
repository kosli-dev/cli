---
title: "merkely pipeline approval request"
---

## merkely pipeline approval request

Request an approval for deploying an artifact in Merkely. 

### Synopsis


Request an approval of a deployment of an artifact in Merkely. The request should be reviewed in Merkely UI. 
The artifact SHA256 fingerprint is calculated or alternatively it can be provided directly.

```shell
merkely pipeline approval request [ARTIFACT-NAME-OR-PATH] [flags]
```

### Examples

```shell

# Request that a file artifact needs approval.
# The approval is for the last 5 git commits
merkely pipeline approval request FILE.tgz \
	--api-token yourAPIToken \
	--owner yourOrgName \
	--pipeline yourPipelineName \
	--artifact-type file \
	--description "An optional description for the requested approval" \
	--newest-commit $(git rev-parse HEAD) \
	--oldest-commit $(git rev-parse HEAD~5)

# Request that an artifact with a sha256 needs approval.
# The approval is for the last 5 git commits
merkely pipeline approval request \
	--api-token yourAPIToken \
	--owner yourOrgName \
	--pipeline yourPipelineName \
	--sha256 yourSha256 \
	--description "An optional description for the requested approval" \
	--newest-commit $(git rev-parse HEAD) \
	--oldest-commit $(git rev-parse HEAD~5)	

```

### Options
| Flag | Description |
| :--- | :--- |
|    -t, --artifact-type string  |  The type of the artifact to calculate its SHA256 fingerprint. One of: [docker, file, dir]  |
|    -d, --description string  |  [optional] The approval description.  |
|    -h, --help  |  help for request  |
|        --newest-commit string  |  The source commit sha for the newest change in the deployment approval. (default "HEAD")  |
|        --oldest-commit string  |  The source commit sha for the oldest change in the deployment approval.  |
|    -p, --pipeline string  |  The Merkely pipeline name.  |
|        --registry-password string  |  The docker registry password or access token.  |
|        --registry-provider string  |  The docker registry provider or url.  |
|        --registry-username string  |  The docker registry username.  |
|        --repo-root string  |  The directory where the source git repository is volume-mounted. (default ".")  |
|    -s, --sha256 string  |  The SHA256 fingerprint for the artifact to be approved. Only required if you don't specify --type.  |
|    -u, --user-data string  |  [optional] The path to a JSON file containing additional data you would like to attach to this approval.  |


### Options inherited from parent commands
| Flag | Description |
| :--- | :--- |
|    -a, --api-token string  |  The merkely API token.  |
|    -c, --config-file string  |  [optional] The merkely config file path. (default "merkely")  |
|    -D, --dry-run  |  Whether to run in dry-run mode. When enabled, data is not sent to Merkely and the CLI exits with 0 exit code regardless of errors.  |
|    -H, --host string  |  The merkely endpoint. (default "https://app.merkely.com")  |
|    -r, --max-api-retries int  |  How many times should API calls be retried when the API host is not reachable. (default 3)  |
|    -o, --owner string  |  The merkely user or organization.  |
|    -v, --verbose  |  Print verbose logs to stdout.  |


