---
title: Following a git commit to service execution
bookCollapseSection: false
weight: 1
draft: true
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

# Following a git commit to service execution

## Overview

In this 5 minute tutorial you'll learn how Kosli tracks "life after git" and shows you events from:
* CI-pipelines (eg, building the docker image, running the unit tests, deploying, etc)
* runtime environments (eg, the blue-green rollover, instance scaling, etc)

You'll follow an actual git commit to an open-source project called **cyber-dojo**.
cyber-dojo's `runner` service performs most of its heavy lifting and
should run with three replicas. Due to an oversight (whilst switching from K8S to AWS)
it was running with just one replica. You will follow the commit that fixed this.

<!-- Do we want to explicitly mention seeing into the runtime environment did not require
     knowledge any secrets nor how to navigate cloud console
-->

## Getting ready

You need to:
* [Install the `kosli` CLI](../../installation).
* [Get your Kosli API token](../../installation#getting-your-kosli-api-token).
* [Set the KOSLI_API_TOKEN environment variable](../../installation#set-the-kosli_api_token-environment-variable).
* Set the KOSLI_OWNER environment variable to `cyber-dojo`.   
  The Kosli `cyber-dojo` organization is public so any authenticated user 
  can read its data.   
  ```shell {.command}
  export KOSLI_OWNER=cyber-dojo
  ```

## Pipeline events

### Listing pipelines

Find out which `cyber-dojo` repositories have a CI pipeline reporting to [Kosli](https://app.kosli.com):

```shell {.command}
kosli pipeline ls
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
shas                    UX for git+image shas               public
web                     UX for practicing TDD               public
```

{{< hint info >}}
## cyber-dojo overview
* [cyber-dojo](https://cyber-dojo.org) is a web platform where teams 
practice TDD without any installation.  
* These docker images run in two AWS environments named 
[aws-beta](https://app.kosli.com/cyber-dojo/environments/aws-beta)
and [aws-prod](https://app.kosli.com/cyber-dojo/environments/aws-prod).
* cyber-dojo has a microservice architecture with a dozen git repositories.
* Each git repository has its own Github Actions CI pipeline producing a docker image as listed above.
{{< /hint >}}


### Following the artifact

The runner service, only had one instance running instead of three.
The commit which fixed the problem was 
[16d9990](https://github.com/cyber-dojo/runner/commit/16d9990ad23a40eecaf087abac2a58a2d2a4b3f4)
in the `runner` repository. We can follow this commit using the `kosli` command:

```shell {.command}
kosli artifact get runner:16d9990
```
You will see:

```plaintext {.light-console}
Name:        cyberdojo/runner:16d9990
SHA256:      9af401c4350b21e3f1df17d6ad808da43d9646e75b6da902cc7c492bcfb9c625
Created on:  Mon, 22 Aug 2022 11:34:59 CEST • 15 days ago
Git commit:  16d9990ad23a40eecaf087abac2a58a2d2a4b3f4
Commit URL:  https://github.com/cyber-dojo/runner/commit/16d9990ad23a40eecaf087abac2a58a2d2a4b3f4
Build URL:   https://github.com/cyber-dojo/runner/actions/runs/2902808452
State:       COMPLIANT
History:
    Artifact created                                     Mon, 22 Aug 2022 11:34:59 CEST
    branch-coverage evidence received                    Mon, 22 Aug 2022 11:36:00 CEST
    Deployment #18 to aws-beta environment               Mon, 22 Aug 2022 11:37:15 CEST
    Deployment #19 to aws-prod environment               Mon, 22 Aug 2022 11:38:19 CEST
    Started running in aws-beta#83 environment           Mon, 22 Aug 2022 11:38:30 CEST
    Started running in aws-prod#63 environment           Mon, 22 Aug 2022 11:39:45 CEST
    Scaled down from 3 to 2 in aws-beta#117 environment  Wed, 24 Aug 2022 18:04:22 CEST
    No longer running in aws-beta#118 environment        Wed, 24 Aug 2022 18:05:22 CEST
    Scaled down from 3 to 1 in aws-prod#91 environment   Wed, 24 Aug 2022 18:10:14 CEST
    No longer running in aws-prod#93 environment         Wed, 24 Aug 2022 18:12:14 CEST
```

Let's look at this output in detail:

* **Name**: The name of the docker image is `cyberdojo/runner:16d9990`. Its image registry is defaulted to
`dockerhub`. Its :tag is the short-sha of the git commit.  
* **SHA256**: The `kosli` CLI knows how to 'fingerprint' any kind of artifact (docker images, zip files, etc) 
  to create a unique tamper-proof SHA. 
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
      * The artifact was **deployed** to [aws-beta](https://app.merkely.com/cyber-dojo/pipelines/runner/deployments/18) on 22nd  August 11:37:17 CEST, and to [aws-prod](https://app.merkely.com/cyber-dojo/pipelines/runner/deployments/19)
     a minute later.
   * **Runtime environment events**
      * The artifact was reported **running** in both environments.
      * The artifact's number of running instances **scaled down**.
      * The artifact was reported **exited**.
     
The information about this artifact is also available through the [web interface](https://app.merkely.com/cyber-dojo/pipelines/runner/artifacts/9af401c4350b21e3f1df17d6ad808da43d9646e75b6da902cc7c492bcfb9c625).

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
If a change is detected, a snapshot of the environment is saved.

{{< hint info >}}
Cyber-dojo runs the `kosli` CLI from inside its AWS runtime environments
using a [lambda function](https://github.com/cyber-dojo/merkely-environment-reporter/tree/main/deployment/terraform/lambda-reporter)
to report the running services to Kosli.
{{< /hint >}}


The **History** of the artifact tells us our artifact started running in snapshot #65 of `aws-prod`.

We query Kosli to see what was running in `aws-prod` snapshot #65:

```shell {.command}
kosli env get aws-prod#65
```

The output will be:

```plaintext {.light-console}
COMMIT   ARTIFACT                                                                    PIPELINE   RUNNING_SINCE  REPLICAS
16d9990  Name: 274425519734.dkr.ecr.eu-central-1.amazonaws.com/runner:16d9990        runner     11 days ago    3
         SHA256: 9af401c4350b21e3f1df17d6ad808da43d9646e75b6da902cc7c492bcfb9c625                              

7c45272  Name: 274425519734.dkr.ecr.eu-central-1.amazonaws.com/shas:7c45272          shas       11 days ago    1
         SHA256: 76c442c04283c4ca1af22d882750eb960cf53c0aa041bbdb2db9df2f2c1282be

...some output elided...

85d83c6  Name: 274425519734.dkr.ecr.eu-central-1.amazonaws.com/runner:85d83c6        runner     13 days ago    1
         SHA256: eeb0cfc9ee7f69fbd9531d5b8c1e8d22a8de119e2a422344a714a868e9a8bfec                              
 
1a2b170  Name: 274425519734.dkr.ecr.eu-central-1.amazonaws.com/differ:1a2b170        differ     13 days ago    1
         SHA256: d8440b94f7f9174c180324ceafd4148360d9d7c916be2b910f132c58b8a943ae
```

We see that in this snapshot, the `runner:16d9990` artifact is indeed running with 3 replicas.
We have proof the git commit has worked. 

{{< hint info >}}
## Blue-green deployment
There were *two* versions of `runner` at this point in time! 
The first, with three replicas (to fix the problem), but also a second (from commit `85d83c6`) with only one replica.

This is because we are in the middle of a **blue-green deployment**.
`runner:85d83c6` is about to be stopped, it will not be reported in
snapshot `aws-prod#66`.
{{< /hint >}}

## Diffing snapshots

Kosli's `env diff` command allows you to see differences between two versions of your
runtime environment.

Let's find out what's *different* between the `aws-prod#64` and `aws-prod#65` snapshots: 

```shell {.command}
kosli env diff aws-prod#64 aws-prod#65
```

The response will be:

```plaintext {.light-console}
only present in aws-prod#65

  Name: 274425519734.dkr.ecr.eu-central-1.amazonaws.com/runner:16d9990
  Sha256: 9af401c4350b21e3f1df17d6ad808da43d9646e75b6da902cc7c492bcfb9c625
  Pipeline: runner
  Commit: https://github.com/cyber-dojo/runner/commit/16d9990ad23a40eecaf087abac2a58a2d2a4b3f4
  Started: Mon, 22 Aug 2022 11:39:17 CEST • 15 days ago
```

<!-- Do we want the label for Commit: to be Commit URL: to match the
     label you see in a `kosli artifact get` command
-->

The ouput above shows that `runner:16d9990` started running in snapshot 65 of `aws-prod` environment.

We have seen how Kosli can follow a git commit on its way into production,
and provide information about the artifacts history, without any access to cyber-dojo's `aws-prod` environment.

<!-- Do we want to explicitly mention seeing into the runtime environment did not require
     knowledge any secrets nor how to navigate cloud console
-->

Next, we will find how to trace a production incident back to a git commit.

{{< button relref="/tracing_a_production_incident_back_to_git_commits" >}}Next >{{< /button >}}