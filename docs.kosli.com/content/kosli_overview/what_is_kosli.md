---
title: 'What is Kosli?'
weight: 120
---
# What is Kosli?

Kosli records data from your CI pipelines and runtime environments, allowing you to query life after git from the command line.

{{<figure src="/images/kosli-overview-docs.jpg" alt="Kosli overview" width="1000">}}

Below you can read about what elements Kosli consists of.

## Organization

An Organization in Kosli "owns" Kosli flows and environments - which means only members of each organization can get access to environments and pipelines that belong to the organization.
By default, when you sign up to Kosli, a personal organization is created for you and the name of the organization matches your user name. Only you can access your personal organization.

### Shared organization

To collaborate with more people (a team or a whole company) you can create shared organizations, and invite other Kosli users as members, so they can see and report to your Kosli flows and environments.

To create a shared organization click on your profile picture (or avatar) in the top right corner of [app.kosli.com](https://app.kosli.com) and select "Add an organization". 

{{<figure src="/images/add-org.png" alt="Add an organization" width="250">}}


You'd be asked to provide the name and the description of your organization. After you click "Create Organization" button the new organization is ready. 

{{<figure src="/images/add-org-form.png" alt="New organization form" width="900">}}

After the page reloads you'll see the "Settings" page for the new organization. 
You can switch between organizations using dropdown menu in the top left corner of the page, under Kosli logo. 

{{<figure src="/images/select-org.png" alt="org page" width="900">}}


### Shared organization members 

To add users to your shared organization make sure you have the right organization selected from the dropdown menu and click "Settings".  
Here you can add users: click on "Add member" button, provide a github username of the user you'd like to share organization with, and select desired role:
* member can create Kosli flows and environments, report to and read from them
* admin can do the same things member can plus they can also add and remove users from the organization 

## Environments

Environments in Kosli provide a place to track how your systems change over time.

{{<figure src="/images/envs.png" alt="Environments" width="900">}}

Each runtime environment you'd like to track in Kosli should have its own Kosli environment created - e.g. if you use k8s cluster to host **qa**, **staging** and **production** versions of your product you create 3 separate environments for those in Kosli. 

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

Once the Kosli environment is ready you can start reporting the status of your actual runtime environment using one of the **kosli snapshot ...** commands - check [client reference](/client_reference) for details

It makes sense to automate reporting - via a cronjob or using your CI. It's up to you to decide how often you want the reports to keep coming. Once the cronjob or CI are set to use the **kosli snapshot ...** command, every time a change in your runtime environment happens a new snapshot capturing the current state of the environment will be created. 

![Diagram of Environment Reporting](/images/environments.svg)

The change could be for example:
* a new artifact started running
* an artifact stopped running
* an artifact was restarted
* a number of instances of the services has changed
* a compliance status of the artifact has changed

### Snapshots

A Snapshot represents a reported status of your runtime environment at a given time. When you click on the name of a specific environment on the **Environments** page at [app.kosli.com](https://app.kosli.com) you are taken to the latest snapshot. 

{{<figure src="/images/snapshot-467.png" alt="Snapshot 467" width="900">}}

You can use the arrow buttons on the right hand side above the running artifacts list to browse older snapshots. 

Once snapshot is reported it can't be modified, that is to secure the integrity of data. Every time the environment report indicates changes in the runtime environment or in the artifact status, a new snapshot is created.

### Compliant Environment

An Environment is **compliant** when two following conditions are met:
1. All the artifacts running in it have provenance (are reported to Kosli) and are compliant themselves OR they were [allow-listed](#allow-list);
2. All the artifacts running in it are [expected to be deployed](/client_reference/kosli_expect_deployment/) to a given environment.

You will see the status of your environment on the environments list. You will also see the status of compliancy of each snapshot when you browse the history of the environment.

If your environment is not compliant check the latest snapshot for more detailed info - each unknown or non-compliant artifact will be marked and the reason for the non-compliance will be provided.

{{<figure src="/images/non-compliant-env.png" alt="Snapshot 467" width="1000">}}


### Allow list 

Not all artifacts that run in your environment must be built by you - these may be publicly available artifacts, or artifacts provided by external vendors. In such case you will likely have no information about these artifacts reported to Kosli. 

These artifacts will by default be marked with "No provenance" red label and it will affect the compliance of the whole environment. If you know how and why these artifacts are present in your environment you can add them to the Allow-list by clicking a button on the snapshot page, or using [kosli allow artifact](/client_reference/kosli_allow_artifact/) command

## Flows and Artifacts

### Flows

Flows in Kosli provide a place to report and track artifacts status and related events from your CI pipelines.

You can create Kosli flow using our CLI with **[kosli create flow](/client_reference/kosli_create_flow/)** command. 

You can run the CLI command manually e.g. using your own computer, but it's also ok to add your flow creation command to your CI pipeline. It's perfectly fine to run it every time you run a CI pipeline. You can also change your [template](/kosli_overview/what_is_kosli/#template) over time, for example by adding new controls. It won't affect the compliance of artifacts reported before the change of the template.

Once your Kosli flow is in place you can start reporting artifacts and evidence of all the events you want to report (matching declared template) from your CI pipelines. Kosli CLI provides a variety of commands to make it possible: 

![Diagram of Flow Reporting](/images/pipelines.svg)

A number of required flags may be defaulted to a set of environment variables, depending on the CI system you use. Check [How to use Kosli in CI Systems](/integrations/ci_cd/) for more details. All flags can be represented by [environment variables](/kosli_overview/kosli_tools/#environment-variables).

### Artifacts

{{<figure src="/images/artifact-view.png" alt="Artifact view" width="1000">}}

Whatever you produce during your build process can be an artifact - a binary file, an archive, a folder, a docker image... sometimes you don't produce anything new while "building" and the folder containing your source code can be the artifact.

Best practice is to create Kosli flow for each type of artifact - e.g. if your CI pipeline produces 3 separate artifacts (that could be 3 different binaries for three different platforms) you'd create 3 different Kosli flows to report artifacts and evidence.

### Template

{{<figure src="/images/artifact-evidence.png" alt="Artifact evidence" width="600">}}


When creating a Kosli flow you need to provide a template - a list of expected controls (evidence) you require for your artifact in order for the artifact to become compliant. That could be for example:
* existing pull request
* code coverage report
* integration test
* unit test 
* and more...

Whenever an event related to your artifact happens, and you want to report it as evidence, you need to tell Kosli which artifact the evidence refers to. You can do it in two ways:

1. You can use `--artifact-type` flag and provide an artifact as an argument to evidence reporting commands (given artifact needs to be available from the location the command is run, so it can be used to calculate artifacts [fingerprint](/kosli_overview/what_is_kosli/#fingerprints));
2. You can use `--fingerprint` (or `--sha256` for older versions of Kosli CLI) to provide a previously calculated fingerprint of an artifact.

You can report absolutely anything as evidence. If there is no support for your specific type of evidence, you can use [generic evidence type](/client_reference/kosli_report_evidence_artifact_generic/).

Evidence is reported as compliant if Kosli determines it as compliant (e.g. analyzing JUnit or Snyk test results). For generic evidence you can implement your own mechanism to determine compliance status and use `--compliant=false` in your evidence reporting command, if you want to send the evidence as non-compliant. 

There are a number of types of evidence with dedicated support:
* [bitbucket](/client_reference/kosli_report_evidence_artifact_pullrequest_bitbucket/), [github](/client_reference/kosli_report_evidence_artifact_pullrequest_github/) and [gitlab](/client_reference/kosli_report_evidence_artifact_pullrequest_gitlab/) pull request - verify and report if a pull request exists for a commit used to build your artifact
* [junit](/client_reference/kosli_report_evidence_artifact_junit/) - report the result of your unit test (requires results as XML in JUnit format)
* [snyk](/client_reference/kosli_report_evidence_artifact_snyk/) - report Snyk vulnerability scan 

### History

{{<figure src="/images/artifact-history.png" alt="Artifact evidence" width="600">}}

At the bottom of an artifact page you can see the artifact timeline: when it was created, when evidence, approvals and deployments were reported, and when the artifact was reported running in each environment.

When you report an event related to a specific environment (expected deployment or environment report) the timeline branches out, for each environment. From now on events related to environments will have different colors - it makes it easier to follow artifacts history in each environment.

### Compliant artifact

Each artifact you report to Kosli will be displayed as being in one of three states in your Kosli pipeline: compliant, non-compliant or incomplete. That status is not reserved for software development in regulated industries. It tells you how far in the process your artifact got and if there are any troubles detected.

#### Compliant

When your artifact was reported to Kosli together with **all** the required (as defined in the template) evidence reported as ***compliant***, it will be displayed in your Kosli flow as **Compliant** artifact:

{{<figure src="/images/artifact-compliant.png" alt="Environment, Snapshot #1" width="900">}}

#### Non-Compliant

When your artifact was reported to Kosli together with **all** the required (as defined in the template) evidence, with **at least one** of the evidence reported as ***non-compliant***, it will be displayed in your Kosli flow as **Non-compliant** artifact:

{{<figure src="/images/artifact-non-compliant.png" alt="Environment, Snapshot #1" width="900">}}

#### Incomplete

When your artifact was reported to kosli but **not all** the required (as defined in the template) evidence were reported yet, it will be displayed in your Kosli flow as **Incomplete** artifact:

{{<figure src="/images/artifact-incomplete.png" alt="Environment, Snapshot #1" width="900">}}


### Deployments

No matter from where and how you deploy your artifacts, you should report to Kosli that you expect an artifact to start running in an environment. You do that using [kosli expect deployment](/client_reference/kosli_expect_deployment/) command. The Environment you're deploying to has to be specified, so if you deploy to more than one environment you need to report each deployment separately.

## Fingerprints 

Fingerprint is a unique identifier of an artifact. It is a calculated SHA256 hash of an artifact. It doesn't matter if the artifact is a single file, a directory or a docker image - we can always calculate its SHA256.

Fingerprint is used to connect the information recorded in Kosli - about environments, deployments and approval - to a matching artifact. 

You can also use the Kosli CLI to calculate the fingerprint of any artifact locally. See [kosli fingerprint](/client_reference/kosli_fingerprint/) for more details.
