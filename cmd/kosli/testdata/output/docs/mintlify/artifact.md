---
title: "artifact"
tag: "DEPRECATED"
description: "Report an artifact creation to a Kosli flow.  "
---

import CliDeprecatedNotice from "/snippets/cli-deprecated-notice.mdx";

<CliDeprecatedNotice />

see kosli attest commands

## Synopsis

```shell
artifact {IMAGE-NAME | FILE-PATH | DIR-PATH} [flags]
```

Report an artifact creation to a Kosli flow.  

The artifact fingerprint can be provided directly with the `--fingerprint` flag, or
calculated based on `--artifact-type` flag.

Artifact type can be one of: "file" for files, "dir" for directories, "oci" for container
images in registries or "docker" for local docker images.

Note: `--artifact-type=docker` reads the image's repo digest via the local Docker daemon.
The image must have been pushed to or pulled from a registry for a repo digest to exist;
a freshly built image (just `docker build`) will not have one. If the image is already in
a registry, prefer `--artifact-type=oci`, which fetches the digest directly from the
registry without needing a local Docker daemon.

For `--artifact-type=oci` (and for `--artifact-type=docker` when `--registry-username`
is set), registry credentials are resolved as follows:
  1) If `--registry-username` (and optionally `--registry-password`) is set, it is used directly.
  2) Otherwise, credentials are discovered automatically from:
     - the Docker config file (`~/.docker/config.json`, populated by `docker login`)
     - the Podman/containers auth file (`~/.config/containers/auth.json`, or `$REGISTRY_AUTH_FILE`)
     - any Docker credential helper configured in that config (e.g. `docker-credential-ecr-login`
       for AWS ECR, `docker-credential-gcloud` for GCR/Artifact Registry, an ACR helper for Azure,
       or a local keychain helper), invoked as an external binary on `$PATH`
     - if none of the above yield credentials, the registry is accessed anonymously, which works
       for public images
  `--registry-provider` is deprecated and no longer used.



## Flags
| Flag | Description |
| :--- | :--- |
|    `-t`, `--artifact-type` string  |  The type of the artifact to calculate its SHA256 fingerprint. One of: [oci, docker, file, dir]. Only required if you want Kosli to calculate the fingerprint for you (i.e. when you don't specify '`--fingerprint`' on commands that allow it).  |
|    `-b`, `--build-url` string  |  The url of CI pipeline that built the artifact. (defaulted in some CIs: [docs](/integrations/ci_cd) ).  |
|    `-u`, `--commit-url` string  |  The url for the git commit that created the artifact. (defaulted in some CIs: [docs](/integrations/ci_cd) ).  |
|    `-D`, `--dry-run`  |  [optional] Run in dry-run mode. When enabled, no data is sent to Kosli and the CLI exits with 0 exit code regardless of any errors.  |
|    `-x`, `--exclude` strings  |  [optional] The comma separated list of directories and files to exclude from fingerprinting. Can take glob patterns. Only applicable for `--artifact-type` dir.  |
|    `-F`, `--fingerprint` string  |  [conditional] The SHA256 fingerprint of the artifact. Only required if you don't specify '`--artifact-type`'.  |
|    `-f`, `--flow` string  |  The Kosli flow name.  |
|    `-g`, `--git-commit` string  |  [defaulted] The git commit from which the artifact was created. (defaulted in some CIs: [docs](/integrations/ci_cd), otherwise defaults to HEAD ).  |
|    `-h`, `--help`  |  help for artifact  |
|    `-n`, `--name` string  |  [optional] Artifact display name, if different from file, image or directory name.  |
|        `--registry-password` string  |  [conditional] The container registry password or access token. Only required if you want to read container image SHA256 digest from a remote container registry and it is not already accessible via Docker/Podman auth files or a credential helper.  |
|        `--registry-username` string  |  [conditional] The container registry username. Only required if you want to read container image SHA256 digest from a remote container registry and it is not already accessible via Docker/Podman auth files or a credential helper.  |
|        `--repo-root` string  |  [defaulted] The directory where the source git repository is available. (default ".")  |


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

