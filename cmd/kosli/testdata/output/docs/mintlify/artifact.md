---
title: "artifact"
beta: false
deprecated: true
description: "Report an artifact creation to a Kosli flow.  "
---

<Warning>
**artifact** is deprecated. see kosli attest commands  Deprecated commands will be removed in a future release.
</Warning>
## Synopsis

```shell
artifact {IMAGE-NAME | FILE-PATH | DIR-PATH} [flags]
```

Report an artifact creation to a Kosli flow.  

The artifact fingerprint can be provided directly with the `--fingerprint` flag, or 
calculated based on `--artifact-type` flag.

Artifact type can be one of: "file" for files, "dir" for directories, "oci" for container
images in registries or "docker" for local docker images.



## Flags
| Flag | Description |
| :--- | :--- |
|    -t, --artifact-type string  |  The type of the artifact to calculate its SHA256 fingerprint. One of: [oci, docker, file, dir]. Only required if you want Kosli to calculate the fingerprint for you (i.e. when you don't specify '--fingerprint' on commands that allow it).  |
|    -b, --build-url string  |  The url of CI pipeline that built the artifact. (defaulted in some CIs: [docs](/ci-defaults) ).  |
|    -u, --commit-url string  |  The url for the git commit that created the artifact. (defaulted in some CIs: [docs](/ci-defaults) ).  |
|    -D, --dry-run  |  [optional] Run in dry-run mode. When enabled, no data is sent to Kosli and the CLI exits with 0 exit code regardless of any errors.  |
|    -x, --exclude strings  |  [optional] The comma separated list of directories and files to exclude from fingerprinting. Can take glob patterns. Only applicable for --artifact-type dir.  |
|    -F, --fingerprint string  |  [conditional] The SHA256 fingerprint of the artifact. Only required if you don't specify '--artifact-type'.  |
|    -f, --flow string  |  The Kosli flow name.  |
|    -g, --git-commit string  |  [defaulted] The git commit from which the artifact was created. (defaulted in some CIs: [docs](/ci-defaults), otherwise defaults to HEAD ).  |
|    -h, --help  |  help for artifact  |
|    -n, --name string  |  [optional] Artifact display name, if different from file, image or directory name.  |
|        --registry-password string  |  [conditional] The container registry password or access token. Only required if you want to read container image SHA256 digest from a remote container registry.  |
|        --registry-username string  |  [conditional] The container registry username. Only required if you want to read container image SHA256 digest from a remote container registry.  |
|        --repo-root string  |  [defaulted] The directory where the source git repository is available. (default ".")  |


## Examples Use Cases

These examples all assume that the flags  `--api-token`, `--org`, `--host`, (and `--flow`, `--trail` when required), are [set/provided](/getting_started/install/#assigning-flags-via-environment-variables). 

<AccordionGroup>
<Accordion title="Report to a Kosli flow that a file type artifact has been created">
```shell
kosli report artifact FILE.tgz 
	--artifact-type file 
	--build-url https://exampleci.com 
	--commit-url https://github.com/YourOrg/YourProject/commit/yourCommitShaThatThisArtifactWasBuiltFrom 
	--git-commit yourCommitShaThatThisArtifactWasBuiltFrom 

```
</Accordion>
<Accordion title="Report to a Kosli flow that an artifact with a provided fingerprint (sha256) has been created">
```shell
kosli report artifact ANOTHER_FILE.txt 
	--build-url https://exampleci.com 
	--commit-url https://github.com/YourOrg/YourProject/commit/yourCommitShaThatThisArtifactWasBuiltFrom 
	--git-commit yourCommitShaThatThisArtifactWasBuiltFrom 
	--fingerprint yourArtifactFingerprint
```
</Accordion>
</AccordionGroup>

