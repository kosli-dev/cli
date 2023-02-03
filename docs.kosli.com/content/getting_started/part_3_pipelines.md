---
title: "Part 3: Pipelines"
bookCollapseSection: false
weight: 230
---
# Part 3: Pipelines

Kosli allows you to connect the development world (commits, builds, tests, approvals, deployments) with what’s happening in operations. There is a variety of commands that let you report all the necessary information to Kosli and - relying on automatically calculated fingerprints of your artifacts - match it with the environments.

## Create a pipeline

To report artifacts to Kosli you need to create a Kosli [pipeline](/kosli_overview/what_is_kosli/#pipelines) first. When you create a pipeline you also define expected controls - a list of evidences you need to be reported in order for the artifact to become compliant. Use the `--template` flag to provide the list of requirements. 

Later, when reporting an evidence for a specific control you will use the same name you used in the template to identify which evidence you are reporting.

It is a normal practice to include `kosli pipeline declare` command in the same CI pipeline you use to build the artifact you want to report to that Kosli pipeline. None of the previously reported artifacts will be overwritten or lost. And if you change the template, by adding or removing required evidence, it won't affect the compliancy status of existing artifacts.

### Example

```
# create/update a Kosli pipeline
kosli pipeline declare \
	--pipeline yourPipelineName \
	--description yourPipelineDescription \
  	--visibility private OR public \
	--template artifact,unit-test,pull-request,code-coverage \
	--api-token yourAPIToken \
	--owner yourOrgName
```
