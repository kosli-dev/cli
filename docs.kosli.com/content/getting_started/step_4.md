---
title: "Step 4: Create a Kosli pipeline"
bookCollapseSection: false
weight: 260
---

# Step 4: Create a Kosli pipeline

A Kosli *pipeline* stores information about what happens in your build system.
The output of the build system is called an *artifact* in Kosli. An artifact could be, for example,
an application binary, a docker image, documentation, or a file. 

Start by creating a new Kosli pipeline:

```shell {.command}
kosli pipeline declare \
    --pipeline quickstart-nginx \
    --description "Pipeline for quickstart nginx image"
```

You can confirm that the Kosli pipeline was created by running:
```shell {.command}
kosli pipeline ls
```
which should produce the following output:
```plaintext {.light-console}
NAME              DESCRIPTION                          VISIBILITY
quickstart-nginx  Pipeline for quickstart nginx image  private
```
{{< hint info >}}
In the web interface you can select the *Pipelines* option on the left.
It will show you that you have a *quickstart-nginx* pipeline.
If you select the pipeline it will show that no artifacts have
been reported yet.
{{< /hint  >}}