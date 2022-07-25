---
title: "kosli pipeline artifact report creation"
---

## kosli pipeline artifact report creation

Report an artifact creation to a Kosli pipeline. 

### Synopsis


   Report an artifact creation to a Kosli pipeline. 
   The artifact SHA256 fingerprint is calculated (based on --artifact-type flag) or alternatively it can be provided directly (with --sha256 flag).

```shell
kosli pipeline artifact report creation ARTIFACT-NAME-OR-PATH [flags]
```

### Flags
| Flag | Description |
| :--- | :--- |
|    -t, --artifact-type string  |  The type of the artifact to calculate its SHA256 fingerprint. One of: [docker, file, dir]. Only required if you don't specify 'sha256'  |
|    -b, --build-url string  |  The url of CI pipeline that built the artifact. (defaulted in some CIs: https://docs.merkely.com/ci-defaults)  |
|    -u, --commit-url string  |  The url for the git commit that created the artifact.  |
|    -C, --compliant  |  Whether the artifact is compliant or not. (default true)  |
|    -d, --description string  |  [optional] The artifact description.  |
|    -g, --git-commit string  |  The git commit from which the artifact was created.  |
|    -h, --help  |  help for creation  |
|    -p, --pipeline string  |  The Kosli pipeline name.  |
|        --registry-password string  |  The docker registry password or access token.  |
|        --registry-provider string  |  The docker registry provider or url.  |
|        --registry-username string  |  The docker registry username.  |
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


### Examples

```shell

# Report to a Kosli pipeline that a file type artifact has been created
kosli pipeline artifact report creation FILE.tgz \
--api-token yourApiToken \
--artifact-type file \
--build-url https://exampleci.com \
--commit-url https://github.com/YourOrg/YourProject/commit/yourCommitShaThatThisArtifactWasBuiltFrom \
--git-commit yourCommitShaThatThisArtifactWasBuiltFrom \
--owner yourOrgName \
--pipeline yourPipelineName 

# Report to a Kosli pipeline that an artifact with a provided fingerprint (sha256) has been created
kosli pipeline artifact report creation \
--api-token yourApiToken \
--build-url https://exampleci.com \
--commit-url https://github.com/YourOrg/YourProject/commit/yourCommitShaThatThisArtifactWasBuiltFrom \
--git-commit yourCommitShaThatThisArtifactWasBuiltFrom \
--owner yourOrgName \
--pipeline yourPipelineName \
--sha256 yourSha256 

```

