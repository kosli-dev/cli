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

![Prod cyber-dojo is down with a 500](/images/cyber-dojo-prod-500-large.png)
![Prod cyber-dojo is down with a 500](/images/cyber-dojo-prod-500-small.png)

<!--
```shell {.command}
kosli env diff aws-prod~1 aws-prod
```
-->

```shell {.command}
kosli env log aws-prod
```

```console
SNAPSHOT  FROM                            TO                              DURATION
177       Tue, 06 Sep 2022 16:53:28 CEST  now                             3 minutes
176       Tue, 06 Sep 2022 16:52:28 CEST  Tue, 06 Sep 2022 16:53:28 CEST  59 seconds
175       Tue, 06 Sep 2022 16:43:28 CEST  Tue, 06 Sep 2022 16:52:28 CEST  9 minutes
174       Tue, 06 Sep 2022 16:42:28 CEST  Tue, 06 Sep 2022 16:43:28 CEST  59 seconds
173       Tue, 06 Sep 2022 15:53:28 CEST  Tue, 06 Sep 2022 16:42:28 CEST  49 minutes
172       Tue, 06 Sep 2022 15:52:28 CEST  Tue, 06 Sep 2022 15:53:28 CEST  about a minute
171       Tue, 06 Sep 2022 15:28:27 CEST  Tue, 06 Sep 2022 15:52:28 CEST  24 minutes
170       Tue, 06 Sep 2022 15:27:28 CEST  Tue, 06 Sep 2022 15:28:27 CEST  59 seconds
169       Mon, 05 Sep 2022 08:00:28 CEST  Tue, 06 Sep 2022 15:27:28 CEST  one day
168       Mon, 05 Sep 2022 07:59:28 CEST  Mon, 05 Sep 2022 08:00:28 CEST  about a minute
167       Sun, 04 Sep 2022 03:11:28 CEST  Mon, 05 Sep 2022 07:59:28 CEST  one day
166       Sun, 04 Sep 2022 03:10:28 CEST  Sun, 04 Sep 2022 03:11:28 CEST  59 seconds
165       Fri, 02 Sep 2022 19:28:28 CEST  Sun, 04 Sep 2022 03:10:28 CEST  one day
164       Fri, 02 Sep 2022 19:27:28 CEST  Fri, 02 Sep 2022 19:28:28 CEST  59 seconds
163       Fri, 02 Sep 2022 05:37:28 CEST  Fri, 02 Sep 2022 19:27:28 CEST  14 hours
```

The ouput for the command above will show more than 177 snapshots because `aws-prod` has moved on since this incident 
(it has been resolved with new commits which have created new deployments).

We look at the diff between the most recent snapshots at this moment:

```shell {.command}
kosli env diff aws-prod#176 aws-prod#177
```

```console
aws-prod#176 only
  Name: 274425519734.dkr.ecr.eu-central-1.amazonaws.com/creator:31dee35
    Sha256: 5d1c926530213dadd5c9fcbf59c8822da56e32a04b0f9c774d7cdde3cf6ba66d
    Pipeline: creator
    Commit: https://github.com/cyber-dojo/creator/commit/31dee357844ab2f58a40d16f945dc5e35dcd6bc0
    Started: Tue, 06 Sep 2022 16:41:44 CEST • 15 minutes ago
```

This tells us artifact `creator:31dee35` is no longer running in the `aws-prod` environment.

<!-- We maybe want to make this even more obvious -->

We look at the diff between the previous snapshots:

```shell {.command}
kosli env diff aws-prod#175 aws-prod#176
```

```console
aws-prod#176 only
  Name: 274425519734.dkr.ecr.eu-central-1.amazonaws.com/creator:b7a5908
    Sha256: 860ad172ace5aee03e6a1e3492a88b3315ecac2a899d4f159f43ca7314290d5a
    Pipeline: creator
    Commit: https://github.com/cyber-dojo/creator/commit/b7a590836cf140e17da3f01eadd5eca17d9efc65
    Started: Tue, 06 Sep 2022 16:52:03 CEST • 6 minutes ago
```
<!-- Do we want the label for Commit: to be Commit URL: to match the
     label you see in a `kosli artifact get` command
-->

This tells us artifact `creator:b7a5908` has started running in the `aws-prod` environment.

Now follow the commit URL.

<!--  TODO: fix cyber-dojo creator repo from the hacks we did to simulate bug -->








{{< hint info >}}
To actually see the same output we had when this document was written run
```shell {.command}
kosli env diff aws-beta#189 aws-prod#169
```
{{< \hint >}}

<!-- We should return the fully expanded snappish as the key in the json -->

```console
aws-beta only
  Name: 244531986313.dkr.ecr.eu-central-1.amazonaws.com/web:c3ada4d
    Sha256: 480735698cb9e468bb16c4265fedb7507640d236b3ab53cf2e3ec09d3bd72063
    Pipeline: web
    Commit: https://github.com/cyber-dojo/web/commit/c3ada4dbd6bb9a66c27f24cec4b5a4c25cf9ce2b
    Started: Tue, 06 Sep 2022 10:28:40 CEST • 4 hours ago

aws-prod only
  Name: 274425519734.dkr.ecr.eu-central-1.amazonaws.com/web:6b1b6bc
    Sha256: d4ab88ce200e88a07eda3c33fb18d7051a586e6b8e900fcea1063a13c4506446
    Pipeline: web
    Commit: https://github.com/cyber-dojo/web/commit/6b1b6bc45af830836838db8644d1388726d8f381
    Started: Fri, 02 Sep 2022 05:32:00 CEST • 4 days ago
```




```shell {.command}
kosli artifact get web@480735698cb9e468bb16c4265fedb7507640d236b3ab53cf2e3ec09d3bd72063
```

```console
Name:        cyberdojo/web:c3ada4d
SHA256:      480735698cb9e468bb16c4265fedb7507640d236b3ab53cf2e3ec09d3bd72063
Created on:  Tue, 06 Sep 2022 10:26:29 CEST • 4 hours ago
Git commit:  c3ada4dbd6bb9a66c27f24cec4b5a4c25cf9ce2b
Commit URL:  https://github.com/cyber-dojo/web/commit/c3ada4dbd6bb9a66c27f24cec4b5a4c25cf9ce2b
Build URL:   https://github.com/cyber-dojo/web/actions/runs/2998616599
State:       COMPLIANT
History:
    Artifact created                             Tue, 06 Sep 2022 10:26:29 CEST
    Deployment #50 to aws-beta environment       Tue, 06 Sep 2022 10:27:42 CEST
    Started running in aws-beta#188 environment  Tue, 06 Sep 2022 10:29:22 CEST    
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