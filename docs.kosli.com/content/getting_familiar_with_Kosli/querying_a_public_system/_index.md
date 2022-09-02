---
title: Querying a public system
bookCollapseSection: false
weight: 2
draft: true
---

In this tutorial you'll learn how Kosli tracks "life after git". 
You'll use the Kosli cli to "follow" a commit to a git repository that is part
of the [cyber-dojo](https://cyber-dojo.org) open source project.
You see the dynamic events related to:
* Its CI-pipeline (eg building docker image, running unit tests, deploying, etc)
* Its AWS runtime environments (eg blue-green rollover, instance scaling, etc)

# Cyber-dojo Introduction

cyber-dojo is an open source platform where teams can practice TDD (in
many languages) directly from a browser without any installation.

cyber-dojo has a standard microservice architecture with a dozen git repositories
(eg [web](https://github.com/cyber-dojo/web), [runner](https://github.com/cyber-dojo/runner)).
Each git repository has its own Github Actions CI pipeline producing a docker image.

These docker images run in two AWS environments whose Kosli names are 
[aws-beta](https://app.kosli.com/cyber-dojo/environments/aws-beta)
and [aws-prod](https://app.kosli.com/cyber-dojo/environments/aws-prod).


# Getting Ready

You run the Kosli CLI commands in this tutorial in a terminal, either in a docker container or
directly on your local machine.
You need to:
* [Install the Kosli cli](../installation)
* [Sign up to Kosli with Github](https://app.kosli.com) so you have a Kosli API token.
* [Get your Kosli API token](../installation#getting-your-kosli-api-token)
* Set the KOSLI_API_TOKEN environment variable. You need this to authenticate.
```shell {.command}
export KOSLI_API_TOKEN=<put your kosli API token here>
```
* Set the KOSLI_OWNER environment variable to `cyber-dojo`. cyber-dojo
is a public Kosli organization and is readable by any authenticated user. 
```shell {.command}
export KOSLI_OWNER=cyber-dojo
```

# CI Pipeline Events

The `kosli` cli automatically uses the `KOSLI_API_TOKEN` to authenticate,
and the `KOSLI_OWNER` environment variable to specify a Kosli organization. 
Let's start by confirming that all 12 `cyber-dojo` repositories have
a CI pipeline reporting to Kosli. 

```shell {.command}
kosli pipeline ls
```
```shell
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

The name of a Kosli pipeline does not have to match the name of a git
repository - but it helps if the relationship is clear.

We will follow the git commit [16d9990](https://github.com/cyber-dojo/runner/commit/16d9990ad23a40eecaf087abac2a58a2d2a4b3f4) 
to the `runner` repository.

Let's find out which artifact was built from this commit.

```shell {.command}
kosli artifact get runner:16d9990ad23a40eecaf087abac2a58a2d2a4b3f4
```

```shell
Name:        cyberdojo/runner:16d9990
SHA256:      9af401c4350b21e3f1df17d6ad808da43d9646e75b6da902cc7c492bcfb9c625
Created on:  Mon, 22 Aug 2022 11:35:00 CEST • 11 days ago
Git commit:  16d9990ad23a40eecaf087abac2a58a2d2a4b3f4
Commit URL:  https://github.com/cyber-dojo/runner/commit/16d9990ad23a40eecaf087abac2a58a2d2a4b3f4
Build URL:   https://github.com/cyber-dojo/runner/actions/runs/2902808452
State:       COMPLIANT
History:
    Artifact created                               Mon, 22 Aug 2022 11:35:00 CEST
    branch-coverage evidence received              Mon, 22 Aug 2022 11:36:02 CEST
    Deployment #18 to aws-beta environment         Mon, 22 Aug 2022 11:37:17 CEST
    Deployment #19 to aws-prod environment         Mon, 22 Aug 2022 11:38:21 CEST
    Reported running in aws-beta#84 environment    Mon, 22 Aug 2022 11:38:28 CEST
    Reported running in aws-prod#65 environment    Mon, 22 Aug 2022 11:39:22 CEST
    Reported running in aws-beta#117 environment   Wed, 24 Aug 2022 18:03:42 CEST
    No longer running in aws-beta#119 environment  Wed, 24 Aug 2022 18:05:42 CEST
    Reported running in aws-prod#94 environment    Wed, 24 Aug 2022 18:10:28 CEST
    No longer running in aws-prod#96 environment   Wed, 24 Aug 2022 18:12:28 CEST

            
```

<!-- Here we could mention the URL for seeing this in app.kosli.com 
     where `aws-prod#65` etc are clickable links (hopefully)!
-->

We can see:
* Name: The name of the docker image `cyberdojo/runner:16d9990`. Its image registry defaults to
`dockerhub`. Its :tag is the short-sha of the git commit. Kosli also supports pipelines building other 
kinds of artifacts (eg zip files).
* SHA256: Kosli knows how to 'fingerprint' any kind of artifact to create a unique tamper-proof digest.
  Note that the first seven characters of the 64 character digest are `9af401c`.
* Created on: The artifact was created on 22nd August 2022, at 11:35 CEST.
* Commit URL: cyber-dojo's git repositories are public; you can follow this link 
  to the actual commit on Github. 
* Build URL: Again, you can follow this link to the actual Github Action for this commit.
* State: COMPLIANT means that all the promised evidence for the artifact was provided before deployment.
* History:
   * Artifact was created on the 22nd August.
     This report came from a simple call to `kosli` in the CI pipeline yml, just after
     the docker image is built.
   * The artifact has attached evidence for `branch-coverage`. 
     This evidence is also reported with a call to `kosli` in exactly the same way, right after 
     the tests pass and the coverage stats are generated.
   * The artifact started deploying to `aws-beta` on 22nd August, and to `aws-prod` one minute later.
     Again, a call to `kosli` reported this just before the terraform deployments.  
     The `runner` service uses Continuous Deployment; if the tests pass the artifact 
     is deployed *directly* to both runtime environments without a manual approval step.
     Some cyber-dojo services (eg web) do have a manual approval step, and Kosli supports this.
   * The artifact was reported running in the `aws-beta` and `aws-prod` environments a minute later.
   * The artifact was reported exited both `aws-beta` and `aws-prod` 2 days later at the times given.
     These last two reports came from calls to `kosli` running inside the environments. 


# Runtime Environment Events

cyber-dojo runs `kosli` from inside the AWS runtime environments
using a lambda function. The lambda function periodically fingerprints 
all the running services and sends a "snapshot" of what is *actually*
running to Kosli. If the snapshot is different to the previous snapshot 
Kosli saves it.

The History tells us the docker image our commit produced was first seen running
in `aws-beta` in that environments `84`th snapshot, and 
in `aws-prod` in that environments `65`th snapshot.

Let's get the whole of `aws-prod`'s `65`th snapshot

```shell {.command}
kosli env get aws-prod#65
```

Some output ...elided for brevity:
```shell
COMMIT   ARTIFACT                                                                              PIPELINE                RUNNING_SINCE  REPLICAS
16d9990  Name: 274425519734.dkr.ecr.eu-central-1.amazonaws.com/runner:16d9990                  runner                  11 days ago    3
         SHA256: 9af401c4350b21e3f1df17d6ad808da43d9646e75b6da902cc7c492bcfb9c625                                                     
                                                                                                                                      
7c45272  Name: 274425519734.dkr.ecr.eu-central-1.amazonaws.com/shas:7c45272                    shas                    11 days ago    1
         SHA256: 76c442c04283c4ca1af22d882750eb960cf53c0aa041bbdb2db9df2f2c1282be
...
85d83c6  Name: 274425519734.dkr.ecr.eu-central-1.amazonaws.com/runner:85d83c6                  runner                  13 days ago    1
         SHA256: eeb0cfc9ee7f69fbd9531d5b8c1e8d22a8de119e2a422344a714a868e9a8bfec                                                     
                                                                                                                                      
1a2b170  Name: 274425519734.dkr.ecr.eu-central-1.amazonaws.com/differ:1a2b170                  differ                  13 days ago    1
         SHA256: d8440b94f7f9174c180324ceafd4148360d9d7c916be2b910f132c58b8a943ae                                                                                                                                                                                  
```

This output reveals two interesting things:

First, the name of the first artifact is `274425519734.dkr.ecr.eu-central-1.amazonaws.com/runner:16d9990`
and *not* `cyberdojo/runner:16d9990` as earlier reported! However, we can see the
commit is the same (`16d9990`) and, more importantly, the SHA256 (from `kosli`s fingerprinting)
is also the same (`9af401c...`).
Why the difference?
Answer: cyber-dojo is an open source project; its git repositories are
public, and the docker images it builds are saved to a public registry (`dockerhub`). 
However, the images running inside the `aws-beta` and `aws-prod` environments 
(which support the `https://cyber-dojo.org` web site) are *private* and are pulled from
a different *private* registry (`274425519734.dkr.ecr.eu-central-1.amazonaws.com`).

Second, there were *two* versions of `runner` at this point in time! 
The first (from commit `16d9990`) has three instances (replicas). 
This is as expected; the `runner` service bears the brunt of cyber-dojo's load.
The second (from commit `85d83c6`) has only one instance.
What is going on?
Let's look at `aws-prod`s *next* snapshot:

```shell {.command}
kosli env get aws-prod#66
```

```shell
COMMIT   ARTIFACT                                                                              PIPELINE                RUNNING_SINCE  REPLICAS
16d9990  Name: 274425519734.dkr.ecr.eu-central-1.amazonaws.com/runner:16d9990                  runner                  11 days ago    3
         SHA256: 9af401c4350b21e3f1df17d6ad808da43d9646e75b6da902cc7c492bcfb9c625                                                     
...
```

We still see the three instances of `runner` from commit `16d9990`.
But the one instance of `runner` from commit `85d83c6` is no longer listed.
It stopped running.
We are seeing a blue-green deployment, mid-flight, in the wild!

We can also find out what changed between snapshot `65` and `66` of `aws-prod`: 

```shell {.command}
kosli env diff aws-prod#65 aws-prod#66
```

```shell
- Name:   274425519734.dkr.ecr.eu-central-1.amazonaws.com/runner:85d83c6
  Sha256: eeb0cfc9ee7f69fbd9531d5b8c1e8d22a8de119e2a422344a714a868e9a8bfec
  Pipeline: runner
  Commit: https://github.com/cyber-dojo/runner/commit/85d83c6ab8e0ce800baeef3dfa4fa9f6eee338a4
  Started: Sat, 20 Aug 2022 22:32:43 CEST • 13 days ago
```
The minus sign in front of the name indicates `runner:85d83c6` has exited (as expected). 
A plus sign would indicate a newly started service.


<!--
Do we want to mention the whole env being compliant?
-->

<!-- 
TODO:
Do we want a command so we can get a list of snapshots that a given artifact was running in?
kosli env get aws-prod@9af401c4350b21e3f1df17d6ad808da43d9646e75b6da902cc7c492bcfb9c625 
-->

<!-- 
This we would like to show the users:
- Kosli gives developers without access to production environment information about what is running.
- Detect that a new "bit-coin miner" is running in your environment. Rogue artifact detection.
- Kosli can show that a deployment is reported, but artifact didn't start. Find this in artifact view.
- Kosli can show that an artifact started, but no deployment was reported for it.
- Detect an artifact that is missing evidence is running in an environment

- Commit makes the server stop working. Use kosli env diff to find out what artifact changed.
It would be good if we had two versions of env where there are several artifacts that change.
(with easter egg)

(- Find out when/where a given commit is running.)

- See what SW is/was running where which is useful in debugging.
  I detect from the web page that there is something wrong with 'saver'. I then want to know
  which version of 'saver' is running now. I want to know what git commit is running.
- List which version of 'saver' is running across all environments.

- We see that beta.cyberdojo.org is not working as expected, but prod is still OK. We do a kosli env diff and
  kosli env log to find out what services has changed.

- Change of K8S infrastructure broke both cyber dojo environments. The fix was to manually change 3 of the
  services on prod. Beta was not fixed and was down for a long period. We might not be able to detect this.

Problems:
- Not every commit generates an artifact. If you only build after 10 commits then 9 will not
be visible.

Things we can do later:
- Find which artifact this "unknown commit" is part of. So we need the git history.
- Kosli can show that an older deployment is running than that is declared. roll-back

 -->