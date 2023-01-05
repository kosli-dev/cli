---
title: Overview
bookCollapseSection: false
weight: 10
---

# Overview

## What is an organization

Organization in Kosli owns Kosli pipelines and environments - which means only members of the organization can get access to them.
By default, when you sign up to Kosli, a personal organization is created for you and the name of the organization matches your user name. Only you can access your personal organization.

You can also create shared organizations, and invite other Kosli users as members, so they can see and report to your Kosli pipelines and environments.

### Shared organization

To create a shared organization click on your profile picture (or avatar) in the top right corner of [app.kosli.com](https://app.kosli.com) and select "Add an organization". You'd be asked to provide the name and the description of your organization. After you click "Create Organization" button the new organization is ready. After the page reloads you'll see the "Settings" page for the new organization. 

You can switch between organizations using dropdown menu in the top left corner of the page, under Kosli logo. 

### Shared organization members 

To add users to your shared organization make sure you have the right organization selected from the dropdown menu and click "Settings". Here you can add users: click on "Add member" button, provide a github username of the user you'd like to share organization with, and select desired role:
* member can create Kosli pipelines and environments, report to and read from them
* admin can do the same things member can plus they can also add and remove users from the organization 

## What are the environments

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

### Snapshots

Snapshot represents a reported status of your runtime environment at a given time. When you click on the name of a specific environment on **Environments** page at [app.kosli.com](https://app.kosli.com) you are taken to the latest snapshot. You can use the arrow buttons to browse older snapshots. 

Once snapshot is reported it can't be modified, that is to secure the integrity of data. Every time the environment report indicates changes in the runtime environment or in the artifact status a new snapshot is created.

### Compliant Environment

Environment is **compliant** when:
1. All the artifacts running in it have provenance and are compliant themselves OR they were [allow-listed](#allow-list)
2. All the artifacts running in it are reported as [deployed](/client_reference/kosli_expect_deployment/) to a given environment

If you're environment is not compliant check the latest snapshot for more detailed info - each unknown or incompliant artifacts will be marked and the reason for the incompliancy will be provided

### Allow list 

Not all the artifacts that run in your environment must be built by you - these may be publicly available artifacts, or artifacts provided by external vendors. In such case you will likely have no information about these artifacts reported to Kosli. 

These artifact will by default be marked with "No provenance" red label and it will affect the compliancy of the whole environment. If you know how and why these artifact are present in your environment you can add them to Allow-list by clicking a button on the snapshot page, or using [kosli environment allowedartifacts add](/client_reference/kosli_environment_allowedartifacts_add/) command

## What are the pipelines

Pipelines in Kosli provide a place to report and track artifact status and related events from your CI pipelines.

You can create Kosli pipeline using our cli with **[kosli pipeline declare](/client_reference/kosli_pipeline_declare/)** command. 

It's normal practice to add your pipeline declaring command to your build pipeline. It's perfectly fine to run it every time you run a build. You can also change your template over time, for example by adding new control. It won't affect the compliancy of artifacts reported before the change of the template.

Once your Kosli pipeline is in place you can start reporting artifacts and evidences of all the events you want to report (matching declared template) from your CI pipelines. Kosli cli provides a variety of commands to make it possible: 

![Diagram of Pipeline Reporting](/images/pipelines.svg)

A number of required flags may be defaulted to a set of environment variables, depending on the CI system you use. Check [How to use Kosli in CI Systems](/getting_started/use_kosli_in_ci_systems/) for more details. All of the flags can be represented by [environment variables](/introducing_kosli/cli/#environment-variables)

### Artifacts

Whatever you produce during your build process can be an artifact - a binary file, an archive, a folder, a docker image... sometimes you don't produce anything new while "building" and the complete code can be your artifact. 

Best practice is to create Kosli pipeline for each type of artifact - e.g. if your CI pipeline produces 3 separate artifacts (that could be 3 different binaries for three different platforms) you'd create 3 different Kosli pipelines to report artifacts and evidences. 

### Evidences

When declaring a pipeline you need to provide a template - a list of required controls (evidences) you required for your artifact in order for the artifact to become compliant. That could be for example:
* existing pull request
* code coverage report
* integration test
* unit test 
* and more...

Whenever an event related to an evidence happens - e.g. test are finished - use Kosli CLI to report the evidence to Kosli. 

### Deployments

No matter if you deploy your artifacts from your build pipeline, or do you have a separate one for that purpose, you should report to Kosli that you expect an artifact to start running in an environment. You do that using [kosli expect deployment](/client_reference/kosli_expect_deployment/) command. Environment that you're deploying to has to be specified, so if you deploy to more than one environment you need to report each deployment separately

## What are the fingerprints 

Every time artifact is reported to Kosli a SHA256 digest of it is calculated. It doesn't matter if the artifact is a single file, a directory or a docker image - we can always calculate SHA256. 

Fingerprints are used to connect the information recorded in Kosli - about environments, deployments and approval - to matching artifact. 

You can also use Kosli CLI to calculated the fingerprint of any artifact locally. See [kosli fingerprint](/client_reference/kosli_fingerprint/) for more details.