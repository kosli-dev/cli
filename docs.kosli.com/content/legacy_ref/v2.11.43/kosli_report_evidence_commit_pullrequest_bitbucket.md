---
title: "kosli report evidence commit pullrequest bitbucket"
beta: false
deprecated: true
summary: "Report Bitbucket pull request evidence for a commit in Kosli flows.  "
---

# kosli report evidence commit pullrequest bitbucket

{{% hint danger %}}
**kosli report evidence commit pullrequest bitbucket** is deprecated. See **kosli attest** commands.  Deprecated commands will be removed in a future release.
{{% /hint %}}
## Synopsis

```shell
kosli report evidence commit pullrequest bitbucket [flags]
```

Report Bitbucket pull request evidence for a commit in Kosli flows.  
It checks if a pull request exists for the git commit and reports the pull-request evidence to the commit in Kosli.

## Flags
| Flag | Description |
| :--- | :--- |
|        --assert  |  [optional] Exit with non-zero code if no pull requests found for the given commit.  |
|        --bitbucket-access-token string  |  Bitbucket repo/project/workspace access token. See https://developer.atlassian.com/cloud/bitbucket/rest/intro/#access-tokens for more details.  |
|        --bitbucket-password string  |  Bitbucket App password. See https://developer.atlassian.com/cloud/bitbucket/rest/intro/#authentication for more details.  |
|        --bitbucket-username string  |  Bitbucket username. Only needed if you use --bitbucket-password  |
|        --bitbucket-workspace string  |  Bitbucket workspace ID.  |
|    -b, --build-url string  |  The url of CI pipeline that generated the evidence. (defaulted in some CIs: https://docs.kosli.com/ci-defaults ).  |
|        --commit string  |  Git commit for which to verify a given evidence. (defaulted in some CIs: https://docs.kosli.com/ci-defaults ).  |
|    -D, --dry-run  |  [optional] Run in dry-run mode. When enabled, no data is sent to Kosli and the CLI exits with 0 exit code regardless of any errors.  |
|        --evidence-fingerprint string  |  [optional] The SHA256 fingerprint of the evidence file or dir.  |
|        --evidence-url string  |  [optional] The external URL where the evidence file or dir is stored.  |
|    -f, --flows strings  |  [defaulted] The comma separated list of Kosli flows. Defaults to all flows of the org.  |
|    -h, --help  |  help for bitbucket  |
|    -n, --name string  |  The name of the evidence.  |
|        --repository string  |  Git repository. (defaulted in some CIs: https://docs.kosli.com/ci-defaults ).  |
|    -u, --user-data string  |  [optional] The path to a JSON file containing additional data you would like to attach to the evidence.  |


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

##### report a pull request evidence to Kosli

```shell
kosli report evidence commit pullrequest bitbucket 
	--commit yourArtifactGitCommit 
	--repository yourBitbucketGitRepository 
	--bitbucket-username yourBitbucketUsername 
	--bitbucket-password yourBitbucketPassword 
	--bitbucket-workspace yourBitbucketWorkspace 
	--name yourEvidenceName 
	--flows yourFlowName1,yourFlowName2 
	--build-url https://exampleci.com 

```

##### fail if a pull request does not exist for your commit

```shell
kosli report evidence commit pullrequest bitbucket 
	--commit yourArtifactGitCommit 
	--repository yourBitbucketGitRepository 
	--bitbucket-username yourBitbucketUsername 
	--bitbucket-password yourBitbucketPassword 
	--bitbucket-workspace yourBitbucketWorkspace 
	--name yourEvidenceName 
	--flows yourFlowName1,yourFlowName2 
	--build-url https://exampleci.com 
	--assert
```

