---
title: "Part 2: Install Kosli"
bookCollapseSection: false
weight: 220
---
# Part 2: Install Kosli

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
version.BuildInfo{Version:"v{{< cli-version >}}", GitCommit:"4058e8932ec093c28f553177e41c906940114866", GitTreeState:"clean", GoVersion:"go1.19.5"}
```

## Using the CLI

The [CLI Reference](/client_reference/) section contains all the information you may need to run the Kosli CLI. The CLI flags offer flexibility for configuration and can be assigned in three distinct manners:

1. Directly on the command line.
2. Via environment variables.
3. Within a config file.
   
Among these options, priority is given in the following order: Option 1 holds the highest precedence, followed by Option 2, with Option 3 being the least prioritized.

### Assigning flags via environment variables

To assign a CLI flag using environment variables, generate a variable prefixed with KOSLI_. Utilize the flag's name in uppercase and substitute any internal dashes with underscores. For instance:


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

```
$ kosli environment ls --config-file=kosli-conf
```
