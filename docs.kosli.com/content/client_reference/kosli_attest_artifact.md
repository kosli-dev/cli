---
title: "kosli attest artifact"
beta: false
deprecated: false
---

# kosli attest artifact

## Synopsis

Attest an artifact creation to a Kosli flow.  
The artifact SHA256 fingerprint is calculated (based on the `--artifact-type` flag and the artifact name/path argument) or can be provided directly (with the `--fingerprint` flag).
This command requires access to a git repo to associate the artifact to the git commit it is originating from. 
You can optionally redact some of the git commit data sent to Kosli using `--redact-commit-info`

```shell
kosli attest artifact {IMAGE-NAME | FILE-PATH | DIR-PATH} [flags]
```

## Flags
| Flag | Description |
| :--- | :--- |
|        --annotate stringToString  |  [optional] Annotate the attestation with data using key=value.  |
|    -t, --artifact-type string  |  [conditional] The type of the artifact to calculate its SHA256 fingerprint. One of: [docker, file, dir]. Only required if you don't specify '--fingerprint'.  |
|    -b, --build-url string  |  The url of CI pipeline that built the artifact. (defaulted in some CIs: https://docs.kosli.com/ci-defaults ).  |
|    -g, --commit string  |  [defaulted] The git commit from which the artifact was created. (defaulted in some CIs: https://docs.kosli.com/ci-defaults, otherwise defaults to HEAD ). (default "HEAD")  |
|    -u, --commit-url string  |  The url for the git commit that created the artifact. (defaulted in some CIs: https://docs.kosli.com/ci-defaults ).  |
|    -N, --display-name string  |  [optional] Artifact display name, if different from file, image or directory name.  |
|    -D, --dry-run  |  [optional] Run in dry-run mode. When enabled, no data is sent to Kosli and the CLI exits with 0 exit code regardless of any errors.  |
|    -x, --exclude strings  |  [optional] The comma separated list of directories and files to exclude from fingerprinting. Can take glob patterns. Only applicable for --artifact-type dir.  |
|        --external-fingerprint stringToString  |  [optional] A SHA256 fingerprint of an external attachment represented by --external-url. The format is label=fingerprint (labels cannot contain '.' or '='). This flag can be set multiple times. There must be an external url with a matching label for each external fingerprint.  |
|        --external-url stringToString  |  [optional] Add labeled reference URL for an external resource. The format is label=url (labels cannot contain '.' or '='). This flag can be set multiple times. If the resource is a file or dir, you can optionally add its fingerprint via --external-fingerprint  |
|    -F, --fingerprint string  |  [conditional] The SHA256 fingerprint of the artifact. Only required if you don't specify '--artifact-type'.  |
|    -f, --flow string  |  The Kosli flow name.  |
|    -h, --help  |  help for artifact  |
|    -n, --name string  |  The name of the artifact in the yml template file.  |
|        --redact-commit-info strings  |  [optional] The list of commit info to be redacted before sending to Kosli. Allowed values are one or more of [author, message, branch].  |
|        --registry-password string  |  [conditional] The docker registry password or access token. Only required if you want to read docker image SHA256 digest from a remote docker registry.  |
|        --registry-provider string  |  [conditional] The docker registry provider or url. Only required if you want to read docker image SHA256 digest from a remote docker registry.  |
|        --registry-username string  |  [conditional] The docker registry username. Only required if you want to read docker image SHA256 digest from a remote docker registry.  |
|        --repo-root string  |  [defaulted] The directory where the source git repository is available. (default ".")  |
|    -T, --trail string  |  The Kosli trail name.  |


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

{{< tabs "live-examples" "col-no-wrap" >}}{{< tab "GitHub" >}}View an example of the `kosli attest artifact` command in GitHub.

In [this YAML file](https://app.kosli.com/api/v2/livedocs/cyber-dojo/yaml?ci=github&command=kosli+attest+artifact), which created [this Kosli Event](https://app.kosli.com/api/v2/livedocs/cyber-dojo/event?ci=github&command=kosli+attest+artifact).{{< /tab >}}{{< tab "GitLab" >}}View an example of the `kosli attest artifact` command in GitLab.

In [this YAML file](https://app.kosli.com/api/v2/livedocs/cyber-dojo/yaml?ci=gitlab&command=kosli+attest+artifact), which created [this Kosli Event](https://app.kosli.com/api/v2/livedocs/cyber-dojo/event?ci=gitlab&command=kosli+attest+artifact).{{< /tab >}}{{< /tabs >}}

## Examples Use Cases

**Attest that a file type artifact has been created, and let Kosli calculate its fingerprint**

```shell
kosli attest artifact FILE.tgz \
	--artifact-type file \
	--build-url https://exampleci.com \
	--commit-url https://github.com/YourOrg/YourProject/commit/yourCommitShaThatThisArtifactWasBuiltFrom \
	--commit yourCommitShaThatThisArtifactWasBuiltFrom \
	--flow yourFlowName \
	--trail yourTrailName \
	--name yourTemplateArtifactName \
	--api-token yourApiToken \
	--org yourOrgName


```

**Attest that an artifact has been created and provide its fingerprint (sha256)**

```shell
kosli attest artifact ANOTHER_FILE.txt \
	--build-url https://exampleci.com \
	--commit-url https://github.com/YourOrg/YourProject/commit/yourCommitShaThatThisArtifactWasBuiltFrom \
	--commit yourCommitShaThatThisArtifactWasBuiltFrom \
	--flow yourFlowName \
	--trail yourTrailName \
	--fingerprint yourArtifactFingerprint \
	--name yourTemplateArtifactName \
	--api-token yourApiToken \
	--org yourOrgName

```

**Attest that an artifact has been created and provide external attachments**

```shell
kosli attest artifact ANOTHER_FILE.txt \
	--build-url https://exampleci.com \
	--commit-url https://github.com/YourOrg/YourProject/commit/yourCommitShaThatThisArtifactWasBuiltFrom \
	--commit yourCommitShaThatThisArtifactWasBuiltFrom \
	--flow yourFlowName \
	--trail yourTrailName \
	--fingerprint yourArtifactFingerprint \
	--external-url label=https://example.com/attachment \
	--external-fingerprint label=yourExternalAttachmentFingerprint \
	--name yourTemplateArtifactName \
	--api-token yourApiToken \
	--org yourOrgName
```

