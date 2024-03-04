---
title: Following a git commit to runtime environments
bookCollapseSection: false
weight: 510
draft: false
---

<!-- The book "Developer Marketing Does Not Exist" by Adam DuVander suggests 
     this tutorial content structure (p49)
     1. Explain the context
     2. Show the end result
     3. Walk through the steps
     4. Help them take the next step

     Quoting from the book...
       "If you find the first three words of your tutorial are
        In this tutorial" then you might have skipped ahead."
     These were our exact first three words!
     I've tried to add initial context.
     I think we are still missing step 2 (see below)
     I think we are still missing step 4, which should probably
       be a simple link to the next tutorial.
-->

# Following a git commit to runtime environments

## Overview

In this 5 minute tutorial you'll learn how Kosli tracks "life after git" and shows you events from:
* CI pipelines (eg, building the docker image, running the unit tests, deploying, etc)
* runtime environments (eg, the blue-green rollover, instance scaling, etc)

You'll follow an actual git commit to an open-source project called **cyber-dojo**. 
In our example cyber-dojo’s `runner` service should run with three replicas. However, due to an oversight while switching
from Google Kubernetes Engine (GKE) to AWS Elastic Container Service (ECS), it was running with just one replica. 
You will follow the commit that fixed this. 

## Getting ready

You need to:
* [Install Kosli CLI](/getting_started/install/).
* [Get a Kosli API token](/getting_started/service-accounts/).
* Set the `KOSLI_ORG` environment variable to `cyber-dojo` (the Kosli `cyber-dojo` organization is public so any authenticated user can read its data) and `KOSLI_API_TOKEN` to your token:
  ```shell {.command}
  export KOSLI_ORG=cyber-dojo
  export KOSLI_API_TOKEN=<your-api-token>
  ```

## CI Pipeline events

### Listing flows

Find out which `cyber-dojo` repositories have a CI pipeline reporting to [Kosli](https://app.kosli.com):

```shell {.command}
kosli ls flows
```

You will see:

```plaintext {.light-console}
NAME                    DESCRIPTION                         VISIBILITY
creator                 UX for Group/Kata creation          public
custom-start-points     Custom exercises choices            public
dashboard               UX for a group practice dashboard   public
differ                  Diff files from two traffic-lights  public
exercises-start-points  Exercises choices                   public
languages-start-points  Language+TestFramework choices      public
nginx                   Reverse proxy                       public
repler                  REPL for Python images              public
runner                  Test runner                         public
saver                   Group/Kata model+persistence        public
version-reporter        UX for git+image version-reporter   public
web                     UX for practicing TDD               public
```

{{< hint info >}}
## cyber-dojo overview
* [cyber-dojo](https://cyber-dojo.org) is a web platform where teams 
practice TDD without any installation.  
* cyber-dojo has a microservice architecture with a dozen git repositories.
* Each git repository has its own Github Actions CI pipeline producing a docker image.
* These docker images run in two AWS environments named 
[aws-beta](https://app.kosli.com/cyber-dojo/environments/aws-beta)
and [aws-prod](https://app.kosli.com/cyber-dojo/environments/aws-prod).
{{< /hint >}}


### Following the artifact

The runner service had one instance running instead of three.
The commit which fixed the problem was 
[16d9990](https://github.com/cyber-dojo/runner/commit/16d9990ad23a40eecaf087abac2a58a2d2a4b3f4)
in the `runner` repository. Follow this commit using the `kosli` command:

```shell {.command}
kosli get artifact runner:16d9990
```
You will see:

```plaintext {.light-console}
Name:         cyberdojo/runner:16d9990
Flow:         runner
Fingerprint:  9af401c4350b21e3f1df17d6ad808da43d9646e75b6da902cc7c492bcfb9c625
Created on:   Mon, 22 Aug 2022 11:35:00 CEST • 15 days ago
Git commit:   16d9990ad23a40eecaf087abac2a58a2d2a4b3f4
Commit URL:   https://github.com/cyber-dojo/runner/commit/16d9990ad23a40eecaf087abac2a58a2d2a4b3f4
Build URL:    https://github.com/cyber-dojo/runner/actions/runs/2902808452
State:        COMPLIANT
History:
    Artifact created                                     Mon, 22 Aug 2022 11:35:00 CEST
    branch-coverage evidence received                    Mon, 22 Aug 2022 11:36:02 CEST
    Deployment #18 to aws-beta environment               Mon, 22 Aug 2022 11:37:17 CEST
    Deployment #19 to aws-prod environment               Mon, 22 Aug 2022 11:38:21 CEST
    Started running in aws-beta#84 environment           Mon, 22 Aug 2022 11:38:28 CEST
    Started running in aws-prod#65 environment           Mon, 22 Aug 2022 11:39:22 CEST
    Scaled down from 3 to 2 in aws-beta#117 environment  Wed, 24 Aug 2022 18:03:42 CEST
    No longer running in aws-beta#119 environment        Wed, 24 Aug 2022 18:05:42 CEST
    Scaled down from 3 to 1 in aws-prod#94 environment   Wed, 24 Aug 2022 18:10:28 CEST
    No longer running in aws-prod#96 environment         Wed, 24 Aug 2022 18:12:28 CEST
```

Let's look at this output in detail:

* **Name**: The name of the docker image is `cyberdojo/runner:16d9990`. Its image registry is defaulted to
`dockerhub`. Its :tag is the short-sha of the git commit.
* **Flow**: The name of the Kosli flow.
* **Fingerprint**: The unique immutable SHA256 fingerprint of the artifact.
* **Created on**: The artifact was created on 22nd August 2022, at 11:35 CEST.
* **Commit URL**: You can follow [the commit URL](https://github.com/cyber-dojo/runner/commit/16d9990ad23a40eecaf087abac2a58a2d2a4b3f4) 
  to the actual commit on Github since cyber-dojo's git repositories are public.
* **Build URL**: You can follow [the build URL](https://github.com/cyber-dojo/runner/actions/runs/2902808452)
  to the actual Github Action for this commit.
* **State**: COMPLIANT means that all the promised evidence for the artifact (in this case `branch-coverage`)
  was provided before deployment.
* **History**:
   * **CI pipeline events**
      * The artifact was **created** on the 22nd August at 11:35:00 CEST.
      * The artifact has `branch-coverage` **evidence**. 
      * The artifact was **deployed** to [aws-beta](https://app.kosli.com/cyber-dojo/flows/runner/deployments/18) on 22nd  August 11:37:17 CEST, and to [aws-prod](https://app.kosli.com/cyber-dojo/flows/runner/deployments/19)
     a minute later.
   * **Runtime environment events**
      * The artifact was reported **running** in both environments.
      * The artifact's number of running instances **scaled down**.
      * The artifact was reported **exited**.
     
The information about this artifact is also available through the [web interface](https://app.kosli.com/cyber-dojo/flows/runner/artifacts/9af401c4350b21e3f1df17d6ad808da43d9646e75b6da902cc7c492bcfb9c625).

{{< hint info >}}
The `runner` service uses [Continuous Deployment](https://en.wikipedia.org/wiki/Continuous_deployment); 
if the tests pass the artifact is [blue-green deployed](https://en.wikipedia.org/wiki/Blue-green_deployment) 
to both its runtime environments *without* any manual approval steps.
Some cyber-dojo services (eg web) have a manual approval step, and Kosli supports this.
{{< /hint >}}

## Environment Snapshots

Kosli environments store information about what is running in your actual runtime environments (eg server, Kubernetes cluster, AWS, ...).
We use one Kosli environment per runtime environment.

The Kosli CLI periodically fingerprints all the running artifacts in a runtime environment and reports them to Kosli.
Whenever a change is detected, a snapshot of the environment is saved.

{{< hint info >}}
Cyber-dojo runs the `kosli` CLI from inside its AWS runtime environments
using a [lambda function](https://github.com/cyber-dojo/kosli-environment-reporter/blob/main/deployment/terraform/deployment.tf)
to report the running services to Kosli.
{{< /hint >}}


The **History** of the artifact tells you your artifact started running in snapshot #65 of `aws-prod`.

You query Kosli to see what was running in `aws-prod` snapshot #65:

```shell {.command}
kosli get snapshot aws-prod#65
```

The output will be:

```plaintext {.light-console}
COMMIT   ARTIFACT                                                                         FLOW       RUNNING_SINCE  REPLICAS
16d9990  Name: 274425519734.dkr.ecr.eu-central-1.amazonaws.com/runner:16d9990             runner     11 days ago    3
         Fingerprint: 9af401c4350b21e3f1df17d6ad808da43d9646e75b6da902cc7c492bcfb9c625                              
                                                                                                    
7c45272  Name: 274425519734.dkr.ecr.eu-central-1.amazonaws.com/shas:7c45272               shas       11 days ago    1
         Fingerprint: 76c442c04283c4ca1af22d882750eb960cf53c0aa041bbdb2db9df2f2c1282be                              

...some output elided...

85d83c6  Name: 274425519734.dkr.ecr.eu-central-1.amazonaws.com/runner:85d83c6             runner     13 days ago    1
         Fingerprint: eeb0cfc9ee7f69fbd9531d5b8c1e8d22a8de119e2a422344a714a868e9a8bfec                              
                                                                                                  
1a2b170  Name: 274425519734.dkr.ecr.eu-central-1.amazonaws.com/differ:1a2b170             differ     13 days ago    1
         Fingerprint: d8440b94f7f9174c180324ceafd4148360d9d7c916be2b910f132c58b8a943ae                              
```

You see in this snapshot, the `runner:16d9990` artifact is indeed running with 3 replicas.
You have proof the git commit has worked. 

{{< hint info >}}
## Blue-green deployment
There were *two* versions of `runner` at this point in time! 
The first had three replicas (to fix the problem), but there was also a second (from commit `85d83c6`) with only one replica.

You are seeing a **blue-green deployment** happening;
`runner:85d83c6` is about to be stopped and will not be reported in
snapshot `aws-prod#66`.
{{< /hint >}}

## Diffing snapshots

Kosli's `env diff` command allows you to see differences between two versions of your
runtime environment.

Let's find out what's *different* between the `aws-prod#64` and `aws-prod#65` snapshots: 

```shell {.command}
kosli diff snapshots aws-prod#64 aws-prod#65
```

The response will be:

```plaintext {.light-console}
Only present in aws-prod#65
                   
     Name:         274425519734.dkr.ecr.eu-central-1.amazonaws.com/runner:16d9990
     Fingerprint:  9af401c4350b21e3f1df17d6ad808da43d9646e75b6da902cc7c492bcfb9c625
     Flow:         runner
     Commit URL:   https://github.com/cyber-dojo/runner/commit/16d9990ad23a40eecaf087abac2a58a2d2a4b3f4
     Started:      Mon, 22 Aug 2022 11:39:17 CEST • 15 days ago
```

The output above shows that `runner:16d9990` started running in snapshot 65 of `aws-prod` environment.

You have seen how Kosli can follow a git commit on its way into production,
and provide information about the artifacts history, without any access to cyber-dojo's `aws-prod` environment.
