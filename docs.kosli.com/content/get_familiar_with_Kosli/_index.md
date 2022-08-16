---
title: Get familiar with Kosli
bookCollapseSection: true
weight: 1
---

# Get familiar with Kosli

Kosli konsists of a database with a WEB UI to store information about the
SW you build and run in your runtime environment, and a Kosli CLI
used to report to and query the database.

Typically all the reporting will be done as part of your CI and runtime systems.
In the getting started you don't need any of this. Local code, git and terminal are enough.

The Kosli CLI is tool agnostics and can run on any major platform (Linux, Mac, Windows).
Kosli does not require you to change your existing process.

The purpose of this guide is to familiarize you with the Kosli tool and concepts.

When you are done with the guide you should be able to start adding Kosli to
your CI system and runtime environment.


## Things to prepare

### CLI

The `kosli` tool can be downloaded from: https://github.com/kosli-dev/cli/releases
Put it in a location you'll be running it from (as `./kosli`) or add it to your PATH so you can use it anywhere (as `kosli`)

### Local setup

For these examples we have a very basic settup with source code, build system and
a server "running" the SW.

```shell
$ mkdir try-kosli
$ cd try-kosli
$ mkdir code server build

# Create version 1 of our source code
$ echo "1" > code/web.src
$ echo "1" > code/db.src

# Create a git repository of the source code
$ cd code
$ git init
$ git add *src
$ git commit -m "Version one of web and database"
$ cd ..

# Build our SW with
$ echo "web version $(cat code/web.src)" > build/web_$(cat code/web.src).bin
$ echo "database version $(cat code/db.src)" > build/db_$(cat code/db.src).bin

# Deploy our buit SW to our server with
$ rm -f server/web_*; cp build/web_$(cat code/web.src).bin server/
$ rm -f server/db_*; cp build/db_$(cat code/db.src).bin server/
``` 

While going through the getting started guids, feel free to explore the
functionality by updating the source code, build and deploy new versions
to server.


### Use of environment variables

All the kosli commands contains some common
flags `--api-token` and `--owner`. By setting
these as environment variables we don't need to specify them. 

You can do that by capitalizing the flag in snake case and adding the `KOSLI_` prefix. For example, to set `--api-token` from an environment variable, you can export KOSLI_API_TOKEN, etc:

```shell
export KOSLI_API_TOKEN=<here you put in your kosli token>
export KOSLI_OWNER=<here you put in your github username>
```

To get the kosli token go to https://app.kosli.com, log in using your github account, and go to your Profile (you'll find it if you click on your avatar in the top right corner of the page)



# Environment

A Kosli environment is a repository for storing information about
what SW is running in your runtime environment (server, Kubernetes cluster, AWS, ...)
We use one Kosli environment per runtime environment. 

A typical setup can be that you report what is running on the 
staging server and on the production server. To report what is 
running you run the Kosli CLI command periodically. The Kosli CLI will
detect the version of the SW you are currently running and report
it to the Kosli environment.


## Create a Kosli environment

To follow the examples make sure you have followed the instructions in Local setup

We create a Kosli environment where we can report what SW are running on our server.
```shell
$ kosli environment declare \
    --name production \
    --environment-type server \
    --description "Production server (for kosli guide)"
```

You can immediately verify that the Kosli environment was created:
```shell
$ kosli environment ls
NAME        TYPE    LAST REPORT  LAST MODIFIED
production  server               2022-08-16T07:53:43+02:00

$ kosli environment get production
Name:              production
Type:              server
Description:       Production server (for kosli guide)
State:             INCOMPLIANT
Last Reported At:  16 Aug 22 07:58 CEST • 25 seconds ago
```


## Report the SW running in your environment

We simulate a report from our server by reporting two dummy files for the web and
database application.
```shell
$ kosli environment report server production --paths $(ls server/web_*bin),$(ls server/db_*.bin)
```

We can see that the current server SW was started, and for how long it has ran.
```shell
$ kosli snapshot ls production
SNAPSHOT  FROM                  TO   DURATION
1         16 Aug 22 07:54 CEST  now  11 seconds
```

We can get a more detailed view of the SW that is currently
running on the server.
```shell
$ kosli snapshot get production
COMMIT  ARTIFACT                                                                  PIPELINE  RUNNING_SINCE  REPLICAS
N/A     Name: /tmp/try-kosli/server/web_1.bin                                     N/A       2 minutes ago  1
        SHA256: a7a87c332500a40f9a01b811ec75f51b40188a3dabd205feb0fa7c3eafb25fbe                           
                                                                                                           
N/A     Name: /tmp/try-kosli/server/db_1.bin                                      N/A       2 minutes ago  1
        SHA256: 0efde582a933f011c3ae9007467a7f973a874517093e9a5a05ea55476f7c91af                           
```

Typically a server would report which SW that is running periodically. The Kosli app
generates a new snapshot if the SW changes, so resending the same environment report
several times will not lead to duplication of a snapshot.
```shell
$ kosli environment report server production --paths $(ls server/web_*bin),$(ls server/db_*.bin)
$ kosli snapshot get production
COMMIT  ARTIFACT                                                                  PIPELINE  RUNNING_SINCE  REPLICAS
N/A     Name: /tmp/try-kosli/server/web_1.bin                                     N/A       2 minutes ago  1
        SHA256: a7a87c332500a40f9a01b811ec75f51b40188a3dabd205feb0fa7c3eafb25fbe                           
                                                                                                           
N/A     Name: /tmp/try-kosli/server/db_1.bin                                      N/A       2 minutes ago  1
        SHA256: 0efde582a933f011c3ae9007467a7f973a874517093e9a5a05ea55476f7c91af                           
```

We simulate a change of the web application from version 1 to version 2
```shell
# Update src
$ echo "2" > code/web.src
$ cd code
$ git add web.src
$ git commit -m "Version two of web"
$ cd ..

# Build
$ echo web version $(cat code/web.src) > build/web_$(cat code/web.src).bin

# Deploy to server
$ rm -f server/web_*; cp build/web_$(cat code/web.src).bin server/

# Report what is now running on server
$ kosli environment report server production --paths $(ls server/web_*bin),$(ls server/db_*.bin)
```

We now see that we have created a new snapshot and that we are now running web version 2.
```shell
$ kosli snapshot ls production
SNAPSHOT  FROM                  TO                    DURATION
2         16 Aug 22 07:58 CEST  now                   9 seconds
1         16 Aug 22 07:54 CEST  16 Aug 22 07:58 CEST  4 minutes

$ kosli snapshot get production
COMMIT  ARTIFACT                                                                  PIPELINE  RUNNING_SINCE       REPLICAS
N/A     Name: /tmp/try-kosli/server/web_2.bin                                     N/A       39 seconds ago  1
        SHA256: cbc92ce1291830382ec23b95efc213d6e1725b5157bcb2927d48296b61c86746                            
                                                                                                            
N/A     Name: /tmp/try-kosli/server/db_1.bin                                      N/A       6 minutes ago   1
        SHA256: 0efde582a933f011c3ae9007467a7f973a874517093e9a5a05ea55476f7c91af                            
```                    

Using environment name always refer to the latest snapshot. We can use the 
Kosli CLI to check which version of the SW that was running in previous snapshots.
Here we look at what was running in snapshot #1 in production
```shell
$ kosli snapshot get production#1
COMMIT  ARTIFACT                                                                  PIPELINE  RUNNING_SINCE  REPLICAS
N/A     Name: /tmp/try-kosli/server/web_1.bin                                     N/A       7 minutes ago  1
        SHA256: a7a87c332500a40f9a01b811ec75f51b40188a3dabd205feb0fa7c3eafb25fbe                           
                                                                                                           
N/A     Name: /tmp/try-kosli/server/db_1.bin                                      N/A       7 minutes ago  1
        SHA256: 0efde582a933f011c3ae9007467a7f973a874517093e9a5a05ea55476f7c91af                           
```


# Pipelines

A Kosli pipeline is a repository for storing the results of your build system.
The output of the build system is called and *artifact* in Kosli. This can be
a application, docker image, documentation, filesystem and so on.

Some organizations have a CI system where one CI pipeline builds one 
artifact, some have a CI system where one CI pipeline builds several
artifacts. For both cases we use one Kosli pipeline for each artifact.
We use Kosli CLI to report information about the creation of an
artifact to the Kosli pipeline.

A Kosli pipeline can also be used to store any information related to 
the artifact that you have built. Like test results, manual approvals, 
pull-request and so on.


## Create a Kosli pipeline

To follow the examples make sure you have followed the instructions in Local setup

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

You can immediately verify that the Kosli pipelines were created:
```shell
$ kosli pipeline ls
NAME             DESCRIPTION                        VISIBILITY
database-server  pipeline to build database-server  private
web-server       pipeline to build web-server       private
```


## Build artifacts and report them to Kosli

We "build" some SW based on the source code
```shell
$ echo "web version $(cat code/web.src)" > build/web_$(cat code/web.src).bin
$ echo "database version $(cat code/db.src)" > build/db_$(cat code/db.src).bin
```

We can now report that we have built the web and database applications
```shell
$ kosli pipeline artifact report creation build/web_$(cat code/web.src).bin \
    --pipeline web-server \
    --artifact-type file \
    --build-url link_to_your_ci_system \
    --commit-url link_to_your_source_repository \
    --git-commit $(cd code; git rev-parse HEAD)

$ kosli pipeline artifact report creation build/db_$(cat code/db.src).bin \
    --pipeline database-server \
    --artifact-type file \
    --build-url link_to_your_ci_system \
    --commit-url link_to_your_source_repository \
    --git-commit $(cd code; git rev-parse HEAD)
```

We can see that we have built one artifact in our *web-server* pipeline
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
Build URL:    link_to_your_ci_system
Commit URL:   link_to_your_source_repository
Created at:   16 Aug 22 08:01 CEST • 2 hours ago
Approvals:    None
Deployments:  None
Evidence:
```


# Deployments

To link the artifact you report to Kosli pipeline to the
artifact you reported running to a Kosli environment we use a Kosli
deployment. So a Kosli deployment is telling which environment
a given aritfact is supose to run in.

We use the Kosli CLI to report when we deploy an artifact to
a given environment.


## Deploy SW to server and report the deployment to Kosli

We assume the user has done both Environments and Pipelines first.

We "deploy" our SW by copying it over to the server
```shell
$ cp build/web_$(cat code/web.src).bin server/
$ cp build/db_$(cat code/db.src).bin server/
```

Now we report to Kosli that the SW has been deployed
```shell
$ kosli pipeline deployment report  build/web_$(cat code/web.src).bin \
    --pipeline web-server \
    --artifact-type file \
    --build-url link_to_your_ci_system \
    --environment production \
    --description "Web server version $(cat code/web.src)"
```

We can verify the deployment with
```shell
$ kosli deployment ls web-server
ID   ARTIFACT                                                                  ENVIRONMENT  REPORTED_AT
1    Name: web_2.bin                                                           production   16 Aug 22 08:02 CEST
     SHA256: cbc92ce1291830382ec23b95efc213d6e1725b5157bcb2927d48296b61c86746               

$ kosli deployment get --pipeline web-server 1
ID:               1
Artifact SHA256:  cbc92ce1291830382ec23b95efc213d6e1725b5157bcb2927d48296b61c86746
Artifact name:    web_2.bin
Build URL:        link_to_your_ci_system
Created at:       16 Aug 22 08:02 CEST • 32 seconds ago
Environment:      production
Runtime state:    The artifact running since 16 Aug 22 07:58 CEST
```


# For developers
You can extract all the commands to execute from this document by running
```shell
cat docs.kosli.com/content/get_familiar_with_Kosli/_index.md | sed -e :a -e '/\\$/N; s/\\\n//; ta' | egrep '^\$ ' | sed "s/^..//" 
```