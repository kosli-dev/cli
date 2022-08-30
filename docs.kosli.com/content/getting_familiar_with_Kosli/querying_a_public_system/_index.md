---
title: Querying a public system
bookCollapseSection: false
weight: 2
draft: false
---

# Cyber-dojo introduction

[Cyber-dojo](https://cyber-dojo.org) is an open source platform where teams can practice TDD in
many different languages without any installation.

Cyber-dojo has a standard micro service architecture with a dozen or so git repositories
(eg [web](https://github.com/cyber-dojo/web), [runner](https://github.com/cyber-dojo/runner)).
Each git repository has its own CI pipeline producing a docker image.

These docker images run in two AWS environments called [aws-beta](https://app.merkely.com/cyber-dojo/environments/aws-beta)
and [aws-prod](https://app.merkely.com/cyber-dojo/environments/aws-prod).

In this tutorial you will learn how Kosli tracks "life after git". In other words, all the dynamic events
after a Cyber-dojo git commit:
* In the CI-pipeline (eg building docker image, running unit tests, deploying, etc)
* In the AWS runtime environment (eg blue-green rollover, instance scaling, etc)


# Getting started

If you want to you can run the actual commands in this tutorial in a bash terminal.
You need to:
* [Install Kosli](../installation)
* [Sign up to Kosli with Github](https://app.kosli.com)
* [Get your Kosli API token](../installation#getting-your-kosli-api-token) and set the following environment variables:
```shell {.command}
export KOSLI_API_TOKEN=<put your kosli API token here>
export KOSLI_OWNER=cyber-dojo
```
<!-- 
You can verify your comand with:
```shell {.command}
kosli env ls 
```
```shell
NAME      TYPE  LAST REPORT                LAST MODIFIED
aws-beta  ECS   2022-08-30T13:18:42+02:00  2022-08-30T13:18:42+02:00
aws-prod  ECS   2022-08-30T13:18:28+02:00  2022-08-30T13:18:28+02:00
beta      K8S   2022-06-15T11:39:59+02:00  2022-06-15T11:39:59+02:00
prod      K8S   2022-06-15T11:40:01+02:00  2022-06-15T11:40:01+02:00
``` -->


# Pipeline events

We will follow a git commit [16d9990](https://github.com/cyber-dojo/runner/commit/16d9990ad23a40eecaf087abac2a58a2d2a4b3f4) 
to the `runner` repository through its CI-pipeline.

The name of a git repository, CI-pipeline and Kosli pipeline does not need to match. But it makes
life a lot easier if the relationship is clear.

We can list the Kosli pipelines so we can match up the name with the git repository name.

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

In this case we have a Kosli pipeline named `runner`.

Lets find out which artifact was built from this commit.
<!-- kosli artifact get runner@9af401c4350b21e3f1df17d6ad808da43d9646e75b6da902cc7c492bcfb9c625 -->
```shell {.command}
kosli artifact get runner{commit=16d9990}
```

```shell
Name:        cyberdojo/runner:16d9990
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

We can see the Name of the artifact, including version number. Cyber-dojo uses the short commit
sha as version number, but semantic or no versioning can also be used.
TODO: State COMPLIANT??
The Build and Commit URL points back to the source code and the build system. We can see
when the artifact was built and that there are no Approvals for this artifact.
This artifact was deployed to both aws-beta and aws-prod and exited 2 days later.

TODO:
We need a `kosli artifact get runner{commit=16d9990}` command
The output of the command that is listed here is missing the artifact sha. Should it also
contain a URL to the artifact?
https://app.merkely.com/cyber-dojo/pipelines/runner/artifacts/9af401c4350b21e3f1df17d6ad808da43d9646e75b6da902cc7c492bcfb9c625


