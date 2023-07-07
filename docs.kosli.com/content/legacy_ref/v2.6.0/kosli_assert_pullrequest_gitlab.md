---
title: "kosli assert pullrequest gitlab"
beta: false
---

# kosli assert pullrequest gitlab

## Synopsis

Assert a Gitlab merge request for a git commit exists.
The command exits with non-zero exit code 
if no merge requests were found for the commit.

```shell
kosli assert pullrequest gitlab [flags]
```

## Flags
| Flag | Description |
| :--- | :--- |
|        --commit string  |  Git commit for which to find pull request evidence. (defaulted in some CIs: https://docs.kosli.com/ci-defaults ).  |
|    -D, --dry-run  |  [optional] Run in dry-run mode. When enabled, no data is sent to Kosli and the CLI exits with 0 exit code regardless of any errors.  |
|        --gitlab-base-url string  |  [optional] Gitlab base URL (only needed for on-prem Gitlab installations).  |
|        --gitlab-org string  |  Gitlab organization. (defaulted if you are running in Gitlab Pipelines: https://docs.kosli.com/ci-defaults ).  |
|        --gitlab-token string  |  Gitlab token.  |
|    -h, --help  |  help for gitlab  |
|        --repository string  |  Git repository. (defaulted in some CIs: https://docs.kosli.com/ci-defaults ).  |


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

kosli assert mergerequest gitlab \
	--github-token yourGithubToken \
	--github-org yourGithubOrg \
	--commit yourArtifactGitCommit \
	--commit yourGitCommit \
	--repository yourGithubGitRepository

```

