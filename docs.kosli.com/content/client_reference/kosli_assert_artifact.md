---
title: "kosli assert artifact"
beta: false
deprecated: false
---

# kosli assert artifact

## Synopsis

Assert the compliance status of an artifact in Kosli.  
Exits with non-zero code if the artifact has a non-compliant status.

```shell
kosli assert artifact [IMAGE-NAME | FILE-PATH | DIR-PATH] [flags]
```

## Flags
| Flag | Description |
| :--- | :--- |
|    -t, --artifact-type string  |  The type of the artifact to calculate its SHA256 fingerprint. One of: [oci, docker, file, dir]. Only required if you want Kosli to calculate the fingerprint for you (i.e. when you don't specify '--fingerprint' on commands that allow it).  |
|    -D, --dry-run  |  [optional] Run in dry-run mode. When enabled, no data is sent to Kosli and the CLI exits with 0 exit code regardless of any errors.  |
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

In [this YAML file](https://app.kosli.com/api/v2/livedocs/cyber-dojo/yaml?ci=github&command=kosli+assert+artifact){{< /tab >}}{{< tab "GitLab" >}}View an example of the `kosli assert artifact` command in GitLab.

In [this YAML file](https://app.kosli.com/api/v2/livedocs/cyber-dojo/yaml?ci=gitlab&command=kosli+assert+artifact){{< /tab >}}{{< /tabs >}}

## Examples Use Cases

**fail if an artifact has a non-compliant status (using the artifact fingerprint)**

```shell
kosli assert artifact \
	--fingerprint 184c799cd551dd1d8d5c5f9a5d593b2e931f5e36122ee5c793c1d08a19839cc0 \
	--flow yourFlowName \
	--api-token yourAPIToken \
	--org yourOrgName 

```

**fail if an artifact has a non-compliant status (using the artifact name and type)**

```shell
kosli assert artifact library/nginx:1.21 \
	--artifact-type docker \
	--flow yourFlowName \
	--api-token yourAPIToken \
	--org yourOrgName
```

