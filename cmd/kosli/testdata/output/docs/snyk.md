---
title: "snyk"
beta: false
deprecated: false
---

# snyk

## Synopsis

Report a snyk attestation to an artifact or a trail in a Kosli flow.  
Only SARIF snyk output is accepted. 
Snyk output can be for "snyk code test", "snyk container test", or "snyk iac test".

The `--scan-results` .json file is analyzed and a summary of the scan results are reported to Kosli.

By default, the `--scan-results` .json file is also uploaded to Kosli's evidence vault.
You can disable that by setting `--upload-results=false`


The attestation can be bound to a trail using the trail name.

If the attestation is for an artifact, the attestation can be bound to the artifact using one of two ways:
- using the artifact's SHA256 fingerprint which is calculated (based on the `--artifact-type` flag and the artifact name/path argument) or can be provided directly (with the `--fingerprint` flag).
- using the artifact's name in the flow yaml template and the git commit from which the artifact is/will be created. Useful when reporting an attestation before creating/reporting the artifact.

You can optionally associate the attestation to a git commit using `--commit` (requires access to a git repo). And you  
can optionally redact some of the git commit data sent to Kosli using `--redact-commit-info`. 
Note that when the attestation is reported for an artifact that does not yet exist in Kosli, `--commit` becomes required to facilitate 
binding the attestation to the right artifact.

```shell
snyk [IMAGE-NAME | FILE-PATH | DIR-PATH] [flags]
```

## Flags
| Flag | Description |
| :--- | :--- |
|        --annotate stringToString  |  [optional] Annotate the attestation with data using key=value.  |
|    -t, --artifact-type string  |  The type of the artifact to calculate its SHA256 fingerprint. One of: [oci, docker, file, dir]. Only required if you want Kosli to calculate the fingerprint for you (i.e. when you don't specify '--fingerprint' on commands that allow it).  |
|        --attachments strings  |  [optional] The comma-separated list of paths of attachments for the reported attestation. Attachments can be files or directories. All attachments are compressed and uploaded to Kosli's evidence vault.  |
|    -g, --commit string  |  [conditional] The git commit for which the attestation is associated to. Becomes required when reporting an attestation for an artifact before reporting it to Kosli. (defaulted in some CIs: https://docs.kosli.com/ci-defaults ).  |
|        --description string  |  [optional] attestation description  |
|    -D, --dry-run  |  [optional] Run in dry-run mode. When enabled, no data is sent to Kosli and the CLI exits with 0 exit code regardless of any errors.  |
|    -x, --exclude strings  |  [optional] The comma separated list of directories and files to exclude from fingerprinting. Can take glob patterns. Only applicable for --artifact-type dir.  |
|        --external-fingerprint stringToString  |  [optional] A SHA256 fingerprint of an external attachment represented by --external-url. The format is label=fingerprint (labels cannot contain '.' or '='). This flag can be set multiple times. There must be an external url with a matching label for each external fingerprint.  |
|        --external-url stringToString  |  [optional] Add labeled reference URL for an external resource. The format is label=url (labels cannot contain '.' or '='). This flag can be set multiple times. If the resource is a file or dir, you can optionally add its fingerprint via --external-fingerprint  |
|    -F, --fingerprint string  |  [conditional] The SHA256 fingerprint of the artifact to attach the attestation to. Only required if the attestation is for an artifact and --artifact-type and artifact name/path are not used.  |
|    -f, --flow string  |  The Kosli flow name.  |
|    -h, --help  |  help for snyk  |
|    -n, --name string  |  The name of the attestation as declared in the flow or trail yaml template.  |
|    -o, --origin-url string  |  [optional] The url pointing to where the attestation came from or is related. (defaulted to the CI url in some CIs: https://docs.kosli.com/ci-defaults ).  |
|        --redact-commit-info strings  |  [optional] The list of commit info to be redacted before sending to Kosli. Allowed values are one or more of [author, message, branch].  |
|        --registry-password string  |  [conditional] The container registry password or access token. Only required if you want to read container image SHA256 digest from a remote container registry.  |
|        --registry-username string  |  [conditional] The container registry username. Only required if you want to read container image SHA256 digest from a remote container registry.  |
|        --repo-root string  |  [defaulted] The directory where the source git repository is available. Only used if --commit is used. (default ".")  |
|    -R, --scan-results string  |  The path to Snyk scan SARIF results file from 'snyk test' and 'snyk container test'. By default, the Snyk results will be uploaded to Kosli's evidence vault.  |
|    -T, --trail string  |  The Kosli trail name.  |
|        --upload-results  |  [defaulted] Whether to upload the provided Snyk results file as an attachment to Kosli or not. (default true)  |
|    -u, --user-data string  |  [optional] The path to a JSON file containing additional data you would like to attach to the attestation.  |


## Examples Use Cases

**report a snyk attestation about a pre-built docker artifact (kosli calculates the fingerprint)**

```shell
kosli attest snyk yourDockerImageName \
	--artifact-type docker \
	--name yourAttestationName \
	--flow yourFlowName \
	--trail yourTrailName \
	--scan-results yourSnykSARIFScanResults \
	--api-token yourAPIToken \
	--org yourOrgName

```

**report a snyk attestation about a pre-built docker artifact (you provide the fingerprint)**

```shell
kosli attest snyk \
	--fingerprint yourDockerImageFingerprint \
	--name yourAttestationName \
	--flow yourFlowName \
	--trail yourTrailName \
	--scan-results yourSnykSARIFScanResults \
	--api-token yourAPIToken \
	--org yourOrgName

```

**report a snyk attestation about a trail**

```shell
kosli attest snyk \
	--name yourAttestationName \
	--flow yourFlowName \
	--trail yourTrailName \
	--scan-results yourSnykSARIFScanResults \
	--api-token yourAPIToken \
	--org yourOrgName

```

**report a snyk attestation about an artifact which has not been reported yet in a trail**

```shell
kosli attest snyk \
	--name yourTemplateArtifactName.yourAttestationName \
	--flow yourFlowName \
	--trail yourTrailName \
	--commit yourArtifactGitCommit \
	--scan-results yourSnykSARIFScanResults \
	--api-token yourAPIToken \
	--org yourOrgName

```

**report a snyk attestation about a trail with an attachment**

```shell
kosli attest snyk \
	--name yourAttestationName \
	--flow yourFlowName \
	--trail yourTrailName \
	--scan-results yourSnykSARIFScanResults \
	--attachments yourEvidencePathName \
	--api-token yourAPIToken \
	--org yourOrgName

```

**report a snyk attestation about a trail without uploading the snyk results file**

```shell
kosli attest snyk \
	--name yourAttestationName \
	--flow yourFlowName \
	--trail yourTrailName \
	--scan-results yourSnykSARIFScanResults \
	--upload-results=false \
	--api-token yourAPIToken \
	--org yourOrgName
```

