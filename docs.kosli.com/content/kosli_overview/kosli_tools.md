---
title: 'How to use Kosli?'
weight: 140
---

# How to use Kosli?

## CLI

In order to [record environments](/getting_started/environments/), [artifacts](/getting_started/artifacts/) and [evidence](/getting_started/evidence/) to Kosli you need to use the [Kosli CLI](https://github.com/kosli-dev/cli). 
The CLI can also be used to [search](/getting_started/querying/) Kosli and find out all you need to know about your runtime environments and artifacts.

Our CLI is an open source and is available for a number of different platforms.

### Installing the Kosli CLI

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


#### Verifying the installation worked

Run this command:
```shell {.command}
kosli version
```
The expected output should be similar to this:
```plaintext {.light-console}
version.BuildInfo{Version:"v{{< cli-version >}}", GitCommit:"4058e8932ec093c28f553177e41c906940114866", GitTreeState:"clean", GoVersion:"go1.19.5"}
```

#### Usage

The [CLI Reference](/client_reference/) section contains all the information you may need to run the Kosli CLI. 

Most of the commands require a number of flags. Some of them are **required**, others are **optional** and some are
 **conditional** - you need to use them if certain conditions occur. Each command doc/help provides details about its flag usage. 

Some flags (including required ones) may be defaulted, depending on the environment 
variables your CI provides. If the flag is defaulted in your CI, you don't have to 
provide it in the command. 
[Here](/ci-defaults) you can find details of all CI flags defaults.

#### Environment variables

Each flag can be provided directly or represented with an environment variable. 
To represent a flag with environment variable create a variable with a `KOSLI_` prefix, followed by the flag name, with all letters capitalized and internal dashes replaced by underscores, e.g.:

* `--api-token` is represented by `KOSLI_API_TOKEN` 
* `--org` is represented by `KOSLI_ORG`


{{< hint info >}}

#### Getting your Kosli API token

<!-- Put this in a separate page? -->
<!-- Add screen shot here? -->

* Go to https://app.kosli.com
* Log in or sign up using your github account
* Open your Profile page (click on your avatar in the top right corner of the page).

{{< /hint >}}

#### Config file

A config file is an alternative to using Kosli flags or environment variables. 
Usually you'd use a config file for the values that rarely change - like api token or org, 
but you can represent all Kosli flags in a config file. The key for each value is the same 
as the flag name, capitalized, so `--api-token` would become `API-TOKEN`, and `--org` would 
become `ORG`, etc. 

You can use JSON, YAML or TOML format for your config file. 

You can use the `--config-file` flag when 
running Kosli commands to let the Kosli CLI know where to look for a config file. 
The file needs a valid format and extension, e.g.:

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

```
$ kosli environment ls --config-file=kosli-conf
```

The `--config-file` flag defaults to `kosli`, so if you name your file `kosli.<yaml|toml|json>` and 
the file is in the same location where you run Kosli CLI commands from, you can 
skip the `--config-file` flag altogether.

#### Dry run

You can use dry run flag to disable reporting to app.kosli.com - e.g. if you're just 
trying things out, or troubleshooting (dry run will print the payload the CLI would send 
in a non dry run mode). 

Here are two ways of enabling a dry run:
1. use the `--dry-run` flag (no value needed) to enable it per command;
2. set the `KOSLI_DRY_RUN` environment variable to `TRUE` to enable it globally (e.g. in your terminal or CI).

## Web UI

[app.kosli.com](https://app.kosli.com) is an easy way to monitor the status of your environments and flows. All you need to log in is a GitHub account.

{{<figure src="/images/envs.png" alt="app.kosli.com" width="900">}}

On the left of the page you can see the menu where you can:

1. Switch to another organization using dropdown menu
2. Switch to the Environments or the Flows view
3. Enter organization settings page

In the top right corner of the page you will see your GitHub avatar where you can 
access your profile settings (containing your Kosli API key). You'll also find there a 
link to the page where you can create a shared organization and 
links to [cyber-dojo demo project](https://app.kosli.com/cyber-dojo/environments/) and to 
log out.
