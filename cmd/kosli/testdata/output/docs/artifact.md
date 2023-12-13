---
title: "artifact"
beta: false
---

# artifact

## Synopsis

Report an artifact creation to a Kosli flow.  
The artifact SHA256 fingerprint is calculated (based on --artifact-type flag) or alternatively it can be provided directly (with --fingerprint flag).

```shell
artifact {IMAGE-NAME | FILE-PATH | DIR-PATH} [flags]
```

## Flags
| Flag | Description |
| :--- | :--- |
|    -t, --artifact-type string  |  [conditional] The type of the artifact to calculate its SHA256 fingerprint. One of: [docker, file, dir]. Only required if you don't specify '--fingerprint'.  |
|    -b, --build-url string  |  The url of CI pipeline that built the artifact. (defaulted in some CIs: https://docs.kosli.com/ci-defaults ).  |
|    -u, --commit-url string  |  The url for the git commit that created the artifact. (defaulted in some CIs: https://docs.kosli.com/ci-defaults ).  |
|    -D, --dry-run  |  [optional] Run in dry-run mode. When enabled, no data is sent to Kosli and the CLI exits with 0 exit code regardless of any errors.  |
|    -x, --exclude strings  |  [optional] The comma separated list of directories and files to exclude from fingerprinting. Only applicable for --artifact-type dir.  |
|    -F, --fingerprint string  |  [conditional] The SHA256 fingerprint of the artifact. Only required if you don't specify '--artifact-type'.  |
|    -f, --flow string  |  The Kosli flow name.  |
|    -g, --git-commit string  |  [defaulted] The git commit from which the artifact was created. (defaulted in some CIs: https://docs.kosli.com/ci-defaults, otherwise defaults to HEAD ).  |
|    -h, --help  |  help for artifact  |
|    -n, --name string  |  [optional] Artifact display name, if different from file, image or directory name.  |
|        --registry-password string  |  [conditional] The docker registry password or access token. Only required if you want to read docker image SHA256 digest from a remote docker registry.  |
|        --registry-provider string  |  [conditional] The docker registry provider or url. Only required if you want to read docker image SHA256 digest from a remote docker registry.  |
|        --registry-username string  |  [conditional] The docker registry username. Only required if you want to read docker image SHA256 digest from a remote docker registry.  |
|        --repo-root string  |  [defaulted] The directory where the source git repository is available. (default ".")  |


## Examples

```shell

# Report to a Kosli flow that a file type artifact has been created
kosli report artifact FILE.tgz \
	--api-token yourApiToken \
	--artifact-type file \
	--build-url https://exampleci.com \
	--commit-url https://github.com/YourOrg/YourProject/commit/yourCommitShaThatThisArtifactWasBuiltFrom \
	--git-commit yourCommitShaThatThisArtifactWasBuiltFrom \
	--org yourOrgName \
	--flow yourFlowName 

# Report to a Kosli flow that an artifact with a provided fingerprint (sha256) has been created
kosli report artifact ANOTHER_FILE.txt \
	--api-token yourApiToken \
	--build-url https://exampleci.com \
	--commit-url https://github.com/YourOrg/YourProject/commit/yourCommitShaThatThisArtifactWasBuiltFrom \
	--git-commit yourCommitShaThatThisArtifactWasBuiltFrom \
	--org yourOrgName \
	--flow yourFlowName \
	--fingerprint yourArtifactFingerprint 

```