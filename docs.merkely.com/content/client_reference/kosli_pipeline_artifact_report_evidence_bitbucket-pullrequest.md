---
title: "kosli pipeline artifact report evidence bitbucket-pullrequest"
---

## kosli pipeline artifact report evidence bitbucket-pullrequest

Report a Bitbucket pull request evidence for an artifact in a Kosli pipeline.

### Synopsis


   Check if a pull request exists for an artifact and report the pull-request evidence to the artifact in Kosli. 
   The artifact SHA256 fingerprint is calculated or alternatively it can be provided directly. 
   

```shell
kosli pipeline artifact report evidence bitbucket-pullrequest [ARTIFACT-NAME-OR-PATH] [flags]
```

### Flags
| Flag | Description |
| :--- | :--- |
|    -t, --artifact-type string  |  The type of the artifact to calculate its SHA256 fingerprint. One of: [docker, file, dir]. Only required if you don't specify 'sha256'  |
|        --assert  |  Exit with non-zero code if no pull requests found for the given commit.  |
|        --bitbucket-password string  |  Bitbucket password.  |
|        --bitbucket-username string  |  Bitbucket user name.  |
|        --bitbucket-workspace string  |  Bitbucket workspace.  |
|    -b, --build-url string  |  The url of CI pipeline that generated the evidence.  |
|        --commit string  |  Git commit for which to find pull request evidence.  |
|    -d, --description string  |  [optional] The evidence description.  |
|    -e, --evidence-type string  |  The type of evidence being reported.  |
|    -h, --help  |  help for bitbucket-pullrequest  |
|    -p, --pipeline string  |  The Kosli pipeline name.  |
|        --registry-password string  |  The docker registry password or access token.  |
|        --registry-provider string  |  The docker registry provider or url.  |
|        --registry-username string  |  The docker registry username.  |
|        --repository string  |  Git repository.  |
|    -s, --sha256 string  |  The SHA256 fingerprint for the artifact. Only required if you don't specify 'artifact-type'.  |


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


