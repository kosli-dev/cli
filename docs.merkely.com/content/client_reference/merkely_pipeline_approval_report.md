---
title: "merkely pipeline approval report"
---

## merkely pipeline approval report

Report to Merkely an approval of deploying an artifact. 

### Synopsis


Report to Merkely an approval of deploying an artifact.
The artifact SHA256 fingerprint is calculated (based on --artifact-type flag) or alternatively it can be provided directly (with --sha256 flag).

```shell
merkely pipeline approval report [ARTIFACT-NAME-OR-PATH] [flags]
```

### Flags
| Flag | Description |
| :--- | :--- |
|    -t, --artifact-type string  |  The type of the artifact to calculate its SHA256 fingerprint. One of: [docker, file, dir]. Only required if you don't specify 'sha256'  |
|    -d, --description string  |  [optional] The approval description.  |
|    -h, --help  |  help for report  |
|        --newest-commit string  |  The source commit sha for the newest change in the deployment. (default "HEAD")  |
|        --oldest-commit string  |  The source commit sha for the oldest change in the deployment.  |
|    -p, --pipeline string  |  The Merkely pipeline name.  |
|        --registry-password string  |  The docker registry password or access token.  |
|        --registry-provider string  |  The docker registry provider or url.  |
|        --registry-username string  |  The docker registry username.  |
|        --repo-root string  |  The directory where the source git repository is volume-mounted. (default ".")  |
|    -s, --sha256 string  |  The SHA256 fingerprint for the artifact. Only required if you don't specify 'artifact-type'.  |
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


### Examples

```shell

# Report that a file type artifact has been approved for deployment.
# The approval is for the last 5 git commits
merkely pipeline approval report FILE.tgz \
	--api-token yourAPIToken \
	--artifact-type file \
	--description "An optional description for the approval" \
	--newest-commit $(git rev-parse HEAD) \
	--oldest-commit $(git rev-parse HEAD~5) \
	--owner yourOrgName \
	--pipeline yourPipelineName 

# Report that an artifact with a provided fingerprint (sha256) has been approved for deployment.
# The approval is for the last 5 git commits
merkely pipeline approval report \
	--api-token yourAPIToken \
	--description "An optional description for the approval" \
	--newest-commit $(git rev-parse HEAD) \
	--oldest-commit $(git rev-parse HEAD~5) \
	--owner yourOrgName \
	--pipeline yourPipelineName \
	--sha256 yourSha256


```

