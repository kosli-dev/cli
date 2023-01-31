---
title: "Part 1: Overview"
bookCollapseSection: false
weight: 210
---

# Part 1: Overview

Kosli is a very flexible tool - you can use it to drive a big transformation but you can also implement it without changing your existing process. 

To start using Kosli you need a [kosli account](https://app.kosli.com/sign-up).  
{{< hint success >}}
If you're eager to start using Kosli right away, check our ["Get familiar with Kosli"](/tutorials/get_familiar_with_kosli/) tutorials that allows you to quickly try out Kosli features without the need to spin up a separate environment. No CI required.
{{< /hint >}}

You can start with reporting your **artifacts** from your **pipelines** and get an overview of what you're building and testing. Or you can start with reporting **environments** and get an overview of what's running and where. 


Once both pipelines and environment are in place you get a full picture - Kosli automatically connects the status of your runtime environments and reported artifacts so you can easily see when not qualified or suspicious artifacts made their way to your environments.

What does *"using Kosli"* really mean? It boils down to running [Kosli CLI](/kosli_overview/kosli_tools/#cli) commands:
* whenever an event related to your code or artifact happens in you CI pipeline - eg.: build, code coverage, static code analysis, security scan, etc (whatever your [template](/kosli_overview/what_is_kosli/#template) requires)
* scheduled to monitor environment - e.g.: as a cron job in your environment, or with CI pipeline (depending on the type of the environment you may need to run it in the actual environment or from any machine able to connect to it)

No matter the order you chose for implementing Kosli in you development process, the end result will be the same, so feel free to start with either environments or pipelines. In this overview we'll explain environments first, before moving to pipelines.

## Reporting environments

All environment reporting commands are described in detail in [Part 2: Environments in Kosli](/getting_started/part_2_environments/) section. And you can find a complete syntax in [Kosli Client Reference](/client_reference/).

Before you start reporting you need to [create an environment](/getting_started/part_2_environments/#create-an-environment) in Kosli. You should have a separate Kosli environment for each runtime environment you're reporting.

What does *"reporting environment"* mean? You can learn more about the concept in [Environments](/kosli_overview/what_is_kosli/#environments).

In practice it means running a cli command. Depending on the type of your environment you would run this command:
* **in your CI**, or on any machine able to access the environment: for *ecs*, *lambda*, *s3* and *k8s* environment types
* **on the actual machine** (or vm) that serves as your environment: for *server*, *docker*, *k8s* environment types (use [helm chart](/helm) to install Kosli reporter as a cronjob)

Once your reporting is up and running you'll see the results under "Environments" at [app.kosli.com](https://app.kosli.com)

{{<figure src="/images/env-reporting.png" alt="Environments at app.kosli.com" width="900">}}

## Reporting artifacts

All artifact/evidence reporting commands are described in detail in [Part 3: Pipelines](/getting_started/part_3_pipelines/) and following sections. And you can find a complete syntax in [Kosli Client Reference](/client_reference/).

Before you start reporting you need to [create a pipeline](/getting_started/part_3_pipelines/#create-a-pipeline) in Kosli. Common practice is to have one Kosli pipeline per artifact type. E.g. if your CI pipeline produces one binary you'd report all builds of that binary to ONE Kosli pipeline. If the same CI pipeline was also producing a docker image or any other artifact you'd report it as an artifact to ANOTHER Kosli pipeline. 

Once your Kosli pipelines are ready you can start reporting your [artifacts](/getting_started/part_4_artifacts/) and artifact related events ([evidences](/getting_started/part_5_evidences/), [approvals](/getting_started/part_6_approvals/), [deployments](/getting_started/part_7_deployments/)).

You can report artifacts and events from wherever you want - including your own machine - but the common practice is to report it from CI immediately after it happens (or, in case of [`kosli expect deployment`](/client_reference/kosli_expect_deployment/) command, right before the deployment starts).

[comment]: <> (When reporting artifact and artifact related events, kosli needs to figure out what's the fingerprint of the artifact the report belongs to.)  

[comment]: <> (It can determine that in two ways:)
* if you provide a name or path to the artifact, relative to the location from which you run 

