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
|    -t, --artifact-type string  |  [conditional] The type of the artifact to calculate its SHA256 fingerprint. One of: [docker, file, dir]. Only required if you don't specify '--sha256'.  |
|        --assert  |  [optional] Exit with non-zero code if no pull requests found for the given commit.  |
|        --bitbucket-password string  |  Bitbucket password.  |
|        --bitbucket-username string  |  Bitbucket user name.  |
|        --bitbucket-workspace string  |  Bitbucket workspace.  |
|    -b, --build-url string  |  The url of CI pipeline that generated the evidence. (defaulted in some CIs: https://docs.kosli.com/ci-defaults ).  |
|        --commit string  |  Git commit for which to find pull request evidence. (defaulted in some CIs: https://docs.kosli.com/ci-defaults ).  |
|    -d, --description string  |  [optional] The evidence description.  |
|    -e, --evidence-type string  |  The type of evidence being reported.  |
|    -h, --help  |  help for bitbucket-pullrequest  |
|    -p, --pipeline string  |  The Kosli pipeline name.  |
|        --registry-password string  |  [conditional] The docker registry password or access token. Only required if you want to read docker image SHA256 digest from a remote docker registry.  |
|        --registry-provider string  |  [conditional] The docker registry provider or url. Only required if you want to read docker image SHA256 digest from a remote docker registry.  |
|        --registry-username string  |  [conditional] The docker registry username. Only required if you want to read docker image SHA256 digest from a remote docker registry.  |
|        --repository string  |  Git repository. (defaulted in some CIs: https://docs.kosli.com/ci-defaults ).  |
|    -s, --sha256 string  |  [conditional] The SHA256 fingerprint for the artifact. Only required if you don't specify '--artifact-type'.  |


### Options inherited from parent commands
| Flag | Description |
| :--- | :--- |
|    -a, --api-token string  |  The Kosli API token.  |
|    -c, --config-file string  |  [optional] The Kosli config file path. (default "kosli")  |
|    -D, --dry-run  |  [optional] Whether to run in dry-run mode. When enabled, data is not sent to Kosli and the CLI exits with 0 exit code regardless of errors.  |
|    -H, --host string  |  [defaulted] The Kosli endpoint. (default "https://app.kosli.com")  |
|    -r, --max-api-retries int  |  [defaulted] How many times should API calls be retried when the API host is not reachable. (default 3)  |
|        --owner string  |  The Kosli user or organization.  |
|    -v, --verbose  |  [optional] Print verbose logs to stdout.  |


