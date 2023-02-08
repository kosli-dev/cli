---
title: "Part 5: Evidence"
bookCollapseSection: false
weight: 250
---
# Part 5: Evidence

Whenever an event related to one of your evidences happens you should report it to Kosli. 

Currently we support following types of evidences:

## Pull request evidence

If you use GitHub or Bitbucket you can use Kosli to verify if the merge commit you used to build your artifact comes from a pull request. Remember to add the pull request evidence to your [pipeline template](/kosli_overview/what_is_kosli/#template) and use the same label for `--evidence-type` you provided in a `template` 

> note that -currently- the status of the PR does NOT impact the compliance status of the evidence.

If there is no pull request for a given commit, the evidence will be reported as incompliant and the pipeline will continue. You can choose to fail the pipeline altogether in case pull request is missing by using the `--assert` flag.

There are two different pull request commands, depending on your CI.

For GitHub: [kosli pipeline artifact report evidence github-pullrequest](/client_reference/kosli_pipeline_artifact_report_evidence_github-pullrequest/) along with the regular flags, you need to provide:
* `--github-org`
* `--github-token` your	Github personal access token with permissions to read PRs.


For Bitbucket: [kosli pipeline artifact report evidence bitbucket-pullrequest](/client_reference/kosli_pipeline_artifact_report_evidence_bitbucket-pullrequest/) along with the regular flags, you need to provide:
*  `--bitbucket-password` - you need to use an api token which is the "App password" you create under "Personal Settings", keep in mind that api tokens you create under "Manage account" won't work for basic auth
* `--bitbucket-username` - you cannot user your email address you use to log in, you have an actual username under "Personal Settings" 
* `--bitbucket-workspace`

### Example
 
```
$ kosli pipeline artifact report evidence github-pullrequest project-a-app.bin \
	--artifact-type file \
	--build-url https://exampleci.com \
	--name pull-request \
	--pipeline project-a \
	--github-token *** \
	--github-org ProjectA \
	--commit e67f2f2b121f9325ebf166b7b3c707f73cb48b14 

github pull request evidence is reported to artifact: 53c97572093cc107c0caa2906d460ccd65083a4c626f68689e57aafa34b14cbf
```
See [kosli pipeline artifact report evidence bitbucket-pullrequest](/client_reference/kosli_pipeline_artifact_report_evidence_bitbucket-pullrequest/) or See [kosli pipeline artifact report evidence github-pullrequest](/client_reference/kosli_pipeline_artifact_report_evidence_github-pullrequest/) for more details. 
for more details. 

## JUnit test evidence

If you produce your test results in JUnit format, you can use `kosli pipeline artifact report evidence junit` command to analyze the results and report it to Kosli. Remember to add the junit test evidence to your [pipeline template](/kosli_overview/what_is_kosli/#template) and use the same label for `--evidence-type` you provided in a `template`.

Use `--results-dir` flag to provide the location of the folder with your junit test results

### Example
 
```
$ kosli pipeline artifact report evidence junit project-a-app.bin \
	--pipeline project-a \
	--artifact-type file \
	--build-url https://exampleci.com \
	--name unit-test \
	--results-dir tests

junit test evidence is reported to artifact: 53c97572093cc107c0caa2906d460ccd65083a4c626f68689e57aafa34b14cbf
```
See [kosli pipeline artifact report evidence junit](/client_reference/kosli_pipeline_artifact_report_evidence_junit/) for more details. 

## Snyk scan evidence

To report results of scan security scan use `kosli pipeline artifact report evidence junit` command to analyze the results and report it to Kosli. Remember to add the snyk scan evidence to your [pipeline template](/kosli_overview/what_is_kosli/#template) and use the same label for `--evidence-type` you provided in a `template`.

Use `--scan-results` flag to provide the location of the json file with your snyk scan results

### Example
 
```
$ kosli pipeline artifact report evidence snyk project-a-app.bin \
	--pipeline project-a \
	--artifact-type file \
	--build-url https://exampleci.com \
	--name snyk \
	--scan-results snyk_scam.json 

snyk scan evidence is reported to artifact: 53c97572093cc107c0caa2906d460ccd65083a4c626f68689e57aafa34b14cbf
```
See [kosli pipeline artifact report evidence snyk](/client_reference/kosli_pipeline_artifact_report_evidence_snyk/) for more details. 

## Generic evidence

If Kosli doesn't support the type of the evidence you'd like to attach, you can use [kosli pipeline artifact report evidence generic](/client_reference/kosli_pipeline_artifact_report_evidence_generic/) command to report such evidence. Remember to add the evidence to your [pipeline template](/kosli_overview/what_is_kosli/#template) and use the same label for `--evidence-type` you provided in a `template`.

Use `--compliant=false` if you want to report given evidence as non-compliant.
### Example
 
```
$ kosli pipeline artifact report evidence generic project-a-app.bin \
	--pipeline project-a \
	--artifact-type file \
	--build-url https://exampleci.com \
	--evidence-type code-coverage \
	--compliant=false

generic evidence 'code-coverage' is reported to artifact: 53c97572093cc107c0caa2906d460ccd65083a4c626f68689e57aafa34b14cbf
```
See [kosli pipeline artifact report evidence generic](/client_reference/kosli_pipeline_artifact_report_evidence_generic/) for more details. 