---
title: Querying a public system
bookCollapseSection: false
weight: 2
draft: true
---

<!-- ?Make this a tutorial just for:
     1. Title?=Following a git commit to service execution

     ?Create a second tutorial for: 
     2. Title?=Tracing a production incident back to git commits
        The stories here would be simulated incidents with Easter-eggs comments.

     Ultimately it would be nice to have a third tutorial which
     traced an incident caused by eg, a change to the network configuration, 
-->

# Overview

In this tutorial you'll learn how Kosli tracks "life after git"!
You'll run `kosli` CLI commands to "follow" an actual git commit and
see dynamic events from:
* CI-pipelines (eg building docker image, running unit tests, deploying, etc)
* AWS runtime environments (eg blue-green rollover, instance scaling, etc)

<!-- Some of the URLs would be better if they opened in their own tab.
     We've looked into this and it does not seem to be supported in MarkDown
     https://stackoverflow.com/questions/4425198/can-i-create-links-with-target-blank-in-markdown
-->

This tutorial is based around the **cyber-dojo** project.

* [https://cyber-dojo.org](https://cyber-dojo.org) is an open source platform where teams 
practice TDD (in many languages) without any installation.  
* cyber-dojo has a microservice architecture with a dozen git repositories
(eg [web](https://github.com/cyber-dojo/web), [runner](https://github.com/cyber-dojo/runner)).  
* Each git repository has its own Github Actions CI pipeline producing a docker image.
* These docker images run in two AWS environments named 
[aws-beta](https://app.kosli.com/cyber-dojo/environments/aws-beta)
and [aws-prod](https://app.kosli.com/cyber-dojo/environments/aws-prod).


# Getting ready

<!-- the copy/paste text is always a single command in this tutorial.
     Can we use CSS to add a leading $ prompt that is not copied?
-->

You need to:
* [Install the `kosli` CLI](../installation).
* [Verify the installation worked](../installation#verifying-the-installation-worked).
* [Sign up to Kosli at https://app.kosli.com with Github](https://app.kosli.com).
* [Get your Kosli API token](../installation#getting-your-kosli-api-token).
* Set the KOSLI_API_TOKEN environment variable.  
  The `kosli` CLI uses this to authenticate you.
  ```shell {.command}
  export KOSLI_API_TOKEN=<paste-your-kosli-API-token-here>
  ```
* Set the KOSLI_OWNER environment variable to `cyber-dojo`.   
  The Kosli `cyber-dojo` organization is public so its readable by any authenticated user.   
  ```shell {.command}
  export KOSLI_OWNER=cyber-dojo
  ```

# Pipeline events

<!-- Do we want this `kosli pipeline ls` ? Does it add much value? 
-->

Find out which `cyber-dojo` repositories have a CI pipeline reporting to https://app.kosli.com:

```shell {.command}
kosli pipeline ls
```
<!-- We want the terminal-output to be visually
distinct to the terminal-input you copy/paste from.
Eg different colour background, no syntax highlighting
-->

You will see:

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

<!-- The name of a Kosli pipeline does not have to match the name of a git
repository - but it helps if the relationship is clear. -->

Find the artifact built from commit
[16d9990](https://github.com/cyber-dojo/runner/commit/16d9990ad23a40eecaf087abac2a58a2d2a4b3f4)
to cyber-dojo's `runner` repository:

<!-- Would be really nice if we had commit completion here so we could use 
     kosli artifact get runner:16d9990
-->

```shell {.command}
kosli artifact get runner:16d9990ad23a40eecaf087abac2a58a2d2a4b3f4
```
You will see:

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

<!-- Should we re-order the lines of this output a bit; based
     on the developer centric focus - starting with the commit?

Git commit:  16d9990ad23a40eecaf087abac2a58a2d2a4b3f4
Commit URL:  https://github.com/cyber-dojo/runner/commit/16d9990ad23a40eecaf087abac2a58a2d2a4b3f4
Build URL:   https://github.com/cyber-dojo/runner/actions/runs/2902808452
Artifact Name:  cyberdojo/runner:16d9990
Artifact SHA256: 9af401c4350b21e3f1df17d6ad808da43d9646e75b6da902cc7c492bcfb9c625
Created on:  Mon, 22 Aug 2022 11:35:00 CEST • 11 days ago
-->

<!-- There is plenty of scope for making various words in the text below into URLs
-->

<!-- We do not comment on the output showing the artifact running TWICE 
      - in aws-beta (84/117)
      - in aws-prod (65/94)
     Do we want to mention that as a third interesting thing? 
     Are missing some "No longer running" reports here...?
-->

<!-- We could mention and create clickable app.kosli.com URLs in the text, eg
     `aws-prod#65` takes you to that snapshot
     `#14` takes you to that deployment event
-->

Look at this output in detail:

* **Name**: The name of the docker image is `cyberdojo/runner:16d9990`. Its image registry is defaulted to
`dockerhub`. Its :tag is the short-sha of the git commit.  
* **SHA256**: The `kosli` CLI knows how to 'fingerprint' any kind of artifact (docker images, zip files, etc) 
  to create a unique tamper-proof SHA. 
  Later on you'll be referring back to this SHA - so note that the first seven characters of 
  this 64 character SHA are `9af401c`.
* **Created on**: The artifact was created on 22nd August 2022, at 11:35 CEST.
* **Commit URL**: You can follow [https://github.com/cyber-dojo/runner/commit/16d9990ad23a40eecaf087abac2a58a2d2a4b3f4](https://github.com/cyber-dojo/runner/commit/16d9990ad23a40eecaf087abac2a58a2d2a4b3f4) 
  to the actual commit on Github since cyber-dojo's git repositories are public.
* **Build URL**: Again, you can follow [https://github.com/cyber-dojo/runner/actions/runs/2902808452](https://github.com/cyber-dojo/runner/actions/runs/2902808452) 
  to the actual Github Action for this commit.
* **State**: COMPLIANT means that all the promised evidence for the artifact (see `branch-coverage` next) 
  was provided before deployment.
* **History**:
   * The artifact was created on the 22nd August at 11:35:00 CEST.
     This report came from a simple call to the `kosli` CLI in the CI pipeline yml, just after
     the docker image was built.
   * The artifact has `branch-coverage` evidence. 
     This evidence was also reported with a call to the `kosli` CLI in exactly the same way, right after 
     the tests passed and the coverage stats were generated.
   * The artifact started deploying to `aws-beta` on 22nd August 11:37:17 CEST, and to `aws-prod` 
     just over one minute later.
     Again, a call to the `kosli` CLI reported this just before the actual terraform deployments.  
     The `runner` service uses [Continuous Deployment](https://en.wikipedia.org/wiki/Continuous_deployment); 
     if the tests pass the artifact is [blue-green deployed](https://en.wikipedia.org/wiki/Blue-green_deployment) 
     to both its runtime environments *without* any manual approval steps.
     Some cyber-dojo services (eg web) have a manual approval step, and Kosli supports this.
   * The artifact was reported running in the `aws-beta` and `aws-prod` environments shortly after.
   * The artifact was reported exited both `aws-beta` and `aws-prod` at the times given.
     
These last two events were reported by the `kosli` CLI running *inside* 
cyber-dojo's AWS runtime environments. 

# Environment Snapshots

<!-- make [lambda function] text a link to the yml that runs the lambda.
     I think this is
     https://github.com/cyber-dojo/merkely-environment-reporter/tree/main/deployment/terraform/lambda-reporter
     Check with Artem
     If it is maybe do this after the repo has been renamed to
     kosli-environment-reporter
-->

<!-- At some point mention that you are getting all this information 
     without having to know anything about AWS, nor how to
     get the secrets needed. 
-->

cyber-dojo runs the `kosli` CLI from inside its AWS runtime environments
using a lambda function. The lambda function periodically fingerprints 
all the running services and sends a "snapshot" of what is *actually*
running to [https://app.kosli.com](https://app.kosli.com). 
If the snapshot is different to the previous snapshot it is saved.

The previous **History** tells us the docker image our commit produced was first seen running
in `aws-beta` in that environment's `84`'th snapshot, and 
in `aws-prod` in that environment's `65`'th snapshot.

<!-- We can add
If your replica-count fix has worked then the runner service will show three replicas
in snapshot `aws-prod#65`.
-->

Get the whole of `aws-prod`'s `65`'th snapshot:

```shell {.command}
kosli env get aws-prod#65
```

You will see:

```shell
COMMIT   ARTIFACT                                                                              PIPELINE                RUNNING_SINCE  REPLICAS
16d9990  Name: 274425519734.dkr.ecr.eu-central-1.amazonaws.com/runner:16d9990                  runner                  11 days ago    3
         SHA256: 9af401c4350b21e3f1df17d6ad808da43d9646e75b6da902cc7c492bcfb9c625                                                     
                                                                                                                                      
7c45272  Name: 274425519734.dkr.ecr.eu-central-1.amazonaws.com/shas:7c45272                    shas                    11 days ago    1
         SHA256: 76c442c04283c4ca1af22d882750eb960cf53c0aa041bbdb2db9df2f2c1282be

...some output elided...

85d83c6  Name: 274425519734.dkr.ecr.eu-central-1.amazonaws.com/runner:85d83c6                  runner                  13 days ago    1
         SHA256: eeb0cfc9ee7f69fbd9531d5b8c1e8d22a8de119e2a422344a714a868e9a8bfec                                                     
                                                                                                                                      
1a2b170  Name: 274425519734.dkr.ecr.eu-central-1.amazonaws.com/differ:1a2b170                  differ                  13 days ago    1
         SHA256: d8440b94f7f9174c180324ceafd4148360d9d7c916be2b910f132c58b8a943ae                                                                                                                                                                                  
```

This output reveals some interesting things:

The name of the first artifact is `274425519734.dkr.ecr.eu-central-1.amazonaws.com/runner:16d9990`
and *not* `cyberdojo/runner:16d9990` as seen earlier! However, we can see the
commit is the same (`16d9990`) and, more importantly, the SHA256 is also the same (`9af401c...`).
Why the difference?
Answer: cyber-dojo's docker images are first saved to a public registry (`dockerhub`)
so anyone can run their own cyber-dojo web site.
However, the images running inside its `aws-beta` and `aws-prod` 
environments (which support the `https://cyber-dojo.org` web site) are are pulled from
a *private* registry (`274425519734.dkr.ecr.eu-central-1.amazonaws.com`).
But the identical SHA256 proves it is the same image with two different names.

Also, there were *two* versions of `runner` at this point in time! 
The first (from commit `16d9990`) has three replicas (you may need
to scroll to the right to see the replica information). 
This is as expected; the `runner` service bears the brunt of cyber-dojo's load.
The second (from commit `85d83c6`) has only one replica.
What is going on?

Look at the snapshot *after* `aws-prod#65`:

```shell {.command}
kosli env get aws-prod#66
```

You will see:

```shell
COMMIT   ARTIFACT                                                                              PIPELINE                RUNNING_SINCE  REPLICAS
16d9990  Name: 274425519734.dkr.ecr.eu-central-1.amazonaws.com/runner:16d9990                  runner                  11 days ago    3
         SHA256: 9af401c4350b21e3f1df17d6ad808da43d9646e75b6da902cc7c492bcfb9c625                                                     
...
```

We still see the three instances of `runner:16d9990`.
But the one instance of `runner:85d83c6` is no longer listed.
Between `aws-prod#65` and `aws-prod#66` it stopped running.
You were seeing the blue-green deployment, mid-flight!

# Diffing snapshots

Find out what's *different* between the `aws-prod#65` and `aws-prod#66` snapshots: 

```shell {.command}
kosli env diff aws-prod#65 aws-prod#66
```

You will see:

<!-- Can we colour this red as it actually appears?
     Use a screenshot?
-->

```shell
- Name:   274425519734.dkr.ecr.eu-central-1.amazonaws.com/runner:85d83c6
  Sha256: eeb0cfc9ee7f69fbd9531d5b8c1e8d22a8de119e2a422344a714a868e9a8bfec
  Pipeline: runner
  Commit: https://github.com/cyber-dojo/runner/commit/85d83c6ab8e0ce800baeef3dfa4fa9f6eee338a4
  Started: Sat, 20 Aug 2022 22:32:43 CEST • 13 days ago
```
The minus sign in front of **Name:** indicates `runner:85d83c6` stopped.
This was the *end* of the blue-green deployment.

Go backwards in time a little and look at the *previous* diff:

```shell {.command}
kosli env diff aws-prod#64 aws-prod#65
```

You will see:

<!-- Can we colour this green as it actually appears? 
     Use a screenshot?
-->

```shell
+ Name:   274425519734.dkr.ecr.eu-central-1.amazonaws.com/runner:16d9990
  Sha256: 9af401c4350b21e3f1df17d6ad808da43d9646e75b6da902cc7c492bcfb9c625
  Pipeline: runner
  Commit: https://github.com/cyber-dojo/runner/commit/16d9990ad23a40eecaf087abac2a58a2d2a4b3f4
  Started: 22 Aug 22 10:39 BST • 11 days ago
```

The plus sign in front of **Name:** indicates `runner:16d9990` started.
This was the *beginning* of the blue-green deployment.

<!-- Note that this diff output does NOT tell us how many instances of
     runner:16d9990 were running. Should it?
     Suppose this was a 2->3 scaling event?
-->

<-- Could we use the kosli CLI now to show the snapshot where
    runner:85d83c6 was initially running with only one instance?
    Again aws-prod#65 isn't really good enough since that could be
    mid-scaling event.
-->

# Diffing snapshots across environments!

The name of an environment without a snapshot number (or the `#` character)
specifies that environment's *latest* snapshot. (You can also use `#-1` if
you want to be explicit).

Is there any different between `aws-beta` and `aws-prod` right now?

```shell {.command}
kosli env diff aws-beta aws-prod
```

You'll probably get no output here, meaning there is no difference.
But if, for example, there is a current deployment to a cyber-dojo 
repository awaiting a manual approval then 
something will be running in `aws-beta` but not in `aws-prod`
and you'll see this difference. 

<!-- Add example of two specific snappishes where this happened or was forced/simulated.
    Make the git commit lead to an Easter-egg with a nice comment/git-message. 
    Maybe the Easter-egg could be the answer to a riddle
    and at the USA conferences we could have a biggish prize for the first person
    who follows this tutorial and finds the answer to the riddle.
-->

If someone has somehow managed to run a rogue service in one of the
environments (but not the other) that will show up in a diff.

<!-- add an example of this that is again forced/simulated -->


<!-- 
This we would like to show the users:
- Kosli gives developers without access to production environment information about what is running.
- Detect that a new "bit-coin miner" is running in your environment. Rogue artifact detection.
- Kosli can show that a deployment is reported, but artifact didn't start. Find this in artifact view.
- Kosli can show that an artifact started, but no deployment was reported for it.
- Detect an artifact that is missing evidence is running in an environment
- Do we want to mention the whole env being compliant?
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