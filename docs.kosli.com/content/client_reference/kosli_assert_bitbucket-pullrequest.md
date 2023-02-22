---
title: "kosli assert bitbucket-pullrequest"
---

# kosli assert bitbucket-pullrequest

## Synopsis

Assert if a Bitbucket pull request for a git commit exists. 
The command exits with non-zero exit code 
if no pull requests were found for the commit.

```shell
kosli assert bitbucket-pullrequest [flags]
```

## Flags
| Flag | Description |
| :--- | :--- |
|        --bitbucket-password string  |  Bitbucket App password. See https://developer.atlassian.com/cloud/bitbucket/rest/intro/#authentication for more details.  |
|        --bitbucket-username string  |  Bitbucket username.  |
|        --bitbucket-workspace string  |  Bitbucket workspace ID.  |
|        --commit string  |  Git commit for which to find pull request evidence. (defaulted in some CIs: https://docs.kosli.com/ci-defaults ).  |
|    -D, --dry-run  |  [optional] Run in dry-run mode. When enabled, no data is sent to Kosli and the CLI exits with 0 exit code regardless of any errors.  |
|    -h, --help  |  help for bitbucket-pullrequest  |
|        --repository string  |  Git repository. (defaulted in some CIs: https://docs.kosli.com/ci-defaults ).  |


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

kosli assert bitbucket-pullrequest  \
	--bitbucket-username yourBitbucketUsername \
	--bitbucket-password yourBitbucketPassword \
	--bitbucket-workspace yourBitbucketWorkspace \
	--commit yourGitCommit \
	--repository yourBitbucketGitRepository

```

