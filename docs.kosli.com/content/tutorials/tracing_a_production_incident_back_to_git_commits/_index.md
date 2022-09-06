---
title: Tracing a production incident back to git commits
bookCollapseSection: false
weight: 3
draft: true
---

<!-- Create a SECOND tutorial for: 
     2. Title?=Tracing a production incident back to git commits
        The stories here would be simulated incidents with Easter-eggs comments.

-->

![Beta cyber-dojo is down with a 500](/images/cyber-dojo-500.png)

The command below will probably give you a different output since the two environments have moved on after this incident 
(the incident has been resolved with new commits which created new deployments).
```
kosli env diff aws-beta aws-prod
- Name:   244531986313.dkr.ecr.eu-central-1.amazonaws.com/web:c3ada4d
  Sha256: 480735698cb9e468bb16c4265fedb7507640d236b3ab53cf2e3ec09d3bd72063
  Pipeline: web
  Commit: https://github.com/cyber-dojo/web/commit/c3ada4dbd6bb9a66c27f24cec4b5a4c25cf9ce2b
  Started: Tue, 06 Sep 2022 10:28:40 CEST • 13 minutes ago

+ Name:   274425519734.dkr.ecr.eu-central-1.amazonaws.com/web:6b1b6bc
  Sha256: d4ab88ce200e88a07eda3c33fb18d7051a586e6b8e900fcea1063a13c4506446
  Pipeline: web
  Commit: https://github.com/cyber-dojo/web/commit/6b1b6bc45af830836838db8644d1388726d8f381
  Started: Fri, 02 Sep 2022 05:32:00 CEST • 4 days ago
```

```
kosli artifact get web@d4ab88ce200e88a07eda3c33fb18d7051a586e6b8e900fcea1063a13c4506446
Name:        cyberdojo/web:6b1b6bc
SHA256:      d4ab88ce200e88a07eda3c33fb18d7051a586e6b8e900fcea1063a13c4506446
Created on:  Fri, 02 Sep 2022 05:26:52 CEST • 4 days ago
Git commit:  6b1b6bc45af830836838db8644d1388726d8f381
Commit URL:  https://github.com/cyber-dojo/web/commit/6b1b6bc45af830836838db8644d1388726d8f381
Build URL:   https://github.com/cyber-dojo/web/actions/runs/2976432579
State:       COMPLIANT
History:
    Artifact created                               Fri, 02 Sep 2022 05:26:52 CEST
    Deployment #48 to aws-beta environment         Fri, 02 Sep 2022 05:28:07 CEST
    Started running in aws-beta#181 environment    Fri, 02 Sep 2022 05:29:42 CEST
    Deployment #49 to aws-prod environment         Fri, 02 Sep 2022 05:31:08 CEST
    Started running in aws-prod#160 environment    Fri, 02 Sep 2022 05:32:28 CEST
    No longer running in aws-beta#187 environment  Tue, 06 Sep 2022 10:30:42 CEST
```
<!-- 
Assume that we are continuing after the following a git commit so we don't need to
explain what cyber dojo is.

During xx we detected that yy did not work.

We diff what was running at that point in time with the previous snapshot

We see that `runner` has changed and the new artifact is zz.

From the artifact we can find the git commit that was used for this build.

We should be able to find out which commit was used for building the previous
version of this artifact.

We now have two git commits and we know that the bug was introduced between those
two commits.
-->

## Diffing snapshots across environments!

<!-- This is really part of a separate tutorial -->

The name of an environment without a snapshot number (or the `#` character)
specifies that environment's *latest* snapshot. (You can also use `#-1` if
you want to be explicit).
<!-- Tore: I thought it was #NOW that is the explicit of latest -->

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