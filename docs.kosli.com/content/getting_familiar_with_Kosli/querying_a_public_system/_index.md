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
State:       COMPLIANT
Git commit:  16d9990ad23a40eecaf087abac2a58a2d2a4b3f4
Build URL:   https://github.com/cyber-dojo/runner/actions/runs/2902808452
Commit URL:  https://github.com/cyber-dojo/runner/commit/16d9990ad23a40eecaf087abac2a58a2d2a4b3f4
Created at:  22 Aug 22 11:35 CEST • 8 days ago
Approvals:   None
Deployments:
     #18 Reported deployment to aws-beta at 22 Aug 22 11:37 CEST (Exited on 24 Aug 22 18:05 CEST)
     #19 Reported deployment to aws-prod at 22 Aug 22 11:38 CEST (Exited on 24 Aug 22 18:12 CEST)
Evidence:
     branch-coverage:  COMPLIANT
```
<!-- I think it makes sense for these to be printed with the two URLs in swapped order -->
<!-- and for the Evidence to come before Approvals -->

We can see:
* Name: The name of the docker image. Its :tag is the short-sha of 
the git commit. Kosli also supports pipelines building other kinds of artifacts, such 
as zip files.
* SHA256: Kosli knows how to 'fingerprint' any kind of artifact to create a unique tamper-proof digest.  
* State.  
* Commit URL: You can follow this link to the actual commit on Github. 
* Build URL: You can follow this link to the actual Github Action for this commit.
* Created at: The artifact was created on 22nd August 2022, at 11:35 CEST.
<!-- It is unfortunate that the day is the same as the year (22). Do we want to print 2022? -->
<!-- There are no Approvals for this artifact. Should we simply not show this? -->
* Deployments. The artifact was deployed to `aws-beta` on 22nd August, and to `aws-prod` one minute later.
It exited both `aws-beta` and `aws-prod` 2 days later at the times given.
* Evidence. The artifact has attached evidence for branch-coverage. This evidence was reported from the CI-pipeline.

<!-- 
TODO:
Do we want a command so we can get a list of snapshots that a given artifact was running in?
kosli env get aws-prod@9af401c4350b21e3f1df17d6ad808da43d9646e75b6da902cc7c492bcfb9c625 
-->

<!-- 
This we would like to show the users:
- Kosli gives developers without access to production environment information about what is running.
- Detect that a new "bit-coin miner" is running in your environment.
- Detect that an unexpected version of an artifact is running.
- Commit makes the server stop working. Use kosli env diff to find out what artifact changed.
It would be good if we had two versions of prod where there are several artifacts that change.
- Kosli could detect that a deployment did not start to run.
- Find out when/where a given commit is running.
- See what SW is/was running where which is useful in debugging.
  I detect from the web page that there is something wrong with 'saver'. I then want to know
  which version of 'saver' is running now.
- We see that staging has stopped working, but prod is still OK. We do a kosli env diff and
  kosli env log to find out what services has changed.
- List which version of 'saver' is running across all environments.

Problems:
- Not every commit generates an artifact. If you only build after 10 commits then 9 will not
be visible.

 -->