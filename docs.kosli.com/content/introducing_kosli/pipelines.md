---
title: 'Pipelines'
weight: 30
---
# Pipelines

Pipelines in Kosli provide a place to report and track artifact status and related events from your CI pipelines.

You can create Kosli pipeline using our cli with **[kosli pipeline declare](/client_reference/kosli_pipeline_declare/)** command. 

It's normal practice to add your pipeline declaring command to your build pipeline. It's perfectly fine to run it every time you run a build. You can also change your template over time, for example by adding new control. It won't affect the compliancy of artifacts reported before the change of the template.

Once your Kosli pipeline is in place you can start reporting artifacts and evidences of all the events you want to report (matching declared template) from your CI pipelines. Kosli cli provides a variety of commands to make it possible: 

![Diagram of Pipeline Reporting](/images/pipelines.svg)

A number of required flags may be defaulted to a set of environment variables, depending on the CI system you use. Check [How to use Kosli in CI Systems](/getting_started/use_kosli_in_ci_systems/) for more details. All of the flags can be represented by [environment variables](/introducing_kosli/cli/#environment-variables)

## Artifacts

Whatever you produce during your build process can be an artifact - a binary file, an archive, a folder, a docker image... sometimes you don't produce anything new while "building" and the complete code can be your artifact. 

Best practice is to create Kosli pipeline for each type of artifact - e.g. if your CI pipeline produces 3 separate artifacts (that could be 3 different binaries for three different platforms) you'd create 3 different Kosli pipelines to report artifacts and evidences. 

## Evidences

When declaring a pipeline you need to provide a template - a list of required controls (evidences) you required for your artifact in order for the artifact to become compliant. That could be for example:
* existing pull request
* code coverage report
* integration test
* unit test 
* and more...

Whenever an event related to an evidence happens - e.g. test are finished - use Kosli CLI to report the evidence to Kosli. 

## Deployments

No matter if you deploy your artifacts from your build pipeline, or do you have a separate one for that purpose, you should report to Kosli that you expect an artifact to start running in an environment. You do that using [kosli expect deployment](/client_reference/kosli_expect_deployment/) command. Environment that you're deploying to has to be specified, so if you deploy to more than one environment you need to report each deployment separately

<!-- 
TODO: 



## Approvals

## Releases 
-->

