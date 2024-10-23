---
title: "Part 8: Environments"
bookCollapseSection: false
weight: 280
---
# Part 8: Environments

Kosli environments allow you to record the artifacts running in your runtime environments and how they change. Every time an environment change (or a set of changes) is reported, Kosli creates a new environment snapshot containing the status of the environment at a given point in time.

## Create an environment

You can create Kosli environments in the app, via CLI or via the API. When you create an environment, you give it a name, a description and select its type. 

{{< hint warning >}}
Make sure that type of Kosli environment matches the type of the environment you'll be reporting from.
{{< /hint >}}

### Via CLI

To create an environment via CLI, you would run a command like this:

```shell {.command}
$ kosli create environment quickstart \
    --environment-type docker \
    --description "quickstart environment for tutorial"
```

See [kosli create environment](/client_reference/kosli_create_environment/) for CLI usage details and examples.

### Via UI

You can also create an environment directly from [app.kosli.com](https://app.kosli.com).

- Make sure you've selected the organization you want to use from the orgs dropdown in the top left corner.
- Click on `Environments` in the left navigation menu.
- Click the `Add new environment` button
- Fill in the environment name and description and select a type, then click `Save Environment`.


After the new environment is created you'll be redirected to its page, which will initially have no snapshots. Once you start reporting your actual runtime environment to Kosli you'll be able to find snapshots and events (such as which artifacts started or stopped running) listed on that page.

## Snapshoting an environment

To record the current status of your environment you need to make Kosli CLI snapshot the running artifacts in it and report it to Kosli. 
When Kosli receives an environment report, if the received list of running artifacts is different than what is in the latest environment snapshot, a new environment snapshot is created. Snapshots are immutable and can't be tampered with.

Currently, the following environment types are supported:
- Kubernetes
- Docker
- Paths on a server
- AWS Simple Storage Service (S3)
- AWS Lambda
- AWS Elastic Container Service (ECS)

You can report environment snapshots manually using the `kosli snapshot [...]` commands for testing. For production use, however,  you would configure the reporting to happen automatically on regular intervals, e.g. via a cron job or scheduled CI job, or on certain events. 

You can follow one of the tutorials below to setup automatic snapshot reporting for your environment:
- [Kubernetes environment reporting](/tutorials/report_k8s_envs)
- [AWS ECS/S3/Lambda environment reporting](/tutorials/report_aws_envs)

### Snapshotting scopes

Depending on the type of your environment, you can scope what to snapshot from the environment. The following table shows the different scoping options currently available for different environment types:

| what to snapshot ->        | all resources | resources by names | resources by Regex | exclude by names | exclude by Regex |
|----------------------------|---------------|--------------------|--------------------|------------------|------------------|
| ECS (clusters)             |       √       |          √         |          √         |         √        |         √        |
| Lambda (functions)         |       √       |          √         |          √         |         √        |         √        |
| S3 (buckets)               |               |                    |                    |                  |                  |
| docker (containers)        |       √       |                    |                    |                  |                  |
| k8s (namespaces)           |       √       |          √         |          √         |         √        |         √        |
| azure (functions and apps) |       √       |                    |                    |                  |                  |


## Logical Environments

Logical environments are a way to group your Kosli environments so you can view all changes happening in your group in the same place. For example, if what you consider to be “Production” is a combination of a Kubernetes cluster, an S3 bucket, and a configuration file, you can combine the reports sent to these Kosli environments into a “Production” logical environment.

A logical environment can be created in the app or the CLI, and physical environments can be assigned to it in the app or with the [`kosli join environment`](/client_reference/kosli_join_environment/) command.
