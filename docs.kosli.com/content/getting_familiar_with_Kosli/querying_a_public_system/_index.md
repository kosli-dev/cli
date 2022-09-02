---
title: Querying a public system
bookCollapseSection: false
weight: 2
draft: true
---

# Cyber-dojo introduction

[Cyber-dojo](https://cyber-dojo.org) is an open source platform where teams can practice TDD in
many different languages directly from the browser without any installation.

Cyber-dojo has a standard microservice architecture with a dozen or so git repositories
(eg [web](https://github.com/cyber-dojo/web), [runner](https://github.com/cyber-dojo/runner)).
Each git repository has its own CI pipeline producing a public docker image.

These docker images run in two AWS environments called [aws-beta](https://app.merkely.com/cyber-dojo/environments/aws-beta)
and [aws-prod](https://app.merkely.com/cyber-dojo/environments/aws-prod).

In this tutorial you will learn how Kosli tracks "life after git". In other words, all the dynamic events
after a Cyber-dojo git commit:
* In the CI-pipeline (eg building docker image, running unit tests, deploying, etc)
* In the AWS runtime environment (eg blue-green rollover, instance scaling, etc)


# Getting started

<!-- When I try to run `docker pull ghcr.io/kosli-dev/cli:v0.1.10` I get
     Error response from daemon: Head "https://ghcr.io/v2/kosli-dev/cli/manifests/v0.1.10": denied: denied
-->

You can run the actual commands in this tutorial in a terminal. Either in a docker container or
on your local machine.
You need to:
* [Install Kosli](../installation)
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

# Pipeline events

The `kosli` cli automatically uses the `KOSLI_API_TOKEN` to authenticate,
and the `KOSLI_OWNER` environment variable to specify a Kosli organization. 
So with these two environment variables set, you can list the 
`cyber-dojo` CI pipelines reporting to Kosli. 

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

The name of a Kosli pipeline does not need to match the name of a git
repository - but it helps if the relationship is clear.

We will follow the git commit [16d9990](https://github.com/cyber-dojo/runner/commit/16d9990ad23a40eecaf087abac2a58a2d2a4b3f4) 
to the `runner` repository through its CI-pipeline.

Lets find out which artifact was built from this commit.
<!-- kosli artifact get runner@9af401c4350b21e3f1df17d6ad808da43d9646e75b6da902cc7c492bcfb9c625 -->
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

We can see:
* Name: The name of the docker image. Its :tag is the short-sha of 
the git commit. Kosli also supports pipelines building other kinds of artifacts, such 
as zip files.
* SHA256: Kosli knows how to 'fingerprint' any kind of artifact to create a unique tamper-proof digest.  
* Created on: The artifact was created on 22nd August 2022, at 11:35 CEST.
* Commit URL: You can follow this link to the actual commit on Github. 
* Build URL: You can follow this link to the actual Github Action for this commit.
* State: COMPLIANT means that the promised evidence for this artifact has been provided.
* History: This shows the history of the artifact:
   * Artifact was created on on 22nd August.
   * The artifact has attached evidence for `branch-coverage`. This evidence was reported from the CI-pipeline.
   * The CI-pipeline reported that the artifact would be deployed to `aws-beta` on 22nd August, and to `aws-prod` one minute later.
   * The artifact was reported to run in `aws-beta` and `aws-prod` a minute later.
   * The artifact exited both `aws-beta` and `aws-prod` 2 days later at the times given.


# Environment events

Cyber-dojo have set up AWS to run a lambda function (same as cron job) periodically. The lambda function
collects the version and finger print for services running and report them to Kosli. If the report of what
is currently running is changing from last report Kosli generates a new snapshot. So the list
of snapshots for an environment describes all the changes to a runtime environment.

We have seen that our commit was deployed to `aws-prod` on 22nd of August and that it was reported
running in the `aws-prod` environment in snapshot number 65. We can show all the services that was
running in that snapshot

```shell {.command}
kosli env get aws-prod#65
```

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

We see that we are actually running two versions of `runner` at this point in time. This is due to blue-green
deployment. In the next snapshot the runner:85d83c6 has stopped.

```shell {.command}
kosli env get aws-prod#66
```
```shell
COMMIT   ARTIFACT                                                                              PIPELINE                RUNNING_SINCE  REPLICAS
16d9990  Name: 274425519734.dkr.ecr.eu-central-1.amazonaws.com/runner:16d9990                  runner                  11 days ago    3
         SHA256: 9af401c4350b21e3f1df17d6ad808da43d9646e75b6da902cc7c492bcfb9c625                                                     
...
```

A more convenient way to find out what changed between snapshot 65 and 66 is to use the Kosli diff tool
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
The minus sign in front of the name indicates a process that has stopped. A plus sign indicates a service that
has started.


<!--
Here we do
kosli env get aws-prod#104
to see what other services it was running in
and whether the whole env was compliant
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