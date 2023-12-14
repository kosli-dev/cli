---
title: 'What is Kosli?'
weight: 120
---
# What is Kosli?

Kosli is a connected collection of attested evidence about any business process.   
Business processes instrumented with Kosli are Always-Audit-Ready.

<!--
{{<figure src="/images/kosli-overview-docs.jpg" alt="Kosli overview" width="1000">}}
Below you can read about what elements Kosli consists of.
-->

## Overview: Organization

A Kosli *Organization* "owns" Environments and Flows - only members of each organization 
can access the Environments and Flows belonging to that organization.
By default, when you sign up to Kosli, a personal organization is created for you.
Its name is your username. Only you can access your personal organization.

## Overview: Environments and Snapshots 

A Kosli *Environment* stores snapshots containing information about
the software Artifacts you are running in your runtime environment.
Kosli supports many kinds of runtime environments; (server, Kubernetes cluster, AWS, etc.)  

## Overview: Flows and Trails

A Kosli *Flow* represents a single business-process, about which you want to attest (record) evidence.   
A Kosli *Trail* is a single instance/run of a Kosli *Flow*.
Here are some examples:

- Required steps when onboarding a new employee
  - The name of the Flow could be *employee-onboarding*
  - Each employee has their own Trail, eg *PeterPan*
  - Attested evidence might relate to documents they have read, eg *DisasterRecoveryPlan*
- How you build, test, and deploy a software Artifact
  - The Flow could be named after the Artifact's git-repository, eg *NeverLand*
  - Each git-commit could have its own Trail, eg *d65da7464db715e0513c5191a90cb546b9d4696b*
  - Each Jira-ticket could have its own Trail, eg *JMB-452*
  - Attested evidence would relate to required SDLC tasks, eg *snyk-scan* 
- Defining and applying changes to a runtime infrastructure
- Recording real-time access to production servers
- Feature-flag changes


## Environments

Environments in Kosli provide a place to track how your software runtime systems change over time.

{{<figure src="/images/envs.png" alt="Environments" width="900">}}

Each runtime environment you'd like to track in Kosli should have its own Kosli environment created. 
Kosli allows you to define the borders of your environments. For example, if you use a Kubernetes 
cluster, you can either treat it as one Kosli environment or treat one or more namespaces in the cluster 
as one Kosli environment.

Kosli supports different types of runtime environments and the reporting command varies for each:
* Kubernetes cluster (k8s)
* Amazon ECS
* Amazon S3
* Amazon Lambda
* Physical/virtual server
* Docker host

You can create a Kosli environment using:
* The CLI's **[kosli create environment](/client_reference/kosli_create_environment/)** command
* The "Add new environment" button under the "Environments" section on [app.kosli.com](https://app.kosli.com) that will take you to environment creation form:

{{<figure src="/images/new-env-form.png" alt="Add environment form" width="900">}}

Once the Kosli environment is ready you can start reporting the status of your actual runtime environment using one of the **kosli snapshot ...** commands - check [CLI Reference](/client_reference) for details.

Reporting your environments should be automated via a cron-like schedule. 
It's up to you to decide how often you want the reports to keep coming, but we recommend 
high frequency to be able to avoid missing short-lived changes. 
Every time a change in your runtime environment is reported, a new snapshot capturing the 
current state of the environment will be created. 

![Diagram of Environment Reporting](/images/environments-cli-v2.svg)

The change could be for example:
* a new artifact started running
* an artifact stopped running
* an artifact was restarted
* the number of instances of a service has changed
* the compliance status of an artifact has changed

### Snapshots

A Snapshot represents the reported status of your runtime environment at a given point in time.

{{<figure src="/images/snapshot-467.png" alt="Snapshot 467" width="900">}}

Using the Kosli UI, you can use the arrow buttons on the right hand side above the running artifacts 
list to browse older snapshots. 

Snapshots are append-only immutable objects. That is once a snapshot is created, it can't be modified.

### Compliant Environment

An Environment is **compliant** when the following conditions are met:
1. All the artifacts running in the environment have provenance (were reported in a Kosli Flow) and are compliant themselves OR they were [allow-listed](#allow-list);
2. All the artifacts running in the environment are [expected to be deployed](/client_reference/kosli_expect_deployment/) to that environment.

You will see the status of your environment on the environments list.
You will also see the compliance status of each snapshot when you browse the snapshots of the environment.

If your environment is not compliant, check the latest snapshot for more detailed info -
each unknown or non-compliant artifact will be marked and the reason for the non-compliance will be provided.

{{<figure src="/images/non-compliant-env.png" alt="Snapshot 467" width="1000">}}


### Allow list 

Not all artifacts that run in your environment must be built by you - these may be publicly 
available artifacts, or artifacts provided by external vendors. In such cases, it is likely that 
those artifacts won't have provenance in Kosli. 

These artifacts will -by default- be marked with "No provenance" red label and it will 
affect the compliance of the whole environment. If you know how and why these artifacts 
are present in your environment you can add them to the Allow-list by clicking a button 
on the snapshot page, or using [kosli allow artifact](/client_reference/kosli_allow_artifact/) command

## Flows and Artifacts

### Flows

Flows in Kosli allow you to track how your value streams produce artifacts. 
They provide a place to report artifact creation events as well as any evidence 
produced from your CI pipelines.

You can create Kosli flow using the **[kosli create flow](/client_reference/kosli_create_flow/)** command. 

When you create a flow, you specify a [template](/kosli_overview/what_is_kosli/#template). The creation of a flow can happen in or out of your CI pipelines.
The template declares what pieces of evidence are required for an artifact produced from that flow to be compliant.

Best practice is to create Kosli flow for each of your value streams regardless of how your CI pipelines are setup. 
For example, if your CI pipeline produces 3 separate artifacts, you'd create 3 different 
Kosli flows to report artifacts and evidence.

Once your Kosli flow is in place you can start reporting artifacts and evidence of all the 
events you want to report (matching declared template) from your CI pipelines. Kosli CLI 
provides a variety of commands to make it possible: 

![Diagram of Flow Reporting](/images/flows-cli-v2.svg)

A number of required flags may be defaulted to a set of environment variables, depending 
on the CI system you use. Check [How to use Kosli in CI Systems](/integrations/ci_cd/) for more details. 
All flags can be represented by [environment variables](/kosli_overview/kosli_tools/#environment-variables).

### Template

When creating a Kosli flow you need to provide a template - a list of expected evidence you 
require for your artifact in order for the artifact to become compliant.
That could be, for example:
* existing pull request
* code coverage report
* integration test
* unit test 
* and more...

The template can be changed over time and it won't affect the compliance of artifacts reported before the change happens.

### Artifacts

{{<figure src="/images/artifact-view.png" alt="Artifact view" width="1000">}}

Whatever you produce during your build process can be an artifact - a binary file, an archive, a folder, 
a docker image... sometimes you don't produce anything new while "building" and the folder containing 
your source code can be the artifact.

An artifact is identified by its [fingerprint](/kosli_overview/what_is_kosli/#fingerprints) which is used to link the artifacts running 
in an environment back to their provenance in a flow.

### Evidence

You can report to Kosli pieces of evidence related to either an artifact or a git commit. 
This gives you flexibility to report evidence before or after you report an artifact.
The evidence names have to match the names you declare in your flow [template](/kosli_overview/what_is_kosli/#template) and 
are then used to evaluate whether an artifact is compliant or not.

Evidence that is reported to a git commit is automatically linked to any artifact produced from that commit.

Kosli supports some types of evidence that you can report, 
but you can also report absolutely anything as an evidence using the [generic evidence type](/client_reference/kosli_report_evidence_artifact_generic/).

The supported types of evidence are:
* [bitbucket](/client_reference/kosli_report_evidence_artifact_pullrequest_bitbucket/), [github](/client_reference/kosli_report_evidence_artifact_pullrequest_github/) and [gitlab](/client_reference/kosli_report_evidence_artifact_pullrequest_gitlab/) pull request - verify and report if a pull request exists for a commit used to build your artifact
* [junit](/client_reference/kosli_report_evidence_artifact_junit/) - report the result of your unit test (requires results as XML in JUnit format)
* [snyk](/client_reference/kosli_report_evidence_artifact_snyk/) - report Snyk vulnerability scan 

For the built-in evidence types, Kosli determines the compliance of the evidence by analyzing the data you provide. 
For generic evidence, however, you need to do the required analysis and inform Kosli whether the evidence is compliant or not. 

### History

{{<figure src="/images/artifact-history.png" alt="Artifact evidence" width="600">}}

At the bottom of an artifact page you can see the artifact timeline: when it was created, when evidence, approvals and deployments were reported, and when the artifact was reported running in each environment.

When you report an event related to a specific environment (expected deployment or environment report) the timeline branches out, for each environment. From now on events related to environments will have different colors - it makes it easier to follow artifacts history in each environment.

### Compliant artifact

Each artifact you report to Kosli will be displayed as being in one of three states in a Kosli flow: **compliant**, **non-compliant** or **incomplete**.
The state depends on how the evidence received for the artifact matches with the template used for the flow which the artifact belongs to. 

#### Compliant

When your artifact was reported to Kosli together with **all** the required (as defined in the template) evidence reported as a ***compliant*** evidence, it will be displayed in your Kosli flow as a **Compliant** artifact:

{{<figure src="/images/artifact-compliant.png" alt="Environment, Snapshot #1" width="900">}}

#### Non-Compliant

When your artifact was reported to Kosli together with **all** the required (as defined in the template) evidence, with **at least one** of the evidence reported as ***non-compliant***, it will be displayed in your Kosli flow as a **Non-compliant** artifact:

{{<figure src="/images/artifact-non-compliant.png" alt="Environment, Snapshot #1" width="900">}}

#### Incomplete

When your artifact was reported to kosli but **not all** the required (as defined in the template) evidence were reported yet, it will be displayed in your Kosli flow as an **Incomplete** artifact:

{{<figure src="/images/artifact-incomplete.png" alt="Environment, Snapshot #1" width="900">}}


### Deployments

No matter from where and how you deploy your artifacts, you should report to Kosli that you expect an artifact to start running in an environment. You do that using [kosli expect deployment](/client_reference/kosli_expect_deployment/) command. The Environment you're deploying to has to be specified, so if you deploy to more than one environment you need to report each deployment separately.

Reporting the expected deployments in an environment ensures that what runs in your environment is what you expect.

## Fingerprints 

Fingerprint is a unique immutable identifier of an artifact. It is a calculated SHA256 hash of an artifact. It doesn't matter if the artifact is a single file, a directory or a docker image - we can always calculate its SHA256.

Fingerprint is used to connect the information recorded in Kosli - about environments, deployments and approvals - to a matching artifact in a flow. 

You can also use the Kosli CLI to calculate the fingerprint of any artifact locally. See [kosli fingerprint](/client_reference/kosli_fingerprint/) for more details.
