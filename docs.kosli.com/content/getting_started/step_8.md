---
title: "Step 8: Report what is running in your environment"
bookCollapseSection: false
weight: 291
---

# Step 8: Report what is running in your environment

Report all the docker containers running on your machine to Kosli:
```shell {.command}
kosli environment report docker quickstart
```
You can confirm that this has created an environment snapshot:
```shell {.command}
kosli environment log quickstart
```
```plaintext {.light-console}
SNAPSHOT  FROM                           TO   DURATION
1         Tue, 01 Nov 2022 15:55:49 CET  now  11 seconds
```

You can get a detailed view of all the docker containers included in the snapshot report:
```shell {.command}
kosli environment get quickstart
```

```plaintext {.light-console}
COMMIT  ARTIFACT                                                                       PIPELINE  RUNNING_SINCE  REPLICAS
N/A     Name: nginx:1.21                                                               N/A       3 minutes ago  1
        Fingerprint: 8f05d73835934b8220e1abd2f157ea4e2260b9c26f6f63a8e3975e7affa46724
```

The `kosli environment report docker` command reports *all* the 
docker containers running in your environment, equivalent to the output from 
`docker ps`. This tutorial only shows the `nginx` container 
in the examples.

{{< hint info >}}
If you refresh the *Environments* web page in your Kosli account, you will see 
that there is now a timestamp for *Last Change At* column. 
Select the *quickstart* link on left for a detailed view of what is currently running.
{{< /hint >}}
