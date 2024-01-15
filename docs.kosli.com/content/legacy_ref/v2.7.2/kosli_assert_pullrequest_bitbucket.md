---
title: "kosli assert pullrequest bitbucket"
beta: false
---

# kosli assert pullrequest bitbucket

## Synopsis

Assert a Bitbucket pull request for a git commit exists.  
The command exits with non-zero exit code 
if no pull requests were found for the commit.

```shell
kosli assert pullrequest bitbucket [flags]
```

## Flags
| Flag | Description |
| :--- | :--- |
|        --bitbucket-password string  |  Bitbucket App password. See https://developer.atlassian.com/cloud/bitbucket/rest/intro/#authentication for more details.  |
|        --bitbucket-username string  |  Bitbucket username.  |
|        --bitbucket-workspace string  |  Bitbucket workspace ID.  |
|        --commit string  |  Git commit for which to find pull request evidence. (defaulted in some CIs: https://docs.kosli.com/ci-defaults ). (default "HEAD")  |
|    -D, --dry-run  |  [optional] Run in dry-run mode. When enabled, no data is sent to Kosli and the CLI exits with 0 exit code regardless of any errors.  |
|    -h, --help  |  help for bitbucket  |
|        --repository string  |  Git repository. (defaulted in some CIs: https://docs.kosli.com/ci-defaults ).  |


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

kosli assert pullrequest bitbucket  \
	--bitbucket-username yourBitbucketUsername \
	--bitbucket-password yourBitbucketPassword \
	--bitbucket-workspace yourBitbucketWorkspace \
	--commit yourGitCommit \
	--repository yourBitbucketGitRepository

```

