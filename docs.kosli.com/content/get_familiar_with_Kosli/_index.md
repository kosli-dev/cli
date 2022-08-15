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

For these examples we have a very basic settup with source code, build system and
a server "running" the SW.

```
$ mkdir try-kosli
$ cd try-kosli
$ mkdir code server

# Make our source code
$ echo "1" > code/web.src
$ echo "1" > code/db.src

# We build our SW with
$ echo web version $(cat code/web.src) > code/web_$(cat code/web.src).bin
$ echo database version $(cat code/db.src) > code/db_$(cat code/db.src).bin

# We deploy our buit SW to our server with
$ rm server/web_*; cp code/web_$(cat code/web.src).bin server/
$ rm server/db_*; cp code/db_$(cat code/db.src).bin server/
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

To follow the examples make sure you have followed the instructions in Local setup

Explain the concept

### Create a Kosli environment

We create a Kosli environment where we can report what SW are running on our server.
```
$ kosli environment declare \
    --name production \
    --environment-type server \
    --description "Production server (for kosli guide)"
```

You can immediately verify that the Kosli environment was created:
```
$ kosli environment ls
NAME        TYPE    LAST REPORT  LAST MODIFIED
production  server               2022-08-15T10:08:08+02:00
```


### Report the SW running in your environment

We simulate a report from our server by reporting two dummy files for the web and
database application.
```
$ kosli environment report server production --paths $(ls server/web_*bin),$(ls server/db_*.bin)
```

We can see that the current server SW was started, and for how long it has ran.
```
$ kosli snapshot ls production
SNAPSHOT  FROM                  TO   DURATION
1         15 Aug 22 10:25 CEST  now  4 seconds
```

We can get a more detailed view of the SW that is currently
running on the server.
```
$ kosli snapshot get production
COMMIT  ARTIFACT                                                                  PIPELINE  RUNNING_SINCE  REPLICAS
N/A     Name: /tmp/try-kosli/server/db_1.bin                                      N/A       8 minutes ago  1
        SHA256: 0e0c5f77db75f32bbe908b70e15f4cb02921d0334bb521775f3c3c3c244df477                           
                                                                                                           
N/A     Name: /tmp/try-kosli/server/web_1.bin                                     N/A       8 minutes ago  1
        SHA256: a7a87c332500a40f9a01b811ec75f51b40188a3dabd205feb0fa7c3eafb25fbe                           
```

Typically a server would report which SW that is running periodically. The Kosli app
generates a new snapshot if the SW changes, so resending the same environment report
several times will not resolve in duplication of a snapshot.
```
$ kosli environment report server production --paths $(ls server/web_*bin),$(ls server/db_*.bin)
$ kosli snapshot get production
COMMIT  ARTIFACT                                                                  PIPELINE  RUNNING_SINCE   REPLICAS
N/A     Name: /tmp/try-kosli/server/db_1.bin                                      N/A       10 minutes ago  1
        SHA256: 0e0c5f77db75f32bbe908b70e15f4cb02921d0334bb521775f3c3c3c244df477                            
                                                                                                            
N/A     Name: /tmp/try-kosli/server/web_1.bin                                     N/A       10 minutes ago  1
        SHA256: a7a87c332500a40f9a01b811ec75f51b40188a3dabd205feb0fa7c3eafb25fbe                            
```

We simulate a change of the web application from version 1 to version 2
```
$ echo 2 > code/web.src
$ echo web version $(cat code/web.src) > code/web_$(cat code/web.src).bin
$ kosli environment report server production --paths $(ls server/web_*bin),$(ls server/db_*.bin)
```

We now see that we have created a new snapshot and that we are now running web version 2.
```
$ kosli snapshot ls production
SNAPSHOT  FROM                  TO                    DURATION
2         15 Aug 22 13:44 CEST  now                   11 seconds
1         15 Aug 22 13:30 CEST  15 Aug 22 13:44 CEST  13 minutes

$ kosli snapshot get production
COMMIT  ARTIFACT                                                                  PIPELINE  RUNNING_SINCE       REPLICAS
N/A     Name: /tmp/try-kosli/server/web_2.bin                                     N/A       about a minute ago  1
        SHA256: cbc92ce1291830382ec23b95efc213d6e1725b5157bcb2927d48296b61c86746                                
                                                                                                                
N/A     Name: /tmp/try-kosli/server/db_1.bin                                      N/A       15 minutes ago      1
        SHA256: 0e0c5f77db75f32bbe908b70e15f4cb02921d0334bb521775f3c3c3c244df477                                
```                    

Using environment name always refer to the latest snapshot. We can use the 
Kosli CLI to check which version of the SW that was running in previous snapshots.
Here we look at what was running in snapshot #1 in production
```
$ kosli snapshot get production#1
COMMIT  ARTIFACT                                                                  PIPELINE  RUNNING_SINCE   REPLICAS
N/A     Name: /tmp/try-kosli/server/db_1.bin                                      N/A       18 minutes ago  1
        SHA256: 0e0c5f77db75f32bbe908b70e15f4cb02921d0334bb521775f3c3c3c244df477                            
                                                                                                            
N/A     Name: /tmp/try-kosli/server/web_1.bin                                     N/A       19 minutes ago  1
        SHA256: a7a87c332500a40f9a01b811ec75f51b40188a3dabd205feb0fa7c3eafb25fbe                          
```



## Pipelines

Explain the concept

To follow the examples make sure you have followed the instructions in Local setup

### Create a Kosli pipeline

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

You can immediately verify that the Kosli pipeline was created:

```
$ kosli pipeline ls
NAME             DESCRIPTION                        VISIBILITY
database-server  pipeline to build database-server  private
web-server       pipeline to build web-server       private
```


### Build artifacts and report them to Kosli

We "build" some new SW based on the source code
```
$ echo web version $(cat code/web.src) > code/web_$(cat code/web.src).bin
$ echo database version $(cat code/db.src) > code/db_$(cat code/db.src).bin
```

We can now report that we have built the web database applications
```
$ kosli pipeline artifact report creation code/web_$(cat code/web.src).bin \
    --pipeline web-server \
    --artifact-type file \
    --build-url link_to_your_ci_system \
    --commit-url link_to_your_source_repository \
    --git-commit fdac8a54

$ kosli pipeline artifact report creation code/db_$(cat code/db.src).bin \
    --pipeline database-server \
    --artifact-type file \
    --build-url link_to_your_ci_system \
    --commit-url link_to_your_source_repository \
    --git-commit fdac8a54
```

We can see that we have built one artifact in our *web-server* pipeline
```
$ kosli artifact ls web-server
COMMIT   ARTIFACT                                                                  STATE      CREATED_AT
fdac8a5  Name: web_1.bin                                                           COMPLIANT  15 Aug 22 14:23 CEST
         SHA256: a7a87c332500a40f9a01b811ec75f51b40188a3dabd205feb0fa7c3eafb25fbe             
```

We do the same for the database
```
$ kosli artifact ls database-server
COMMIT   ARTIFACT                                                                  STATE      CREATED_AT
fdac8a5  Name: db_1.bin                                                            COMPLIANT  15 Aug 22 14:35 CEST
         SHA256: 0efde582a933f011c3ae9007467a7f973a874517093e9a5a05ea55476f7c91af             
```


## Deployments

We assume the user has done both Environments and Pipelines first.

Explain the concept

### Deploy SW to server and report the deployment to Kosli

We "deploy" our SW by copying it over to the server
```
$ cp code/web_$(cat code/web.src).bin server/
$ cp code/db_$(cat code/db.src).bin server/
```

Now we report to Kosli that the SW has been deployed
```
$ kosli pipeline deployment report  code/web_$(cat code/web.src).bin \
    --pipeline web-server \
    --artifact-type file \
    --build-url link_to_your_ci_system \
    --environment production \
    --description "Web server version $(cat code/web.src)"
```

We can verify the deployment with
```
$ kosli deployment ls web-server
ID   ARTIFACT                                                                  ENVIRONMENT  REPORTED_AT
1    Name: web_1.bin                                                           production   15 Aug 22 14:47 CEST
     SHA256: a7a87c332500a40f9a01b811ec75f51b40188a3dabd205feb0fa7c3eafb25fbe               

$ kosli deployment get --pipeline web-server 1
ID:               1
Artifact SHA256:  a7a87c332500a40f9a01b811ec75f51b40188a3dabd205feb0fa7c3eafb25fbe
Artifact name:    web_1.bin
Build URL:        link_to_your_ci_system
Created at:       15 Aug 22 14:47 CEST • 2 minutes ago
Environment:      production
Runtime state:    The artifact exited on 15 Aug 22 13:44 CEST
```
