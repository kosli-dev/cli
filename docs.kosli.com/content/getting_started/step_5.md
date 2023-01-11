---
title: "Step 5: Create a Kosli environment"
bookCollapseSection: false
weight: 270
---

# Step 5: Create a Kosli environment

A Kosli *environment* stores snapshots containing information about
the software artifacts that you are running in your runtime environments.

Create a Kosli environment:

```shell {.command}
kosli environment declare \
    --name quickstart \
    --environment-type docker \
    --description "quickstart environment for tutorial"
```

You can verify that the Kosli environment was created:

```shell {.command}
kosli environment ls
```

```plaintext {.light-console}
NAME        TYPE    LAST REPORT  LAST MODIFIED
quickstart  docker               2022-11-01T15:30:56+01:00
```

{{< hint info >}}
If you refresh the *Environments* web page in your Kosli account, 
it will show you that you have a *quickstart* environment and that
no reports have been received.
{{< /hint >}}