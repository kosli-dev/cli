---
title: Get Familiar with Kosli
bookCollapseSection: false
weight: 1
---

# Get Familiar with Kosli

Kosli stores information about the SW you build in CI pipelines
and run in your runtime environment. The Kosli CLI is used for reporting
and querying the information.

Typically all the reporting will be done as part of your CI and runtime systems.
In the getting started you don't need any of this. Local code, git and a terminal are enough.

The Kosli CLI is tool agnostic and can run on any major platform (Linux, Mac, Windows).
Kosli does not require you to change your existing process.

The purpose of this guide is to familiarize you with the Kosli CLI and concepts.

When you are done with the guide you should be able to start adding Kosli to
your CI system and runtime environment.


## Things to prepare

### CLI

The Kosli CLI can be downloaded from: https://github.com/kosli-dev/cli/releases
Put it in a location you'll be running it from (as `./kosli`) or add it to your PATH so you can use it anywhere (as `kosli`)

### Local setup

For these examples we are simulating a system with source code, build and a running server.
We have a script to help you run these simulations so you don't need to type so many commands.

You can download the file from here:
https://raw.githubusercontent.com/kosli-dev/cli/main/simulation_commands.bash

or from the command line with
```shell
$ cd /tmp
$ curl -O https://raw.githubusercontent.com/kosli-dev/cli/main/simulation_commands.bash
```

Source the simulation commands so you can use them later on in the examples.
```shell
$ source simulation_commands.bash
```

To see what a command (eg create_git_repo_in_tmp) in the simulation is actually doing run 
```shell
$ type create_git_repo_in_tmp
create_git_repo_in_tmp is a function
create_git_repo_in_tmp () 
{ 
    pushd /tmp;
    mkdir try-kosli;
    ...
```

Create the git repo and simulate a build and deployment to server.
```shell
$ create_git_repo_in_tmp
$ simulate_build
$ simulate_deployment
``` 

While going through the getting started guide, feel free to explore the
functionality by updating the source code, building and deploying new versions.


### Use of environment variables

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

### Web interface

As you go through the guide you can also check your progress from
the [web interface](https://app.kosli.com).

In the upper left corner there is a house icon. Next to it you can select
which organization you want to view. Your personal organization
has the same name as your github login name, and is the organization you will
be using in this guide.


# Environment

A Kosli environment stores information about
what SW is running in your actual runtime environment (server, Kubernetes cluster, AWS, ...)
We use one Kosli environment per runtime environment. 

A typical setup reports what is running on the 
staging server and on the production server. To report what is 
running you run the Kosli CLI command periodically. The Kosli CLI will
detect the version of the SW you are currently running and report
it to the Kosli environment.


## Create a Kosli environment

To follow the examples make sure you have followed the instructions in Local setup

We create a Kosli environment.
```shell
$ kosli environment declare \
    --name production \
    --environment-type server \
    --description "Production server (for kosli getting started)"
```

You can immediately verify that the Kosli environment was created:
```shell
$ kosli environment ls
NAME        TYPE    LAST REPORT  LAST MODIFIED
production  server               2022-08-16T07:53:43+02:00

$ kosli environment get production
Name:              production
Type:              server
Description:       Production server (for kosli getting started)
State:             N/A
Last Reported At:  N/A
```

In the web interface you can select the **Environments** menu on the left.
It will show you that you have a *production* environment and that
no reports have been received.


## Report the SW running in your environment

We simulate a report from our server by reporting two dummy files for the web and
database applications.
```shell
$ kosli environment report server production \
    --paths /tmp/try-kosli/server/web_*bin \
    --paths /tmp/try-kosli/server/db_*.bin
```

We can see that the server has started, and how long it has run.
```shell
$ kosli snapshot ls production
SNAPSHOT  FROM                  TO   DURATION
1         16 Aug 22 07:54 CEST  now  11 seconds
```

We can get a more detailed view of what is currently running on the server.
```shell
$ kosli snapshot get production
COMMIT  ARTIFACT                                                                  PIPELINE  RUNNING_SINCE  REPLICAS
N/A     Name: /tmp/try-kosli/server/web_1.bin                                     N/A       2 minutes ago  1
        SHA256: a7a87c332500a40f9a01b811ec75f51b40188a3dabd205feb0fa7c3eafb25fbe                           
                                                                                                           
N/A     Name: /tmp/try-kosli/server/db_1.bin                                      N/A       2 minutes ago  1
        SHA256: 0efde582a933f011c3ae9007467a7f973a874517093e9a5a05ea55476f7c91af                           
```

If you refresh the environment page in the web browser you can see that we have
a time-stamp for when the environment changed. Pressing the *production* link
gives you a detailed view of what is running now.

Typically a server periodically sends a snapshot of what is currently running to Kosli. Kosli
only stores the new snapshot if the snapshot changes, so resending the same environment report
several times will not lead to duplication of snapshots.
```shell
$ kosli environment report server production \
    --paths /tmp/try-kosli/server/web_*bin \
    --paths /tmp/try-kosli/server/db_*.bin
$ kosli snapshot get production
COMMIT  ARTIFACT                                                                  PIPELINE  RUNNING_SINCE  REPLICAS
N/A     Name: /tmp/try-kosli/server/web_1.bin                                     N/A       2 minutes ago  1
        SHA256: a7a87c332500a40f9a01b811ec75f51b40188a3dabd205feb0fa7c3eafb25fbe                           
                                                                                                           
N/A     Name: /tmp/try-kosli/server/db_1.bin                                      N/A       2 minutes ago  1
        SHA256: 0efde582a933f011c3ae9007467a7f973a874517093e9a5a05ea55476f7c91af                           
```

We simulate an update of the web application to a new version, build and deploy it
```shell
$ update_web_src
$ simulate_build
$ simulate_deployment
```

Report what is now running on server
```shell
$ kosli environment report server production \
    --paths /tmp/try-kosli/server/web_*bin \
    --paths /tmp/try-kosli/server/db_*.bin
```

We can see we have created a new snapshot.
```shell
$ kosli snapshot ls production
SNAPSHOT  FROM                  TO                    DURATION
2         16 Aug 22 07:58 CEST  now                   9 seconds
1         16 Aug 22 07:54 CEST  16 Aug 22 07:58 CEST  4 minutes
```

We can see that we are currently running web version 2 in production.
```shell
$ kosli snapshot get production
COMMIT  ARTIFACT                                                                  PIPELINE  RUNNING_SINCE   REPLICAS
N/A     Name: /tmp/try-kosli/server/web_2.bin                                     N/A       39 seconds ago  1
        SHA256: cbc92ce1291830382ec23b95efc213d6e1725b5157bcb2927d48296b61c86746                            
                                                                                                            
N/A     Name: /tmp/try-kosli/server/db_1.bin                                      N/A       6 minutes ago   1
        SHA256: 0efde582a933f011c3ae9007467a7f973a874517093e9a5a05ea55476f7c91af                            
```                    

Here, using the bare environment name (eg production) always refer to the latest snapshot
in that environment. We can also use the 
Kosli CLI to check what was running in previous snapshots.
Here we look at what was running in snapshot #1 in production.
```shell
$ kosli snapshot get production#1
COMMIT  ARTIFACT                                                                  PIPELINE  RUNNING_SINCE  REPLICAS
N/A     Name: /tmp/try-kosli/server/web_1.bin                                     N/A       7 minutes ago  1
        SHA256: a7a87c332500a40f9a01b811ec75f51b40188a3dabd205feb0fa7c3eafb25fbe                           
                                                                                                           
N/A     Name: /tmp/try-kosli/server/db_1.bin                                      N/A       7 minutes ago  1
        SHA256: 0efde582a933f011c3ae9007467a7f973a874517093e9a5a05ea55476f7c91af                           
```

In the web interface you should now also be able to see 2 snapshots. The Log
tab (TODO: add icon) should show what changed in snapshot 1 and snapshot 2.


# Pipelines

A Kosli pipeline stores information about what happens in your build system.
The output of the build system is called an *artifact* in Kosli. This can be
an application, docker image, documentation, filesystem and so on.

Some organizations have a CI system where one CI pipeline builds one 
artifact, some have a CI system where one CI pipeline builds several
artifacts. For both cases we use one Kosli pipeline for each artifact.
We use the Kosli CLI to report information about the creation of an
artifact to the Kosli pipeline.

A Kosli pipeline can also be used to store any information related to 
the artifact you have built, like test results, manual approvals, 
pull-requests and so on.


## Create a Kosli pipeline

To follow the examples make sure you have followed the instructions in Local setup.

We create a Kosli pipeline where we can report what SW our CI system
is building. Since we are building two applications we are making
two Kosli pipelines `web-server` and `database-server`.

Create your new pipelines:
```shell
$ kosli pipeline declare \
    --pipeline web-server \
    --description "pipeline to build web-server" \
    --visibility private \
    --template artifact

$ kosli pipeline declare \
    --pipeline database-server \
    --description "pipeline to build database-server" \
    --visibility private \
    --template artifact
```

You can immediately verify the Kosli pipelines were created:
```shell
$ kosli pipeline ls
NAME             DESCRIPTION                        VISIBILITY
database-server  pipeline to build database-server  private
web-server       pipeline to build web-server       private
```

In the web interface you can select the **Pipelines** menu on the left.
It will show you that you have a *web-server* and *database-server* pipeline.
If you press either of the pipelines they will show that no artifacts have
been reported for the pipelines.


## Build artifacts and report them to Kosli

Simulate building our SW
```shell
$ simulate_build
```

We can now report we have built the web and database applications. We are using
a dummy `--build-url`, in real life it would be a CI build URL.
```shell
$ kosli pipeline artifact report creation /tmp/try-kosli/build/web_$(cat /tmp/try-kosli/code/web.src).bin \
    --pipeline web-server \
    --artifact-type file \
    --build-url file://dummy \
    --commit-url file:///tmp/try-kosli/code \
    --git-commit $(cd /tmp/try-kosli/code; git rev-parse HEAD)

$ kosli pipeline artifact report creation /tmp/try-kosli/build/db_$(cat /tmp/try-kosli/code/db.src).bin \
    --pipeline database-server \
    --artifact-type file \
    --build-url file://dummy \
    --commit-url file:///tmp/try-kosli/code \
    --git-commit $(cd /tmp/try-kosli/code; git rev-parse HEAD)
```

We can see we have built one artifact in our *web-server* pipeline
```shell
$ kosli artifact ls web-server
COMMIT   ARTIFACT                                                                  STATE      CREATED_AT
5187374  Name: web_2.bin                                                           COMPLIANT  16 Aug 22 08:00 CEST
         SHA256: cbc92ce1291830382ec23b95efc213d6e1725b5157bcb2927d48296b61c86746             
```

And one for the *database-server* pipeline
```shell
$ kosli artifact ls database-server
COMMIT   ARTIFACT                                                                  STATE      CREATED_AT
5187374  Name: db_1.bin                                                            COMPLIANT  16 Aug 22 08:01 CEST
         SHA256: 0efde582a933f011c3ae9007467a7f973a874517093e9a5a05ea55476f7c91af             
```

We can also get detailed information about each artifact that has been reported.
```shell
$ kosli artifact get --pipeline database-server 0efde582a933f011c3ae9007467a7f973a874517093e9a5a05ea55476f7c91af
Name:         db_1.bin
State:        COMPLIANT
Git commit:   518737485e5150ee6255a1c74749997d380c1708
Build URL:    file://dummy
Commit URL:   file:///tmp/try-kosli/code
Created at:   16 Aug 22 08:01 CEST • 2 hours ago
Approvals:    None
Deployments:  None
Evidence:
```

In the web interface you can select the *database-server* pipeline and then the *db_1.bin*
artifact to get more details.


# Deployments

We assume the user has done both Environments and Pipelines first.

A Kosli deployment command is used to indicate an aritfact is
being deployed to a given runtime environment. 


## Deploy SW to server and report the deployment to Kosli

Simulate deploying our SW to the server
```shell
$ simulate_deployment
```

Report to Kosli that the web SW has been deployed.
```shell
$ kosli pipeline deployment report /tmp/try-kosli/build/web_$(cat /tmp/try-kosli/code/web.src).bin \
    --pipeline web-server \
    --artifact-type file \
    --build-url file://dummy \
    --environment production \
    --description "Web server version $(cat /tmp/try-kosli/code/web.src)"
```

We can verify the deployment with
```shell
$ kosli deployment ls web-server
ID   ARTIFACT                                                                  ENVIRONMENT  REPORTED_AT
1    Name: web_2.bin                                                           production   16 Aug 22 08:02 CEST
     SHA256: cbc92ce1291830382ec23b95efc213d6e1725b5157bcb2927d48296b61c86746               
```

We can also get detailed information about a deployment.
```shell
$ kosli deployment get --pipeline web-server 1
ID:               1
Artifact SHA256:  cbc92ce1291830382ec23b95efc213d6e1725b5157bcb2927d48296b61c86746
Artifact name:    web_2.bin
Build URL:        file://dummy
Created at:       16 Aug 22 08:02 CEST • 32 seconds ago
Environment:      production
Runtime state:    The artifact running since 16 Aug 22 07:58 CEST
```

If you select the *web_2.bin* artifact in the web interface it will show
that it was part of Deployment #1 to *production* environment.


# For developers
You can extract all the commands to execute from this document by running
```shell
cat docs.kosli.com/content/get_familiar_with_Kosli/_index.md | sed -e :a -e '/\\$/N; s/\\\n//; ta' | egrep '^\$ ' | sed "s/^..//" 
```
