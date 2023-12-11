---
title: "kosli report evidence artifact pullrequest github"
beta: false
---

# kosli report evidence artifact pullrequest github

## Synopsis

Report a Github pull request evidence for an artifact in a Kosli flow.  
It checks if a pull request exists for the artifact (based on its git commit) and reports the pull-request evidence to the artifact in Kosli.  
The artifact SHA256 fingerprint is calculated (based on --artifact-type flag) or alternatively it can be provided directly (with --fingerprint flag).

```shell
kosli report evidence artifact pullrequest github [IMAGE-NAME | FILE-PATH | DIR-PATH] [flags]
```

## Flags
| Flag | Description |
| :--- | :--- |
|    -t, --artifact-type string  |  [conditional] The type of the artifact to calculate its SHA256 fingerprint. One of: [docker, file, dir]. Only required if you don't specify '--fingerprint'.  |
|        --assert  |  [optional] Exit with non-zero code if no pull requests found for the given commit.  |
|    -b, --build-url string  |  The url of CI pipeline that generated the evidence. (defaulted in some CIs: https://docs.kosli.com/ci-defaults ).  |
|        --commit string  |  Git commit for which to find pull request evidence. (defaulted in some CIs: https://docs.kosli.com/ci-defaults ).  |
|    -D, --dry-run  |  [optional] Run in dry-run mode. When enabled, no data is sent to Kosli and the CLI exits with 0 exit code regardless of any errors.  |
|        --evidence-fingerprint string  |  [optional] The SHA256 fingerprint of the evidence file or dir.  |
|        --evidence-url string  |  [optional] The external URL where the evidence file or dir is stored.  |
|    -x, --exclude strings  |  [optional] The comma separated list of directories and files to exclude from fingerprinting. Only applicable for --artifact-type dir.  |
|    -F, --fingerprint string  |  [conditional] The SHA256 fingerprint of the artifact. Only required if you don't specify '--artifact-type'.  |
|    -f, --flow string  |  The Kosli flow name.  |
|        --github-base-url string  |  [optional] GitHub base URL (only needed for GitHub Enterprise installations).  |
|        --github-org string  |  Github organization. (defaulted if you are running in GitHub Actions: https://docs.kosli.com/ci-defaults ).  |
|        --github-token string  |  Github token.  |
|    -h, --help  |  help for github  |
|    -n, --name string  |  The name of the evidence.  |
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
|        --debug  |  [optional] Print debug logs to stdout. A boolean flag https://docs.kosli.com/faq/#boolean-flags (default false)  |
|    -H, --host string  |  [defaulted] The Kosli endpoint. (default "https://app.kosli.com")  |
|    -r, --max-api-retries int  |  [defaulted] How many times should API calls be retried when the API host is not reachable. (default 3)  |
|        --org string  |  The Kosli organization.  |


## Examples

```shell

# report a pull request evidence to kosli for a docker image
kosli report evidence artifact pullrequest github yourDockerImageName \
	--artifact-type docker \
	--build-url https://exampleci.com \
	--name yourEvidenceName \
	--flow yourFlowName \
	--github-token yourGithubToken \
	--github-org yourGithubOrg \
	--commit yourArtifactGitCommit \
	--repository yourGithubGitRepository \
	--org yourOrgName \
	--api-token yourAPIToken
	
# fail if a pull request does not exist for your artifact
kosli report evidence artifact pullrequest github yourDockerImageName \
	--artifact-type docker \
	--build-url https://exampleci.com \
	--name yourEvidenceName \
	--flow yourFlowName \
	--github-token yourGithubToken \
	--github-org yourGithubOrg \
	--commit yourArtifactGitCommit \
	--repository yourGithubGitRepository \
	--org yourOrgName \
	--api-token yourAPIToken \
	--assert

```

