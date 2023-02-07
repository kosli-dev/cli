---
title: 'How to use Kosli?'
weight: 140
---
# How to use Kosli?

## CLI

In order to [record environments](/getting_started/part_2_environments/), [artifacts](/getting_started/part_4_artifacts/) and [evidence](/getting_started/part_5_evidence/) to Kosli you need to use the [Kosli CLI](https://github.com/kosli-dev/cli). 
The CLI can be used to [search](/getting_started/part_8_querying/) Kosli and find out all you need to know about your runtime environments and artifacts.

Our CLI is an open source tool written in go and it's available for a number of different platforms.

### Installing the Kosli CLI

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
curl -L https://github.com/kosli-dev/cli/releases/download/v0.1.35/kosli_0.1.35_darwin_amd64.tar.gz | tar zx
sudo mv kosli /usr/local/bin/kosli
```
{{< /tab >}}

{{< tab "Docker" >}}
You can run the Kosli CLI in this docker container:
```shell {.command}
docker run -it --rm ghcr.io/kosli-dev/cli:v0.1.35 bash
```
{{< /tab >}}


{{< /tabs >}}


#### Verifying the installation worked

Run this command:
```shell {.command}
kosli version
```
The expected output should be similar to this:
```plaintext {.light-console}
version.BuildInfo{Version:"v0.1.35", GitCommit:"4058e8932ec093c28f553177e41c906940114866", GitTreeState:"clean", GoVersion:"go1.19.5"}
```

#### Usage

<!-- TODO:

explain kosli version and kosli status commands -->

The [Reference](/client_reference/) section contains all the information you may need to run the Kosli CLI. 

Most of the commands require a number of flags. Some of them are **required**, some are **optional** - you don't need to use them if you don't want to and some are **conditional** - you need to use them if a certain conditions occurs, e.g.:
* if you use `--sha256` flag it means you provide artifact's fingerprint on your own and we don't need to calculate it, so the flag `--artifact-type` is not needed
* if you want to read a docker digest from registry without pulling the image you need to provide registry information: `--registry-password`, `--registry-username` and `--registry-provider`

Each conditional flag is explained in its description.

Some of the flags are **defaulted**, and the default value will be always printed in the description. You can skip the flag if the default value is what you choose to use.

Depending on the CI tool you are using, some of the flags (including required ones) may also be defaulted, depending on the environment variables provided by the tool. If the flag is defaulted in your CI you don't have to provide it in the command. [Here](/ci-defaults) you can find more details about flags defaulted depending on CI.

#### Environment variables

Each flag can be provided directly or represented with environment variable. In order to represent a flag with environment variable you need to create a variable with a `KOSLI_` prefix, followed by the flag name capitalized and internal dashes replaced by underscores, e.g.:

* `--api-token` can be represented by `KOSLI_API_TOKEN` 
* `--owner` can be represented by `KOSLI_OWNER`

etc.


{{< hint warning >}}

#### Getting your Kosli API token

<!-- Put this in a separate page? -->
<!-- Add screen shot here? -->

* Go to https://app.kosli.com
* Log in or sign up using your github account
* Open your Profile page (click on your avatar in the top right corner of the page).

{{< /hint >}}

#### Config file

A config file is an alternative for using Kosli flags or Environment variables. Usually you'd use a config file for the values that rarely change - like api token or owner, but you can represent all Kosli flags with config file. The key for each value is the same as the flag name, capitalized, so `--api-token` would become `API-TOKEN`, and `--owner` would become `OWNER`, etc. 

You can use JSON, YAML or TOML format for your config file. 

If you want to keep certain Kosli configuration in a file use `--config-file` flag when running Kosli commands to let the cli tool know where to look for the file. The path given to `--config-file` flag should be a path relative to the location you're running kosli from. The file needs a valid format and extension, e.g.:

**kosli-conf.json:**
```
{
  "OWNER": "my-org",
  "API-TOKEN": "123456abcdef"
}
```

**kosli-conf.yaml:**
```
OWNER: "my-org"
API-TOKEN: "123456abcdef"
```

**kosli-conf.toml:**
```
OWNER = "my-org"
API-TOKEN = "123456abcdef"
```

When calling Kosli command you can skip file extension. For example, to list environments with `owner` and `api-token` in the configuration file you would run:

```
$ kosli environment ls --config-file kosli-conf
```

`--config-file` defaults to `kosli`, so if you name your file `kosli.<yaml|toml|json>` and the file is in the same location as where you run Kosli commands from, you can skip the `--config-file` altogether.

#### Dry run

You can use dry run to disable reporting to app.kosli.com - e.g. if you're just trying things out, or troubleshooting (dry run will print the payload the cli would send in a non dry run mode). 

Here are two ways of enabling a dry run:
1. use the `--dry-run` flag (no value needed) to enable it per command
1. set the `KOSLI_API_TOKEN` environment variable to `DRY_RUN` to enable it globally (e.g. in your terminal or CI)

## Web UI

[app.kosli.com](https://app.kosli.com) is an easy way to monitor the status of your environments and pipelines. All you need to log in is a GitHub account.

![app.kosli.com](/images/app.png)

On the left of the page you can see the menu where you can:

1. Switch to another organization using dropdown menu
2. Switch to the Environments or the Pipelines view
3. Enter organization settings page
4. Access this documentation page

In the top right corner of the page you will see your GitHub avatar where you can access your profile settings (containing your Kosli api key). You'll also find there a link to the page where you can create a shared organization and links to [cyber-dojo demo project](https://app.kosli.com/cyber-dojo/environments/) and to log out. 