---
title: "kosli attest jira"
beta: false
deprecated: false
summary: "Report a jira attestation to an artifact or a trail in a Kosli flow.  "
---

# kosli attest jira

## Synopsis

```shell
kosli attest jira [IMAGE-NAME | FILE-PATH | DIR-PATH] [flags]
```

Report a jira attestation to an artifact or a trail in a Kosli flow.  
Parses the given commit's message, current branch name or the content of the `--jira-secondary-source`
argument for Jira issue references of the form:  
'at least 2 characters long, starting with an uppercase letter project key followed by
dash and one or more digits'.

If you want to restrict the Jira issue matching to a specific project, use the
`--jira-project-key` flag to specify your own project key. You can specify multiple project keys if needed.

If the `--ignore-branch-match` is set, the branch name is not parsed for a match.

The found issue references will be checked against Jira to confirm their existence.
The attestation is reported in all cases, and its compliance status depends on referencing
existing Jira issues.  
If you have wrong Jira credentials or wrong Jira-base-url it will be reported as non existing Jira issue.
This is because Jira returns same 404 error code in all cases.

The `--jira-issue-fields` can be used to include fields from the jira issue. By default no fields
are included. `*all` will give all fields. Using `--jira-issue-fields "*all" --dry-run` will give you
the complete list so you can select the once you need. The issue fields uses the jira API that is documented here:
https://developer.atlassian.com/cloud/jira/platform/rest/v2/api-group-issues/#api-rest-api-2-issue-issueidorkey-get-request


The attestation can be bound to a *trail* using the trail name.  
The attestation can be bound to an *artifact* in two ways:
- using the artifact's SHA256 fingerprint which is calculated (based on the `--artifact-type` flag and the artifact name/path argument) or can be provided directly (with the `--fingerprint` flag).
- using the artifact's name in the flow yaml template and the git commit from which the artifact is/will be created. Useful when reporting an attestation before creating/reporting the artifact.

You can optionally associate the attestation to a git commit using `--commit` (requires access to a git repo).
You can optionally redact some of the git commit data sent to Kosli using `--redact-commit-info`.
Note that when the attestation is reported for an artifact that does not yet exist in Kosli, `--commit` is required to facilitate
binding the attestation to the right artifact.

## Flags
| Flag | Description |
| :--- | :--- |
|        --annotate stringToString  |  [optional] Annotate the attestation with data using key=value.  |
|    -t, --artifact-type string  |  The type of the artifact to calculate its SHA256 fingerprint. One of: [oci, docker, file, dir]. Only required if you want Kosli to calculate the fingerprint for you (i.e. when you don't specify '--fingerprint' on commands that allow it).  |
|        --assert  |  [optional] Exit with non-zero code if the attestation is non-compliant  |
|        --attachments strings  |  [optional] The comma-separated list of paths of attachments for the reported attestation. Attachments can be files or directories. All attachments are compressed and uploaded to Kosli's evidence vault.  |
|    -g, --commit string  |  [conditional] The git commit for which the attestation is associated to. Becomes required when reporting an attestation for an artifact before reporting it to Kosli. (defaulted in some CIs: https://docs.kosli.com/ci-defaults ).  |
|        --description string  |  [optional] attestation description  |
|    -D, --dry-run  |  [optional] Run in dry-run mode. When enabled, no data is sent to Kosli and the CLI exits with 0 exit code regardless of any errors.  |
|    -x, --exclude strings  |  [optional] The comma separated list of directories and files to exclude from fingerprinting. Can take glob patterns. Only applicable for --artifact-type dir.  |
|        --external-fingerprint stringToString  |  [optional] A SHA256 fingerprint of an external attachment represented by --external-url. The format is label=fingerprint (labels cannot contain '.' or '='). This flag can be set multiple times. There must be an external url with a matching label for each external fingerprint.  |
|        --external-url stringToString  |  [optional] Add labeled reference URL for an external resource. The format is label=url (labels cannot contain '.' or '='). This flag can be set multiple times. If the resource is a file or dir, you can optionally add its fingerprint via --external-fingerprint  |
|    -F, --fingerprint string  |  [conditional] The SHA256 fingerprint of the artifact to attach the attestation to. Only required if the attestation is for an artifact and --artifact-type and artifact name/path are not used.  |
|    -f, --flow string  |  The Kosli flow name.  |
|    -h, --help  |  help for jira  |
|        --ignore-branch-match  |  Ignore branch name when searching for Jira ticket reference.  |
|        --jira-api-token string  |  Jira API token (for Jira Cloud)  |
|        --jira-base-url string  |  The base url for the jira project, e.g. 'https://kosli.atlassian.net'  |
|        --jira-issue-fields string  |  [optional] The comma separated list of fields to include from the Jira issue. Default no fields are included. '*all' will give all fields.  |
|        --jira-pat string  |  Jira personal access token (for self-hosted Jira)  |
|        --jira-project-key strings  |  [optional] Jira project key to match against. Can be repeated. Defaults to matching any jira project key.  |
|        --jira-secondary-source string  |  [optional] An optional string to search for Jira ticket reference, e.g. '--jira-secondary-source ${{ github.head_ref }}'  |
|        --jira-username string  |  Jira username (for Jira Cloud)  |
|    -n, --name string  |  The name of the attestation as declared in the flow or trail yaml template.  |
|    -o, --origin-url string  |  [optional] The url pointing to where the attestation came from or is related. (defaulted to the CI url in some CIs: https://docs.kosli.com/integrations/ci_cd/#defaulted-kosli-command-flags-from-ci-variables ).  |
|        --redact-commit-info strings  |  [optional] The list of commit info to be redacted before sending to Kosli. Allowed values are one or more of [author, message, branch].  |
|        --registry-password string  |  [conditional] The container registry password or access token. Only required if you want to read container image SHA256 digest from a remote container registry.  |
|        --registry-username string  |  [conditional] The container registry username. Only required if you want to read container image SHA256 digest from a remote container registry.  |
|        --repo-root string  |  [defaulted] The directory where the source git repository is available. Only used if --commit is used or defaulted in CI, see https://docs.kosli.com/integrations/ci_cd/#defaulted-kosli-command-flags-from-ci-variables . (default ".")  |
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

**report a jira attestation about a pre-built docker artifact (kosli calculates the fingerprint)**

```shell
kosli attest jira yourDockerImageName 
	--artifact-type docker 
	--name yourAttestationName 
	--jira-base-url https://kosli.atlassian.net 
	--jira-username user@domain.com 
	--jira-api-token yourJiraAPIToken 

```

**report a jira attestation about a pre-built docker artifact (you provide the fingerprint)**

```shell
kosli attest jira 
	--fingerprint yourDockerImageFingerprint 
	--name yourAttestationName 
	--jira-base-url https://kosli.atlassian.net 
	--jira-username user@domain.com 
	--jira-api-token yourJiraAPIToken 

```

**report a jira attestation about a trail**

```shell
kosli attest jira 
	--name yourAttestationName 
	--jira-base-url https://kosli.atlassian.net 
	--jira-username user@domain.com 
	--jira-api-token yourJiraAPIToken 

```

**report a jira attestation matching a specific jira project key**

```shell
kosli attest jira 
	--name yourAttestationName 
	--jira-base-url https://kosli.atlassian.net 
	--jira-username user@domain.com 
	--jira-api-token yourJiraAPIToken 
	--jira-project-key ABC 

```

**report a jira attestation about a trail and include jira issue summary, description and creator**

```shell
kosli attest jira 
	--name yourAttestationName 
	--jira-base-url https://kosli.atlassian.net 
	--jira-username user@domain.com 
	--jira-api-token yourJiraAPIToken 
	--jira-issue-fields "summary,description,creator"

```

**report a jira attestation about an artifact which has not been reported yet in a trail**

```shell
kosli attest jira 
	--name yourTemplateArtifactName.yourAttestationName 
	--commit yourArtifactGitCommit 
	--jira-base-url https://kosli.atlassian.net 
	--jira-username user@domain.com 
	--jira-api-token yourJiraAPIToken 

```

**report a jira attestation about a trail with an attachment**

```shell
kosli attest jira 
	--name yourAttestationName 
	--jira-base-url https://kosli.atlassian.net 
	--jira-username user@domain.com 
	--jira-api-token yourJiraAPIToken 
	--attachments yourAttachmentPathName 

```

**fail if no issue reference is found, or the issue is not found in your jira instance**

```shell
kosli attest jira 
	--name yourAttestationName 
	--jira-base-url https://kosli.atlassian.net 
	--jira-username user@domain.com 
	--jira-api-token yourJiraAPIToken 
	--assert

```

**get jira reference from original branch name in a GitHub Pull Request merge job**

```shell
kosli attest jira 
	--name yourAttestationName 
	--jira-secondary-source ${{ github.head_ref }} 
	--jira-base-url https://kosli.atlassian.net 
	--jira-username user@domain.com 
	--jira-api-token yourJiraAPIToken 
```

