---
title: Installing the Kosli CLI
bookCollapseSection: false
weight: 1
---


## Installing the Kosli CLI

Kosli CLI can be installed from package managers, 
by Curling pre-built binaries, or by running inside a Docker container.  
We recommend using a Docker container for the tutorials.
{{< tabs "installKosli" >}}

{{< tab "Homebrew" >}}
If you have [Homebrew](https://brew.sh/) (available on MacOS, Linux or Windows Subsystem for Linux), 
you can install the Kosli CLI by running: 

```shell {.command}
brew install kosli-dev/tap/kosli
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
curl -L https://github.com/kosli-dev/cli/releases/download/v0.1.10/kosli_0.1.10_darwin_amd64.tar.gz | tar zx
sudo mv kosli /usr/local/bin/kosli
```
{{< /tab >}}

{{< tab "Docker" >}}
You can run the Kosli CLI in this docker container:
```shell {.command}
docker run -it --rm ghcr.io/kosli-dev/cli:v0.1.10 bash
```
{{< /tab >}}


{{< /tabs >}}


## Verifying the installation worked

Run this command:
```shell {.command}
kosli version
```
The expected output should be similar to this:
```
version.BuildInfo{Version:"v0.1.10", GitCommit:"9c623f1e6c293235ddc8de1e347bf99a1b356e48", GitTreeState:"clean", GoVersion:"go1.17.11"}
```

## Getting your Kosli API token

<!-- Put this in a separate page? -->
<!-- Add screen shot here? -->

* Go to https://app.kosli.com
* Log in using your github account
* Open your Profile page (click on your avatar in the top right corner of the page).

## Using environment variables

<!-- Put this in a separate page? -->

The `--api-token` and `--owner` flags are used in every `kosli` CLI command.  
Rather than retyping these every time you run `kosli`, you can set them as environment variables.

Simply capitalize the flag in snake case and add the `KOSLI_` prefix.  
For example, after this:

```shell
export KOSLI_API_TOKEN=abcdefg
export KOSLI_OWNER=cyber-dojo
```

Then instead of:

```shell
kosli pipeline ls --api-token abcdefg --owner cyber-dojo 
```

You can use:

```shell
kosli pipeline ls 
```




