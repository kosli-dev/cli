---
title: Quick start
bookCollapseSection: false
weight: 10
---

# Getting started with Kosli

In this tutorial, you will see how Kosli allows you to follow a source code change to runtime environments.
You will use Kosli to record build and deployment events, and to
track what artifacts are running in your runtime environments. 

This tutorial uses the `docker` Kosli environment type, but the same steps can be applied to other supported environment types.

## Prerequisites

To follow the tutorial, you will need to:

- Install both `Docker` and `docker-compose`.
- [Install the Kosli CLI](/getting_started/installation) and [set the `KOSLI_API_TOKEN` and `KOSLI_OWNER` environment variables](/getting_started/installation#getting-your-kosli-api-token).
- You can check your Kosli set up by running: 
    ```shell {.command}
    kosli pipeline ls
    ```
    which should return a list of pipelines or the message "No pipelines were found".

- Clone our quickstart-docker repository:
    ```shell {.command}
    git clone https://github.com/kosli-dev/quickstart-docker-example.git
    cd quickstart-docker-example
    ```

{{< hint info >}}
As you go through the guide you can also check your progress from 
[your browser](https://app.kosli.com).

In the upper left corner there is a house icon. Next to it you can select
which organization you want to view. Your personal organization
has the same name as your github login name, and is the organization you will
be using in this guide.
{{< /hint >}}

## Kosli setup

### Creating a Kosli pipeline

A Kosli *pipeline* stores information about what happens in your build system.
The output of the build system is called an *artifact* in Kosli. An artifact could be, for example,
an application binary, a docker image, documentation, or a file. 

Start by creating a new Kosli pipeline:

```shell {.command}
kosli pipeline declare \
    --pipeline quickstart-nginx \
    --description "Pipeline for quickstart nginx image"
```

You can confirm that the Kosli pipeline was created by running:
```shell {.command}
kosli pipeline ls
```
which should produce the following output:
```plaintext {.light-console}
NAME              DESCRIPTION                          VISIBILITY
quickstart-nginx  Pipeline for quickstart nginx image  private
```
{{< hint info >}}
In the web interface you can select the *Pipelines* option on the left.
It will show you that you have a *quickstart-nginx* pipeline.
If you select the pipeline it will show that no artifacts have
been reported yet.
{{< /hint  >}}

### Creating a Kosli environment

A Kosli *environment* stores snapshots containing information about
the software artifacts that you are running in your runtime environments.

Create a Kosli environment:

```shell {.command}
kosli environment declare \
    --name quickstart \
    --environment-type docker \
    --description "quickstart environment for tutorial"
```

You can verify that the Kosli environment was created:

```shell {.command}
kosli environment ls
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

## Artifacts

### Reporting artifacts to Kosli

Typically, you would build an artifact in your CI system. 
The quickstart-docker repository contains a `docker-compose.yml` file which uses an [nginx](nginx.org) docker image 
which you will be using as your artifact in this tutorial instead.

Pull the docker image - Kosli CLI needs the artifact to be locally present to 
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
this would be a link to a build service (e.g. Github Actions).

```shell {.command}
kosli pipeline artifact report creation nginx:1.21 \
    --pipeline quickstart-nginx \
    --artifact-type docker \
    --build-url https://example.com \
    --commit-url https://github.com/kosli-dev/quickstart-docker-example/commit/9f14efa0c91807da9a8b1d1d6332c5b3aa24a310 \
    --git-commit 9f14efa0c91807da9a8b1d1d6332c5b3aa24a310
```

You can verify that you have reported the artifact in your *quickstart-nginx* pipeline:

```shell {.command}
kosli artifact ls quickstart-nginx
```

```plaintext {.light-console}
COMMIT   ARTIFACT                                                                       STATE      CREATED_AT
9f14efa  Name: nginx:1.21                                                               COMPLIANT  Tue, 01 Nov 2022 15:46:59 CET
         Fingerprint: 2bcabc23b45489fb0885d69a06ba1d648aeda973fae7bb981bafbb884165e514                 
```

### Deploying the artifact

Before you run the nginx docker image (the artifact) on your docker host, you need to report 
to Kosli your intention of deploying that image. This allows Kosli to match what you 
expect to run in your environment with what is actually running, and flag any mismatches.  

```shell {.command}
kosli expect deployment nginx:1.21 \
    --pipeline quickstart-nginx \
    --artifact-type docker \
    --build-url https://example.com \
    --environment quickstart \
    --description "quickstart-nginx artifact deployed to quickstart env"
```

You can verify the deployment with:

```shell {.command}
kosli deployment ls quickstart-nginx
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

### Reporting what is running in your environment

Report all the docker containers running on your machine to Kosli:
```shell {.command}
kosli environment report docker quickstart
```
You can confirm that this has created an environment snapshot:
```shell {.command}
kosli environment log quickstart
```
```plaintext {.light-console}
SNAPSHOT  FROM                           TO   DURATION
1         Tue, 01 Nov 2022 15:55:49 CET  now  11 seconds
```

You can get a detailed view of all the docker containers included in the snapshot report:
```shell {.command}
kosli environment get quickstart
```

```plaintext {.light-console}
COMMIT  ARTIFACT                                                                       PIPELINE  RUNNING_SINCE  REPLICAS
N/A     Name: nginx:1.21                                                               N/A       3 minutes ago  1
        Fingerprint: 8f05d73835934b8220e1abd2f157ea4e2260b9c26f6f63a8e3975e7affa46724
```

The `kosli environment report docker` command reports *all* the 
docker containers running in your environment, equivalent to the output from 
`docker ps`. This tutorial only shows the `nginx` container 
in the examples.

{{< hint info >}}
If you refresh the *Environments* web page in your Kosli account, you will see 
that there is now a timestamp for *Last Change At* column. 
Select the *quickstart* link on left for a detailed view of what is currently running.
{{< /hint >}}

## Searching Kosli

Now that you have reported our artifact and what's running in our runtime environment,
you can use the `kosli search` command to find everything Kosli knows about an artifact or a git commit.

For example, you can give Kosli search the git commit SHA which you used when you reported the artifact:: 

```shell {.command}
kosli search 9f14efa0c91807da9a8b1d1d6332c5b3aa24a310
```

```plaintext {.light-console}
Search result resolved to commit 9f14efa0c91807da9a8b1d1d6332c5b3aa24a310
Name:              nginx:1.21
Fingerprint:       2bcabc23b45489fb0885d69a06ba1d648aeda973fae7bb981bafbb884165e514
Has provenance:    true
Pipeline:          quickstart-nginx
Git commit:        9f14efa0c91807da9a8b1d1d6332c5b3aa24a310
Commit URL:        https://github.com/kosli-dev/quickstart-docker-example/commit/9f14efa0c91807da9a8b1d1d6332c5b3aa24a310
Build URL:         https://example.com
Compliance state:  COMPLIANT
History:
    Artifact created                             Tue, 01 Nov 2022 15:46:59 CET
    Deployment #1 to quickstart environment      Tue, 01 Nov 2022 15:48:47 CET
    Started running in quickstart#1 environment  Tue, 01 Nov 2022 15:55:49 CET
```