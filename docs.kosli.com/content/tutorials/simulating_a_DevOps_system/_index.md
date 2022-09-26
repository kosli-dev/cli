---
title: Simulating a DevOps system
bookCollapseSection: false
weight: 3
---

# Preparing the tutorial
To follow the tutorial you need to:
* [Install the `kosli` CLI](../../installation).
* [Get your Kosli API token](../../installation#getting-your-kosli-api-token).
* Set the KOSLI_API_TOKEN environment variable:
  ```shell {.command}
  export KOSLI_API_TOKEN=<paste-your-kosli-API-token-here>
  ```
* Set the KOSLI_OWNER environment variable to your Kosli organization name:
  ```shell {.command}
  export KOSLI_OWNER=<paste-your-kosli-organization-name>
  ```

For this tutorial you will simulate a system with source code, a build system, and a running server.
There is a script to help you run these simulations so you don't need to type so many commands.

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

To see what a command (eg create_git_repo_in_tmp) in the simulation is actually doing run:

```shell {.command}
type create_git_repo_in_tmp
```
```plaintext {.light-console}
create_git_repo_in_tmp is a function
create_git_repo_in_tmp () 
{ 
    pushd /tmp;
    mkdir try-kosli;
    ...
```

Create the git repo and simulate a build and deployment to server:

```shell {.command}
create_git_repo_in_tmp
simulate_build
simulate_deployment
``` 

While going through the getting started guide, feel free to explore the
functionality by updating the source code, building and deploying new versions.


{{< hint info >}}
## Using a web browser

As you go through the guide you can also check your progress from 
[your browser](https://app.kosli.com).

In the upper left corner there is a house icon. Next to it you can select
which organization you want to view. Your personal organization
has the same name as your github login name, and is the organization you will
be using in this guide.
{{< /hint >}}



# Environment

A Kosli environment stores information about
what software is running in your actual runtime environment (server, Kubernetes cluster, AWS, ...)

A typical setup reports what is running on the 
staging server and on the production server. To report what is 
running you run the Kosli CLI command periodically. The Kosli CLI will
detect the version of the software you are currently running and report
it to the Kosli environment.


## Creating a Kosli environment

To follow the examples make sure you have followed the instructions in Local setup

Create a Kosli environment:

```shell {.command}
kosli environment declare \
    --name production \
    --environment-type server \
    --description "Production server (for kosli getting started)"
```

You can immediately verify that the Kosli environment was created:

```shell {.command}
kosli environment ls
```

```plaintext {.light-console}
NAME        TYPE    LAST REPORT  LAST MODIFIED
production  server               2022-08-16T07:53:43+02:00
```

```shell {.command}
kosli environment inspect production
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
kosli environment report server production \
    --paths /tmp/try-kosli/server/web_*bin \
    --paths /tmp/try-kosli/server/db_*.bin
```

You can see that the server has started, and how long it has run:

```shell {.command}
kosli environment log production
```
```plaintext {.light-console}
SNAPSHOT  FROM                  TO   DURATION
1         16 Aug 22 07:54 CEST  now  11 seconds
```

Get a more detailed view of what is currently running on the server:

```shell {.command}
kosli environment get production
```
```plaintext {.light-console}
COMMIT  ARTIFACT                                                                       PIPELINE  RUNNING_SINCE  REPLICAS
N/A     Name: /tmp/try-kosli/server/web_1.bin                                          N/A       2 minutes ago  1
        Fingerprint: a7a87c332500a40f9a01b811ec75f51b40188a3dabd205feb0fa7c3eafb25fbe                           
                                                                                                           
N/A     Name: /tmp/try-kosli/server/db_1.bin                                           N/A       2 minutes ago  1
        Fingerprint: 0efde582a933f011c3ae9007467a7f973a874517093e9a5a05ea55476f7c91af                           
```

If you refresh the environment page in the web browser you can see that there is
a time-stamp for when the environment changed. Pressing the *production* link
gives you a detailed view of what is running now.

Typically a server periodically sends a report of what is currently running to Kosli. Kosli
only creates a new snapshot if the report has changes compared to previous snapshot, so resending the same environment report
several times will not lead to duplication of snapshots.

Send an environment report:

```shell {.command}
kosli environment report server production \
    --paths /tmp/try-kosli/server/web_*bin \
    --paths /tmp/try-kosli/server/db_*.bin
```
```shell {.command}
kosli environment log production
```
```plaintext {.light-console}
SNAPSHOT  FROM                  TO   DURATION
1         16 Aug 22 07:54 CEST  now  11 seconds
```

Simulate an update of the web application to a new version, build and deploy it:

```shell {.command}
update_web_src
simulate_build
simulate_deployment
```

Report what is now running on server:

```shell {.command}
kosli environment report server production \
    --paths /tmp/try-kosli/server/web_*bin \
    --paths /tmp/try-kosli/server/db_*.bin
```

You can see Kosli has created a new snapshot:

```shell {.command}
kosli environment log production
```

```plaintext {.light-console}
SNAPSHOT  FROM                  TO                    DURATION
2         16 Aug 22 07:58 CEST  now                   9 seconds
1         16 Aug 22 07:54 CEST  16 Aug 22 07:58 CEST  4 minutes
```

You can see that you are currently running web version 2 in production:

```shell {.command}
kosli environment get production
```

```plaintext {.light-console}
COMMIT  ARTIFACT                                                                       PIPELINE  RUNNING_SINCE   REPLICAS
N/A     Name: /tmp/try-kosli/server/web_2.bin                                          N/A       39 seconds ago  1
        Fingerprint: cbc92ce1291830382ec23b95efc213d6e1725b5157bcb2927d48296b61c86746                            
                                                                                                            
N/A     Name: /tmp/try-kosli/server/db_1.bin                                           N/A       6 minutes ago   1
        Fingerprint: 0efde582a933f011c3ae9007467a7f973a874517093e9a5a05ea55476f7c91af                            
```                    

Here, using the bare environment name (eg production) always refers to the latest snapshot
in that environment. You can also use the 
Kosli CLI to check what was running in previous snapshots.
Find what was running in snapshot #1 in production:

```shell {.command}
kosli environment get production#1
```

```plaintext {.light-console}
COMMIT  ARTIFACT                                                                       PIPELINE  RUNNING_SINCE  REPLICAS
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


# Pipelines

A Kosli pipeline stores information about what happens in your build system.
The output of the build system is called an *artifact* in Kosli. This can be
an application, docker image, documentation, filesystem and so on.

<!-- TODO: Do we need this??
Some organizations have a CI system where one CI pipeline builds one 
artifact, some have a CI system where one CI pipeline builds several
artifacts. For both cases you use one Kosli pipeline for each artifact. -->

You use the Kosli CLI to report information about the creation of an
artifact to the Kosli pipeline.

A Kosli pipeline can also be used to store any information related to 
the artifact you have built, like test results, manual approvals, 
pull-requests, and so on.


## Creating a Kosli pipeline

To follow the examples make sure you have followed the instructions in Local setup.

Create a Kosli pipeline where you can report what software your CI system
is building. Since you are building two applications you are making
two Kosli pipelines `web-server` and `database-server`.

Create your new pipelines:

```shell {.command}
kosli pipeline declare \
    --pipeline web-server \
    --description "pipeline to build web-server" \
    --visibility private \
    --template artifact
```

```shell {.command}
kosli pipeline declare \
    --pipeline database-server \
    --description "pipeline to build database-server" \
    --visibility private \
    --template artifact
```

You can immediately verify the Kosli pipelines were created:

```shell {.command}
kosli pipeline ls
```

```plaintext {.light-console}
NAME             DESCRIPTION                        VISIBILITY
database-server  pipeline to build database-server  private
web-server       pipeline to build web-server       private
```

In the web interface you can select the **Pipelines** menu on the left.
It will show you that you have a *web-server* and *database-server* pipeline.
If you press either of the pipelines they will show that no artifacts have
been reported for the pipelines.


## Building artifacts and reporting them to Kosli

Simulate building your software:

```shell {.command}
simulate_build
```

Report you have built the web and database applications. You are using
a dummy `--build-url`, in real life it would be a CI build URL:

```shell {.command}
kosli pipeline artifact report creation /tmp/try-kosli/build/web_$(cat /tmp/try-kosli/code/web.src).bin \
    --pipeline web-server \
    --artifact-type file \
    --build-url file://dummy \
    --commit-url file:///tmp/try-kosli/code \
    --git-commit $(cd /tmp/try-kosli/code; git rev-parse HEAD)
```

```shell {.command}
kosli pipeline artifact report creation /tmp/try-kosli/build/db_$(cat /tmp/try-kosli/code/db.src).bin \
    --pipeline database-server \
    --artifact-type file \
    --build-url file://dummy \
    --commit-url file:///tmp/try-kosli/code \
    --git-commit $(cd /tmp/try-kosli/code; git rev-parse HEAD)
```

You can see you have built one artifact in your *web-server* pipeline:

```shell {.command}
kosli artifact ls web-server
```

```plaintext {.light-console}
COMMIT   ARTIFACT                                                                       STATE      CREATED_AT
5187374  Name: web_2.bin                                                                COMPLIANT  16 Aug 22 08:00 CEST
         Fingerprint: cbc92ce1291830382ec23b95efc213d6e1725b5157bcb2927d48296b61c86746             
```

And one for the *database-server* pipeline:

```shell {.command}
kosli artifact ls database-server
```

```plaintext {.light-console}
COMMIT   ARTIFACT                                                                        STATE      CREATED_AT
5187374  Name: db_1.bin                                                                  COMPLIANT  16 Aug 22 08:01 CEST
         Fingerprint: 0efde582a933f011c3ae9007467a7f973a874517093e9a5a05ea55476f7c91af             
```

You can also get detailed information about each artifact that has been reported:

```shell {.command}
kosli artifact get database-server@0efde582a933f011c3ae9007467a7f973a874517093e9a5a05ea55476f7c91af
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

In the web interface you can select the *database-server* pipeline and then the *db_1.bin*
artifact to get more details.


# Deployments

A Kosli deployment command is used to indicate an artifact is
being deployed to a given runtime environment. 


## Deploying software to the server and reporting the deployment to Kosli

Simulate deploying your software to the server:

```shell {.command}
simulate_deployment
```

Report to Kosli that the web software has been deployed:

```shell {.command}
kosli pipeline deployment report /tmp/try-kosli/build/web_$(cat /tmp/try-kosli/code/web.src).bin \
    --pipeline web-server \
    --artifact-type file \
    --build-url file://dummy \
    --environment production \
    --description "Web server version $(cat /tmp/try-kosli/code/web.src)"
```

You can verify the deployment with:

```shell {.command}
kosli deployment ls web-server
```

```plaintext {.light-console}
ID   ARTIFACT                                                                        ENVIRONMENT  REPORTED_AT
1    Name: web_2.bin                                                                 production   16 Aug 22 08:02 CEST
     Fingerprint: cbc92ce1291830382ec23b95efc213d6e1725b5157bcb2927d48296b61c86746               
```

Get detailed information about a deployment:

```shell {.command}
kosli deployment get web-server#1
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

### See also the other tutorials:
- [Following a git commit to runtime environments](../following_a_git_commit_to_runtime_environments/)
- [Tracing a production incident back to git commits](../tracing_a_production_incident_back_to_git_commits/)
