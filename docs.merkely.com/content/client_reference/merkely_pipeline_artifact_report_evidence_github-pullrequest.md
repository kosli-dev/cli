---
title: "merkely pipeline artifact report evidence github-pullrequest"
---

## merkely pipeline artifact report evidence github-pullrequest

Report a Github pull request evidence for an artifact in a Merkely pipeline.

### Synopsis


   Check if a pull request exists for an artifact and report the pull-request evidence to the artifact in Merkely. 
   The artifact SHA256 fingerprint is calculated or alternatively it can be provided directly. 
   

```shell
merkely pipeline artifact report evidence github-pullrequest [ARTIFACT-NAME-OR-PATH] [flags]
```

### Flags
| Flag | Description |
| :--- | :--- |
|    -t, --artifact-type string  |  The type of the artifact to calculate its SHA256 fingerprint. One of: [docker, file, dir]. Only required if you don't specify 'sha256'  |
|        --assert  |  Exit with non-zero code if no pull requests found for the given commit.  |
|    -b, --build-url string  |  The url of CI pipeline that generated the evidence.  |
|        --commit string  |  Git commit for which to find pull request evidence.  |
|    -d, --description string  |  [optional] The evidence description.  |
|    -e, --evidence-type string  |  The type of evidence being reported.  |
|        --github-org string  |  Github organization.  |
|        --github-token string  |  Github token.  |
|    -h, --help  |  help for github-pullrequest  |
|    -p, --pipeline string  |  The Merkely pipeline name.  |
|        --registry-password string  |  The docker registry password or access token.  |
|        --registry-provider string  |  The docker registry provider or url.  |
|        --registry-username string  |  The docker registry username.  |
|        --repository string  |  Git repository.  |
|    -s, --sha256 string  |  The SHA256 fingerprint for the artifact. Only required if you don't specify 'artifact-type'.  |


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


