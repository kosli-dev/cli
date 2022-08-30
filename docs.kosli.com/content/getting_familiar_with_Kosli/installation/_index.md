---
title: Installing Kosli CLI
bookCollapseSection: false
weight: 1
---


## Installing Kosli CLI

This guide shows you how to install the Kosli CLI. Kosli CLI can be installed from package managers or pre-built binaries.

{{< tabs "installKosli" >}}

{{< tab "Homebrew" >}}
If you have [Homebrew](https://brew.sh/) (available on MacOS, Linux or Windows Subsystem for Linux), 
you can install the Kosli CLI by running: 

```shell {.command}
brew install kosli-dev/tap/kosli
```
{{< /tab >}}

{{< tab "APT" >}}
If you are using Ubuntu or Debian Linux, you can install the Kosli CLI by running the following commands:
```shell {.command}
sudo sh -c 'echo "deb [trusted=yes] https://apt.fury.io/kosli/ /"  > /etc/apt/sources.list.d/fury.list'
# if you are on a clean debian container/machine, you will need to install ca-certificates, otherwise ignore that step
sudo apt install ca-certificates

sudo apt update
sudo apt install kosli
```
{{< /tab >}}

{{< tab "YUM" >}}
If you have RedHat Linux, you can use YUM to install the Kosli CLI.
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
Alternatively, the Kosli CLI binary can be downloaded from [GitHub](https://github.com/kosli-dev/cli/releases)
Make sure to choose the correct tar file for your system.

For example, on Mac with AMD:
```shell {.command}
curl -L https://github.com/kosli-dev/cli/releases/download/v0.1.10/kosli_0.1.10_darwin_amd64.tar.gz | tar zx
sudo mv kosli /usr/local/bin/kosli
```

{{< /tab >}}

{{< /tabs >}}

## Verify the installation worked

To verify that Kosli CLI is successfully installed run the command below.
```shell {.command}
kosli version
```
The expected output should be similar to the one below
```
version.BuildInfo{Version:"v0.1.10", GitCommit:"9c623f1e6c293235ddc8de1e347bf99a1b356e48", GitTreeState:"clean", GoVersion:"go1.17.11"}
```

## Using environment variables

All the kosli commands contain some common
flags `--api-token` and `--owner`. By setting
these as environment variables we don't need to specify them. 

You do this by capitalizing the flag in snake case and adding the `KOSLI_` prefix. 
For example, to set `--api-token xx` from an environment variable, you can `export KOSLI_API_TOKEN=xx`, etc:

```shell
export KOSLI_API_TOKEN=<put your kosli API token here>
export KOSLI_OWNER=<put your github username here>
```

To get the kosli API token go to https://app.kosli.com, log in using your github account, and go to your Profile (you'll find it if you click on your avatar in the top right corner of the page)



