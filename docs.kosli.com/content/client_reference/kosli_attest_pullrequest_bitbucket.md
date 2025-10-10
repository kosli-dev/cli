---
title: "kosli attest pullrequest bitbucket"
beta: false
deprecated: false
summary: "Report a Bitbucket pull request attestation to an artifact or a trail in a Kosli flow.  "
---

# kosli attest pullrequest bitbucket

## Synopsis

```shell
kosli attest pullrequest bitbucket [IMAGE-NAME | FILE-PATH | DIR-PATH] [flags]
```

Report a Bitbucket pull request attestation to an artifact or a trail in a Kosli flow.  
It checks if a pull request exists for a given merge commit and reports the pull-request attestation to Kosli.
Authentication to Bitbucket can be done with access token (recommended) or app passwords. Credentials need to have read access for both repos and pull requests.


The attestation can be bound to a *trail* using the trail name.  
The attestation can be bound to an *artifact* in two ways:
- using the artifact's SHA256 fingerprint which is calculated (based on the `--artifact-type` flag and the artifact name/path argument) or can be provided directly (with the `--fingerprint` flag).
- using the artifact's name in the flow yaml template and the git commit from which the artifact is/will be created. Useful when reporting an attestation before creating/reporting the artifact.

## Flags
| Flag | Description |
| :--- | :--- |
|        --annotate stringToString  |  [optional] Annotate the attestation with data using key=value.  |
|    -t, --artifact-type string  |  The type of the artifact to calculate its SHA256 fingerprint. One of: [oci, docker, file, dir]. Only required if you want Kosli to calculate the fingerprint for you (i.e. when you don't specify '--fingerprint' on commands that allow it).  |
|        --assert  |  [optional] Exit with non-zero code if no pull requests found for the given commit.  |
|        --attachments strings  |  [optional] The comma-separated list of paths of attachments for the reported attestation. Attachments can be files or directories. All attachments are compressed and uploaded to Kosli's evidence vault.  |
|        --bitbucket-access-token string  |  Bitbucket repo/project/workspace access token. See https://developer.atlassian.com/cloud/bitbucket/rest/intro/#access-tokens for more details.  |
|        --bitbucket-password string  |  Bitbucket App password. See https://developer.atlassian.com/cloud/bitbucket/rest/intro/#authentication for more details.  |
|        --bitbucket-username string  |  Bitbucket username. Only needed if you use --bitbucket-password  |
|        --bitbucket-workspace string  |  Bitbucket workspace ID.  |
|    -g, --commit string  |  the git merge commit to be checked for associated pull requests.  |
|        --description string  |  [optional] attestation description  |
|    -D, --dry-run  |  [optional] Run in dry-run mode. When enabled, no data is sent to Kosli and the CLI exits with 0 exit code regardless of any errors.  |
|    -x, --exclude strings  |  [optional] The comma separated list of directories and files to exclude from fingerprinting. Can take glob patterns. Only applicable for --artifact-type dir.  |
|        --external-fingerprint stringToString  |  [optional] A SHA256 fingerprint of an external attachment represented by --external-url. The format is label=fingerprint (labels cannot contain '.' or '='). This flag can be set multiple times. There must be an external url with a matching label for each external fingerprint.  |
|        --external-url stringToString  |  [optional] Add labeled reference URL for an external resource. The format is label=url (labels cannot contain '.' or '='). This flag can be set multiple times. If the resource is a file or dir, you can optionally add its fingerprint via --external-fingerprint  |
|    -F, --fingerprint string  |  [conditional] The SHA256 fingerprint of the artifact to attach the attestation to. Only required if the attestation is for an artifact and --artifact-type and artifact name/path are not used.  |
|    -f, --flow string  |  The Kosli flow name.  |
|    -h, --help  |  help for bitbucket  |
|    -n, --name string  |  The name of the attestation as declared in the flow or trail yaml template.  |
|    -o, --origin-url string  |  [optional] The url pointing to where the attestation came from or is related. (defaulted to the CI url in some CIs: https://docs.kosli.com/integrations/ci_cd/#defaulted-kosli-command-flags-from-ci-variables ).  |
|        --redact-commit-info strings  |  [optional] The list of commit info to be redacted before sending to Kosli. Allowed values are one or more of [author, message, branch].  |
|        --registry-password string  |  [conditional] The container registry password or access token. Only required if you want to read container image SHA256 digest from a remote container registry.  |
|        --registry-username string  |  [conditional] The container registry username. Only required if you want to read container image SHA256 digest from a remote container registry.  |
|        --repo-root string  |  [defaulted] The directory where the source git repository is available. Only used if --commit is used or defaulted in CI, see https://docs.kosli.com/integrations/ci_cd/#defaulted-kosli-command-flags-from-ci-variables . (default ".")  |
|        --repository string  |  Git repository. (defaulted in some CIs: https://docs.kosli.com/ci-defaults ).  |
|    -T, --trail string  |  The Kosli trail name.  |
|    -u, --user-data string  |  [optional] The path to a JSON file containing additional data you would like to attach to the attestation.  |


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

**report a Bitbucket pull request attestation about a pre-built docker artifact (kosli calculates the fingerprint)**

```shell
kosli attest pullrequest bitbucket yourDockerImageName 
	--artifact-type docker 
	--name yourAttestationName 
	--bitbucket-access-token yourBitbucketAccessToken 
	--bitbucket-workspace yourBitbucketWorkspace 
	--commit yourArtifactGitCommit 
	--repository yourBitbucketGitRepository 

```

**report a Bitbucket pull request attestation about a pre-built docker artifact (you provide the fingerprint)**

```shell
kosli attest pullrequest bitbucket 
	--fingerprint yourDockerImageFingerprint 
	--name yourAttestationName 
	--bitbucket-access-token yourBitbucketAccessToken 
	--bitbucket-workspace yourBitbucketWorkspace 
	--commit yourArtifactGitCommit 
	--repository yourBitbucketGitRepository 

```

**report a Bitbucket pull request attestation about a trail**

```shell
kosli attest pullrequest bitbucket 
	--name yourAttestationName 
	--bitbucket-access-token yourBitbucketAccessToken 
	--bitbucket-workspace yourBitbucketWorkspace 
	--commit yourArtifactGitCommit 
	--repository yourBitbucketGitRepository 

```

**report a Bitbucket pull request attestation about an artifact which has not been reported yet in a trail**

```shell
kosli attest pullrequest bitbucket 
	--name yourTemplateArtifactName.yourAttestationName 
	--bitbucket-access-token yourBitbucketAccessToken 
	--bitbucket-workspace yourBitbucketWorkspace 
	--commit yourArtifactGitCommit 
	--repository yourBitbucketGitRepository 

```

**report a Bitbucket pull request attestation about a trail with an attachment**

```shell
kosli attest pullrequest bitbucket 
	--name yourAttestationName 
	--bitbucket-access-token yourBitbucketAccessToken 
	--bitbucket-workspace yourBitbucketWorkspace 
	--commit yourArtifactGitCommit 
	--repository yourBitbucketGitRepository 
	--attachments=yourAttachmentPathName 

```

**fail if a pull request does not exist for your artifact**

```shell
kosli attest pullrequest bitbucket 
	--name yourTemplateArtifactName.yourAttestationName 
	--bitbucket-access-token yourBitbucketAccessToken 
	--bitbucket-workspace yourBitbucketWorkspace 
	--commit yourArtifactGitCommit 
	--repository yourBitbucketGitRepository 
	--assert
```

