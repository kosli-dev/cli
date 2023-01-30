---
title: "Part 5: Evidences"
bookCollapseSection: false
weight: 250
---
# Part 5: Evidences

## Report evidences

Whenever an event related to one of your evidences happens you should report it to Kosli. 

Currently we support following types of evidences:

### Pull request

If you use GitHub or Bitbucket you can use Kosli to verify if the merge commit you used to build your artifact comes from a pull request. Remember to add the pull request evidence to your [pipeline template](/how_to/connect/#create-a-pipeline) and use the same label for `--evidence-type` you provided in a `template` 

> note that -currently- the status of the PR does NOT impact the compliance status of the evidence.

If there is no pull request for a given commit, the evidence will be reported as incompliant and the pipeline will continue. You can choose to fail the pipeline altogether in case pull request is missing by using the `--assert` flag.

There are two different commands for that, depending on your CI.

For GitHub: [kosli pipeline artifact report evidence github-pullrequest](/client_reference/kosli_pipeline_artifact_report_evidence_github-pullrequest/) along regular flags, you need to provide:
* `--github-org`
* `--github-token` your	Github personal access token with permissions to read PRs.


For Bitbucket: [kosli pipeline artifact report evidence bitbucket-pullrequest](/client_reference/kosli_pipeline_artifact_report_evidence_bitbucket-pullrequest/) along regular flags, you need to provide:
*  `--bitbucket-password` - you need to use an api token which is the "App password" you create under "Personal Settings", keep in mind that api tokens you create under "Manage account" won't work for basic auth
* `--bitbucket-username` - you cannot user your email address you use to log in, you have an actual username under "Personal Settings") 
* `--bitbucket-workspace`


### JUnit test 

If you produce your test results in JUnit format, you can use `kosli pipeline artifact report evidence test` command to analyze the results and report it to Kosli. Remember to add the junit test evidence to your [pipeline template](/how_to/connect/#create-a-pipeline) and use the same label for `--evidence-type` you provided in a `template` 

### Example

```
# report a JUnit test evidence about a file artifact:
kosli pipeline artifact report evidence test FILE.tgz \
	--artifact-type file \
	--evidence-type yourEvidenceType \
	--pipeline yourPipelineName \
	--build-url https://exampleci.com \
	--api-token yourAPIToken \
	--owner yourOrgName	\
	--results-dir yourFolderWithJUnitResults

# report a JUnit test evidence about an artifact using an available Sha256 digest:
kosli pipeline artifact report evidence test \
	--sha256 yourSha256 \
	--evidence-type yourEvidenceType \
	--pipeline yourPipelineName \
	--build-url https://exampleci.com \
	--api-token yourAPIToken \
	--owner yourOrgName	\
	--results-dir yourFolderWithJUnitResults
```

[junit test](/client_reference/kosli_pipeline_artifact_report_evidence_test/) 

### Generic

