---
title: "Part 4: Flows"
bookCollapseSection: false
weight: 230
---
# Part 4: Flows

Kosli allows you to connect the development world (commits, builds, tests, approvals, deployments) with what’s happening in operations. There is a variety of commands that let you report all the necessary information to Kosli and - relying on automatically calculated fingerprints of your artifacts - match it with the environments.

{{< hint warning >}}
In all the commands below we skip required `--api-token` and `--org` flags - these can be easily configured via [config file](/kosli_overview/kosli_tools/#config-file) or [environment variables](/kosli_overview/kosli_tools/#environment-variables) so you don't have type them over and over again.
{{< /hint >}}

## Create a flow

To report artifacts to Kosli you need to create a Kosli [flow](/kosli_overview/what_is_kosli/#flows) first. When you create a flow you also define a [template](/kosli_overview/what_is_kosli/#template) - a list of types of evidence (controls) you need to be reported in order for the artifact to become compliant. Use the `--template` flag to provide the list of controls. 

Later, when reporting evidence for a specific control you will use the same name you used in the template to identify which evidence you are reporting.

It is a normal practice to include `kosli create flow` command in the same CI pipeline you use to build the artifact you want to report to that Kosli flow. None of the previously reported artifacts will be overwritten or lost. And if you change the template, by adding or removing required evidence, it won't affect the compliance status of existing artifacts.

### Example

{{< tabs "commands" "col-no-wrap" >}}

{{< tab "v2" >}}
```
$ kosli create flow project-a \
	--description "Project A artifacts" \
	--template artifact,unit-test,pull-request,snyk,code-coverage

flow 'project-a' was created
```
{{< /tab >}}

{{< tab "v0.1.x" >}}
```
$ kosli pipeline declare \
	--pipeline project-a \
	--description "Project A artifacts" \
	--template artifact,unit-test,pull-request,snyk,code-coverage

pipeline 'project-a' created
```
{{< /tab >}}

{{< /tabs >}}

See [kosli create flow](/client_reference/kosli_create_flow/) for more details. 
