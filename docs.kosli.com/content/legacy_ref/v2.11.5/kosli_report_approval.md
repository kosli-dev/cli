---
title: "kosli report approval"
beta: false
deprecated: false
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

