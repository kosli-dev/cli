---
title: "kosli commit report evidence bitbucket-pullrequest"
---

# kosli commit report evidence bitbucket-pullrequest

## Synopsis

Report a Bitbucket pull request evidence for a commit in a Kosli pipeline.
It checks if a pull request exists for the git commit and reports the pull-request evidence to the commit in Kosli.

```shell
kosli commit report evidence bitbucket-pullrequest [flags]
```

## Flags
| Flag | Description |
| :--- | :--- |
|        --assert  |  [optional] Exit with non-zero code if no pull requests found for the given commit.  |
|        --bitbucket-password string  |  Bitbucket App password. See https://developer.atlassian.com/cloud/bitbucket/rest/intro/#authentication for more details.  |
|        --bitbucket-username string  |  Bitbucket username.  |
|        --bitbucket-workspace string  |  Bitbucket workspace ID.  |
|    -b, --build-url string  |  The url of CI pipeline that generated the evidence. (defaulted in some CIs: https://docs.kosli.com/ci-defaults ).  |
|        --commit string  |  Git commit for which to find pull request evidence. (defaulted in some CIs: https://docs.kosli.com/ci-defaults ).  |
|    -D, --dry-run  |  [optional] Run in dry-run mode. When enabled, no data is sent to Kosli and the CLI exits with 0 exit code regardless of any errors.  |
|        --evidence-fingerprint string  |  The fingerprint of the evidence.  |
|        --evidence-url string  |  The URL to the evidence.  |
|    -h, --help  |  help for bitbucket-pullrequest  |
|    -n, --name string  |  The name of the evidence.  |
|    -p, --pipelines strings  |  The comma separated list of pipelines for which a commit evidence belongs.  |
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

# report a pull request evidence to Kosli
kosli commit report evidence bitbucket-pullrequest \
	--commit yourArtifactGitCommit \
	--repository yourBitbucketGitRepository \
	--bitbucket-username yourBitbucketUsername \
	--bitbucket-password yourBitbucketPassword \
	--bitbucket-workspace yourBitbucketWorkspace \
	--name yourEvidenceName \
	--pipelines yourPipelineName \
	--build-url https://exampleci.com \
	--owner yourOrgName \
	--api-token yourAPIToken
	
# fail if a pull request does not exist for your commit
kosli commit report evidence bitbucket-pullrequest \
	--commit yourArtifactGitCommit \
	--repository yourBitbucketGitRepository \
	--bitbucket-username yourBitbucketUsername \
	--bitbucket-password yourBitbucketPassword \
	--bitbucket-workspace yourBitbucketWorkspace \
	--name yourEvidenceName \
	--pipelines yourPipelineName \
	--build-url https://exampleci.com \
	--owner yourOrgName \
	--api-token yourAPIToken \
	--assert

```

