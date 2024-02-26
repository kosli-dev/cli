---
title: Tracing a production incident back to git commits
bookCollapseSection: false
weight: 520
draft: false
---

<!-- Add Easter-eggs comments? -->

# Tracing a production incident back to git commits

In this 5 minute tutorial you'll learn how Kosli can track a production incident in Cyber-dojo back to git commits.

Something has gone wrong and [https://cyber-dojo.org](https://cyber-dojo.org) is displaying a 500 error!


{{< figure src="/images/cyber-dojo-prod-500-large.png" alt="Prod cyber-dojo is down with a 500" width="90%" >}}

It was working an hour ago. What has happened in the last hour?

## Getting ready

You need to:
* [Install Kosli CLI](/getting_started/install/).
* [Get a Kosli API token](/getting_started/service-accounts/).
* Set the `KOSLI_ORG` environment variable to `cyber-dojo` (the Kosli `cyber-dojo` organization is public so any authenticated user can read its data) and `KOSLI_API_TOKEN` to your token:
  ```shell {.command}
  export KOSLI_ORG=cyber-dojo
  export KOSLI_API_TOKEN=<your-api-token>
  ```

## Start with the environment

[https://cyber-dojo.org](https://cyber-dojo.org) is running in an AWS environment
that reports to Kosli as `aws-prod`.  
Get a log of this environment's changes:

```shell {.command}
kosli log env aws-prod
```

At the time this tutorial was written the output of this command
displayed the first page of 177 snapshots. 
You will see the first page of considerably more than 177 snapshots because 
`aws-prod` has moved on since this incident (it has been resolved with new 
commits which have created new deployments). 
To limit the output you can set the interval for the command:

```shell {.command}
kosli log env aws-prod --interval 176..177
```

The output should be:

```plaintext {.light-console}
SNAPSHOT  EVENT                                                                          FLOW      DEPLOYMENTS
#177      Artifact: 274425519734.dkr.ecr.eu-central-1.amazonaws.com/creator:31dee35      creator   #87 
          Fingerprint: 5d1c926530213dadd5c9fcbf59c8822da56e32a04b0f9c774d7cdde3cf6ba66d             
          Description: 1 instance stopped running (from 1 to 0).                               
          Reported at: Tue, 06 Sep 2022 16:53:28 CEST                                          
                                                                                               
#176      Artifact: 274425519734.dkr.ecr.eu-central-1.amazonaws.com/creator:b7a5908      creator   #89 
          Fingerprint: 860ad172ace5aee03e6a1e3492a88b3315ecac2a899d4f159f43ca7314290d5a             
          Description: 1 instance started running (from 0 to 1).                               
          Reported at: Tue, 06 Sep 2022 16:52:28 CEST
```

These two snapshots belong to the same blue-green deployment.
You see artifact `creator:b7a5908` starting in snapshot #176, and artifact
`creator:31dee35` exiting in snapshot #177.

## Dig into the artifact

You are interested in #176, showing the newly running artifact, `creator:b7a5908`,
with the fingerprint starting `860ad17`.

Let's learn more about this artifact:

```shell {.command}
kosli get artifact creator@860ad17
```

```plaintext {.light-console}
Name:        cyberdojo/creator:b7a5908
Flow:        creator
Fingerprint: 860ad172ace5aee03e6a1e3492a88b3315ecac2a899d4f159f43ca7314290d5a
Created on:  Tue, 06 Sep 2022 16:48:07 CEST • 21 hours ago
Git commit:  b7a590836cf140e17da3f01eadd5eca17d9efc65
Commit URL:  https://github.com/cyber-dojo/creator/commit/b7a590836cf140e17da3f01eadd5eca17d9efc65
Build URL:   https://github.com/cyber-dojo/creator/actions/runs/3001102984
State:       COMPLIANT
History:  
    Artifact created                               Tue, 06 Sep 2022 16:48:07 CEST
    Deployment #88 to aws-beta environment         Tue, 06 Sep 2022 16:49:59 CEST
    Deployment #89 to aws-prod environment         Tue, 06 Sep 2022 16:51:12 CEST
    Started running in aws-beta#196 environment    Tue, 06 Sep 2022 16:51:42 CEST
    Started running in aws-prod#176 environment    Tue, 06 Sep 2022 16:52:28 CEST
```

## Follow to the commit

You can follow the [commit URL](https://github.com/cyber-dojo/creator/commit/b7a590836cf140e17da3f01eadd5eca17d9efc65).

{{< figure src="/images/cyber-dojo-github-diff.png" alt="cyber-dojo github diff" width="500" >}}

The incident was caused by a simple typo in the `app.rb` file!

Perhaps someone accidentally inserted the "s" while trying to save the file?
Either way, this is clearly the problem because the function is called `respond_to` without the `s`.

You were able to trace the problem back to a specific commit without any access to cyber-dojo's `aws-prod` environment.

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

- See what software is/was running where which is useful in debugging.
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