---
title: "kosli assert github-pullrequest"
---

## kosli assert github-pullrequest

Assert if a Github pull request for a git commit exists.

### Synopsis

Assert if a Github pull request for a git commit exists.
The command exits with non-zero exit code 
if no pull requests were found for the commit.

```shell
kosli assert github-pullrequest [flags]
```

### Flags
| Flag | Description |
| :--- | :--- |
|        --commit string  |  Git commit for which to find pull request evidence. (defaulted in some CIs: https://docs.kosli.com/ci-defaults ).  |
|    -D, --dry-run  |  [optional] Run in dry-run mode. When enabled, no data is sent to Kosli and the CLI exits with 0 exit code regardless of any errors.  |
|        --github-base-url string  |  [optional] GitHub base URL (only needed for GitHub Enterprise installations).  |
|        --github-org string  |  Github organization. (defaulted if you are running in GitHub Actions: https://docs.kosli.com/ci-defaults ).  |
|        --github-token string  |  Github token.  |
|    -h, --help  |  help for github-pullrequest  |
|        --repository string  |  Git repository. (defaulted in some CIs: https://docs.kosli.com/ci-defaults ).  |


### Options inherited from parent commands
| Flag | Description |
| :--- | :--- |
|    -a, --api-token string  |  The Kosli API token.  |
|    -c, --config-file string  |  [optional] The Kosli config file path. (default "kosli")  |
|        --debug  |  [optional] Print debug logs to stdout.  |
|    -H, --host string  |  [defaulted] The Kosli endpoint. (default "https://app.kosli.com")  |
|    -r, --max-api-retries int  |  [defaulted] How many times should API calls be retried when the API host is not reachable. (default 3)  |
|        --owner string  |  The Kosli user or organization.  |


### Examples

```shell

kosli assert github-pullrequest  \
	--github-token yourGithubToken \
	--github-org yourGithubOrg \
	--commit yourArtifactGitCommit \
	--commit yourGitCommit \
	--repository yourGithubGitRepository

```

