---
title: "kosli attest sonar"
beta: false
deprecated: false
---

# kosli attest sonar

## Synopsis

Report a sonarcloud or sonarqube attestation to an artifact or a trail in a Kosli flow.  

The attestation can be bound to a trail using the trail name.

If the attestation is for an artifact, the attestation can be bound to the artifact using one of two ways:
- using the artifact's SHA256 fingerprint which is calculated (based on the `--artifact-type` flag and the artifact name/path argument) or can be provided directly (with the `--fingerprint` flag).
- using the artifact's name in the flow yaml template and the git commit from which the artifact is/will be created. Useful when reporting an attestation before creating/reporting the artifact.

```shell
kosli attest sonar [IMAGE-NAME | FILE-PATH | DIR-PATH] [flags]
```

## Flags
| Flag | Description |
| :--- | :--- |
|        --annotate stringToString  |  [optional] Annotate the attestation with data using key=value.  |
|    -t, --artifact-type string  |  [conditional] The type of the artifact to calculate its SHA256 fingerprint. One of: [docker, file, dir]. Only required if you don't specify '--fingerprint'.  |
|        --attachments strings  |  [optional] The comma-separated list of paths of attachments for the reported attestation. Attachments can be files or directories. All attachments are compressed and uploaded to Kosli's evidence vault.  |
|        --branch-name string  |  CI Branch Name  |
|    -g, --commit string  |  [conditional] The git commit for which the attestation is associated to. Becomes required when reporting an attestation for an artifact before reporting it to Kosli. (defaulted in some CIs: https://docs.kosli.com/ci-defaults ).  |
|        --description string  |  [optional] attestation description  |
|    -D, --dry-run  |  [optional] Run in dry-run mode. When enabled, no data is sent to Kosli and the CLI exits with 0 exit code regardless of any errors.  |
|    -x, --exclude strings  |  [optional] The comma separated list of directories and files to exclude from fingerprinting. Can take glob patterns. Only applicable for --artifact-type dir.  |
|        --external-fingerprint stringToString  |  [optional] A SHA256 fingerprint of an external attachment represented by --external-url. The format is label=fingerprint (labels cannot contain '.' or '='). This flag can be set multiple times. There must be an external url with a matching label for each external fingerprint.  |
|        --external-url stringToString  |  [optional] Add labeled reference URL for an external resource. The format is label=url (labels cannot contain '.' or '='). This flag can be set multiple times. If the resource is a file or dir, you can optionally add its fingerprint via --external-fingerprint  |
|    -F, --fingerprint string  |  [conditional] The SHA256 fingerprint of the artifact to attach the attestation to. Only required if the attestation is for an artifact and --artifact-type and artifact name/path are not used.  |
|    -f, --flow string  |  The Kosli flow name.  |
|    -h, --help  |  help for sonar  |
|    -n, --name string  |  The name of the attestation as declared in the flow or trail yaml template.  |
|    -o, --origin-url string  |  [optional] The url pointing to where the attestation came from or is related. (defaulted to the CI url in some CIs: https://docs.kosli.com/ci-defaults ).  |
|        --pull-request-id string  |  Pull Request ID  |
|        --redact-commit-info strings  |  [optional] The list of commit info to be redacted before sending to Kosli. Allowed values are one or more of [author, message, branch].  |
|        --registry-password string  |  [conditional] The docker registry password or access token. Only required if you want to read docker image SHA256 digest from a remote docker registry.  |
|        --registry-provider string  |  [conditional] The docker registry provider or url. Only required if you want to read docker image SHA256 digest from a remote docker registry.  |
|        --registry-username string  |  [conditional] The docker registry username. Only required if you want to read docker image SHA256 digest from a remote docker registry.  |
|        --repo-root string  |  [defaulted] The directory where the source git repository is available. Only used if --commit is used. (default ".")  |
|        --sonar-api-token string  |  SonarCloud API token  |
|        --sonar-project-key string  |  SonarCloud project key  |
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


## Live Examples in different CI systems

{{< tabs "live-examples" "col-no-wrap" >}}{{< tab "GitHub" >}}View an example of the `kosli attest sonar` command in GitHub.

In [this YAML file](https://app.kosli.com/api/v2/livedocs/cyber-dojo/yaml?ci=github&command=kosli+attest+sonar), which created [this Kosli Event](https://app.kosli.com/api/v2/livedocs/cyber-dojo/event?ci=github&command=kosli+attest+sonar).{{< /tab >}}{{< /tabs >}}

## Examples Use Cases

**report a sonarcloud attestation about a trail**

```shell
kosli attest sonar \
	--name yourAttestationName \
	--flow yourFlowName \
	--trail yourTrailName \
	--sonar-project-key yourSonarProjectKey \
	--sonar-api-token yourSonarAPIToken \
	--api-token yourAPIToken \
	--org yourOrgName \

```

**report a sonarcloud attestation for a specific branch about a trail**

```shell
kosli attest sonar \
	--name yourAttestationName \
	--flow yourFlowName \
	--trail yourTrailName \
	--sonar-project-key yourSonarProjectKey \
	--sonar-api-token yourSonarAPIToken \
	--branch-name yourBranchName \
	--api-token yourAPIToken \
	--org yourOrgName \

```

**report a sonarcloud attestation for a pull-request about a trail**

```shell
kosli attest sonar \
	--name yourAttestationName \
	--flow yourFlowName \
	--trail yourTrailName \
	--sonar-project-key yourSonarProjectKey \
	--sonar-api-token yourSonarAPIToken \
	--pull-request-id yourPullRequestID \
	--api-token yourAPIToken \
	--org yourOrgName \

```

**report a sonarcloud attestation about a trail with an attachment**

```shell
kosli attest sonar \
	--name yourAttestationName \
	--flow yourFlowName \
	--trail yourTrailName \
	--sonar-project-key yourSonarProjectKey \
	--sonar-api-token yourSonarAPIToken \
	--attachment yourAttachmentPath \
	--api-token yourAPIToken \
	--org yourOrgName
```

