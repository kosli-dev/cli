---
title: Simulating a DevOps system
bookCollapseSection: false
weight: 530
---

# Simulating a DevOps system

## Pre-requisites 

To follow the simulation you need to:
* [Install the `kosli` CLI](/kosli_overview/kosli_tools/#installing-the-kosli-cli).
* [Get your Kosli API token](/kosli_overview/kosli_tools/#getting-your-kosli-api-token).
* Set the KOSLI_API_TOKEN environment variable:
  ```shell {.command}
  export KOSLI_API_TOKEN=<paste-your-kosli-API-token-here>
  ```
* Set the KOSLI_ORG environment variable to your Kosli organization name:
  ```shell {.command}
  export KOSLI_ORG=<paste-your-kosli-organization-name>
  ```
## Overview

You will simulate a system with source code in a git repo, building a web service
and a database service, and deploying them both to a server.

There is a script to help you run these simulations, so you won't need to type too many commands.
You can download the script from here:
https://raw.githubusercontent.com/kosli-dev/cli/main/simulation_commands.bash

or from the command line with:

```shell {.command}
cd /tmp
curl -O https://raw.githubusercontent.com/kosli-dev/cli/main/simulation_commands.bash
```

Source the simulation commands so you can use them later on in the examples:

```shell {.command}
source simulation_commands.bash
```

Use `type` to see what a simulation command is actually doing (in zsh you will need `type -f`):

```shell {.command}
type create_git_repo_in_tmp
```
```plaintext {.light-console}
create_git_repo_in_tmp ()
{ 
    pushd /tmp &> /dev/null
    mkdir try-kosli
    ...
	popd &> /dev/null
}
```

Create the git repo and simulate a build and deployment to the server:

```shell {.command}
create_git_repo_in_tmp
simulate_build
simulate_deployment
``` 

Feel free to explore the functionality by updating the source code, 
building and deploying new versions.


{{< hint info >}}
## Using a web browser

As you go through the guide you can also check your progress from 
[your browser](https://app.kosli.com).

In the upper left corner there is a house icon. Next to it you can select
which organization you want to view. Your personal organization
has the same name as your github login name, and is the organization you will
be using in this guide.
{{< /hint >}}


# Environments

A Kosli *Environment* stores information about
what software is running in your actual runtime environment (server, Kubernetes cluster, AWS, etc.)

A typical setup reports what is running on a 
staging server and also on a production server. To report what is 
running you execute the Kosli CLI command periodically. The Kosli CLI will
detect precisely what software is currently running and report
it to the Kosli *Environment*.


## Creating a Kosli Environment

Create a Kosli *Environment*:

```shell {.command}
kosli create environment production \
    --type server \
    --description "Production server (for kosli getting started)"
```

Verify the Kosli Environment was created:

```shell {.command}
kosli ls environments
```

```plaintext {.light-console}
NAME        TYPE    LAST REPORT  LAST MODIFIED
production  server               2022-08-16T07:53:43+02:00
```

Verify there are no reports to it yet:

```shell {.command}
kosli get environment production
```

```plaintext {.light-console}
Name:              production
Type:              server
Description:       Production server (for kosli getting started)
State:             N/A
Last Reported At:  N/A
```


In the web interface you can select the **Environments** menu on the left.
It will show you that you have a *production* environment and that
no reports have been received.


## Reporting the software running in your environment

Simulate a report from your server by reporting two dummy files for the web and
database applications:

```shell {.command}
kosli snapshot server production \
    --paths /tmp/try-kosli/server/web_*.bin \
    --paths /tmp/try-kosli/server/db_*.bin
```

You can see that the server has started, and how long it has been running for:

```shell {.command}
kosli list snapshots production
```
```plaintext {.light-console}
SNAPSHOT  FROM                  TO   DURATION    COMPLIANT
1         16 Aug 22 07:54 CEST  now  11 seconds  false
```

Get a more detailed view of what is currently running on the server:

```shell {.command}
kosli get snapshot production
```
```plaintext {.light-console}
COMMIT  ARTIFACT                                                                       FLOW      RUNNING_SINCE  REPLICAS
N/A     Name: /tmp/try-kosli/server/web_1.bin                                          N/A       2 minutes ago  1
        Fingerprint: a7a87c332500a40f9a01b811ec75f51b40188a3dabd205feb0fa7c3eafb25fbe                           
                                                                                                           
N/A     Name: /tmp/try-kosli/server/db_1.bin                                           N/A       2 minutes ago  1
        Fingerprint: 0efde582a933f011c3ae9007467a7f973a874517093e9a5a05ea55476f7c91af                           
```

If you refresh the environment page in the web browser you can see that there is
a timestamp for when the environment changed. Pressing the *production* link
gives you a detailed view of what is running now.

Typically, a server periodically sends a report of what is currently running to Kosli. But Kosli
will only create a new snapshot if the report shows changes compared to the previous snapshot, so 
resending the same environment report several times will not lead to duplication of snapshots.

Resend the environment report:

```shell {.command}
kosli snapshot server production \
    --paths /tmp/try-kosli/server/web_*.bin \
    --paths /tmp/try-kosli/server/db_*.bin
```

Confirm there is still a single snapshot:

```shell {.command}
kosli list snapshots production
```

```plaintext {.light-console}
SNAPSHOT  FROM                  TO   DURATION    COMPLIANT
1         16 Aug 22 07:54 CEST  now  11 seconds  false
```

Simulate an update of the web application to a new version, build and deploy it:

```shell {.command}
update_web_src
simulate_build
simulate_deployment
```

Report what is now running on the server:

```shell {.command}
kosli snapshot server production \
    --paths /tmp/try-kosli/server/web_*.bin \
    --paths /tmp/try-kosli/server/db_*.bin
```

You can see Kosli has created a new snapshot:

```shell {.command}
kosli list snapshots production
```

```plaintext {.light-console}
SNAPSHOT  FROM                  TO                    DURATION
2         16 Aug 22 07:58 CEST  now                   9 seconds
1         16 Aug 22 07:54 CEST  16 Aug 22 07:58 CEST  4 minutes
```

You can see that you are currently running web version 2 in production:

```shell {.command}
kosli get snapshot production
```

```plaintext {.light-console}
COMMIT  ARTIFACT                                                                       FLOW      RUNNING_SINCE   REPLICAS
N/A     Name: /tmp/try-kosli/server/web_2.bin                                          N/A       39 seconds ago  1
        Fingerprint: cbc92ce1291830382ec23b95efc213d6e1725b5157bcb2927d48296b61c86746                            
                                                                                                            
N/A     Name: /tmp/try-kosli/server/db_1.bin                                           N/A       6 minutes ago   1
        Fingerprint: 0efde582a933f011c3ae9007467a7f973a874517093e9a5a05ea55476f7c91af                            
```                    

Here, using the bare environment name (eg production) always refers to the latest snapshot
in that Environment (currently #2). You can also use the 
Kosli CLI to check what was running in previous snapshots.

Find what was running in snapshot #1 in production:

```shell {.command}
kosli get snapshot production#1
```

```plaintext {.light-console}
COMMIT  ARTIFACT                                                                       FLOW      RUNNING_SINCE  REPLICAS
N/A     Name: /tmp/try-kosli/server/web_1.bin                                          N/A       7 minutes ago  1
        Fingerprint: a7a87c332500a40f9a01b811ec75f51b40188a3dabd205feb0fa7c3eafb25fbe                           
                                                                                                           
N/A     Name: /tmp/try-kosli/server/db_1.bin                                           N/A       7 minutes ago  1
        Fingerprint: 0efde582a933f011c3ae9007467a7f973a874517093e9a5a05ea55476f7c91af                           
```

<!---
TODO: add icon of the log tab in the text below
-->

In the web interface you should now also be able to see 2 snapshots. The Log
tab should show what changed in snapshot 1 and snapshot 2.


# Flows

For this tutorial we are simulating building and deploying two
artifacts; a web-server and a db-server.

## Creating Kosli Flows

When attesting evidence, the target of the attestation must be named.
These names are defined in a yml file.
Prepared yml files already exist in the git repository, one
for the web-server, called `web.yml`, and one for the db-server,
called `db.yml`.


<!--
A Kosli flow stores information about what happens in your build system.
The output of the build system is called an *artifact* in Kosli. This can be
an application, a docker image, documentation, a filesystem, etc.

You use the Kosli CLI to report information about the creation of an
artifact to the Kosli flow.

A Kosli flow can also be used to store any information related to 
the artifact you have built, like test results, manual approvals, 
pull-requests, and so on.
-->

You are building two applications, so create
two Kosli Flows called `web-server` and `database-server`
specifying these two yml files:


```shell {.command}
kosli create flow2 web-server \
    --description "flow to build web-server" \
    --template-file try-kosli/web.yml
```

```shell {.command}
kosli create flow2 database-server \
    --description "flow to build database-server" \
    --template-file try-kosli/db.yml
```

You can immediately verify that the Kosli flows were created:

```shell {.command}
kosli ls flows
```

```plaintext {.light-console}
NAME             DESCRIPTION                    VISIBILITY
database-server  flow to build database-server  private
web-server       flow to build web-server       private
```

In the web interface you can select the **Flows** menu on the left.
It will show you that you have a *web-server* and *database-server* flow.
If you select either of the flows they will show that no artifacts have
been reported for the flows.

# Trails
...

## Creating Kosli Trails
...

## Building artifacts and reporting them to Kosli

Simulate building your software:

```shell {.command}
simulate_build
```

Report that you have built the web and database applications. You are using
a dummy `--build-url`, but in reality it would be a CI build URL:

```shell {.command}
kosli report artifact /tmp/try-kosli/build/web_$(cat /tmp/try-kosli/code/web.src).bin \
    --flow web-server \
    --artifact-type file \
    --build-url file://dummy \
    --commit-url file:///tmp/try-kosli/code \
    --repo-root /tmp/try-kosli/code \
    --git-commit $(cd /tmp/try-kosli/code; git rev-parse HEAD)
```

```shell {.command}
kosli report artifact /tmp/try-kosli/build/db_$(cat /tmp/try-kosli/code/db.src).bin \
    --flow database-server \
    --artifact-type file \
    --build-url file://dummy \
    --commit-url file:///tmp/try-kosli/code \
    --repo-root /tmp/try-kosli/code \
    --git-commit $(cd /tmp/try-kosli/code; git rev-parse HEAD)
```

You can see you have built one artifact in your *web-server* flow:

```shell {.command}
kosli ls artifacts --flow web-server
```

```plaintext {.light-console}
COMMIT   ARTIFACT                                                                       STATE      CREATED_AT
5187374  Name: web_2.bin                                                                COMPLIANT  16 Aug 22 08:00 CEST
         Fingerprint: cbc92ce1291830382ec23b95efc213d6e1725b5157bcb2927d48296b61c86746             
```

And one for the *database-server* flow:

```shell {.command}
kosli ls artifacts --flow database-server
```

```plaintext {.light-console}
COMMIT   ARTIFACT                                                                        STATE      CREATED_AT
5187374  Name: db_1.bin                                                                  COMPLIANT  16 Aug 22 08:01 CEST
         Fingerprint: 0efde582a933f011c3ae9007467a7f973a874517093e9a5a05ea55476f7c91af             
```

You can also get detailed information about each artifact that has been reported:

```shell {.command}
kosli get artifact database-server@0efde582a933f011c3ae9007467a7f973a874517093e9a5a05ea55476f7c91af
```

```plaintext {.light-console}
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

In the web interface you can select the *database-server* flow and then the *db_1.bin*
artifact to get more details.


# Deployments

The Kosli expect deployment command is used to indicate an artifact is
about to be deployed to a given runtime environment. 


## Deploying software to the server and reporting the deployment to Kosli

Report to Kosli that the web software is expected to be deployed:

```shell {.command}
kosli expect deployment /tmp/try-kosli/build/web_$(cat /tmp/try-kosli/code/web.src).bin \
    --flow web-server \
    --artifact-type file \
    --build-url file://dummy \
    --environment production \
    --description "Web server version $(cat /tmp/try-kosli/code/web.src)"
```

Simulate deploying your software to the server:

```shell {.command}
simulate_deployment
```

You can verify the deployment with:

```shell {.command}
kosli ls deployments --flow web-server
```

```plaintext {.light-console}
ID   ARTIFACT                                                                        ENVIRONMENT  REPORTED_AT
1    Name: web_2.bin                                                                 production   16 Aug 22 08:02 CEST
     Fingerprint: cbc92ce1291830382ec23b95efc213d6e1725b5157bcb2927d48296b61c86746               
```

Get detailed information about a deployment:

```shell {.command}
kosli get deployment web-server#1
```

```plaintext {.light-console}
ID:                    1
Artifact fingerprint:  cbc92ce1291830382ec23b95efc213d6e1725b5157bcb2927d48296b61c86746
Artifact name:         web_2.bin
Build URL:             file://dummy
Created at:            16 Aug 22 08:02 CEST • 32 seconds ago
Environment:           production
Runtime state:         The artifact running since 16 Aug 22 07:58 CEST
```

If you select the *web_2.bin* artifact in the web interface it will show
that it was part of Deployment #1 to *production* environment.
