---
title: "kosli attest pullrequest github"
beta: true
deprecated: false
---

# kosli attest pullrequest github

{{< hint warning >}}**kosli attest pullrequest github** is a beta feature. Beta features provide early access to product functionality.  These features may change between releases without warning, or can be removed in a future release.
Please contact us to enable this feature for your organization.{{< /hint >}}
## Synopsis

Report a Github pull request attestation to an artifact or a trail in a Kosli flow.  
It checks if a pull request exists for the artifact (based on its git commit) and reports the pull-request evidence to the artifact in Kosli.
The artifact SHA256 fingerprint is calculated (based on --artifact-type flag) or alternatively it can be provided directly (with --fingerprint flag).

```shell
kosli attest pullrequest github [IMAGE-NAME | FILE-PATH | DIR-PATH] [flags]
```

## Flags
| Flag | Description |
| :--- | :--- |
|    -t, --artifact-type string  |  [conditional] The type of the artifact to calculate its SHA256 fingerprint. One of: [docker, file, dir]. Only required if you don't specify '--fingerprint'.  |
|        --assert  |  [optional] Exit with non-zero code if no pull requests found for the given commit.  |
|    -g, --commit string  |  [optional] The git commit associated to the attestation. (defaulted in some CIs: https://docs.kosli.com/ci-defaults ).  |
|    -D, --dry-run  |  [optional] Run in dry-run mode. When enabled, no data is sent to Kosli and the CLI exits with 0 exit code regardless of any errors.  |
|        --evidence-fingerprint string  |  [optional] The SHA256 fingerprint of the evidence file or dir.  |
|    -e, --evidence-paths strings  |  [optional] The comma-separated list of paths containing supporting proof for the reported evidence. Paths can be for files or directories. All provided proofs will be uploaded to Kosli's evidence vault.  |
|        --evidence-url string  |  [optional] The external URL where the evidence file or dir is stored.  |
|    -x, --exclude strings  |  [optional] The comma separated list of directories and files to exclude from fingerprinting. Only applicable for --artifact-type dir.  |
|    -F, --fingerprint string  |  [optional] The SHA256 fingerprint of the artifact to attach the attestation to.  |
|    -f, --flow string  |  The Kosli flow name.  |
|        --github-base-url string  |  [optional] GitHub base URL (only needed for GitHub Enterprise installations).  |
|        --github-org string  |  Github organization. (defaulted if you are running in GitHub Actions: https://docs.kosli.com/ci-defaults ).  |
|        --github-token string  |  Github token.  |
|    -h, --help  |  help for github  |
|    -n, --name string  |  The name of the attestation as declared in the flow or trail yaml template.  |
|        --registry-password string  |  [conditional] The docker registry password or access token. Only required if you want to read docker image SHA256 digest from a remote docker registry.  |
|        --registry-provider string  |  [conditional] The docker registry provider or url. Only required if you want to read docker image SHA256 digest from a remote docker registry.  |
|        --registry-username string  |  [conditional] The docker registry username. Only required if you want to read docker image SHA256 digest from a remote docker registry.  |
|        --repo-root string  |  [defaulted] The directory where the source git repository is available. Only used if --commit is used. (default ".")  |
|        --repository string  |  Git repository. (defaulted in some CIs: https://docs.kosli.com/ci-defaults ).  |
|    -T, --trail string  |  The Kosli trail name.  |
|    -b, --url string  |  The url pointing to where the attestation came from or is related. (defaulted to the CI url in some CIs: https://docs.kosli.com/ci-defaults ).  |
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

# report a Github pull request attestation about a pre-built docker artifact (kosli calculates the fingerprint):
kosli attest pullrequest github yourDockerImageName \
	--artifact-type docker \
	--name yourAttestationName \
	--flow yourFlowName \
	--trail yourTrailName \
	--github-token yourGithubToken \
	--github-org yourGithubOrg \
	--commit yourArtifactGitCommit \
	--repository yourGithubGitRepository \
	--api-token yourAPIToken \
	--org yourOrgName

# report a Github pull request attestation about a pre-built docker artifact (you provide the fingerprint):
kosli attest pullrequest github \
	--fingerprint yourDockerImageFingerprint \
	--name yourAttestationName \
	--flow yourFlowName \
	--trail yourTrailName \
	--github-token yourGithubToken \
	--github-org yourGithubOrg \
	--commit yourArtifactGitCommit \
	--repository yourGithubGitRepository \
	--api-token yourAPIToken \
	--org yourOrgName

# report a Github pull request attestation about a trail:
kosli attest pullrequest github \
	--name yourAttestationName \
	--flow yourFlowName \
	--trail yourTrailName \
	--github-token yourGithubToken \
	--github-org yourGithubOrg \
	--commit yourArtifactGitCommit \
	--repository yourGithubGitRepository \
	--api-token yourAPIToken \
	--org yourOrgName

# report a Github pull request attestation about an artifact which has not been reported yet in a trail:
kosli attest pullrequest github \
	--name yourTemplateArtifactName.yourAttestationName \
	--flow yourFlowName \
	--trail yourTrailName \
	--github-token yourGithubToken \
	--github-org yourGithubOrg \
	--commit yourArtifactGitCommit \
	--repository yourGithubGitRepository \
	--api-token yourAPIToken \
	--org yourOrgName

# report a Github pull request attestation about a trail with an evidence file:
kosli attest pullrequest github \
	--name yourAttestationName \
	--flow yourFlowName \
	--trail yourTrailName \
	--github-token yourGithubToken \
	--github-org yourGithubOrg \
	--commit yourArtifactGitCommit \
	--repository yourGithubGitRepository \
	--evidence-paths=yourEvidencePathName \
	--api-token yourAPIToken \
	--org yourOrgName

# fail if a pull request does not exist for your artifact
kosli attest pullrequest github \
	--name yourTemplateArtifactName.yourAttestationName \
	--flow yourFlowName \
	--trail yourTrailName \
	--github-token yourGithubToken \
	--github-org yourGithubOrg \
	--commit yourArtifactGitCommit \
	--repository yourGithubGitRepository \
	--api-token yourAPIToken \
	--org yourOrgName \
	--assert

```
