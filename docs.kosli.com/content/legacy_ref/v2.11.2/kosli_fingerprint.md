---
title: "kosli fingerprint"
beta: false
deprecated: false
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
Excluded paths are relative to the DIR-PATH and can be literal paths or
glob patterns.  
The supported glob pattern syntax is what is documented here: https://pkg.go.dev/path/filepath#Match , 
plus the ability to use recursive globs "**"

To specify paths in a directory artifact that should always be excluded from the SHA256 calculation, you can add a `.kosli_ignore` file to the root of the artifact.
Each line should specify a relative path or path glob to be ignored. You can include comments in this file, using `#`.
The `.kosli_ignore` will be treated as part of the artifact like any other file,unless it is explicitly ignored itself.

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

In [this YAML file](https://app.kosli.com/api/v2/livedocs/cyber-dojo/yaml?ci=github&command=kosli+fingerprint), which created [this Kosli Event](https://app.kosli.com/api/v2/livedocs/cyber-dojo/event?ci=github&command=kosli+fingerprint).{{< /tab >}}{{< /tabs >}}

## Examples Use Cases

**fingerprint a file**

```shell
kosli fingerprint --artifact-type file file.txt

```

**fingerprint a dir**

```shell
kosli fingerprint --artifact-type dir mydir

```

**fingerprint a dir while excluding paths**

```shell
kosli fingerprint --artifact-type dir --exclude logs --exclude *.exe mydir

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

