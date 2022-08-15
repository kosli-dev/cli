---
title: Get familiar with Kosli
bookCollapseSection: true
weight: 1
---

# Get familiar with Kosli

Here you will learn what is Kosli and how to start using it. 
You don't need CI system to start with. Local code (and a git repository later) and terminal are enough. 

### Things to prepare

#### CLI

The `kosli` tool can be downloaded from: https://github.com/kosli-dev/cli/releases
Put it in a location you'll be running it from (as `./kosli`) or add it to your PATH so you can use it anywhere (as `kosli`)

#### Local setup

Create folders for your code and "server"

```
$ mkdir -p try-kosli/server try-kosli/code
``` 
Put files in server, so you can see "what's running" in your environment in the next stage of this quide. We'll pretend we're running a web service that uses a local database, so we expect to see two services running in our environment:

```
$ cd try-kosli/server
$ echo "web version 1" > web.bin
$ echo "db version 1" > db.bin
``` 

#### Use of environment variables

All the kosli commands contains some common
flags `--api-token` and `--owner`. By setting
these as environment variables we don't need to specify them. 

You can do that by capitalizing the flag in snake case and adding the `KOSLI_` prefix. For example, to set `--api-token` from an environment variable, you can export KOSLI_API_TOKEN, etc:

```shell
export KOSLI_API_TOKEN=<here you put in your kosli token>
export KOSLI_OWNER=<here you put in your github username>
```

To get the kosli token go to https://app.kosli.com, log in using your github account, and go to your Profile (you'll find it if you click on your avatar in the top right corner of the page)

## Environment

Explain the concept

### Create an environment

```
$ kosli environment declare \
    --name production \
    --environment-type server \
    --description "Production server (for kosli guide)"
```
And you can immediately verify that the environment was created:
```
$ kosli environment ls
NAME        TYPE    LAST REPORT  LAST MODIFIED
production  server               2022-08-15T10:08:08+02:00
```

### Report an environment
```
$ kosli environment report server production --paths server/web.bin,server/db.bin
```

How to see what's running then?
```
$ kosli snapshot ls production
SNAPSHOT  FROM                  TO   DURATION
1         15 Aug 22 10:25 CEST  now  4 seconds

$ kosli snapshot get production
COMMIT  ARTIFACT                                                                  PIPELINE  RUNNING_SINCE   REPLICAS
N/A     Name: /Users/try-kosli/server/db.bin                       N/A       23 minutes ago  1
        SHA256: 0e0c5f77db75f32bbe908b70e15f4cb02921d0334bb521775f3c3c3c244df477                            
                                                                                                            
N/A     Name: /Users/try-kosli/server/web.bin                      N/A       23 minutes ago  1
        SHA256: a7a87c332500a40f9a01b811ec75f51b40188a3dabd205feb0fa7c3eafb25fbe     
```


Explain how to see it - terminal and show ls/get commands on the way

## Pipelines

Explain the concept

- declare a pipeline 
- report an artifact
- report evidences
- request/report approvals (?)

Show some get/ls commands on the way (pipeline, artifact, etc)

## Deployments

Explain the concept

- report a deployment to an env

Show get/ls commands


