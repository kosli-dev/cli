---
title: "kosli attest pullrequest azure"
beta: false
deprecated: false
---

# kosli attest pullrequest azure

## Synopsis

Report an Azure Devops pull request attestation to an artifact or a trail in a Kosli flow.  
It checks if a pull request exists for the artifact (based on its git commit) and reports the pull-request attestation to the artifact in Kosli.
The artifact SHA256 fingerprint is calculated (based on --artifact-type flag) or alternatively it can be provided directly (with --fingerprint flag).

```shell
kosli attest pullrequest azure [IMAGE-NAME | FILE-PATH | DIR-PATH] [flags]
```

## Flags
| Flag | Description |
| :--- | :--- |
|    -t, --artifact-type string  |  [conditional] The type of the artifact to calculate its SHA256 fingerprint. One of: [docker, file, dir]. Only required if you don't specify '--fingerprint'.  |
|        --assert  |  [optional] Exit with non-zero code if no pull requests found for the given commit.  |
|        --attachments strings  |  [optional] The comma-separated list of paths of attachments for the reported attestation. Attachments can be files or directories. All attachments are compressed and uploaded to Kosli's evidence vault.  |
|        --azure-org-url string  |  Azure organization url. E.g. "https://dev.azure.com/myOrg" (defaulted if you are running in Azure Devops pipelines: https://docs.kosli.com/ci-defaults ).  |
|        --azure-token string  |  Azure Personal Access token.  |
|    -g, --commit string  |  [optional] The git commit associated to the attestation. (defaulted in some CIs: https://docs.kosli.com/ci-defaults ).  |
|        --description string  |  [optional] attestation description  |
|    -D, --dry-run  |  [optional] Run in dry-run mode. When enabled, no data is sent to Kosli and the CLI exits with 0 exit code regardless of any errors.  |
|    -x, --exclude strings  |  [optional] The comma separated list of directories and files to exclude from fingerprinting. Can take glob patterns. Only applicable for --artifact-type dir.  |
|        --external-fingerprint stringToString  |  [optional] A SHA256 fingerprint of an external attachment represented by --external-url. The format is label=fingerprint (labels cannot contain '.' or '='). This flag can be set multiple times. There must be an external url with a matching label for each external fingerprint.  |
|        --external-url stringToString  |  [optional] Add labeled reference URL for an external resource. The format is label=url (labels cannot contain '.' or '='). This flag can be set multiple times. If the resource is a file or dir, you can optionally add its fingerprint via --external-fingerprint  |
|    -F, --fingerprint string  |  [optional] The SHA256 fingerprint of the artifact to attach the attestation to.  |
|    -f, --flow string  |  The Kosli flow name.  |
|    -h, --help  |  help for azure  |
|    -n, --name string  |  The name of the attestation as declared in the flow or trail yaml template.  |
|    -o, --origin-url string  |  [optional] The url pointing to where the attestation came from or is related. (defaulted to the CI url in some CIs: https://docs.kosli.com/ci-defaults ).  |
|        --project string  |  Azure project.(defaulted if you are running in Azure Devops pipelines: https://docs.kosli.com/ci-defaults ).  |
|        --registry-password string  |  [conditional] The docker registry password or access token. Only required if you want to read docker image SHA256 digest from a remote docker registry.  |
|        --registry-provider string  |  [conditional] The docker registry provider or url. Only required if you want to read docker image SHA256 digest from a remote docker registry.  |
|        --registry-username string  |  [conditional] The docker registry username. Only required if you want to read docker image SHA256 digest from a remote docker registry.  |
|        --repo-root string  |  [defaulted] The directory where the source git repository is available. Only used if --commit is used. (default ".")  |
|        --repository string  |  Git repository. (defaulted in some CIs: https://docs.kosli.com/ci-defaults ).  |
|    -T, --trail string  |  The Kosli trail name.  |
|    -u, --user-data string  |  [optional] The path to a JSON file containing additional data you would like to attach to the attestation.  |


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

# report an Azure Devops pull request attestation about a pre-built docker artifact (kosli calculates the fingerprint):
kosli attest pullrequest azure yourDockerImageName \
	--artifact-type docker \
	--name yourAttestationName \
	--flow yourFlowName \
	--trail yourTrailName \
	--azure-org-url https://dev.azure.com/myOrg \
	--project yourAzureDevOpsProject \
	--azure-token yourAzureToken \
	--commit yourGitCommitSha1 \
	--repository yourAzureGitRepository \
	--api-token yourAPIToken \
	--org yourOrgName

# report an Azure Devops pull request attestation about a pre-built docker artifact (you provide the fingerprint):
kosli attest pullrequest azure \
	--fingerprint yourDockerImageFingerprint \
	--name yourAttestationName \
	--flow yourFlowName \
	--trail yourTrailName \
	--azure-org-url https://dev.azure.com/myOrg \
	--project yourAzureDevOpsProject \
	--azure-token yourAzureToken \
	--commit yourGitCommitSha1 \
	--repository yourAzureGitRepository \
	--api-token yourAPIToken \
	--org yourOrgName

# report an Azure Devops pull request attestation about a trail:
kosli attest pullrequest azure \
	--name yourAttestationName \
	--flow yourFlowName \
	--trail yourTrailName \
	--azure-org-url https://dev.azure.com/myOrg \
	--project yourAzureDevOpsProject \
	--azure-token yourAzureToken \
	--commit yourGitCommitSha1 \
	--repository yourAzureGitRepository \
	--api-token yourAPIToken \
	--org yourOrgName

# report an Azure Devops pull request attestation about an artifact which has not been reported yet in a trail:
kosli attest pullrequest azure \
	--name yourTemplateArtifactName.yourAttestationName \
	--flow yourFlowName \
	--trail yourTrailName \
	--azure-org-url https://dev.azure.com/myOrg \
	--project yourAzureDevOpsProject \
	--azure-token yourAzureToken \
	--commit yourGitCommitSha1 \
	--repository yourAzureGitRepository \
	--api-token yourAPIToken \
	--org yourOrgName

# report an Azure Devops pull request attestation about a trail with an attachment:
kosli attest pullrequest azure \
	--name yourAttestationName \
	--flow yourFlowName \
	--trail yourTrailName \
	--azure-org-url https://dev.azure.com/myOrg \
	--project yourAzureDevOpsProject \
	--azure-token yourAzureToken \
	--commit yourGitCommitSha1 \
	--repository yourAzureGitRepository \
	--attachments=yourAttachmentPathName \
	--api-token yourAPIToken \
	--org yourOrgName

# fail if a pull request does not exist for your artifact
kosli attest pullrequest azure \
	--name yourTemplateArtifactName.yourAttestationName \
	--flow yourFlowName \
	--trail yourTrailName \
	--azure-org-url https://dev.azure.com/myOrg \
	--project yourAzureDevOpsProject \
	--azure-token yourAzureToken \
	--commit yourGitCommitSha1 \
	--repository yourAzureGitRepository \
	--api-token yourAPIToken \
	--org yourOrgName \
	--assert

```
