---
title: Get familiar with Kosli
bookCollapseSection: false
weight: 505
---

# Get familiar with Kosli

> The following guide is the easiest and quickest way to try Kosli out and understand it's features. 
It is made to run from your local machine, but the same concepts and steps apply to using Kosli in a production setup.

In this tutorial, you'll learn how Kosli allows you to follow a source code change to runtime environments.
You'll set up a `docker` environment, use Kosli to record build and deployment events, and track what artifacts are running in your runtime environment. 

This tutorial uses the `docker` Kosli environment type, but the same steps can be applied to other supported environment types.

{{< hint info >}}
As you go through the guide you can also check your progress from 
[your browser](https://app.kosli.com).

In the upper left corner there is a house icon. Next to it you can select
which organization you want to view. Your personal organization
has the same name as your github login name, and is the organization you will
be using in this guide.
{{< /hint >}}

## Step 1: Prerequisites and Kosli account

To follow the tutorial, you will need to:

- Install both `Docker` and `docker-compose`.
- [Create a Kosli account](https://app.kosli.com/sign-up) if you have not got one already.
- [Install the Kosli CLI](/kosli_overview/kosli_tools/#installing-the-kosli-cli) and [set the `KOSLI_API_TOKEN` and `KOSLI_ORG` environment variables](/kosli_overview/kosli_tools/#getting-your-kosli-api-token).
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

## Step 2: Create a Kosli trail

A Kosli *trail* stores information about what happens in your build system.
The output of the build system is called an *artifact* in Kosli. An artifact could be, for example,
an application binary, a docker image, a directory, or a file.

When attesting artifacts and evidence to a Kosli trail, each attestation must be named.
These names are defined in a yml file.
You will be using the file called `kosli_trail.yml` in the root of the git repo
you cloned in the previous step. Confirm this file exists by catting it:

```shell {.command}
cat kosli_trail.yml
```

which should produce the following output:
```plaintext {.light-console}
version: 1

trail:
  artifacts:
    - name: nginx
```

A trail lives inside a Kosli flow.
Start by creating a new Kosli flow called `quickstart-nginx`
based on this yml file:

```shell {.command}
kosli create flow2 quickstart-nginx \
    --description "Flow for quickstart nginx image" \
    --template-file kosli_trail.yml
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

Now create a Kosli trail, in this flow, whose name is the current git-commit:

```shell {.command}
kosli begin trail $(git rev-parse HEAD) \
    --flow quickstart-nginx
```

## Step 3: Create a Kosli environment

A Kosli *environment* stores snapshots containing information about
the software artifacts you are running in your runtime environment.

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

## Step 4: Attest an artifact to Kosli

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

Now you can attest the artifact to Kosli. 
This tutorial uses a dummy value for the `--build-url` flag, in a real installation 
this would be a defaulted link to a build service (e.g. Github Actions).

```shell {.command}
kosli attest artifact nginx:1.21 \  
    --name nginx \
    --flow quickstart-nginx \
    --trail $(git rev-parse HEAD) \
    --artifact-type docker \
    --build-url https://example.com \
    --commit-url https://github.com/kosli-dev/quickstart-docker-example/commit/9f14efa0c91807da9a8b1d1d6332c5b3aa24a310 \
    --git-commit $(git rev-parse HEAD)    
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

## Step 5: Report expected deployment of the artifact

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

## Step 6: Report what is running in your environment

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

## Step 7: Searching Kosli

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

Visit the [Kosli Querying](/getting_started/querying/) guide to learn more about the search command.