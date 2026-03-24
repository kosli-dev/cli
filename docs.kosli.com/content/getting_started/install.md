---
title: "Part 2: Install Kosli CLI"
bookCollapseSection: false
weight: 220
---
# Part 2: Install Kosli CLI

{{< tabs "installKosli" >}}

{{< tab "macOS / Linux (Homebrew)" >}}

The simplest way to install the Kosli CLI is via [Homebrew](https://brew.sh/):

```shell {.command}
brew install kosli-dev/tap/kosli
```

Once installed, verify it works:

```shell {.command}
kosli version
```


{{< /tab >}}


{{< tab "macOS (manual)" >}}

1. Download the latest binary from [GitHub releases](https://github.com/kosli-dev/cli/releases)

For Intel Macs:
```shell {.command}
curl -L https://github.com/kosli-dev/cli/releases/latest/download/kosli_latest_darwin_amd64.tar.gz | tar zx
sudo mv kosli /usr/local/bin/kosli
```

For Apple Silicon:
```shell {.command}
curl -L https://github.com/kosli-dev/cli/releases/latest/download/kosli_latest_darwin_arm64.tar.gz | tar zx
sudo mv kosli /usr/local/bin/kosli
```

{{< /tab >}}


{{< tab "Linux (manual)" >}}

1. Download the latest binary from [GitHub releases](https://github.com/kosli-dev/cli/releases)

For AMD64:
```shell {.command}
curl -L https://github.com/kosli-dev/cli/releases/latest/download/kosli_latest_linux_amd64.tar.gz | tar zx
sudo mv kosli /usr/local/bin/kosli
```

For ARM64:
```shell {.command}
curl -L https://github.com/kosli-dev/cli/releases/latest/download/kosli_latest_linux_arm64.tar.gz | tar zx
sudo mv kosli /usr/local/bin/kosli
```

{{< /tab >}}

{{< tab "Windows" >}}

Download the latest release from [GitHub releases](https://github.com/kosli-dev/cli/releases) and add the binary to your PATH.

You can also use [Scoop](https://scoop.sh/):

```shell {.command}
scoop bucket add kosli-dev https://github.com/kosli-dev/scoop-kosli.git
scoop install kosli
```

{{< /tab >}}

{{< tab "Docker" >}}
We offer the Kosli CLI as a Docker image. Grab the latest from [Docker Hub](https://hub.docker.com/r/kosli/cli).

```shell {.command}
docker run --rm kosli/cli version
```

{{< /tab >}}

{{< tab "From source" >}}
You can build Kosli CLI from source by running:
```shell {.command}
git clone https://github.com/kosli-dev/cli.git && cd cli
make build
```

{{< /tab >}}

{{< /tabs >}}


## Verifying the installation worked

Run this command:

```shell {.command}
kosli version
```

You should see an output like this:

```plaintext
version.BuildInfo{Version:"v2.11.1", GitCommit:"e34ea09fe4c7c4e2b3b0b0e1e3e1cfe3d56be7c3", GitTreeState:"clean"}
```

## Configure your API key and Org

You can configure the API key and the Kosli organization using the config file, environment variables, or command flags.  
See [CLI Configuration](/getting_started/cli-configuration/) for details.

{{< hint info >}}

You can find the API key in the Kosli app at **Settings > Your Profile**.

{{< /hint >}}

## Configuring request timeouts

The Kosli CLI supports configurable request timeouts via environment variables:

```shell {.command}
export KOSLI_REQUEST_TIMEOUT=60   # default request timeout in seconds
export KOSLI_UPLOAD_TIMEOUT=600   # upload timeout in seconds
```

You can also set the timeout using the `--request-timeout` flag:

```shell {.command}
kosli attest artifact --request-timeout 120 ...
```

The default timeout is 30 seconds for standard requests and 300 seconds for uploads.


## What's next?

Now that the CLI is installed, continue to [Part 3: your first Kosli commands](/getting_started/part_3/).
