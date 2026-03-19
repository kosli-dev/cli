---
title: "kosli assert pullrequest bitbucket"
beta: false
deprecated: false
summary: "Assert a Bitbucket pull request for a git commit exists.  "
---

# kosli assert pullrequest bitbucket

## Synopsis

```shell
kosli assert pullrequest bitbucket [flags]
```

Assert a Bitbucket pull request for a git commit exists.  
The command exits with non-zero exit code if no pull requests were found for the commit.
Authentication to Bitbucket can be done with access token (recommended) or app passwords. Credentials need to have read access for both repos and pull requests.

## Flags
| Flag | Description |
| :--- | :--- |
|        --bitbucket-access-token string  |  Bitbucket repo/project/workspace access token. See https://developer.atlassian.com/cloud/bitbucket/rest/intro/#access-tokens for more details.  |
|        --bitbucket-password string  |  Bitbucket App password. See https://developer.atlassian.com/cloud/bitbucket/rest/intro/#authentication for more details.  |
|        --bitbucket-username string  |  Bitbucket username. Only needed if you use --bitbucket-password  |
|        --bitbucket-workspace string  |  Bitbucket workspace ID.  |
|        --commit string  |  Git commit for which to find pull request evidence. (defaulted in some CIs: https://docs.kosli.com/ci-defaults ). (default "HEAD")  |
|    -D, --dry-run  |  [optional] Run in dry-run mode. When enabled, no data is sent to Kosli and the CLI exits with 0 exit code regardless of any errors.  |
|    -h, --help  |  help for bitbucket  |
|        --repository string  |  Git repository. (defaulted in some CIs: https://docs.kosli.com/ci-defaults ).  |


## Flags inherited from parent commands
| Flag | Description |
| :--- | :--- |
|    -a, --api-token string  |  The Kosli API token.  |
|    -c, --config-file string  |  [optional] The Kosli config file path. (default "kosli")  |
|        --debug  |  [optional] Print debug logs to stdout. A boolean flag https://docs.kosli.com/faq/#boolean-flags (default false)  |
|    -H, --host string  |  [defaulted] The Kosli endpoint. (default "https://app.kosli.com")  |
|        --http-proxy string  |  [optional] The HTTP proxy URL including protocol and port number. e.g. 'http://proxy-server-ip:proxy-port'  |
|    -r, --max-api-retries int  |  [defaulted] How many times should API calls be retried when the API host is not reachable. (default 3)  |
|        --org string  |  The Kosli organization.  |


## Examples Use Cases

These examples all assume that the flags  `--api-token`, `--org`, `--host`, (and `--flow`, `--trail` when required), are [set/provided](https://docs.kosli.com/getting_started/install/#assigning-flags-via-environment-variables). 

```shell
kosli assert pullrequest bitbucket  \
	--bitbucket-access-token yourBitbucketAccessToken \
	--bitbucket-workspace yourBitbucketWorkspace \
	--commit yourGitCommit \
	--repository yourBitbucketGitRepository
```

