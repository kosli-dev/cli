---
title: "Part 3: Pipelines"
bookCollapseSection: false
weight: 230
---
# Part 3: Pipelines

Kosli allows you to connect the development world (commits, builds, tests, approvals, deployments) with what’s happening in operations. There is a variety of commands that let you report all the necessary information to Kosli and - relying on automatically calculated fingerprints of your artifacts - match it with the environments.

{{< hint warning >}}
In all the commands below we skip required `--api-token` and `--owner` flags - these can be easily configured via [config file](/kosli_overview/kosli_tools/#config-file) or [environment variables](/kosli_overview/kosli_tools/#environment-variables) so you don't have type them over and over again.
{{< /hint >}}

## Create a pipeline

To report artifacts to Kosli you need to create a Kosli [pipeline](/kosli_overview/what_is_kosli/#pipelines) first. When you create a pipeline you also define a [template](/kosli_overview/what_is_kosli/#template) - a list of types of evidence (controls) you need to be reported in order for the artifact to become compliant. Use the `--template` flag to provide the list of controls. 

Later, when reporting evidence for a specific control you will use the same name you used in the template to identify what evidence you are reporting.

It is a normal practice to include `kosli pipeline declare` command in the same CI pipeline you use to build the artifact you want to report to that Kosli pipeline. None of the previously reported artifacts will be overwritten or lost. And if you change the template, by adding or removing required evidence, it won't affect the compliancy status of existing artifacts.

### Example

```
$ kosli pipeline declare \
	--pipeline project-a \
	--description "Project A artifacts" \
	--template artifact,unit-test,pull-request,snyk,code-coverage

pipeline 'project-a' created
```
See [kosli pipeline declare](/client_reference/kosli_pipeline_declare/) for more details. 
