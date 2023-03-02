---
title: "kosli pipeline artifact report evidence gitlab-mergerequest"
---

# kosli pipeline artifact report evidence gitlab-mergerequest

## Synopsis

Report a Gitlab merge request evidence for an artifact in a Kosli pipeline.
It checks if a merge request exists for the artifact (based on its git commit) and report the merge request evidence to the artifact in Kosli. 
The artifact SHA256 fingerprint is calculated (based on --artifact-type flag) or alternatively it can be provided directly (with --sha256 flag).

```shell
kosli pipeline artifact report evidence gitlab-mergerequest [IMAGE-NAME | FILE-PATH | DIR-PATH] [flags]
```

## Flags
| Flag | Description |
| :--- | :--- |
|    -t, --artifact-type string  |  [conditional] The type of the artifact to calculate its SHA256 fingerprint. One of: [docker, file, dir]. Only required if you don't specify '--sha256' or '--fingerprint'.  |
|        --assert  |  [optional] Exit with non-zero code if no pull requests found for the given commit.  |
|    -b, --build-url string  |  The url of CI pipeline that generated the evidence. (defaulted in some CIs: https://docs.kosli.com/ci-defaults ).  |
|        --commit string  |  Git commit for which to find pull request evidence. (defaulted in some CIs: https://docs.kosli.com/ci-defaults ).  |
|    -D, --dry-run  |  [optional] Run in dry-run mode. When enabled, no data is sent to Kosli and the CLI exits with 0 exit code regardless of any errors.  |
|        --evidence-fingerprint string  |  The fingerprint of the evidence.  |
|        --evidence-url string  |  The URL to the evidence.  |
|    -f, --fingerprint string  |  [conditional] The SHA256 fingerprint for the artifact. Only required if you don't specify '--artifact-type'.  |
|        --gitlab-base-url string  |  [optional] Gitlab base URL (only needed for on-prem Gitlab installations).  |
|        --gitlab-org string  |  Gitlab organization. (defaulted if you are running in Gitlab Pipelines: https://docs.kosli.com/ci-defaults ).  |
|        --gitlab-token string  |  Gitlab token.  |
|    -h, --help  |  help for gitlab-mergerequest  |
|    -n, --name string  |  The name of the evidence.  |
|    -p, --pipeline string  |  The Kosli pipeline name.  |
|        --registry-password string  |  [conditional] The docker registry password or access token. Only required if you want to read docker image SHA256 digest from a remote docker registry.  |
|        --registry-provider string  |  [conditional] The docker registry provider or url. Only required if you want to read docker image SHA256 digest from a remote docker registry.  |
|        --registry-username string  |  [conditional] The docker registry username. Only required if you want to read docker image SHA256 digest from a remote docker registry.  |
|        --repository string  |  Git repository. (defaulted in some CIs: https://docs.kosli.com/ci-defaults ).  |
|    -u, --user-data string  |  [optional] The path to a JSON file containing additional data you would like to attach to this evidence.  |


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

# report a merge request evidence to kosli for a docker image
kosli pipeline artifact report evidence gitlab-mergerequest yourDockerImageName \
	--artifact-type docker \
	--build-url https://exampleci.com \
	--name yourEvidenceName \
	--pipeline yourPipelineName \
	--gitlab-token yourGitlabToken \
	--gitlab-org yourGitlabOrg \
	--commit yourArtifactGitCommit \
	--repository yourGithubGitRepository \
	--owner yourOrgName \
	--api-token yourAPIToken

# report a merge request evidence (from an on-prem Gitlab) to kosli for a docker image 
kosli pipeline artifact report evidence gitlab-mergerequest yourDockerImageName \
	--artifact-type docker \
	--build-url https://exampleci.com \
	--name yourEvidenceName \
	--pipeline yourPipelineName \
	--gitlab-base-url https://gitlab.example.org \
	--gitlab-token yourGitlabToken \
	--gitlab-org yourGitlabOrg \
	--commit yourArtifactGitCommit \
	--repository yourGithubGitRepository \
	--owner yourOrgName \
	--api-token yourAPIToken
	
# fail if a merge request does not exist for your artifact
kosli pipeline artifact report evidence gitlab-mergerequest yourDockerImageName \
	--artifact-type docker \
	--build-url https://exampleci.com \
	--name yourEvidenceName \
	--pipeline yourPipelineName \
	--gitlab-token yourGitlabToken \
	--gitlab-org yourGitlabOrg \
	--commit yourArtifactGitCommit \
	--repository yourGithubGitRepository \
	--owner yourOrgName \
	--api-token yourAPIToken \
	--assert

```

