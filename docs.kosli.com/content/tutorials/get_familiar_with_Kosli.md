---
title: Get familiar with Kosli
bookCollapseSection: false
weight: 505
---

# Get familiar with Kosli

The following guide is the easiest and quickest way to try Kosli out and understand it's features. But is not a real life use case for Kosli - usually you'd run Kosli in your CI and remote environments.  

So you can try it out using just your local machine and `docker`. In our ***Guides*** and ***Kosli integrations*** sections you'll find all the information needed to run it in actual projects.

In this tutorial, you'll learn how Kosli allows you to follow a source code change to runtime environments.
You'll set up a `docker` environment, use Kosli to record build and deployment events, and track what artifacts are running in your runtime environments. 

This tutorial uses the `docker` Kosli environment type, but the same steps can be applied to other supported environment types.

{{< hint info >}}
As you go through the guide you can also check your progress from 
[your browser](https://app.kosli.com).

In the upper left corner there is a house icon. Next to it you can select
which organization you want to view. Your personal organization
has the same name as your github login name, and is the organization (in the context of Kosli CLI called "owner") you will
be using in this guide.
{{< /hint >}}

## Step 1: Prerequisites and Kosli account

To follow the tutorial, you will need to:

- Install both `Docker` and `docker-compose`.
- [Install the Kosli CLI](/kosli_overview/kosli_tools/#installing-the-kosli-cli) and [set the `KOSLI_API_TOKEN` and `KOSLI_OWNER` environment variables](/kosli_overview/kosli_tools/#getting-your-kosli-api-token).
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

### Create Kosli account

You need a GitHub account to be able to use Kosli.  
Go to [app.kosli.com](https://app.kosli.com) and use "Sign up with GitHub" button to create a Kosli account. 


## Step 2: Install the Kosli CLI

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

### Verifying the installation worked

Run this command:
```shell {.command}
kosli version
```
The expected output should be similar to this:
```plaintext {.light-console}
version.BuildInfo{Version:"v0.1.35", GitCommit:"4058e8932ec093c28f553177e41c906940114866", GitTreeState:"clean", GoVersion:"go1.19.5"}
```

## Step 3: Configure your working environment

### Getting your Kosli API token

<!-- Put this in a separate page? -->
<!-- Add screen shot here? -->

To be able to run Kosli commands (from your local machine, but the same goes for any CI/CD system you use) you need a Kosli API Token to be able to authenticate. It's a common practice to configure the token as an environment variable (or e.g. a secret in GitHub Actions or Bitbucket, etc)

To retrieve your API Token:

* Go to https://app.kosli.com
* Log in or sign up using your github account
* Open your Profile page (click on your avatar in the top right corner of the page) and copy the API Key

### Using environment variables

<!-- Put this in a separate page? -->

The `--api-token` and `--owner` flags are used in every `kosli` CLI command.  
Rather than retyping these every time you run `kosli`, you can set them as environment variables.

The owner is the name of the organization you intend to use - it is either your private organization, which has exactly the same name as your GitHub username, or a shared organization (if you created or have been invited to one).

By setting the environment variables:
```shell {.command}
export KOSLI_API_TOKEN=abcdefg
export KOSLI_OWNER=cyber-dojo
```

you can use

```shell {.command}
kosli list flows 
```

instead of

```shell {.command}
kosli list flows --api-token abcdefg --owner cyber-dojo 
```

You can represent **ANY** flag as an environment variable. To do that you need to capitalize the words in the flag, replacing dashes with underscores, and add the `KOSLI_` prefix. For example, `--api-token` becomes `KOSLI_API_TOKEN`.

## Step 4: Create a Kosli flow

A Kosli *flow* stores information about what happens in your build system.
The output of the build system is called an *artifact* in Kosli. An artifact could be, for example,
an application binary, a docker image, a directory, or a file. 

Start by creating a new Kosli flow:

```shell {.command}
kosli create flow quickstart-nginx \
    --description "Flow for quickstart nginx image"
```

You can confirm that the Kosli flow was created by running:
```shell {.command}
kosli list flows
```
which should produce the following output:
```plaintext {.light-console}
NAME              DESCRIPTION                          VISIBILITY
quickstart-nginx  Flow for quickstart nginx image      private
```
{{< hint info >}}
In the web interface you can select the *Flows* option on the left.
It will show you that you have a *quickstart-nginx* flow.
If you select the flow it will show that no artifacts have
been reported yet.
{{< /hint  >}}

## Step 5: Create a Kosli environment

A Kosli *environment* stores snapshots containing information about
the software artifacts you are running in your runtime environments.

Create a Kosli environment:

```shell {.command}
kosli create environment quickstart \
    --type docker \
    --description "quickstart environment for tutorial"
```

You can verify that the Kosli environment was created:

```shell {.command}
kosli list environments
```

```plaintext {.light-console}
NAME        TYPE    LAST REPORT  LAST MODIFIED
quickstart  docker               2022-11-01T15:30:56+01:00
```

{{< hint info >}}
If you refresh the *Environments* web page in your Kosli account, 
it will show you that you have a *quickstart* environment and that
no reports have been received.
{{< /hint >}}

## Step 6: Report artifacts to Kosli

Typically, you would build an artifact in your CI system. 
The quickstart-docker repository contains a `docker-compose.yml` file which uses an [nginx](https://nginx.org/) docker image 
which you will be using as your artifact in this tutorial instead.

Pull the docker image - the Kosli CLI needs the artifact to be locally present to 
generate a "fingerprint" to identify it:

```shell {.command}
docker-compose pull
```

You can check that this has worked by typing: 
```shell {.command}
docker images nginx
```
The output should look like this:
```plaintext {.light-console}
REPOSITORY   TAG       IMAGE ID       CREATED        SIZE
nginx        1.21      8f05d7383593   5 months ago   134MB
```

Now you can report the artifact to Kosli. 
This tutorial uses a dummy value for the `--build-url` flag, in a real installation 
this would be a defaulted link to a build service (e.g. Github Actions).

```shell {.command}
kosli report artifact nginx:1.21 \
    --flow quickstart-nginx \
    --artifact-type docker \
    --build-url https://example.com \
    --commit-url https://github.com/kosli-dev/quickstart-docker-example/commit/9f14efa0c91807da9a8b1d1d6332c5b3aa24a310 \
    --git-commit 9f14efa0c91807da9a8b1d1d6332c5b3aa24a310
```

You can verify that you have reported the artifact in your *quickstart-nginx* flow:

```shell {.command}
kosli list artifacts --flow quickstart-nginx
```

```plaintext {.light-console}
COMMIT   ARTIFACT                                                                       STATE      CREATED_AT
9f14efa  Name: nginx:1.21                                                               COMPLIANT  Tue, 01 Nov 2022 15:46:59 CET
         Fingerprint: 2bcabc23b45489fb0885d69a06ba1d648aeda973fae7bb981bafbb884165e514                 
```

## Step 7: Report expected deployment of the artifact

Before you run the nginx docker image (the artifact) on your docker host, you need to report 
to Kosli your intention of deploying that image. This allows Kosli to match what you 
expect to run in your environment with what is actually running, and flag any mismatches.  

```shell {.command}
kosli expect deployment nginx:1.21 \
    --flow quickstart-nginx \
    --artifact-type docker \
    --build-url https://example.com \
    --environment quickstart \
    --description "quickstart-nginx artifact deployed to quickstart env"
```

You can verify the deployment with:

```shell {.command}
kosli list deployments --flow quickstart-nginx
```

```plaintext {.light-console}
ID   ARTIFACT                                                                       ENVIRONMENT  REPORTED_AT
1    Name: nginx:1.21                                                               quickstart   Tue, 01 Nov 2022 15:48:47 CET
     Fingerprint: 2bcabc23b45489fb0885d69a06ba1d648aeda973fae7bb981bafbb884165e514  
```

Now run the artifact:
```shell {.command}
docker-compose up -d
```

You can confirm the container is running:
```shell {.command}
docker ps
```
The output should include an entry similar to this:
```plaintext {.light-console}
CONTAINER ID  IMAGE      COMMAND                 CREATED         STATUS         PORTS                  NAMES
6330e545b532  nginx:1.21 "/docker-entrypoint.…"  35 seconds ago  Up 34 seconds  0.0.0.0:8080->80/tcp   quickstart-nginx
```

## Step 8: Report what is running in your environment

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

{{< hint info >}}
If you refresh the *Environments* web page in your Kosli account, you will see 
that there is now a timestamp for *Last Change At* column. 
Select the *quickstart* link on left for a detailed view of what is currently running.
{{< /hint >}}

## Searching Kosli

Now that you have reported your artifact and what's running in our runtime environment,
you can use the `kosli search` command to find everything Kosli knows about an artifact or a git commit.

For example, you can give Kosli search the git commit SHA which you used when you reported the artifact: 

```shell {.command}
kosli search 9f14efa0c91807da9a8b1d1d6332c5b3aa24a310
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

Visit [Part 9: Querying](/getting_started/part_9_querying/) section to learn more