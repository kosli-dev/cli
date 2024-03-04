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


After the new environment is created you'll be redirected to its page - with "No events have been found for [...]" message. Once you start reporting your actual runtime environment to Kosli you'll see all the events (such as which artifacts started or stopped running) listed on that page.

## Snapshoting an environment

There is range of `kosli snapshot [...]` commands, allowing you to report a variety of environments. To record the current status of your environment you simply run one of them. While you can do it manually, typically you would run the commands automatically, e.g. via a cron job or scheduled CI job.

Whenever an environment report is received, if the received list of running artifacts is different than what is in the latest snapshot, a new snapshot is created. Snapshots are immutable and can't be tampered with.

Currently, the following environment types are supported:

- Kubernetes: see [kosli snapshot kubernetes](/client_reference/kosli_snapshot_k8s/) for usage details and examples.
- Docker: see [kosli snapshot docker](/client_reference/kosli_snapshot_docker/) for usage details and examples.
- Physical/Virtual Server: see [kosli snapshot server](/client_reference/kosli_snapshot_server/) for usage details and examples.
- AWS Simple Storage Service (S3): see [kosli snapshot s3](/client_reference/kosli_snapshot_s3/) for usage details and examples.
- AWS Lambda: see [kosli snapshot lambda](/client_reference/kosli_snapshot_lambda/) for usage details and examples.
- AWS Elastic Container Service (ECS): see [kosli snapshot ecs](/client_reference/kosli_snapshot_ecs/) for usage details and examples.
