---
title: 'Environments'
weight: 20
---
# Environments

Environments in Kosli provide a place to track how your systems change over time.

Each runtime environment you'd like to track in Kosli should have its own Kosli environment created - e.g. if you use k8s cluster to host **qa**, **staging** and **production** versions of your product you create 3 separate environments for those in Kosli. 

Kosli supports different type of runtime environments and the reporting command varies for each:
* Kubernetes cluster (k8s)
* Amazon ECS
* Amazon S3
* Amazon Lambda
* Physical/virtual server

You can create Kosli environment using:
* "Add new environment" button under "Environments" section on [app.kosli.com](https://app.kosli.com) 
* cli with **[kosli environment declare](/client_reference/kosli_environment_declare/)** command

Once the Kosli environment is ready you can start reporting the status of your actual runtime environment using one of the **kosli environment report ...** commands - check [client reference](/client_reference) for details

It makes sense to automate reporting - via cronjob or using your CI. It's up to you to decide how often you want the reports to keep coming. Once the cronjob or CI are set to use **kosli environment report ...** command, every time a change in your runtime environment happens a new snapshot capturing current state of the environment will be created. 

![Diagram of Environment Reporting](/images/environments.svg)

The change could be for example:
* a new artifact started running
* an artifact stopped running
* a number of instances of the services has changed
* a compliance status of the artifact has changed

## Snapshots

Snapshot represents a reported status of your runtime environment at a given time. When you click on the name of a specific environment on **Environments** page at [app.kosli.com](https://app.kosli.com) you are taken to the latest snapshot. You can use the arrow buttons to browse older snapshots. 

Once snapshot is reported it can't be modified, that is to secure the integrity of data. Every time the environment report indicates changes in the runtime environment or in the artifact status a new snapshot is created.

## Compliant Environment

Environment is **compliant** when:
1. All the artifacts running in it have provenance and are compliant themselves OR they were [allow-listed](#allow-list)
2. All the artifacts running in it are reported as [deployed](/client_reference/kosli_expect_deployment/) to a given environment

If you're environment is not compliant check the latest snapshot for more detailed info - each unknown or incompliant artifacts will be marked and the reason for the incompliancy will be provided

## Allow list 

Not all the artifacts that run in your environment must be built by you - these may be publicly available artifacts, or artifacts provided by external vendors. In such case you will likely have no information about these artifacts reported to Kosli. 

These artifact will by default be marked with "No provenance" red label and it will affect the compliancy of the whole environment. If you know how and why these artifact are present in your environment you can add them to Allow-list by clicking a button on the snapshot page, or using [kosli environment allowedartifacts add](/client_reference/kosli_environment_allowedartifacts_add/) command




