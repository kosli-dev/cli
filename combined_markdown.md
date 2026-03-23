## _index.md
---
title: Welcome to Kosli Docs 
seo_title: Welcome to Kosli Docs 
description: Record all of the changes in your software and business processes so you can prove compliance and maintain security without slowing down.
hideToC: true

hero:
    title: Welcome to Kosli Docs
    link_text: Read the Kosli overview >
    url: /introducing_kosli/
    image: /images/home/artie-hero.svg
    alt_text: Kosli artie reading a book

paragraph: >
    Record all of the changes in your software and business processes so you can prove compliance and maintain security without slowing down. Track and query every change from the command line or browser.

sections:
    title: Dive right in…
    blocks:
        - title: What is Kosli
          image: /images/home/home-concepts.svg
          alt_text: Introducing Kosli icon
          description: Understand what Kosli is and how it works
          link_text: View >
          url: /understand_kosli/what_is_kosli/
        - title: Kosli environments
          image: /images/home/home-environments.svg
          alt_text: Kosli environments icon
          description: Environment reporting explained
          link_text: View >
          url: /getting_started/environments/
        - title: Kosli flows
          image: /images/home/home-flows.svg
          alt_text: Flows and artifact reporting explained
          description: Artifact reporting explained
          link_text: View >
          url: /getting_started/flows/
        - title: Get familiar with Kosli
          image: /images/home/home-quickstart.svg
          alt_text: Use cases icon
          description: Learn how to use Kosli with simple examples
          link_text: View >
          url: /tutorials/get_familiar_with_kosli/
        - title: Command reference
          image: /images/home/home-commands.svg
          alt_text: Command reference icon
          description: All Kosli commands in one place
          link_text: View >
          url: /client_reference/
        - title: Support on Slack
          image: /images/home/home-community.svg
          alt_text: Slack community icon
          description: Join the Kosli Community
          link_text: Join the Kosli Slack Community >
          url: https://www.kosli.com/community/
          new_page: true
---

## _index.md
---
title: API Reference
bookCollapseSection: true
weight: 610
---


## _index.md
---
title: API V1
directLink: "/api_v1.html"
weight: 2
---

## _index.md
---
title: API V2
directLink: "/api_v2.html"
weight: 1
---

## _index.md
---
title: CLI Reference
bookCollapseSection: true
weight: 600
---

# CLI Reference

## kosli.md
---
title: "kosli"
beta: false
deprecated: false
summary: "The Kosli CLI."
---

# kosli

## Synopsis

The Kosli CLI.

Environment variables:
You can set any flag from an environment variable by capitalizing it in snake case and adding the KOSLI_ prefix.
For example, to set --api-token from an environment variable, you can export KOSLI_API_TOKEN=YOUR_API_TOKEN.

Setting the API token to DRY_RUN sets the --dry-run flag.


## Flags
| Flag | Description |
| :--- | :--- |
|    -a, --api-token string  |  The Kosli API token.  |
|    -c, --config-file string  |  [optional] The Kosli config file path. (default "kosli")  |
|        --debug  |  [optional] Print debug logs to stdout. A boolean flag https://docs.kosli.com/faq/#boolean-flags (default false)  |
|    -h, --help  |  help for kosli  |
|    -H, --host string  |  [defaulted] The Kosli endpoint. (default "https://app.kosli.com")  |
|        --http-proxy string  |  [optional] The HTTP proxy URL including protocol and port number. e.g. 'http://proxy-server-ip:proxy-port'  |
|    -r, --max-api-retries int  |  [defaulted] How many times should API calls be retried when the API host is not reachable. (default 3)  |
|        --org string  |  The Kosli organization.  |




## kosli_allow_artifact.md
---
title: "kosli allow artifact"
beta: false
deprecated: false
summary: "Add an artifact to an environment's allowlist.  "
---

# kosli allow artifact

## Synopsis

Add an artifact to an environment's allowlist.  

The artifact fingerprint can be provided directly with the `--fingerprint` flag, or 
calculated based on `--artifact-type` flag.

Artifact type can be one of: "file" for files, "dir" for directories, "oci" for container
images in registries or "docker" for local docker images.



```shell
kosli allow artifact [IMAGE-NAME | FILE-PATH | DIR-PATH] [flags]
```

## Flags
| Flag | Description |
| :--- | :--- |
|    -t, --artifact-type string  |  The type of the artifact to calculate its SHA256 fingerprint. One of: [oci, docker, file, dir]. Only required if you want Kosli to calculate the fingerprint for you (i.e. when you don't specify '--fingerprint' on commands that allow it).  |
|    -D, --dry-run  |  [optional] Run in dry-run mode. When enabled, no data is sent to Kosli and the CLI exits with 0 exit code regardless of any errors.  |
|    -e, --environment string  |  The environment name for which the artifact is allowlisted.  |
|    -x, --exclude strings  |  [optional] The comma separated list of directories and files to exclude from fingerprinting. Can take glob patterns. Only applicable for --artifact-type dir.  |
|    -F, --fingerprint string  |  [conditional] The SHA256 fingerprint of the artifact. Only required if you don't specify '--artifact-type'.  |
|    -h, --help  |  help for artifact  |
|        --reason string  |  The reason why this artifact is allowlisted.  |
|        --registry-password string  |  [conditional] The container registry password or access token. Only required if you want to read container image SHA256 digest from a remote container registry.  |
|        --registry-username string  |  [conditional] The container registry username. Only required if you want to read container image SHA256 digest from a remote container registry.  |


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




## kosli_archive_attestation-type.md
---
title: "kosli archive attestation-type"
beta: false
deprecated: false
summary: "Archive a custom Kosli attestation type."
---

# kosli archive attestation-type

## Synopsis

Archive a custom Kosli attestation type.
New custom attestations using this type cannot be made, but existing attestations will still be visible.


```shell
kosli archive attestation-type TYPE-NAME [flags]
```

## Flags
| Flag | Description |
| :--- | :--- |
|    -D, --dry-run  |  [optional] Run in dry-run mode. When enabled, no data is sent to Kosli and the CLI exits with 0 exit code regardless of any errors.  |
|    -h, --help  |  help for attestation-type  |


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

These examples all assume that the flags  `--api-token`, `--org`, `--host`, (and `--flow`, `--trail` when required), are set/provided. 

**archive a Kosli custom attestation type**

```shell
kosli archive attestation-type yourAttestationTypeName 
```



## kosli_archive_environment.md
---
title: "kosli archive environment"
beta: false
deprecated: false
summary: "Archive a Kosli environment."
---

# kosli archive environment

## Synopsis

Archive a Kosli environment.
The environment will no longer be visible in list of environments, data is still stored in the database.


```shell
kosli archive environment ENVIRONMENT-NAME [flags]
```

## Flags
| Flag | Description |
| :--- | :--- |
|    -D, --dry-run  |  [optional] Run in dry-run mode. When enabled, no data is sent to Kosli and the CLI exits with 0 exit code regardless of any errors.  |
|    -h, --help  |  help for environment  |


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

These examples all assume that the flags  `--api-token`, `--org`, `--host`, (and `--flow`, `--trail` when required), are set/provided. 

**archive a Kosli environment**

```shell
kosli archive environment yourEnvironmentName 
```



## kosli_archive_flow.md
---
title: "kosli archive flow"
beta: false
deprecated: false
summary: "Archive a Kosli flow."
---

# kosli archive flow

## Synopsis

Archive a Kosli flow.
The flow will no longer be visible in list of flows, data is still stored in the database.


```shell
kosli archive flow FLOW-NAME [flags]
```

## Flags
| Flag | Description |
| :--- | :--- |
|    -D, --dry-run  |  [optional] Run in dry-run mode. When enabled, no data is sent to Kosli and the CLI exits with 0 exit code regardless of any errors.  |
|    -h, --help  |  help for flow  |


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

These examples all assume that the flags  `--api-token`, `--org`, `--host`, (and `--flow`, `--trail` when required), are set/provided. 

**archive a Kosli flow**

```shell
kosli archive flow yourFlowName 
```



## kosli_assert_approval.md
---
title: "kosli assert approval"
beta: false
deprecated: false
summary: "Assert an artifact in Kosli has been approved for deployment.  "
---

# kosli assert approval

## Synopsis

Assert an artifact in Kosli has been approved for deployment.  
Exits with non-zero code if the artifact has not been approved.  

The artifact fingerprint can be provided directly with the `--fingerprint` flag, or 
calculated based on `--artifact-type` flag.

Artifact type can be one of: "file" for files, "dir" for directories, "oci" for container
images in registries or "docker" for local docker images.



```shell
kosli assert approval [IMAGE-NAME | FILE-PATH | DIR-PATH] [flags]
```

## Flags
| Flag | Description |
| :--- | :--- |
|    -t, --artifact-type string  |  The type of the artifact to calculate its SHA256 fingerprint. One of: [oci, docker, file, dir]. Only required if you want Kosli to calculate the fingerprint for you (i.e. when you don't specify '--fingerprint' on commands that allow it).  |
|    -D, --dry-run  |  [optional] Run in dry-run mode. When enabled, no data is sent to Kosli and the CLI exits with 0 exit code regardless of any errors.  |
|    -x, --exclude strings  |  [optional] The comma separated list of directories and files to exclude from fingerprinting. Can take glob patterns. Only applicable for --artifact-type dir.  |
|    -F, --fingerprint string  |  [conditional] The SHA256 fingerprint of the artifact. Only required if you don't specify '--artifact-type'.  |
|    -f, --flow string  |  The Kosli flow name.  |
|    -h, --help  |  help for approval  |
|        --registry-password string  |  [conditional] The container registry password or access token. Only required if you want to read container image SHA256 digest from a remote container registry.  |
|        --registry-username string  |  [conditional] The container registry username. Only required if you want to read container image SHA256 digest from a remote container registry.  |


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

These examples all assume that the flags  `--api-token`, `--org`, `--host`, (and `--flow`, `--trail` when required), are set/provided. 

**Assert that a file type artifact has been approved**

```shell
kosli assert approval FILE.tgz 
	--artifact-type file 


```

**Assert that an artifact with a provided fingerprint (sha256) has been approved**

```shell
kosli assert approval 
	--fingerprint yourArtifactFingerprint
```



## kosli_assert_artifact.md
---
title: "kosli assert artifact"
beta: false
deprecated: false
summary: "Assert the compliance status of an artifact in Kosli (in its flow or against an environment).  "
---

# kosli assert artifact

## Synopsis

Assert the compliance status of an artifact in Kosli (in its flow or against an environment).  
Exits with non-zero code if the artifact has a non-compliant status.

```shell
kosli assert artifact [IMAGE-NAME | FILE-PATH | DIR-PATH] [flags]
```

## Flags
| Flag | Description |
| :--- | :--- |
|    -t, --artifact-type string  |  The type of the artifact to calculate its SHA256 fingerprint. One of: [oci, docker, file, dir]. Only required if you want Kosli to calculate the fingerprint for you (i.e. when you don't specify '--fingerprint' on commands that allow it).  |
|    -D, --dry-run  |  [optional] Run in dry-run mode. When enabled, no data is sent to Kosli and the CLI exits with 0 exit code regardless of any errors.  |
|        --environment string  |  The Kosli environment name to assert the artifact against.  |
|    -x, --exclude strings  |  [optional] The comma separated list of directories and files to exclude from fingerprinting. Can take glob patterns. Only applicable for --artifact-type dir.  |
|    -F, --fingerprint string  |  [conditional] The SHA256 fingerprint of the artifact. Only required if you don't specify '--artifact-type'.  |
|    -f, --flow string  |  The Kosli flow name.  |
|    -h, --help  |  help for artifact  |
|        --registry-password string  |  [conditional] The container registry password or access token. Only required if you want to read container image SHA256 digest from a remote container registry.  |
|        --registry-username string  |  [conditional] The container registry username. Only required if you want to read container image SHA256 digest from a remote container registry.  |


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

{{< tabs "live-examples" "col-no-wrap" >}}{{< tab "GitHub" >}}View an example of the `kosli assert artifact` command in GitHub.

In [this YAML file](https://app.kosli.com/api/v2/livedocs/cyber-dojo/yaml?ci=github&command=kosli+assert+artifact), which created [this Kosli Event](https://app.kosli.com/api/v2/livedocs/cyber-dojo/event?ci=github&command=kosli+assert+artifact).{{< /tab >}}{{< tab "GitLab" >}}View an example of the `kosli assert artifact` command in GitLab.

In [this YAML file](https://app.kosli.com/api/v2/livedocs/cyber-dojo/yaml?ci=gitlab&command=kosli+assert+artifact), which created [this Kosli Event](https://app.kosli.com/api/v2/livedocs/cyber-dojo/event?ci=gitlab&command=kosli+assert+artifact).{{< /tab >}}{{< /tabs >}}

## Examples Use Cases

These examples all assume that the flags  `--api-token`, `--org`, `--host`, (and `--flow`, `--trail` when required), are set/provided. 

**assert that an artifact meets all compliance requirements for an environment**

```shell
kosli assert artifact 
	--fingerprint 184c799cd551dd1d8d5c5f9a5d593b2e931f5e36122ee5c793c1d08a19839cc0 
	--environment prod 

```

**fail if an artifact has a non-compliant status (using the artifact fingerprint)**

```shell
kosli assert artifact 
	--fingerprint 184c799cd551dd1d8d5c5f9a5d593b2e931f5e36122ee5c793c1d08a19839cc0 

```

**fail if an artifact has a non-compliant status (using the artifact name and type)**

```shell
kosli assert artifact library/nginx:1.21 
	--artifact-type docker 
```



## kosli_assert_pullrequest_azure.md
---
title: "kosli assert pullrequest azure"
beta: false
deprecated: false
summary: "Assert an Azure DevOps pull request for a git commit exists.  "
---

# kosli assert pullrequest azure

## Synopsis

Assert an Azure DevOps pull request for a git commit exists.  
The command exits with non-zero exit code 
if no pull requests were found for the commit.

```shell
kosli assert pullrequest azure [flags]
```

## Flags
| Flag | Description |
| :--- | :--- |
|        --azure-org-url string  |  Azure organization url. E.g. "https://dev.azure.com/myOrg" (defaulted if you are running in Azure Devops pipelines: https://docs.kosli.com/ci-defaults ).  |
|        --azure-token string  |  Azure Personal Access token.  |
|        --commit string  |  Git commit for which to find pull request evidence. (defaulted in some CIs: https://docs.kosli.com/ci-defaults ). (default "HEAD")  |
|    -D, --dry-run  |  [optional] Run in dry-run mode. When enabled, no data is sent to Kosli and the CLI exits with 0 exit code regardless of any errors.  |
|    -h, --help  |  help for azure  |
|        --project string  |  Azure project.(defaulted if you are running in Azure Devops pipelines: https://docs.kosli.com/ci-defaults ).  |
|        --repository string  |  Git repository. (defaulted in some CIs: https://docs.kosli.com/ci-defaults ).  |


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

These examples all assume that the flags  `--api-token`, `--org`, `--host`, (and `--flow`, `--trail` when required), are set/provided. 

```shell
kosli assert pullrequest azure \
	--azure-token yourAzureToken \
	--azure-org-url yourAzureOrgUrl \
	--commit yourGitCommit \
	--project yourAzureDevopsProject \
	--repository yourAzureDevOpsGitRepository
```



## kosli_assert_pullrequest_bitbucket.md
---
title: "kosli assert pullrequest bitbucket"
beta: false
deprecated: false
summary: "Assert a Bitbucket pull request for a git commit exists.  "
---

# kosli assert pullrequest bitbucket

## Synopsis

Assert a Bitbucket pull request for a git commit exists.  
The command exits with non-zero exit code if no pull requests were found for the commit.
Authentication to Bitbucket can be done with access token (recommended) or app passwords. Credentials need to have read access for both repos and pull requests.

```shell
kosli assert pullrequest bitbucket [flags]
```

## Flags
| Flag | Description |
| :--- | :--- |
|        --bitbucket-access-token string  |  Bitbucket repo/project/workspace access token. See https://developer.atlassian.com/cloud/bitbucket/rest/intro/#access-tokens for more details.  |
|        --bitbucket-password string  |  Bitbucket App password. See https://developer.atlassian.com/cloud/bitbucket/rest/intro/#authentication for more details.  |
|        --bitbucket-username string  |  Bitbucket username. Only needed if you use --bitbucket-password  |
|        --bitbucket-workspace string  |  Bitbucket workspace ID.  |
|        --commit string  |  Git commit for which to find pull request evidence. (defaulted in some CIs: https://docs.kosli.com/ci-defaults ). (default "HEAD")  |
|    -D, --dry-run  |  [optional] Run in dry-run mode. When enabled, no data is sent to Kosli and the CLI exits with 0 exit code regardless of any errors.  |
|    -h, --help  |  help for bitbucket  |
|        --repository string  |  Git repository. (defaulted in some CIs: https://docs.kosli.com/ci-defaults ).  |


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

These examples all assume that the flags  `--api-token`, `--org`, `--host`, (and `--flow`, `--trail` when required), are set/provided. 

```shell
kosli assert pullrequest bitbucket  \
	--bitbucket-access-token yourBitbucketAccessToken \
	--bitbucket-workspace yourBitbucketWorkspace \
	--commit yourGitCommit \
	--repository yourBitbucketGitRepository
```



## kosli_assert_pullrequest_github.md
---
title: "kosli assert pullrequest github"
beta: false
deprecated: false
summary: "Assert a Github pull request for a git commit exists.  "
---

# kosli assert pullrequest github

## Synopsis

Assert a Github pull request for a git commit exists.  
The command exits with non-zero exit code 
if no pull requests were found for the commit.

```shell
kosli assert pullrequest github [flags]
```

## Flags
| Flag | Description |
| :--- | :--- |
|        --commit string  |  Git commit for which to find pull request evidence. (defaulted in some CIs: https://docs.kosli.com/ci-defaults ). (default "HEAD")  |
|    -D, --dry-run  |  [optional] Run in dry-run mode. When enabled, no data is sent to Kosli and the CLI exits with 0 exit code regardless of any errors.  |
|        --github-base-url string  |  [optional] GitHub base URL (only needed for GitHub Enterprise installations).  |
|        --github-org string  |  Github organization. (defaulted if you are running in GitHub Actions: https://docs.kosli.com/ci-defaults ).  |
|        --github-token string  |  Github token.  |
|    -h, --help  |  help for github  |
|        --repository string  |  Git repository. (defaulted in some CIs: https://docs.kosli.com/ci-defaults ).  |


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

These examples all assume that the flags  `--api-token`, `--org`, `--host`, (and `--flow`, `--trail` when required), are set/provided. 

```shell
kosli assert pullrequest github \
	--github-token yourGithubToken \
	--github-org yourGithubOrg \
	--commit yourGitCommit \
	--repository yourGithubGitRepository
```



## kosli_assert_pullrequest_gitlab.md
---
title: "kosli assert pullrequest gitlab"
beta: false
deprecated: false
summary: "Assert a Gitlab merge request for a git commit exists.  "
---

# kosli assert pullrequest gitlab

## Synopsis

Assert a Gitlab merge request for a git commit exists.  
The command exits with non-zero exit code 
if no merge requests were found for the commit.

```shell
kosli assert pullrequest gitlab [flags]
```

## Flags
| Flag | Description |
| :--- | :--- |
|        --commit string  |  Git commit for which to find pull request evidence. (defaulted in some CIs: https://docs.kosli.com/ci-defaults ). (default "HEAD")  |
|    -D, --dry-run  |  [optional] Run in dry-run mode. When enabled, no data is sent to Kosli and the CLI exits with 0 exit code regardless of any errors.  |
|        --gitlab-base-url string  |  [optional] Gitlab base URL (only needed for on-prem Gitlab installations).  |
|        --gitlab-org string  |  Gitlab organization. (defaulted if you are running in Gitlab Pipelines: https://docs.kosli.com/ci-defaults ).  |
|        --gitlab-token string  |  Gitlab token.  |
|    -h, --help  |  help for gitlab  |
|        --repository string  |  Git repository. (defaulted in some CIs: https://docs.kosli.com/ci-defaults ).  |


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

These examples all assume that the flags  `--api-token`, `--org`, `--host`, (and `--flow`, `--trail` when required), are set/provided. 

```shell
kosli assert mergerequest gitlab \
	--github-token yourGithubToken \
	--github-org yourGithubOrg \
	--commit yourGitCommit \
	--repository yourGithubGitRepository
```



## kosli_assert_snapshot.md
---
title: "kosli assert snapshot"
beta: false
deprecated: false
summary: "Assert the compliance status of an environment in Kosli."
---

# kosli assert snapshot

## Synopsis

Assert the compliance status of an environment in Kosli.
Exits with non-zero code if the environment has a non-compliant status.
The expected argument is an expression to specify the specific environment snapshot to assert.
It has the format <ENVIRONMENT_NAME>[SEPARATOR][SNAPSHOT_REFERENCE] 

Separators can be:
- '#' to specify a specific snapshot number for the environment that is being asserted.
- '~' to get N-th behind the latest snapshot.

Examples of valid expressions are: 
- prod (latest snapshot of prod)
- prod#10 (snapshot number 10 of prod)
- prod~2 (third latest snapshot of prod)


```shell
kosli assert snapshot ENVIRONMENT-NAME-OR-EXPRESSION [flags]
```

## Flags
| Flag | Description |
| :--- | :--- |
|    -D, --dry-run  |  [optional] Run in dry-run mode. When enabled, no data is sent to Kosli and the CLI exits with 0 exit code regardless of any errors.  |
|    -h, --help  |  help for snapshot  |


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

These examples all assume that the flags  `--api-token`, `--org`, `--host`, (and `--flow`, `--trail` when required), are set/provided. 

```shell
kosli assert snapshot prod#5 \
	--api-token yourAPIToken \
	--org yourOrgName
```



## kosli_assert_status.md
---
title: "kosli assert status"
beta: false
deprecated: false
summary: "Assert the status of a Kosli server."
---

# kosli assert status

## Synopsis

Assert the status of a Kosli server.
Exits with non-zero code if the Kosli server down.

```shell
kosli assert status [flags]
```

## Flags
| Flag | Description |
| :--- | :--- |
|    -h, --help  |  help for status  |


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




## kosli_attach-policy.md
---
title: "kosli attach-policy"
beta: false
deprecated: false
summary: "Attach a policy to one or more Kosli environments.  "
---

# kosli attach-policy

## Synopsis

Attach a policy to one or more Kosli environments.  

```shell
kosli attach-policy POLICY-NAME [flags]
```

## Flags
| Flag | Description |
| :--- | :--- |
|    -D, --dry-run  |  [optional] Run in dry-run mode. When enabled, no data is sent to Kosli and the CLI exits with 0 exit code regardless of any errors.  |
|    -e, --environment strings  |  the list of environment names to attach the policy to  |
|    -h, --help  |  help for attach-policy  |


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

These examples all assume that the flags  `--api-token`, `--org`, `--host`, (and `--flow`, `--trail` when required), are set/provided. 

**attach a previously created policy to multiple environment**

```shell
kosli attach-policy yourPolicyName 
	--environment yourFirstEnvironmentName 
	--environment yourSecondEnvironmentName 
```



## kosli_attest_artifact.md
---
title: "kosli attest artifact"
beta: false
deprecated: false
summary: "Attest an artifact creation to a Kosli flow.  "
---

# kosli attest artifact

## Synopsis

Attest an artifact creation to a Kosli flow.  

The artifact fingerprint can be provided directly with the `--fingerprint` flag, or 
calculated based on `--artifact-type` flag.

Artifact type can be one of: "file" for files, "dir" for directories, "oci" for container
images in registries or "docker" for local docker images.

To specify paths in a directory artifact that should always be excluded from the SHA256 calculation, you can add a `.kosli_ignore` file to the root of the artifact.
Each line should specify a relative path or path glob to be ignored. You can include comments in this file, using `#`.
The `.kosli_ignore` will be treated as part of the artifact like any other file, unless it is explicitly ignored itself.
This command requires access to a git repo to associate the artifact to the git commit it is originating from. 
You can optionally redact some of the git commit data sent to Kosli using `--redact-commit-info`

```shell
kosli attest artifact {IMAGE-NAME | FILE-PATH | DIR-PATH} [flags]
```

## Flags
| Flag | Description |
| :--- | :--- |
|        --annotate stringToString  |  [optional] Annotate the attestation with data using key=value.  |
|    -t, --artifact-type string  |  The type of the artifact to calculate its SHA256 fingerprint. One of: [oci, docker, file, dir]. Only required if you want Kosli to calculate the fingerprint for you (i.e. when you don't specify '--fingerprint' on commands that allow it).  |
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
|        --registry-password string  |  [conditional] The container registry password or access token. Only required if you want to read container image SHA256 digest from a remote container registry.  |
|        --registry-username string  |  [conditional] The container registry username. Only required if you want to read container image SHA256 digest from a remote container registry.  |
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

These examples all assume that the flags  `--api-token`, `--org`, `--host`, (and `--flow`, `--trail` when required), are set/provided. 

**Attest that a file type artifact has been created, and let Kosli calculate its fingerprint**

```shell
kosli attest artifact FILE.tgz 
	--artifact-type file 
	--build-url https://exampleci.com 
	--commit-url https://github.com/YourOrg/YourProject/commit/yourCommitShaThatThisArtifactWasBuiltFrom 
	--commit yourCommitShaThatThisArtifactWasBuiltFrom 
	--name yourTemplateArtifactName 


```

**Attest that an artifact has been created and provide its fingerprint (sha256)**

```shell
kosli attest artifact ANOTHER_FILE.txt 
	--build-url https://exampleci.com 
	--commit-url https://github.com/YourOrg/YourProject/commit/yourCommitShaThatThisArtifactWasBuiltFrom 
	--commit yourCommitShaThatThisArtifactWasBuiltFrom 
	--fingerprint yourArtifactFingerprint 
	--name yourTemplateArtifactName 

```

**Attest that an artifact has been created and provide external attachments**

```shell
kosli attest artifact ANOTHER_FILE.txt 
	--build-url https://exampleci.com 
	--commit-url https://github.com/YourOrg/YourProject/commit/yourCommitShaThatThisArtifactWasBuiltFrom 
	--commit yourCommitShaThatThisArtifactWasBuiltFrom 
	--fingerprint yourArtifactFingerprint 
	--external-url label=https://example.com/attachment 
	--external-fingerprint label=yourExternalAttachmentFingerprint 
	--name yourTemplateArtifactName 
```



## kosli_attest_custom.md
---
title: "kosli attest custom"
beta: false
deprecated: false
summary: "Report a custom attestation to an artifact or a trail in a Kosli flow. "
---

# kosli attest custom

## Synopsis

Report a custom attestation to an artifact or a trail in a Kosli flow. 
The name of the custom attestation type is specified using the `--type` flag.
The path to the JSON file the custom type will evaluate is specified using the `--attestation-data` flag.


The attestation can be bound to a *trail* using the trail name.  
The attestation can be bound to an *artifact* in two ways:
- using the artifact's SHA256 fingerprint which is calculated (based on the `--artifact-type` flag and the artifact name/path argument) or can be provided directly (with the `--fingerprint` flag).
- using the artifact's name in the flow yaml template and the git commit from which the artifact is/will be created. Useful when reporting an attestation before creating/reporting the artifact.

You can optionally associate the attestation to a git commit using `--commit` (requires access to a git repo).
You can optionally redact some of the git commit data sent to Kosli using `--redact-commit-info`.
Note that when the attestation is reported for an artifact that does not yet exist in Kosli, `--commit` is required to facilitate
binding the attestation to the right artifact.

```shell
kosli attest custom [IMAGE-NAME | FILE-PATH | DIR-PATH] [flags]
```

## Flags
| Flag | Description |
| :--- | :--- |
|        --annotate stringToString  |  [optional] Annotate the attestation with data using key=value.  |
|    -t, --artifact-type string  |  The type of the artifact to calculate its SHA256 fingerprint. One of: [oci, docker, file, dir]. Only required if you want Kosli to calculate the fingerprint for you (i.e. when you don't specify '--fingerprint' on commands that allow it).  |
|        --attachments strings  |  [optional] The comma-separated list of paths of attachments for the reported attestation. Attachments can be files or directories. All attachments are compressed and uploaded to Kosli's evidence vault.  |
|        --attestation-data string  |  The filepath of a json file containing the custom attestation data.  |
|    -g, --commit string  |  [conditional] The git commit for which the attestation is associated to. Becomes required when reporting an attestation for an artifact before reporting it to Kosli. (defaulted in some CIs: https://docs.kosli.com/ci-defaults ).  |
|        --description string  |  [optional] attestation description  |
|    -D, --dry-run  |  [optional] Run in dry-run mode. When enabled, no data is sent to Kosli and the CLI exits with 0 exit code regardless of any errors.  |
|    -x, --exclude strings  |  [optional] The comma separated list of directories and files to exclude from fingerprinting. Can take glob patterns. Only applicable for --artifact-type dir.  |
|        --external-fingerprint stringToString  |  [optional] A SHA256 fingerprint of an external attachment represented by --external-url. The format is label=fingerprint (labels cannot contain '.' or '='). This flag can be set multiple times. There must be an external url with a matching label for each external fingerprint.  |
|        --external-url stringToString  |  [optional] Add labeled reference URL for an external resource. The format is label=url (labels cannot contain '.' or '='). This flag can be set multiple times. If the resource is a file or dir, you can optionally add its fingerprint via --external-fingerprint  |
|    -F, --fingerprint string  |  [conditional] The SHA256 fingerprint of the artifact to attach the attestation to. Only required if the attestation is for an artifact and --artifact-type and artifact name/path are not used.  |
|    -f, --flow string  |  The Kosli flow name.  |
|    -h, --help  |  help for custom  |
|    -n, --name string  |  The name of the attestation as declared in the flow or trail yaml template.  |
|    -o, --origin-url string  |  [optional] The url pointing to where the attestation came from or is related. (defaulted to the CI url in some CIs: https://docs.kosli.com/ci-defaults ).  |
|        --redact-commit-info strings  |  [optional] The list of commit info to be redacted before sending to Kosli. Allowed values are one or more of [author, message, branch].  |
|        --registry-password string  |  [conditional] The container registry password or access token. Only required if you want to read container image SHA256 digest from a remote container registry.  |
|        --registry-username string  |  [conditional] The container registry username. Only required if you want to read container image SHA256 digest from a remote container registry.  |
|        --repo-root string  |  [defaulted] The directory where the source git repository is available. Only used if --commit is used. (default ".")  |
|    -T, --trail string  |  The Kosli trail name.  |
|        --type string  |  The name of the custom attestation type.  |
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

{{< tabs "live-examples" "col-no-wrap" >}}{{< tab "GitHub" >}}View an example of the `kosli attest custom` command in GitHub.

In [this YAML file](https://app.kosli.com/api/v2/livedocs/cyber-dojo/yaml?ci=github&command=kosli+attest+custom), which created [this Kosli Event](https://app.kosli.com/api/v2/livedocs/cyber-dojo/event?ci=github&command=kosli+attest+custom).{{< /tab >}}{{< /tabs >}}

## Examples Use Cases

These examples all assume that the flags  `--api-token`, `--org`, `--host`, (and `--flow`, `--trail` when required), are set/provided. 

**report a custom attestation about a pre-built container image artifact (kosli finds the fingerprint)**

```shell
kosli attest custom yourDockerImageName 
	--artifact-type oci 
	--type customTypeName 
	--name yourAttestationName 
	--attestation-data yourJsonFilePath 

```

**report a custom attestation about a pre-built docker artifact (you provide the fingerprint)**

```shell
kosli attest custom 
	--fingerprint yourDockerImageFingerprint 
	--type customTypeName 
	--name yourAttestationName 
	--attestation-data yourJsonFilePath 

```

**report a custom attestation about a trail**

```shell
kosli attest custom 
	--type customTypeName 
	--name yourAttestationName 
	--attestation-data yourJsonFilePath 

```

**report a custom attestation about an artifact which has not been reported yet in a trail**

```shell
kosli attest custom 
	--type customTypeName 
	--name yourTemplateArtifactName.yourAttestationName 
	--attestation-data yourJsonFilePath 
	--commit yourArtifactGitCommit 

```

**report a custom attestation about a trail with an attachment**

```shell
kosli attest custom 
    --type customTypeName 
	--name yourAttestationName 
	--attestation-data yourJsonFilePath 
	--attachments yourAttachmentPathName 
```



## kosli_attest_generic.md
---
title: "kosli attest generic"
beta: false
deprecated: false
summary: "Report a generic attestation to an artifact or a trail in a Kosli flow.  "
---

# kosli attest generic

## Synopsis

Report a generic attestation to an artifact or a trail in a Kosli flow.  

The attestation can be bound to a *trail* using the trail name.  
The attestation can be bound to an *artifact* in two ways:
- using the artifact's SHA256 fingerprint which is calculated (based on the `--artifact-type` flag and the artifact name/path argument) or can be provided directly (with the `--fingerprint` flag).
- using the artifact's name in the flow yaml template and the git commit from which the artifact is/will be created. Useful when reporting an attestation before creating/reporting the artifact.

You can optionally associate the attestation to a git commit using `--commit` (requires access to a git repo).
You can optionally redact some of the git commit data sent to Kosli using `--redact-commit-info`.
Note that when the attestation is reported for an artifact that does not yet exist in Kosli, `--commit` is required to facilitate
binding the attestation to the right artifact.

```shell
kosli attest generic [IMAGE-NAME | FILE-PATH | DIR-PATH] [flags]
```

## Flags
| Flag | Description |
| :--- | :--- |
|        --annotate stringToString  |  [optional] Annotate the attestation with data using key=value.  |
|    -t, --artifact-type string  |  The type of the artifact to calculate its SHA256 fingerprint. One of: [oci, docker, file, dir]. Only required if you want Kosli to calculate the fingerprint for you (i.e. when you don't specify '--fingerprint' on commands that allow it).  |
|        --attachments strings  |  [optional] The comma-separated list of paths of attachments for the reported attestation. Attachments can be files or directories. All attachments are compressed and uploaded to Kosli's evidence vault.  |
|    -g, --commit string  |  [conditional] The git commit for which the attestation is associated to. Becomes required when reporting an attestation for an artifact before reporting it to Kosli. (defaulted in some CIs: https://docs.kosli.com/ci-defaults ).  |
|    -C, --compliant  |  [defaulted] Whether the attestation is compliant or not. A boolean flag https://docs.kosli.com/faq/#boolean-flags (default true)  |
|        --description string  |  [optional] attestation description  |
|    -D, --dry-run  |  [optional] Run in dry-run mode. When enabled, no data is sent to Kosli and the CLI exits with 0 exit code regardless of any errors.  |
|    -x, --exclude strings  |  [optional] The comma separated list of directories and files to exclude from fingerprinting. Can take glob patterns. Only applicable for --artifact-type dir.  |
|        --external-fingerprint stringToString  |  [optional] A SHA256 fingerprint of an external attachment represented by --external-url. The format is label=fingerprint (labels cannot contain '.' or '='). This flag can be set multiple times. There must be an external url with a matching label for each external fingerprint.  |
|        --external-url stringToString  |  [optional] Add labeled reference URL for an external resource. The format is label=url (labels cannot contain '.' or '='). This flag can be set multiple times. If the resource is a file or dir, you can optionally add its fingerprint via --external-fingerprint  |
|    -F, --fingerprint string  |  [conditional] The SHA256 fingerprint of the artifact to attach the attestation to. Only required if the attestation is for an artifact and --artifact-type and artifact name/path are not used.  |
|    -f, --flow string  |  The Kosli flow name.  |
|    -h, --help  |  help for generic  |
|    -n, --name string  |  The name of the attestation as declared in the flow or trail yaml template.  |
|    -o, --origin-url string  |  [optional] The url pointing to where the attestation came from or is related. (defaulted to the CI url in some CIs: https://docs.kosli.com/ci-defaults ).  |
|        --redact-commit-info strings  |  [optional] The list of commit info to be redacted before sending to Kosli. Allowed values are one or more of [author, message, branch].  |
|        --registry-password string  |  [conditional] The container registry password or access token. Only required if you want to read container image SHA256 digest from a remote container registry.  |
|        --registry-username string  |  [conditional] The container registry username. Only required if you want to read container image SHA256 digest from a remote container registry.  |
|        --repo-root string  |  [defaulted] The directory where the source git repository is available. Only used if --commit is used. (default ".")  |
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

{{< tabs "live-examples" "col-no-wrap" >}}{{< tab "GitHub" >}}View an example of the `kosli attest generic` command in GitHub.

In [this YAML file](https://app.kosli.com/api/v2/livedocs/cyber-dojo/yaml?ci=github&command=kosli+attest+generic), which created [this Kosli Event](https://app.kosli.com/api/v2/livedocs/cyber-dojo/event?ci=github&command=kosli+attest+generic).{{< /tab >}}{{< tab "GitLab" >}}View an example of the `kosli attest generic` command in GitLab.

In [this YAML file](https://app.kosli.com/api/v2/livedocs/cyber-dojo/yaml?ci=gitlab&command=kosli+attest+generic), which created [this Kosli Event](https://app.kosli.com/api/v2/livedocs/cyber-dojo/event?ci=gitlab&command=kosli+attest+generic).{{< /tab >}}{{< /tabs >}}

## Examples Use Cases

These examples all assume that the flags  `--api-token`, `--org`, `--host`, (and `--flow`, `--trail` when required), are set/provided. 

**report a generic attestation about a pre-built docker artifact (kosli calculates the fingerprint)**

```shell
kosli attest generic yourDockerImageName 
	--artifact-type docker 
	--name yourAttestationName 

```

**report a generic attestation about a pre-built docker artifact (you provide the fingerprint)**

```shell
kosli attest generic 
	--fingerprint yourDockerImageFingerprint 
	--name yourAttestationName 

```

**report a generic attestation about a trail**

```shell
kosli attest generic 
	--name yourAttestationName 

```

**report a generic attestation about an artifact which has not been reported yet in a trail**

```shell
kosli attest generic 
	--name yourTemplateArtifactName.yourAttestationName 
	--commit yourArtifactGitCommit 

```

**report a generic attestation about a trail with an attachment**

```shell
kosli attest generic 
	--name yourAttestationName 
	--attachments yourAttachmentPathName 

```

**report a non-compliant generic attestation about a trail**

```shell
kosli attest generic 
	--name yourAttestationName 
	--compliant=false 
```



## kosli_attest_jira.md
---
title: "kosli attest jira"
beta: false
deprecated: false
summary: "Report a jira attestation to an artifact or a trail in a Kosli flow.  "
---

# kosli attest jira

## Synopsis

Report a jira attestation to an artifact or a trail in a Kosli flow.  
Parses the given commit's message, current branch name or the content of the `--jira-secondary-source`
argument for Jira issue references of the form:  
'at least 2 characters long, starting with an uppercase letter project key followed by
dash and one or more digits'. 

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

```shell
kosli attest jira [IMAGE-NAME | FILE-PATH | DIR-PATH] [flags]
```

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
|        --jira-secondary-source string  |  [optional] An optional string to search for Jira ticket reference, e.g. '--jira-secondary-source ${{ github.head_ref }}'  |
|        --jira-username string  |  Jira username (for Jira Cloud)  |
|    -n, --name string  |  The name of the attestation as declared in the flow or trail yaml template.  |
|    -o, --origin-url string  |  [optional] The url pointing to where the attestation came from or is related. (defaulted to the CI url in some CIs: https://docs.kosli.com/ci-defaults ).  |
|        --redact-commit-info strings  |  [optional] The list of commit info to be redacted before sending to Kosli. Allowed values are one or more of [author, message, branch].  |
|        --registry-password string  |  [conditional] The container registry password or access token. Only required if you want to read container image SHA256 digest from a remote container registry.  |
|        --registry-username string  |  [conditional] The container registry username. Only required if you want to read container image SHA256 digest from a remote container registry.  |
|        --repo-root string  |  [defaulted] The directory where the source git repository is available. Only used if --commit is used. (default ".")  |
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

These examples all assume that the flags  `--api-token`, `--org`, `--host`, (and `--flow`, `--trail` when required), are set/provided. 

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



## kosli_attest_junit.md
---
title: "kosli attest junit"
beta: false
deprecated: false
summary: "Report a junit attestation to an artifact or a trail in a Kosli flow.
JUnit xml files are read from the ^--results-dir^ directory which defaults to the current directory.
The xml files are automatically uploaded as ^--attachments^ via the ^--upload-results^ flag which defaults to ^true^.  "
---

# kosli attest junit

## Synopsis

Report a junit attestation to an artifact or a trail in a Kosli flow.
JUnit xml files are read from the `--results-dir` directory which defaults to the current directory.
The xml files are automatically uploaded as `--attachments` via the `--upload-results` flag which defaults to `true`.  

The attestation can be bound to a *trail* using the trail name.  
The attestation can be bound to an *artifact* in two ways:
- using the artifact's SHA256 fingerprint which is calculated (based on the `--artifact-type` flag and the artifact name/path argument) or can be provided directly (with the `--fingerprint` flag).
- using the artifact's name in the flow yaml template and the git commit from which the artifact is/will be created. Useful when reporting an attestation before creating/reporting the artifact.

You can optionally associate the attestation to a git commit using `--commit` (requires access to a git repo).
You can optionally redact some of the git commit data sent to Kosli using `--redact-commit-info`.
Note that when the attestation is reported for an artifact that does not yet exist in Kosli, `--commit` is required to facilitate
binding the attestation to the right artifact.

```shell
kosli attest junit [IMAGE-NAME | FILE-PATH | DIR-PATH] [flags]
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
|    -h, --help  |  help for junit  |
|    -n, --name string  |  The name of the attestation as declared in the flow or trail yaml template.  |
|    -o, --origin-url string  |  [optional] The url pointing to where the attestation came from or is related. (defaulted to the CI url in some CIs: https://docs.kosli.com/ci-defaults ).  |
|        --redact-commit-info strings  |  [optional] The list of commit info to be redacted before sending to Kosli. Allowed values are one or more of [author, message, branch].  |
|        --registry-password string  |  [conditional] The container registry password or access token. Only required if you want to read container image SHA256 digest from a remote container registry.  |
|        --registry-username string  |  [conditional] The container registry username. Only required if you want to read container image SHA256 digest from a remote container registry.  |
|        --repo-root string  |  [defaulted] The directory where the source git repository is available. Only used if --commit is used. (default ".")  |
|    -R, --results-dir string  |  [defaulted] The path to a directory with JUnit test results. By default, the directory will be uploaded to Kosli's evidence vault. (default ".")  |
|    -T, --trail string  |  The Kosli trail name.  |
|        --upload-results  |  [defaulted] Whether to upload the provided Junit results directory as an attachment to Kosli or not. (default true)  |
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

{{< tabs "live-examples" "col-no-wrap" >}}{{< tab "GitHub" >}}View an example of the `kosli attest junit` command in GitHub.

In [this YAML file](https://app.kosli.com/api/v2/livedocs/cyber-dojo/yaml?ci=github&command=kosli+attest+junit), which created [this Kosli Event](https://app.kosli.com/api/v2/livedocs/cyber-dojo/event?ci=github&command=kosli+attest+junit).{{< /tab >}}{{< tab "GitLab" >}}View an example of the `kosli attest junit` command in GitLab.

In [this YAML file](https://app.kosli.com/api/v2/livedocs/cyber-dojo/yaml?ci=gitlab&command=kosli+attest+junit), which created [this Kosli Event](https://app.kosli.com/api/v2/livedocs/cyber-dojo/event?ci=gitlab&command=kosli+attest+junit).{{< /tab >}}{{< /tabs >}}

## Examples Use Cases

These examples all assume that the flags  `--api-token`, `--org`, `--host`, (and `--flow`, `--trail` when required), are set/provided. 

**report a junit attestation about a pre-built docker artifact (kosli calculates the fingerprint)**

```shell
kosli attest junit yourDockerImageName 
	--artifact-type docker 
	--name yourAttestationName 
	--results-dir yourFolderWithJUnitResults 

```

**report a junit attestation about a pre-built docker artifact (you provide the fingerprint)**

```shell
kosli attest junit 
	--fingerprint yourDockerImageFingerprint 
	--name yourAttestationName 
	--results-dir yourFolderWithJUnitResults 

```

**report a junit attestation about a trail**

```shell
kosli attest junit 
	--name yourAttestationName 
	--results-dir yourFolderWithJUnitResults 

```

**report a junit attestation about an artifact which has not been reported yet in a trail**

```shell
kosli attest junit 
	--name yourTemplateArtifactName.yourAttestationName 
	--commit yourArtifactGitCommit 
	--results-dir yourFolderWithJUnitResults 

```

**report a junit attestation about a trail with an attachment**

```shell
kosli attest junit 
	--name yourAttestationName 
	--results-dir yourFolderWithJUnitResults 
	--attachments yourAttachmentPathName 
```



## kosli_attest_pullrequest_azure.md
---
title: "kosli attest pullrequest azure"
beta: false
deprecated: false
summary: "Report an Azure Devops pull request attestation to an artifact or a trail in a Kosli flow.  "
---

# kosli attest pullrequest azure

## Synopsis

Report an Azure Devops pull request attestation to an artifact or a trail in a Kosli flow.  
It checks if a pull request exists for the artifact (based on its git commit) and reports the pull-request attestation to the artifact in Kosli.


The attestation can be bound to a *trail* using the trail name.  
The attestation can be bound to an *artifact* in two ways:
- using the artifact's SHA256 fingerprint which is calculated (based on the `--artifact-type` flag and the artifact name/path argument) or can be provided directly (with the `--fingerprint` flag).
- using the artifact's name in the flow yaml template and the git commit from which the artifact is/will be created. Useful when reporting an attestation before creating/reporting the artifact.

```shell
kosli attest pullrequest azure [IMAGE-NAME | FILE-PATH | DIR-PATH] [flags]
```

## Flags
| Flag | Description |
| :--- | :--- |
|        --annotate stringToString  |  [optional] Annotate the attestation with data using key=value.  |
|    -t, --artifact-type string  |  The type of the artifact to calculate its SHA256 fingerprint. One of: [oci, docker, file, dir]. Only required if you want Kosli to calculate the fingerprint for you (i.e. when you don't specify '--fingerprint' on commands that allow it).  |
|        --assert  |  [optional] Exit with non-zero code if no pull requests found for the given commit.  |
|        --attachments strings  |  [optional] The comma-separated list of paths of attachments for the reported attestation. Attachments can be files or directories. All attachments are compressed and uploaded to Kosli's evidence vault.  |
|        --azure-org-url string  |  Azure organization url. E.g. "https://dev.azure.com/myOrg" (defaulted if you are running in Azure Devops pipelines: https://docs.kosli.com/ci-defaults ).  |
|        --azure-token string  |  Azure Personal Access token.  |
|    -g, --commit string  |  [conditional] The git commit for which the attestation is associated to. Becomes required when reporting an attestation for an artifact before reporting it to Kosli. (defaulted in some CIs: https://docs.kosli.com/ci-defaults ).  |
|        --description string  |  [optional] attestation description  |
|    -D, --dry-run  |  [optional] Run in dry-run mode. When enabled, no data is sent to Kosli and the CLI exits with 0 exit code regardless of any errors.  |
|    -x, --exclude strings  |  [optional] The comma separated list of directories and files to exclude from fingerprinting. Can take glob patterns. Only applicable for --artifact-type dir.  |
|        --external-fingerprint stringToString  |  [optional] A SHA256 fingerprint of an external attachment represented by --external-url. The format is label=fingerprint (labels cannot contain '.' or '='). This flag can be set multiple times. There must be an external url with a matching label for each external fingerprint.  |
|        --external-url stringToString  |  [optional] Add labeled reference URL for an external resource. The format is label=url (labels cannot contain '.' or '='). This flag can be set multiple times. If the resource is a file or dir, you can optionally add its fingerprint via --external-fingerprint  |
|    -F, --fingerprint string  |  [conditional] The SHA256 fingerprint of the artifact to attach the attestation to. Only required if the attestation is for an artifact and --artifact-type and artifact name/path are not used.  |
|    -f, --flow string  |  The Kosli flow name.  |
|    -h, --help  |  help for azure  |
|    -n, --name string  |  The name of the attestation as declared in the flow or trail yaml template.  |
|    -o, --origin-url string  |  [optional] The url pointing to where the attestation came from or is related. (defaulted to the CI url in some CIs: https://docs.kosli.com/ci-defaults ).  |
|        --project string  |  Azure project.(defaulted if you are running in Azure Devops pipelines: https://docs.kosli.com/ci-defaults ).  |
|        --redact-commit-info strings  |  [optional] The list of commit info to be redacted before sending to Kosli. Allowed values are one or more of [author, message, branch].  |
|        --registry-password string  |  [conditional] The container registry password or access token. Only required if you want to read container image SHA256 digest from a remote container registry.  |
|        --registry-username string  |  [conditional] The container registry username. Only required if you want to read container image SHA256 digest from a remote container registry.  |
|        --repo-root string  |  [defaulted] The directory where the source git repository is available. Only used if --commit is used. (default ".")  |
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

These examples all assume that the flags  `--api-token`, `--org`, `--host`, (and `--flow`, `--trail` when required), are set/provided. 

**report an Azure Devops pull request attestation about a pre-built docker artifact (kosli calculates the fingerprint)**

```shell
kosli attest pullrequest azure yourDockerImageName 
	--artifact-type docker 
	--name yourAttestationName 
	--azure-org-url https://dev.azure.com/myOrg 
	--project yourAzureDevOpsProject 
	--azure-token yourAzureToken 
	--commit yourGitCommitSha1 
	--repository yourAzureGitRepository 

```

**report an Azure Devops pull request attestation about a pre-built docker artifact (you provide the fingerprint)**

```shell
kosli attest pullrequest azure 
	--fingerprint yourDockerImageFingerprint 
	--name yourAttestationName 
	--azure-org-url https://dev.azure.com/myOrg 
	--project yourAzureDevOpsProject 
	--azure-token yourAzureToken 
	--commit yourGitCommitSha1 
	--repository yourAzureGitRepository 

```

**report an Azure Devops pull request attestation about a trail**

```shell
kosli attest pullrequest azure 
	--name yourAttestationName 
	--azure-org-url https://dev.azure.com/myOrg 
	--project yourAzureDevOpsProject 
	--azure-token yourAzureToken 
	--commit yourGitCommitSha1 
	--repository yourAzureGitRepository 

```

**report an Azure Devops pull request attestation about an artifact which has not been reported yet in a trail**

```shell
kosli attest pullrequest azure 
	--name yourTemplateArtifactName.yourAttestationName 
	--azure-org-url https://dev.azure.com/myOrg 
	--project yourAzureDevOpsProject 
	--azure-token yourAzureToken 
	--commit yourGitCommitSha1 
	--repository yourAzureGitRepository 

```

**report an Azure Devops pull request attestation about a trail with an attachment**

```shell
kosli attest pullrequest azure 
	--name yourAttestationName 
	--azure-org-url https://dev.azure.com/myOrg 
	--project yourAzureDevOpsProject 
	--azure-token yourAzureToken 
	--commit yourGitCommitSha1 
	--repository yourAzureGitRepository 
	--attachments=yourAttachmentPathName 

```

**fail if a pull request does not exist for your artifact**

```shell
kosli attest pullrequest azure 
	--name yourTemplateArtifactName.yourAttestationName 
	--azure-org-url https://dev.azure.com/myOrg 
	--project yourAzureDevOpsProject 
	--azure-token yourAzureToken 
	--commit yourGitCommitSha1 
	--repository yourAzureGitRepository 
	--assert
```



## kosli_attest_pullrequest_bitbucket.md
---
title: "kosli attest pullrequest bitbucket"
beta: false
deprecated: false
summary: "Report a Bitbucket pull request attestation to an artifact or a trail in a Kosli flow.  "
---

# kosli attest pullrequest bitbucket

## Synopsis

Report a Bitbucket pull request attestation to an artifact or a trail in a Kosli flow.  
It checks if a pull request exists for a given merge commit and reports the pull-request attestation to Kosli.
Authentication to Bitbucket can be done with access token (recommended) or app passwords. Credentials need to have read access for both repos and pull requests.


The attestation can be bound to a *trail* using the trail name.  
The attestation can be bound to an *artifact* in two ways:
- using the artifact's SHA256 fingerprint which is calculated (based on the `--artifact-type` flag and the artifact name/path argument) or can be provided directly (with the `--fingerprint` flag).
- using the artifact's name in the flow yaml template and the git commit from which the artifact is/will be created. Useful when reporting an attestation before creating/reporting the artifact.

```shell
kosli attest pullrequest bitbucket [IMAGE-NAME | FILE-PATH | DIR-PATH] [flags]
```

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
|    -o, --origin-url string  |  [optional] The url pointing to where the attestation came from or is related. (defaulted to the CI url in some CIs: https://docs.kosli.com/ci-defaults ).  |
|        --redact-commit-info strings  |  [optional] The list of commit info to be redacted before sending to Kosli. Allowed values are one or more of [author, message, branch].  |
|        --registry-password string  |  [conditional] The container registry password or access token. Only required if you want to read container image SHA256 digest from a remote container registry.  |
|        --registry-username string  |  [conditional] The container registry username. Only required if you want to read container image SHA256 digest from a remote container registry.  |
|        --repo-root string  |  [defaulted] The directory where the source git repository is available. Only used if --commit is used. (default ".")  |
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

These examples all assume that the flags  `--api-token`, `--org`, `--host`, (and `--flow`, `--trail` when required), are set/provided. 

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



## kosli_attest_pullrequest_github.md
---
title: "kosli attest pullrequest github"
beta: false
deprecated: false
summary: "Report a Github pull request attestation to an artifact or a trail in a Kosli flow.  "
---

# kosli attest pullrequest github

## Synopsis

Report a Github pull request attestation to an artifact or a trail in a Kosli flow.  
It checks if a pull request exists for a given merge commit and reports the pull-request attestation to Kosli.


The attestation can be bound to a *trail* using the trail name.  
The attestation can be bound to an *artifact* in two ways:
- using the artifact's SHA256 fingerprint which is calculated (based on the `--artifact-type` flag and the artifact name/path argument) or can be provided directly (with the `--fingerprint` flag).
- using the artifact's name in the flow yaml template and the git commit from which the artifact is/will be created. Useful when reporting an attestation before creating/reporting the artifact.

```shell
kosli attest pullrequest github [IMAGE-NAME | FILE-PATH | DIR-PATH] [flags]
```

## Flags
| Flag | Description |
| :--- | :--- |
|        --annotate stringToString  |  [optional] Annotate the attestation with data using key=value.  |
|    -t, --artifact-type string  |  The type of the artifact to calculate its SHA256 fingerprint. One of: [oci, docker, file, dir]. Only required if you want Kosli to calculate the fingerprint for you (i.e. when you don't specify '--fingerprint' on commands that allow it).  |
|        --assert  |  [optional] Exit with non-zero code if no pull requests found for the given commit.  |
|        --attachments strings  |  [optional] The comma-separated list of paths of attachments for the reported attestation. Attachments can be files or directories. All attachments are compressed and uploaded to Kosli's evidence vault.  |
|    -g, --commit string  |  the git merge commit to be checked for associated pull requests.  |
|        --description string  |  [optional] attestation description  |
|    -D, --dry-run  |  [optional] Run in dry-run mode. When enabled, no data is sent to Kosli and the CLI exits with 0 exit code regardless of any errors.  |
|    -x, --exclude strings  |  [optional] The comma separated list of directories and files to exclude from fingerprinting. Can take glob patterns. Only applicable for --artifact-type dir.  |
|        --external-fingerprint stringToString  |  [optional] A SHA256 fingerprint of an external attachment represented by --external-url. The format is label=fingerprint (labels cannot contain '.' or '='). This flag can be set multiple times. There must be an external url with a matching label for each external fingerprint.  |
|        --external-url stringToString  |  [optional] Add labeled reference URL for an external resource. The format is label=url (labels cannot contain '.' or '='). This flag can be set multiple times. If the resource is a file or dir, you can optionally add its fingerprint via --external-fingerprint  |
|    -F, --fingerprint string  |  [conditional] The SHA256 fingerprint of the artifact to attach the attestation to. Only required if the attestation is for an artifact and --artifact-type and artifact name/path are not used.  |
|    -f, --flow string  |  The Kosli flow name.  |
|        --github-base-url string  |  [optional] GitHub base URL (only needed for GitHub Enterprise installations).  |
|        --github-org string  |  Github organization. (defaulted if you are running in GitHub Actions: https://docs.kosli.com/ci-defaults ).  |
|        --github-token string  |  Github token.  |
|    -h, --help  |  help for github  |
|    -n, --name string  |  The name of the attestation as declared in the flow or trail yaml template.  |
|    -o, --origin-url string  |  [optional] The url pointing to where the attestation came from or is related. (defaulted to the CI url in some CIs: https://docs.kosli.com/ci-defaults ).  |
|        --redact-commit-info strings  |  [optional] The list of commit info to be redacted before sending to Kosli. Allowed values are one or more of [author, message, branch].  |
|        --registry-password string  |  [conditional] The container registry password or access token. Only required if you want to read container image SHA256 digest from a remote container registry.  |
|        --registry-username string  |  [conditional] The container registry username. Only required if you want to read container image SHA256 digest from a remote container registry.  |
|        --repo-root string  |  [defaulted] The directory where the source git repository is available. Only used if --commit is used. (default ".")  |
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


## Live Examples in different CI systems

{{< tabs "live-examples" "col-no-wrap" >}}{{< tab "GitHub" >}}View an example of the `kosli attest pullrequest github` command in GitHub.

In [this YAML file](https://app.kosli.com/api/v2/livedocs/cyber-dojo/yaml?ci=github&command=kosli+attest+pullrequest+github), which created [this Kosli Event](https://app.kosli.com/api/v2/livedocs/cyber-dojo/event?ci=github&command=kosli+attest+pullrequest+github).{{< /tab >}}{{< /tabs >}}

## Examples Use Cases

These examples all assume that the flags  `--api-token`, `--org`, `--host`, (and `--flow`, `--trail` when required), are set/provided. 

**report a Github pull request attestation about a pre-built docker artifact (kosli calculates the fingerprint)**

```shell
kosli attest pullrequest github yourDockerImageName 
	--artifact-type docker 
	--name yourAttestationName 
	--github-token yourGithubToken 
	--github-org yourGithubOrg 
	--commit yourArtifactGitCommit 
	--repository yourGithubGitRepository 

```

**report a Github pull request attestation about a pre-built docker artifact (you provide the fingerprint)**

```shell
kosli attest pullrequest github 
	--fingerprint yourDockerImageFingerprint 
	--name yourAttestationName 
	--github-token yourGithubToken 
	--github-org yourGithubOrg 
	--commit yourArtifactGitCommit 
	--repository yourGithubGitRepository 

```

**report a Github pull request attestation about a trail**

```shell
kosli attest pullrequest github 
	--name yourAttestationName 
	--github-token yourGithubToken 
	--github-org yourGithubOrg 
	--commit yourArtifactGitCommit 
	--repository yourGithubGitRepository 

```

**report a Github pull request attestation about an artifact which has not been reported yet in a trail**

```shell
kosli attest pullrequest github 
	--name yourTemplateArtifactName.yourAttestationName 
	--github-token yourGithubToken 
	--github-org yourGithubOrg 
	--commit yourArtifactGitCommit 
	--repository yourGithubGitRepository 

```

**report a Github pull request attestation about a trail with an attachment**

```shell
kosli attest pullrequest github 
	--name yourAttestationName 
	--github-token yourGithubToken 
	--github-org yourGithubOrg 
	--commit yourArtifactGitCommit 
	--repository yourGithubGitRepository 
	--attachments=yourAttachmentPathName 

```

**fail if a pull request does not exist for your artifact**

```shell
kosli attest pullrequest github 
	--name yourTemplateArtifactName.yourAttestationName 
	--github-token yourGithubToken 
	--github-org yourGithubOrg 
	--commit yourArtifactGitCommit 
	--repository yourGithubGitRepository 
	--assert
```



## kosli_attest_pullrequest_gitlab.md
---
title: "kosli attest pullrequest gitlab"
beta: false
deprecated: false
summary: "Report a Gitlab merge request attestation to an artifact or a trail in a Kosli flow.  "
---

# kosli attest pullrequest gitlab

## Synopsis

Report a Gitlab merge request attestation to an artifact or a trail in a Kosli flow.  
It checks if a merge request exists for a given merge commit and reports the merge request attestation to Kosli.


The attestation can be bound to a *trail* using the trail name.  
The attestation can be bound to an *artifact* in two ways:
- using the artifact's SHA256 fingerprint which is calculated (based on the `--artifact-type` flag and the artifact name/path argument) or can be provided directly (with the `--fingerprint` flag).
- using the artifact's name in the flow yaml template and the git commit from which the artifact is/will be created. Useful when reporting an attestation before creating/reporting the artifact.

```shell
kosli attest pullrequest gitlab [IMAGE-NAME | FILE-PATH | DIR-PATH] [flags]
```

## Flags
| Flag | Description |
| :--- | :--- |
|        --annotate stringToString  |  [optional] Annotate the attestation with data using key=value.  |
|    -t, --artifact-type string  |  The type of the artifact to calculate its SHA256 fingerprint. One of: [oci, docker, file, dir]. Only required if you want Kosli to calculate the fingerprint for you (i.e. when you don't specify '--fingerprint' on commands that allow it).  |
|        --assert  |  [optional] Exit with non-zero code if no pull requests found for the given commit.  |
|        --attachments strings  |  [optional] The comma-separated list of paths of attachments for the reported attestation. Attachments can be files or directories. All attachments are compressed and uploaded to Kosli's evidence vault.  |
|    -g, --commit string  |  the git merge commit to be checked for associated pull requests.  |
|        --description string  |  [optional] attestation description  |
|    -D, --dry-run  |  [optional] Run in dry-run mode. When enabled, no data is sent to Kosli and the CLI exits with 0 exit code regardless of any errors.  |
|    -x, --exclude strings  |  [optional] The comma separated list of directories and files to exclude from fingerprinting. Can take glob patterns. Only applicable for --artifact-type dir.  |
|        --external-fingerprint stringToString  |  [optional] A SHA256 fingerprint of an external attachment represented by --external-url. The format is label=fingerprint (labels cannot contain '.' or '='). This flag can be set multiple times. There must be an external url with a matching label for each external fingerprint.  |
|        --external-url stringToString  |  [optional] Add labeled reference URL for an external resource. The format is label=url (labels cannot contain '.' or '='). This flag can be set multiple times. If the resource is a file or dir, you can optionally add its fingerprint via --external-fingerprint  |
|    -F, --fingerprint string  |  [conditional] The SHA256 fingerprint of the artifact to attach the attestation to. Only required if the attestation is for an artifact and --artifact-type and artifact name/path are not used.  |
|    -f, --flow string  |  The Kosli flow name.  |
|        --gitlab-base-url string  |  [optional] Gitlab base URL (only needed for on-prem Gitlab installations).  |
|        --gitlab-org string  |  Gitlab organization. (defaulted if you are running in Gitlab Pipelines: https://docs.kosli.com/ci-defaults ).  |
|        --gitlab-token string  |  Gitlab token.  |
|    -h, --help  |  help for gitlab  |
|    -n, --name string  |  The name of the attestation as declared in the flow or trail yaml template.  |
|    -o, --origin-url string  |  [optional] The url pointing to where the attestation came from or is related. (defaulted to the CI url in some CIs: https://docs.kosli.com/ci-defaults ).  |
|        --redact-commit-info strings  |  [optional] The list of commit info to be redacted before sending to Kosli. Allowed values are one or more of [author, message, branch].  |
|        --registry-password string  |  [conditional] The container registry password or access token. Only required if you want to read container image SHA256 digest from a remote container registry.  |
|        --registry-username string  |  [conditional] The container registry username. Only required if you want to read container image SHA256 digest from a remote container registry.  |
|        --repo-root string  |  [defaulted] The directory where the source git repository is available. Only used if --commit is used. (default ".")  |
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


## Live Examples in different CI systems

{{< tabs "live-examples" "col-no-wrap" >}}{{< tab "GitLab" >}}View an example of the `kosli attest pullrequest gitlab` command in GitLab.

In [this YAML file](https://app.kosli.com/api/v2/livedocs/cyber-dojo/yaml?ci=gitlab&command=kosli+attest+pullrequest+gitlab), which created [this Kosli Event](https://app.kosli.com/api/v2/livedocs/cyber-dojo/event?ci=gitlab&command=kosli+attest+pullrequest+gitlab).{{< /tab >}}{{< /tabs >}}

## Examples Use Cases

These examples all assume that the flags  `--api-token`, `--org`, `--host`, (and `--flow`, `--trail` when required), are set/provided. 

**report a Gitlab merge request attestation about a pre-built docker artifact (kosli calculates the fingerprint)**

```shell
kosli attest pullrequest gitlab yourDockerImageName 
	--artifact-type docker 
	--name yourAttestationName 
	--gitlab-token yourGitlabToken 
	--gitlab-org yourGitlabOrg 
	--commit yourArtifactGitCommit 
	--repository yourGithubGitRepository 

```

**report a Gitlab merge request attestation about a pre-built docker artifact (you provide the fingerprint)**

```shell
kosli attest pullrequest gitlab 
	--fingerprint yourDockerImageFingerprint 
	--name yourAttestationName 
	--gitlab-token yourGitlabToken 
	--gitlab-org yourGitlabOrg 
	--commit yourArtifactGitCommit 
	--repository yourGithubGitRepository 

```

**report a Gitlab merge request attestation about a trail**

```shell
kosli attest pullrequest gitlab 
	--name yourAttestationName 
	--gitlab-token yourGitlabToken 
	--gitlab-org yourGitlabOrg 
	--commit yourArtifactGitCommit 
	--repository yourGithubGitRepository 

```

**report a Gitlab merge request attestation about an artifact which has not been reported yet in a trail**

```shell
kosli attest pullrequest gitlab 
	--name yourTemplateArtifactName.yourAttestationName 
	--gitlab-token yourGitlabToken 
	--gitlab-org yourGitlabOrg 
	--commit yourArtifactGitCommit 
	--repository yourGithubGitRepository 

```

**report a Gitlab merge request attestation about a trail with an attachment**

```shell
kosli attest pullrequest gitlab 
	--name yourAttestationName 
	--gitlab-token yourGitlabToken 
	--gitlab-org yourGitlabOrg 
	--commit yourArtifactGitCommit 
	--repository yourGithubGitRepository 
	--attachments=yourAttachmentPathName 

```

**fail if a merge request does not exist for your artifact**

```shell
kosli attest pullrequest gitlab 
	--name yourTemplateArtifactName.yourAttestationName 
	--gitlab-token yourGitlabToken 
	--gitlab-org yourGitlabOrg 
	--commit yourArtifactGitCommit 
	--repository yourGithubGitRepository 
	--assert
```



## kosli_attest_snyk.md
---
title: "kosli attest snyk"
beta: false
deprecated: false
summary: "Report a snyk attestation to an artifact or a trail in a Kosli flow.  "
---

# kosli attest snyk

## Synopsis

Report a snyk attestation to an artifact or a trail in a Kosli flow.  
Only SARIF snyk output is accepted. 
Snyk output can be for "snyk code test", "snyk container test", or "snyk iac test".

The `--scan-results` .json file is analyzed and a summary of the scan results are reported to Kosli.

By default, the `--scan-results` .json file is also uploaded to Kosli's evidence vault.
You can disable that by setting `--upload-results=false`


The attestation can be bound to a *trail* using the trail name.  
The attestation can be bound to an *artifact* in two ways:
- using the artifact's SHA256 fingerprint which is calculated (based on the `--artifact-type` flag and the artifact name/path argument) or can be provided directly (with the `--fingerprint` flag).
- using the artifact's name in the flow yaml template and the git commit from which the artifact is/will be created. Useful when reporting an attestation before creating/reporting the artifact.

You can optionally associate the attestation to a git commit using `--commit` (requires access to a git repo).
You can optionally redact some of the git commit data sent to Kosli using `--redact-commit-info`.
Note that when the attestation is reported for an artifact that does not yet exist in Kosli, `--commit` is required to facilitate
binding the attestation to the right artifact.

```shell
kosli attest snyk [IMAGE-NAME | FILE-PATH | DIR-PATH] [flags]
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

{{< tabs "live-examples" "col-no-wrap" >}}{{< tab "GitHub" >}}View an example of the `kosli attest snyk` command in GitHub.

In [this YAML file](https://app.kosli.com/api/v2/livedocs/cyber-dojo/yaml?ci=github&command=kosli+attest+snyk), which created [this Kosli Event](https://app.kosli.com/api/v2/livedocs/cyber-dojo/event?ci=github&command=kosli+attest+snyk).{{< /tab >}}{{< tab "GitLab" >}}View an example of the `kosli attest snyk` command in GitLab.

In [this YAML file](https://app.kosli.com/api/v2/livedocs/cyber-dojo/yaml?ci=gitlab&command=kosli+attest+snyk), which created [this Kosli Event](https://app.kosli.com/api/v2/livedocs/cyber-dojo/event?ci=gitlab&command=kosli+attest+snyk).{{< /tab >}}{{< /tabs >}}

## Examples Use Cases

These examples all assume that the flags  `--api-token`, `--org`, `--host`, (and `--flow`, `--trail` when required), are set/provided. 

**report a snyk attestation about a pre-built docker artifact (kosli calculates the fingerprint)**

```shell
kosli attest snyk yourDockerImageName 
	--artifact-type docker 
	--name yourAttestationName 
	--scan-results yourSnykSARIFScanResults 

```

**report a snyk attestation about a pre-built docker artifact (you provide the fingerprint)**

```shell
kosli attest snyk 
	--fingerprint yourDockerImageFingerprint 
	--name yourAttestationName 
	--scan-results yourSnykSARIFScanResults 

```

**report a snyk attestation about a trail**

```shell
kosli attest snyk 
	--name yourAttestationName 
	--scan-results yourSnykSARIFScanResults 

```

**report a snyk attestation about an artifact which has not been reported yet in a trail**

```shell
kosli attest snyk 
	--name yourTemplateArtifactName.yourAttestationName 
	--commit yourArtifactGitCommit 
	--scan-results yourSnykSARIFScanResults 

```

**report a snyk attestation about a trail with an attachment**

```shell
kosli attest snyk 
	--name yourAttestationName 
	--scan-results yourSnykSARIFScanResults 
	--attachments yourEvidencePathName 

```

**report a snyk attestation about a trail without uploading the snyk results file**

```shell
kosli attest snyk 
	--name yourAttestationName 
	--scan-results yourSnykSARIFScanResults 
	--upload-results=false 
```



## kosli_attest_sonar.md
---
title: "kosli attest sonar"
beta: false
deprecated: false
summary: "Report a SonarQube attestation to an artifact or a trail in a Kosli flow.  "
---

# kosli attest sonar

## Synopsis

Report a SonarQube attestation to an artifact or a trail in a Kosli flow.  
Retrieves results for the specified scan from SonarQube Cloud or SonarQube Server and attests them to Kosli.
The results are parsed to find the status of the project's quality gate which is used to determine the attestation's compliance status.

The scan to be retrieved can be specified in two ways:
1. (Default) Using metadata created by the Sonar scanner. By default this is located within a temporary .scannerwork folder in the repo base directory.
If you have overriden the location of this folder by passing parameters to the Sonar scanner, or are running Kosli's CLI locally outside the repo's base directory,
you can provide the correct path using the --sonar-working-dir flag. This metadata is generated by a specific scan, allowing Kosli to retrieve the results of that scan.
2. Providing the Sonar project key and the revision of the scan (plus the SonarQube server URL if relevant). If running the Kosli CLI in some CI/CD pipeline, the revision
is defaulted to the commit SHA. If you are running the command locally, or have overriden the revision in SonarQube via parameters to the Sonar scanner, you can
provide the correct revision using the --sonar-revision flag. Kosli then finds the scan results for the specified project key and revision.

Note that if your project is very large and you are using SonarQube Cloud's automatic analysis, it is possible for the attest sonar command to run before the SonarQube Cloud scan is completed.
In this case, we recommend using Kosli's Sonar webhook integration ( https://docs.kosli.com/integrations/sonar/ ) rather than the CLI to attest the scan results.


The attestation can be bound to a *trail* using the trail name.  
The attestation can be bound to an *artifact* in two ways:
- using the artifact's SHA256 fingerprint which is calculated (based on the `--artifact-type` flag and the artifact name/path argument) or can be provided directly (with the `--fingerprint` flag).
- using the artifact's name in the flow yaml template and the git commit from which the artifact is/will be created. Useful when reporting an attestation before creating/reporting the artifact.

```shell
kosli attest sonar [IMAGE-NAME | FILE-PATH | DIR-PATH] [flags]
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
|    -h, --help  |  help for sonar  |
|    -n, --name string  |  The name of the attestation as declared in the flow or trail yaml template.  |
|    -o, --origin-url string  |  [optional] The url pointing to where the attestation came from or is related. (defaulted to the CI url in some CIs: https://docs.kosli.com/ci-defaults ).  |
|        --redact-commit-info strings  |  [optional] The list of commit info to be redacted before sending to Kosli. Allowed values are one or more of [author, message, branch].  |
|        --registry-password string  |  [conditional] The container registry password or access token. Only required if you want to read container image SHA256 digest from a remote container registry.  |
|        --registry-username string  |  [conditional] The container registry username. Only required if you want to read container image SHA256 digest from a remote container registry.  |
|        --repo-root string  |  [defaulted] The directory where the source git repository is available. Only used if --commit is used. (default ".")  |
|        --sonar-api-token string  |  [required] SonarCloud/SonarQube API token.  |
|        --sonar-project-key string  |  [conditional] The project key of the SonarCloud/SonarQube project. Only required if you want to use the project key/revision to get the scan results rather than using Sonar's metadata file.  |
|        --sonar-revision string  |  [conditional] The revision of the SonarCloud/SonarQube project. Only required if you want to use the project key/revision to get the scan results rather than using Sonar's metadata file and you have overridden the default revision, or you aren't using a CI. Defaults to the value of the git commit flag.  |
|        --sonar-server-url string  |  [conditional] The URL of your SonarQube server. Only required if you are using SonarQube and not using SonarQube's metadata file to get scan results. (default "https://sonarcloud.io")  |
|        --sonar-working-dir string  |  [conditional] The base directory of the repo scanned by SonarCloud/SonarQube. Only required if you have overriden the default in the sonar scanner or you are running the CLI locally in a separate folder from the repo. (default ".scannerwork")  |
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

These examples all assume that the flags  `--api-token`, `--org`, `--host`, (and `--flow`, `--trail` when required), are set/provided. 

**report a SonarQube Cloud attestation about a trail using Sonar's metadata**

```shell
kosli attest sonar 
	--name yourAttestationName 
	--sonar-api-token yourSonarAPIToken 
	--sonar-working-dir yourSonarWorkingDirPath 

```

**report a SonarQube Server attestation about a trail using Sonar's metadata**

```shell
kosli attest sonar 
	--name yourAttestationName 
	--sonar-api-token yourSonarAPIToken 
	--sonar-working-dir yourSonarWorkingDirPath 

```

**report a SonarQube Cloud attestation for a specific branch about a trail using key/revision**

```shell
kosli attest sonar 
	--name yourAttestationName 
	--sonar-api-token yourSonarAPIToken 
	--sonar-project-key yourSonarProjectKey 
	--sonar-revision yourSonarRevision 
	--branch-name yourBranchName 

```

**report a SonarQube Server attestation for a pull-request about a trail using key/revision**

```shell
kosli attest sonar 
	--name yourAttestationName 
	--sonar-api-token yourSonarAPIToken 
	--sonarqube-url yourSonarQubeURL 
	--sonar-project-key yourSonarProjectKey 
	--sonar-revision yourSonarRevision 
	--pull-request-id yourPullRequestID 

```

**report a SonarQube Cloud attestation about a trail with an attachment using Sonar's metadata**

```shell
kosli attest sonar 
	--name yourAttestationName 
	--sonar-api-token yourSonarAPIToken 
	--sonar-working-dir yourSonarWorkingDirPath 
	--attachment yourAttachmentPath 
```



## kosli_begin_trail.md
---
title: "kosli begin trail"
beta: false
deprecated: false
summary: "Begin or update a Kosli flow trail."
---

# kosli begin trail

## Synopsis

Begin or update a Kosli flow trail.

You can optionally associate the trail to a git commit using `--commit` (requires access to a git repo). And you  
can optionally redact some of the git commit data sent to Kosli using `--redact-commit-info`.

`TRAIL-NAME`s must start with a letter or number, and only contain letters, numbers, `.`, `-`, `_`, and `~`.


```shell
kosli begin trail TRAIL-NAME [flags]
```

## Flags
| Flag | Description |
| :--- | :--- |
|    -g, --commit string  |  [defaulted] The git commit from which the trail is begun. (defaulted in some CIs: https://docs.kosli.com/ci-defaults, otherwise defaults to HEAD ).  |
|        --description string  |  [optional] The Kosli trail description.  |
|    -D, --dry-run  |  [optional] Run in dry-run mode. When enabled, no data is sent to Kosli and the CLI exits with 0 exit code regardless of any errors.  |
|        --external-fingerprint stringToString  |  [optional] A SHA256 fingerprint of an external attachment represented by --external-url. The format is label=fingerprint (labels cannot contain '.' or '='). This flag can be set multiple times. There must be an external url with a matching label for each external fingerprint.  |
|        --external-url stringToString  |  [optional] Add labeled reference URL for an external resource. The format is label=url (labels cannot contain '.' or '='). This flag can be set multiple times. If the resource is a file or dir, you can optionally add its fingerprint via --external-fingerprint  |
|        --flow string  |  The Kosli flow name.  |
|    -h, --help  |  help for trail  |
|    -o, --origin-url string  |  [optional] The url pointing to where the attestation came from or is related. (defaulted to the CI url in some CIs: https://docs.kosli.com/ci-defaults ).  |
|        --redact-commit-info strings  |  [optional] The list of commit info to be redacted before sending to Kosli. Allowed values are one or more of [author, message, branch].  |
|        --repo-root string  |  [defaulted] The directory where the source git repository is available. Only used if --commit is used. (default ".")  |
|    -f, --template-file string  |  [optional] The path to a yaml template file.  |
|    -u, --user-data string  |  [optional] The path to a JSON file containing additional data you would like to attach to the flow trail.  |


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

{{< tabs "live-examples" "col-no-wrap" >}}{{< tab "GitHub" >}}View an example of the `kosli begin trail` command in GitHub.

In [this YAML file](https://app.kosli.com/api/v2/livedocs/cyber-dojo/yaml?ci=github&command=kosli+begin+trail), which created [this Kosli Event](https://app.kosli.com/api/v2/livedocs/cyber-dojo/event?ci=github&command=kosli+begin+trail).{{< /tab >}}{{< tab "GitLab" >}}View an example of the `kosli begin trail` command in GitLab.

In [this YAML file](https://app.kosli.com/api/v2/livedocs/cyber-dojo/yaml?ci=gitlab&command=kosli+begin+trail), which created [this Kosli Event](https://app.kosli.com/api/v2/livedocs/cyber-dojo/event?ci=gitlab&command=kosli+begin+trail).{{< /tab >}}{{< /tabs >}}

## Examples Use Cases

These examples all assume that the flags  `--api-token`, `--org`, `--host`, (and `--flow`, `--trail` when required), are set/provided. 

**begin/update a Kosli flow trail**

```shell
kosli begin trail yourTrailName 
	--description yourTrailDescription 
	--template-file /path/to/your/template/file.yml 
	--user-data /path/to/your/user-data/file.json 
```



## kosli_completion.md
---
title: "kosli completion"
beta: false
deprecated: false
summary: "Generate completion script"
---

# kosli completion

## Synopsis

To load completions:

  ### Bash

```
  $ source <(kosli completion bash)
```
  To load completions for each session, execute once:  

  On Linux:
  ```
  $ kosli completion bash > /etc/bash_completion.d/kosli
  ``` 
  On macOS:
  ```
  $ kosli completion bash > $(brew --prefix)/etc/bash_completion.d/kosli
  ```
  ### Zsh

  If shell completion is not already enabled in your environment,  
you will need to enable it.  You can execute the following once:
  ```
  $ echo "autoload -U compinit; compinit" >> ~/.zshrc
  ```
  To load completions for each session, execute once:
  ```
  $ kosli completion zsh > "${fpath[1]}/_kosli"
  ```
  You will need to start a new shell for this setup to take effect.

  ### fish
  ```
  $ kosli completion fish | source
  ```
  To load completions for each session, execute once:
  ``` 
  $ kosli completion fish > ~/.config/fish/completions/kosli.fish
  ```
  ### PowerShell
  ```
  PS> kosli completion powershell | Out-String | Invoke-Expression
  ```
 To load completions for every new session, run:
 ```
 PS> kosli completion powershell > kosli.ps1
 ``` 
 and source this file from your PowerShell profile.


```shell
kosli completion [bash|zsh|fish|powershell]
```

## Flags
| Flag | Description |
| :--- | :--- |
|    -h, --help  |  help for completion  |


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




## kosli_config.md
---
title: "kosli config"
beta: false
deprecated: false
summary: "Config global Kosli flags values and store them in $HOME/.kosli .  "
---

# kosli config

## Synopsis

Config global Kosli flags values and store them in $HOME/.kosli .  

Flag values are determined in the following order (highest precedence first):
- command line flags on each executed command.
- environment variables.
- custom config file provided with --config-file flag.
- default config file in $HOME/.kosli

You can configure global Kosli flags (the ones that apply to all/most commands) using their dedicated
convenience flags (e.g. --org). 

API tokens are stored in the suitable credentials manager on your machine. 

Other Kosli flags can be configured using the --set flag which takes a comma-separated list of key=value pairs.
Keys correspond to the specific flag name, capitalized. For instance: --flow would be set using --set FLOW=value


```shell
kosli config [flags]
```

## Flags
| Flag | Description |
| :--- | :--- |
|    -h, --help  |  help for config  |
|        --set stringToString  |  [optional] The key-value pairs to tag the resource with. The format is: key=value  |
|        --unset strings  |  [optional] The list of tag keys to remove from the resource.  |


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

These examples all assume that the flags  `--api-token`, `--org`, `--host`, (and `--flow`, `--trail` when required), are set/provided. 

**configure global flags in your default config file**

```shell
kosli config --org=yourOrg 
	--api-token=yourAPIToken 
	--host=https://app.kosli.com 
	--debug=false 
	--max-api-retries=3 
	--http-proxy=http://192.0.0.1:8080

```

**configure non-global flags in your default config file**

```shell
kosli config --set FLOW=yourFlowName

```

**remove a key from the default config file**

```shell
kosli config --unset FLOW
```



## kosli_create_attestation-type.md
---
title: "kosli create attestation-type"
beta: false
deprecated: false
summary: "Create or update a Kosli custom attestation type."
---

# kosli create attestation-type

## Synopsis

Create or update a Kosli custom attestation type.
You can specify attestation type parameters in flags.

`TYPE-NAME` must start with a letter or number, and only contain letters, numbers, `.`, `-`, `_`, and `~`.

`--schema` is a path to a file containing a JSON schema which will be used to validate attestations made using this type.  
The schema is used to specify the structure of the attestation data, e.g. any fields that are required or 
the expected type of the data.
See an example schema file 
[here](https://github.com/cyber-dojo/kosli-attestation-types/blob/f9130c58d3a8151b0b0e7c5db284e4380eb2d2cf/metrics-coverage.schema.json).

`--jq` defines an evaluation rule, given in jq-format, for this attestation type. The flag can be repeated in order to add additional rules.  
These rules specify acceptable values for attestation data, e.g. `.age >= 21` or `.failing_tests == 0`.  
When a custom attestation is reported, the provided data is evaluated according to the rules defined in its attestation-type. 
All rules must return `true` for the evaluation to pass and the attestation to be determined compliant.


```shell
kosli create attestation-type TYPE-NAME [flags]
```

## Flags
| Flag | Description |
| :--- | :--- |
|    -d, --description string  |  [optional] The attestation type description.  |
|    -D, --dry-run  |  [optional] Run in dry-run mode. When enabled, no data is sent to Kosli and the CLI exits with 0 exit code regardless of any errors.  |
|    -h, --help  |  help for attestation-type  |
|        --jq stringArray  |  [optional] The attestation type evaluation JQ rules.  |
|    -s, --schema string  |  [optional] Path to the attestation type schema in JSON Schema format.  |


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

{{< tabs "live-examples" "col-no-wrap" >}}{{< tab "GitHub" >}}View an example of the `kosli create attestation-type` command in GitHub.

In [this YAML file](https://app.kosli.com/api/v2/livedocs/cyber-dojo/yaml?ci=github&command=kosli+create+attestation-type){{< /tab >}}{{< /tabs >}}

## Examples Use Cases

These examples all assume that the flags  `--api-token`, `--org`, `--host`, (and `--flow`, `--trail` when required), are set/provided. 

**create/update a custom attestation type with no schema no evaluation rules**

```shell
kosli create attestation-type customTypeName

```

**create/update a custom attestation type with schema and jq evaluation rules**

```shell
kosli create attestation-type customTypeName 
    --description "Attest that a person meets the age requirements." 
    --schema person-schema.json 
    --jq ".age >= 18"
    --jq ".age < 65"
```



## kosli_create_environment.md
---
title: "kosli create environment"
beta: false
deprecated: false
summary: "Create or update a Kosli environment."
---

# kosli create environment

## Synopsis

Create or update a Kosli environment.

``--type`` must match the type of environment you wish to record snapshots from.
The following types are supported:
  - k8s        - Kubernetes
  - ecs        - Amazon Elastic Container Service
  - s3         - Amazon S3 object storage
  - lambda     - AWS Lambda serverless
  - docker     - Docker images
  - azure-apps - Azure app services
  - server     - Generic type
  - logical    - Logical grouping of real environments

By default, the environment does not require artifacts provenance (i.e. environment snapshots will not 
become non-compliant because of artifacts that do not have provenance). You can require provenance for all artifacts
by setting --require-provenance=true

Also, by default, kosli will not make new snapshots for scaling events (change in number of instances running).
For large clusters the scaling events will often outnumber the actual change of SW.

It is possible to enable new snapshots for scaling events with the --include-scaling flag, or turn
it off again with the --exclude-scaling.

Logical environments are used for grouping of physical environments. For instance **prod-aws** and **prod-s3** can
be grouped into logical environment **prod**. Logical environments are view-only, you can not report snapshots
to them.

`ENVIRONMENT-NAME`s must start with a letter or number, and only contain letters, numbers, `.`, `-`, `_`, and `~`.


```shell
kosli create environment ENVIRONMENT-NAME [flags]
```

## Flags
| Flag | Description |
| :--- | :--- |
|    -d, --description string  |  [optional] The environment description.  |
|    -D, --dry-run  |  [optional] Run in dry-run mode. When enabled, no data is sent to Kosli and the CLI exits with 0 exit code regardless of any errors.  |
|        --exclude-scaling  |  [optional] Exclude scaling events for snapshots. Snapshots with scaling changes will not result in new environment records.  |
|    -h, --help  |  help for environment  |
|        --include-scaling  |  [optional] Include scaling events for snapshots. Snapshots with scaling changes will result in new environment records.  |
|        --included-environments strings  |  [optional] Comma separated list of environments to include in logical environment  |
|        --require-provenance  |  [defaulted] Require provenance for all artifacts running in environment snapshots.  |
|    -t, --type string  |  The type of environment. Valid types are: [K8S, ECS, server, S3, lambda, docker, azure-apps, logical].  |


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

These examples all assume that the flags  `--api-token`, `--org`, `--host`, (and `--flow`, `--trail` when required), are set/provided. 

**create a Kosli environment**

```shell
kosli create environment yourEnvironmentName
	--type K8S 
	--description "my new env" 


kosli create environment yourLogicalEnvironmentName
	--type logical 
	--included-environments realEnv1,realEnv2,realEnv3
	--description "my full prod" 
```



## kosli_create_flow.md
---
title: "kosli create flow"
beta: false
deprecated: false
summary: "Create or update a Kosli flow."
---

# kosli create flow

## Synopsis

Create or update a Kosli flow.
You can specify flow parameters in flags.

`FLOW-NAME`s must start with a letter or number, and only contain letters, numbers, `.`, `-`, `_`, and `~`.


```shell
kosli create flow FLOW-NAME [flags]
```

## Flags
| Flag | Description |
| :--- | :--- |
|        --description string  |  [optional] The Kosli flow description.  |
|    -D, --dry-run  |  [optional] Run in dry-run mode. When enabled, no data is sent to Kosli and the CLI exits with 0 exit code regardless of any errors.  |
|    -h, --help  |  help for flow  |
|    -t, --template strings  |  [defaulted] The comma-separated list of required compliance controls names.  |
|    -f, --template-file string  |  [optional] The path to a yaml template file. Cannot be used together with --use-empty-template  |
|        --use-empty-template  |  Use an empty template for the flow creation without specifying a file. Cannot be used together with --template or --template-file  |
|        --visibility string  |  [defaulted] The visibility of the Kosli flow. Valid visibilities are [public, private]. (default "private")  |


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

{{< tabs "live-examples" "col-no-wrap" >}}{{< tab "GitHub" >}}View an example of the `kosli create flow` command in GitHub.

In [this YAML file](https://app.kosli.com/api/v2/livedocs/cyber-dojo/yaml?ci=github&command=kosli+create+flow), which created [this Kosli Event](https://app.kosli.com/api/v2/livedocs/cyber-dojo/event?ci=github&command=kosli+create+flow).{{< /tab >}}{{< tab "GitLab" >}}View an example of the `kosli create flow` command in GitLab.

In [this YAML file](https://app.kosli.com/api/v2/livedocs/cyber-dojo/yaml?ci=gitlab&command=kosli+create+flow), which created [this Kosli Event](https://app.kosli.com/api/v2/livedocs/cyber-dojo/event?ci=gitlab&command=kosli+create+flow).{{< /tab >}}{{< /tabs >}}

## Examples Use Cases

These examples all assume that the flags  `--api-token`, `--org`, `--host`, (and `--flow`, `--trail` when required), are set/provided. 

**create/update a Kosli flow (with empty template)**

```shell
kosli create flow yourFlowName 
	--description yourFlowDescription 
	--visibility private OR public 
	--use-empty-template 

```

**create/update a Kosli flow (with template file)**

```shell
kosli create flow yourFlowName 
	--description yourFlowDescription 
	--visibility private OR public 
	--template-file /path/to/your/template/file.yml 
```



## kosli_create_policy.md
---
title: "kosli create policy"
beta: false
deprecated: false
summary: "Create or update a Kosli policy."
---

# kosli create policy

## Synopsis

Updating policy content creates a new version of the policy.

```shell
kosli create policy POLICY-NAME POLICY-FILE-PATH [flags]
```

## Flags
| Flag | Description |
| :--- | :--- |
|        --comment string  |  [optional] comment about the change made in a policy file when updating a policy.  |
|        --description string  |  [optional] policy description.  |
|    -D, --dry-run  |  [optional] Run in dry-run mode. When enabled, no data is sent to Kosli and the CLI exits with 0 exit code regardless of any errors.  |
|    -h, --help  |  help for policy  |
|        --type string  |  [defaulted] the type of policy. One of: [env] (default "env")  |


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

These examples all assume that the flags  `--api-token`, `--org`, `--host`, (and `--flow`, `--trail` when required), are set/provided. 

**create a Kosli policy**

```shell
kosli create policy yourPolicyName yourPolicyFile.yml 
	--description yourPolicyDescription 
	--type env 

```

**update a Kosli policy**

```shell
kosli create policy yourPolicyName yourPolicyFile.yml 
	--description yourPolicyDescription 
	--type env 
	--comment yourChangeComment 
```



## kosli_detach-policy.md
---
title: "kosli detach-policy"
beta: false
deprecated: false
summary: "Detach a policy from one or more Kosli environments.  "
---

# kosli detach-policy

## Synopsis

If the environment has no more policies attached to it, then its snapshots' status will become "unknown".

```shell
kosli detach-policy POLICY-NAME [flags]
```

## Flags
| Flag | Description |
| :--- | :--- |
|    -D, --dry-run  |  [optional] Run in dry-run mode. When enabled, no data is sent to Kosli and the CLI exits with 0 exit code regardless of any errors.  |
|    -e, --environment strings  |  the list of environment names to detach the policy from  |
|    -h, --help  |  help for detach-policy  |


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

These examples all assume that the flags  `--api-token`, `--org`, `--host`, (and `--flow`, `--trail` when required), are set/provided. 

**detach policy from multiple environment**

```shell
kosli detach-policy yourPolicyName 
	--environment yourFirstEnvironmentName 
	--environment yourSecondEnvironmentName 
```



## kosli_diff_snapshots.md
---
title: "kosli diff snapshots"
beta: false
deprecated: false
summary: "Diff environment snapshots.  "
---

# kosli diff snapshots

## Synopsis

Diff environment snapshots.  
Specify SNAPPISH_1 and SNAPPISH_2 by:
- environmentName
    - the latest snapshot for environmentName, at the time of the request
    - e.g., **prod**
- environmentName#N
    - the Nth snapshot, counting from 1
    - e.g., **prod#42**
- environmentName~N
    - the Nth snapshot behind the latest, at the time of the request
    - e.g., **prod~5**
- environmentName@{YYYY-MM-DDTHH:MM:SS}
    - the snapshot at specific moment in time in UTC
    - e.g., **prod@{2023-10-02T12:00:00}**
- environmentName@{N.<hours|days|weeks|months>.ago}
    - the snapshot at a time relative to the time of the request
    - e.g., **prod@{2.hours.ago}**


```shell
kosli diff snapshots SNAPPISH_1 SNAPPISH_2 [flags]
```

## Flags
| Flag | Description |
| :--- | :--- |
|    -h, --help  |  help for snapshots  |
|    -o, --output string  |  [defaulted] The format of the output. Valid formats are: [table, json]. (default "table")  |
|    -u, --show-unchanged  |  [defaulted] Show the unchanged artifacts present in both snapshots within the diff output.  |


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


## Live Example

{{< raw-html >}}To view a live example of 'kosli diff snapshots' you can run the commands below (for the <a href="https://app.kosli.com/cyber-dojo/environments/aws-prod/snapshots/">cyber-dojo</a> demo organization).<br/><a href="https://app.kosli.com/api/v2/livedocs/cyber-dojo/cli?command=kosli+diff+snapshots+aws-beta+aws-prod+--output=json">Run the commands below and view the output.</a><pre>export KOSLI_ORG=cyber-dojo
export KOSLI_API_TOKEN=Pj_XT2deaVA6V1qrTlthuaWsmjVt4eaHQwqnwqjRO3A  # read-only
kosli diff snapshots aws-beta aws-prod --output=json</pre>{{< / raw-html >}}

## Examples Use Cases

These examples all assume that the flags  `--api-token`, `--org`, `--host`, (and `--flow`, `--trail` when required), are set/provided. 

**compare the third latest snapshot in an environment to the latest**

```shell
kosli diff snapshots envName~3 envName 

```

**compare snapshots of two different environments of the same type**

```shell
kosli diff snapshots envName1 envName2 

```

**show the not-changed artifacts in both snapshots**

```shell
kosli diff snapshots envName1 envName2 
	--show-unchanged 

```

**compare the snapshot from 2 weeks ago in an environment to the latest**

```shell
kosli diff snapshots envName@{2.weeks.ago} envName 
```



## kosli_disable_beta.md
---
title: "kosli disable beta"
beta: false
deprecated: false
summary: "Disable beta features for an organization."
---

# kosli disable beta

## Synopsis

Disable beta features for an organization.

```shell
kosli disable beta [flags]
```

## Flags
| Flag | Description |
| :--- | :--- |
|    -h, --help  |  help for beta  |


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




## kosli_enable_beta.md
---
title: "kosli enable beta"
beta: false
deprecated: false
summary: "Enable beta features for an organization."
---

# kosli enable beta

## Synopsis

Enable beta features for an organization.

```shell
kosli enable beta [flags]
```

## Flags
| Flag | Description |
| :--- | :--- |
|    -h, --help  |  help for beta  |


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




## kosli_expect_deployment.md
---
title: "kosli expect deployment"
beta: false
deprecated: true
summary: "Report the expectation of an upcoming deployment of an artifact to an environment.  "
---

# kosli expect deployment

{{% hint danger %}}
**kosli expect deployment** is deprecated. deployment expectation is no longer required for compliance.  Deprecated commands will be removed in a future release.
{{% /hint %}}
## Synopsis

Report the expectation of an upcoming deployment of an artifact to an environment.  

The artifact fingerprint can be provided directly with the `--fingerprint` flag, or 
calculated based on `--artifact-type` flag.

Artifact type can be one of: "file" for files, "dir" for directories, "oci" for container
images in registries or "docker" for local docker images.



```shell
kosli expect deployment [IMAGE-NAME | FILE-PATH | DIR-PATH] [flags]
```

## Flags
| Flag | Description |
| :--- | :--- |
|    -t, --artifact-type string  |  The type of the artifact to calculate its SHA256 fingerprint. One of: [oci, docker, file, dir]. Only required if you want Kosli to calculate the fingerprint for you (i.e. when you don't specify '--fingerprint' on commands that allow it).  |
|    -b, --build-url string  |  The url of CI pipeline that built the artifact. (defaulted in some CIs: https://docs.kosli.com/ci-defaults ).  |
|    -d, --description string  |  [optional] The deployment description.  |
|    -D, --dry-run  |  [optional] Run in dry-run mode. When enabled, no data is sent to Kosli and the CLI exits with 0 exit code regardless of any errors.  |
|    -e, --environment string  |  The environment name.  |
|    -x, --exclude strings  |  [optional] The comma separated list of directories and files to exclude from fingerprinting. Can take glob patterns. Only applicable for --artifact-type dir.  |
|    -F, --fingerprint string  |  [conditional] The SHA256 fingerprint of the artifact. Only required if you don't specify '--artifact-type'.  |
|    -f, --flow string  |  The Kosli flow name.  |
|    -h, --help  |  help for deployment  |
|        --registry-password string  |  [conditional] The container registry password or access token. Only required if you want to read container image SHA256 digest from a remote container registry.  |
|        --registry-username string  |  [conditional] The container registry username. Only required if you want to read container image SHA256 digest from a remote container registry.  |
|    -u, --user-data string  |  [optional] The path to a JSON file containing additional data you would like to attach to the deployment.  |


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




## kosli_fingerprint.md
---
title: "kosli fingerprint"
beta: false
deprecated: false
summary: "Calculate the SHA256 fingerprint of an artifact."
---

# kosli fingerprint

## Synopsis

Calculate the SHA256 fingerprint of an artifact.
Requires `--artifact-type` flag to be set.
Artifact type can be one of: "file" for files, "dir" for directories, "oci" for container
images in registries or "docker" for local docker images.

Fingerprinting container images can be done using the local docker daemon or the fingerprint can be fetched
from a remote registry.

When fingerprinting a 'dir' artifact, you can exclude certain paths from fingerprint calculation 
using the `--exclude` flag.
Excluded paths are relative to the DIR-PATH and can be literal paths or glob patterns.
With a directory structure like this `foo/bar/zam/file.txt` if you are calculating the fingerprint of `foo/bar` you need to
exclude `zam/file.txt` which is relative to the DIR-PATH.
The supported glob pattern syntax is what is documented here: https://pkg.go.dev/path/filepath#Match , 
plus the ability to use recursive globs "**"

To specify paths in a directory artifact that should always be excluded from the SHA256 calculation, you can add a `.kosli_ignore` file to the root of the artifact.
Each line should specify a relative path or path glob to be ignored. You can include comments in this file, using `#`.
The `.kosli_ignore` will be treated as part of the artifact like any other file, unless it is explicitly ignored itself.

```shell
kosli fingerprint {IMAGE-NAME | FILE-PATH | DIR-PATH} [flags]
```

## Flags
| Flag | Description |
| :--- | :--- |
|    -t, --artifact-type string  |  The type of the artifact to calculate its SHA256 fingerprint. One of: [oci, docker, file, dir]. Only required if you want Kosli to calculate the fingerprint for you (i.e. when you don't specify '--fingerprint' on commands that allow it).  |
|    -x, --exclude strings  |  [optional] The comma separated list of directories and files to exclude from fingerprinting. Can take glob patterns. Only applicable for --artifact-type dir.  |
|    -h, --help  |  help for fingerprint  |
|        --registry-password string  |  [conditional] The container registry password or access token. Only required if you want to read container image SHA256 digest from a remote container registry.  |
|        --registry-username string  |  [conditional] The container registry username. Only required if you want to read container image SHA256 digest from a remote container registry.  |


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

{{< tabs "live-examples" "col-no-wrap" >}}{{< tab "GitHub" >}}View an example of the `kosli fingerprint` command in GitHub.

In [this YAML file](https://app.kosli.com/api/v2/livedocs/cyber-dojo/yaml?ci=github&command=kosli+fingerprint){{< /tab >}}{{< /tabs >}}

## Examples Use Cases

These examples all assume that the flags  `--api-token`, `--org`, `--host`, (and `--flow`, `--trail` when required), are set/provided. 

**fingerprint a file**

```shell
kosli fingerprint --artifact-type file file.txt

```

**fingerprint a dir**

```shell
kosli fingerprint --artifact-type dir mydir

```

**fingerprint a dir while excluding paths ^mydir/logs^ and ^mydir/*exe^**

```shell
kosli fingerprint --artifact-type dir --exclude logs --exclude *.exe mydir

```

**fingerprint a dir while excluding all ^.pyc^ files**

```shell
kosli fingerprint --artifact-type dir  --exclude **/*.pyc mydir

```

**fingerprint a dir while excluding paths in .kosli_ignore file**

```shell
echo bar/file.txt > mydir/.kosli_ignore
kosli fingerprint --artifact-type dir mydir

```

**fingerprint a locally available docker image (requires docker daemon running)**

```shell
kosli fingerprint --artifact-type docker nginx:latest

```

**fingerprint a public image from a remote registry**

```shell
kosli fingerprint --artifact-type oci nginx:latest

```

**fingerprint a private image from a remote registry**

```shell
kosli fingerprint --artifact-type oci private:latest --registry-username YourUsername --registry-password YourPassword
```



## kosli_get_approval.md
---
title: "kosli get approval"
beta: false
deprecated: false
summary: "Get an approval from a specified flow."
---

# kosli get approval

## Synopsis

Get an approval from a specified flow.
EXPRESSION can be specified as follows:
- flowName
    - the latest approval to flowName, at the time of the request
    - e.g., **creator**
- flowName#N
    - the Nth approval, counting from 1
    - e.g., **creator#453**
- flowName~N
    - the Nth approval behind the latest, at the time of the request
    - e.g., **creator~56**


```shell
kosli get approval EXPRESSION [flags]
```

## Flags
| Flag | Description |
| :--- | :--- |
|    -h, --help  |  help for approval  |
|    -o, --output string  |  [defaulted] The format of the output. Valid formats are: [table, json]. (default "table")  |


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

These examples all assume that the flags  `--api-token`, `--org`, `--host`, (and `--flow`, `--trail` when required), are set/provided. 

**get second behind the latest approval from a flow**

```shell
kosli get approval flowName~1 

```

**get the 10th approval from a flow**

```shell
kosli get approval flowName#10 

```

**get the latest approval from a flow**

```shell
kosli get approval flowName 
```



## kosli_get_artifact.md
---
title: "kosli get artifact"
beta: false
deprecated: false
summary: "Get artifact from a specified flow"
---

# kosli get artifact

## Synopsis

Get artifact from a specified flow
You can get an artifact by its fingerprint or by its git commit sha.
In case of using the git commit, it is possible to get multiple artifacts matching the git commit.

The expected argument is an expression to specify the artifact to get.
It has the format <FLOW_NAME><SEPARATOR><COMMIT_SHA1|ARTIFACT_FINGERPRINT> 

Expression can be specified as follows:
- flowName@<fingerprint>  artifact with a given fingerprint. The fingerprint can be short or complete.
- flowName:<commit_sha>   artifact with a given commit SHA. The commit sha can be short or complete.

Examples of valid expressions are:
- flow@184c799cd551dd1d8d5c5f9a5d593b2e931f5e36122ee5c793c1d08a19839cc0
- flow@184c7
- flow:110d048bf1fce72ba546cbafc4427fb21b958dee
- flow:110d0


```shell
kosli get artifact EXPRESSION [flags]
```

## Flags
| Flag | Description |
| :--- | :--- |
|    -h, --help  |  help for artifact  |
|    -o, --output string  |  [defaulted] The format of the output. Valid formats are: [table, json]. (default "table")  |
|    -t, --trail string  |  [optional] The Kosli trail name.  |


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

These examples all assume that the flags  `--api-token`, `--org`, `--host`, (and `--flow`, `--trail` when required), are set/provided. 

**get an artifact with a given fingerprint from a flow**

```shell
kosli get artifact flowName@fingerprint 

```

**get the latest artifact with a given fingerprint from a flow in a specific trail**

```shell
kosli get artifact flowName@fingerprint 

```

**get an artifact with a given commit SHA from a flow**

```shell
kosli get artifact flowName:commitSHA 

```

**get a list of artifacts with a given commit SHA from a flow in a particular trail**

```shell
kosli get artifact flowName:commitSHA 
```



## kosli_get_attestation-type.md
---
title: "kosli get attestation-type"
beta: false
deprecated: false
summary: "Get a custom Kosli attestation type.  "
---

# kosli get attestation-type

## Synopsis

Get a custom Kosli attestation type.  
The TYPE-NAME can be specified as follows:
- customTypeName
	- Returns the unversioned custom attestation type, containing details of all versions of the type.
	- e.g. `custom-type`
- customTypeName@vN
	- Returns the Nth version of the custom attestation type.
	- If a non-integer version number is given, the unversioned custom attestation type is returned.
	- e.g. `custom-type@v4`


```shell
kosli get attestation-type TYPE-NAME [flags]
```

## Flags
| Flag | Description |
| :--- | :--- |
|    -h, --help  |  help for attestation-type  |
|    -o, --output string  |  [defaulted] The format of the output. Valid formats are: [table, json]. (default "table")  |


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

These examples all assume that the flags  `--api-token`, `--org`, `--host`, (and `--flow`, `--trail` when required), are set/provided. 

**get an unversioned custom attestation type**

```shell
kosli get attestation-type customTypeName

```

**get version 1 of a custom attestation type**

```shell
kosli get attestation-type customTypeName@v1
```



## kosli_get_deployment.md
---
title: "kosli get deployment"
beta: false
deprecated: false
summary: "Get a deployment from a specified flow."
---

# kosli get deployment

## Synopsis

Get a deployment from a specified flow.
EXPRESSION can be specified as follows:
- flowName
    - the latest deployment to flowName, at the time of the request
    - e.g., **dashboard**
- flowName#N
    - the Nth deployment, counting from 1
    - e.g., **dashboard#453**
- flowName~N
    - the Nth deployment behind the latest, at the time of the request
    - e.g., **dashboard~56**


```shell
kosli get deployment EXPRESSION [flags]
```

## Flags
| Flag | Description |
| :--- | :--- |
|    -h, --help  |  help for deployment  |
|    -o, --output string  |  [defaulted] The format of the output. Valid formats are: [table, json]. (default "table")  |


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

These examples all assume that the flags  `--api-token`, `--org`, `--host`, (and `--flow`, `--trail` when required), are set/provided. 

**get previous deployment in a flow**

```shell
kosli get deployment flowName~1 

```

**get the 10th deployment in a flow**

```shell
kosli get deployment flowName#10 

```

**get the latest deployment in a flow**

```shell
kosli get deployment flowName 
```



## kosli_get_environment.md
---
title: "kosli get environment"
beta: false
deprecated: false
summary: "Get an environment's metadata."
---

# kosli get environment

## Synopsis

Get an environment's metadata.

```shell
kosli get environment ENVIRONMENT-NAME [flags]
```

## Flags
| Flag | Description |
| :--- | :--- |
|    -h, --help  |  help for environment  |
|    -o, --output string  |  [defaulted] The format of the output. Valid formats are: [table, json]. (default "table")  |


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


## Live Example

{{< raw-html >}}To view a live example of 'kosli get environment' you can run the commands below (for the <a href="https://app.kosli.com/cyber-dojo/environments/aws-prod/snapshots/">cyber-dojo</a> demo organization).<br/><a href="https://app.kosli.com/api/v2/livedocs/cyber-dojo/cli?command=kosli+get+environment+aws-prod+--output=json">Run the commands below and view the output.</a><pre>export KOSLI_ORG=cyber-dojo
export KOSLI_API_TOKEN=Pj_XT2deaVA6V1qrTlthuaWsmjVt4eaHQwqnwqjRO3A  # read-only
kosli get environment aws-prod --output=json</pre>{{< / raw-html >}}



## kosli_get_flow.md
---
title: "kosli get flow"
beta: false
deprecated: false
summary: "Get the metadata of a specific flow."
---

# kosli get flow

## Synopsis

Get the metadata of a specific flow.

```shell
kosli get flow FLOW-NAME [flags]
```

## Flags
| Flag | Description |
| :--- | :--- |
|    -h, --help  |  help for flow  |
|    -o, --output string  |  [defaulted] The format of the output. Valid formats are: [table, json]. (default "table")  |


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


## Live Example

{{< raw-html >}}To view a live example of 'kosli get flow' you can run the commands below (for the <a href="https://app.kosli.com/cyber-dojo/environments/aws-prod/snapshots/">cyber-dojo</a> demo organization).<br/><a href="https://app.kosli.com/api/v2/livedocs/cyber-dojo/cli?command=kosli+get+flow+dashboard-ci+--output=json">Run the commands below and view the output.</a><pre>export KOSLI_ORG=cyber-dojo
export KOSLI_API_TOKEN=Pj_XT2deaVA6V1qrTlthuaWsmjVt4eaHQwqnwqjRO3A  # read-only
kosli get flow dashboard-ci --output=json</pre>{{< / raw-html >}}



## kosli_get_policy.md
---
title: "kosli get policy"
beta: false
deprecated: false
summary: "Get a policy's metadata."
---

# kosli get policy

## Synopsis

Get a policy's metadata.

```shell
kosli get policy POLICY-NAME [flags]
```

## Flags
| Flag | Description |
| :--- | :--- |
|    -h, --help  |  help for policy  |
|    -o, --output string  |  [defaulted] The format of the output. Valid formats are: [table, json]. (default "table")  |


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




## kosli_get_snapshot.md
---
title: "kosli get snapshot"
beta: false
deprecated: false
summary: "Get a specified environment snapshot.  "
---

# kosli get snapshot

## Synopsis

Get a specified environment snapshot.  
ENVIRONMENT-NAME-OR-EXPRESSION can be specified as follows:
- environmentName
    - the latest snapshot for environmentName, at the time of the request
    - e.g., **prod**
- environmentName#N
    - the Nth snapshot, counting from 1
    - e.g., **prod#42**
- environmentName~N
    - the Nth snapshot behind the latest, at the time of the request
    - e.g., **prod~5**
- environmentName@{YYYY-MM-DDTHH:MM:SS}
    - the snapshot at specific moment in time in UTC
    - e.g., **prod@{2023-10-02T12:00:00}**
- environmentName@{N.<hours|days|weeks|months>.ago}
    - the snapshot at a time relative to the time of the request
    - e.g., **prod@{2.hours.ago}**


```shell
kosli get snapshot ENVIRONMENT-NAME-OR-EXPRESSION [flags]
```

## Flags
| Flag | Description |
| :--- | :--- |
|    -h, --help  |  help for snapshot  |
|    -o, --output string  |  [defaulted] The format of the output. Valid formats are: [table, json]. (default "table")  |


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


## Live Example

{{< raw-html >}}To view a live example of 'kosli get snapshot' you can run the commands below (for the <a href="https://app.kosli.com/cyber-dojo/environments/aws-prod/snapshots/">cyber-dojo</a> demo organization).<br/><a href="https://app.kosli.com/api/v2/livedocs/cyber-dojo/cli?command=kosli+get+snapshot+aws-prod+--output=json">Run the commands below and view the output.</a><pre>export KOSLI_ORG=cyber-dojo
export KOSLI_API_TOKEN=Pj_XT2deaVA6V1qrTlthuaWsmjVt4eaHQwqnwqjRO3A  # read-only
kosli get snapshot aws-prod --output=json</pre>{{< / raw-html >}}

## Examples Use Cases

These examples all assume that the flags  `--api-token`, `--org`, `--host`, (and `--flow`, `--trail` when required), are set/provided. 

**get the latest snapshot of an environment**

```shell
kosli get snapshot yourEnvironmentName

```

**get the SECOND latest snapshot of an environment**

```shell
kosli get snapshot yourEnvironmentName~1

```

**get the snapshot number 23 of an environment**

```shell
kosli get snapshot yourEnvironmentName#23

```

**get the environment snapshot at midday (UTC), on valentine's day of 2023**

```shell
kosli get snapshot yourEnvironmentName@{2023-02-14T12:00:00}

```

**get the environment snapshot based on a relative time**

```shell
kosli get snapshot yourEnvironmentName@{3.weeks.ago}
```



## kosli_get_trail.md
---
title: "kosli get trail"
beta: false
deprecated: false
summary: "Get the metadata of a specific trail."
---

# kosli get trail

## Synopsis

Get the metadata of a specific trail.

```shell
kosli get trail TRAIL-NAME [flags]
```

## Flags
| Flag | Description |
| :--- | :--- |
|    -f, --flow string  |  The Kosli flow name.  |
|    -h, --help  |  help for trail  |
|    -o, --output string  |  [defaulted] The format of the output. Valid formats are: [table, json]. (default "table")  |


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


## Live Example

{{< raw-html >}}To view a live example of 'kosli get trail' you can run the commands below (for the <a href="https://app.kosli.com/cyber-dojo/environments/aws-prod/snapshots/">cyber-dojo</a> demo organization).<br/><a href="https://app.kosli.com/api/v2/livedocs/cyber-dojo/cli?command=kosli+get+trail+dashboard-ci+1159a6f1193150681b8484545150334e89de6c1c+--output=json">Run the commands below and view the output.</a><pre>export KOSLI_ORG=cyber-dojo
export KOSLI_API_TOKEN=Pj_XT2deaVA6V1qrTlthuaWsmjVt4eaHQwqnwqjRO3A  # read-only
kosli get trail dashboard-ci 1159a6f1193150681b8484545150334e89de6c1c --output=json</pre>{{< / raw-html >}}



## kosli_join_environment.md
---
title: "kosli join environment"
beta: false
deprecated: false
summary: "Join a physical environment to a logical environment."
---

# kosli join environment

## Synopsis

Join a physical environment to a logical environment.

```shell
kosli join environment [flags]
```

## Flags
| Flag | Description |
| :--- | :--- |
|    -D, --dry-run  |  [optional] Run in dry-run mode. When enabled, no data is sent to Kosli and the CLI exits with 0 exit code regardless of any errors.  |
|    -h, --help  |  help for environment  |
|        --logical string  |  [required] The logical environment.  |
|        --physical string  |  [required] The physical environment.  |


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

These examples all assume that the flags  `--api-token`, `--org`, `--host`, (and `--flow`, `--trail` when required), are set/provided. 

**join a physical environment to a logical environment**

```shell
kosli join environment 
	--physical prod-k8 
	--logical prod 
```



## kosli_list_approvals.md
---
title: "kosli list approvals"
beta: false
deprecated: false
summary: "List approvals in a flow."
---

# kosli list approvals

## Synopsis

List approvals in a flow.
The results are paginated and ordered from latest to oldest.
By default, the page limit is 15 approvals per page.  


```shell
kosli list approvals [flags]
```

## Flags
| Flag | Description |
| :--- | :--- |
|    -f, --flow string  |  The Kosli flow name.  |
|    -h, --help  |  help for approvals  |
|    -o, --output string  |  [defaulted] The format of the output. Valid formats are: [table, json]. (default "table")  |
|        --page int  |  [defaulted] The page number of a response. (default 1)  |
|    -n, --page-limit int  |  [defaulted] The number of elements per page. (default 15)  |


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

These examples all assume that the flags  `--api-token`, `--org`, `--host`, (and `--flow`, `--trail` when required), are set/provided. 

**list the last 15 approvals for a flow**

```shell
kosli list approvals 

```

**list the last 30 approvals for a flow**

```shell
kosli list approvals 
	--page-limit 30 

```

**list the last 30 approvals for a flow (in JSON)**

```shell
kosli list approvals 
	--page-limit 30 
	--output json
```



## kosli_list_artifacts.md
---
title: "kosli list artifacts"
beta: false
deprecated: false
summary: "List artifacts in a flow. "
---

# kosli list artifacts

## Synopsis

List artifacts in a flow. The results are paginated and ordered from latest to oldest.
By default, the page limit is 15 artifacts per page.


```shell
kosli list artifacts [flags]
```

## Flags
| Flag | Description |
| :--- | :--- |
|    -f, --flow string  |  The Kosli flow name.  |
|    -h, --help  |  help for artifacts  |
|    -o, --output string  |  [defaulted] The format of the output. Valid formats are: [table, json]. (default "table")  |
|        --page int  |  [defaulted] The page number of a response. (default 1)  |
|    -n, --page-limit int  |  [defaulted] The number of elements per page. (default 15)  |


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

These examples all assume that the flags  `--api-token`, `--org`, `--host`, (and `--flow`, `--trail` when required), are set/provided. 

**list the last 15 artifacts for a flow**

```shell
kosli list artifacts 

```

**list the last 30 artifacts for a flow**

```shell
kosli list artifacts 
	--page-limit 30 

```

**list the last 30 artifacts for a flow (in JSON)**

```shell
kosli list artifacts 
	--page-limit 30 
	--output json
```



## kosli_list_attestation-types.md
---
title: "kosli list attestation-types"
beta: false
deprecated: false
summary: "List all Kosli attestation types for an org."
---

# kosli list attestation-types

## Synopsis

List all Kosli attestation types for an org.

```shell
kosli list attestation-types [flags]
```

## Flags
| Flag | Description |
| :--- | :--- |
|    -h, --help  |  help for attestation-types  |
|    -o, --output string  |  [defaulted] The format of the output. Valid formats are: [table, json]. (default "table")  |


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




## kosli_list_deployments.md
---
title: "kosli list deployments"
beta: false
deprecated: false
summary: "List deployments in a flow."
---

# kosli list deployments

## Synopsis

List deployments in a flow.
The results are paginated and ordered from latest to oldest.
By default, the page limit is 15 deployments per page.


```shell
kosli list deployments [flags]
```

## Flags
| Flag | Description |
| :--- | :--- |
|    -f, --flow string  |  The Kosli flow name.  |
|    -h, --help  |  help for deployments  |
|    -o, --output string  |  [defaulted] The format of the output. Valid formats are: [table, json]. (default "table")  |
|        --page int  |  [defaulted] The page number of a response. (default 1)  |
|    -n, --page-limit int  |  [defaulted] The number of elements per page. (default 15)  |


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

These examples all assume that the flags  `--api-token`, `--org`, `--host`, (and `--flow`, `--trail` when required), are set/provided. 

**list the last 15 deployments for a flow**

```shell
kosli list deployments 

```

**list the last 30 deployments for a flow**

```shell
kosli list deployments 
	--page-limit 30 

```

**list the last 30 deployments for a flow (in JSON)**

```shell
kosli list deployments 
	--page-limit 30 
	--output json
```



## kosli_list_environments.md
---
title: "kosli list environments"
beta: false
deprecated: false
summary: "List environments for an org."
---

# kosli list environments

## Synopsis

List environments for an org.

```shell
kosli list environments [flags]
```

## Flags
| Flag | Description |
| :--- | :--- |
|    -h, --help  |  help for environments  |
|    -o, --output string  |  [defaulted] The format of the output. Valid formats are: [table, json]. (default "table")  |


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


## Live Example

{{< raw-html >}}To view a live example of 'kosli list environments' you can run the commands below (for the <a href="https://app.kosli.com/cyber-dojo/environments/aws-prod/snapshots/">cyber-dojo</a> demo organization).<br/><a href="https://app.kosli.com/api/v2/livedocs/cyber-dojo/cli?command=kosli+list+environments+--output=json">Run the commands below and view the output.</a><pre>export KOSLI_ORG=cyber-dojo
export KOSLI_API_TOKEN=Pj_XT2deaVA6V1qrTlthuaWsmjVt4eaHQwqnwqjRO3A  # read-only
kosli list environments --output=json</pre>{{< / raw-html >}}



## kosli_list_flows.md
---
title: "kosli list flows"
beta: false
deprecated: false
summary: "List flows for an org."
---

# kosli list flows

## Synopsis

List flows for an org.

```shell
kosli list flows [flags]
```

## Flags
| Flag | Description |
| :--- | :--- |
|    -h, --help  |  help for flows  |
|    -o, --output string  |  [defaulted] The format of the output. Valid formats are: [table, json]. (default "table")  |


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


## Live Example

{{< raw-html >}}To view a live example of 'kosli list flows' you can run the commands below (for the <a href="https://app.kosli.com/cyber-dojo/environments/aws-prod/snapshots/">cyber-dojo</a> demo organization).<br/><a href="https://app.kosli.com/api/v2/livedocs/cyber-dojo/cli?command=kosli+list+flows+--output=json">Run the commands below and view the output.</a><pre>export KOSLI_ORG=cyber-dojo
export KOSLI_API_TOKEN=Pj_XT2deaVA6V1qrTlthuaWsmjVt4eaHQwqnwqjRO3A  # read-only
kosli list flows --output=json</pre>{{< / raw-html >}}



## kosli_list_policies.md
---
title: "kosli list policies"
beta: false
deprecated: false
summary: "List environment policies for an org."
---

# kosli list policies

## Synopsis

List environment policies for an org.

```shell
kosli list policies [flags]
```

## Flags
| Flag | Description |
| :--- | :--- |
|    -h, --help  |  help for policies  |
|    -o, --output string  |  [defaulted] The format of the output. Valid formats are: [table, json]. (default "table")  |


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




## kosli_list_snapshots.md
---
title: "kosli list snapshots"
beta: false
deprecated: false
summary: "List environment snapshots."
---

# kosli list snapshots

## Synopsis

List environment snapshots.
The results are paginated and ordered from latest to oldest.
By default, the page limit is 15 snapshots per page.

You can optionally specify an INTERVAL between two snapshot expressions with [expression]..[expression]. 

Expressions can be:
* ~N   N'th behind the latest snapshot  
* N    snapshot number N  
* NOW  the latest snapshot  

Either expression can be omitted to default to NOW.


```shell
kosli list snapshots ENV_NAME [flags]
```

## Flags
| Flag | Description |
| :--- | :--- |
|    -h, --help  |  help for snapshots  |
|    -i, --interval string  |  [optional] Expression to define specified snapshots range.  |
|    -o, --output string  |  [defaulted] The format of the output. Valid formats are: [table, json]. (default "table")  |
|        --page int  |  [defaulted] The page number of a response. (default 1)  |
|    -n, --page-limit int  |  [defaulted] The number of elements per page. (default 15)  |
|        --reverse  |  [defaulted] Reverse the order of output list.  |


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


## Live Example

{{< raw-html >}}To view a live example of 'kosli list snapshots' you can run the commands below (for the <a href="https://app.kosli.com/cyber-dojo/environments/aws-prod/snapshots/">cyber-dojo</a> demo organization).<br/><a href="https://app.kosli.com/api/v2/livedocs/cyber-dojo/cli?command=kosli+list+snapshots+aws-prod+--output=json">Run the commands below and view the output.</a><pre>export KOSLI_ORG=cyber-dojo
export KOSLI_API_TOKEN=Pj_XT2deaVA6V1qrTlthuaWsmjVt4eaHQwqnwqjRO3A  # read-only
kosli list snapshots aws-prod --output=json</pre>{{< / raw-html >}}

## Examples Use Cases

These examples all assume that the flags  `--api-token`, `--org`, `--host`, (and `--flow`, `--trail` when required), are set/provided. 

**list the last 15 snapshots for an environment**

```shell
kosli list snapshots yourEnvironmentName 

```

**list the last 30 snapshots for an environment**

```shell
kosli list snapshots yourEnvironmentName 
	--page-limit 30 

```

**list the last 30 snapshots for an environment (in JSON)**

```shell
kosli list snapshots yourEnvironmentName 
	--page-limit 30 
	--output json
```



## kosli_list_trails.md
---
title: "kosli list trails"
beta: false
deprecated: false
summary: "List Trails for a Flow in an org."
---

# kosli list trails

## Synopsis

List Trails for a Flow in an org.The results are ordered from latest to oldest.  
If the `page-limit` flag is provided, the results will be paginated, otherwise all results will be 
returned.  
If `page-limit` is set to 0, all results will be returned.

```shell
kosli list trails [flags]
```

## Flags
| Flag | Description |
| :--- | :--- |
|    -f, --flow string  |  The Kosli flow name.  |
|    -h, --help  |  help for trails  |
|    -o, --output string  |  [defaulted] The format of the output. Valid formats are: [table, json]. (default "table")  |
|        --page int  |  [defaulted] The page number of a response. (default 1)  |
|    -n, --page-limit int  |  [defaulted] The number of elements per page.  |


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

These examples all assume that the flags  `--api-token`, `--org`, `--host`, (and `--flow`, `--trail` when required), are set/provided. 

**list all trails for a flow**

```shell
kosli list trails 

```

**list the most recent 30 trails for a flow**

```shell
kosli list trails 
	--page-limit 30 

```

**show the second page of trails for a flow**

```shell
kosli list trails 
	--page-limit 30 
	--page 2 

```

**list all trails for a flow (in JSON)**

```shell
kosli list trails 
	--output json
```



## kosli_log_environment.md
---
title: "kosli log environment"
beta: false
deprecated: false
summary: "List environment events."
---

# kosli log environment

## Synopsis

List environment events.
The results are paginated and ordered from latest to oldest.
By default, the page limit is 15 events per page.

You can optionally specify an INTERVAL between two snapshot expressions with [expression]..[expression]. 

Expressions can be:
* ~N   N'th behind the latest snapshot  
* N    snapshot number N  
* NOW  the latest snapshot  

Either expression can be omitted to default to NOW.


```shell
kosli log environment ENV_NAME [flags]
```

## Flags
| Flag | Description |
| :--- | :--- |
|    -h, --help  |  help for environment  |
|    -i, --interval string  |  [optional] Expression to define specified snapshots range.  |
|    -o, --output string  |  [defaulted] The format of the output. Valid formats are: [table, json]. (default "table")  |
|        --page int  |  [defaulted] The page number of a response. (default 1)  |
|    -n, --page-limit int  |  [defaulted] The number of elements per page. (default 15)  |
|        --reverse  |  [defaulted] Reverse the order of output list.  |


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


## Live Example

{{< raw-html >}}To view a live example of 'kosli log environment' you can run the commands below (for the <a href="https://app.kosli.com/cyber-dojo/environments/aws-prod/snapshots/">cyber-dojo</a> demo organization).<br/><a href="https://app.kosli.com/api/v2/livedocs/cyber-dojo/cli?command=kosli+log+environment+aws-prod+--output=json">Run the commands below and view the output.</a><pre>export KOSLI_ORG=cyber-dojo
export KOSLI_API_TOKEN=Pj_XT2deaVA6V1qrTlthuaWsmjVt4eaHQwqnwqjRO3A  # read-only
kosli log environment aws-prod --output=json</pre>{{< / raw-html >}}

## Examples Use Cases

These examples all assume that the flags  `--api-token`, `--org`, `--host`, (and `--flow`, `--trail` when required), are set/provided. 

**list the last 15 events for an environment**

```shell
kosli log environment yourEnvironmentName 

```

**list the last 30 events for an environment**

```shell
kosli log environment yourEnvironmentName 
	--page-limit 30 

```

**list the last 30 events for an environment (in JSON)**

```shell
kosli log environment yourEnvironmentName 
	--page-limit 30 
	--output json
```



## kosli_rename_environment.md
---
title: "kosli rename environment"
beta: false
deprecated: false
summary: "Rename a Kosli environment."
---

# kosli rename environment

## Synopsis

Rename a Kosli environment.
The environment will remain accessible under its old name until that name is taken by another environment.


```shell
kosli rename environment OLD_NAME NEW_NAME [flags]
```

## Flags
| Flag | Description |
| :--- | :--- |
|    -D, --dry-run  |  [optional] Run in dry-run mode. When enabled, no data is sent to Kosli and the CLI exits with 0 exit code regardless of any errors.  |
|    -h, --help  |  help for environment  |


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

These examples all assume that the flags  `--api-token`, `--org`, `--host`, (and `--flow`, `--trail` when required), are set/provided. 

**rename a Kosli environment**

```shell
kosli rename environment oldName newName 
```



## kosli_rename_flow.md
---
title: "kosli rename flow"
beta: false
deprecated: false
summary: "Rename a Kosli flow."
---

# kosli rename flow

## Synopsis

Rename a Kosli flow.
The flow will remain accessible under its old name until that name is taken by another flow.


```shell
kosli rename flow OLD_NAME NEW_NAME [flags]
```

## Flags
| Flag | Description |
| :--- | :--- |
|    -D, --dry-run  |  [optional] Run in dry-run mode. When enabled, no data is sent to Kosli and the CLI exits with 0 exit code regardless of any errors.  |
|    -h, --help  |  help for flow  |


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

These examples all assume that the flags  `--api-token`, `--org`, `--host`, (and `--flow`, `--trail` when required), are set/provided. 

**rename a Kosli flow**

```shell
kosli rename flow oldName newName 
```



## kosli_report_approval.md
---
title: "kosli report approval"
beta: false
deprecated: false
summary: "Report an approval of deploying an artifact to an environment to Kosli.  "
---

# kosli report approval

## Synopsis

Report an approval of deploying an artifact to an environment to Kosli.  

The artifact fingerprint can be provided directly with the `--fingerprint` flag, or 
calculated based on `--artifact-type` flag.

Artifact type can be one of: "file" for files, "dir" for directories, "oci" for container
images in registries or "docker" for local docker images.



```shell
kosli report approval [IMAGE-NAME | FILE-PATH | DIR-PATH] [flags]
```

## Flags
| Flag | Description |
| :--- | :--- |
|        --approver string  |  [optional] The user approving an approval.  |
|    -t, --artifact-type string  |  The type of the artifact to calculate its SHA256 fingerprint. One of: [oci, docker, file, dir]. Only required if you want Kosli to calculate the fingerprint for you (i.e. when you don't specify '--fingerprint' on commands that allow it).  |
|    -d, --description string  |  [optional] The approval description.  |
|    -D, --dry-run  |  [optional] Run in dry-run mode. When enabled, no data is sent to Kosli and the CLI exits with 0 exit code regardless of any errors.  |
|    -e, --environment string  |  [defaulted] The environment the artifact is approved for. (defaults to all environments)  |
|    -x, --exclude strings  |  [optional] The comma separated list of directories and files to exclude from fingerprinting. Can take glob patterns. Only applicable for --artifact-type dir.  |
|    -F, --fingerprint string  |  [conditional] The SHA256 fingerprint of the artifact. Only required if you don't specify '--artifact-type'.  |
|    -f, --flow string  |  The Kosli flow name.  |
|    -h, --help  |  help for approval  |
|        --newest-commit string  |  [defaulted] The source commit sha for the newest change in the deployment. Can be any commit-ish. (default "HEAD")  |
|        --oldest-commit string  |  [conditional] The source commit sha for the oldest change in the deployment. Can be any commit-ish. Only required if you don't specify '--environment'.  |
|        --registry-password string  |  [conditional] The container registry password or access token. Only required if you want to read container image SHA256 digest from a remote container registry.  |
|        --registry-username string  |  [conditional] The container registry username. Only required if you want to read container image SHA256 digest from a remote container registry.  |
|        --repo-root string  |  [defaulted] The directory where the source git repository is available. (default ".")  |
|    -u, --user-data string  |  [optional] The path to a JSON file containing additional data you would like to attach to the approval.  |


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

{{< tabs "live-examples" "col-no-wrap" >}}{{< tab "GitHub" >}}View an example of the `kosli report approval` command in GitHub.

In [this YAML file](https://app.kosli.com/api/v2/livedocs/cyber-dojo/yaml?ci=github&command=kosli+report+approval), which created [this Kosli Event](https://app.kosli.com/api/v2/livedocs/cyber-dojo/event?ci=github&command=kosli+report+approval).{{< /tab >}}{{< tab "GitLab" >}}View an example of the `kosli report approval` command in GitLab.

In [this YAML file](https://app.kosli.com/api/v2/livedocs/cyber-dojo/yaml?ci=gitlab&command=kosli+report+approval), which created [this Kosli Event](https://app.kosli.com/api/v2/livedocs/cyber-dojo/event?ci=gitlab&command=kosli+report+approval).{{< /tab >}}{{< /tabs >}}

## Examples Use Cases

These examples all assume that the flags  `--api-token`, `--org`, `--host`, (and `--flow`, `--trail` when required), are set/provided. 

```shell
# Report that an artifact with a provided fingerprint (sha256) has been approved for 
# deployment to environment <yourEnvironmentName>.
# The approval is for all git commits since the last approval to this environment.
kosli report approval \
	--api-token yourAPIToken \
	--description "An optional description for the approval" \
	--environment yourEnvironmentName \
	--approver username \
	--org yourOrgName \
	--flow yourFlowName \
	--fingerprint yourArtifactFingerprint

# Report that a file type artifact has been approved for deployment to environment <yourEnvironmentName>.
# The approval is for all git commits since the last approval to this environment.
kosli report approval FILE.tgz \
	--api-token yourAPIToken \
	--artifact-type file \
	--description "An optional description for the approval" \
	--environment yourEnvironmentName \
	--newest-commit HEAD \
	--approver username \
	--org yourOrgName \
	--flow yourFlowName 

# Report that an artifact with a provided fingerprint (sha256) has been approved for deployment.
# The approval is for all environments.
# The approval is for all commits since the git commit of origin/production branch.
kosli report approval \
	--api-token yourAPIToken \
	--description "An optional description for the approval" \
	--newest-commit HEAD \
	--oldest-commit origin/production \
	--approver username \
	--org yourOrgName \
	--flow yourFlowName \
	--fingerprint yourArtifactFingerprint
```



## kosli_report_artifact.md
---
title: "kosli report artifact"
beta: false
deprecated: true
summary: "Report an artifact creation to a Kosli flow.  "
---

# kosli report artifact

{{% hint danger %}}
**kosli report artifact** is deprecated. see kosli attest commands  Deprecated commands will be removed in a future release.
{{% /hint %}}
## Synopsis

Report an artifact creation to a Kosli flow.  

The artifact fingerprint can be provided directly with the `--fingerprint` flag, or 
calculated based on `--artifact-type` flag.

Artifact type can be one of: "file" for files, "dir" for directories, "oci" for container
images in registries or "docker" for local docker images.



```shell
kosli report artifact {IMAGE-NAME | FILE-PATH | DIR-PATH} [flags]
```

## Flags
| Flag | Description |
| :--- | :--- |
|    -t, --artifact-type string  |  The type of the artifact to calculate its SHA256 fingerprint. One of: [oci, docker, file, dir]. Only required if you want Kosli to calculate the fingerprint for you (i.e. when you don't specify '--fingerprint' on commands that allow it).  |
|    -b, --build-url string  |  The url of CI pipeline that built the artifact. (defaulted in some CIs: https://docs.kosli.com/ci-defaults ).  |
|    -u, --commit-url string  |  The url for the git commit that created the artifact. (defaulted in some CIs: https://docs.kosli.com/ci-defaults ).  |
|    -D, --dry-run  |  [optional] Run in dry-run mode. When enabled, no data is sent to Kosli and the CLI exits with 0 exit code regardless of any errors.  |
|    -x, --exclude strings  |  [optional] The comma separated list of directories and files to exclude from fingerprinting. Can take glob patterns. Only applicable for --artifact-type dir.  |
|    -F, --fingerprint string  |  [conditional] The SHA256 fingerprint of the artifact. Only required if you don't specify '--artifact-type'.  |
|    -f, --flow string  |  The Kosli flow name.  |
|    -g, --git-commit string  |  [defaulted] The git commit from which the artifact was created. (defaulted in some CIs: https://docs.kosli.com/ci-defaults, otherwise defaults to HEAD ).  |
|    -h, --help  |  help for artifact  |
|    -n, --name string  |  [optional] Artifact display name, if different from file, image or directory name.  |
|        --registry-password string  |  [conditional] The container registry password or access token. Only required if you want to read container image SHA256 digest from a remote container registry.  |
|        --registry-username string  |  [conditional] The container registry username. Only required if you want to read container image SHA256 digest from a remote container registry.  |
|        --repo-root string  |  [defaulted] The directory where the source git repository is available. (default ".")  |


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

These examples all assume that the flags  `--api-token`, `--org`, `--host`, (and `--flow`, `--trail` when required), are set/provided. 

**Report to a Kosli flow that a file type artifact has been created**

```shell
kosli report artifact FILE.tgz 
	--artifact-type file 
	--build-url https://exampleci.com 
	--commit-url https://github.com/YourOrg/YourProject/commit/yourCommitShaThatThisArtifactWasBuiltFrom 
	--git-commit yourCommitShaThatThisArtifactWasBuiltFrom 

```

**Report to a Kosli flow that an artifact with a provided fingerprint (sha256) has been created**

```shell
kosli report artifact ANOTHER_FILE.txt 
	--build-url https://exampleci.com 
	--commit-url https://github.com/YourOrg/YourProject/commit/yourCommitShaThatThisArtifactWasBuiltFrom 
	--git-commit yourCommitShaThatThisArtifactWasBuiltFrom 
	--fingerprint yourArtifactFingerprint
```



## kosli_report_evidence_artifact_generic.md
---
title: "kosli report evidence artifact generic"
beta: false
deprecated: true
summary: "Report generic evidence to an artifact in a Kosli flow.  "
---

# kosli report evidence artifact generic

{{% hint danger %}}
**kosli report evidence artifact generic** is deprecated. See **kosli attest** commands.  Deprecated commands will be removed in a future release.
{{% /hint %}}
## Synopsis

Report generic evidence to an artifact in a Kosli flow.  

The artifact fingerprint can be provided directly with the `--fingerprint` flag, or 
calculated based on `--artifact-type` flag.

Artifact type can be one of: "file" for files, "dir" for directories, "oci" for container
images in registries or "docker" for local docker images.



```shell
kosli report evidence artifact generic [IMAGE-NAME | FILE-PATH | DIR-PATH] [flags]
```

## Flags
| Flag | Description |
| :--- | :--- |
|    -t, --artifact-type string  |  The type of the artifact to calculate its SHA256 fingerprint. One of: [oci, docker, file, dir]. Only required if you want Kosli to calculate the fingerprint for you (i.e. when you don't specify '--fingerprint' on commands that allow it).  |
|    -b, --build-url string  |  The url of CI pipeline that generated the evidence. (defaulted in some CIs: https://docs.kosli.com/ci-defaults ).  |
|    -C, --compliant  |  [defaulted] Whether the evidence is compliant or not. A boolean flag https://docs.kosli.com/faq/#boolean-flags (default true)  |
|    -d, --description string  |  [optional] The evidence description.  |
|    -D, --dry-run  |  [optional] Run in dry-run mode. When enabled, no data is sent to Kosli and the CLI exits with 0 exit code regardless of any errors.  |
|        --evidence-fingerprint string  |  [optional] The SHA256 fingerprint of the evidence file or dir.  |
|    -e, --evidence-paths strings  |  [optional] The comma-separated list of paths containing supporting proof for the reported evidence. Paths can be for files or directories. All provided proofs will be uploaded to Kosli's evidence vault.  |
|        --evidence-url string  |  [optional] The external URL where the evidence file or dir is stored.  |
|    -x, --exclude strings  |  [optional] The comma separated list of directories and files to exclude from fingerprinting. Can take glob patterns. Only applicable for --artifact-type dir.  |
|    -F, --fingerprint string  |  [conditional] The SHA256 fingerprint of the artifact. Only required if you don't specify '--artifact-type'.  |
|    -f, --flow string  |  The Kosli flow name.  |
|    -h, --help  |  help for generic  |
|    -n, --name string  |  The name of the evidence.  |
|        --registry-password string  |  [conditional] The container registry password or access token. Only required if you want to read container image SHA256 digest from a remote container registry.  |
|        --registry-username string  |  [conditional] The container registry username. Only required if you want to read container image SHA256 digest from a remote container registry.  |
|    -u, --user-data string  |  [optional] The path to a JSON file containing additional data you would like to attach to the evidence.  |


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

These examples all assume that the flags  `--api-token`, `--org`, `--host`, (and `--flow`, `--trail` when required), are set/provided. 

**report a generic evidence about a pre-built docker image**

```shell
kosli report evidence artifact generic yourDockerImageName 
	--artifact-type docker 
	--build-url https://exampleci.com 
	--name yourEvidenceName 

```

**report a generic evidence about a directory type artifact**

```shell
kosli report evidence artifact generic /path/to/your/dir 
	--artifact-type dir 
	--build-url https://exampleci.com 
	--name yourEvidenceName 

```

**report a generic evidence about an artifact with a provided fingerprint (sha256)**

```shell
kosli report evidence artifact generic 
	--build-url https://exampleci.com 
	--name yourEvidenceName 
	--fingerprint yourArtifactFingerprint

```

**report a generic evidence about an artifact with evidence file upload**

```shell
kosli report evidence artifact generic 
	--build-url https://exampleci.com 
	--name yourEvidenceName 
	--fingerprint yourArtifactFingerprint 
	--evidence-paths=yourEvidencePathName

```

**report a generic evidence about an artifact with evidence file upload via API**

```shell
curl -X 'POST' 
	'https://app.kosli.com/api/v2/evidence/yourOrgName/artifact/yourFlowName/generic' 
	-H 'accept: application/json' 
	-H 'Content-Type: multipart/form-data' 
	-F 'evidence_json={
  	  "artifact_fingerprint": "yourArtifactFingerprint",
	  "name": "yourEvidenceName",
      "build_url": "https://exampleci.com",
      "is_compliant": true
    }' 
	-F 'evidence_file=@yourEvidencePathName'
```



## kosli_report_evidence_artifact_junit.md
---
title: "kosli report evidence artifact junit"
beta: false
deprecated: true
summary: "Report JUnit test evidence for an artifact in a Kosli flow.  "
---

# kosli report evidence artifact junit

{{% hint danger %}}
**kosli report evidence artifact junit** is deprecated. See **kosli attest** commands.  Deprecated commands will be removed in a future release.
{{% /hint %}}
## Synopsis

Report JUnit test evidence for an artifact in a Kosli flow.    
All .xml files from --results-dir are parsed and uploaded to Kosli's evidence vault.  
If there are no failing tests and no errors the evidence is reported as compliant. Otherwise the evidence is reported as non-compliant.  

The artifact fingerprint can be provided directly with the `--fingerprint` flag, or 
calculated based on `--artifact-type` flag.

Artifact type can be one of: "file" for files, "dir" for directories, "oci" for container
images in registries or "docker" for local docker images.



```shell
kosli report evidence artifact junit [IMAGE-NAME | FILE-PATH | DIR-PATH] [flags]
```

## Flags
| Flag | Description |
| :--- | :--- |
|    -t, --artifact-type string  |  The type of the artifact to calculate its SHA256 fingerprint. One of: [oci, docker, file, dir]. Only required if you want Kosli to calculate the fingerprint for you (i.e. when you don't specify '--fingerprint' on commands that allow it).  |
|    -b, --build-url string  |  The url of CI pipeline that generated the evidence. (defaulted in some CIs: https://docs.kosli.com/ci-defaults ).  |
|    -D, --dry-run  |  [optional] Run in dry-run mode. When enabled, no data is sent to Kosli and the CLI exits with 0 exit code regardless of any errors.  |
|        --evidence-fingerprint string  |  [optional] The SHA256 fingerprint of the evidence file or dir.  |
|        --evidence-url string  |  [optional] The external URL where the evidence file or dir is stored.  |
|    -x, --exclude strings  |  [optional] The comma separated list of directories and files to exclude from fingerprinting. Can take glob patterns. Only applicable for --artifact-type dir.  |
|    -F, --fingerprint string  |  [conditional] The SHA256 fingerprint of the artifact. Only required if you don't specify '--artifact-type'.  |
|    -f, --flow string  |  The Kosli flow name.  |
|    -h, --help  |  help for junit  |
|    -n, --name string  |  The name of the evidence.  |
|        --registry-password string  |  [conditional] The container registry password or access token. Only required if you want to read container image SHA256 digest from a remote container registry.  |
|        --registry-username string  |  [conditional] The container registry username. Only required if you want to read container image SHA256 digest from a remote container registry.  |
|    -R, --results-dir string  |  [defaulted] The path to a directory with JUnit test results. By default, the directory will be uploaded to Kosli's evidence vault. (default ".")  |
|    -u, --user-data string  |  [optional] The path to a JSON file containing additional data you would like to attach to the evidence.  |


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

These examples all assume that the flags  `--api-token`, `--org`, `--host`, (and `--flow`, `--trail` when required), are set/provided. 

**report JUnit test evidence about a file artifact**

```shell
kosli report evidence artifact junit FILE.tgz 
	--artifact-type file 
	--name yourEvidenceName 
	--build-url https://exampleci.com 
	--results-dir yourFolderWithJUnitResults

```

**report JUnit test evidence about an artifact using an available Sha256 digest**

```shell
kosli report evidence artifact junit 
	--fingerprint yourSha256 
	--name yourEvidenceName 
	--build-url https://exampleci.com 
	--results-dir yourFolderWithJUnitResults
```



## kosli_report_evidence_artifact_pullrequest_azure.md
---
title: "kosli report evidence artifact pullrequest azure"
beta: false
deprecated: true
summary: "Report an Azure Devops pull request evidence for an artifact in a Kosli flow.  "
---

# kosli report evidence artifact pullrequest azure

{{% hint danger %}}
**kosli report evidence artifact pullrequest azure** is deprecated. See **kosli attest** commands.  Deprecated commands will be removed in a future release.
{{% /hint %}}
## Synopsis

Report an Azure Devops pull request evidence for an artifact in a Kosli flow.  
It checks if a pull request exists for the artifact (based on its git commit) and reports the pull-request evidence to the artifact in Kosli.  

The artifact fingerprint can be provided directly with the `--fingerprint` flag, or 
calculated based on `--artifact-type` flag.

Artifact type can be one of: "file" for files, "dir" for directories, "oci" for container
images in registries or "docker" for local docker images.



```shell
kosli report evidence artifact pullrequest azure [IMAGE-NAME | FILE-PATH | DIR-PATH] [flags]
```

## Flags
| Flag | Description |
| :--- | :--- |
|    -t, --artifact-type string  |  The type of the artifact to calculate its SHA256 fingerprint. One of: [oci, docker, file, dir]. Only required if you want Kosli to calculate the fingerprint for you (i.e. when you don't specify '--fingerprint' on commands that allow it).  |
|        --assert  |  [optional] Exit with non-zero code if no pull requests found for the given commit.  |
|        --azure-org-url string  |  Azure organization url. E.g. "https://dev.azure.com/myOrg" (defaulted if you are running in Azure Devops pipelines: https://docs.kosli.com/ci-defaults ).  |
|        --azure-token string  |  Azure Personal Access token.  |
|    -b, --build-url string  |  The url of CI pipeline that generated the evidence. (defaulted in some CIs: https://docs.kosli.com/ci-defaults ).  |
|        --commit string  |  Git commit for which to find pull request evidence. (defaulted in some CIs: https://docs.kosli.com/ci-defaults ).  |
|    -D, --dry-run  |  [optional] Run in dry-run mode. When enabled, no data is sent to Kosli and the CLI exits with 0 exit code regardless of any errors.  |
|        --evidence-fingerprint string  |  [optional] The SHA256 fingerprint of the evidence file or dir.  |
|        --evidence-url string  |  [optional] The external URL where the evidence file or dir is stored.  |
|    -x, --exclude strings  |  [optional] The comma separated list of directories and files to exclude from fingerprinting. Can take glob patterns. Only applicable for --artifact-type dir.  |
|    -F, --fingerprint string  |  [conditional] The SHA256 fingerprint of the artifact. Only required if you don't specify '--artifact-type'.  |
|    -f, --flow string  |  The Kosli flow name.  |
|    -h, --help  |  help for azure  |
|    -n, --name string  |  The name of the evidence.  |
|        --project string  |  Azure project.(defaulted if you are running in Azure Devops pipelines: https://docs.kosli.com/ci-defaults ).  |
|        --registry-password string  |  [conditional] The container registry password or access token. Only required if you want to read container image SHA256 digest from a remote container registry.  |
|        --registry-username string  |  [conditional] The container registry username. Only required if you want to read container image SHA256 digest from a remote container registry.  |
|        --repository string  |  Git repository. (defaulted in some CIs: https://docs.kosli.com/ci-defaults ).  |
|    -u, --user-data string  |  [optional] The path to a JSON file containing additional data you would like to attach to the evidence.  |


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

These examples all assume that the flags  `--api-token`, `--org`, `--host`, (and `--flow`, `--trail` when required), are set/provided. 

**report a pull request evidence to kosli for a docker image**

```shell
kosli report evidence artifact pullrequest azure yourDockerImageName 
	--artifact-type docker 
	--azure-org-url https://dev.azure.com/myOrg 
	--project yourAzureDevOpsProject 
	--commit yourGitCommitSha1 
	--repository yourAzureGitRepository 
	--azure-token yourAzureToken 
	--name yourEvidenceName 
	--flows yourFlowName1,yourFlowName2 
	--build-url https://exampleci.com 

```

**fail if a pull request does not exist for your artifact**

```shell
kosli report evidence artifact pullrequest azure yourDockerImageName 
	--artifact-type docker 
	--azure-org-url https://dev.azure.com/myOrg 
	--project yourAzureDevOpsProject 
	--commit yourGitCommitSha1 
	--repository yourAzureGitRepository 
	--azure-token yourAzureToken 
	--name yourEvidenceName 
	--flows yourFlowName1,yourFlowName2 
	--build-url https://exampleci.com 
	--assert
```



## kosli_report_evidence_artifact_pullrequest_bitbucket.md
---
title: "kosli report evidence artifact pullrequest bitbucket"
beta: false
deprecated: true
summary: "Report a Bitbucket pull request evidence for an artifact in a Kosli flow.  "
---

# kosli report evidence artifact pullrequest bitbucket

{{% hint danger %}}
**kosli report evidence artifact pullrequest bitbucket** is deprecated. See **kosli attest** commands.  Deprecated commands will be removed in a future release.
{{% /hint %}}
## Synopsis

Report a Bitbucket pull request evidence for an artifact in a Kosli flow.  
It checks if a pull request exists for the artifact (based on its git commit) and reports the pull-request evidence to the artifact in Kosli.  

The artifact fingerprint can be provided directly with the `--fingerprint` flag, or 
calculated based on `--artifact-type` flag.

Artifact type can be one of: "file" for files, "dir" for directories, "oci" for container
images in registries or "docker" for local docker images.



```shell
kosli report evidence artifact pullrequest bitbucket [IMAGE-NAME | FILE-PATH | DIR-PATH] [flags]
```

## Flags
| Flag | Description |
| :--- | :--- |
|    -t, --artifact-type string  |  The type of the artifact to calculate its SHA256 fingerprint. One of: [oci, docker, file, dir]. Only required if you want Kosli to calculate the fingerprint for you (i.e. when you don't specify '--fingerprint' on commands that allow it).  |
|        --assert  |  [optional] Exit with non-zero code if no pull requests found for the given commit.  |
|        --bitbucket-access-token string  |  Bitbucket repo/project/workspace access token. See https://developer.atlassian.com/cloud/bitbucket/rest/intro/#access-tokens for more details.  |
|        --bitbucket-password string  |  Bitbucket App password. See https://developer.atlassian.com/cloud/bitbucket/rest/intro/#authentication for more details.  |
|        --bitbucket-username string  |  Bitbucket username. Only needed if you use --bitbucket-password  |
|        --bitbucket-workspace string  |  Bitbucket workspace ID.  |
|    -b, --build-url string  |  The url of CI pipeline that generated the evidence. (defaulted in some CIs: https://docs.kosli.com/ci-defaults ).  |
|        --commit string  |  Git commit for which to find pull request evidence. (defaulted in some CIs: https://docs.kosli.com/ci-defaults ).  |
|    -D, --dry-run  |  [optional] Run in dry-run mode. When enabled, no data is sent to Kosli and the CLI exits with 0 exit code regardless of any errors.  |
|        --evidence-fingerprint string  |  [optional] The SHA256 fingerprint of the evidence file or dir.  |
|        --evidence-url string  |  [optional] The external URL where the evidence file or dir is stored.  |
|    -x, --exclude strings  |  [optional] The comma separated list of directories and files to exclude from fingerprinting. Can take glob patterns. Only applicable for --artifact-type dir.  |
|    -F, --fingerprint string  |  [conditional] The SHA256 fingerprint of the artifact. Only required if you don't specify '--artifact-type'.  |
|    -f, --flow string  |  The Kosli flow name.  |
|    -h, --help  |  help for bitbucket  |
|    -n, --name string  |  The name of the evidence.  |
|        --registry-password string  |  [conditional] The container registry password or access token. Only required if you want to read container image SHA256 digest from a remote container registry.  |
|        --registry-username string  |  [conditional] The container registry username. Only required if you want to read container image SHA256 digest from a remote container registry.  |
|        --repository string  |  Git repository. (defaulted in some CIs: https://docs.kosli.com/ci-defaults ).  |
|    -u, --user-data string  |  [optional] The path to a JSON file containing additional data you would like to attach to the evidence.  |


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

These examples all assume that the flags  `--api-token`, `--org`, `--host`, (and `--flow`, `--trail` when required), are set/provided. 

**report a pull request evidence to kosli for a docker image**

```shell
kosli report evidence artifact pullrequest bitbucket yourDockerImageName 
	--artifact-type docker 
	--build-url https://exampleci.com 
	--name yourEvidenceName 
	--bitbucket-username yourBitbucketUsername 
	--bitbucket-password yourBitbucketPassword 
	--bitbucket-workspace yourBitbucketWorkspace 
	--commit yourArtifactGitCommit 
	--repository yourBitbucketGitRepository 

```

**fail if a pull request does not exist for your artifact**

```shell
kosli report evidence artifact pullrequest bitbucket yourDockerImageName 
	--artifact-type docker 
	--build-url https://exampleci.com 
	--name yourEvidenceName 
	--bitbucket-username yourBitbucketUsername 
	--bitbucket-password yourBitbucketPassword 
	--bitbucket-workspace yourBitbucketWorkspace 
	--commit yourArtifactGitCommit 
	--repository yourBitbucketGitRepository 
	--assert
```



## kosli_report_evidence_artifact_pullrequest_github.md
---
title: "kosli report evidence artifact pullrequest github"
beta: false
deprecated: true
summary: "Report a Github pull request evidence for an artifact in a Kosli flow.  "
---

# kosli report evidence artifact pullrequest github

{{% hint danger %}}
**kosli report evidence artifact pullrequest github** is deprecated. See **kosli attest** commands.  Deprecated commands will be removed in a future release.
{{% /hint %}}
## Synopsis

Report a Github pull request evidence for an artifact in a Kosli flow.  
It checks if a pull request exists for the artifact (based on its git commit) and reports the pull-request evidence to the artifact in Kosli.  

The artifact fingerprint can be provided directly with the `--fingerprint` flag, or 
calculated based on `--artifact-type` flag.

Artifact type can be one of: "file" for files, "dir" for directories, "oci" for container
images in registries or "docker" for local docker images.



```shell
kosli report evidence artifact pullrequest github [IMAGE-NAME | FILE-PATH | DIR-PATH] [flags]
```

## Flags
| Flag | Description |
| :--- | :--- |
|    -t, --artifact-type string  |  The type of the artifact to calculate its SHA256 fingerprint. One of: [oci, docker, file, dir]. Only required if you want Kosli to calculate the fingerprint for you (i.e. when you don't specify '--fingerprint' on commands that allow it).  |
|        --assert  |  [optional] Exit with non-zero code if no pull requests found for the given commit.  |
|    -b, --build-url string  |  The url of CI pipeline that generated the evidence. (defaulted in some CIs: https://docs.kosli.com/ci-defaults ).  |
|        --commit string  |  Git commit for which to find pull request evidence. (defaulted in some CIs: https://docs.kosli.com/ci-defaults ).  |
|    -D, --dry-run  |  [optional] Run in dry-run mode. When enabled, no data is sent to Kosli and the CLI exits with 0 exit code regardless of any errors.  |
|        --evidence-fingerprint string  |  [optional] The SHA256 fingerprint of the evidence file or dir.  |
|        --evidence-url string  |  [optional] The external URL where the evidence file or dir is stored.  |
|    -x, --exclude strings  |  [optional] The comma separated list of directories and files to exclude from fingerprinting. Can take glob patterns. Only applicable for --artifact-type dir.  |
|    -F, --fingerprint string  |  [conditional] The SHA256 fingerprint of the artifact. Only required if you don't specify '--artifact-type'.  |
|    -f, --flow string  |  The Kosli flow name.  |
|        --github-base-url string  |  [optional] GitHub base URL (only needed for GitHub Enterprise installations).  |
|        --github-org string  |  Github organization. (defaulted if you are running in GitHub Actions: https://docs.kosli.com/ci-defaults ).  |
|        --github-token string  |  Github token.  |
|    -h, --help  |  help for github  |
|    -n, --name string  |  The name of the evidence.  |
|        --registry-password string  |  [conditional] The container registry password or access token. Only required if you want to read container image SHA256 digest from a remote container registry.  |
|        --registry-username string  |  [conditional] The container registry username. Only required if you want to read container image SHA256 digest from a remote container registry.  |
|        --repository string  |  Git repository. (defaulted in some CIs: https://docs.kosli.com/ci-defaults ).  |
|    -u, --user-data string  |  [optional] The path to a JSON file containing additional data you would like to attach to the evidence.  |


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

These examples all assume that the flags  `--api-token`, `--org`, `--host`, (and `--flow`, `--trail` when required), are set/provided. 

**report a pull request evidence to kosli for a docker image**

```shell
kosli report evidence artifact pullrequest github yourDockerImageName 
	--artifact-type docker 
	--build-url https://exampleci.com 
	--name yourEvidenceName 
	--github-token yourGithubToken 
	--github-org yourGithubOrg 
	--commit yourArtifactGitCommit 
	--repository yourGithubGitRepository 

```

**fail if a pull request does not exist for your artifact**

```shell
kosli report evidence artifact pullrequest github yourDockerImageName 
	--artifact-type docker 
	--build-url https://exampleci.com 
	--name yourEvidenceName 
	--github-token yourGithubToken 
	--github-org yourGithubOrg 
	--commit yourArtifactGitCommit 
	--repository yourGithubGitRepository 
	--assert
```



## kosli_report_evidence_artifact_pullrequest_gitlab.md
---
title: "kosli report evidence artifact pullrequest gitlab"
beta: false
deprecated: true
summary: "Report a Gitlab merge request evidence for an artifact in a Kosli flow.  "
---

# kosli report evidence artifact pullrequest gitlab

{{% hint danger %}}
**kosli report evidence artifact pullrequest gitlab** is deprecated. See **kosli attest** commands.  Deprecated commands will be removed in a future release.
{{% /hint %}}
## Synopsis

Report a Gitlab merge request evidence for an artifact in a Kosli flow.  
It checks if a merge request exists for the artifact (based on its git commit) and reports the merge request evidence to the artifact in Kosli.  

The artifact fingerprint can be provided directly with the `--fingerprint` flag, or 
calculated based on `--artifact-type` flag.

Artifact type can be one of: "file" for files, "dir" for directories, "oci" for container
images in registries or "docker" for local docker images.



```shell
kosli report evidence artifact pullrequest gitlab [IMAGE-NAME | FILE-PATH | DIR-PATH] [flags]
```

## Flags
| Flag | Description |
| :--- | :--- |
|    -t, --artifact-type string  |  The type of the artifact to calculate its SHA256 fingerprint. One of: [oci, docker, file, dir]. Only required if you want Kosli to calculate the fingerprint for you (i.e. when you don't specify '--fingerprint' on commands that allow it).  |
|        --assert  |  [optional] Exit with non-zero code if no pull requests found for the given commit.  |
|    -b, --build-url string  |  The url of CI pipeline that generated the evidence. (defaulted in some CIs: https://docs.kosli.com/ci-defaults ).  |
|        --commit string  |  Git commit for which to find pull request evidence. (defaulted in some CIs: https://docs.kosli.com/ci-defaults ).  |
|    -D, --dry-run  |  [optional] Run in dry-run mode. When enabled, no data is sent to Kosli and the CLI exits with 0 exit code regardless of any errors.  |
|        --evidence-fingerprint string  |  [optional] The SHA256 fingerprint of the evidence file or dir.  |
|        --evidence-url string  |  [optional] The external URL where the evidence file or dir is stored.  |
|    -x, --exclude strings  |  [optional] The comma separated list of directories and files to exclude from fingerprinting. Can take glob patterns. Only applicable for --artifact-type dir.  |
|    -F, --fingerprint string  |  [conditional] The SHA256 fingerprint of the artifact. Only required if you don't specify '--artifact-type'.  |
|    -f, --flow string  |  The Kosli flow name.  |
|        --gitlab-base-url string  |  [optional] Gitlab base URL (only needed for on-prem Gitlab installations).  |
|        --gitlab-org string  |  Gitlab organization. (defaulted if you are running in Gitlab Pipelines: https://docs.kosli.com/ci-defaults ).  |
|        --gitlab-token string  |  Gitlab token.  |
|    -h, --help  |  help for gitlab  |
|    -n, --name string  |  The name of the evidence.  |
|        --registry-password string  |  [conditional] The container registry password or access token. Only required if you want to read container image SHA256 digest from a remote container registry.  |
|        --registry-username string  |  [conditional] The container registry username. Only required if you want to read container image SHA256 digest from a remote container registry.  |
|        --repository string  |  Git repository. (defaulted in some CIs: https://docs.kosli.com/ci-defaults ).  |
|    -u, --user-data string  |  [optional] The path to a JSON file containing additional data you would like to attach to the evidence.  |


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

These examples all assume that the flags  `--api-token`, `--org`, `--host`, (and `--flow`, `--trail` when required), are set/provided. 

**report a merge request evidence to kosli for a docker image**

```shell
kosli report evidence artifact mergerequest gitlab yourDockerImageName 
	--artifact-type docker 
	--build-url https://exampleci.com 
	--name yourEvidenceName 
	--gitlab-token yourGitlabToken 
	--gitlab-org yourGitlabOrg 
	--commit yourArtifactGitCommit 
	--repository yourGithubGitRepository 

```

**report a merge request evidence (from an on-prem Gitlab) to kosli for a docker image**

```shell
kosli report evidence artifact mergerequest gitlab yourDockerImageName 
	--artifact-type docker 
	--build-url https://exampleci.com 
	--name yourEvidenceName 
	--gitlab-base-url https://gitlab.example.org 
	--gitlab-token yourGitlabToken 
	--gitlab-org yourGitlabOrg 
	--commit yourArtifactGitCommit 
	--repository yourGithubGitRepository 

```

**fail if a merge request does not exist for your artifact**

```shell
kosli report evidence artifact mergerequest gitlab yourDockerImageName 
	--artifact-type docker 
	--build-url https://exampleci.com 
	--gitlab-token yourGitlabToken 
	--gitlab-org yourGitlabOrg 
	--commit yourArtifactGitCommit 
	--repository yourGithubGitRepository 
	--assert
```



## kosli_report_evidence_artifact_snyk.md
---
title: "kosli report evidence artifact snyk"
beta: false
deprecated: true
summary: "Report Snyk vulnerability scan evidence for an artifact in a Kosli flow.  "
---

# kosli report evidence artifact snyk

{{% hint danger %}}
**kosli report evidence artifact snyk** is deprecated. See **kosli attest** commands.  Deprecated commands will be removed in a future release.
{{% /hint %}}
## Synopsis

Report Snyk vulnerability scan evidence for an artifact in a Kosli flow.    
The --scan-results .json file is parsed and uploaded to Kosli's evidence vault.

In CLI <v2.8.2, Snyk results could only be in the Snyk JSON output format. "snyk code test" results were not supported by 
this command and could be reported as generic evidence.

Starting from v2.8.2, the Snyk results can be in Snyk JSON or SARIF output format for "snyk container test". 
"snyk code test" is now supported but only in the SARIF format.

If no vulnerabilities are detected, the evidence is reported as compliant. Otherwise the evidence is reported as non-compliant.


The artifact fingerprint can be provided directly with the `--fingerprint` flag, or 
calculated based on `--artifact-type` flag.

Artifact type can be one of: "file" for files, "dir" for directories, "oci" for container
images in registries or "docker" for local docker images.



```shell
kosli report evidence artifact snyk [IMAGE-NAME | FILE-PATH | DIR-PATH] [flags]
```

## Flags
| Flag | Description |
| :--- | :--- |
|    -t, --artifact-type string  |  The type of the artifact to calculate its SHA256 fingerprint. One of: [oci, docker, file, dir]. Only required if you want Kosli to calculate the fingerprint for you (i.e. when you don't specify '--fingerprint' on commands that allow it).  |
|    -b, --build-url string  |  The url of CI pipeline that generated the evidence. (defaulted in some CIs: https://docs.kosli.com/ci-defaults ).  |
|    -D, --dry-run  |  [optional] Run in dry-run mode. When enabled, no data is sent to Kosli and the CLI exits with 0 exit code regardless of any errors.  |
|        --evidence-fingerprint string  |  [optional] The SHA256 fingerprint of the evidence file or dir.  |
|        --evidence-url string  |  [optional] The external URL where the evidence file or dir is stored.  |
|    -x, --exclude strings  |  [optional] The comma separated list of directories and files to exclude from fingerprinting. Can take glob patterns. Only applicable for --artifact-type dir.  |
|    -F, --fingerprint string  |  [conditional] The SHA256 fingerprint of the artifact. Only required if you don't specify '--artifact-type'.  |
|    -f, --flow string  |  The Kosli flow name.  |
|    -h, --help  |  help for snyk  |
|    -n, --name string  |  The name of the evidence.  |
|        --registry-password string  |  [conditional] The container registry password or access token. Only required if you want to read container image SHA256 digest from a remote container registry.  |
|        --registry-username string  |  [conditional] The container registry username. Only required if you want to read container image SHA256 digest from a remote container registry.  |
|    -R, --scan-results string  |  The path to Snyk SARIF or JSON scan results file from 'snyk test' and 'snyk container test'. By default, the Snyk results will be uploaded to Kosli's evidence vault.  |
|        --upload-results  |  [defaulted] Whether to upload the provided Snyk results file as an attachment to Kosli or not. (default true)  |
|    -u, --user-data string  |  [optional] The path to a JSON file containing additional data you would like to attach to the evidence.  |


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

These examples all assume that the flags  `--api-token`, `--org`, `--host`, (and `--flow`, `--trail` when required), are set/provided. 

**report Snyk vulnerability scan evidence about a file artifact**

```shell
kosli report evidence artifact snyk FILE.tgz 
	--artifact-type file 
	--name yourEvidenceName 
	--build-url https://exampleci.com 
	--scan-results yourSnykJSONScanResults

```

**report Snyk vulnerability scan evidence about an artifact using an available Sha256 digest**

```shell
kosli report evidence artifact snyk 
	--fingerprint yourSha256 
	--name yourEvidenceName 
	--build-url https://exampleci.com 
	--scan-results yourSnykJSONScanResults
```



## kosli_report_evidence_commit_generic.md
---
title: "kosli report evidence commit generic"
beta: false
deprecated: true
summary: "Report Generic evidence for a commit in Kosli flows.  "
---

# kosli report evidence commit generic

{{% hint danger %}}
**kosli report evidence commit generic** is deprecated. See **kosli attest** commands.  Deprecated commands will be removed in a future release.
{{% /hint %}}
## Synopsis

Report Generic evidence for a commit in Kosli flows.  

```shell
kosli report evidence commit generic [flags]
```

## Flags
| Flag | Description |
| :--- | :--- |
|    -b, --build-url string  |  The url of CI pipeline that generated the evidence. (defaulted in some CIs: https://docs.kosli.com/ci-defaults ).  |
|        --commit string  |  Git commit for which to verify a given evidence. (defaulted in some CIs: https://docs.kosli.com/ci-defaults ).  |
|    -C, --compliant  |  [defaulted] Whether the evidence is compliant or not. A boolean flag https://docs.kosli.com/faq/#boolean-flags  |
|    -d, --description string  |  [optional] The evidence description.  |
|    -D, --dry-run  |  [optional] Run in dry-run mode. When enabled, no data is sent to Kosli and the CLI exits with 0 exit code regardless of any errors.  |
|        --evidence-fingerprint string  |  [optional] The SHA256 fingerprint of the evidence file or dir.  |
|    -e, --evidence-paths strings  |  [optional] The comma-separated list of paths containing supporting proof for the reported evidence. Paths can be for files or directories. All provided proofs will be uploaded to Kosli's evidence vault.  |
|        --evidence-url string  |  [optional] The external URL where the evidence file or dir is stored.  |
|    -f, --flows strings  |  [defaulted] The comma separated list of Kosli flows. Defaults to all flows of the org.  |
|    -h, --help  |  help for generic  |
|    -n, --name string  |  The name of the evidence.  |
|    -u, --user-data string  |  [optional] The path to a JSON file containing additional data you would like to attach to the evidence.  |


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

These examples all assume that the flags  `--api-token`, `--org`, `--host`, (and `--flow`, `--trail` when required), are set/provided. 

**report Generic evidence for a commit related to one Kosli flow**

```shell
kosli report evidence commit generic 
	--commit yourGitCommitSha1 
	--name yourEvidenceName 
	--description "some description" 
	--compliant 
	--flows yourFlowName 
	--build-url https://exampleci.com 

```

**report Generic evidence for a commit related to multiple Kosli flows with user-data**

```shell
kosli report evidence commit generic 
	--commit yourGitCommitSha1 
	--name yourEvidenceName 
	--description "some description" 
	--compliant 
	--flows yourFlowName1,yourFlowName2 
	--build-url https://exampleci.com 
	--user-data /path/to/json/file.json
```



## kosli_report_evidence_commit_jira.md
---
title: "kosli report evidence commit jira"
beta: false
deprecated: true
summary: "Report Jira evidence for a commit in Kosli flows."
---

# kosli report evidence commit jira

{{% hint danger %}}
**kosli report evidence commit jira** is deprecated. See **kosli attest** commands.  Deprecated commands will be removed in a future release.
{{% /hint %}}
## Synopsis

Report Jira evidence for a commit in Kosli flows.  
Parses the given commit's message or current branch name for Jira issue references of the 
form:  
'at least 2 characters long, starting with an uppercase letter project key followed by
dash and one or more digits'. 

The found issue references will be checked against Jira to confirm their existence.
The evidence is reported in all cases, and its compliance status depends on referencing
existing Jira issues.  
If you have wrong Jira credentials or wrong Jira-base-url it will be reported as non existing Jira issue.
This is because Jira returns same 404 error code in all cases.


```shell
kosli report evidence commit jira [flags]
```

## Flags
| Flag | Description |
| :--- | :--- |
|        --assert  |  [optional] Exit with non-zero code if no jira issue reference found, or jira issue does not exist, for the given commit or branch.  |
|    -b, --build-url string  |  The url of CI pipeline that generated the evidence. (defaulted in some CIs: https://docs.kosli.com/ci-defaults ).  |
|        --commit string  |  Git commit for which to verify a given evidence. (defaulted in some CIs: https://docs.kosli.com/ci-defaults ).  |
|    -D, --dry-run  |  [optional] Run in dry-run mode. When enabled, no data is sent to Kosli and the CLI exits with 0 exit code regardless of any errors.  |
|        --evidence-fingerprint string  |  [optional] The SHA256 fingerprint of the evidence file or dir.  |
|    -e, --evidence-paths strings  |  [optional] The comma-separated list of paths containing supporting proof for the reported evidence. Paths can be for files or directories. All provided proofs will be uploaded to Kosli's evidence vault.  |
|        --evidence-url string  |  [optional] The external URL where the evidence file or dir is stored.  |
|    -f, --flows strings  |  [defaulted] The comma separated list of Kosli flows. Defaults to all flows of the org.  |
|    -h, --help  |  help for jira  |
|        --jira-api-token string  |  Jira API token (for Jira Cloud)  |
|        --jira-base-url string  |  The base url for the jira project, e.g. 'https://kosli.atlassian.net'  |
|        --jira-pat string  |  Jira personal access token (for self-hosted Jira)  |
|        --jira-username string  |  Jira username (for Jira Cloud)  |
|    -n, --name string  |  The name of the evidence.  |
|        --repo-root string  |  [defaulted] The directory where the source git repository is available. (default ".")  |
|    -u, --user-data string  |  [optional] The path to a JSON file containing additional data you would like to attach to the evidence.  |


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

These examples all assume that the flags  `--api-token`, `--org`, `--host`, (and `--flow`, `--trail` when required), are set/provided. 

**report Jira evidence for a commit related to one Kosli flow (with Jira Cloud)**

```shell
kosli report evidence commit jira 
	--commit yourGitCommitSha1 
	--name yourEvidenceName 
	--jira-base-url https://kosli.atlassian.net 
	--jira-username user@domain.com 
	--jira-api-token yourJiraAPIToken 
	--flows yourFlowName 
	--build-url https://exampleci.com 

```

**report Jira evidence for a commit related to one Kosli flow (with self-hosted Jira)**

```shell
kosli report evidence commit jira 
	--commit yourGitCommitSha1 
	--name yourEvidenceName 
	--jira-base-url https://jira.example.com 
	--jira-pat yourJiraPATToken 
	--flows yourFlowName 
	--build-url https://exampleci.com 

```

**report Jira  evidence for a commit related to multiple Kosli flows with user-data (with Jira Cloud)**

```shell
kosli report evidence commit jira 
	--commit yourGitCommitSha1 
	--name yourEvidenceName 
	--jira-base-url https://kosli.atlassian.net 
	--jira-username user@domain.com 
	--jira-api-token yourJiraAPIToken 
	--flows yourFlowName1,yourFlowName2 
	--build-url https://exampleci.com 
	--user-data /path/to/json/file.json


```

**fail if no issue reference is found, or the issue is not found in your jira instance**

```shell
kosli report evidence commit jira 
	--commit yourGitCommitSha1 
	--name yourEvidenceName 
	--jira-base-url https://kosli.atlassian.net 
	--jira-username user@domain.com 
	--jira-api-token yourJiraAPIToken 
	--flows yourFlowName 
	--build-url https://exampleci.com 
	--assert
```



## kosli_report_evidence_commit_junit.md
---
title: "kosli report evidence commit junit"
beta: false
deprecated: true
summary: "Report JUnit test evidence for a commit in Kosli flows.  "
---

# kosli report evidence commit junit

{{% hint danger %}}
**kosli report evidence commit junit** is deprecated. See **kosli attest** commands.  Deprecated commands will be removed in a future release.
{{% /hint %}}
## Synopsis

Report JUnit test evidence for a commit in Kosli flows.    
All .xml files from --results-dir are parsed and uploaded to Kosli's evidence vault.  
If there are no failing tests and no errors the evidence is reported as compliant. Otherwise the evidence is reported as non-compliant.


```shell
kosli report evidence commit junit [flags]
```

## Flags
| Flag | Description |
| :--- | :--- |
|    -b, --build-url string  |  The url of CI pipeline that generated the evidence. (defaulted in some CIs: https://docs.kosli.com/ci-defaults ).  |
|        --commit string  |  Git commit for which to verify a given evidence. (defaulted in some CIs: https://docs.kosli.com/ci-defaults ).  |
|    -D, --dry-run  |  [optional] Run in dry-run mode. When enabled, no data is sent to Kosli and the CLI exits with 0 exit code regardless of any errors.  |
|        --evidence-fingerprint string  |  [optional] The SHA256 fingerprint of the evidence file or dir.  |
|        --evidence-url string  |  [optional] The external URL where the evidence file or dir is stored.  |
|    -f, --flows strings  |  [defaulted] The comma separated list of Kosli flows. Defaults to all flows of the org.  |
|    -h, --help  |  help for junit  |
|    -n, --name string  |  The name of the evidence.  |
|    -R, --results-dir string  |  [defaulted] The path to a directory with JUnit test results. By default, the directory will be uploaded to Kosli's evidence vault. (default ".")  |
|    -u, --user-data string  |  [optional] The path to a JSON file containing additional data you would like to attach to the evidence.  |


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

These examples all assume that the flags  `--api-token`, `--org`, `--host`, (and `--flow`, `--trail` when required), are set/provided. 

**report JUnit test evidence for a commit related to one Kosli flow**

```shell
kosli report evidence commit junit 
	--commit yourGitCommitSha1 
	--name yourEvidenceName 
	--flows yourFlowName 
	--build-url https://exampleci.com 
	--results-dir yourFolderWithJUnitResults

```

**report JUnit test evidence for a commit related to multiple Kosli flows**

```shell
kosli report evidence commit junit 
	--commit yourGitCommitSha1 
	--name yourEvidenceName 
	--flows yourFlowName1,yourFlowName2 
	--build-url https://exampleci.com 
	--results-dir yourFolderWithJUnitResults
```



## kosli_report_evidence_commit_pullrequest_azure.md
---
title: "kosli report evidence commit pullrequest azure"
beta: false
deprecated: true
summary: "Report Azure Devops pull request evidence for a git commit in Kosli flows.  "
---

# kosli report evidence commit pullrequest azure

{{% hint danger %}}
**kosli report evidence commit pullrequest azure** is deprecated. See **kosli attest** commands.  Deprecated commands will be removed in a future release.
{{% /hint %}}
## Synopsis

Report Azure Devops pull request evidence for a git commit in Kosli flows.  
It checks if a pull request exists for a commit and report the pull-request evidence to the commit in Kosli. 


```shell
kosli report evidence commit pullrequest azure [flags]
```

## Flags
| Flag | Description |
| :--- | :--- |
|        --assert  |  [optional] Exit with non-zero code if no pull requests found for the given commit.  |
|        --azure-org-url string  |  Azure organization url. E.g. "https://dev.azure.com/myOrg" (defaulted if you are running in Azure Devops pipelines: https://docs.kosli.com/ci-defaults ).  |
|        --azure-token string  |  Azure Personal Access token.  |
|    -b, --build-url string  |  The url of CI pipeline that generated the evidence. (defaulted in some CIs: https://docs.kosli.com/ci-defaults ).  |
|        --commit string  |  Git commit for which to verify a given evidence. (defaulted in some CIs: https://docs.kosli.com/ci-defaults ).  |
|    -D, --dry-run  |  [optional] Run in dry-run mode. When enabled, no data is sent to Kosli and the CLI exits with 0 exit code regardless of any errors.  |
|        --evidence-fingerprint string  |  [optional] The SHA256 fingerprint of the evidence file or dir.  |
|        --evidence-url string  |  [optional] The external URL where the evidence file or dir is stored.  |
|    -f, --flows strings  |  [defaulted] The comma separated list of Kosli flows. Defaults to all flows of the org.  |
|    -h, --help  |  help for azure  |
|    -n, --name string  |  The name of the evidence.  |
|        --project string  |  Azure project.(defaulted if you are running in Azure Devops pipelines: https://docs.kosli.com/ci-defaults ).  |
|        --repository string  |  Git repository. (defaulted in some CIs: https://docs.kosli.com/ci-defaults ).  |
|    -u, --user-data string  |  [optional] The path to a JSON file containing additional data you would like to attach to the evidence.  |


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

These examples all assume that the flags  `--api-token`, `--org`, `--host`, (and `--flow`, `--trail` when required), are set/provided. 

**report a pull request commit evidence to Kosli**

```shell
kosli report evidence commit pullrequest azure 
	--commit yourGitCommitSha1 
	--azure-org-url https://dev.azure.com/myOrg 
	--project yourAzureDevOpsProject 
	--repository yourAzureGitRepository 
	--azure-token yourAzureToken 
	--name yourEvidenceName 
	--flows yourFlowName1,yourFlowName2 
	--build-url https://exampleci.com 

```

**fail if a pull request does not exist for your commit**

```shell
kosli report evidence commit pullrequest azure 
	--commit yourGitCommitSha1 
	--azure-org-url https://dev.azure.com/myOrg 
	--project yourAzureDevOpsProject 
	--repository yourAzureGitRepository 
	--azure-token yourAzureToken 
	--name yourEvidenceName 
	--flows yourFlowName1,yourFlowName2 
	--build-url https://exampleci.com 
	--assert
```



## kosli_report_evidence_commit_pullrequest_bitbucket.md
---
title: "kosli report evidence commit pullrequest bitbucket"
beta: false
deprecated: true
summary: "Report Bitbucket pull request evidence for a commit in Kosli flows.  "
---

# kosli report evidence commit pullrequest bitbucket

{{% hint danger %}}
**kosli report evidence commit pullrequest bitbucket** is deprecated. See **kosli attest** commands.  Deprecated commands will be removed in a future release.
{{% /hint %}}
## Synopsis

Report Bitbucket pull request evidence for a commit in Kosli flows.  
It checks if a pull request exists for the git commit and reports the pull-request evidence to the commit in Kosli.

```shell
kosli report evidence commit pullrequest bitbucket [flags]
```

## Flags
| Flag | Description |
| :--- | :--- |
|        --assert  |  [optional] Exit with non-zero code if no pull requests found for the given commit.  |
|        --bitbucket-access-token string  |  Bitbucket repo/project/workspace access token. See https://developer.atlassian.com/cloud/bitbucket/rest/intro/#access-tokens for more details.  |
|        --bitbucket-password string  |  Bitbucket App password. See https://developer.atlassian.com/cloud/bitbucket/rest/intro/#authentication for more details.  |
|        --bitbucket-username string  |  Bitbucket username. Only needed if you use --bitbucket-password  |
|        --bitbucket-workspace string  |  Bitbucket workspace ID.  |
|    -b, --build-url string  |  The url of CI pipeline that generated the evidence. (defaulted in some CIs: https://docs.kosli.com/ci-defaults ).  |
|        --commit string  |  Git commit for which to verify a given evidence. (defaulted in some CIs: https://docs.kosli.com/ci-defaults ).  |
|    -D, --dry-run  |  [optional] Run in dry-run mode. When enabled, no data is sent to Kosli and the CLI exits with 0 exit code regardless of any errors.  |
|        --evidence-fingerprint string  |  [optional] The SHA256 fingerprint of the evidence file or dir.  |
|        --evidence-url string  |  [optional] The external URL where the evidence file or dir is stored.  |
|    -f, --flows strings  |  [defaulted] The comma separated list of Kosli flows. Defaults to all flows of the org.  |
|    -h, --help  |  help for bitbucket  |
|    -n, --name string  |  The name of the evidence.  |
|        --repository string  |  Git repository. (defaulted in some CIs: https://docs.kosli.com/ci-defaults ).  |
|    -u, --user-data string  |  [optional] The path to a JSON file containing additional data you would like to attach to the evidence.  |


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

These examples all assume that the flags  `--api-token`, `--org`, `--host`, (and `--flow`, `--trail` when required), are set/provided. 

**report a pull request evidence to Kosli**

```shell
kosli report evidence commit pullrequest bitbucket 
	--commit yourArtifactGitCommit 
	--repository yourBitbucketGitRepository 
	--bitbucket-username yourBitbucketUsername 
	--bitbucket-password yourBitbucketPassword 
	--bitbucket-workspace yourBitbucketWorkspace 
	--name yourEvidenceName 
	--flows yourFlowName1,yourFlowName2 
	--build-url https://exampleci.com 

```

**fail if a pull request does not exist for your commit**

```shell
kosli report evidence commit pullrequest bitbucket 
	--commit yourArtifactGitCommit 
	--repository yourBitbucketGitRepository 
	--bitbucket-username yourBitbucketUsername 
	--bitbucket-password yourBitbucketPassword 
	--bitbucket-workspace yourBitbucketWorkspace 
	--name yourEvidenceName 
	--flows yourFlowName1,yourFlowName2 
	--build-url https://exampleci.com 
	--assert
```



## kosli_report_evidence_commit_pullrequest_github.md
---
title: "kosli report evidence commit pullrequest github"
beta: false
deprecated: true
summary: "Report Github pull request evidence for a git commit in Kosli flows.  "
---

# kosli report evidence commit pullrequest github

{{% hint danger %}}
**kosli report evidence commit pullrequest github** is deprecated. See **kosli attest** commands.  Deprecated commands will be removed in a future release.
{{% /hint %}}
## Synopsis

Report Github pull request evidence for a git commit in Kosli flows.  
It checks if a pull request exists for a commit and report the pull-request evidence to the commit in Kosli. 


```shell
kosli report evidence commit pullrequest github [flags]
```

## Flags
| Flag | Description |
| :--- | :--- |
|        --assert  |  [optional] Exit with non-zero code if no pull requests found for the given commit.  |
|    -b, --build-url string  |  The url of CI pipeline that generated the evidence. (defaulted in some CIs: https://docs.kosli.com/ci-defaults ).  |
|        --commit string  |  Git commit for which to verify a given evidence. (defaulted in some CIs: https://docs.kosli.com/ci-defaults ).  |
|    -D, --dry-run  |  [optional] Run in dry-run mode. When enabled, no data is sent to Kosli and the CLI exits with 0 exit code regardless of any errors.  |
|        --evidence-fingerprint string  |  [optional] The SHA256 fingerprint of the evidence file or dir.  |
|        --evidence-url string  |  [optional] The external URL where the evidence file or dir is stored.  |
|    -f, --flows strings  |  [defaulted] The comma separated list of Kosli flows. Defaults to all flows of the org.  |
|        --github-base-url string  |  [optional] GitHub base URL (only needed for GitHub Enterprise installations).  |
|        --github-org string  |  Github organization. (defaulted if you are running in GitHub Actions: https://docs.kosli.com/ci-defaults ).  |
|        --github-token string  |  Github token.  |
|    -h, --help  |  help for github  |
|    -n, --name string  |  The name of the evidence.  |
|        --repository string  |  Git repository. (defaulted in some CIs: https://docs.kosli.com/ci-defaults ).  |
|    -u, --user-data string  |  [optional] The path to a JSON file containing additional data you would like to attach to the evidence.  |


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

These examples all assume that the flags  `--api-token`, `--org`, `--host`, (and `--flow`, `--trail` when required), are set/provided. 

**report a pull request commit evidence to Kosli**

```shell
kosli report evidence commit pullrequest github 
	--commit yourGitCommitSha1 
	--repository yourGithubGitRepository 
	--github-token yourGithubToken 
	--github-org yourGithubOrg 
	--name yourEvidenceName 
	--flows yourFlowName1,yourFlowName2 
	--build-url https://exampleci.com 

```

**fail if a pull request does not exist for your commit**

```shell
kosli report evidence commit pullrequest github 
	--commit yourGitCommitSha1 
	--repository yourGithubGitRepository 
	--github-token yourGithubToken 
	--github-org yourGithubOrg 
	--name yourEvidenceName 
	--flows yourFlowName1,yourFlowName2 
	--build-url https://exampleci.com 
	--assert
```



## kosli_report_evidence_commit_pullrequest_gitlab.md
---
title: "kosli report evidence commit pullrequest gitlab"
beta: false
deprecated: true
summary: "Report Gitlab merge request evidence for a commit in Kosli flows.  "
---

# kosli report evidence commit pullrequest gitlab

{{% hint danger %}}
**kosli report evidence commit pullrequest gitlab** is deprecated. See **kosli attest** commands.  Deprecated commands will be removed in a future release.
{{% /hint %}}
## Synopsis

Report Gitlab merge request evidence for a commit in Kosli flows.  
It checks if a merge request exists for the git commit and reports the merge-request evidence to the commit in Kosli.

```shell
kosli report evidence commit pullrequest gitlab [flags]
```

## Flags
| Flag | Description |
| :--- | :--- |
|        --assert  |  [optional] Exit with non-zero code if no pull requests found for the given commit.  |
|    -b, --build-url string  |  The url of CI pipeline that generated the evidence. (defaulted in some CIs: https://docs.kosli.com/ci-defaults ).  |
|        --commit string  |  Git commit for which to verify a given evidence. (defaulted in some CIs: https://docs.kosli.com/ci-defaults ).  |
|    -D, --dry-run  |  [optional] Run in dry-run mode. When enabled, no data is sent to Kosli and the CLI exits with 0 exit code regardless of any errors.  |
|        --evidence-fingerprint string  |  [optional] The SHA256 fingerprint of the evidence file or dir.  |
|        --evidence-url string  |  [optional] The external URL where the evidence file or dir is stored.  |
|    -f, --flows strings  |  [defaulted] The comma separated list of Kosli flows. Defaults to all flows of the org.  |
|        --gitlab-base-url string  |  [optional] Gitlab base URL (only needed for on-prem Gitlab installations).  |
|        --gitlab-org string  |  Gitlab organization. (defaulted if you are running in Gitlab Pipelines: https://docs.kosli.com/ci-defaults ).  |
|        --gitlab-token string  |  Gitlab token.  |
|    -h, --help  |  help for gitlab  |
|    -n, --name string  |  The name of the evidence.  |
|        --repository string  |  Git repository. (defaulted in some CIs: https://docs.kosli.com/ci-defaults ).  |
|    -u, --user-data string  |  [optional] The path to a JSON file containing additional data you would like to attach to the evidence.  |


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

These examples all assume that the flags  `--api-token`, `--org`, `--host`, (and `--flow`, `--trail` when required), are set/provided. 

**report a merge request evidence to Kosli**

```shell
kosli report evidence commit pullrequest gitlab 
	--commit yourArtifactGitCommit 
	--repository yourBitbucketGitRepository 
	--gitlab-token yourGitlabToken 
	--gitlab-org yourGitlabOrg 
	--name yourEvidenceName 
	--flows yourFlowName1,yourFlowName2 
	--build-url https://exampleci.com 

```

**fail if a pull request does not exist for your commit**

```shell
kosli report evidence commit pullrequest gitlab 
	--commit yourArtifactGitCommit 
	--repository yourBitbucketGitRepository 
	--gitlab-token yourGitlabToken 
	--gitlab-org yourGitlabOrg 
	--name yourEvidenceName 
	--flows yourFlowName1,yourFlowName2 
	--build-url https://exampleci.com 
	--assert
```



## kosli_report_evidence_commit_snyk.md
---
title: "kosli report evidence commit snyk"
beta: false
deprecated: true
summary: "Report Snyk vulnerability scan evidence for a commit in Kosli flows.  "
---

# kosli report evidence commit snyk

{{% hint danger %}}
**kosli report evidence commit snyk** is deprecated. See **kosli attest** commands.  Deprecated commands will be removed in a future release.
{{% /hint %}}
## Synopsis

Report Snyk vulnerability scan evidence for a commit in Kosli flows.    
The --scan-results .json file is parsed and uploaded to Kosli's evidence vault.

In CLI <v2.8.2, Snyk results could only be in the Snyk JSON output format. "snyk code test" results were not supported by 
this command and could be reported as generic evidence.

Starting from v2.8.2, the Snyk results can be in Snyk JSON or SARIF output format for "snyk container test". 
"snyk code test" is now supported but only in the SARIF format.

If no vulnerabilities are detected the evidence is reported as compliant. Otherwise the evidence is reported as non-compliant.


```shell
kosli report evidence commit snyk [flags]
```

## Flags
| Flag | Description |
| :--- | :--- |
|    -b, --build-url string  |  The url of CI pipeline that generated the evidence. (defaulted in some CIs: https://docs.kosli.com/ci-defaults ).  |
|        --commit string  |  Git commit for which to verify a given evidence. (defaulted in some CIs: https://docs.kosli.com/ci-defaults ).  |
|    -D, --dry-run  |  [optional] Run in dry-run mode. When enabled, no data is sent to Kosli and the CLI exits with 0 exit code regardless of any errors.  |
|        --evidence-fingerprint string  |  [optional] The SHA256 fingerprint of the evidence file or dir.  |
|        --evidence-url string  |  [optional] The external URL where the evidence file or dir is stored.  |
|    -f, --flows strings  |  [defaulted] The comma separated list of Kosli flows. Defaults to all flows of the org.  |
|    -h, --help  |  help for snyk  |
|    -n, --name string  |  The name of the evidence.  |
|    -R, --scan-results string  |  The path to Snyk SARIF or JSON scan results file from 'snyk test' and 'snyk container test'. By default, the Snyk results will be uploaded to Kosli's evidence vault.  |
|        --upload-results  |  [defaulted] Whether to upload the provided Snyk results file as an attachment to Kosli or not. (default true)  |
|    -u, --user-data string  |  [optional] The path to a JSON file containing additional data you would like to attach to the evidence.  |


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

These examples all assume that the flags  `--api-token`, `--org`, `--host`, (and `--flow`, `--trail` when required), are set/provided. 

**report Snyk evidence for a commit related to one Kosli flow**

```shell
kosli report evidence commit snyk 
	--commit yourGitCommitSha1 
	--name yourEvidenceName 
	--flows yourFlowName1 
	--build-url https://exampleci.com 
	--scan-results yourSnykJSONScanResults

```

**report Snyk evidence for a commit related to multiple Kosli flows**

```shell
kosli report evidence commit snyk 
	--commit yourGitCommitSha1 
	--name yourEvidenceName 
	--flows yourFlowName1,yourFlowName2 
	--build-url https://exampleci.com 
	--scan-results yourSnykJSONScanResults
```



## kosli_request_approval.md
---
title: "kosli request approval"
beta: false
deprecated: false
summary: "Request an approval of a deployment of an artifact to an environment in Kosli.  "
---

# kosli request approval

## Synopsis

Request an approval of a deployment of an artifact to an environment in Kosli.  
The request should be reviewed in the Kosli UI.  

The artifact fingerprint can be provided directly with the `--fingerprint` flag, or 
calculated based on `--artifact-type` flag.

Artifact type can be one of: "file" for files, "dir" for directories, "oci" for container
images in registries or "docker" for local docker images.



```shell
kosli request approval [IMAGE-NAME | FILE-PATH | DIR-PATH] [flags]
```

## Flags
| Flag | Description |
| :--- | :--- |
|    -t, --artifact-type string  |  The type of the artifact to calculate its SHA256 fingerprint. One of: [oci, docker, file, dir]. Only required if you want Kosli to calculate the fingerprint for you (i.e. when you don't specify '--fingerprint' on commands that allow it).  |
|    -d, --description string  |  [optional] The approval description.  |
|    -D, --dry-run  |  [optional] Run in dry-run mode. When enabled, no data is sent to Kosli and the CLI exits with 0 exit code regardless of any errors.  |
|    -e, --environment string  |  [defaulted] The environment the artifact is approved for. (defaults to all environments)  |
|    -x, --exclude strings  |  [optional] The comma separated list of directories and files to exclude from fingerprinting. Can take glob patterns. Only applicable for --artifact-type dir.  |
|    -F, --fingerprint string  |  [conditional] The SHA256 fingerprint of the artifact. Only required if you don't specify '--artifact-type'.  |
|    -f, --flow string  |  The Kosli flow name.  |
|    -h, --help  |  help for approval  |
|        --newest-commit string  |  [defaulted] The source commit sha for the newest change in the deployment. Can be any commit-ish. (default "HEAD")  |
|        --oldest-commit string  |  [conditional] The source commit sha for the oldest change in the deployment. Can be any commit-ish. Only required if you don't specify '--environment'.  |
|        --registry-password string  |  [conditional] The container registry password or access token. Only required if you want to read container image SHA256 digest from a remote container registry.  |
|        --registry-username string  |  [conditional] The container registry username. Only required if you want to read container image SHA256 digest from a remote container registry.  |
|        --repo-root string  |  [defaulted] The directory where the source git repository is available. (default ".")  |
|    -u, --user-data string  |  [optional] The path to a JSON file containing additional data you would like to attach to the approval.  |


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

These examples all assume that the flags  `--api-token`, `--org`, `--host`, (and `--flow`, `--trail` when required), are set/provided. 

```shell
# Request an approval for an artifact with a provided fingerprint (sha256)
# for deployment to environment <yourEnvironmentName>.
# The approval is for all git commits since the last approval to this environment.
kosli request approval \
	--api-token yourAPIToken \
	--description "An optional description for the approval" \
	--environment yourEnvironmentName \
	--org yourOrgName \
	--flow yourFlowName \
	--fingerprint yourArtifactFingerprint

# Request that a file type artifact needs approval for deployment to environment <yourEnvironmentName>.
# The approval is for all git commits since the last approval to this environment.
kosli request approval FILE.tgz \
	--api-token yourAPIToken \
	--artifact-type file \
	--description "An optional description for the requested approval" \
	--environment yourEnvironmentName \
	--newest-commit HEAD \
	--org yourOrgName \
	--flow yourFlowName 

# Request an approval for an artifact with a provided fingerprint (sha256).
# The approval is for all environments.
# The approval is for all commits since the git commit of origin/production branch.
kosli request approval \
	--api-token yourAPIToken \
	--description "An optional description for the requested approval" \
	--newest-commit HEAD \
	--oldest-commit origin/production \
	--org yourOrgName \
	--flow yourFlowName \
	--fingerprint yourArtifactFingerprint
```



## kosli_search.md
---
title: "kosli search"
beta: false
deprecated: false
summary: "Search for a git commit or an artifact fingerprint in Kosli.  "
---

# kosli search

## Synopsis

Search for a git commit or an artifact fingerprint in Kosli.   
You can use short git commit or artifact fingerprint shas, but you must provide at least 5 characters.

```shell
kosli search {GIT-COMMIT | FINGERPRINT} [flags]
```

## Flags
| Flag | Description |
| :--- | :--- |
|    -h, --help  |  help for search  |
|    -o, --output string  |  [defaulted] The format of the output. Valid formats are: [table, json]. (default "table")  |


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

These examples all assume that the flags  `--api-token`, `--org`, `--host`, (and `--flow`, `--trail` when required), are set/provided. 

**Search for a git commit in Kosli**

```shell
kosli search YOUR_GIT_COMMIT 

```

**Search for an artifact fingerprint in Kosli**

```shell
kosli search YOUR_ARTIFACT_FINGERPRINT 
```



## kosli_snapshot_azure.md
---
title: "kosli snapshot azure"
beta: false
deprecated: false
summary: "Report a snapshot of running Azure web apps and function apps in an Azure resource group to Kosli.  "
---

# kosli snapshot azure

## Synopsis

Report a snapshot of running Azure web apps and function apps in an Azure resource group to Kosli.  
The reported data includes Azure app names, container image digests and creation timestamps.

For Azure Function apps or Web apps which uses zip deployment the fingerprint is calculated based on the
content of the zip file. This is the same as unzipping the file and then running `kosli fingerprint -t dir yourDirName`.
When doing zip deployment the WEBSITE_RUN_FROM_PACKAGE must NOT be set to 1. This will cause the azure
API calls to not return the content of what is running on the server and fingerprint calculations
will not match. See 
https://learn.microsoft.com/en-us/azure/azure-functions/functions-app-settings#website_run_from_package

To authenticate to Azure, you need to create Azure service principal with a secret  
and provide these Azure credentials via flags or by exporting the equivalent KOSLI env vars (e.g. KOSLI_AZURE_CLIENT_ID).  
The service principal needs to have the following permissions:  
  1) Microsoft.Web/sites/Read  
  2) Microsoft.ContainerRegistry/registries/pull/read  

	

```shell
kosli snapshot azure ENVIRONMENT-NAME [flags]
```

## Flags
| Flag | Description |
| :--- | :--- |
|        --azure-client-id string  |  Azure client ID.  |
|        --azure-client-secret string  |  Azure client secret.  |
|        --azure-resource-group-name string  |  Azure resource group name.  |
|        --azure-subscription-id string  |  Azure subscription ID.  |
|        --azure-tenant-id string  |  Azure tenant ID.  |
|        --digests-source string  |  [defaulted] Where to get the digests from. Valid values are 'acr' and 'logs'. (default "acr")  |
|    -D, --dry-run  |  [optional] Run in dry-run mode. When enabled, no data is sent to Kosli and the CLI exits with 0 exit code regardless of any errors.  |
|    -h, --help  |  help for azure  |
|        --zip  |  Download logs from Azure as zip files  |


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

These examples all assume that the flags  `--api-token`, `--org`, `--host`, (and `--flow`, `--trail` when required), are set/provided. 

**Use Azure Container Registry to get the digests for artifacts in a snapshot**

```shell
kosli snapshot azure yourEnvironmentName 
	--azure-client-id yourAzureClientID 
	--azure-client-secret yourAzureClientSecret 
	--azure-tenant-id yourAzureTenantID 
	--azure-subscription-id yourAzureSubscriptionID 
	--azure-resource-group-name yourAzureResourceGroupName 
	--digests-source acr 

```

**Use Docker logs of Azure apps to get the digests for artifacts in a snapshot**

```shell
kosli snapshot azure yourEnvironmentName 
	--azure-client-id yourAzureClientID 
	--azure-client-secret yourAzureClientSecret 
	--azure-tenant-id yourAzureTenantID 
	--azure-subscription-id yourAzureSubscriptionID 
	--azure-resource-group-name yourAzureResourceGroupName 
	--digests-source logs 

```

**Report digest of an Azure Function app**

```shell
kosli snapshot azure yourEnvironmentName 
	--azure-client-id yourAzureClientID 
	--azure-client-secret yourAzureClientSecret 
	--azure-tenant-id yourAzureTenantID 
	--azure-subscription-id yourAzureSubscriptionID 
	--azure-resource-group-name yourAzureResourceGroupName 
```



## kosli_snapshot_docker.md
---
title: "kosli snapshot docker"
beta: false
deprecated: false
summary: "Report a snapshot of running containers from docker host to Kosli.  "
---

# kosli snapshot docker

## Synopsis

Report a snapshot of running containers from docker host to Kosli.  
The reported data includes container image digests 
and creation timestamps. Containers running images which have not
been pushed to or pulled from a registry will be ignored.

```shell
kosli snapshot docker ENVIRONMENT-NAME [flags]
```

## Flags
| Flag | Description |
| :--- | :--- |
|    -D, --dry-run  |  [optional] Run in dry-run mode. When enabled, no data is sent to Kosli and the CLI exits with 0 exit code regardless of any errors.  |
|    -h, --help  |  help for docker  |


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

These examples all assume that the flags  `--api-token`, `--org`, `--host`, (and `--flow`, `--trail` when required), are set/provided. 

**report what is running in a docker host**

```shell
kosli snapshot docker yourEnvironmentName 
```



## kosli_snapshot_ecs.md
---
title: "kosli snapshot ecs"
beta: false
deprecated: false
summary: "Report a snapshot of running containers in one or more AWS ECS cluster(s) to Kosli.  "
---

# kosli snapshot ecs

## Synopsis

Report a snapshot of running containers in one or more AWS ECS cluster(s) to Kosli.  
Skip `--clusters` and `--clusters-regex` to report all clusters in a given AWS account. Or use `--exclude` and/or `--exclude-regex` to report all clusters excluding some.
The reported data includes container image digests and creation timestamps.

To authenticate to AWS, you can either:  
  1) provide the AWS static credentials via flags or by exporting the equivalent KOSLI env vars (e.g. KOSLI_AWS_KEY_ID)  
  2) export the AWS env vars (e.g. AWS_ACCESS_KEY_ID).  
  3) Use a shared config/credentials file under the $HOME/.aws  
  
Option 1 takes highest precedence, while option 3 is the lowest.  
More details can be found here: https://aws.github.io/aws-sdk-go-v2/docs/configuring-sdk/#specifying-credentials
	

```shell
kosli snapshot ecs ENVIRONMENT-NAME [flags]
```

## Flags
| Flag | Description |
| :--- | :--- |
|        --aws-key-id string  |  The AWS access key ID.  |
|        --aws-region string  |  The AWS region.  |
|        --aws-secret-key string  |  The AWS secret access key.  |
|        --clusters strings  |  [optional] The comma-separated list of ECS cluster names to snapshot. Can't be used together with --exclude or --exclude-regex.  |
|        --clusters-regex strings  |  [optional] The comma-separated list of ECS cluster name regex patterns to snapshot. Can't be used together with --exclude or --exclude-regex.  |
|    -D, --dry-run  |  [optional] Run in dry-run mode. When enabled, no data is sent to Kosli and the CLI exits with 0 exit code regardless of any errors.  |
|        --exclude strings  |  [optional] The comma-separated list of ECS cluster names to exclude. Can't be used together with --exclude or --exclude-regex.  |
|        --exclude-regex strings  |  [optional] The comma-separated list of ECS cluster name regex patterns to exclude. Can't be used together with --clusters or --clusters-regex.  |
|    -h, --help  |  help for ecs  |


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

These examples all assume that the flags  `--api-token`, `--org`, `--host`, (and `--flow`, `--trail` when required), are set/provided. 

**report what is running in an entire AWS ECS cluster**

```shell
export AWS_REGION=yourAWSRegion
export AWS_ACCESS_KEY_ID=yourAWSAccessKeyID
export AWS_SECRET_ACCESS_KEY=yourAWSSecretAccessKey

kosli snapshot ecs yourEnvironmentName 
	--clusters yourECSClusterName 

```

**report what is running in a specific AWS ECS service within a cluster**

```shell
export AWS_REGION=yourAWSRegion
export AWS_ACCESS_KEY_ID=yourAWSAccessKeyID
export AWS_SECRET_ACCESS_KEY=yourAWSSecretAccessKey

kosli snapshot ecs yourEnvironmentName 
	--clusters yourECSClusterName 
	--service-name yourECSServiceName 

```

**report what is running in all ECS clusters in an AWS account (AWS auth provided in flags)**

```shell
kosli snapshot ecs yourEnvironmentName 
	--aws-key-id yourAWSAccessKeyID 
	--aws-secret-key yourAWSSecretAccessKey 
	--aws-region yourAWSRegion 

```

**report what is running in all ECS clusters in an AWS account except for clusters with names matching given regex patterns**

```shell
kosli snapshot ecs yourEnvironmentName 
	--aws-key-id yourAWSAccessKeyID 
	--aws-secret-key yourAWSSecretAccessKey 
	--aws-region yourAWSRegion 
	--exclude-regex "those-names.*" 
```



## kosli_snapshot_k8s.md
---
title: "kosli snapshot k8s"
beta: false
deprecated: false
summary: "Report a snapshot of running pods in a K8S cluster or namespace(s) to Kosli.  "
---

# kosli snapshot k8s

## Synopsis

Report a snapshot of running pods in a K8S cluster or namespace(s) to Kosli.  
Skip `--namespaces` and `--namespaces-regex` to report all pods in all namespaces in a cluster.
The reported data includes pod container images digests and creation timestamps. You can customize the scope of reporting
to include or exclude namespaces.

```shell
kosli snapshot k8s ENVIRONMENT-NAME [flags]
```

## Flags
| Flag | Description |
| :--- | :--- |
|    -D, --dry-run  |  [optional] Run in dry-run mode. When enabled, no data is sent to Kosli and the CLI exits with 0 exit code regardless of any errors.  |
|    -x, --exclude-namespaces strings  |  [optional] The comma separated list of namespaces names to exclude from reporting artifacts info from. Requires cluster-wide read permissions for pods and namespaces. Can't be used together with --namespaces or --namespaces-regex.  |
|        --exclude-namespaces-regex strings  |  [optional] The comma separated list of namespaces regex patterns to exclude from reporting artifacts info from. Requires cluster-wide read permissions for pods and namespaces. Can't be used together with --namespaces or --namespaces-regex.  |
|    -h, --help  |  help for k8s  |
|    -k, --kubeconfig string  |  [defaulted] The kubeconfig path for the target cluster. (default "$HOME/.kube/config")  |
|    -n, --namespaces strings  |  [optional] The comma separated list of namespaces names to report artifacts info from. Can't be used together with --exclude-namespaces or --exclude-namespaces-regex.  |
|        --namespaces-regex strings  |  [optional] The comma separated list of namespaces regex patterns to report artifacts info from. Requires cluster-wide read permissions for pods and namespaces. Can't be used together with --exclude-namespaces --exclude-namespaces-regex.  |


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

These examples all assume that the flags  `--api-token`, `--org`, `--host`, (and `--flow`, `--trail` when required), are set/provided. 

**report what is running in an entire cluster using kubeconfig at $HOME/.kube/config**

```shell
kosli snapshot k8s yourEnvironmentName 

```

**report what is running in an entire cluster using kubeconfig at $HOME/.kube/config**

```shell
(with global flags defined in environment or in a config file):
export KOSLI_API_TOKEN=yourAPIToken
export KOSLI_ORG=yourOrgName

kosli snapshot k8s yourEnvironmentName

```

**report what is running in an entire cluster excluding some namespaces using kubeconfig at $HOME/.kube/config**

```shell
kosli snapshot k8s yourEnvironmentName 
    --exclude-namespaces kube-system,utilities 

```

**report what is running in a given namespace in the cluster using kubeconfig at $HOME/.kube/config**

```shell
kosli snapshot k8s yourEnvironmentName 
	--namespaces your-namespace 

```

**report what is running in a cluster using kubeconfig at a custom path**

```shell
kosli snapshot k8s yourEnvironmentName 
	--kubeconfig /path/to/kube/config 
```



## kosli_snapshot_lambda.md
---
title: "kosli snapshot lambda"
beta: false
deprecated: false
summary: "Report a snapshot of artifacts deployed as one or more AWS Lambda functions and their digests to Kosli."
---

# kosli snapshot lambda

## Synopsis

Report a snapshot of artifacts deployed as one or more AWS Lambda functions and their digests to Kosli.  
Skip `--function-names` and `--function-names-regex` to report all functions in a given AWS account. Or use `--exclude` and/or `--exclude-regex` to report all functions excluding some.

To authenticate to AWS, you can either:  
  1) provide the AWS static credentials via flags or by exporting the equivalent KOSLI env vars (e.g. KOSLI_AWS_KEY_ID)  
  2) export the AWS env vars (e.g. AWS_ACCESS_KEY_ID).  
  3) Use a shared config/credentials file under the $HOME/.aws  
  
Option 1 takes highest precedence, while option 3 is the lowest.  
More details can be found here: https://aws.github.io/aws-sdk-go-v2/docs/configuring-sdk/#specifying-credentials
	

```shell
kosli snapshot lambda ENVIRONMENT-NAME [flags]
```

## Flags
| Flag | Description |
| :--- | :--- |
|        --aws-key-id string  |  The AWS access key ID.  |
|        --aws-region string  |  The AWS region.  |
|        --aws-secret-key string  |  The AWS secret access key.  |
|    -D, --dry-run  |  [optional] Run in dry-run mode. When enabled, no data is sent to Kosli and the CLI exits with 0 exit code regardless of any errors.  |
|        --exclude strings  |  [optional] The comma-separated list of AWS Lambda function names to be excluded. Cannot be used together with --function-names  |
|        --exclude-regex strings  |  [optional] The comma-separated list of name regex patterns for AWS Lambda functions to be excluded. Cannot be used together with --function-names. Allowed regex patterns are described in https://github.com/google/re2/wiki/Syntax  |
|        --function-names strings  |  [optional] The comma-separated list of AWS Lambda function names to be reported. Cannot be used together with --exclude or --exclude-regex.  |
|        --function-names-regex strings  |  [optional] The comma-separated list of AWS Lambda function names regex patterns to be reported. Cannot be used together with --exclude or --exclude-regex.  |
|    -h, --help  |  help for lambda  |


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

These examples all assume that the flags  `--api-token`, `--org`, `--host`, (and `--flow`, `--trail` when required), are set/provided. 

**report all Lambda functions running in an AWS account (AWS auth provided in env variables)**

```shell
export AWS_REGION=yourAWSRegion
export AWS_ACCESS_KEY_ID=yourAWSAccessKeyID
export AWS_SECRET_ACCESS_KEY=yourAWSSecretAccessKey

kosli snapshot lambda yourEnvironmentName 

```

**report all (excluding some) Lambda functions running in an AWS account (AWS auth provided in env variables)**

```shell
export AWS_REGION=yourAWSRegion
export AWS_ACCESS_KEY_ID=yourAWSAccessKeyID
export AWS_SECRET_ACCESS_KEY=yourAWSSecretAccessKey

kosli snapshot lambda yourEnvironmentName 
    --exclude function1,function2 
	--exclude-regex "^not-wanted.*" 

```

**report what is running in the latest version of an AWS Lambda function (AWS auth provided in env variables)**

```shell
export AWS_REGION=yourAWSRegion
export AWS_ACCESS_KEY_ID=yourAWSAccessKeyID
export AWS_SECRET_ACCESS_KEY=yourAWSSecretAccessKey

kosli snapshot lambda yourEnvironmentName 
	--function-names yourFunctionName 

```

**report what is running in the latest version of AWS Lambda functions that match a name regex**

```shell
export AWS_REGION=yourAWSRegion
export AWS_ACCESS_KEY_ID=yourAWSAccessKeyID
export AWS_SECRET_ACCESS_KEY=yourAWSSecretAccessKey

kosli snapshot lambda yourEnvironmentName 
	--function-names-regex yourFunctionNameRegexPattern 

```

**report what is running in the latest version of multiple AWS Lambda functions (AWS auth provided in env variables)**

```shell
export AWS_REGION=yourAWSRegion
export AWS_ACCESS_KEY_ID=yourAWSAccessKeyID
export AWS_SECRET_ACCESS_KEY=yourAWSSecretAccessKey

kosli snapshot lambda yourEnvironmentName 
	--function-names yourFirstFunctionName,yourSecondFunctionName 

```

**report what is running in the latest version of an AWS Lambda function (AWS auth provided in flags)**

```shell
kosli snapshot lambda yourEnvironmentName 
	--function-names yourFunctionName 
	--aws-key-id yourAWSAccessKeyID 
	--aws-secret-key yourAWSSecretAccessKey 
	--aws-region yourAWSRegion 
```



## kosli_snapshot_path.md
---
title: "kosli snapshot path"
beta: false
deprecated: false
summary: "Report a snapshot of a single artifact running in a specific filesystem path to Kosli.  "
---

# kosli snapshot path

## Synopsis

Report a snapshot of a single artifact running in a specific filesystem path to Kosli.  
You can report a directory or file artifact. For reporting multiple artifacts in one go, use "kosli snapshot paths".
You can exclude certain paths or patterns from the artifact fingerprint using `--exclude`.
The supported glob pattern syntax is documented here: https://pkg.go.dev/path/filepath#Match ,
plus the ability to use recursive globs "**"

To specify paths in a directory artifact that should always be excluded from the SHA256 calculation, you can add a `.kosli_ignore` file to the root of the artifact.
Each line should specify a relative path or path glob to be ignored. You can include comments in this file, using `#`.
The `.kosli_ignore` will be treated as part of the artifact like any other file, unless it is explicitly ignored itself.

```shell
kosli snapshot path ENVIRONMENT-NAME [flags]
```

## Flags
| Flag | Description |
| :--- | :--- |
|    -D, --dry-run  |  [optional] Run in dry-run mode. When enabled, no data is sent to Kosli and the CLI exits with 0 exit code regardless of any errors.  |
|    -x, --exclude strings  |  [optional] The comma-separated list of literal paths or glob patterns to exclude when fingerprinting the artifact.  |
|    -h, --help  |  help for path  |
|        --name string  |  The reported name of the artifact.  |
|        --path string  |  The base path for the artifact to snapshot.  |
|        --watch  |  [optional] Watch the filesystem for changes and report snapshots of artifacts running in specific filesystem paths to Kosli.  |


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

These examples all assume that the flags  `--api-token`, `--org`, `--host`, (and `--flow`, `--trail` when required), are set/provided. 

**report one artifact running in a specific path in a filesystem**

```shell
kosli snapshot path yourEnvironmentName 
	--path path/to/your/artifact/dir/or/file 
	--name yourArtifactDisplayName 

```

**report one artifact running in a specific path in a filesystem AND exclude certain path patterns**

```shell
kosli snapshot path yourEnvironmentName 
	--path path/to/your/artifact/dir 
	--name yourArtifactDisplayName 
	--exclude **/log,unwanted.txt,path/**/output.txt
```



## kosli_snapshot_paths.md
---
title: "kosli snapshot paths"
beta: false
deprecated: false
summary: "Report a snapshot of artifacts running in specific filesystem paths to Kosli.  "
---

# kosli snapshot paths

## Synopsis

Report a snapshot of artifacts running in specific filesystem paths to Kosli.  
You can report directory or file artifacts in one or more filesystem paths. 
Artifacts names and the paths to include and exclude when fingerprinting them can be 
defined in a paths file which can be provided using `--paths-file`.

Paths files can be in YAML, JSON or TOML formats.
They specify a list of artifacts to fingerprint. For each artifact, the file specifies a base path to look for the artifact in 
and (optionally) a list of paths to exclude. Excluded paths are relative to the artifact path(s) and can be literal paths or
glob patterns.  
The supported glob pattern syntax is documented here: https://pkg.go.dev/path/filepath#Match ,
plus the ability to use recursive globs "**"

To specify paths in a directory artifact that should always be excluded from the SHA256 calculation, you can add a `.kosli_ignore` file to the root of the artifact.
Each line should specify a relative path or path glob to be ignored. You can include comments in this file, using `#`.
The `.kosli_ignore` will be treated as part of the artifact like any other file, unless it is explicitly ignored itself.

This is an example YAML paths spec file:
```yaml
version: 1
artifacts:
  artifact_name_a:
    path: dir1
    exclude: [subdir1, **/log]
```

```shell
kosli snapshot paths ENVIRONMENT-NAME [flags]
```

## Flags
| Flag | Description |
| :--- | :--- |
|    -D, --dry-run  |  [optional] Run in dry-run mode. When enabled, no data is sent to Kosli and the CLI exits with 0 exit code regardless of any errors.  |
|    -h, --help  |  help for paths  |
|        --paths-file string  |  The path to a paths file in YAML/JSON/TOML format. Cannot be used together with --path .  |
|        --watch  |  [optional] Watch the filesystem for changes and report snapshots of artifacts running in specific filesystem paths to Kosli.  |


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

These examples all assume that the flags  `--api-token`, `--org`, `--host`, (and `--flow`, `--trail` when required), are set/provided. 

**report one or more artifacts running in a filesystem using a path spec file**

```shell
kosli snapshot paths yourEnvironmentName 
	--paths-file path/to/your/paths/file 
```



## kosli_snapshot_s3.md
---
title: "kosli snapshot s3"
beta: false
deprecated: false
summary: "Report a snapshot of the content of an AWS S3 bucket to Kosli."
---

# kosli snapshot s3

## Synopsis

Report a snapshot of the content of an AWS S3 bucket to Kosli.

To authenticate to AWS, you can either:  
  1) provide the AWS static credentials via flags or by exporting the equivalent KOSLI env vars (e.g. KOSLI_AWS_KEY_ID)  
  2) export the AWS env vars (e.g. AWS_ACCESS_KEY_ID).  
  3) Use a shared config/credentials file under the $HOME/.aws  
  
Option 1 takes highest precedence, while option 3 is the lowest.  
More details can be found here: https://aws.github.io/aws-sdk-go-v2/docs/configuring-sdk/#specifying-credentials
	
You can report the entire bucket content, or filter some of the content using `--include` and `--exclude`.
In all cases, the content is reported as one artifact. If you wish to report separate files/dirs within the same bucket as separate artifacts, you need to run the command twice.

To specify paths in a directory artifact that should always be excluded from the SHA256 calculation, you can add a `.kosli_ignore` file to the root of the artifact.
Each line should specify a relative path or path glob to be ignored. You can include comments in this file, using `#`.
The `.kosli_ignore` will be treated as part of the artifact like any other file, unless it is explicitly ignored itself.

```shell
kosli snapshot s3 ENVIRONMENT-NAME [flags]
```

## Flags
| Flag | Description |
| :--- | :--- |
|        --aws-key-id string  |  The AWS access key ID.  |
|        --aws-region string  |  The AWS region.  |
|        --aws-secret-key string  |  The AWS secret access key.  |
|        --bucket string  |  The name of the S3 bucket.  |
|    -D, --dry-run  |  [optional] Run in dry-run mode. When enabled, no data is sent to Kosli and the CLI exits with 0 exit code regardless of any errors.  |
|    -x, --exclude strings  |  [optional] The comma separated list of file and/or directory paths in the S3 bucket to exclude when fingerprinting. Cannot be used together with --include.  |
|    -h, --help  |  help for s3  |
|    -i, --include strings  |  [optional] The comma separated list of file and/or directory paths in the S3 bucket to include when fingerprinting. Cannot be used together with --exclude.  |


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

These examples all assume that the flags  `--api-token`, `--org`, `--host`, (and `--flow`, `--trail` when required), are set/provided. 

**report the contents of an entire AWS S3 bucket (AWS auth provided in env variables)**

```shell
export AWS_REGION=yourAWSRegion
export AWS_ACCESS_KEY_ID=yourAWSAccessKeyID
export AWS_SECRET_ACCESS_KEY=yourAWSSecretAccessKey

kosli snapshot s3 yourEnvironmentName 
	--bucket yourBucketName 

```

**report what is running in an AWS S3 bucket (AWS auth provided in flags)**

```shell
kosli snapshot s3 yourEnvironmentName 
	--bucket yourBucketName 
	--aws-key-id yourAWSAccessKeyID 
	--aws-secret-key yourAWSSecretAccessKey 
	--aws-region yourAWSRegion 

```

**report a subset of contents of an AWS S3 bucket (AWS auth provided in env variables)**

```shell
export AWS_REGION=yourAWSRegion
export AWS_ACCESS_KEY_ID=yourAWSAccessKeyID
export AWS_SECRET_ACCESS_KEY=yourAWSSecretAccessKey

kosli snapshot s3 yourEnvironmentName 
	--bucket yourBucketName 
	--include file.txt,path/within/bucket 

```

**report contents of an entire AWS S3 bucket, except for some paths (AWS auth provided in env variables)**

```shell
export AWS_REGION=yourAWSRegion
export AWS_ACCESS_KEY_ID=yourAWSAccessKeyID
export AWS_SECRET_ACCESS_KEY=yourAWSSecretAccessKey

kosli snapshot s3 yourEnvironmentName 
	--bucket yourBucketName 
	--exclude file.txt,path/within/bucket 
```



## kosli_snapshot_server.md
---
title: "kosli snapshot server"
beta: false
deprecated: true
summary: "Report a snapshot of artifacts running in a server environment to Kosli.  "
---

# kosli snapshot server

{{% hint danger %}}
**kosli snapshot server** is deprecated. use 'kosli snapshot paths' instead  Deprecated commands will be removed in a future release.
{{% /hint %}}
## Synopsis

Report a snapshot of artifacts running in a server environment to Kosli.  
You can report directory or file artifacts in one or more server paths.

When fingerprinting a 'dir' artifact, you can exclude certain paths from fingerprint calculation 
using the `--exclude` flag.
Excluded paths are relative to the DIR-PATH and can be literal paths or glob patterns.
With a directory structure like this `foo/bar/zam/file.txt` if you are calculating the fingerprint of `foo/bar` you need to
exclude `zam/file.txt` which is relative to the DIR-PATH.
The supported glob pattern syntax is what is documented here: https://pkg.go.dev/path/filepath#Match , 
plus the ability to use recursive globs "**"

To specify paths in a directory artifact that should always be excluded from the SHA256 calculation, you can add a `.kosli_ignore` file to the root of the artifact.
Each line should specify a relative path or path glob to be ignored. You can include comments in this file, using `#`.
The `.kosli_ignore` will be treated as part of the artifact like any other file, unless it is explicitly ignored itself.

```shell
kosli snapshot server ENVIRONMENT-NAME [flags]
```

## Flags
| Flag | Description |
| :--- | :--- |
|    -D, --dry-run  |  [optional] Run in dry-run mode. When enabled, no data is sent to Kosli and the CLI exits with 0 exit code regardless of any errors.  |
|    -x, --exclude strings  |  [optional] The comma separated list of directories and files to exclude from fingerprinting. Can take glob patterns.  |
|    -h, --help  |  help for server  |
|    -p, --paths strings  |  The comma separated list of absolute or relative paths of artifact directories or files. Can take glob patterns, but be aware that each matching path will be reported as an artifact.  |


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

These examples all assume that the flags  `--api-token`, `--org`, `--host`, (and `--flow`, `--trail` when required), are set/provided. 

```shell
# report directory artifacts running in a server at a list of paths:
kosli snapshot server yourEnvironmentName \
	--paths a/b/c,e/f/g \
	--api-token yourAPIToken \
	--org yourOrgName  
	
# exclude certain paths when reporting directory artifacts: 
# in the example below, any path matching [a/b/c/logs, a/b/c/*/logs, a/b/c/*/*/logs]
# will be skipped when calculating the fingerprint
kosli snapshot server yourEnvironmentName \
	--paths a/b/c \
	--exclude logs,"*/logs","*/*/logs"
	--api-token yourAPIToken \
	--org yourOrgName 
	
# use glob pattern to match paths to report them as directory artifacts: 
# in the example below, any path matching "*/*/src" under top-dir/ will be reported as a separate artifact.
kosli snapshot server yourEnvironmentName \
	--paths "top-dir/*/*/src" \
	--api-token yourAPIToken \
	--org yourOrgName
```



## kosli_status.md
---
title: "kosli status"
beta: false
deprecated: false
summary: "Check the status of a Kosli server.  "
---

# kosli status

## Synopsis

Check the status of a Kosli server.  
The status is logged and the command always exits with 0 exit code.  
If you like to assert the Kosli server status, you can use the `--assert` flag or the "kosli assert status" command.

```shell
kosli status [flags]
```

## Flags
| Flag | Description |
| :--- | :--- |
|        --assert  |  [optional] Exit with non-zero code if Kosli server is not responding.  |
|    -h, --help  |  help for status  |


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




## kosli_tag.md
---
title: "kosli tag"
beta: false
deprecated: false
summary: "Tag a resource in Kosli with key-value pairs.  "
---

# kosli tag

## Synopsis

Tag a resource in Kosli with key-value pairs.  
use --set to add or update tags, and --unset to remove tags.


```shell
kosli tag RESOURCE-TYPE RESOURCE-ID [flags]
```

## Flags
| Flag | Description |
| :--- | :--- |
|    -D, --dry-run  |  [optional] Run in dry-run mode. When enabled, no data is sent to Kosli and the CLI exits with 0 exit code regardless of any errors.  |
|    -h, --help  |  help for tag  |
|        --set stringToString  |  [optional] The key-value pairs to tag the resource with. The format is: key=value  |
|        --unset strings  |  [optional] The list of tag keys to remove from the resource.  |


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

{{< tabs "live-examples" "col-no-wrap" >}}{{< tab "GitHub" >}}View an example of the `kosli tag` command in GitHub.

In [this YAML file](https://app.kosli.com/api/v2/livedocs/cyber-dojo/yaml?ci=github&command=kosli+tag), which created [this Kosli Event](https://app.kosli.com/api/v2/livedocs/cyber-dojo/event?ci=github&command=kosli+tag).{{< /tab >}}{{< tab "GitLab" >}}View an example of the `kosli tag` command in GitLab.

In [this YAML file](https://app.kosli.com/api/v2/livedocs/cyber-dojo/yaml?ci=gitlab&command=kosli+tag), which created [this Kosli Event](https://app.kosli.com/api/v2/livedocs/cyber-dojo/event?ci=gitlab&command=kosli+tag).{{< /tab >}}{{< /tabs >}}

## Examples Use Cases

These examples all assume that the flags  `--api-token`, `--org`, `--host`, (and `--flow`, `--trail` when required), are set/provided. 

**add/update tags to a flow**

```shell
kosli tag flow yourFlowName 
	--set key1=value1 
	--set key2=value2 

```

**tag an environment**

```shell
kosli tag env yourEnvironmentName 
	--set key1=value1 
	--set key2=value2 

```

**add/update tags to an environment**

```shell
kosli tag env yourEnvironmentName 
	--set key1=value1 
	--set key2=value2 

```

**remove tags from an environment**

```shell
kosli tag env yourEnvironmentName 
	--unset key1=value1 
```



## kosli_version.md
---
title: "kosli version"
beta: false
deprecated: false
summary: "Print the version of a Kosli CLI.  "
---

# kosli version

## Synopsis

Print the version of a Kosli CLI.  
The output will look something like this:  
version.BuildInfo{Version:"v0.0.1", GitCommit:"fe51cd1e31e6a202cba7dead9552a6d418ded79a", GitTreeState:"clean", GoVersion:"go1.16.3"}

- Version is the semantic version of the release.
- GitCommit is the SHA for the commit that this version was built from.
- GitTreeState is "clean" if there are no local code changes when this binary was
  built, and "dirty" if the binary was built from locally modified code.
- GoVersion is the version of Go that was used to compile Kosli CLI.


```shell
kosli version [flags]
```

## Flags
| Flag | Description |
| :--- | :--- |
|    -h, --help  |  help for version  |
|    -s, --short  |  [optional] Print only the Kosli CLI version number.  |


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




## _index.md
---
title: FAQ
bookCollapseSection: false
weight: 700
summary: "Frequently asked questions "
---

# Frequently asked questions

If you can't find the answer you're looking for please:

* email us at [support@kosli.com](mailto:support@kosli.com)
* join our slack community [here](https://join.slack.com/t/koslicommunity/shared_invite/zt-1dlchm3s7-DEP6TKjP3Mr58OZVB3hCBw)

## What do I do if Kosli is down?

There is a [tutorial](/tutorials/what_do_i_do_if_kosli_is_down/) dedicated to this.

## Why am I getting "Error response from daemon: client version 1.47 is too new. Maximum supported API version is 1.45" error in my GitHub Action Workflow?

The latest Kosli CLI defaults to using version 1.47 of the Docker API and
on Github Action Workflows, the maximum supported Docker API version is currently 1.45

You can tell the Kosli CLI to use version 1.45 by setting the
`DOCKER_API_VERSION` environment-variable. For example:

```yaml
env:
  DOCKER_API_VERSION: "1.45"
```


## Why am I getting "unknown flag" error?

If you see an error like below (or similar, with a different flag):
```
Error: unknown flag: --artifact-type
```
It most likely means you misspelled a flag.

## "unknown command" errors
E.g.
```
kosli expect deploymenct abc.exe --artifact-type file
Error: unknown command: deploymenct
available subcommands are: deployment
```

Note that there is a typo in deploymen**c**t.
This error will pop up if you're trying to use a command that is not present in the version of the kosli CLI you are using.

## zsh: no such user or named directory

When running commands with an argument starting with `~` you can encounter following problem:

```shell {.command}
kosli list snapshots prod ~3..NOW
```
```plaintext {.light-console}
zsh: no such user or named directory: 3..NOW
```

To help ZShell interpret the argument correctly, wrap it in quotation marks (single or double): 
```shell {.command}
kosli list snapshots prod '~3..NOW'
```
or
```shell {.command}
kosli list snapshots prod "~3..NOW"
```

## Github can't see KOSLI_API_TOKEN secret

Secrets in Github actions are not automatically exported as environment variables. You need to add required secrets to your GITHUB environment explicitly. E.g. to make kosli_api_token secret available for all cli commands as an environment variable use the following:

```yaml
env:
  KOSLI_API_TOKEN: ${{ secrets.kosli_api_token }}
```

## I'm running the Kosli CLI in a subshell and the captured output includes stderr!

The Kosli CLI writes debug information to `stderr`, and all other output to `stdout`.
Normally, in a bash $(subshell), only `stdout` is captured. 
In the following example, the `DIGEST` variable captures _only_ the 64 character digest of the docker image; 
the extra debug information is printed to the terminal.

```shell {.command}
# In a local terminal
KOSLI_DEBUG=true
DIGEST="$(kosli fingerprint "${IMAGE_NAME}" --artifact-type=docker)"

[debug] calculated fingerprint: 2c6079df58292ed10e8074adcb74be549b7f841a1bd8266f06bb5c518643193e for artifact: 244531986313.dkr.ecr.eu-central-1.amazonaws.com/exercises-start-points:86f9052

echo "DIGEST=${DIGEST}"
DIGEST=2c6079df58292ed10e8074adcb74be549b7f841a1bd8266f06bb5c518643193e
```

However, in many CI workflows (including Github and Gitlab), `stdout` and `stderr` are multiplexed together.
This means `DIGEST` will contain _both_ the 64 character digest _and_ the debug information. 
For example:

```shell {.command}
# In a CI workflow
KOSLI_DEBUG=true
DIGEST="$(kosli fingerprint "${IMAGE_NAME}" --artifact-type=docker)"

echo "DIGEST=${DIGEST}"
DIGEST=[debug] calculated fingerprint: 2c6079df58292ed10e8074adcb74be549b7f841a1bd8266f06bb5c518643193e for artifact: 244531986313.dkr.ecr.eu-central-1.amazonaws.com/exercises-start-points:86f9052
2c6079df58292ed10e8074adcb74be549b7f841a1bd8266f06bb5c518643193e
```

When running the Kosli CLI in a subshell, in a CI workflow, we recommend explicitly setting the `--debug` flag to false.

```shell {.command}
# In a CI workflow
KOSLI_DEBUG=true
DIGEST="$(kosli fingerprint "${IMAGE_NAME}" --artifact-type=docker --debug=false)"

echo "DIGEST=${DIGEST}"
DIGEST=2c6079df58292ed10e8074adcb74be549b7f841a1bd8266f06bb5c518643193e
```


## Where can I find API documentation?

Kosli API documentation is available for logged in Kosli users here: https://app.kosli.com/api/v2/doc/  
You can find the link at [app.kosli.com](https://app.kosli.com) after clicking at your avatar (top-right corner of the page)

<!-- 
### Do you support uploading a spdx or sbom as evidence?

We are working on providing that functionality in a near future. -->

## Do I have to provide all the flags all the time? 

A number of flags won't change their values often (or at all) between commands, like `--org` or `--api-token`.  Some will differ between e.g. workflows, like `--flow`. You can define them as environment variable to avoid unnecessary redundancy. Check [Environment variables](/getting_started/install/#assigning-flags-via-environment-variables) section to learn more.

## What is dry run and how to use it?

You can use dry run to disable writing to app.kosli.com - e.g. if you're just trying things out, or troubleshooting (dry run will print the payload the CLI would send in a non dry run mode). 

Here are three possible ways of enabling a dry run:
1. use the `--dry-run` flag (no value needed) to enable it per command
2. set the `KOSLI_DRY_RUN` environment variable to `true` to enable it globally (e.g. in your terminal or CI)
3. set the `KOSLI_API_TOKEN` environment variable to `DRY_RUN` to enable it globally (e.g. in your terminal or CI)

## What is the `--config-file` flag?

A config file is an alternative for using Kosli flags or Environment variables. Usually you'd use a config file for the values that rarely change - like api token or org, but you can represent all Kosli flags with config file. The key for each value is the same as the flag name, capitalized, so `--api-token` would become `API-TOKEN`, and `--org` would become `ORG`, etc. 

You can use JSON, YAML or TOML format for your config file. 

If you want to keep certain Kosli configuration in a file use `--config-file` flag when running Kosli commands to let the CLI know where to look for the file. The path given to `--config-file` flag should be a path relative to the location you're running kosli from. The file needs a valid format and extension, e.g.:

**kosli-conf.json:**
```
{
  "ORG": "my-org",
  "API-TOKEN": "123456abcdef"
}
```

**kosli-conf.yaml:**
```
ORG: "my-org"
API-TOKEN: "123456abcdef"
```

**kosli-conf.toml:**
```
ORG = "my-org"
API-TOKEN = "123456abcdef"
```

When calling Kosli command you can skip the file extension. For example, to list environments with `org` and `api-token` in the configuration file you would run:

```shell {.command}
kosli list environments --config-file kosli-conf
```

`--config-file` defaults to `kosli`, so if you name your file `kosli.<yaml|toml|json>` and the file is in the same location as where you run Kosli commands from, you can skip the `--config-file` altogether.


## Reporting the same artifact and evidence multiple times
If an artifact or evidence is reported multiple times there are a few corner cases. 
The issues are described here:

### Template
When an artifact is reported, the template for the flow is stored together with the artifact. 
If the template has changed between the times the same artifact is reported, it is the last 
template that is considered the template for that artifact.

### Evidence
If a given named evidence is reported multiple times it is the compliance status of the last 
reported version of the evidence that is considered the compliance state of that evidence.

If an artifact is reported multiple times with different git-commit, we can have the same named 
commit-evidence being attached to the artifact through multiple git-commits. It is the last
reported version of the named commit-evidence that is considered the compliance state of that evidence.

### Evidence outside the template
If an artifact has evidence, either commit evidence or artifact evidence, that is not 
part of the template, the state of the extra evidence will affect the overall compliance of the artifact.

## How to set compliant status of generic evidence

The `--compliant` flag is a [boolean flag](#boolean-flags). 
To report generic evidence as non-compliant use `--compliant=false`, as in this example:
```shell {.command}
kosli report evidence artifact generic server:1.0 \
  --artifact-type docker \
  --name test \
  --description "generic test evidence" \
  --compliant=false \
  --flow server
```

Keep in mind a number of flags, usually represented with environment variables, are omitted in this example.  
`--compliance` flag is set to `true` by default, so if you want to report generic evidence as compliant, simply skip providing the flag altogether.

## Boolean flags

Flags with values can usually be specified with an `=` or with a **space** as a separator.
For example, `--artifact-type=file` or `--artifact-type file`.
However, an explicitly specified boolean flag value **must** use an `=`.
For example, if you try this:
```
kosli attest generic Dockerfile --artifact-type file  --compliant true ...
```
You will get an error stating:
```
Error: accepts at most 1 arg(s), received 2
```
Here, `--artifact-type file` is parsed as if it was `--artifact-type=file`, leaving:
```
kosli attest generic Dockerfile --compliant true ...
```
Then `--compliant` is parsed as if *implicitly* defaulting to `--compliant=true`, leaving:
```
kosli attest generic Dockerfile true ...
```
The parser then sees `Dockerfile` and `true` as the two
arguments to `kosli attest generic`.


## _index.md
---
title: Getting started
bookCollapseSection: true
weight: 200
aliases:
    - /getting_started
---

## approvals.md
---
title: "Part 10: Approvals"
bookCollapseSection: false
weight: 300
summary: "When an artifact is ready to be deployed to a given environment, an approval may be reported to Kosli. An approval can be requested which will require a manual action, or reported automatically. This will be recorded in Kosli so the decision made outside your CI system won't be lost."
---
# Part 10: Approvals

When an artifact is ready to be deployed to a given [environment](/getting_started/environments/), an approval may be reported to Kosli. An approval can be requested which will require a manual action, or reported automatically. This will be recorded in Kosli so the decision made outside your CI system won't be lost.

When an approval is created for an artifact to a specific environment with the `--environment` flag, Kosli will generate a list of commits to be approved. By default, this list will contain all commits between `HEAD` and the commit of the most recent artifact coming from the same [flow](/getting_started/flows/) found in the given environment. The list can also be specified by providing values for `--newest-commit` and `--oldest-commit`. If you are providing these commits yourself, keep in mind that `--oldest-commit` has to be an ancestor of `--newest-commit`.

See [request approval](/client_reference/kosli_request_approval/) and [report approval](/client_reference/kosli_report_approval/) for usage details and examples. 


## artifacts.md
---
title: "Part 6: Artifacts"
bookCollapseSection: false
weight: 260
summary: "In software processes, you typically generate one or more artifacts that are deployed or distributed, such as docker images, archives, binaries, etc. You can ensure traceability for the creation of these artifacts by attesting them to Kosli, thereby establishing a binary provenance for each one."
---
# Part 6: Artifacts

In software processes, you typically generate one or more artifacts that are deployed or distributed, such as docker images, archives, binaries, etc. You can ensure traceability for the creation of these artifacts by attesting them to Kosli, thereby establishing a binary provenance for each one.

## Binary provenance

Binary provenance for artifacts refers to the ability to trace and verify the origins, history, and journey of the artifacts throughout their lifecycle. This involves recording immutable attestations about the artifact creation, risk controls performed on it, deployments, and execution/usage.

Artifacts are uniquely identified by their SHA256 fingerprints. When attesting an artifact to Kosli, you have the option to either provide the fingerprint manually or allow Kosli CLI to calculate it automatically for you.

By leveraging the artifact's fingerprint, Kosli can establish connections between the creation of the artifact and its runtime-related events, such as when the artifact starts or ceases execution within a specific environment.

By establishing and maintaining binary provenance for artifacts, Kosli enables you to:

1. **Track Changes**: Trace how your Flow artifacts change over time.
2. **Identify Sources**: Understand where your artifacts originated from, which can help in identifying vulnerabilities or issues.
3. **Monitor Compliance**: Ensure that the artifacts adhere to your compliance requirements.
4. **Enable Audits**: Access audit packages on demand allowing audits and investigations into the software supply chain.
5. **Enhance Trust**: Build trust among users, customers, and stakeholders by providing transparent and verified information about the software's history.

## Attesting artifacts

To attest an artifact, you can run a command similar to the one below:

```shell {.command}
kosli attest artifact project-a-app.bin \
	--artifact-type file \
	--build-url https://exampleci.com \
	--commit-url https://github.com/ProjectA/ProjectAApp/commit/e67f2f2b121f9325ebf166b7b3c707f73cb48b14 \
	--commit e67f2f2b121f9325ebf166b7b3c707f73cb48b14 \
	--flow project-a \
	--trail trail-1 \
	--name backend
```

The `--artifact-type` flag is used to determine the type of artifact being attested. The following types are supported:

- **file**: for any single file artifacts (e.g. a binary, Jar file, etc.)
- **dir**: for directory artifacts.
- **docker**: for docker images that are pulled on the machine. This option depends on having a running Docker daemon on the machine.
- **oci**: for container images in docker or OCI format. The fingerprint is fetched directly from the registry.

See [kosli attest artifact](/client_reference/kosli_attest_artifact/) for more details. 


## The --dry-run flag

All Kosli CLI commands which write data accept the `--dry-run` [boolean flag](/faq/#boolean-flags). 
When this flag is used, a CLI command:
* Does not communicate with Kosli at all
* Prints the payload it would have sent
* Exits with a zero status code

We recommend using the `KOSLI_DRY_RUN` environment variable to automatically set the `--dry-run` flag. 
This will allow you to instantly turn off all Kosli CLI commands if Kosli is down, as detailed in
[this tutorial](/tutorials/what_do_i_do_if_kosli_is_down/).

The `--dry-run` flag is also useful when trying commands locally. For example:

```shell {.command}
kosli attest artifact cyberdojo/differ:dde3b2a \
  --artifact-type=docker \
  --org=cyber-dojo \
  --flow=differ-ci \
  --trail=$(git rev-parse HEAD) \
  --dry-run \
  ...

 {
    "fingerprint": "0f53b5b9e7c266defe6984deafe039b116295b2df4a409ba6288c403f2451a9f",
    "filename": "cyberdojo/differ:dde3b2a",
    "git_commit": "fbb9e8000e2344323040e348a54b33ecbf67f273",
    "git_commit_info": {
        "sha1": "fbb9e8000e2344323040e348a54b33ecbf67f273",
        "message": "improve coverage report info (#2796)",
        "author": "Jon Jagger \u003cjon@jaggersoft.com\u003e",
        "timestamp": 1733724563,
        "branch": "master",
        "url": "https://github.com/kosli-dev/server/commit/fbb9e8000e2344323040e348a54b33ecbf67f273"
    },
    "build_url": "https://github.com/cyber-dojo/differ/actions/runs/11777650898",
    "commit_url": "https://github.com/cyber-dojo/differ/commit/dde3b2a7dab8e4567038e4c66ac68f0f01d0f704",
    "repo_url": "https://github.com/kosli-dev/server",
    "template_reference_name": "differ",
    "trail_name": "dde3b2a7dab8e4567038e4c66ac68f0f01d0f704"
}

$ echo $?
0
```



## attestations.md
---
title: "Part 7: Attestations"
bookCollapseSection: false
weight: 270
summary: "Attestations are how you record the facts you care about in your software supply chain. 
They are the evidence that you have performed certain activities, such as running tests, security scans, or ensuring that a certain requirement is met."
---
# Part 7: Attestations

Attestations are how you record the facts you care about in your software supply chain. 
They are the evidence that you have performed certain activities, such as running tests, security scans, or ensuring that a certain requirement is met.

Kosli allows you to report different types of attestations about artifacts and trails. 
Kosli will process the evidence you provide and conclude whether the evidence proves compliance or otherwise. 

Let's take a look at how to make attestations to Kosli.

The following compliance template is expecting 4 attestations, each with its own `name`.

```yml
version: 1
trail:
  attestations:
  - name: jira-ticket
    type: jira
  artifacts:
  - name: backend
    attestations:
    - name: unit-tests
      type: junit
    - name: security-scan
      type: snyk
```

It expects `jira-ticket` on the trail, the `backend` artifact, with `unit-tests` and `security-scan` attached to it. 
When you make an attestation, you have the choice of what `name` to attach it to.

## Make the `jira-ticket` attestation to a trail

The `jira-ticket` attestation belongs to a single trail and is not linked to a specific artifact. In this example, the id of the trail is the git commit.

```shell {.command}
kosli attest jira \
    --flow backend-ci \
	--trail $(git rev-parse HEAD) \	
    --name jira-ticket 
    ...
```

## Make the `unit-test` attestation to the `backend` artifact

Some attestations are attached to a specific artifact, like the unit tests for the `backend` artifact. Often, evidence like unit tests are created _before_ the artifact is built. To attach the evidence to the artifact before its creation, use `backend` (the artifact's `name` from the template), as well as `unit-tests` (the attestation's `name` from the template).

```shell {.command}
kosli attest junit \
    --name backend.unit-tests \
    --flow backend-ci \
    --trail $(git rev-parse HEAD) \
    ...
```

This attestation belongs to any artifact attested with the matching `name` from the template (in this example `backend`) and a matching git commit. 

## Make the `backend` artifact attestation

Once the artifact has been built, it can be attested with the following command.

```shell {.command}
kosli attest artifact my_company/backend:latest \
	--artifact-type docker \
    --flow backend-ci \
	--trail $(git rev-parse HEAD) \	
    --name backend 
    ...
```

In this case the Kosli CLI will calculate the fingerprint of the docker image called `my_company/backend:latest` and attest it as the `backend` artifact `name` in the trail.

{{% hint info %}}
### Automatically gather git commit and CI environment information
In all attestation commands the Kosli CLI will automatically gather the git commit and other information from the current git repository and the [CI environment](https://docs.kosli.com/integrations/ci_cd/).
This is how the git commit is used to match attestations to artifacts.
{{% /hint %}}

## Make the `security-scan` attestation to the `backend` artifact

Often, evidence like snyk reports are created _after_ the artifact is built. In this case, you can attach the evidence to the artifact after its creation. Use `backend` (the artifact's `name` from the template), as well as `security-scan` (the attestation's `name` from the template) to name the attestation.

The following attestation will only belong to the artifact `my_company/backend:latest` attested above and its fingerprint, in this case calculated by the Kosli CLI.

```shell {.command}
kosli attest snyk \
    --artifact-type docker my_company/backend:latest \
    --name backend.security-scan \
    --flow backend-ci \
    --trail $(git rev-parse HEAD)
    ...
```


## Compliance

### Attesting with a template

The four attestations above are all made against a Flow named `backend-ci` and a Trail named after the git commit.
Typically, the Flow and Trail are explicitly setup before making the attestations (e.g. at the start of a CI workflow).
This is done with the `create flow` and `begin trail` commands, either of which can specify the name of the template yaml file above 
(e.g. `.kosli.yml`) whose contents define overall compliance. For example:

```shell {.command}
kosli create flow backend-ci \
    --template-file .kosli.yml
    ...
    
kosli begin trail $(git rev-parse HEAD) \
    --flow backend-ci \
    ...    
```

An attested `backend` artifact is then compliant if and only if all the template attestations have been made
against it and are themselves compliant:
- `jira-ticket` on its Trail 
- `backend.unit-tests` for its junit evidence 
- `backend.security-scan` for its snyk evidence

If any of these attestations are missing, or are individually non-compliant then the `backend` artifact is non-compliant.

### Attesting without a template

An attestation can also be made against a Flow and Trail **not** previously explicitly setup.
In this case a Flow and Trail will be automatically setup but there will be no template yaml file defining
overall compliance. The compliance of any attested artifact will depend only on the compliance of the attestations actually made
and never because a specific attestation is missing.

### Attestation immutability

You can set/edit the template yml file for the Flow/Trail at any time.
This will affect compliance evaluations made after the edit.
It will not affect earlier records of compliance evaluations (e.g. in Environment Snapshots). 

Attestations are append-only immutable records. You can report the same attestation multiple times, and each report will be recorded.
However, only the latest version of the attestation is considered when evaluating compliance.


## Evidence Vault

Along with attestations data, you can attach additional supporting evidence files. These will be securely stored in Kosli's **Evidence Vault** and can easily be retrieved when needed. Alternatively, you can store the evidence files in your own preferred storage and only attach links to it in the Kosli attestation.

{{% hint info %}}
For `JUnit` attestations (see below), Kosli automatically stores the JUnit XML results files in the Evidence Vault. You can disable this by setting `--upload-results=false`
{{% /hint %}}

## Attestation types

Currently, we support the following types of evidence:

### Pull requests

If you use GitHub, Bitbucket, Gitlab or Azure DevOps you can use Kosli to verify if a given git commit comes from a pull/merge request. 

{{% hint warning %}}
Currently, the status of the PR does NOT impact the compliance status of the attestation.
{{% /hint %}}

If there is no pull request for the commit, the attestation will be reported as `non-compliant`. You can choose to short-circuit execution in case pull request is missing by using the `--assert` flag.

See the CLI reference for the following commands for more details and examples:

- [attest Github PR ](/client_reference/kosli_attest_pullrequest_github/) 
- [attest Bitbucket PR ](/client_reference/kosli_attest_pullrequest_bitbucket/)
- [attest Gitlab PR ](/client_reference/kosli_attest_pullrequest_gitlab/)
- [attest Azure Devops PR ](/client_reference/kosli_attest_pullrequest_azure/)


### JUnit test results

If you produce your test results in JUnit format, you can attest the test results to Kosli. Kosli will analyze the JUnit results and determine the compliance status based on whether any tests have failed and/or errored or not.

See [attest JUnit results to an artifact or a trail](/client_reference/kosli_attest_junit/) for usage details and examples.

### Snyk security scans 

You can report results of a Snyk security scan to Kosli and it will analyze the Snyk scan results and determine the compliance status based on whether vulnerabilities were found or not.

See [attest Snyk results to an artifact or a trail](/client_reference/kosli_attest_snyk/) for usage details and examples.

### Jira issues

You can use the Jira attestation to verify that a git commit or branch contains a reference to a Jira issue and that an issue with the same reference does exist in Jira.

If Jira reference is found in a commit message, that reference will be reported as evidence. If the reference is not found in the commit message, Kosli CLI will check if it's a part of a branch name.

Kosli CLI will also verify and report if the detected issue reference is found and accessible on Jira (reported as compliant) or not (reported as non compliant). 

See [attest Jira issue to an artifact or a trail](/client_reference/kosli_attest_jira/) for usage details and examples.

### SonarQube scan results

You can report the results of a SonarQube Server or SonarQube Cloud scan to Kosli. Kosli will use the status of the scan's Quality Gate (passing or failing) to determine the compliance status. 

These scan result can be attested in two ways:
- Using Kosli's [webhook integration](/integrations/sonar) with Sonar
- Using [Kosli's CLI](/client_reference/kosli_attest_sonar)

### Custom

The above attestations are all "fully typed" - each one knows how to interpret its own particular kind of input.
For example, `kosli attest snyk` interprets the sarif file produced by a snyk container scan to determine the `true/false` value. 
If you're using a tool that does not yet have a corresponding kosli attest command we recommend creating your own custom attestation type.

A custom attestation type specifies one or more arbitrary evaluation rules.
These rules can have an optional schema specifying the types of the names used in the rules, whether they are required, whether they have defaults, etc.
When a custom attestation is made using this type its rules are applied to the provided custom attestation data to determine its `true/false` compliance status.

For example, suppose you wish to attest coverage metrics captured as part of a unit-test run.
The coverage metrics are being saved in a file called `unit-test-coverage.json` as follows:
```json
{
  "code": {
    "lines": {
      "missed": 32,
      "total": 1209
    }
  },
  ...
}
```
You could create a custom attestation type called `coverage-metrics` using a [jq expression](https://jqlang.org/manual/) rule defining a minimum line coverage of 95%: 

```bash
kosli create attestation-type coverage-metrics
  --jq=".code.lines.missed / .code.lines.total * 100 <= 5"
```

You could then make your custom attestation with the json file:
```bash
kosli attest custom 
  --type=coverage-metrics
  --attestation-data=unit-test-coverage.json
  ...
```

For this attestation, Kosli would:
- Evaluate the rule `.code.lines.missed / .code.lines.total * 100 <= 5`
- Using the values from the file `unit-test-coverage.json`
  - `.code.lines.missed` is `32`
  - `.code.lines.total` is `1209`
- So `32 / 1209 * 100 <= 5` evaluates to `2.64 <= 5` which is `true`


See:
* [create custom attestation type](/client_reference/kosli_create_attestation-type) and
* [report custom attestation to an artifact or a trail](/client_reference/kosli_attest_custom/) for usage details and examples.

### Generic

{{% hint warning %}}
Generic attestations are an earlier, much less sophisticated version of custom attestations.
We recommend using custom attestations instead of generic attestations.
{{% /hint %}}

See [report generic attestation to an artifact or a trail](/client_reference/kosli_attest_generic/) for usage details and examples.



## environments.md
---
title: "Part 8: Environments"
bookCollapseSection: false
weight: 280
summary: "Kosli environments allow you to record the artifacts running in your runtime environments and how they change. Every time an environment change (or a set of changes) is reported, Kosli creates a new environment snapshot containing the status of the environment at a given point in time. The change record created in Kosli enables you to retrospectively perform runtime forensics about what ran where and when."
---
# Part 8: Environments

Kosli environments allow you to record the artifacts running in your runtime environments and how they change. Every time an environment change (or a set of changes) is reported, Kosli creates a new environment snapshot containing the status of the environment at a given point in time. The change record created in Kosli enables you to retrospectively perform runtime forensics about what ran where and when.

## Create an environment

You can create Kosli environments in the app, via CLI or via the API. When you create an environment, you give it a name, a description and select its type. 

{{% hint warning %}}
Make sure that type of Kosli environment matches the type of the environment you'll be reporting from.
{{% /hint %}}

### Via CLI

To create an environment via CLI, you would run a command like this:

```shell {.command}
kosli create environment quickstart \
    --environment-type docker \
    --description "quickstart environment for tutorial"
```

See [kosli create environment](/client_reference/kosli_create_environment/) for CLI usage details and examples.

### Via UI

You can also create an environment directly from [app.kosli.com](https://app.kosli.com).

- Make sure you've selected the organization you want to use from the orgs dropdown in the top left corner.
- Click on `Environments` in the left navigation menu.
- Click the `Add new environment` button
- Fill in the environment name and description and select a type, then click `Save Environment`.


After the new environment is created you'll be redirected to its page, which will initially have no snapshots. Once you start reporting your actual runtime environment to Kosli you'll be able to find snapshots and events (such as which artifacts started or stopped running) listed on that page.

## Snapshoting an environment

To record the current status of your environment you need to use the Kosli CLI to snapshot the running artifacts in it and report it to Kosli. 
When Kosli receives an environment report, if the received list of running artifacts is different than what is in the latest environment snapshot, a new snapshot is created. Snapshots are immutable and can't be tampered with.

Currently, the following environment types are supported:
- Kubernetes
- Docker
- Paths on a server
- AWS Simple Storage Service (S3)
- AWS Lambda
- AWS Elastic Container Service (ECS)
- Azure Web Apps and Function Apps

You can report environment snapshots manually using the `kosli snapshot [...]` commands for testing. For production use, however,  you would configure the reporting to happen automatically on regular intervals, e.g. via a cron job or scheduled CI job, or on certain events. 

You can follow one of the tutorials below to setup automatic snapshot reporting for your environment:
- [Kubernetes environment reporting](/tutorials/report_k8s_envs)
- [AWS ECS/S3/Lambda environment reporting](/tutorials/report_aws_envs)

### Snapshotting scopes

Depending on the type of your environment, you can scope what to snapshot from the environment. The following table shows the different scoping options currently available for different environment types:

| what to snapshot ->        | all resources | resources by names | resources by Regex | exclude by names | exclude by Regex |
|----------------------------|---------------|--------------------|--------------------|------------------|------------------|
| ECS (clusters)             |       √       |          √         |          √         |         √        |         √        |
| Lambda (functions)         |       √       |          √         |          √         |         √        |         √        |
| S3 (buckets)               |               |                    |                    |                  |                  |
| docker (containers)        |       √       |                    |                    |                  |                  |
| k8s (namespaces)           |       √       |          √         |          √         |         √        |         √        |
| azure (functions and apps) |       √       |                    |                    |                  |                  |


## Logical Environments

Logical environments are a way to group your Kosli environments so you can view all changes happening in your group in the same place. For example, if what you consider to be “Production” is a combination of a Kubernetes cluster, an S3 bucket, and a configuration file, you can combine the reports sent to these Kosli environments into a “Production” logical environment.

A logical environment can be created in the app or the CLI, and physical environments can be assigned to it in the app or with the [`kosli join environment`](/client_reference/kosli_join_environment/) command.


## flows.md
---
title: "Part 4: Flows"
bookCollapseSection: false
weight: 240
summary: "A Kosli Flow represents a business or software process that requires change tracking. It allows you to monitor changes across all steps within a process or focus specifically on a subset of critical steps."
---
# Part 4: Flows

A Kosli Flow represents a business or software process that requires change tracking. It allows you to monitor changes across all steps within a process or focus specifically on a subset of critical steps.

{{% hint info %}}
In all the commands below we skip the required `--api-token` and `--org` flags for brevity. These can be set as described [here](/getting_started/install#assigning-flags-via-config-files).
{{% /hint %}}

## Create a flow

To create a Flow, you can run:

```shell {.command}
kosli create flow process-a --description "My SW delivery process" \
    --use-empty-template
```

## Flow template

When creating a Flow, you can optionally provide a `Flow Template`. This template defines the necessary steps within the business or software process represented by a Kosli Flow. The compliance of Flow trails and artifacts will be assessed using the template.

A Flow template is a YAML file following the syntax outlined in the [flow template spec](/template_ref).

Here is an example, `sw-delivery-template.yml`:

```yml
version: 1
trail:
  attestations:
  - name: jira-ticket
    type: jira
  artifacts:
  - name: backend
    attestations:
    - name: unit-tests
      type: junit
```

### Create a Flow with a template

To create a Flow with a template, you can run:

```shell {.command}
kosli create flow process-a --description "My SW delivery process" \
 --template-file sw-delivery-template.yml
```

## Update a Flow

Rerunning the command with different description or template file will update the Flow. 

See [kosli create flow](/client_reference/kosli_create_flow/) for more details. 


## install.md
---
title: "Part 2: Install Kosli CLI"
bookCollapseSection: false
weight: 220
summary: "Kosli CLI can be installed from package managers, 
by Curling pre-built binaries, or can be used from the distributed Docker images."
---
# Part 2: Install Kosli CLI

Kosli CLI can be installed from package managers, 
by Curling pre-built binaries, or can be used from the distributed Docker images.
{{< tabs "installKosli" >}}

{{< tab "Homebrew" >}}
If you have [Homebrew](https://brew.sh/) (available on MacOS, Linux or Windows Subsystem for Linux), 
you can install the Kosli CLI by running: 

```shell {.command}
brew install kosli-cli
```
{{< /tab >}}

{{< tab "APT" >}}
On Ubuntu or Debian Linux, you can use APT to install the Kosli CLI by running:
```shell {.command}
sudo sh -c 'echo "deb [trusted=yes] https://apt.fury.io/kosli/ /"  > /etc/apt/sources.list.d/fury.list'
# On a clean debian container/machine, you need ca-certificates
sudo apt install ca-certificates
sudo apt update
sudo apt install kosli
```
{{< /tab >}}

{{< tab "YUM" >}}
On RedHat Linux, you can use YUM to install the Kosli CLI by running:
```shell {.command}
cat <<EOT >> /etc/yum.repos.d/kosli.repo
[kosli]
name=Kosli public Repo
baseurl=https://yum.fury.io/kosli/
enabled=1
gpgcheck=0
EOT
```
If you get mirrorlist errors (likely if you are on a clean centos container):

```shell {.command}
cd /etc/yum.repos.d/
sed -i 's/mirrorlist/#mirrorlist/g' /etc/yum.repos.d/CentOS-*
sed -i 's|#baseurl=http://mirror.centos.org|baseurl=http://vault.centos.org|g' /etc/yum.repos.d/CentOS-*
```

```shell {.command}
yum update -y
yum install kosli
```
{{< /tab >}}

{{< tab "Curl" >}}
You can download the Kosli CLI from [GitHub](https://github.com/kosli-dev/cli/releases).  
Make sure to choose the correct tar file for your system.  
For example, on Mac with AMD:
```shell {.command}
curl -L https://github.com/kosli-dev/cli/releases/download/v{{< cli-version >}}/kosli_{{< cli-version >}}_darwin_amd64.tar.gz | tar zx
sudo mv kosli /usr/local/bin/kosli
```
{{< /tab >}}

{{< tab "Docker" >}}
You can run the Kosli CLI with docker:
```shell {.command}
docker run --rm ghcr.io/kosli-dev/cli:v{{< cli-version >}}
```
The `entrypoint` for this container is the kosli command.

To run any kosli command you append it to the `docker run` command above –
without the `kosli` keyword. For example to run `kosli version`:
```shell {.command}
docker run --rm ghcr.io/kosli-dev/cli:v{{< cli-version >}} version
```
{{< /tab >}}

{{< tab "From source" >}}
You can build Kosli CLI from source by running:
```shell {.command}
git clone git@github.com:kosli-dev/cli.git
cd cli
make build
```
{{< /tab >}}

{{< /tabs >}}


## Verifying the installation worked

Run this command:
```shell {.command}
kosli version
```
The expected output should be similar to this:
```plaintext {.light-console}
version.BuildInfo{Version:"{{< cli-version >}}", GitCommit:"Homebrew", GitTreeState:"clean", GoVersion:"go1.23.4"}
```

## Using the CLI

The [CLI Reference](/client_reference/) section contains all the information you may need to run the Kosli CLI. The CLI flags offer flexibility for configuration and can be assigned in three distinct manners:

1. Directly on the command line.
2. Via environment variables.
3. Within a config file.
   
Among these options, priority is given in the following order: Option 1 holds the highest precedence, followed by Option 2, with Option 3 being the least prioritized.

### Assigning flags via environment variables

To assign a CLI flag using environment variables, generate a variable prefixed with KOSLI_. Use the flag's name in uppercase and substitute any internal dashes with underscores. For instance:


* `--api-token` corresponds to `KOSLI_API_TOKEN` 
* `--org` corresponds to `KOSLI_ORG`


### Assigning flags via config files

A config file is an alternative to using Kosli flags or environment variables. 
You could use a config file for the values that rarely change - like API token or org, 
but you can represent all Kosli flags in a config file. 

Each key in the config file corresponds to the flag name, capitalized. For instance:

* `--api-token` would become `API-TOKEN`.
* `--org` would become `ORG`.

Config files can be written in JSON, YAML, or TOML formats.

To direct Kosli CLI to use a config file, employ the --config-file flag when executing Kosli commands. By default, the CLI looks for a config file called `kosli.<yaml/yml/json/toml>`

Below are examples of different config file formats:


**kosli-conf.json:**
```
{
  "ORG": "my-org",
  "API-TOKEN": "123456abcdef"
}
```

**kosli-conf.yaml:**
```
ORG: "my-org"
API-TOKEN: "123456abcdef"
```

**kosli-conf.toml:**
```
ORG = "my-org"
API-TOKEN = "123456abcdef"
```

When using the `--config-file` flag you can skip the file extension. For example, 
to list environments with `org` and `api-token` in the configuration file you would run:

```shell {.command}
kosli list environments --config-file=kosli-conf
```


## next.md
---
title: "Part 11: Next Steps"
bookCollapseSection: false
weight: 310
summary: "In the previous chapters, you explored Kosli Flows and Environments and have reported some data to Kosli. 
The next steps would be to harness the benefits of your hard work. Here are a few areas to look at next:"
---
# Part 11: Next Steps

In the previous chapters, you explored Kosli Flows and Environments and have reported some data to Kosli. 
The next steps would be to harness the benefits of your hard work. Here are a few areas to look at next:

- [What do I do if Kosli is down?](/tutorials/what_do_i_do_if_kosli_is_down/)
- [Querying Kosli](/tutorials/querying_kosli/)
- [Setup Actions on Environment changes](/integrations/actions/)
- [Integrate Slack and Kosli](/integrations/slack/)


## overview.md
---
title: "Part 1: Overview"
bookCollapseSection: false
weight: 210
summary: "The \"Getting Started\" section contains the steps you can follow to implement Kosli in your organization. It focuses on general instructions for using Kosli and doesn't delve into specific tutorials for integrating Kosli with particular tools or runtime environments."
---
# Part 1: Overview

The "Getting Started" section contains the steps you can follow to implement Kosli in your organization. It focuses on general instructions for using Kosli and doesn't delve into specific tutorials for integrating Kosli with particular tools or runtime environments. For tailored tutorials catering to specific integrations, please refer to [tutorials](/tutorials).
  
{{% hint success %}}
If you're eager to start using Kosli right away, check our ["Get familiar with Kosli"](/tutorials/get_familiar_with_kosli/) tutorial that allows you to quickly try out Kosli features without the need to spin up a separate environment. No CI required.
{{% /hint %}}

The guide initially presents steps associated with **Flows** followed by **Environments**. However, if preferred, you can commence with Environments before exploring Flows. The guide allows flexibility in the order of exploration based on individual preferences.


## policies.md
---
title: "Part 9: Environment Policies"
bookCollapseSection: false
weight: 290
summary: "Environment Policies enable you to define and enforce compliance requirements for artifact deployments across different environments."
---
# Part 9: Environment Policies

{{% hint warning %}}
Environment policies is in alpha. It is subject to change, including naming, syntax, CLI commands, etc.
If you want to try this feature, create a policy and attach it to an environment. 
{{% /hint %}}
{{% hint warning %}}
Note that once an environment starts using policies, it is not possible to go back to not using them.
{{% /hint %}}

Environment Policies enable you to define and enforce compliance requirements for artifact deployments across different environments. With Environment Policies, you can:
- Define specific requirements for each environment (e.g, dev, staging, prod)
- Enforce consistent compliance standards across your deployment pipeline
- Prevent non-compliant artifacts from being deployed (via admission controllers)

Policies are written in YAML and are immutable (updating a policy creates a new version). They can be attached to one or more environments, and an environment can have one or more policies attached to it.

## Create a Policy

You can create a policy via CLI or via the API. Here is a basic policy that requires provenance and specific attestations:

```yaml {.command}
# prod-policy.yaml
_schema: https://kosli.com/schemas/policy/environment/v1
artifacts: # the rules apply to artifacts in an environment snapshot
  provenance:
    required: true # all artifacts must have provenance
  attestations:
    - name: dependency-scan # all artifacts must have dependency-scan attestation
      type: '*' # any attestation type
    - name: unit-test # all artifacts must have unit-test attestation
      type: junit # must be a 'junit' attestation type
```

You can create and manage policies using the Kosli CLI (global flags like org and api-token are omitted for brevity):

```shell {.command}
kosli create policy prod-requirements prod-policy.yaml
```

```shell {.command}
kosli create get policy prod-requirements
```

See [kosli create policy](/client_reference/kosli_create_policy/) for usage details and examples.

{{% hint info %}}
Once you create a policy, you will be able to see it in the UI under `policies` in the left navigation menu.
{{% /hint %}} 

## Declarative Policy Syntax

A Policy is declaratively defined according to the following schema:

```yaml {.command}
_schema: https://kosli.com/schemas/policy/environment/v1

artifacts:
  provenance:
    required: true | false (default = false)
    exceptions: (default [])
    - if: ${{ expression }}

  trail-compliance: 
    required: true | false (default = false)
    exceptions: (default [])
    - if: ${{ expression }}

  attestations: (default [])
    - if: ${{ expression }} (default = true)
      name: str (default = "*") # cannot have both name and type as *
      type: oneOf ['*', 'junit', 'jira', 'pull_request', 'snyk', 'sonar', 'generic', 'custom:<custom-type-name>'] (default = "*") # cannot have both name and type as *
```

### Policy Rules

A policy consists of `rules` which are applied to artifacts in an environment snapshot.

#### Provenance 

```yaml {.command}
artifacts:
  provenance:
    required: true  # Requires artifact to be part of a Kosli Flow
```

#### Trail Compliance 

```yaml {.command}
artifacts:
  trail-compliance:
    required: true  # Requires the trail in which the artifact is attested to be compliant
```

#### Specific Attestations

```yaml {.command}
artifacts:
  attestations:
    - name: '*' # attestation name can be anything
      type: pull-request
    - name: acceptance-test
      type: '*' # attestation type can be any built-in or existing custom type
    - name: security-scan
      type: snyk
    - name: coverage-metrics
      type: custom:my-coverage-metrics # custom attestation type
```

### Policy Rules Exceptions

You can add exceptions to policy rules using expressions.

```yaml
_schema: https://kosli.com/schemas/policy/environment/v1

artifacts
  provenance:
    required: true 
    exceptions:
    # provenance is required except when one of the expressions evaluates to true
    - if: ${{ expression1 }} 
    - if: ${{ expression2 }} 

  trail-compliance: 
    required: true 
    exceptions: 
    # trail-compliance is required except when one of the expressions evaluates to true
    - if: ${{ expression1 }} 
    - if: ${{ expression2 }} 

  attestations: 
    - if: ${{ expression }} # this attestation is only required when expression evaluates to true
      name: unit-tests
      type: junit
```

#### Policy Expressions

Policy expressions allow you to create conditional rules using a simple and powerful syntax. Expressions are wrapped in `${{ }}` and can be used in policy rules to create dynamic conditions. An expression consists of operands and operators:

**Operators**

Expressions support these operators:
- Comparison: `==, !=, <, >, <=, >=`
- Logical: `and, or, not`
- List membership: `in`

**Operands**

Operands can be:
- Literal string
- List 
- Context variable
- Function call


**Available Contexts**

Contexts are built-in objects which are accessible from an expression. Expressions can access two main contexts:
- `flow` - Information about the Kosli Flow:
    - `flow.name` - Name of the flow
    - `flow.tags` - Flow tags (accessed via flow.tags.tag_name)
- `artifact` - Information about the artifact:
    - `artifact.name` - Name of the artifact
    - `artifact.fingerprint` - SHA256 fingerprint

**Functions**

Functions are helpers that can be used when constructing conditions. They may or may not accept arguments. Arguments can be literals or context variables. Expressions can use following functions:

- `exists(arg)` : checks whether the value of arg is not None/Null
- `matches(input, regex)` : checks if input matches regex


**Example Expressions**

- ${{ exists(flow) }}
- ${{ flow.name in ["runner", 'saver', differ] }}
- ${{ matches(artifact.name, "^datadog:.*") }}
- ${{ flow.name == "runner" and matches(artifact.name, "^runner:.*") }}
- ${{ flow.tags.risk-level == "high" or matches(artifact.name, "^runner:.*") }}
- ${{ not flow.tags.risk-level == "high"}}
- ${{ flow.tags.risk-level != "high"}}
- ${{ flow.tags.key.with.dots == "value"}}
- ${{ flow.tags.risk-level >= 2 }}
- ${{ flow.name == 'prod' and (flow.tags.key_name == "value" or artifact.name == 'critical-service') }}
- ${{ flow.name == 'HIGH-RISK' and artifact.fingerprint == "37193ba1f3da2581e93ff1a9bba523241a7982a6c01dd311494b0aff6d349462" }}


## Attaching/Detaching Policies to/from Environments

Once you define your policies, you can attach them to environments via CLI or API:

```shell {.command}
kosli attach-policy prod-requirements --environment=aws-production
```

To detach a policy from an environment:

```shell {.command}
kosli detach-policy prod-requirements --environment=aws-production
```

Any attachment/detachment operation automatically triggers an evaluation of the latest environment snapshot and creates a new one with an updated compliance status.

{{% hint info %}}
If you detach all attached policies from an environment, the environment will have no defined requirements for artifacts running in it, and therefore, new environment snapshots will have status `unknown` 
{{% /hint %}} 


## Policy Enforcement Gates

Environment policies enable you to proactively block deploying a non-compliant artifact into an environment. This can be done as a deployment gate in your delivery pipeline or as an admission controller in your environment. 

Regardless of where you place your policy enforcement gate, it will be using the `assert artifact` Kosli CLI command or its equivalent API call.

```shell {.command}
kosli assert artifact --fingerprint=$SHA256 --environment=aws-production
```

## service-accounts.md
---
title: "Part 3: Service Accounts"
bookCollapseSection: false
weight: 230
summary: "Prior to engaging with Kosli, authentication is necessary. There are two methods to achieve this:

1. Using a service account API key (recommended).
2. Using a personal API key."
---
# Part 3: Create a Service Account

Prior to engaging with Kosli, authentication is necessary. There are two methods to achieve this:

1. Using a service account API key (recommended).
2. Using a personal API key.

## Service Accounts

{{% hint warning %}}
Service accounts are exclusively available within shared organizations.
{{% /hint %}}

A service account represents a machine user designed for interactions with Kosli from external systems, such as CI or runtime environments.

To create a service account:

- Log in to Kosli.
- From the left navigation menu, choose the organization where you wish to create the service account.
- Navigate to `Settings` in the left navigation menu.
- Select `Service accounts` from the settings sub-menu.
- Click `Add new service account`, provide a name for the service account, and click Add.
- Once created, generate an API key for the service account by clicking `Add API Key`.
- Choose a Time-To-Live (TTL) for the key, add a descriptive label, and then click `Add`.
- Ensure to copy the generated key as it won't be retrievable later. This key serves as the authentication token.


## Personal API Keys

{{% hint warning %}}
Personal API keys possess equivalent permissions to your user account, encompassing access to multiple organizations. Therefore, exercise caution while using personal API keys. These keys grant access and perform actions as per the associated user's permissions across various organizations.
{{% /hint %}}

To create a personal API key:
- Login to Kosli 
- From your user menu in the top right corner, click `Profile`
- In the API Keys section, click `Add API Key`, select a Time-To-Live (TTL) for the key, add a descriptive label, and then click `Add`
- Ensure to copy the generated key as it won't be retrievable later. This key serves as the authentication token.


## API Keys rotation

You can execute a zero-downtime API key rotation by following these steps:

- **Generate a New Key**: 
Create a new API key that will replace the existing key.

- **Replace the Old Key Where Used**: 
Implement the new key in all areas where the old key is currently used for authentication or access.

- **Delete the Old Key:**
Once the new key is in place and operational, remove or delete the old key from the system or applications where it was previously employed for security or authentication purposes.

By systematically following these steps, you can ensure a seamless API key rotation without causing any downtime or interruptions in service.


## Using API Keys

### In CLI

you can assign an API key to any CLI command by one of the following options:
- using the `--api-token` flag
- exporting an environment variable called `KOSLI_API_TOKEN`
- setting it in a config file and passing the config file using `--config-file` (see [here](/getting_started/install#assigning-flags-via-config-files))

### In API

When making requests against the Kosli API directly (e.g. using curl), you can authenticate your requests by setting the bearer token in the request Authorization header to your API key.

```shell
curl -H "Authorization: Bearer <<your-api-key>>" http://app.kosli.com/api/v2/environments/<<your-org-name>>
```

## trails.md
---
title: "Part 5: Trails"
bookCollapseSection: false
weight: 250
summary: "Every time you execute a process represented by a Kosli Flow, you would initiate a `trail` to record the changes made during that specific execution."
---
# Part 5: Trails

Every time you execute a process represented by a Kosli Flow, you would initiate a `trail` to record the changes made during that specific execution.

You have the flexibility to determine the boundaries of what you consider a single execution of your process. For instance, in a software delivery process, an execution instance might be defined by: 

- **Git commits**: the trail represents changes recorded from a single commit (as reported from CI).
- **Pull requests**: the trail represents changes recorded throughout the life of a single pull request (can span multiple commits).
- **Jira or Github issues**: the trail represents changes recorded throughout the life of a single ticket/issue (can span multiple pull requests and commits).

Each trail must possess a unique name within the Flow. This name typically follows a custom pattern, depending on how you define the scope of a single process execution.

## Begin a trail 

To begin a Trail, you can run a command similar to the one below:

```shell {.command}
kosli begin trail trail-1 --flow process-1 --description "My first trail"
```

Rerunning the command with different description or template file will update the Trail. 

See [kosli begin trail](/client_reference/kosli_begin_trail/) for more details. 

{{% hint info %}}
You can overwrite the flow template for each trail using `--template-file`.
By default, the trail inherits the template from its Flow.
{{% /hint %}}


## _index.md
---
title: Kubernetes Reporter Helm Chart
summary: "A Helm chart for installing the Kosli K8S reporter as a cronjob.
The chart allows you to create a Kubernetes cronjob and all its necessary RBAC to report running images to Kosli at a given cron schedule."
---

# k8s-reporter

![Version: 1.6.0](https://img.shields.io/badge/Version-1.6.0-informational?style=flat-square)

A Helm chart for installing the Kosli K8S reporter as a cronjob.
The chart allows you to create a Kubernetes cronjob and all its necessary RBAC to report running images to Kosli at a given cron schedule.

## Prerequisites

- A Kubernetes cluster (minimum supported version is `v1.21`)
- Helm v3.0+
- If you want to report artifacts from just one namespace, you need to have permissions to `get` and `list` pods in that namespace.
- If you want to report artifacts from multiple namespaces or entire cluster, you need to have cluster-wide permissions to `get` and `list` pods.

## Installing the chart

To install this chart via the Helm chart repository:

1. Add the Kosli helm repo
```shell {.command}
helm repo add kosli https://charts.kosli.com/ && helm repo update
```

2. Create a secret for the Kosli API token
```shell {.command}
kubectl create secret generic kosli-api-token --from-literal=key=<your-api-key>
```

3. Install the helm chart

A. To report artifacts running in entire cluster (requires cluster-wide read permissions):

```shell {.command}
helm install kosli-reporter kosli/k8s-reporter \
    --set reporterConfig.kosliOrg=<your-org> \
    --set reporterConfig.kosliEnvironmentName=<your-env-name>
```

B. To report artifacts running in multiple namespaces (requires cluster-wide read permissions):

```shell {.command}
helm install kosli-reporter kosli/k8s-reporter \
    --set reporterConfig.kosliOrg=<your-org> \
    --set reporterConfig.kosliEnvironmentName=<your-env-name> \
    --set reporterConfig.namespaces=<namespace1,namespace2>
```

C. To report artifacts running in one namespace (requires namespace-scoped read permissions):

```shell {.command}
helm install kosli-reporter kosli/k8s-reporter \
    --set reporterConfig.kosliOrg=<your-org> \
    --set reporterConfig.kosliEnvironmentName=<your-env-name> \
    --set reporterConfig.namespaces=<namespace1> \
    --set serviceAccount.permissionScope=namespace
```

> Chart source can be found at https://github.com/kosli-dev/cli/tree/main/charts/k8s-reporter

> See all available [configuration options](#configurations) below.

## Upgrading the chart

```shell {.command}
helm upgrade kosli-reporter kosli/k8s-reporter ...
```

## Uninstalling chart

```shell {.command}
helm uninstall kosli-reporter
```

## Configurations
| Key | Type | Default | Description |
|-----|------|---------|-------------|
| cronSchedule | string | `"*/5 * * * *"` | the cron schedule at which the reporter is triggered to report to Kosli   |
| fullnameOverride | string | `""` | overrides the fullname used for the created k8s resources. It has higher precedence than `nameOverride` |
| image.pullPolicy | string | `"IfNotPresent"` | the kosli reporter image pull policy |
| image.repository | string | `"ghcr.io/kosli-dev/cli"` | the kosli reporter image repository |
| image.tag | string | `"v2.11.3"` | the kosli reporter image tag, overrides the image tag whose default is the chart appVersion. |
| kosliApiToken.secretKey | string | `"key"` | the name of the key in the secret data which contains the Kosli API token |
| kosliApiToken.secretName | string | `"kosli-api-token"` | the name of the secret containing the kosli API token |
| nameOverride | string | `""` | overrides the name used for the created k8s resources. If `fullnameOverride` is provided, it has higher precedence than this one |
| podAnnotations | object | `{}` |  |
| reporterConfig.dryRun | bool | `false` | whether the dry run mode is enabled or not. In dry run mode, the reporter logs the reports to stdout and does not send them to kosli. |
| reporterConfig.httpProxy | string | `""` | the http proxy url |
| reporterConfig.kosliEnvironmentName | string | `""` | the name of Kosli environment that the k8s cluster/namespace correlates to |
| reporterConfig.kosliOrg | string | `""` | the name of the Kosli org |
| reporterConfig.namespaces | string | `""` | the namespaces which represent the environment. It is a comma separated list of namespace names. leave this unset if you want to report what is running in the entire cluster |
| resources.limits.cpu | string | `"100m"` | the cpu limit |
| resources.limits.memory | string | `"256Mi"` | the memory limit |
| resources.requests.memory | string | `"64Mi"` | the memory request |
| serviceAccount.annotations | object | `{}` | annotations to add to the service account |
| serviceAccount.create | bool | `true` | specifies whether a service account should be created |
| serviceAccount.name | string | `""` | the name of the service account to use. If not set and create is true, a name is generated using the fullname template |
| serviceAccount.permissionScope | string | `"cluster"` | specifies whether to create a cluster-wide permissions for the service account or namespace-scoped permissions. allowed values are: [cluster, namespace] |

----------------------------------------------
Autogenerated from chart metadata using [helm-docs v1.5.0](https://github.com/norwoodj/helm-docs/releases/v1.5.0)



## _index.md
---
title: Kosli integrations
bookCollapseSection: true
weight: 300
---

## actions.md
---
title: Actions
bookCollapseSection: false
weight: 320
summary: "Actions enable you to automate the execution of if-this-do-that workflows based on Kosli events. You can configure actions to either receive a Slack notification or a JSON payload on a custom webhook when certain Kosli events happen.  "
---

# Kosli Actions

Actions enable you to automate the execution of if-this-do-that workflows based on Kosli events. You can configure actions to either receive a Slack notification or a JSON payload on a custom webhook when certain Kosli events happen.   

You can configure actions to be triggered by one or more of the following events occurring in one or more environments:

- When a new artifact starts execution in an environment.
- When an artifact ceases execution in an environment.
- When instances of an artifact are scaled up or down.
- When an artifact is added to the allow-list in an environment.
- When an environment transitions from a **Compliant** state to a **Non-Compliant** state.
- When an environment changes from a **Non-Compliant** state to a **Compliant** state.


## Slack Notifications

To receive Kosli notifications in Slack, you have two options:

1) Using Kosli Slack App (recommended)

Subscribe to Kosli notifications using the [Kosli Slack App](/integrations/slack/). This method is recommended for a seamless integration.
Use the app to create notification settings by running the `/kosli subscribe` slash command.

2) Using Slack Incoming Webhooks

- Create a [Slack incoming webhook](https://api.slack.com/messaging/webhooks#create_a_webhook).
- Use this webhook to [create a notification settings in the Kosli UI](/integrations/actions/#manage-actions-in-the-ui).
  
Both approaches allow you to configure Kosli notifications in Slack, offering flexibility based on your preferences.

## Custom Webhook Notifications

Custom webhook notifications empower you to implement automation workflows for "if-this-then-that" scenarios. Whenever an event that matches your specified notification settings occurs, a JSON payload, as outlined below, is transmitted to your designated custom webhook:

```json
{
    "version": "1.0",
    "timestamp": "1692616493",
    "org": "cyber-dojo",
    "environment": "aws-prod",
    "event_type": "ARTIFACT_STARTED",
    "description": "1 instance started running (from 0 to 1)",
    "snapshot":  {
           "index": "1035",
           "status": "compliant",
           "html_url": "https://app.kosli.com/cyber-dojo/environments/aws-prod/snapshots/1035",
           "api_url": "https://app.kosli.com/api/v2/snapshots/cyber-dojo/aws-prod/1035"
     },
    "artifact": {
        "name": "runner",
        "fingerprint": "719defb995c86ad7c406ad74258fe98b9ebd71dfa80cd786870c967cb6c1f08d",
        "provenance": {
            "flow": "runner",
            "status": "compliant",
            "commit": "1ac157003dd6fb9ec764daa47726b7bfed65c312",
            "commit_url": "https://github.com/cyber-dojo/runner/commit/1ac157003dd6fb9ec764daa47726b7bfed65c312",
            "html_url": "https://app.kosli.com/cyber-dojo/runner/719defb995c86ad7c406ad74258fe98b9ebd71dfa80cd786870c967cb6c1f08d",
            "api_url": "https://app.kosli.com/api/v2/artifacts/cyber-dojo/runner/fingerprint/719defb995c86ad7c406ad74258fe98b9ebd71dfa80cd786870c967cb6c1f08d",
            "build_url": "https://github.com/cyber-dojo/runner/actions/runs/5891969166",
            "deployments": [ 
                {
                   "number": "44",
                   "timestamp": "1692618644",
                   "build_url": "https://github.com/cyber-dojo/runner/actions/runs/5891969166",
                   "html_url": "https://app.kosli.com/cyber-dojo/flows/runner/deployments/44",
                   "api_url": "https://app.kosli.com/api/v2/deployments/cyber-dojo/runner/44"
               }
            ],
            "approvals": [
                {
                   "number": "42",
                   "timestamp": "1692617329",
                   "state": "approved",
                   "latest_reviewer": "username",
                   "latest_review_comment": "lgtm",
                    "html_url": "https://app.kosli.com/cyber-dojo/flows/runner/approvals/42",
                    "api_url": "https://app.kosli.com/api/v2/approvals/cyber-dojo/runner/42"
               }
            ]
        }
    }
}
```

# Manage Actions in the UI

You can manage Actions for your organization in the Kosli UI from the `Actions` section in the left navigation menu. The Actions sections enables you to:
- Create Notifications: Create a new notifications settings.
- Delete Notifications: Remove existing notification settings that are no longer needed.
- Update Notifications: Modify notification settings as needed.



## ci_cd.md
---
title: CI/CD
bookCollapseSection: false
weight: 310
summary: "This section provides how-to guides showing you how to use Kosli to report changes from
different CI systems."
aliases:
    - /ci-defaults  # To keep short URL in docs and help in the CLI
---
# Use Kosli in CI Systems

This section provides how-to guides showing you how to use Kosli to report changes from
different CI systems.

{{% hint info %}}
Note that **all** CLI command flags can be set as environment variables by adding the the `KOSLI_` prefix and capitalizing them. 
{{% /hint %}}

## Defaulted Kosli command flags from CI variables

The following flags are **defaulted** (which means you don't need to provide the flags, they'll be automatically set to values listed below) as follows in the CI systems below:

{{< tabs "ci-defaults" "col-no-wrap" >}}

{{< tab "Azure DevOps" >}}
| Flag | Default |
| :--- | :--- |
| --build-url | ${SYSTEM_COLLECTIONURI}/${SYSTEM_TEAMPROJECT}/_build/results?buildId=${BUILD_BUILDID} |
| --commit-url | ${SYSTEM_COLLECTIONURI}/${SYSTEM_TEAMPROJECT}/_git/${BUILD_REPOSITORY_NAME}/commit/${BUILD_SOURCEVERSION} |
| --commit | ${BUILD_SOURCEVERSION} |
| --git-commit | ${BUILD_SOURCEVERSION} |
| --repository | ${BUILD_REPOSITORY_NAME} |
| --project | ${SYSTEM_TEAMPROJECT} |
| --azure-org-url | ${SYSTEM_COLLECTIONURI} |
{{< /tab >}}

{{< tab "Bitbucket Cloud" >}}
| Flag | Default |
| :--- | :--- |
| --build-url | https://bitbucket&#46;org/${BITBUCKET_WORKSPACE}/${BITBUCKET_REPO_SLUG}/addon/pipelines/home#!/results/${BITBUCKET_BUILD_NUMBER} |
| --commit-url | https://bitbucket&#46;org/${BITBUCKET_WORKSPACE}/${BITBUCKET_REPO_SLUG}/commits/${BITBUCKET_COMMIT} |
| --commit | ${BITBUCKET_COMMIT} |
| --git-commit | ${BITBUCKET_COMMIT} |
| --repository | ${BITBUCKET_REPO_SLUG} |
| --bitbucket-workspace |  ${BITBUCKET_WORKSPACE} |
{{< /tab >}}

{{< tab "CodeBuild" >}}
| Flag | Default |
| :--- | :--- |
| --build-url | ${CODEBUILD_BUILD_URL} |
| --commit-url | ${CODEBUILD_SOURCE_REPO_URL}/commit(s)/${CODEBUILD_RESOLVED_SOURCE_VERSION} |
| --commit | ${CODEBUILD_RESOLVED_SOURCE_VERSION} |
| --git-commit | ${CODEBUILD_RESOLVED_SOURCE_VERSION} |
{{< /tab >}}

{{< tab "Github" >}}
| Flag | Default |
| :--- | :--- |
| --build-url | ${GITHUB_SERVER_URL}/${GITHUB_REPOSITORY}/actions/runs/${GITHUB_RUN_ID} |
| --commit-url | ${GITHUB_SERVER_URL}/${GITHUB_REPOSITORY}/commit/${GITHUB_SHA} |
| --commit | ${GITHUB_SHA} |
| --git-commit | ${GITHUB_SHA} |
| --repository | ${GITHUB_REPOSITORY} |
| --github-org | ${GITHUB_REPOSITORY_OWNER} |
{{< /tab >}}

{{< tab "Gitlab" >}}
| Flag | Default |
| :--- | :--- |
| --build-url | ${CI_JOB_URL} |
| --commit-url | ${CI_PROJECT_URL}/-/commit/${CI_COMMIT_SHA} |
| --commit | ${CI_COMMIT_SHA} |
| --git-commit | ${CI_COMMIT_SHA} |
| --repository | ${CI_PROJECT_NAME} |
| --gitlab-org | ${CI_PROJECT_NAMESPACE} |
{{< /tab >}}

{{< tab "CircleCI" >}}
| Flag | Default |
| :--- | :--- |
| --build-url | ${CIRCLE_BUILD_URL} |
| --commit-url | ${CIRCLE_REPOSITORY_URL}(converted to https url)/commit(s)/${CIRCLE_SHA1} |
| --git-commit | ${CIRCLE_SHA1} |
{{< /tab >}}

{{< tab "Teamcity" >}}
| Flag | Default |
| :--- | :--- |
| --git-commit | ${BUILD_VCS_NUMBER} |
{{< /tab >}}

{{< /tabs >}}


## Use Kosli in Github Actions

To use Kosli in [Github Actions](https://docs.github.com/en/actions) workflows, you can use the kosli [CLI setup action](https://github.com/marketplace/actions/setup-kosli-cli) to install the CLI on your Github Actions Runner.
Then, you can use all the [CLI commands](/client_reference) in your workflows.

### GitHub Secrets 

Keep in mind that secrets in Github actions are not automatically exported as environment variables. You need to add required secrets to your GITHUB environment explicitly. E.g. to make kosli_api_token secret available for all cli commands as an environment variable use following:

```yaml
env:
  KOSLI_API_TOKEN: ${{ secrets.kosli_api_token }}
```

### Example

Here is an example Github Actions workflow snippet using `kosli-dev/setup-cli-action` running `kosli create flow` command:

```yaml
jobs:
  example:
    runs-on: ubuntu-latest
    env:
      KOSLI_API_TOKEN: ${{ secrets.MY_KOSLI_API_TOKEN }}
      KOSLI_ORG: my-org
    steps:
      - name: setup kosli
        uses: kosli-dev/setup-cli-action@v2
      - name: create flow
        run: kosli create flow my-flow --template pull-request,artifact,test
```

For a complete example of a Github workflow using Kosli, please check the Kosli CLI's [own workflow](https://github.com/kosli-dev/cli/blob/main/.github/workflows/docker.yml). 


## Use Kosli in Gitlab pipelines

For a complete example of a Gitlab pipeline using Kosli, please check [this cyber-dojo pipeline](https://gitlab.com/cyber-dojo/creator/-/blob/main/.gitlab/workflows/main.yml). 


## launchdarkly.md
---
title: LaunchDarkly
bookCollapseSection: false
weight: 340
summary: "LaunchDarkly feature flag changes can be tracked in Kosli trails."
---
# LaunchDarkly in Kosli

LaunchDarkly feature flag changes can be tracked in [Kosli trails](/getting_started/trails/).

## Setting up in Kosli

To set up the integration, navigate to the LaunchDarkly integration page of your org in the [Kosli app](https://app.kosli.com/).

![Kosli App LaunchDarkly Integration page](/images/launchdarkly-integration.png)

After switching on the integration, you will be provided with a webhook and a secret.

## Setting up in LaunchDarkly

You're now just a few steps away from connecting LaunchDarkly to Kosli.
In [LaunchDarkly](https://app.launchdarkly.com/):
- Navigate to the "Integrations" tab
- Create a new webhook integration
- Enter the webhook url and secret in the relevant fields
- Add policy statements for flags and environments for which you'd like to send information Kosli. By leaving these policy statements blank, all flag changes in all environments will report back to Kosli.
- Save the settings

## Testing the integration

To make sure the integration is configured properly, switch a feature flag on or off.
The first time a flag is changed in a LaunchDarkly environment, a [Flow](/getting_started/flows/) will be created in Kosli titled `launch-darkly-<your_environment_name>`, and inside this flow a trail will be created named after the name of your feature flag.
All changes to this flag will be found in the trail.
Subsequently, any change to a feature flag in this environment will be tracked in the appropriate trail.


## slack.md
---
title: Slack integration
bookCollapseSection: false
weight: 330
summary: "Kosli Slack App allows you to configure and receive notifications about changes in your environments and query Kosli about your environments and artifacts without leaving Slack window."
---
# Kosli Slack App
[Kosli Slack App](#kosli-slack-app) allows you to configure and receive notifications about changes in your environments
and query Kosli about your environments and artifacts without leaving Slack window.

## Installation

Visit https://slack.kosli.com to add Kosli Slack App to your Slack workspace.
## Usage

Now that Kosli Slack App is installed you can start using all `/kosli` commands in any channel.

At any time you can run `/kosli help` to see which commands are available.

The next step is connecting your Slack user with your Kosli user, use the command below to do that:
```
/kosli login
```

After that you may want to set up default Kosli organization, so you don't have to provide it every time you want to run `/kosli` commands from slack.  
E.g. if the organization name is **my-org**: 
```
/kosli config org my-org
```

In case of commands referring to snapshots you can specify snapshot(s) you're interested in multiple ways:
- environmentName~N *N'th behind the latest snapshot*
- environmentName#N *snapshot number N*
- environmentName@{YYYY-MM-DDTHH:MM:SS} *snapshot at specific moment in time in UTC*
- environmentName *the latest snapshot*

### Example

Here is an example of *search* command and the response:  

`/kosli search edb1a262`
{{<figure src="/images/slack-kosli-search.png" alt="Kosli search slack message" width="700">}}





## sonar.md
---
title: Sonar
bookCollapseSection: false
weight: 340
summary: "The results of SonarQube Server and SonarQube Cloud scans can be tracked in Kosli trails. This integration involves setting up a Sonar webhook in Kosli and a corresponding webhook in SonarQube. When you run a scan of your SonarQube project, the webhook is triggered and the results of the scan are sent to Kosli."
---
# Record Sonar scan results in Kosli

The results of SonarQube Server and SonarQube Cloud scans can be tracked in [Kosli trails](/getting_started/trails/).  
This integration involves setting up a Sonar webhook in Kosli and a corresponding webhook in SonarQube. When you run a scan of your SonarQube project, the webhook is triggered and the results of the scan are sent to Kosli.  
Some parameters must be passed to the Sonar scanner when it is run (e.g. the name of the Flow corresponding to the project, and the name of the trail the results should be attested to); these are sent with the scan results, and allow Kosli to determine the compliance status of the results and attest them to the correct trail/artifact.

## Setting up in Kosli

To set up the integration, navigate to the Sonar integration page for your org in the [Kosli app](https://app.kosli.com/).

After switching on the integration, you will be provided with a webhook and a secret.

## Setting up Sonar Webhooks

You're now just a few steps away from connecting SonarQube to Kosli.

Both SonarQube Server and SonarQube Cloud provide two types of webhooks: global (which are triggered when any project in your organization is scanned) and project-specific (which are triggered by a scan for that project only). Kosli supports both types of webhooks.

In [SonarQube Cloud](https://sonarcloud.io/) or [SonarQube Server](https://sonarqube.org):

### To create a global webhook:

- In SonarQube Cloud: Go to your Organization, then Administration > Webhooks
- In SonarQube Server: Go to Administration > Configuration > Webhooks
- Create a new Webhook
- Add the Kosli webhook URL and secret provided
- Click Create

![SonarQube Cloud Global Webhook page](/images/sonarqube-cloud-integration-global.png)
![SonarQube Server Global Webhook page](/images/sonarqube-integration-global.png)

### To create a project-specific webhook:

- Go to the project you want to create a webhook for
- Click on Administration (SonarQube Cloud) or Project Settings (SonarQube Server) and go to Webhooks in the dropdown menu
- Create a new Webhook
- Add the Kosli webhook URL and secret provided
- Click Create

![SonarQube Cloud Project Webhook page](/images/sonarqube-cloud-integration-project.png)
![SonarQube Server Project Webhook page](/images/sonarqube-integration-project.png)

## Setting up the SonarScanner

In order for Kosli to know where the scan results should be attested, certain parameters can be passed to the SonarScanner. Note that parameters cannot be passed with SonarQube Cloud's Automatic Analysis - in this case, Kosli determines the relevant Flow and Trail as described below.

These parameters can be passed to the scanner in three ways:
- As part of the sonar-project.properties file used in CI analysis
- As arguments to the scanner in your CI pipeline's YML file
```shell
    - name: SonarQube Scan
        uses: SonarSource/sonarqube-scan-action@master
        with:
          args: >
            -Dsonar.analysis.kosli_flow=<YourFlowName>
            -Dsonar.analysis.kosli_trail=<YourTrailName>
```
- As arguments to the CLI scanner
```shell
$ sonar-scanner \
  -Dsonar.analysis.kosli_flow=<YourFlowName> \
  -Dsonar.analysis.kosli_trail=<YourTrailName> 
```


### Scanner parameters:
- `sonar.analysis.kosli_flow=<YourFlowName>`
    - The name of the Flow relevant to your project. If a Flow does not already exist with the given name, it is created. If no Flow name is provided, the project key of your project in SonarQube is used as the name (with any invalid symbols replaced by '-').
- `sonar.analysis.kosli_trail=<YourTrailName>`
    - The name of the Trail to attest the scan results. If a Trail does not already exist with the given name it is created. If no Trail name is provided, the revision ID of the SonarQube project (typically defaulted to the Git SHA) is used as the name.
- `sonar.analysis.kosli_attestation=<YourAttestationName>`
    - The name you want to give to the attestation. If not provided, a default name "sonar" is used. If using dot-notation (of the form `<YourTargetArtifact.YourAttestationName>`), either the artifact fingerprint or git commit is also required (see below).
- `sonar.analysis.kosli_git_commit=<GitCommitSHA>`
    - The git commit for the attestation. If not provided the revision ID of the SonarQube project is used (provided it has the correct format for a git SHA).
- `sonar.analysis.kosli_artifact_fingerprint=<YourArtifactFingerprint>`
    - The fingerprint of the artifact you want the attestation to be attached to. Requires that the artifact has already been reported to Kosli.
- `sonar.analysis.kosli_flow_description=<DescriptionOfYourKosliFlow>`
    - The description for the Kosli Flow being created by this webhook. This will not be used if attesting to an already-existing Flow (i.e. will not change any existing descriptions).
- `sonar.analysis.kosli_trail_description=<DescriptionOfYourKosliTrail>`
    - The description for the Kosli Trail being created by this webhook. This will not be used if attesting to an already-existing Trail (i.e. will not change any existing descriptions).

## Testing the integration

To test the webhook once configured, simply scan a project in SonarQube. If successful, the results of the scan will be attested to the relevant Flow and Trail (and artifact, if applicable) as a sonar attestation. <br>
If the webhook fails, check that you have passed the parameters to the scanner correctly, and that the trail name, attestation name and artifact fingerprint are valid.

## Live Example in CI system
View an example of a sonar attestation via webhook in Github.

In [this YAML file](https://app.kosli.com/api/v2/livedocs/cyber-dojo/yaml?ci=github&command=-Dsonar.analysis.kosli_flow), which created [this Kosli event](https://app.kosli.com/api/v2/livedocs/cyber-dojo/event?ci=github&command=-Dsonar.analysis.kosli_flow). 


## Alternatives:
If you'd rather not use webhooks, or they don't quite fit your use-case, we also have a [CLI command](/client_reference/kosli_attest_sonar/) for attesting Sonar scan results to Kosli.

## _index.md
---
title: Older versions
bookCollapseSection: true
weight: 650
---

## _index.md
---
title: Migrations
bookCollapseSection: true
weight: 620
summary: " "
---

# Migrations

## cli_v1_to_v2.md
---
title: "CLI v0.1.x => 2.0.0 migration"
weight: 1
summary: "If you decided to migrate Kosli cli from version v0.1.x to v2.0.0 or later the table below can help you with figuring out how the commands have changed. "
---

# CLI v0.1.x => 2.0.0 migration

If you decided to migrate Kosli cli from version v0.1.x to v2.0.0 or later the table below can help you with figuring out how the commands have changed.  

{{% hint info %}}

Keep in mind that for some commands the [flag names or argument types](#flagsarguments) are also updated, so have a look at documentation for each command before switching.  
Reach out to us using [Slack](https://www.kosli.com/community/) if you find yourself in trouble.

{{% /hint %}}

## Commands

| v0.1.x                                                        | v2.0.0                                               |
|---------------------------------------------------------------|------------------------------------------------------|
| kosli approval get                                            | [kosli get approval](https://docs.kosli.com/client_reference/kosli_get_approval/)                                   |
| kosli approval ls                                             | [kosli list approvals](https://docs.kosli.com/client_reference/kosli_list_approvals/)                                   |
| kosli artifact get                                            | [kosli get artifact](https://docs.kosli.com/client_reference/kosli_get_artifact/)                                   |
| kosli artifact ls                                             | [kosli list artifacts](https://docs.kosli.com/client_reference/kosli_list_artifacts/)                                   |
| kosli assert artifact                                         | [kosli assert artifact](https://docs.kosli.com/client_reference/kosli_assert_artifact/)                                |
| kosli assert bitbucket-pullrequest                            | [kosli assert pullrequest bitbucket](https://docs.kosli.com/client_reference/kosli_assert_pullrequest_bitbucket/)                   |
| kosli assert environment                                      | [kosli assert snapshot](https://docs.kosli.com/client_reference/kosli_assert_snapshot/)                                |
| kosli assert github-pullrequest                               | [kosli assert pullrequest github](https://docs.kosli.com/client_reference/kosli_assert_pullrequest_github/)                      |
| kosli assert gitlab-mergerequest                              | [kosli assert pullrequest gitlab](https://docs.kosli.com/client_reference/kosli_assert_pullrequest_gitlab/)         |
| kosli assert status                                           | [kosli assert status](https://docs.kosli.com/client_reference/kosli_assert_status/)                                  |
| kosli commit report evidence bitbucket-pullrequest            | [kosli report evidence commit pullrequest bitbucket](https://docs.kosli.com/client_reference/kosli_report_evidence_commit_pullrequest_bitbucket/)                 |
| kosli commit report evidence generic                          | [kosli report evidence commit generic](https://docs.kosli.com/client_reference/kosli_report_evidence_commit_generic/)                    |
| kosli commit report evidence github-pullrequest               | [kosli report evidence commit pullrequest github](https://docs.kosli.com/client_reference/kosli_report_evidence_commit_pullrequest_github/)                   |
| kosli commit report evidence gitlab-mergerequest              | [kosli report evidence commit pullrequest gitlab](https://docs.kosli.com/client_reference/kosli_report_evidence_commit_pullrequest_gitlab/)                   |
| kosli commit report evidence junit                            | [kosli report evidence commit junit](https://docs.kosli.com/client_reference/kosli_report_evidence_commit_junit/)                    |
| kosli commit report evidence snyk                             | [kosli report evidence commit snyk](https://docs.kosli.com/client_reference/kosli_report_evidence_commit_snyk/)                    |
| kosli completion                                              | [kosli completion](https://docs.kosli.com/client_reference/kosli_completion/)                                     |
| kosli deployment get                                          | [kosli get deployment](https://docs.kosli.com/client_reference/kosli_get_deployment/)                                 |
| kosli deployment ls                                           | [kosli list deployments](https://docs.kosli.com/client_reference/kosli_list_deployments/)                                 |
| kosli environment allowedartifacts add                        | [kosli allow artifact](https://docs.kosli.com/client_reference/kosli_allow_artifact/)                                 |
| kosli environment declare                                     | [kosli create environment](https://docs.kosli.com/client_reference/kosli_create_environment/)                             |
| kosli environment diff                                        | [kosli diff snapshots](https://docs.kosli.com/client_reference/kosli_diff_snapshots/)                                 |
| kosli environment get                                         | [kosli get snapshot](https://docs.kosli.com/client_reference/kosli_get_snapshot/)                                   |
| kosli environment inspect                                     | [kosli get environment](https://docs.kosli.com/client_reference/kosli_get_environment/)                                |
| kosli environment log                                         | [kosli list snapshots](https://docs.kosli.com/client_reference/kosli_list_snapshots/)                                   |
| kosli environment log --long                                  | [kosli log environment](https://docs.kosli.com/client_reference/kosli_log_environment/)                               |
| kosli environment ls                                          | [kosli list environments](https://docs.kosli.com/client_reference/kosli_list_environments/)                                |
| kosli environment rename                                      | [kosli rename environment](https://docs.kosli.com/client_reference/kosli_rename_environment/)                             |
| kosli environment report docker                               | [kosli snapshot docker](https://docs.kosli.com/client_reference/kosli_snapshot_docker/)                                |
| kosli environment report ecs                                  | [kosli snapshot ecs](https://docs.kosli.com/client_reference/kosli_snapshot_ecs/)                                   |
| kosli environment report k8s                                  | [kosli snapshot k8s](https://docs.kosli.com/client_reference/kosli_snapshot_k8s/)                                   |
| kosli environment report lambda                               | [kosli snapshot lambda](https://docs.kosli.com/client_reference/kosli_snapshot_lambda/)                                |
| kosli environment report s3                                   | [kosli snapshot s3](https://docs.kosli.com/client_reference/kosli_snapshot_s3/)                                    |
| kosli environment report server                               | [kosli snapshot server](https://docs.kosli.com/client_reference/kosli_snapshot_server/)                                |
| kosli expect deployment                                       | [kosli expect deployment](https://docs.kosli.com/client_reference/kosli_expect_deployment/)                              |
| kosli pipeline deployment report                              | [kosli expect deployment](https://docs.kosli.com/client_reference/kosli_expect_deployment/)                              |
| kosli fingerprint                                             | [kosli fingerprint](https://docs.kosli.com/client_reference/kosli_fingerprint/)                                    |
| kosli pipeline approval assert                                | [kosli assert approval](https://docs.kosli.com/client_reference/kosli_assert_approval/)                                |
| kosli pipeline approval report                                | [kosli report approval](https://docs.kosli.com/client_reference/kosli_report_approval/)                                |
| kosli pipeline approval request                               | [kosli request approval](https://docs.kosli.com/client_reference/kosli_request_approval/)                               |
| kosli pipeline artifact report creation                       | [kosli report artifact](https://docs.kosli.com/client_reference/kosli_report_artifact/)                                |
| kosli pipeline artifact report evidence bitbucket-pullrequest | [kosli report evidence artifact pullrequest bitbucket](https://docs.kosli.com/client_reference/kosli_report_evidence_artifact_pullrequest_bitbucket/) |
| kosli pipeline artifact report evidence generic               | [kosli report evidence artifact generic](https://docs.kosli.com/client_reference/kosli_report_evidence_artifact_generic/)               |
| kosli pipeline artifact report evidence github-pullrequest    | [kosli report evidence artifact pullrequest github](https://docs.kosli.com/client_reference/kosli_report_evidence_artifact_pullrequest_github/)    |
| kosli pipeline artifact report evidence gitlab-mergerequest   | [kosli report evidence artifact pullrequest gitlab](https://docs.kosli.com/client_reference/kosli_report_evidence_artifact_pullrequest_gitlab/)    |
| kosli pipeline artifact report evidence junit                 | [kosli report evidence artifact junit](https://docs.kosli.com/client_reference/kosli_report_evidence_artifact_junit/)                 |
| pipeline artifact report evidence test                        | [kosli report evidence artifact junit](https://docs.kosli.com/client_reference/kosli_report_evidence_artifact_junit/)                 |
| kosli pipeline artifact report evidence snyk                  | [kosli report evidence artifact snyk](https://docs.kosli.com/client_reference/kosli_report_evidence_artifact_snyk/)                  |
| kosli pipeline declare                                        | [kosli create flow](https://docs.kosli.com/client_reference/kosli_create_flow/)                                    |
| kosli pipeline inspect                                        | [kosli get flow](https://docs.kosli.com/client_reference/kosli_get_flow/)                                       |
| kosli pipeline ls                                             | [kosli list flows](https://docs.kosli.com/client_reference/kosli_list_flows/)                                       |
| kosli search                                                  | [kosli search](https://docs.kosli.com/client_reference/kosli_search/)                                         |
| kosli status                                                  | [kosli status](https://docs.kosli.com/client_reference/kosli_status/)                                         |
| kosli version                                                 | [kosli version](https://docs.kosli.com/client_reference/kosli_version/)                                        |

## Flags/Arguments

| v0.1.x                                                        | v2.0.0                                               |
|---------------------------------------------------------------|------------------------------------------------------|
| Pipeline as argument (for some commands)               | **--flow**                              |
|  **--owner**                                                   | **--org**           |
|  **--sha256**                                                   | **--fingerprint**           |
|  **--pipeline**                                                   | **--flow**           |
|  **--pipelines**                                                   | **--flows**           |
|  **--evidence-type**                                                   | **--name**           |

## migrate_to_flows_with_trails.md
---
title: "Migrate your flows reporting to use trails and attestations"
weight: 2
summary: "Initially, flows in Kosli represented artifacts and their evidence. Trails was then introduced to give users more flexibility to model business and/or software workflows they care about in Kosli flows. "
---

# Migrate your flows reporting to use trails and attestations

Initially, flows in Kosli represented artifacts and their evidence. [Trails was then introduced](https://www.kosli.com/blog/how-to-record-an-audit-trail-for-any-devops-process-with-kosli-trails/) to give users more flexibility to model business and/or software workflows they care about in Kosli flows. 

In October 2024, we will start migrating all flows data to use trails and all evidence to attestations. During (and after) the migration, deprecated CLI commands and API endpoints continue to work and get converted on-the-fly to use trails and attestations.

This guide aims to help users switch from the deprecated CLI commands to the newer first-class commands for flows with trails and attestations.

## CLI commands

Replace the commands in the first column with their counterpart from the middle column. Be sure to check each command documentation in the [CLI reference](https://docs.kosli.com/client_reference/) as it may have new or changed flags compared to the deprecated commands.

| Deprecated Command                        | Use this command instead              | Remarks                                                                                                    |
|-------------------------------------------|---------------------------------------|------------------------------------------------------------------------------------------------------------|
| kosli create flow --template ...          | kosli create flow --template-file ... | --template is deprecated. use --template-file instead.  Template files adhere to the format defined [here](https://docs.kosli.com/template_ref/).  Creating the flow can be skipped and Kosli will auto-create one for you when you report trails, artifacts or attestations on a flow name that does not exist. |
|           | kosli begin trail ... | Creating the trail can be skipped and Kosli will auto-create one for you when you report artifacts or attestations on a trail name that does not exist. |
| kosli report artifact ...                 | kosli attest artifact ...             |                                                                                                            |
| kosli report evidence artifact <type> ... | kosli attest <type> ...               |                                                                                                            |
| kosli report evidence commit <type> ...   | kosli attest <type> ...               |                                                                                                            |

## _index.md
---
title: Search results
hideToC: true
bookhidden: true
---

## _index.md
---
title: Flow Template Specification
summary: "This document describes the specification for how to write your Flow Template files in [YAML](http://yaml.org/). The template file contains the following fields:"
---
# Flow Template Specification

This document describes the specification for how to write your Flow Template files in [YAML](http://yaml.org/). The template file contains the following fields:

```yml
version: The version of the specification schema. Allowed values are [1]. (required)
trail: # the trail specification (optional)
  attestations: # what attestations are required for the trail to be compliant (optional)
  - name: the attestation name (required)
    type: the attestation type. One of [generic, jira, junit, pull_request, snyk, sonar, '*'] (required)
  artifacts: # what artifacts are expected to be produced in the trail (optional)
  - name: reference name for the artifact (e.g. frontend-app) (required)
    attestations: # what attestations are required for the artifact to be compliant
    - name: the attestation name (required)
      type: the attestation type. One of [generic, jira, junit, pull_request, snyk, sonar, custom:<custom-type-name>] (required)
```
 
## Example:

```yaml
version: 1
trail:
  attestations:
  - name: jira-ticket
    type: jira
  - name: risk-level-assessment
    type: generic
  artifacts:
  - name: backend
    attestations:
    - name: unit-tests
      type: junit
    - name: security-scan
      type: snyk
  - name: frontend
    attestations:
    - name: manual-ui-test
      type: generic
    - name: coverage-metrics
      type: custom:coverage-metrics
```


## _index.md
---
title: Tutorials
bookCollapseSection: true
weight: 500
---

## attest_snyk.md
---
title: "Attesting Snyk scans"
bookCollapseSection: false
weight: 507
summary: "In this tutorial, we will see how you can run and attest different types of Snyk scans to Kosli. We will run the scans on the Kosli CLI git repo"
---

# Attesting Snyk Scans

Snyk scans analyze your source code, docker images and IaC source for security issues and vulnerabilities. Reporting these results to Kosli is beneficial for: 
- Tracking whether the snyk scan happened on a given artifact or trail or not.
- Keeping a record of the findings.  

In this tutorial, we will see how you can run and attest different types of Snyk scans to Kosli. We will run the scans on the [Kosli CLI git repo](https://github.com/kosli-dev/cli).

{{<hint info>}}
While snyk attestations can be bound to a trail or an artifact in a trail, this tutorial
demonstrates it only on trails for simplicity.
{{</hint>}}

## Getting ready

To follow the steps in this tutorial, you need to:
* [Setup Snyk on your machine](https://docs.snyk.io/snyk-cli/getting-started-with-the-snyk-cli#install-the-snyk-cli-and-authenticate-your-machine).
* [Install Helm](https://helm.sh/docs/intro/install/) if you want to try Snyk IaC attestations, otherwise skip.
* [Install Docker](https://docs.docker.com/engine/install/) if you want to try Snyk container attestations, otherwise skip.
* [Create a Kosli account](https://app.kosli.com/) (Skip if you already have one).
* [Install Kosli CLI](/getting_started/install/).
* [Get a Kosli API token](/getting_started/service-accounts/).
* Set the `KOSLI_ORG` environment variable to your personal org name and `KOSLI_API_TOKEN` to your token:
  ```shell {.command}
  export KOSLI_ORG=<your-personal-kosli-org-name>
  export KOSLI_API_TOKEN=<your-api-token>
  ```
* Clone the Kosli CLI git repo
  ```shell {.command}
  git clone https://github.com/kosli-dev/cli.git 
  cd cli
  ```

## Creating a Flow and Trail

We will start by creating a flow in Kosli to contain Trails and Artifacts for this demo.

```shell {.command}
kosli create flow snyk-demo --use-empty-template
```

{{<hint info>}}
`--use-empty-template` indicates that this flow does not have a predefined set of required attestations.
{{</hint>}}

Then, we can start a trail to bind our snyk attestations to.

```shell {.command}
kosli begin trail test-1 --flow snyk-demo
```

Now we can start running Snyk scans and attest them to this trail.

{{<hint info>}}
After each attestation in the sections below, you can navigate to:
**https://app.kosli.com/\<your-personal-org-name\>/flows/snyk-demo/trails/test-1** to view the status of the trail in Kosli.
{{</hint>}}

## Snyk Open source scan

[Snyk Open Source](https://docs.snyk.io/scan-using-snyk/snyk-open-source) allows you to find and fix vulnerabilities in the open-source libraries used by your applications. 

You can run a snyk opens source scan and report it to Kosli as follows:
```shell {.command}
snyk test --sarif-file-output=os.json

kosli attest snyk --flow snyk-demo --trail test-1 --name open-source-scan --scan-results os.json --commit HEAD
```

{{<hint info>}}
`--commit` allows you to relate the attestation to a specific git commit.
{{</hint>}}


## Snyk Code scan

[Snyk Code](https://docs.snyk.io/scan-using-snyk/snyk-code) lets you scan your source code for security issues. 

You can run a snyk code scan and report it to Kosli as follows:
```shell {.command}
snyk code test --sarif-file-output=code.json

kosli attest snyk --flow snyk-demo --trail test-1 --name code-scan --scan-results code.json --commit HEAD
```

## Snyk Container scan

[Snyk Container](https://docs.snyk.io/scan-using-snyk/snyk-container) lets you scan your container images for security issues. 

You can run a snyk container scan and report it to Kosli as follows:
```shell {.command}
# pull the cli docker image before scanning it
docker pull ghcr.io/kosli-dev/cli:v2.8.3
snyk container test ghcr.io/kosli-dev/cli:v2.8.3  --file=Dockerfile --sarif-file-output=container.json

kosli attest snyk --flow snyk-demo --trail test-1 --name container-scan --scan-results container.json --commit HEAD
```

## Snyk IaC scan

[Snyk IaC](https://docs.snyk.io/scan-using-snyk/snyk-iac) lets you scan various types of IaC configuration files (e.g. Terraform, Kubernetes, Helm) for security issues. 

We can run a snyk IaC scan on the K8S reporter Helm chart and report it to Kosli as follows:
```shell {.command}
helm template ./charts/k8s-reporter --output-dir helm \
  --set kosliApiToken.secretName=secret \
  --set reporterConfig.kosliEnvironmentName=foo \
  --set reporterConfig.kosliOrg=bar

snyk iac test helm  --sarif-file-output=helm.json

kosli attest snyk --flow snyk-demo --trail test-1 --name helm-scan --scan-results helm.json --commit HEAD
```

You can refer to the [Snyk docs](https://docs.snyk.io/snyk-cli/scan-and-maintain-projects-using-the-cli/snyk-cli-for-iac/test-your-iac-files) for more information on supported IaC configuration formats and how you can run snyk scans on them.

For more details about the `kosli attest snyk` command, please refer to [its CLI reference](/client_reference/kosli_attest_snyk/).

## cli_and_http_proxy.md
---
title: "Using Kosli CLI with an HTTP proxy"
bookCollapseSection: false
weight: 511
summary: "In enterprises with strict network policies, you might want to communicate with Kosli via an HTTP proxy as the single point of egress communication with Kosli. This tutorial shows how you can setup an HTTP proxy and use it when communicating with Kosli via the CLI."
---

# Using Kosli CLI with an HTTP proxy

In enterprises with strict network policies, you might want to communicate with Kosli via an HTTP proxy as the single point of egress communication with Kosli. 

This tutorial shows how you can setup an HTTP proxy and use it when communicating with Kosli via the CLI.

## TLDR

If you already have an HTTP proxy, [start using it with Kosli CLI](#use-the-http-proxy-with-kosli-cli)


{{<hint info>}}
In this tutorial, we will setup Tinyproxy (in docker) as an HTTP proxy on a Mac machine.
The same steps apply for different HTTP proxies and machines, but commands will differ.
{{</hint>}}


## Start the HTTP proxy

1. Start Tinyproxy using docker:

```shell {.command}
cat <<EOF > tinyproxy.conf
User nobody
Group nobody
Port 8888
EOF

docker run -p 8888:8888 -v $(PWD)/tinyproxy.conf:/etc/tinyproxy/tinyproxy.conf:ro kalaksi/tinyproxy
```



Now you have an HTTP proxy running at http://localhost:8888

## Use the HTTP proxy with Kosli CLI

To verify if the setup works, you can run this command to list environments of the public demo org `Cyber Dojo`:

```shell {.command}
kosli list envs --org cyber-dojo --http-proxy http://localhost:8888 --api-token <<your-token>>
```

Your request goes through the HTTP proxy and is then forwarded to Kosli. If successful, you should see a similar output to this:

```
NAME                         TYPE  LAST REPORT                LAST MODIFIED              TAGS
aws-beta                     ECS   2024-04-18T15:17:54+02:00  2024-04-18T15:17:54+02:00  [url=https://beta.cyber-dojo.org/]
aws-prod                     ECS   2024-04-18T15:17:57+02:00  2024-04-18T15:17:57+02:00  [url=https://cyber-dojo.org/]
terraform-state-differ-beta  S3    2024-04-18T15:18:23+02:00  2024-04-18T15:18:23+02:00  
terraform-state-differ-prod  S3    2024-04-18T15:18:17+02:00  2024-04-18T15:18:17+02:00 
```

All you need to do now is to use `--http-proxy http://localhost:8888` with your CLI commands.
Alternatively, you can add this to your kosli config so that you don't type it on each command:
`kosli config --http-proxy=http://localhost:8888`


## following_a_git_commit_to_runtime_environments.md
---
title: Following a git commit to runtime environments
bookCollapseSection: false
weight: 510
draft: false
summary: "In this 5 minute tutorial you'll learn how Kosli tracks \"life after git\" and shows you events from CI pipelines (eg, building the docker image, running the unit tests, deploying, etc) and runtime environments (eg, the blue-green rollover, instance scaling, etc)"
---

<!-- The book "Developer Marketing Does Not Exist" by Adam DuVander suggests 
     this tutorial content structure (p49)
     1. Explain the context
     2. Show the end result
     3. Walk through the steps
     4. Help them take the next step

     Quoting from the book...
       "If you find the first three words of your tutorial are
        In this tutorial" then you might have skipped ahead."
     These were our exact first three words!
     I've tried to add initial context.
     I think we are still missing step 2 (see below)
     I think we are still missing step 4, which should probably
       be a simple link to the next tutorial.
-->

# Following a git commit to runtime environments

## Overview

In this 5 minute tutorial you'll learn how Kosli tracks "life after git" and shows you events from:
* CI pipelines (eg, building the docker image, running the unit tests, deploying, etc)
* runtime environments (eg, the blue-green rollover, instance scaling, etc)

You'll follow an actual git commit to an open-source project called **cyber-dojo**. 
In our example cyber-dojo’s `runner` service should run with three replicas. However, due to an oversight while switching
from Google Kubernetes Engine (GKE) to AWS Elastic Container Service (ECS), it was running with just one replica. 
You will follow the commit that fixed this. 

## Getting ready

You need to:
* [Install Kosli CLI](/getting_started/install/).
* [Get a Kosli API token](/getting_started/service-accounts/).
* Set the `KOSLI_ORG` environment variable to `cyber-dojo` (the Kosli `cyber-dojo` organization is public so any authenticated user can read its data) and `KOSLI_API_TOKEN` to your token:
  ```shell {.command}
  export KOSLI_ORG=cyber-dojo
  export KOSLI_API_TOKEN=<your-api-token>
  ```

## CI Pipeline events

### Listing flows

Find out which `cyber-dojo` repositories have a CI pipeline reporting to [Kosli](https://app.kosli.com):

```shell {.command}
kosli ls flows
```

You will see:

```plaintext {.light-console}
NAME                    DESCRIPTION                         VISIBILITY
creator                 UX for Group/Kata creation          public
custom-start-points     Custom exercises choices            public
dashboard               UX for a group practice dashboard   public
differ                  Diff files from two traffic-lights  public
exercises-start-points  Exercises choices                   public
languages-start-points  Language+TestFramework choices      public
nginx                   Reverse proxy                       public
repler                  REPL for Python images              public
runner                  Test runner                         public
saver                   Group/Kata model+persistence        public
version-reporter        UX for git+image version-reporter   public
web                     UX for practicing TDD               public
```

{{% hint info %}}
## cyber-dojo overview
* [cyber-dojo](https://cyber-dojo.org) is a web platform where teams 
practice TDD without any installation.  
* cyber-dojo has a microservice architecture with a dozen git repositories.
* Each git repository has its own Github Actions CI pipeline producing a docker image.
* These docker images run in two AWS environments named 
[aws-beta](https://app.kosli.com/cyber-dojo/environments/aws-beta)
and [aws-prod](https://app.kosli.com/cyber-dojo/environments/aws-prod).
{{% /hint %}}


### Following the artifact

The runner service had one instance running instead of three.
The commit which fixed the problem was 
[16d9990](https://github.com/cyber-dojo/runner/commit/16d9990ad23a40eecaf087abac2a58a2d2a4b3f4)
in the `runner` repository. Follow this commit using the `kosli` command:

```shell {.command}
kosli get artifact runner:16d9990
```
You will see:

```plaintext {.light-console}
Name:         cyberdojo/runner:16d9990
Flow:         runner
Fingerprint:  9af401c4350b21e3f1df17d6ad808da43d9646e75b6da902cc7c492bcfb9c625
Created on:   Mon, 22 Aug 2022 11:35:00 CEST • 15 days ago
Git commit:   16d9990ad23a40eecaf087abac2a58a2d2a4b3f4
Commit URL:   https://github.com/cyber-dojo/runner/commit/16d9990ad23a40eecaf087abac2a58a2d2a4b3f4
Build URL:    https://github.com/cyber-dojo/runner/actions/runs/2902808452
State:        COMPLIANT
History:
    Artifact created                                     Mon, 22 Aug 2022 11:35:00 CEST
    branch-coverage evidence received                    Mon, 22 Aug 2022 11:36:02 CEST
    Deployment #18 to aws-beta environment               Mon, 22 Aug 2022 11:37:17 CEST
    Deployment #19 to aws-prod environment               Mon, 22 Aug 2022 11:38:21 CEST
    Started running in aws-beta#84 environment           Mon, 22 Aug 2022 11:38:28 CEST
    Started running in aws-prod#65 environment           Mon, 22 Aug 2022 11:39:22 CEST
    Scaled down from 3 to 2 in aws-beta#117 environment  Wed, 24 Aug 2022 18:03:42 CEST
    No longer running in aws-beta#119 environment        Wed, 24 Aug 2022 18:05:42 CEST
    Scaled down from 3 to 1 in aws-prod#94 environment   Wed, 24 Aug 2022 18:10:28 CEST
    No longer running in aws-prod#96 environment         Wed, 24 Aug 2022 18:12:28 CEST
```

Let's look at this output in detail:

* **Name**: The name of the docker image is `cyberdojo/runner:16d9990`. Its image registry is defaulted to
`dockerhub`. Its :tag is the short-sha of the git commit.
* **Flow**: The name of the Kosli flow.
* **Fingerprint**: The unique immutable SHA256 fingerprint of the artifact.
* **Created on**: The artifact was created on 22nd August 2022, at 11:35 CEST.
* **Commit URL**: You can follow [the commit URL](https://github.com/cyber-dojo/runner/commit/16d9990ad23a40eecaf087abac2a58a2d2a4b3f4) 
  to the actual commit on Github since cyber-dojo's git repositories are public.
* **Build URL**: You can follow [the build URL](https://github.com/cyber-dojo/runner/actions/runs/2902808452)
  to the actual Github Action for this commit.
* **State**: COMPLIANT means that all the promised evidence for the artifact (in this case `branch-coverage`)
  was provided before deployment.
* **History**:
   * **CI pipeline events**
      * The artifact was **created** on the 22nd August at 11:35:00 CEST.
      * The artifact has `branch-coverage` **evidence**. 
      * The artifact was **deployed** to [aws-beta](https://app.kosli.com/cyber-dojo/flows/runner/deployments/18) on 22nd  August 11:37:17 CEST, and to [aws-prod](https://app.kosli.com/cyber-dojo/flows/runner/deployments/19)
     a minute later.
   * **Runtime environment events**
      * The artifact was reported **running** in both environments.
      * The artifact's number of running instances **scaled down**.
      * The artifact was reported **exited**.
     
The information about this artifact is also available through the [web interface](https://app.kosli.com/cyber-dojo/flows/runner/artifacts/9af401c4350b21e3f1df17d6ad808da43d9646e75b6da902cc7c492bcfb9c625).

{{% hint info %}}
The `runner` service uses [Continuous Deployment](https://en.wikipedia.org/wiki/Continuous_deployment); 
if the tests pass the artifact is [blue-green deployed](https://en.wikipedia.org/wiki/Blue-green_deployment) 
to both its runtime environments *without* any manual approval steps.
Some cyber-dojo services (eg web) have a manual approval step, and Kosli supports this.
{{% /hint %}}

## Environment Snapshots

Kosli environments store information about what is running in your actual runtime environments (eg server, Kubernetes cluster, AWS, ...).
We use one Kosli environment per runtime environment.

The Kosli CLI periodically fingerprints all the running artifacts in a runtime environment and reports them to Kosli.
Whenever a change is detected, a snapshot of the environment is saved.

{{% hint info %}}
Cyber-dojo runs the `kosli` CLI from inside its AWS runtime environments
using a [lambda function](https://github.com/cyber-dojo/kosli-environment-reporter/blob/main/deployment/terraform/deployment.tf)
to report the running services to Kosli.
{{% /hint %}}


The **History** of the artifact tells you your artifact started running in snapshot #65 of `aws-prod`.

You query Kosli to see what was running in `aws-prod` snapshot #65:

```shell {.command}
kosli get snapshot aws-prod#65
```

The output will be:

```plaintext {.light-console}
COMMIT   ARTIFACT                                                                         FLOW       RUNNING_SINCE  REPLICAS
16d9990  Name: 274425519734.dkr.ecr.eu-central-1.amazonaws.com/runner:16d9990             runner     11 days ago    3
         Fingerprint: 9af401c4350b21e3f1df17d6ad808da43d9646e75b6da902cc7c492bcfb9c625                              
                                                                                                    
7c45272  Name: 274425519734.dkr.ecr.eu-central-1.amazonaws.com/shas:7c45272               shas       11 days ago    1
         Fingerprint: 76c442c04283c4ca1af22d882750eb960cf53c0aa041bbdb2db9df2f2c1282be                              

...some output elided...

85d83c6  Name: 274425519734.dkr.ecr.eu-central-1.amazonaws.com/runner:85d83c6             runner     13 days ago    1
         Fingerprint: eeb0cfc9ee7f69fbd9531d5b8c1e8d22a8de119e2a422344a714a868e9a8bfec                              
                                                                                                  
1a2b170  Name: 274425519734.dkr.ecr.eu-central-1.amazonaws.com/differ:1a2b170             differ     13 days ago    1
         Fingerprint: d8440b94f7f9174c180324ceafd4148360d9d7c916be2b910f132c58b8a943ae                              
```

You see in this snapshot, the `runner:16d9990` artifact is indeed running with 3 replicas.
You have proof the git commit has worked. 

{{% hint info %}}
## Blue-green deployment
There were *two* versions of `runner` at this point in time! 
The first had three replicas (to fix the problem), but there was also a second (from commit `85d83c6`) with only one replica.

You are seeing a **blue-green deployment** happening;
`runner:85d83c6` is about to be stopped and will not be reported in
snapshot `aws-prod#66`.
{{% /hint %}}

## Diffing snapshots

Kosli's `env diff` command allows you to see differences between two versions of your
runtime environment.

Let's find out what's *different* between the `aws-prod#64` and `aws-prod#65` snapshots: 

```shell {.command}
kosli diff snapshots aws-prod#64 aws-prod#65
```

The response will be:

```plaintext {.light-console}
Only present in aws-prod#65
                   
     Name:         274425519734.dkr.ecr.eu-central-1.amazonaws.com/runner:16d9990
     Fingerprint:  9af401c4350b21e3f1df17d6ad808da43d9646e75b6da902cc7c492bcfb9c625
     Flow:         runner
     Commit URL:   https://github.com/cyber-dojo/runner/commit/16d9990ad23a40eecaf087abac2a58a2d2a4b3f4
     Started:      Mon, 22 Aug 2022 11:39:17 CEST • 15 days ago
```

The output above shows that `runner:16d9990` started running in snapshot 65 of `aws-prod` environment.

You have seen how Kosli can follow a git commit on its way into production,
and provide information about the artifacts history, without any access to cyber-dojo's `aws-prod` environment.


## get_familiar_with_Kosli.md
---
title: Get familiar with Kosli
bookCollapseSection: false
weight: 504
summary: "The following guide is the easiest and quickest way to try Kosli out and understand its features. 
It is made to run from your local machine, but the same concepts and steps apply to using Kosli in a production setup."
---

# Get familiar with Kosli

> The following guide is the easiest and quickest way to try Kosli out and understand its features. 
It is made to run from your local machine, but the same concepts and steps apply to using Kosli in a production setup.

In this tutorial, you'll learn how Kosli allows you to track a source code change from runtime environments.
You'll set up a `docker` environment, use Kosli to record build and deployment events, and track what 
artifacts are running in your runtime environment. 

This tutorial uses the `docker` Kosli environment type, but the same steps can be applied to 
other supported environment types.

{{% hint info %}}
As you go through the guide you can also check your progress from 
[your browser](https://app.kosli.com).

In the upper left corner there is a house icon. Next to it you can select
which organization you want to view. Your personal organization
has the same name as your GitHub login name, and is the organization you will
be using in this guide.
{{% /hint %}}

> [Playground](https://github.com/kosli-dev/playground?tab=readme-ov-file#kosli-playground) is an alternative version of this tutorial in which you
embed the Kosli commands in a GitHub CI Workflow (in a clone of the playground repo) rather than running them 
directly from your terminal. 


## Step 1: Prerequisites and Kosli account

To follow the tutorial, you will need to:

- Install `Docker`.
- [Create a Kosli account](https://app.kosli.com/sign-up) if you have not got one already.
- [Install Kosli CLI](/getting_started/install/).
- [Get a Kosli API token](/getting_started/service-accounts/).
- Set the `KOSLI_ORG` and `KOSLI_API_TOKEN` environment variables:
  ```shell {.command}
  export KOSLI_ORG=<your-org>
  export KOSLI_API_TOKEN=<your-api-token>
  ```
- You can check your Kosli set up by running: 
    ```shell {.command}
    kosli list flows
    ```
    which should return a list of flows or the message "No flows were found".

- Clone our quickstart-docker repository:
    ```shell {.command}
    git clone https://github.com/kosli-dev/quickstart-docker-example.git
    cd quickstart-docker-example
    ```
- Export the head commit in a variable (will be used in several of the commands below):
  ```shell {.command}
  export GIT_COMMIT=$(git rev-parse HEAD)
  ```

## Step 2: Create a Kosli Flow

<!--
For this tutorial we are using the first kind:
- the *Flow* corresponds to the git repository
- the Artifact being built and deployed is an `nginx` docker image
When attesting evidence, the target of the attestation must be named.
These names are defined in a yml file.
-->

The Flow's yml template-file exists in the git repository.
Confirm this yml file exists by catting it:

```shell {.command}
cat kosli.yml
```

You will see the following output, specifying the existence of an Artifact named `nginx`:


```plaintext {.light-console}
trail:
  artifacts:
    - name: nginx
```

Create a Kosli *Flow* called `quickstart-nginx` using this yml template-file:

```shell {.command}
kosli create flow quickstart-nginx \
    --description "Flow for quickstart nginx image" \
    --template-file kosli.yml
```

Confirm the Kosli Flow called `quickstart-nginx` was created:

```shell {.command}
kosli list flows
```

which will produce the following output:

```plaintext {.light-console}
NAME              DESCRIPTION                          VISIBILITY
quickstart-nginx  Flow for quickstart nginx image      private
```
{{% hint info %}}
In the web interface you can select *Flows* on the left.
It will show you that you have a *quickstart-nginx* Flow.
If you select the Flow it will show that no Artifacts have
been reported yet.
{{% /hint %}}


## Step 3: Create a Kosli Trail

Create a Kosli *Trail*, in the `quickstart-nginx` Flow, whose 
name is the repository's current git-commit:

```shell {.command}
kosli begin trail ${GIT_COMMIT} \
    --flow quickstart-nginx
```

<!--
Step to confirm the Trail exists?
-->

## Step 4: Attest an Artifact to Kosli

Typically, you would build an Artifact in your CI system, in response to a git-commit being pushed.
The quickstart-docker repository contains a `docker-compose.yml` file that uses a public [nginx](https://nginx.org/) 
docker image which you will be using as your Artifact in this tutorial instead.

<!-- Pull the docker image - the Kosli CLI needs the Artifact to be locally present to 
generate a "fingerprint" to identify it:

```shell {.command}
docker compose pull
```

You can check that this has worked by typing: 
```shell {.command}
docker images nginx
```
The output should look like this:
```plaintext {.light-console}
REPOSITORY   TAG       IMAGE ID       CREATED        SIZE
nginx        1.21      8f05d7383593   5 months ago   134MB
``` -->

Now report the artifact to Kosli using the `kosli attest artifact` command.

Note:
- The `--name` flag has the value `nginx` which is the (only) artifact
name defined in the `kosli.yml` file from step 2.
- The `--build-url` and `--commit-url` flags have dummy values;
in a real call these would be the CI and git hosting provider URLs respectively.

```shell {.command}
kosli attest artifact nginx:1.21 \
    --name nginx \
    --flow quickstart-nginx \
    --trail ${GIT_COMMIT} \
    --artifact-type oci \
    --build-url https://example.com \
    --commit-url https://github.com/kosli-dev/quickstart-docker-example/commit/9f14efa0c91807da9a8b1d1d6332c5b3aa24a310 \
    --commit $(git rev-parse HEAD)    
```

<!--
It is noticeable here that we are providing the git-commit twice;
once for the name of the trail, and once for the actual git-commit.
It is also noticeable that the git-commit is hard-wired to 9f14efa...
in several places, and it will be incorrect whenever the repo gets new git
commit (eg to add the kosli.yml file)
-->

You can verify that you have reported the Artifact in your *quickstart-nginx* flow:

```shell {.command}
kosli list artifacts --flow quickstart-nginx
```

```plaintext {.light-console}
COMMIT   ARTIFACT                                                                       STATE      CREATED_AT
9f14efa  Name: nginx:1.21                                                               COMPLIANT  Tue, 01 Nov 2022 15:46:59 CET
         Fingerprint: 2bcabc23b45489fb0885d69a06ba1d648aeda973fae7bb981bafbb884165e514                 
```


## Step 5: Create a Kosli environment

<!--
A Kosli *Environment* stores snapshots containing information about
the software Artifacts you are running in your runtime environment.
Kosli supports many kinds of runtime environments; (server, Kubernetes cluster, AWS, etc.)  
-->

Create a Kosli *Environment* called `quickstart` whose type is `docker`:

```shell {.command}
kosli create environment quickstart \
    --type docker \
    --description "quickstart environment for tutorial"
```

You can verify that the Kosli *Environment* was created:

```shell {.command}
kosli list environments
```

```plaintext {.light-console}
NAME        TYPE    LAST REPORT  LAST MODIFIED
quickstart  docker               2022-11-01T15:30:56+01:00
```

{{% hint info %}}
If you refresh the *Environments* web page in your Kosli account, 
it will show you that you have a *quickstart* environment and that
no snapshot reports have been received yet.
{{% /hint %}}


## Step 6: Report what is running in your environment

First, run the artifact:
```shell {.command}
docker compose up -d
```

Confirm the container is running:

```shell {.command}
docker ps
```
The output should include an entry similar to this:

```plaintext {.light-console}
CONTAINER ID  IMAGE      COMMAND                 CREATED         STATUS         PORTS                  NAMES
6330e545b532  nginx:1.21 "/docker-entrypoint.…"  35 seconds ago  Up 34 seconds  0.0.0.0:8080->80/tcp   quickstart-nginx
```

Report all the docker containers running on your machine to Kosli:

```shell {.command}
kosli snapshot docker quickstart
```

You can confirm this has created an environment snapshot:

```shell {.command}
kosli list snapshots quickstart
```
```plaintext {.light-console}
SNAPSHOT  FROM                           TO   DURATION
1         Tue, 01 Nov 2022 15:55:49 CET  now  11 seconds
```

You can get a detailed view of all the docker containers included in the snapshot report:

```shell {.command}
kosli get snapshot quickstart
```

```plaintext {.light-console}
COMMIT  ARTIFACT                                                                       FLOW  RUNNING_SINCE  REPLICAS
N/A     Name: nginx:1.21                                                               N/A   3 minutes ago  1
        Fingerprint: 8f05d73835934b8220e1abd2f157ea4e2260b9c26f6f63a8e3975e7affa46724
```

The `kosli snapshot docker` command reports *all* the 
docker containers running in your environment, equivalent to the output from 
`docker ps`. This tutorial only shows the `nginx` container 
in the examples.

{{% hint info %}}
If you refresh the *Environments* web page in your Kosli account, you will see 
that there is now a timestamp for *Last Change At* column. 
Select the *quickstart* link on left for a detailed view of what is currently running.
{{% /hint %}}

## Step 7: Searching Kosli

Now that you have reported your Artifact and what's running in your runtime environment,
you can use the `kosli search` command to find everything Kosli knows about an Artifact or a git-commit.

For example, you can give Kosli search the git-commit whose CI run built and deployed the Artifact: 

```shell {.command}
kosli search ${GIT_COMMIT}
```

```plaintext {.light-console}
Search result resolved to commit 9f14efa0c91807da9a8b1d1d6332c5b3aa24a310
Name:              nginx:1.21
Fingerprint:       2bcabc23b45489fb0885d69a06ba1d648aeda973fae7bb981bafbb884165e514
Has provenance:    true
Flow:              quickstart-nginx
Git commit:        9f14efa0c91807da9a8b1d1d6332c5b3aa24a310
Commit URL:        https://github.com/kosli-dev/quickstart-docker-example/commit/9f14efa0c91807da9a8b1d1d6332c5b3aa24a310
Build URL:         https://example.com
Compliance state:  COMPLIANT
History:
    Artifact created                             Tue, 01 Nov 2022 15:46:59 CET
    Deployment #1 to quickstart environment      Tue, 01 Nov 2022 15:48:47 CET
    Started running in quickstart#1 environment  Tue, 01 Nov 2022 15:55:49 CET
```

Visit the [Kosli Querying](/tutorials/querying_kosli/) guide to learn more about the search command.

## querying_kosli.md
---
title: "Querying Kosli"
bookCollapseSection: false
weight: 505
summary: "All the information stored in Kosli may be helpful both for operations and development. A set of `get`, `list`, `log` and `assert` commands allows you to quickly access the information about your environments, artifacts and deployments, without leaving your development environment."
---

# Querying Kosli

All the information stored in Kosli may be helpful both for operations and development. A set of `get`, `list`, `log` and `assert` commands allows you to quickly access the information about your environments, artifacts and deployments, without leaving your development environment.

## Getting ready

You need to:
* [Install Kosli CLI](/getting_started/install/).
* [Get a Kosli API token](/getting_started/service-accounts/).
* Set the `KOSLI_ORG` environment variable to `cyber-dojo` (the Kosli `cyber-dojo` organization is public so any authenticated user can read its data) and `KOSLI_API_TOKEN` to your token:
  ```shell {.command}
  export KOSLI_ORG=cyber-dojo # cyber-dojo is a public demo org
  export KOSLI_API_TOKEN=<your-api-token>
  ```

## Search with git commit sha

You can use `kosli search` command to find out if Kosli knows of any artifact that was build using that commit - both short and full shas are accepted:

```shell {.command}
kosli search 0f5c9e1
```

```
Search result resolved to commit 0f5c9e19c4d4f948d19ce4c8495b2a44745cda96
Name:              cyberdojo/web:0f5c9e1
Fingerprint:       62e1d2909cc59193b31bfd120276fcb8ba5e42dd6becd873218a41e4ce022505
Has provenance:    true
Flow:              web
Git commit:        0f5c9e19c4d4f948d19ce4c8495b2a44745cda96
Commit URL:        https://github.com/cyber-dojo/web/commit/0f5c9e19c4d4f948d19ce4c8495b2a44745cda96
Build URL:         https://github.com/cyber-dojo/web/actions/runs/3021563461
Compliance state:  COMPLIANT
History:
    Artifact created                                   Fri, 09 Sep 2022 11:59:50 CEST
    Deployment #59 to aws-beta environment             Fri, 09 Sep 2022 12:01:12 CEST
    Started running in aws-beta#217 environment        Fri, 09 Sep 2022 12:02:42 CEST
    Deployment #60 to aws-prod environment             Fri, 09 Sep 2022 12:06:37 CEST
    Started running in aws-prod#202 environment        Fri, 09 Sep 2022 12:07:28 CEST
    Scaled up from 1 to 3 in aws-prod#203 environment  Fri, 09 Sep 2022 12:08:28 CEST
    No longer running in aws-beta#222 environment      Sat, 10 Sep 2022 08:44:42 CEST
    No longer running in aws-prod#210 environment      Sat, 10 Sep 2022 08:49:28 CEST
```

The information returned by `kosli search` - like Flow, Fingerprint or History - can be used to run more dedicated searches in Kosli. 

## Search for a flow

When you search in Kosli you often need to refer to a specific flow. If you don't remember all the flows' names it is easy to list them with `kosli list flows` command:

```shell {.command}
kosli list flows
```

```
NAME                    DESCRIPTION                         VISIBILITY
creator                 UX for Group/Kata creation          public
custom-start-points     Custom exercises choices            public
dashboard               UX for a group practice dashboard   public
differ                  Diff files from two traffic-lights  public
exercises-start-points  Exercises choices                   public
languages-start-points  Language+TestFramework choices      public
nginx                   Reverse proxy                       public
repler                  REPL for Python images              public
runner                  Test runner                         public
saver                   Group/Kata model+persistence        public
shas                    UX for git+image shas               public
web                     UX for practicing TDD               public
```

And if you want to check metadata of a specific flow (like description or template) use `kosli get flow`

```shell {.command}
kosli get flow creator
```

```
Name:                creator
Description:         UX for Group/Kata creation
Visibility:          public
Template:            [artifact, branch-coverage]
Last Deployment At:  Wed, 14 Sep 2022 10:51:43 CEST • one month ago
```

## List artifacts

To find the information about artifacts reported to a specific flow in Kosli use `kosli list artifacts` command

```shell {.command}
kosli list artifacts --flow creator
```

```
COMMIT   ARTIFACT                                  STATE       CREATED_AT
344430d  Name: cyberdojo/creator:344430d           COMPLIANT   Wed, 14 Sep 2022 10:48:09 CEST
         Fingerprint: 817a72(...)6b5a273399c693             
                                                                                                    
41bfb7b  Name: cyberdojo/creator:41bfb7b           COMPLIANT   Sat, 10 Sep 2022 08:41:15 CEST
         Fingerprint: 8d6fef(...)b84c281f712ef8             
                                                                                                    
aa0a3d3  Name: cyberdojo/creator:aa0a3d3           COMPLIANT   Fri, 09 Sep 2022 11:58:56 CEST
         Fingerprint: 3ede07(...)238845a631e96a             
                                                                                                    
[...]
```

The output of the command is shortened above, for readability purposes. 

The amount of artifacts may be really long and by default you can see the last 15 artifacts - the first page of the result list. You can use `-n` flag to limit the amount of artifacts displayed per page, and `--page` to select which page of the result list you want to see.

E.g. to see last five artifacts you'd use:
```shell {.command}
kosli list artifacts --flow creator -n 5
```

And to see the next page:
```shell {.command}
kosli list artifacts --flow creator -n 5 --page 2
```

You can also use the `--output` flag to change the format of the response. By default the response comes in a *table* format, but you can choose to switch to *json*:
```shell {.command}
kosli list artifacts --flow creator --output json
```
## Get artifact

To get more detailed information about a given artifact use `kosli get artifact`. To identify the artifact you need to use:
* flow name followed by `@` and artifact fingerprint
OR
* flow name followed by `:` and commit sha

Both are available in the output of `kosli list artifacts` command

```shell {.command}
# search for an artifact by its fingerprint
kosli get artifact creator@817a72609041c51cd2a3bbbcbeb048c687677986b5a273399c6938b5e6aa1ded
```

```
Name:                     cyberdojo/creator:344430d
Flow:                     creator
Fingerprint:              817a72609041c51cd2a3bbbcbeb048c687677986b5a273399c6938b5e6aa1ded
Created on:               Wed, 14 Sep 2022 10:48:09 CEST • 2 months ago
Git commit:               344430d530d26068aa1f39760a9c094c989382f3
Commit URL:               https://github.com/cyber-dojo/creator/commit/344430d530d26068aa1f39760a9c094c989382f3
Build URL:                https://github.com/cyber-dojo/creator/actions/runs/3051390570
State:                    COMPLIANT
Running in environments:  aws-beta#265, aws-prod#259
History:
    Artifact created                               Wed, 14 Sep 2022 10:48:09 CEST
    branch-coverage evidence received              Wed, 14 Sep 2022 10:49:11 CEST
    Deployment #100 to aws-beta environment        Wed, 14 Sep 2022 10:50:40 CEST
    Deployment #101 to aws-prod environment        Wed, 14 Sep 2022 10:51:43 CEST
    Started running in aws-beta#229 environment    Wed, 14 Sep 2022 10:52:42 CEST
    Started running in aws-prod#217 environment    Wed, 14 Sep 2022 10:53:28 CEST
    No longer running in aws-prod#252 environment  Fri, 14 Oct 2022 08:17:28 CEST
    Started running in aws-prod#254 environment    Fri, 14 Oct 2022 08:22:28 CEST
    No longer running in aws-beta#254 environment  Fri, 14 Oct 2022 16:35:42 CEST
    Started running in aws-beta#256 environment    Fri, 14 Oct 2022 16:38:42 CEST
    No longer running in aws-beta#257 environment  Sun, 16 Oct 2022 07:45:42 CEST
    Started running in aws-beta#259 environment    Sun, 16 Oct 2022 07:49:42 CEST
    No longer running in aws-beta#260 environment  Wed, 19 Oct 2022 09:28:42 CEST
    Started running in aws-beta#262 environment    Wed, 19 Oct 2022 09:32:42 CEST
    No longer running in aws-beta#263 environment  Wed, 19 Oct 2022 09:42:42 CEST
    Started running in aws-beta#265 environment    Wed, 19 Oct 2022 09:46:42 CEST
    No longer running in aws-prod#257 environment  Fri, 21 Oct 2022 11:02:28 CEST
    Started running in aws-prod#259 environment    Fri, 21 Oct 2022 11:05:28 CEST
```

```shell {.command}
# search for an artifact by its commit sha
kosli get artifact creator:344430d
```

```
Name:                     cyberdojo/creator:344430d
Flow:                     creator
Fingerprint:              817a72609041c51cd2a3bbbcbeb048c687677986b5a273399c6938b5e6aa1ded
Created on:               Wed, 14 Sep 2022 10:48:09 CEST • 2 months ago
Git commit:               344430d530d26068aa1f39760a9c094c989382f3
Commit URL:               https://github.com/cyber-dojo/creator/commit/344430d530d26068aa1f39760a9c094c989382f3
Build URL:                https://github.com/cyber-dojo/creator/actions/runs/3051390570
State:                    COMPLIANT
Running in environments:  aws-beta#265, aws-prod#259
History:
    Artifact created                               Wed, 14 Sep 2022 10:48:09 CEST
    branch-coverage evidence received              Wed, 14 Sep 2022 10:49:11 CEST
    Deployment #100 to aws-beta environment        Wed, 14 Sep 2022 10:50:40 CEST
    Deployment #101 to aws-prod environment        Wed, 14 Sep 2022 10:51:43 CEST
    Started running in aws-beta#229 environment    Wed, 14 Sep 2022 10:52:42 CEST
    Started running in aws-prod#217 environment    Wed, 14 Sep 2022 10:53:28 CEST
    No longer running in aws-prod#252 environment  Fri, 14 Oct 2022 08:17:28 CEST
    Started running in aws-prod#254 environment    Fri, 14 Oct 2022 08:22:28 CEST
    No longer running in aws-beta#254 environment  Fri, 14 Oct 2022 16:35:42 CEST
    Started running in aws-beta#256 environment    Fri, 14 Oct 2022 16:38:42 CEST
    No longer running in aws-beta#257 environment  Sun, 16 Oct 2022 07:45:42 CEST
    Started running in aws-beta#259 environment    Sun, 16 Oct 2022 07:49:42 CEST
    No longer running in aws-beta#260 environment  Wed, 19 Oct 2022 09:28:42 CEST
    Started running in aws-beta#262 environment    Wed, 19 Oct 2022 09:32:42 CEST
    No longer running in aws-beta#263 environment  Wed, 19 Oct 2022 09:42:42 CEST
    Started running in aws-beta#265 environment    Wed, 19 Oct 2022 09:46:42 CEST
    No longer running in aws-prod#257 environment  Fri, 21 Oct 2022 11:02:28 CEST
    Started running in aws-prod#259 environment    Fri, 21 Oct 2022 11:05:28 CEST
```

## Search for an environment

As is the case for flows and artifacts, you can list all the Kosli environments you created under your organization

```shell {.command}
kosli list environments
```

```
NAME      TYPE  LAST REPORT                LAST MODIFIED
aws-beta  ECS   2022-10-30T14:51:42+01:00  2022-10-30T14:51:42+01:00
aws-prod  ECS   2022-10-30T14:51:28+01:00  2022-10-30T14:51:28+01:00
beta      K8S   2022-06-15T11:39:59+02:00  2022-06-15T11:39:59+02:00
prod      K8S   2022-06-15T11:40:01+02:00  2022-06-15T11:40:01+02:00
```

And get the metadata (including the type) of each environment:

```shell {.command}
kosli get environment aws-beta
```

```
Name:              aws-beta
Type:              ECS
Description:       The ECS beta namespace
State:             COMPLIANT
Last Reported At:  Sun, 30 Oct 2022 14:55:42 CET • 5 seconds ago
```

## Get environment events

When you have the name of the environment you want to dig into use `kosli list snapshots` or `kosli log environment` to browse snapshots and changes in the environment, or `kosli get snapshot` to have a look at a specific snapshot.

```shell {.command}
kosli list snapshots aws-beta
```

```
SNAPSHOT  FROM                            TO                              DURATION
266       Wed, 19 Oct 2022 09:47:42 CEST  now                             11 days
265       Wed, 19 Oct 2022 09:46:42 CEST  Wed, 19 Oct 2022 09:47:42 CEST  59 seconds
264       Wed, 19 Oct 2022 09:45:42 CEST  Wed, 19 Oct 2022 09:46:42 CEST  about a minute
263       Wed, 19 Oct 2022 09:42:42 CEST  Wed, 19 Oct 2022 09:45:42 CEST  3 minutes
262       Wed, 19 Oct 2022 09:32:42 CEST  Wed, 19 Oct 2022 09:42:42 CEST  10 minutes
261       Wed, 19 Oct 2022 09:31:42 CEST  Wed, 19 Oct 2022 09:32:42 CEST  about a minute
260       Wed, 19 Oct 2022 09:28:42 CEST  Wed, 19 Oct 2022 09:31:42 CEST  3 minutes
259       Sun, 16 Oct 2022 07:49:42 CEST  Wed, 19 Oct 2022 09:28:42 CEST  3 days
258       Sun, 16 Oct 2022 07:48:42 CEST  Sun, 16 Oct 2022 07:49:42 CEST  59 seconds
257       Sun, 16 Oct 2022 07:45:42 CEST  Sun, 16 Oct 2022 07:48:42 CEST  3 minutes
256       Fri, 14 Oct 2022 16:38:42 CEST  Sun, 16 Oct 2022 07:45:42 CEST  2 days
255       Fri, 14 Oct 2022 16:37:42 CEST  Fri, 14 Oct 2022 16:38:42 CEST  about a minute
254       Fri, 14 Oct 2022 16:35:42 CEST  Fri, 14 Oct 2022 16:37:42 CEST  2 minutes
253       Thu, 13 Oct 2022 09:04:42 CEST  Fri, 14 Oct 2022 16:35:42 CEST  one day
252       Mon, 10 Oct 2022 08:47:42 CEST  Thu, 13 Oct 2022 09:04:42 CEST  3 days
```

By default, you can see the last 15 changes to the environment. You can choose to only print e.g. last 3 events (`-n` flag).

You can also choose to see the actual events from each snapshot, using `kosli log environment` command:

```shell {.command}
kosli log environment aws-beta
```

```
SNAPSHOT  EVENT                                                                          FLOW       DEPLOYMENTS
#266      Artifact: 244531986313.dkr.ecr.eu-central-1.amazonaws.com/dashboard:d90a3e4    dashboard  #15 
          Fingerprint: dd5308fdcda117c1ff3963e192a069ae390c2fe9e10e8abfa2430224265efe98             
          Description: 1 instance started running (from 0 to 1)                                     
          Reported at: Wed, 19 Oct 2022 09:47:42 CEST                                               
                                                                                                    
#265      Artifact: 244531986313.dkr.ecr.eu-central-1.amazonaws.com/web:7ac7cdc          web        #63 
          Fingerprint: 88c082eee192653ea5826d14f714bcfbdadbd1827a7a29416bfddbdff2b69507             
          Description: 3 instances started running (from 0 to 3)                                    
          Reported at: Wed, 19 Oct 2022 09:46:42 CEST                                               
                                                                                                    
#265      Artifact: 244531986313.dkr.ecr.eu-central-1.amazonaws.com/runner:2872115       runner     #24 
          Fingerprint: 9461946e43393404ce744292331e7efbfe4e17cc2e5a32972169a90c81ec875c             
          Description: 3 instances started running (from 0 to 3)                                    
          Reported at: Wed, 19 Oct 2022 09:46:42 CEST  
```

You can also use an *interval* expression, like `262..264` (to see specified snapshot list)

```shell {.command}
kosli log environment aws-beta 262..264
```

```
SNAPSHOT  FROM                            TO                              DURATION
264       Wed, 19 Oct 2022 09:45:42 CEST  Wed, 19 Oct 2022 09:46:42 CEST  about a minute
263       Wed, 19 Oct 2022 09:42:42 CEST  Wed, 19 Oct 2022 09:45:42 CEST  3 minutes
262       Wed, 19 Oct 2022 09:32:42 CEST  Wed, 19 Oct 2022 09:42:42 CEST  10 minutes
```

or `~4..NOW` (to get a list of snapshots starting from 4 behind a currently running one and the current one)

```shell {.command}
kosli log environment aws-beta ~4..NOW
```

```
SNAPSHOT  FROM                            TO                              DURATION
266       Wed, 19 Oct 2022 09:47:42 CEST  now                             11 days
265       Wed, 19 Oct 2022 09:46:42 CEST  Wed, 19 Oct 2022 09:47:42 CEST  59 seconds
264       Wed, 19 Oct 2022 09:45:42 CEST  Wed, 19 Oct 2022 09:46:42 CEST  about a minute
263       Wed, 19 Oct 2022 09:42:42 CEST  Wed, 19 Oct 2022 09:45:42 CEST  3 minutes
262       Wed, 19 Oct 2022 09:32:42 CEST  Wed, 19 Oct 2022 09:42:42 CEST  10 minutes
```

## Get a snapshot 

To have a look at what is or was running in a given snapshot use `kosli get snapshot` command. You can use just the environment name as the argument, which will give you the latest snapshot, add `#` and snapshot number, to get a specific one, or `~n` where *n* is a number, to get *n-th* snapshot behind a current one:

``` shell {.command}
kosli get snapshot aws-beta
```

```
COMMIT   ARTIFACT                                                                              FLOW      RUNNING_SINCE  REPLICAS
d90a3e4  Name: 244531986313.dkr.ecr.eu-central-1.amazonaws.com/dashboard:d90a3e4               N/A       11 days ago    1
         Fingerprint: dd5308fdcda117c1ff3963e192a069ae390c2fe9e10e8abfa2430224265efe98                                  
                                                                                                                        
9f669e5  Name: 244531986313.dkr.ecr.eu-central-1.amazonaws.com/languages-start-points:9f669e5  N/A       11 days ago    1
         Fingerprint: e6b72f6a41d0944824538334120804ccde795b4b5aeb8aa311dbc0721b4e40fd                                  
                                                                                                                        
1c162e4  Name: 244531986313.dkr.ecr.eu-central-1.amazonaws.com/differ:1c162e4                  N/A       11 days ago    1
         Fingerprint: b7fd766dd2514b2610c0c8d70d8f762de4921931f97fdd6fbbfcc9745ac3ce3b                                  
[...]
```

```shell {.command}
kosli get snapshot aws-beta#256
```

```
COMMIT   ARTIFACT                                                                              FLOW      RUNNING_SINCE  REPLICAS
6fe0d30  Name: 244531986313.dkr.ecr.eu-central-1.amazonaws.com/repler:6fe0d30                  N/A       16 days ago    1
         Fingerprint: a0c03099c832e4ce5f23f5e33dac9889c0b7ccd61297fffdaf1c67e7b99e6f8f                                  
                                                                                                                        
d90a3e4  Name: 244531986313.dkr.ecr.eu-central-1.amazonaws.com/dashboard:d90a3e4               N/A       16 days ago    1
         Fingerprint: dd5308fdcda117c1ff3963e192a069ae390c2fe9e10e8abfa2430224265efe98                                  
                                                                                                                        
1c162e4  Name: 244531986313.dkr.ecr.eu-central-1.amazonaws.com/differ:1c162e4                  N/A       16 days ago    1
         Fingerprint: b7fd766dd2514b2610c0c8d70d8f762de4921931f97fdd6fbbfcc9745ac3ce3b                                  
[...]
```

```shell {.command}
kosli get snapshot aws-beta~19
```

```
COMMIT   ARTIFACT                                                                              FLOW      RUNNING_SINCE  REPLICAS
2e8646c  Name: 244531986313.dkr.ecr.eu-central-1.amazonaws.com/shas:2e8646c                    N/A       one month ago  1
         Fingerprint: a3158c3e79c83905fd3613e06b8cf5a45141c50cf49d4f99de90a2d081b77771                                  
                                                                                                                        
344430d  Name: 244531986313.dkr.ecr.eu-central-1.amazonaws.com/creator:344430d                 N/A       2 months ago   1
         Fingerprint: 817a72609041c51cd2a3bbbcbeb048c687677986b5a273399c6938b5e6aa1ded                                  
                                                                                                                        
7ac7cdc  Name: 244531986313.dkr.ecr.eu-central-1.amazonaws.com/web:7ac7cdc                     N/A       2 months ago   3
         Fingerprint: 88c082eee192653ea5826d14f714bcfbdadbd1827a7a29416bfddbdff2b69507                                 

```

The same expressions (with `#` and `~`) can be used to reference snapshots when diffing environment.

In the example below there was only one difference between snapshots: one new artifact started running in the latest snapshot. 

```shell {.command}
kosli diff snapshots aws-beta aws-beta~1
```

```
Only present in aws-beta (snapshot: aws-beta#266)
                   
     Name:         244531986313.dkr.ecr.eu-central-1.amazonaws.com/dashboard:d90a3e4
     Fingerprint:  dd5308fdcda117c1ff3963e192a069ae390c2fe9e10e8abfa2430224265efe98
     Flow:         dashboard
     Commit URL:   https://github.com/cyber-dojo/dashboard/commit/d90a3e481d57023816f6694ba4252342889405eb
     Started:      Wed, 19 Oct 2022 09:47:33 CEST • 11 days ago
```

## Diff environments/snapshots

You can use `diff` to compare snapshots of two different environments or different snapshots of the same environment:

```shell {.command}
kosli diff snapshots aws-beta~3 aws-prod
```

```
Only present in aws-prod (snapshot: aws-prod#261)
                   
     Name:         274425519734.dkr.ecr.eu-central-1.amazonaws.com/saver:8d724a1
     Fingerprint:  3e52f9b838cbb4e31455524c908eb8dd878b2ae25144427de8160f6658ee191f
     Flow:         saver
     Commit URL:   https://github.com/cyber-dojo/saver/commit/8d724a14c6e95947f0c56ad6af8251bca656a599
     Started:      Fri, 21 Oct 2022 11:04:59 CEST • 9 days ago
                   
     Name:         274425519734.dkr.ecr.eu-central-1.amazonaws.com/nginx:d491f5c
     Fingerprint:  4f66ab1b0a7a9f7ed064a3b1033a53ec7dd99359ff68d509ab555dcf4516b23e
     Flow:         nginx
     Commit URL:   https://github.com/cyber-dojo/nginx/commit/d491f5c06babe70bfebe2f9df0a9a66db7957f17
     Started:      Fri, 21 Oct 2022 11:03:53 CEST • 9 days ago
                   
     Name:         274425519734.dkr.ecr.eu-central-1.amazonaws.com/custom-start-points:8c551d3
     Fingerprint:  76ad6ffc1828d8213a39bc39c879b3c35a75d4705d1d8df5977a87a11e6ae25e
     Flow:         custom-start-points
     Commit URL:   https://github.com/cyber-dojo/custom-start-points/commit/8c551d378051b6ef1fde7fd58aaced1047264405
     Started:      Fri, 21 Oct 2022 11:04:30 CEST • 9 days ago
[...]
```

## report_aws_envs.md
---
title: "How to report ECS, Lambda and S3 environments"
bookCollapseSection: false
weight: 509
summary: "Kosli environments allow you to track changes in your physical/virtual runtime environments. Such changes must be reported from the runtime environment to Kosli. This tutorial shows you how to set up reporting of running artifacts from a Kubernetes cluster to Kosli."
---

# How to report ECS, Lambda and S3 environments

Kosli environments allow you to track changes in your physical/virtual runtime environments. Such changes must be reported from the runtime environment to Kosli.

This tutorial shows you how to set up reporting of running artifacts from a Kubernetes cluster to Kosli.


## Different ways for reporting

There are two different ways to report what's running in a Kubernetes cluster:

- Using Kosli CLI (suitable for testing only)
- Using the [Kosli terraform module](https://registry.terraform.io/modules/kosli-dev/kosli-reporter/aws/latest) to setup a Lambda function to be triggered on AWS changes and report to Kosli.

We describe how to use the different options below and you can choose what suites your needs.

## Prerequisites

To follow this tutorial, you will need to:

- Have access to AWS.
- [Create a Kosli account](https://app.kosli.com/sign-up) if you have not got one already.
- [Create an ECS, Lambda or S3 Kosli environment](/getting_started/environments/#create-an-environment) named `aws-env-tutorial` 
- [Get a Kosli API token](/getting_started/service-accounts/)
- [Install Kosli CLI](/getting_started/install/) (only needed if you will report using CLI)
- [Install Terraform](https://developer.hashicorp.com/terraform/install) (only needed if you will use the Kosli terraform module)

## Report snapshots using Kosli CLI

This option is **only suitable for testing purposes**.  
You need to create an AWS static credentials or equivalent and export the following environments variables:

```shell {.command}
export AWS_REGION=yourAWSRegion
export AWS_ACCESS_KEY_ID=yourAWSAccessKeyID
export AWS_SECRET_ACCESS_KEY=yourAWSSecretAccessKey
```

{{< tabs "snapshot env" "col-no-wrap" >}}

{{< tab "ECS" >}}
```shell {.command}
kosli snapshot ecs aws-env-tutorial \
    --cluster <your-ecs-cluster-name> \
	--api-token <your-api-token-here> \
    --org <your-kosli-org-name>
```
{{< /tab >}}

{{< tab "Lambda" >}}
```shell {.command}
kosli snapshot lambda aws-env-tutorial \
    --function-names function1,function2 \
	--api-token <your-api-token-here> \
    --org <your-kosli-org-name>
```
{{< /tab >}}

{{< tab "S3" >}}
```shell {.command}
kosli snapshot s3 aws-env-tutorial \
    --bucket <your-bucket-name> \
	--api-token <your-api-token-here> \
    --org <your-kosli-org-name>
```
{{< /tab >}}

{{< /tabs >}}


## Report snapshots using Terraform module

You can use the Kosli reporter terraform module to setup a Lambda function which is triggered every time your ECS cluster, Lambda function(s) or S3 bucket changes. The Lambda function will report the running artifacts to Kosli by running the Kosli CLI.

To setup the Lambda function using terraform, you need to follow these steps:

1. [Authenticate to AWS](https://docs.aws.amazon.com/cli/latest/userguide/cli-chap-configure.html)
   
2. Store the Kosli API key value in the parameter store in AWS Systems Manager (SSM) (use SecureString type). By default, the Lambda Reporter function will search for the `kosli_api_token` SSM parameter, but it is also possible to set custom parameter name using `kosli_api_token_ssm_parameter_name` terraform variable.
   
3. Create a Terraform configuration by copying one of the examples below into a `main.tf` file.

{{< tabs "terraform aws env" "col-no-wrap" >}}

{{< tab "ECS" >}}
```hcl {.command}
terraform {
  required_version = ">= 1.0.0"

  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = ">= 4.63"
    }
    random = {
      source  = "hashicorp/random"
      version = ">= 3.5.1"
    }
  }
}

provider "aws" {
  region = local.region

  # Make it faster by skipping some checks
  skip_metadata_api_check     = true
  skip_region_validation      = true
  skip_credentials_validation = true
  skip_requesting_account_id  = true
}

locals {
  reporter_name = "reporter-${random_pet.this.id}"
  region        = "eu-central-1"
}

data "aws_caller_identity" "current" {}

data "aws_canonical_user_id" "current" {}

resource "random_pet" "this" {
  length = 2
}

module "lambda_reporter" {
  source  = "kosli-dev/kosli-reporter/aws"
  version = "0.5.7"

  name                              = local.reporter_name
  kosli_environment_type            = "ecs"
  kosli_cli_version                 = "v2.11.0"
  kosli_environment_name            = "aws-env-tutorial"
  kosli_org                         = "<your-org-name>"
  reported_aws_resource_name        = "<your-ecs-cluster-name>"
}
```
{{< /tab >}}


{{< tab "Lambda" >}}
```hcl {.command}
terraform {
  required_version = ">= 1.0.0"

  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = ">= 4.63"
    }
    random = {
      source  = "hashicorp/random"
      version = ">= 3.5.1"
    }
  }
}

provider "aws" {
  region = local.region

  # Make it faster by skipping some checks
  skip_metadata_api_check     = true
  skip_region_validation      = true
  skip_credentials_validation = true
  skip_requesting_account_id  = true
}

locals {
  reporter_name = "reporter-${random_pet.this.id}"
  region        = "eu-central-1"
}

data "aws_caller_identity" "current" {}

data "aws_canonical_user_id" "current" {}

resource "random_pet" "this" {
  length = 2
}

variable "my_lambda_functions" {
  type    = string
  default = "function_name1, function_name2"
}

module "lambda_reporter" {
  source  = "kosli-dev/kosli-reporter/aws"
  version = "0.5.7"

  name                           = local.reporter_name
  kosli_environment_type         = "lambda"
  kosli_cli_version              = "v2.11.0"
  kosli_environment_name         = "aws-env-tutorial"
  kosli_org                      = "<your-org-name>"
  reported_aws_resource_name     = var.my_lambda_functions
  use_custom_eventbridge_pattern = true
  custom_eventbridge_pattern     = local.custom_event_pattern
}

locals {
  lambda_function_names_list = split(",", var.my_lambda_functions)

  custom_event_pattern = jsonencode({
    source      = ["aws.lambda"]
    detail-type = ["AWS API Call via CloudTrail"]
    detail = {
      requestParameters = {
        functionName = local.lambda_function_names_list
      }
      responseElements = {
        functionName = local.lambda_function_names_list
      }
    }
  })
}
```
{{< /tab >}}

{{< tab "S3" >}}
```hcl {.command}
terraform {
  required_version = ">= 1.0.0"

  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = ">= 4.63"
    }
    random = {
      source  = "hashicorp/random"
      version = ">= 3.5.1"
    }
  }
}

provider "aws" {
  region = local.region

  # Make it faster by skipping some checks
  skip_metadata_api_check     = true
  skip_region_validation      = true
  skip_credentials_validation = true
  skip_requesting_account_id  = true
}

locals {
  reporter_name = "reporter-${random_pet.this.id}"
  region        = "eu-central-1"
}

data "aws_caller_identity" "current" {}

data "aws_canonical_user_id" "current" {}

resource "random_pet" "this" {
  length = 2
}

variable "my_lambda_functions" {
  type    = string
  default = "my_lambda_function1, my_lambda_function_name2"
}

module "lambda_reporter" {
  source  = "kosli-dev/kosli-reporter/aws"
  version = "0.5.7"

  name                       = local.reporter_name
  kosli_environment_type     = "s3"
  kosli_cli_version          = "v2.11.0"
  kosli_environment_name     = "aws-env-tutorial"
  kosli_org                  = "<your-org-name>"
  reported_aws_resource_name = "<your-s3-bucket-name>"
}
```
{{< /tab >}}

{{< /tabs >}}

4. Initialize and run Terraform by running:

```shell {.command}
terraform init
terraform apply
```

5. To check Lambda reporter logs you can go to the AWS console -> Lambda service -> choose your lambda reporter function -> Monitor tab -> Logs tab.

## report_k8s_envs.md
---
title: "How to report Kubernetes Clusters"
bookCollapseSection: false
weight: 508
summary: "Kosli environments allow you to track changes in your physical/virtual runtime environments. Such changes must be reported from the runtime environment to Kosli. This tutorial shows you how to set up reporting of running artifacts from a Kubernetes cluster to Kosli."
---

# How to report Kubernetes Clusters to Kosli

Kosli environments allow you to track changes in your physical/virtual runtime environments. Such changes must be reported from the runtime environment to Kosli.

This tutorial shows you how to set up reporting of running artifacts from a Kubernetes cluster to Kosli.


## Different ways for reporting

There are 3 different ways to report what's running in a Kubernetes cluster:

- Using Kosli CLI (suitable for testing only)
- Using a Kubernetes cronjob configured with a helm chart (recommended for production use).
- Using an externally scheduled cron process (e.g. a scheduled CI workflow)

We describe how to use the different options below and you can choose what suites your needs.

## Prerequisites

To follow this tutorial, you will need to:

- Have access to a Kubernetes cluster.
- [Create a Kosli account](https://app.kosli.com/sign-up) if you have not got one already.
- [Create a Kubernetes Kosli environment](/getting_started/environments/#create-an-environment) named `k8s-tutorial` 
- [Get a Kosli API token](/getting_started/service-accounts/)
- [Install Kosli CLI](/getting_started/install/) (only needed if you will report using CLI)
- [Install Helm](https://helm.sh/docs/intro/install/) (only needed if you will use the Kosli helm chart)

## Report snapshots using Kosli CLI

This option is **only suitable for testing purposes**. 

> All the commands below will use the default `kubecontext` in "$HOME/.kube/config". You can change it with `--kubeconfig` 

To report the **artifacts running in an entire cluster**, you can run the following command:

```shell {.command}
kosli snapshot k8s k8s-tutorial \
    --api-token <your-api-token-here> \
    --org <your-kosli-org-name>
```

To report **artifacts running in one or more namespaces**, you can run the following command:

```shell {.command}
kosli snapshot k8s k8s-tutorial \
    --namespaces namespace1,namespace2 \
    --api-token <your-api-token-here> \
    --org <your-kosli-org-name>
```

To report **artifacts running in the entire cluster except from some namespaces**, you can run the following command:

```shell {.command}
kosli snapshot k8s k8s-tutorial \
    --exclude-namespaces namespace1,namespace2 \
    --api-token <your-api-token-here> \
    --org <your-kosli-org-name>
```

## Report snapshots using the Kosli K8S reporter helm chart

The recommended way to regularly report artifacts running in a cluster to Kosli is to use the [K8S reporter helm chart](/helm).

The chart creates a cronjob that will run the Kosli CLI inside a pod to report the artifacts running in the cluster.

1. Create a K8S secret to contain your Kosli API token.

```shell {.command}
kubectl create secret generic kosli-api-token --from-literal=apikey=<your-kosli-api-token>
```

> Make sure the secret value does not contain any trailing whitespace.

2. Prepare the settings for the helm chart

To customize how the helm chart creates the cronjob, you can create your own values file by copying and modifying the [default values file](https://github.com/kosli-dev/cli/blob/main/charts/k8s-reporter/values.yaml).

We will use this file (named `tutorial-values.yaml`):

```yaml {.command}
# -- the cron schedule at which the reporter is triggered to report to kosli  
cronSchedule: "*/5 * * * *"

kosliApiToken:
  # -- the name of the secret containing the kosli API token
  secretName: "kosli-api-token"
  # -- the name of the key in the secret data which contains the kosli API token
  secretKey: "apikey"

reporterConfig:
  # -- the name of the kosli org
  kosliOrg: "<your-kosli-org-name>"
  # -- the name of kosli environment that the k8s cluster/namespace correlates to
  kosliEnvironmentName: "k8s-tutorial"
  # -- the namespaces which represent the environment.
  # It is a comma separated list of namespace name regex patterns.
  # e.g. `^prod$,^dev-*` reports for the `prod` namespace and any namespace that starts with `dev-`
  # leave this unset if you want to report what is running in the entire cluster
  namespaces: ""
```

3. Install the Kosli helm chart

```shell {.command}
helm repo add kosli https://charts.kosli.com/
helm repo update
helm install kosli-reporter kosli/k8s-reporter -f tutorial-values.yaml
```

4. Confirm the cronjob is created in the cluster:

```shell {.command}
kubectl get cronjobs
```

Now, the cronjob will run every 5 minutes and report what is running in the entire cluster to Kosli.


## Report snapshots using externally scheduled cronjobs

If you do not wish to run the Kosli reporter inside the cluster, you can run it from outside the cluster. This requires opening access to the cluster from the place you will run the CLI regularly. 

One option to send reports regularly from outside the cluster is to use Github Actions scheduled workflows. Here is an example workflow definition:

> Note that the workflow below needs secrets to be added in Github actions.

```yaml {.command}
name: Regular Kubernetes reports to Kosli

on:
  workflow_dispatch: 
  schedule: 
    - cron: '0 * * * *' # every one hour

jobs:
  k8s-report:
    runs-on: ubuntu-latest
    permissions:
      id-token: write
      contents: write
    env:
      KOSLI_API_TOKEN: ${{ secrets.MY_KOSLI_API_TOKEN }}

    steps:
      - name: install kosli
        uses: kosli-dev/setup-cli-action@v2
      
      # connect to your cluster
      # if not using GKE, replace this step with one that connects to your cluster
      - name: Connect to GKE
        uses: 'Swibi/connect-to-gke'
        with:
          GCP_SA_KEY: ${{ secrets.GKE_SA_KEY }}
          GCP_PROJECT_ID: ${{ secrets.GKE_PROJECT }}
          GKE_CLUSTER: <your-cluster-name>
          GKE_ZONE: <your-cluster-zone>
      
      - name: Scan artifacts and send K8S report to Kosli
        run: 
          kosli snapshot k8s k8s-tutorial --org <your-kosli-org-name>
          
      # send slack notifications on failure to report
      - name: Slack Notification on Failure
        if: ${{ failure() }}
        uses: rtCamp/action-slack-notify@v2
        env:
          SLACK_CHANNEL: kosli-reports-failure
          SLACK_COLOR: ${{ job.status }}
          SLACK_TITLE: Reporting K8S artifacts to Kosli has failed
          SLACK_USERNAME: GithubActions
          SLACK_WEBHOOK: ${{ secrets.SLACK_CI_FAILURES_WEBHOOK }}
          SLACK_MESSAGE: "Reporting K8S artifacts to Kosli has failed. Please check the logs for more details."
```


## tracing_a_production_incident_back_to_git_commits.md
---
title: Tracing a production incident back to git commits
bookCollapseSection: false
weight: 520
draft: false
summary: "In this 5 minute tutorial you'll learn how Kosli can track a production incident in Cyber-dojo back to git commits."
---

<!-- Add Easter-eggs comments? -->

# Tracing a production incident back to git commits

In this 5 minute tutorial you'll learn how Kosli can track a production incident in Cyber-dojo back to git commits.

Something has gone wrong and [https://cyber-dojo.org](https://cyber-dojo.org) is displaying a 500 error!


{{< figure src="/images/cyber-dojo-prod-500-large.png" alt="Prod cyber-dojo is down with a 500" width="90%" >}}

It was working an hour ago. What has happened in the last hour?

## Getting ready

You need to:
* [Install Kosli CLI](/getting_started/install/).
* [Get a Kosli API token](/getting_started/service-accounts/).
* Set the `KOSLI_ORG` environment variable to `cyber-dojo` (the Kosli `cyber-dojo` organization is public so any authenticated user can read its data) and `KOSLI_API_TOKEN` to your token:
  ```shell {.command}
  export KOSLI_ORG=cyber-dojo
  export KOSLI_API_TOKEN=<your-api-token>
  ```

## Start with the environment

[https://cyber-dojo.org](https://cyber-dojo.org) is running in an AWS environment
that reports to Kosli as `aws-prod`.  
Get a log of this environment's changes:

```shell {.command}
kosli log env aws-prod
```

At the time this tutorial was written the output of this command
displayed the first page of 177 snapshots. 
You will see the first page of considerably more than 177 snapshots because 
`aws-prod` has moved on since this incident (it has been resolved with new 
commits which have created new deployments). 
To limit the output you can set the interval for the command:

```shell {.command}
kosli log env aws-prod --interval 176..177
```

The output should be:

```plaintext {.light-console}
SNAPSHOT  EVENT                                                                          FLOW      DEPLOYMENTS
#177      Artifact: 274425519734.dkr.ecr.eu-central-1.amazonaws.com/creator:31dee35      creator   #87 
          Fingerprint: 5d1c926530213dadd5c9fcbf59c8822da56e32a04b0f9c774d7cdde3cf6ba66d             
          Description: 1 instance stopped running (from 1 to 0).                               
          Reported at: Tue, 06 Sep 2022 16:53:28 CEST                                          
                                                                                               
#176      Artifact: 274425519734.dkr.ecr.eu-central-1.amazonaws.com/creator:b7a5908      creator   #89 
          Fingerprint: 860ad172ace5aee03e6a1e3492a88b3315ecac2a899d4f159f43ca7314290d5a             
          Description: 1 instance started running (from 0 to 1).                               
          Reported at: Tue, 06 Sep 2022 16:52:28 CEST
```

These two snapshots belong to the same blue-green deployment.
You see artifact `creator:b7a5908` starting in snapshot #176, and artifact
`creator:31dee35` exiting in snapshot #177.

## Dig into the artifact

You are interested in #176, showing the newly running artifact, `creator:b7a5908`,
with the fingerprint starting `860ad17`.

Let's learn more about this artifact:

```shell {.command}
kosli get artifact creator@860ad17
```

```plaintext {.light-console}
Name:        cyberdojo/creator:b7a5908
Flow:        creator
Fingerprint: 860ad172ace5aee03e6a1e3492a88b3315ecac2a899d4f159f43ca7314290d5a
Created on:  Tue, 06 Sep 2022 16:48:07 CEST • 21 hours ago
Git commit:  b7a590836cf140e17da3f01eadd5eca17d9efc65
Commit URL:  https://github.com/cyber-dojo/creator/commit/b7a590836cf140e17da3f01eadd5eca17d9efc65
Build URL:   https://github.com/cyber-dojo/creator/actions/runs/3001102984
State:       COMPLIANT
History:  
    Artifact created                               Tue, 06 Sep 2022 16:48:07 CEST
    Deployment #88 to aws-beta environment         Tue, 06 Sep 2022 16:49:59 CEST
    Deployment #89 to aws-prod environment         Tue, 06 Sep 2022 16:51:12 CEST
    Started running in aws-beta#196 environment    Tue, 06 Sep 2022 16:51:42 CEST
    Started running in aws-prod#176 environment    Tue, 06 Sep 2022 16:52:28 CEST
```

## Follow to the commit

You can follow the [commit URL](https://github.com/cyber-dojo/creator/commit/b7a590836cf140e17da3f01eadd5eca17d9efc65).

{{< figure src="/images/cyber-dojo-github-diff.png" alt="cyber-dojo github diff" width="500" >}}

The incident was caused by a simple typo in the `app.rb` file!

Perhaps someone accidentally inserted the "s" while trying to save the file?
Either way, this is clearly the problem because the function is called `respond_to` without the `s`.

You were able to trace the problem back to a specific commit without any access to cyber-dojo's `aws-prod` environment.

<!-- 
This we would like to show the users:
- Kosli gives developers without access to production environment information about what is running.
- Detect that a new "bit-coin miner" is running in your environment. Rogue artifact detection.
- Kosli can show that a deployment is reported, but artifact didn't start. Find this in artifact view.
- Kosli can show that an artifact started, but no deployment was reported for it.
- Detect an artifact that is missing evidence is running in an environment
- Do we want to mention the whole env being compliant?
- Commit makes the server stop working. Use kosli env diff to find out what artifact changed.
It would be good if we had two versions of env where there are several artifacts that change.
(with easter egg)

(- Find out when/where a given commit is running.)

- See what software is/was running where which is useful in debugging.
  I detect from the web page that there is something wrong with 'saver'. I then want to know
  which version of 'saver' is running now. I want to know what git commit is running.
- List which version of 'saver' is running across all environments.

- We see that beta.cyberdojo.org is not working as expected, but prod is still OK. We do a kosli env diff and
  kosli env log to find out what services has changed.

- Change of K8S infrastructure broke both cyber dojo environments. The fix was to manually change 3 of the
  services on prod. Beta was not fixed and was down for a long period. We might not be able to detect this.

Problems:
- Not every commit generates an artifact. If you only build after 10 commits then 9 will not
be visible.

Things we can do later:
- Find which artifact this "unknown commit" is part of. So we need the git history.
- Kosli can show that an older deployment is running than that is declared. roll-back

 -->

## unauthorized_iac_changes.md
---
title: "Detecting unauthorized Terraform IaC changes"
bookCollapseSection: false
weight: 506
summary: "Authorized Terraform changes follow a predefined process that maintains a certain level of quality, security and safety for the underlying infrastructure. Unauthorized changes, however, can undermine the integrity and reliability of the infrastructure. Hence the importance of prompt detection of such changes."
---

# Detecting unauthorized Terraform IaC changes

Authorized Terraform changes follow a predefined process that maintains a certain level of quality, security and safety for the underlying infrastructure. Unauthorized changes, however, can undermine the integrity and reliability of the infrastructure. Hence the importance of prompt detection of such changes.

Unauthorized Terraform changes happen in one of two ways:
1. Bypassing Terraform and making direct changes via cloud APIs, clients or UI consoles. This leads to drift, where the desired state does not match the actual state. This kind of unauthorized change can be detected and corrected with [Terraform drift detection](https://developer.hashicorp.com/terraform/tutorials/state/resource-drift).
2. Bypassing the predefined process for Terraform changes. For example, a developer running terraform directly from their machine without going through CI.

This tutorial shows how you can use Kosli to track and detect the second type of unauthorized changes.

## Prerequisites

To follow the steps in this tutorial, you need to:
* [Install Terraform on your machine](https://developer.hashicorp.com/terraform/install).
* (Optional)[Setup Snyk on your machine](https://docs.snyk.io/snyk-cli/getting-started-with-the-snyk-cli#install-the-snyk-cli-and-authenticate-your-machine).
* [Create a Kosli account](https://app.kosli.com/) (Skip if you already have one).
* [Install Kosli CLI](/getting_started/install/).
* [Get a Kosli API token](/getting_started/service-accounts/).
* Set the `KOSLI_ORG` environment variable to your personal org name and `KOSLI_API_TOKEN` to your token:
  ```shell {.command}
  export KOSLI_ORG=<your-personal-kosli-org-name>
  export KOSLI_API_TOKEN=<your-api-token>
  ```
* Clone the tutorial git repo
  ```shell {.command}
  git clone https://github.com/kosli-dev/iac-changes-tutorial.git 
  cd iac-changes-tutorial
  ```

## Creating a Kosli flow

We will start by creating a Kosli flow to represent the process for authorized Terraform changes.
For simplicity, we will not define any requirements for this process by using `--use-empty-template`

```shell {.command}
kosli create flow tf-tutorial --use-empty-template
```

## Making and tracking an authorized change

{{<hint info>}}
In production, an authorized change will normally go though CI.
In this tutorial, however, we run the commands that you would otherwise do in CI locally for simplicity.
{{</hint>}}

Let's create a trail to represent a single instance of making an authorized change. We will call it `authorized-1`.

```shell {.command}
kosli begin trail authorized-1 --flow=tf-tutorial
```
Next, we can scan our terraform config scripts for security issues. We capture the SARIF output from the scan and attest it to Kosli.

```shell {.command}
snyk iac test main.tf --sarif-file-output=sarif.json
kosli attest snyk --name=security --flow=tf-tutorial --trail=authorized-1 --scan-results=sarif.json
```

We are now ready to run terraform. We create a plan and save it to a file. Then attest the plan file to Kosli to build a historical audit log. 

```shell {.command}
terraform init
terraform plan -out=tf.plan
kosli attest generic --name=tf-plan --flow=tf-tutorial --trail=authorized-1 --attachments=tf.plan
```

Finally, we apply the terraform plan, and attest the produced terraform state file as an artifact.
This will calculate a SHA256 fingerprint for the state file based on its contents. The fingerprint will later be used to determine if a change is 
authorized or not.

{{<hint info>}}
In this tutorial, we use a simple setup where the terraform state file is stored locally.
In production cases, however, the state file would be stored in some cloud storage (e.g. AWS S3). 
In such cases, you would need to download the state file from the remote backend after it was updated by the authorized change.

Note that we set both `--build-url` and `--commit-url` to fake URLs. These are normally defaulted in CI.
{{</hint>}}

```shell {.command}
terraform apply -auto-approve tf.plan
kosli attest artifact terraform.tfstate --name=state-file --artifact-type=file --flow=tf-tutorial --trail=authorized-1 \
   --build-url=https://example.com --commit-url=https://example.com --commit=HEAD
```

## Monitoring the state file

Every time a change to the infrastructure happens via Terraform, the state file content would be changed. 
To detect when an **unauthorized** change happens, we can monitor the state file for changes and record those changes in
a Kosli environment.

Let's start by creating an environment of type `server`. 

```shell {.command}
kosli create env terraform-state --type=server
```

We can report the state file to the environment we created:

{{<hint info>}}
In this tutorial, we run the environment reporting manually. 
In production, you would configure the environment reporting to run periodically or on changes. 
See [reporting AWS environments](../report_aws_envs) if you are using S3 as a backend for your state files.
{{</hint>}}

```shell {.command}
kosli snapshot path terraform-state --name=tf-state --path=terraform.tfstate
```

You can get the latest snapshot of the environment by running:

```shell
kosli get snapshot terraform-state
COMMIT   ARTIFACT                                                                       FLOW         COMPLIANCE     RUNNING_SINCE  REPLICAS
d881b2f  Name: tf-state                                                                 tf-tutorial  NON-COMPLIANT  28 minutes ago   1
         Fingerprint: a57667a7b921b91d438631afa1a1fe35300b4da909a19d2b61196580f30f1d0c
```

Note that the `FLOW` column indicates that this artifact came from the `tf-tutorial` flow which means Kosli has provenance for 
where this change came from.

You can also view the environment status in the Kosli UI by navigating to: `Environments > terraform-state`.
At this point you should see one artifact with a compliant status since we have provenance for the change that happened.

{{< figure src="/images/tutorials/iac-changes/authorized-iac-change.png" alt="Environment shows an authorized change" width="90%" >}}

## Introducing an unauthorized change

Now let's see how Kosli can help catching an unauthorized change. 
We can simulate such change by modifying the `random_pet_result` output on line 6 in main.tf to `random_pet_name` and running:

```shell {.command}
terraform apply --auto-approve
```

This updates the state file. Let's report the updated state file to the Kosli environment.

{{<hint info>}}
In production, this step won't be necessary because you would have configured environment reporting to happen
automatically (either on state file change or periodically).
{{</hint>}}

```shell {.command}
kosli snapshot path terraform-state --name=tf-state --path=terraform.tfstate
```

Getting the latest snapshot of the environment by running the command below shows that the `FLOW` is unknown. 
This means that Kosli does not have provenance for that change (i.e. it is an unauthorized change).

```shell
kosli get snapshot terraform-state
COMMIT  ARTIFACT                                                                       FLOW  COMPLIANCE     RUNNING_SINCE   REPLICAS
N/A     Name: tf-state                                                                 N/A   NON-COMPLIANT  8 minutes ago  1
        Fingerprint: edd93dcde27718ed493222ceb218275655555f3f3bfefa95628c599e678ac325
```

When you navigate to the environment page again, you will see a non-compliant artifact running.

{{< figure src="/images/tutorials/iac-changes/unauthorized-iac-change.png" alt="Environment shows an unauthorized change" width="90%" >}}

## Next steps

Now that we can detect unauthorized changes in Terraform IaC, the next step would be to receive notifications or
trigger automated actions when this happens. You can achieve that by configuring [Kosli actions](/integrations/actions/).


## what_do_i_do_if_kosli_is_down.md
---
title: "What do I do if Kosli is down?"
bookCollapseSection: false
weight: 507
summary: "Customers use Kosli to attest evidence of their business and software processes. If Kosli is down, these attestations will fail. In this situation there is a built-in mechanism to instantly turn Kosli off and keep the pipelines flowing. When Kosli is back up, you can instantly turn Kosli back on."
---

# What do I do if Kosli is down?

Customers use Kosli to attest evidence of their business and software processes.
If Kosli is down, these attestations will fail.
This will break CI workflow pipelines, blocking artifacts from being deployed.
In this situation there is a built-in mechanism to instantly turn Kosli off and keep the pipelines flowing.
When Kosli is back up, you can instantly turn Kosli back on.

## Turning Kosli CLI calls on and off instantly

If the `KOSLI_DRY_RUN` environment variable is set to `true` then all Kosli CLI commands will:
* Not communicate with Kosli at all
* Print the payload they would have sent
* Exit with a zero status code

We recommend creating an Org-level KOSLI_DRY_RUN variable in your CI system and, in all CI workflows,
ensuring there is an environment variable set from it. 

For example, in a [Github Action workflow](https://github.com/cyber-dojo/differ/blob/main/.github/workflows/main.yml):

```yaml
name: Main
...
env:
  KOSLI_DRY_RUN: ${{ vars.KOSLI_DRY_RUN }}           # true iff Kosli is down
```


## Turning Kosli API calls on and off instantly

If you are using the Kosli API in your workflows (e.g. using `curl`), we recommend using the same Org-level `KOSLI_DRY_RUN` 
environment variable and guarding the `curl` call with a simple if statement. For example:

```shell
#!/usr/bin/env bash

kosli_curl()
{
  local URL="${1}"
  local JSON_PAYLOAD="${2}"

  if [ "${KOSLI_DRY_RUN:-}" == "true" ]; then
    echo KOSLI_DRY_RUN is set to true. This is the payload that would have been sent
    echo "${JSON_PAYLOAD}" | jq .
  else
    curl ... --data="${JSON_PAYLOAD}" "${URL}"
  fi
}
```






## _index.md
---
title: Understand Kosli
bookCollapseSection: true
weight: 100
---

## concepts.md
---
title: 'Concepts'
weight: 130
summary: "This section helps you understand the concepts Kosli is built on. The figure below gives an overview of the main Kosli concepts and how they are related to each other."
---

# Concepts

This section helps you understand the concepts Kosli is built on. The figure below gives an overview of the main Kosli concepts and how they are related to each other.

{{<figure src="/images/kosli_concepts.png" alt="Kosli Concepts" width="900">}}

## Organization

A Kosli organization is an account that owns Kosli resources, such as Flows and Environments. Only members within an organization can access its resources.

When signing up for Kosli, a personal organization is automatically created for you, bearing your username. This personal organization is exclusively accessible to you. Additionally, you can create `Shared` organizations and invite multiple team members to collaborate on different Flows and Environments.

## Flow

A Kosli Flow represents a business or software process for which you want to track changes and monitor compliance.

### Trail

A Kosli Trail represents a single execution instance of a process represented by a Kosli Flow.
Each Trail must have a unique identifier of your choice, based on your process and domain. Example identifiers include git commits or pull request numbers.
  
#### Artifact

Kosli Artifacts represent the software artifacts generated from every execution, portrayed as a Trail, of your software process depicted as a Flow. These artifacts play a crucial role in enabling **Binary Provenance**, providing a comprehensive chain of custody that records the origin, history and distribution of each artifact.

Each Artifact is uniquely identified by its SHA256 fingerprint. Using this fingerprint, Kosli can link the creation of the Artifact with its runtime-related events, such as when the artifact starts or concludes execution within a specific Environment.

#### Attestation

An Attestation is a record of compliance checks or controls that have been performed a particular Artifact or Trail. It is normally reported after performing a specific risk control or quality check (e.g. running tests). The attestation encompasses the procedure's results.

Kosli provides specific built-in types of attestations (e.g., a snyk scan, sonar scan, junit tests) and allows to define your own custom types. 

##### Evidence Vault

Attestations in Kosli have the capability to contain additional evidence files attached to them. This supporting evidence is securely stored within Kosli's evidence vault and is retrievable on demand.

## Audit package

During an audit process, Kosli enables you to download an audit package for a Trail, Artifact, or an individual Attestation. This package comprises a tar file containing metadata related to the selected resource, alongside any evidence files that have been attached. The audit package serves as a comprehensive collection of information aiding in audit-related investigations or reviews.

## Flow Template

A Flow Template defines the expected attestations for Flow Trails and Artifacts to be considered compliant. While each Flow has its own Template, each Trail in a Flow can override the Flow Template with its own.

## Environment

Environments in Kosli monitor changes in your software runtime systems.

Each physical or virtual runtime environment you want to track in Kosli should have its own Kosli Environment created. Kosli allows you to portray your environments precisely. For instance, with a Kubernetes cluster, you can treat it as one Kosli Environment or designate one or more namespaces in the cluster as separate Kosli Environments.

Kosli supports various types of runtime environments:
* Kubernetes cluster (K8S)
* Amazon ECS
* Amazon S3
* Amazon Lambda
* Physical/virtual server
* Docker host
* Azure Web Apps and Function Apps

### Environment Snapshot

An Environment Snapshot represents the reported status (running Artifacts) of your runtime environment at a specific point in time. Snapshots are immutable, append-only objects. Once a snapshot is created, it cannot be modified.

In each snapshot, Kosli links the running artifacts to the Flows and Trails that produced them. Snapshot compliance relies on the compliance status of each running artifact, while Environment compliance depends on its latest snapshot compliance.

Running artifacts that come from 3rd party sources, can be `allow-listed` in an Environment to make them compliant. 


### Environment Policy

Environment Policy enables you to define and enforce compliance requirements for artifact deployments across different environments.

## what_is_kosli.md
---
title: 'What is Kosli?'
weight: 120
summary: "Kosli is a change recording and compliance monitoring platform which allows you to record, track and query changes about any software or business process so you can prove compliance and maintain security without slowing down."
---
# What is Kosli?

Kosli is a change recording and compliance monitoring platform which allows you to record, track and query changes about any software or business process so you can prove compliance and maintain security without slowing down.

Kosli connects the recorded changes to establish immutable "chains of custody" which enables you to:

1. **Track Changes**: Trace how your business or software processes change over time.
2. **Identify Sources**: Understand where changes originated from, which can help in identifying issues.
3. **Continuous Compliance**: Ensure that you continuously adhere to your compliance requirements.
4. **Enable Audits**: Access audit packages on demand allowing audits and investigations into the software supply chain.
5. **Enhance Trust**: Build trust among users, customers, and stakeholders by providing transparent and verified information about the change history.
6. **Remove Friction**: Make changes at the DevOps speed with continuous compliance. No spreadsheets, paperwork or CAB meetings required.

# When to use Kosli?

Kosli serves as a versatile solution for a variety of use cases. Two primary scenarios which customers find Kosli valuable are:

- Monitoring and Ensuring Compliance in Software and Business Processes:
  Regardless of your specific process requirements and tools, Kosli empowers you to streamline and automate your change management and evidence recording processes. This capability ensures that organizations are consistently prepared for audits.

  Examples of software process requirements that Kosli can assist with are:
  - Verification that all artifacts running in production have undergone a specific risk control (e.g. security scanning).
  - Mandatory code review in a pull request before deployment to production.
  
  Examples of business process requirements that Kosli can address include:
  - Confirmation that employee onboarding/off-boarding aligns with relevant policies.
  - Logging and adherence to relevant policies for accessing production servers.
  
- Enhancing Overall Observability of Changes in Complex Systems and Environments:
  Beyond compliance considerations, Kosli remains valuable as a comprehensive platform, offering a unified view to different stakeholders. It provides visibility into changes occurring across various components of complex systems, even when there are no specific compliance requirements in place. This single pane of glass enhances overall observability.

Feel free to contact us with any additional questions or if you require further information regarding the capabilities of Kosli. Alternatively, [explore how our customers have successfully implemented and used our services](https://www.kosli.com/case-studies/).

# Where does Kosli fit in the growing tools landscape?

Kosli is tool-agnostic, specifically crafted to seamlessly integrate with various tools, including CI systems, code analysis tools, runtime environments, and more. Serving as a comprehensive compliance and change data hub, Kosli consolidates information from all your tools into a unified platform. It acts as your singular compliance and change management interface, providing a consolidated and streamlined view of data from diverse sources.

# How does Kosli work?

Kosli operates akin to a black box recorder, functioning as an append-only repository for immutable change records. Users report specific changes of interest through the command line interface (CLI) or API. Kosli, in turn, captures and stores these changes while actively monitoring compliance with designated policies.

Notably, change sources can originate from diverse environments, including build systems (such as CI systems) and runtime environments (for instance, a Kubernetes cluster). This flexibility ensures that Kosli effectively captures and monitors changes across various stages of development and deployment.

{{<figure src="/images/kosli-overview-docs.jpg" alt="Kosli overview" width="1000">}}



