---
title: "Querying Kosli"
bookCollapseSection: false
weight: 506
---

# Querying Kosli

All the information stored in Kosli may be helpful both for operations and development. A set of `get`, `list`, `log` and `assert` commands allows you to quickly access the information about your environments, artifacts and deployments, without leaving your development environment.

## Getting ready

You need to:
* [Install Kosli CLI](/getting_started/install/).
* [Get a Kosli API token](/getting_started/service-accounts/).
* Set the `KOSLI_ORG` environment variable to `cyber-dojo` (the Kosli `cyber-dojo` organization is public so any authenticated user can read its data) and `KOSLI_API_TOKEN` to your token:
  ```shell {.command}
  export KOSLI_ORG=cyber-dojo # cyber-dojo is a public demo org
  export KOSLI_API_TOKEN=<your-api-token>
  ```

## Search with git commit sha

You can use `kosli search` command to find out if Kosli knows of any artifact that was build using that commit - both short and full shas are accepted:

```
$ kosli search 0f5c9e1
Search result resolved to commit 0f5c9e19c4d4f948d19ce4c8495b2a44745cda96
Name:              cyberdojo/web:0f5c9e1
Fingerprint:       62e1d2909cc59193b31bfd120276fcb8ba5e42dd6becd873218a41e4ce022505
Has provenance:    true
Flow:              web
Git commit:        0f5c9e19c4d4f948d19ce4c8495b2a44745cda96
Commit URL:        https://github.com/cyber-dojo/web/commit/0f5c9e19c4d4f948d19ce4c8495b2a44745cda96
Build URL:         https://github.com/cyber-dojo/web/actions/runs/3021563461
Compliance state:  COMPLIANT
History:
    Artifact created                                   Fri, 09 Sep 2022 11:59:50 CEST
    Deployment #59 to aws-beta environment             Fri, 09 Sep 2022 12:01:12 CEST
    Started running in aws-beta#217 environment        Fri, 09 Sep 2022 12:02:42 CEST
    Deployment #60 to aws-prod environment             Fri, 09 Sep 2022 12:06:37 CEST
    Started running in aws-prod#202 environment        Fri, 09 Sep 2022 12:07:28 CEST
    Scaled up from 1 to 3 in aws-prod#203 environment  Fri, 09 Sep 2022 12:08:28 CEST
    No longer running in aws-beta#222 environment      Sat, 10 Sep 2022 08:44:42 CEST
    No longer running in aws-prod#210 environment      Sat, 10 Sep 2022 08:49:28 CEST
```

The information returned by `kosli search` - like Flow, Fingerprint or History - can be used to run more dedicated searches in Kosli. 

## Search for a flow

When you search in Kosli you often need to refer to a specific flow. If you don't remember all the flows' names it is easy to list them with `kosli list flows` command:

```
$ kosli list flows 
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

And if you want to check metadata of a specific flow (like description or template) use `kosli get flow`

```
$ kosli get flow creator
Name:                creator
Description:         UX for Group/Kata creation
Visibility:          public
Template:            [artifact, branch-coverage]
Last Deployment At:  Wed, 14 Sep 2022 10:51:43 CEST • one month ago
```

## List artifacts

To find the information about artifacts reported to a specific flow in Kosli use `kosli list artifacts` command

```
$ kosli list artifacts --flow creator
COMMIT   ARTIFACT                                  STATE       CREATED_AT
344430d  Name: cyberdojo/creator:344430d           COMPLIANT   Wed, 14 Sep 2022 10:48:09 CEST
         Fingerprint: 817a72(...)6b5a273399c693             
                                                                                                    
41bfb7b  Name: cyberdojo/creator:41bfb7b           COMPLIANT   Sat, 10 Sep 2022 08:41:15 CEST
         Fingerprint: 8d6fef(...)b84c281f712ef8             
                                                                                                    
aa0a3d3  Name: cyberdojo/creator:aa0a3d3           COMPLIANT   Fri, 09 Sep 2022 11:58:56 CEST
         Fingerprint: 3ede07(...)238845a631e96a             
                                                                                                    
[...]
```

The output of the command is shortened above, for readability purposes. 

The amount of artifacts may be really long and by default you can see the last 15 artifacts - the first page of the result list. You can use `-n` flag to limit the amount of artifacts displayed per page, and `--page` to select which page of the result list you want to see.

E.g. to see last five artifacts you'd use:
```
$ kosli list artifacts --flow creator -n 5
```

And to see the next page:
```
$ kosli list artifacts --flow creator -n 5 --page 2
```

You can also use the `--output` flag to change the format of the response. By default the response comes in a *table* format, but you can choose to switch to *json*:
```
$ kosli list artifacts --flow creator --output json
```
## Get artifact

To get more detailed information about a given artifact use `kosli get artifact`. To identify the artifact you need to use:
* flow name followed by `@` and artifact fingerprint
OR
* flow name followed by `:` and commit sha

Both are available in the output of `kosli list artifacts` command

```
# search for artifact by its fingerprint
$ kosli get artifact creator@817a72609041c51cd2a3bbbcbeb048c687677986b5a273399c6938b5e6aa1ded
Name:                     cyberdojo/creator:344430d
Flow:                     creator
Fingerprint:              817a72609041c51cd2a3bbbcbeb048c687677986b5a273399c6938b5e6aa1ded
Created on:               Wed, 14 Sep 2022 10:48:09 CEST • 2 months ago
Git commit:               344430d530d26068aa1f39760a9c094c989382f3
Commit URL:               https://github.com/cyber-dojo/creator/commit/344430d530d26068aa1f39760a9c094c989382f3
Build URL:                https://github.com/cyber-dojo/creator/actions/runs/3051390570
State:                    COMPLIANT
Running in environments:  aws-beta#265, aws-prod#259
History:
    Artifact created                               Wed, 14 Sep 2022 10:48:09 CEST
    branch-coverage evidence received              Wed, 14 Sep 2022 10:49:11 CEST
    Deployment #100 to aws-beta environment        Wed, 14 Sep 2022 10:50:40 CEST
    Deployment #101 to aws-prod environment        Wed, 14 Sep 2022 10:51:43 CEST
    Started running in aws-beta#229 environment    Wed, 14 Sep 2022 10:52:42 CEST
    Started running in aws-prod#217 environment    Wed, 14 Sep 2022 10:53:28 CEST
    No longer running in aws-prod#252 environment  Fri, 14 Oct 2022 08:17:28 CEST
    Started running in aws-prod#254 environment    Fri, 14 Oct 2022 08:22:28 CEST
    No longer running in aws-beta#254 environment  Fri, 14 Oct 2022 16:35:42 CEST
    Started running in aws-beta#256 environment    Fri, 14 Oct 2022 16:38:42 CEST
    No longer running in aws-beta#257 environment  Sun, 16 Oct 2022 07:45:42 CEST
    Started running in aws-beta#259 environment    Sun, 16 Oct 2022 07:49:42 CEST
    No longer running in aws-beta#260 environment  Wed, 19 Oct 2022 09:28:42 CEST
    Started running in aws-beta#262 environment    Wed, 19 Oct 2022 09:32:42 CEST
    No longer running in aws-beta#263 environment  Wed, 19 Oct 2022 09:42:42 CEST
    Started running in aws-beta#265 environment    Wed, 19 Oct 2022 09:46:42 CEST
    No longer running in aws-prod#257 environment  Fri, 21 Oct 2022 11:02:28 CEST
    Started running in aws-prod#259 environment    Fri, 21 Oct 2022 11:05:28 CEST

# search for artifact by its commit sha
$ kosli get artifact creator:344430d
Name:                     cyberdojo/creator:344430d
Flow:                     creator
Fingerprint:              817a72609041c51cd2a3bbbcbeb048c687677986b5a273399c6938b5e6aa1ded
Created on:               Wed, 14 Sep 2022 10:48:09 CEST • 2 months ago
Git commit:               344430d530d26068aa1f39760a9c094c989382f3
Commit URL:               https://github.com/cyber-dojo/creator/commit/344430d530d26068aa1f39760a9c094c989382f3
Build URL:                https://github.com/cyber-dojo/creator/actions/runs/3051390570
State:                    COMPLIANT
Running in environments:  aws-beta#265, aws-prod#259
History:
    Artifact created                               Wed, 14 Sep 2022 10:48:09 CEST
    branch-coverage evidence received              Wed, 14 Sep 2022 10:49:11 CEST
    Deployment #100 to aws-beta environment        Wed, 14 Sep 2022 10:50:40 CEST
    Deployment #101 to aws-prod environment        Wed, 14 Sep 2022 10:51:43 CEST
    Started running in aws-beta#229 environment    Wed, 14 Sep 2022 10:52:42 CEST
    Started running in aws-prod#217 environment    Wed, 14 Sep 2022 10:53:28 CEST
    No longer running in aws-prod#252 environment  Fri, 14 Oct 2022 08:17:28 CEST
    Started running in aws-prod#254 environment    Fri, 14 Oct 2022 08:22:28 CEST
    No longer running in aws-beta#254 environment  Fri, 14 Oct 2022 16:35:42 CEST
    Started running in aws-beta#256 environment    Fri, 14 Oct 2022 16:38:42 CEST
    No longer running in aws-beta#257 environment  Sun, 16 Oct 2022 07:45:42 CEST
    Started running in aws-beta#259 environment    Sun, 16 Oct 2022 07:49:42 CEST
    No longer running in aws-beta#260 environment  Wed, 19 Oct 2022 09:28:42 CEST
    Started running in aws-beta#262 environment    Wed, 19 Oct 2022 09:32:42 CEST
    No longer running in aws-beta#263 environment  Wed, 19 Oct 2022 09:42:42 CEST
    Started running in aws-beta#265 environment    Wed, 19 Oct 2022 09:46:42 CEST
    No longer running in aws-prod#257 environment  Fri, 21 Oct 2022 11:02:28 CEST
    Started running in aws-prod#259 environment    Fri, 21 Oct 2022 11:05:28 CEST
```

## Search for an environment

As is the case for flows and artifacts, you can list all the Kosli environments you created under your organization

```
$ kosli list environments
NAME      TYPE  LAST REPORT                LAST MODIFIED
aws-beta  ECS   2022-10-30T14:51:42+01:00  2022-10-30T14:51:42+01:00
aws-prod  ECS   2022-10-30T14:51:28+01:00  2022-10-30T14:51:28+01:00
beta      K8S   2022-06-15T11:39:59+02:00  2022-06-15T11:39:59+02:00
prod      K8S   2022-06-15T11:40:01+02:00  2022-06-15T11:40:01+02:00
```

And get the metadata (including the type) of each environment:

```
$ kosli get environment aws-beta
Name:              aws-beta
Type:              ECS
Description:       The ECS beta namespace
State:             COMPLIANT
Last Reported At:  Sun, 30 Oct 2022 14:55:42 CET • 5 seconds ago
```

## Get environment events

When you have the name of the environment you want to dig into use `kosli list snapshots` or `kosli log environment` to browse snapshots and changes in the environment, or `kosli get snapshot` to have a look at a specific snapshot.

```
$ kosli list snapshots aws-beta
SNAPSHOT  FROM                            TO                              DURATION
266       Wed, 19 Oct 2022 09:47:42 CEST  now                             11 days
265       Wed, 19 Oct 2022 09:46:42 CEST  Wed, 19 Oct 2022 09:47:42 CEST  59 seconds
264       Wed, 19 Oct 2022 09:45:42 CEST  Wed, 19 Oct 2022 09:46:42 CEST  about a minute
263       Wed, 19 Oct 2022 09:42:42 CEST  Wed, 19 Oct 2022 09:45:42 CEST  3 minutes
262       Wed, 19 Oct 2022 09:32:42 CEST  Wed, 19 Oct 2022 09:42:42 CEST  10 minutes
261       Wed, 19 Oct 2022 09:31:42 CEST  Wed, 19 Oct 2022 09:32:42 CEST  about a minute
260       Wed, 19 Oct 2022 09:28:42 CEST  Wed, 19 Oct 2022 09:31:42 CEST  3 minutes
259       Sun, 16 Oct 2022 07:49:42 CEST  Wed, 19 Oct 2022 09:28:42 CEST  3 days
258       Sun, 16 Oct 2022 07:48:42 CEST  Sun, 16 Oct 2022 07:49:42 CEST  59 seconds
257       Sun, 16 Oct 2022 07:45:42 CEST  Sun, 16 Oct 2022 07:48:42 CEST  3 minutes
256       Fri, 14 Oct 2022 16:38:42 CEST  Sun, 16 Oct 2022 07:45:42 CEST  2 days
255       Fri, 14 Oct 2022 16:37:42 CEST  Fri, 14 Oct 2022 16:38:42 CEST  about a minute
254       Fri, 14 Oct 2022 16:35:42 CEST  Fri, 14 Oct 2022 16:37:42 CEST  2 minutes
253       Thu, 13 Oct 2022 09:04:42 CEST  Fri, 14 Oct 2022 16:35:42 CEST  one day
252       Mon, 10 Oct 2022 08:47:42 CEST  Thu, 13 Oct 2022 09:04:42 CEST  3 days
```

By default you can see the last 15 changes to the environment. You can choose to only print e.g. last 3 events (`-n` flag).

You can also choose to see the actual events from each snapshot, using `kosli log environment` command:

```
$ kosli log environment aws-beta
SNAPSHOT  EVENT                                                                          FLOW       DEPLOYMENTS
#266      Artifact: 244531986313.dkr.ecr.eu-central-1.amazonaws.com/dashboard:d90a3e4    dashboard  #15 
          Fingerprint: dd5308fdcda117c1ff3963e192a069ae390c2fe9e10e8abfa2430224265efe98             
          Description: 1 instance started running (from 0 to 1)                                     
          Reported at: Wed, 19 Oct 2022 09:47:42 CEST                                               
                                                                                                    
#265      Artifact: 244531986313.dkr.ecr.eu-central-1.amazonaws.com/web:7ac7cdc          web        #63 
          Fingerprint: 88c082eee192653ea5826d14f714bcfbdadbd1827a7a29416bfddbdff2b69507             
          Description: 3 instances started running (from 0 to 3)                                    
          Reported at: Wed, 19 Oct 2022 09:46:42 CEST                                               
                                                                                                    
#265      Artifact: 244531986313.dkr.ecr.eu-central-1.amazonaws.com/runner:2872115       runner     #24 
          Fingerprint: 9461946e43393404ce744292331e7efbfe4e17cc2e5a32972169a90c81ec875c             
          Description: 3 instances started running (from 0 to 3)                                    
          Reported at: Wed, 19 Oct 2022 09:46:42 CEST  
```

You can also use an *interval* expression, like `262..264` (to see specified snapshot list) , or `~4..NOW` (to get a list of snapshots starting from 4 behind a currently running one and the current one)

```
$ kosli log environment aws-beta 262..264
SNAPSHOT  FROM                            TO                              DURATION
264       Wed, 19 Oct 2022 09:45:42 CEST  Wed, 19 Oct 2022 09:46:42 CEST  about a minute
263       Wed, 19 Oct 2022 09:42:42 CEST  Wed, 19 Oct 2022 09:45:42 CEST  3 minutes
262       Wed, 19 Oct 2022 09:32:42 CEST  Wed, 19 Oct 2022 09:42:42 CEST  10 minutes

$ kosli log environment aws-beta ~4..NOW
SNAPSHOT  FROM                            TO                              DURATION
266       Wed, 19 Oct 2022 09:47:42 CEST  now                             11 days
265       Wed, 19 Oct 2022 09:46:42 CEST  Wed, 19 Oct 2022 09:47:42 CEST  59 seconds
264       Wed, 19 Oct 2022 09:45:42 CEST  Wed, 19 Oct 2022 09:46:42 CEST  about a minute
263       Wed, 19 Oct 2022 09:42:42 CEST  Wed, 19 Oct 2022 09:45:42 CEST  3 minutes
262       Wed, 19 Oct 2022 09:32:42 CEST  Wed, 19 Oct 2022 09:42:42 CEST  10 minutes
```

## Get a snapshot 

To have a look at what is or was running in a given snapshot use `kosli get snapshot` command. You can use just the environment name as the argument, which will give you the latest snapshot, add `#` and snapshot number, to get a specific one, or `~n` where *n* is a number, to get *n-th* snapshot behind a current one:

``` 
$ kosli get snapshot aws-beta
COMMIT   ARTIFACT                                                                              FLOW      RUNNING_SINCE  REPLICAS
d90a3e4  Name: 244531986313.dkr.ecr.eu-central-1.amazonaws.com/dashboard:d90a3e4               N/A       11 days ago    1
         Fingerprint: dd5308fdcda117c1ff3963e192a069ae390c2fe9e10e8abfa2430224265efe98                                  
                                                                                                                        
9f669e5  Name: 244531986313.dkr.ecr.eu-central-1.amazonaws.com/languages-start-points:9f669e5  N/A       11 days ago    1
         Fingerprint: e6b72f6a41d0944824538334120804ccde795b4b5aeb8aa311dbc0721b4e40fd                                  
                                                                                                                        
1c162e4  Name: 244531986313.dkr.ecr.eu-central-1.amazonaws.com/differ:1c162e4                  N/A       11 days ago    1
         Fingerprint: b7fd766dd2514b2610c0c8d70d8f762de4921931f97fdd6fbbfcc9745ac3ce3b                                  
[...]

$ kosli get snapshot aws-beta#256
COMMIT   ARTIFACT                                                                              FLOW      RUNNING_SINCE  REPLICAS
6fe0d30  Name: 244531986313.dkr.ecr.eu-central-1.amazonaws.com/repler:6fe0d30                  N/A       16 days ago    1
         Fingerprint: a0c03099c832e4ce5f23f5e33dac9889c0b7ccd61297fffdaf1c67e7b99e6f8f                                  
                                                                                                                        
d90a3e4  Name: 244531986313.dkr.ecr.eu-central-1.amazonaws.com/dashboard:d90a3e4               N/A       16 days ago    1
         Fingerprint: dd5308fdcda117c1ff3963e192a069ae390c2fe9e10e8abfa2430224265efe98                                  
                                                                                                                        
1c162e4  Name: 244531986313.dkr.ecr.eu-central-1.amazonaws.com/differ:1c162e4                  N/A       16 days ago    1
         Fingerprint: b7fd766dd2514b2610c0c8d70d8f762de4921931f97fdd6fbbfcc9745ac3ce3b                                  
[...]

$ kosli get snapshot aws-beta~19
COMMIT   ARTIFACT                                                                              FLOW      RUNNING_SINCE  REPLICAS
2e8646c  Name: 244531986313.dkr.ecr.eu-central-1.amazonaws.com/shas:2e8646c                    N/A       one month ago  1
         Fingerprint: a3158c3e79c83905fd3613e06b8cf5a45141c50cf49d4f99de90a2d081b77771                                  
                                                                                                                        
344430d  Name: 244531986313.dkr.ecr.eu-central-1.amazonaws.com/creator:344430d                 N/A       2 months ago   1
         Fingerprint: 817a72609041c51cd2a3bbbcbeb048c687677986b5a273399c6938b5e6aa1ded                                  
                                                                                                                        
7ac7cdc  Name: 244531986313.dkr.ecr.eu-central-1.amazonaws.com/web:7ac7cdc                     N/A       2 months ago   3
         Fingerprint: 88c082eee192653ea5826d14f714bcfbdadbd1827a7a29416bfddbdff2b69507                                 

```

The same expressions (with `#` and `~`) can be used to reference snapshots when diffing environment.

In the example below there was only one difference between snapshots: one new artifact started running in the latest snapshot. 

```
$ kosli diff snapshots aws-beta aws-beta~1
Only present in aws-beta (snapshot: aws-beta#266)
                   
     Name:         244531986313.dkr.ecr.eu-central-1.amazonaws.com/dashboard:d90a3e4
     Fingerprint:  dd5308fdcda117c1ff3963e192a069ae390c2fe9e10e8abfa2430224265efe98
     Flow:         dashboard
     Commit URL:   https://github.com/cyber-dojo/dashboard/commit/d90a3e481d57023816f6694ba4252342889405eb
     Started:      Wed, 19 Oct 2022 09:47:33 CEST • 11 days ago
```

## Diff environments/snapshots

You can use `diff` to compare snapshots of two different environments or different snapshots of the same environment:

```
$ kosli diff snapshots aws-beta~3 aws-prod
Only present in aws-prod (snapshot: aws-prod#261)
                   
     Name:         274425519734.dkr.ecr.eu-central-1.amazonaws.com/saver:8d724a1
     Fingerprint:  3e52f9b838cbb4e31455524c908eb8dd878b2ae25144427de8160f6658ee191f
     Flow:         saver
     Commit URL:   https://github.com/cyber-dojo/saver/commit/8d724a14c6e95947f0c56ad6af8251bca656a599
     Started:      Fri, 21 Oct 2022 11:04:59 CEST • 9 days ago
                   
     Name:         274425519734.dkr.ecr.eu-central-1.amazonaws.com/nginx:d491f5c
     Fingerprint:  4f66ab1b0a7a9f7ed064a3b1033a53ec7dd99359ff68d509ab555dcf4516b23e
     Flow:         nginx
     Commit URL:   https://github.com/cyber-dojo/nginx/commit/d491f5c06babe70bfebe2f9df0a9a66db7957f17
     Started:      Fri, 21 Oct 2022 11:03:53 CEST • 9 days ago
                   
     Name:         274425519734.dkr.ecr.eu-central-1.amazonaws.com/custom-start-points:8c551d3
     Fingerprint:  76ad6ffc1828d8213a39bc39c879b3c35a75d4705d1d8df5977a87a11e6ae25e
     Flow:         custom-start-points
     Commit URL:   https://github.com/cyber-dojo/custom-start-points/commit/8c551d378051b6ef1fde7fd58aaced1047264405
     Started:      Fri, 21 Oct 2022 11:04:30 CEST • 9 days ago
[...]
```