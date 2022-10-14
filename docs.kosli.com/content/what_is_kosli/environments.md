---
title: 'Environments'
weight: 20
---

## Environments

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

<!-- 
TODO:

## Snapshots

## Compliant Environment

## Allow list 
-->



