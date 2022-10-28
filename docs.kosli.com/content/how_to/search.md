---
title: Search
bookCollapseSection: false
weight: 50
---
# Get your artifacts and environments information from Kosli

All the information stored in Kosli may be helpful both for operations and development. A set of `get`, `ls`, `log` and `inspect` commands allows you to quickly access the information about your environments, artifacts and deployments, without leaving your development environment.

The same CLI you use to record and connect your changes can be use to search for and browse information in Kosli.

To make it easier to run Kosli search command with the CLI you can export the `owner` and `api-token` as environment variables, so you don't have to provide them every time you run commands. This approach is valid for [any flag](/introducing_kosli/cli/#environment-variables) 

```
export KOSLI_OWNER=yourOrganizationName
export KOSLI_API_TOKEN=yourKosliApiToken
```

You can try all the commands below by setting the `owner` to `cyber-dojo`.  
the Kosli cyber-dojo organization is public so any authenticated user can read its data:

```
export KOSLI_OWNER=cyber-dojo
```

## Search for a pipeline

When you search in Kosli you often need to refer to a specific pipeline. If you don't remember all the pipelines' names it is easy to find with `kosli pipeline ls` command:

```
% kosli pipeline ls 
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

And if you want to check metadata of a specific pipeline (like description or template) use `kosli pipeline inspect`

```
% kosli pipeline inspect creator
Name:                creator
Description:         UX for Group/Kata creation
Visibility:          public
Template:            [artifact, branch-coverage]
Last Deployment At:  Wed, 14 Sep 2022 10:51:43 CEST • one month ago
```

## Search for an artifact

To find the information about artifacts reported to a specific pipeline in Kosli use `kosli artifact ls` command

```
% kosli artifact ls creator
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

The amount of artifacts may be really long and by default you can see last 15 artifacts - the first page of the result list. You can use `-n` flag to limit the amount of artifacts displayed per page, and `--page` to select which page of the result list you want to see.

E.g. to see last five artifacts you'd use:
```
kosli artifact ls creator -n 5
```

And to see the next page:
```
kosli artifact ls creator -n 5 --page 2
```

You can also use the `--output` flag to change the format of the response. By default the response comes in a *table* format, but you can choose to switch to *json*:
```
kosli artifact ls creator --output json
```
