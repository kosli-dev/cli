---
title: "kosli assert gitlab-mergerequest"
---

## kosli assert gitlab-mergerequest

Assert if a Gitlab pull request for a git commit exists.

### Synopsis

Assert if a Gitlab pull request for a git commit exists.
The command exits with non-zero exit code 
if no pull requests were found for the commit.

```shell
kosli assert gitlab-mergerequest [flags]
```

### Flags
| Flag | Description |
| :--- | :--- |
|        --commit string  |  Git commit for which to find pull request evidence. (defaulted in some CIs: https://docs.kosli.com/ci-defaults ).  |
|    -D, --dry-run  |  [optional] Run in dry-run mode. When enabled, no data is sent to Kosli and the CLI exits with 0 exit code regardless of any errors.  |
|        --gitlab-org string  |  Gitlab organization.  |
|        --gitlab-token string  |  Gitlab token.  |
|    -h, --help  |  help for gitlab-mergerequest  |
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

kosli assert gitlab-mergerequest \
	--github-token yourGithubToken \
	--github-org yourGithubOrg \
	--commit yourArtifactGitCommit \
	--commit yourGitCommit \
	--repository yourGithubGitRepository

```

