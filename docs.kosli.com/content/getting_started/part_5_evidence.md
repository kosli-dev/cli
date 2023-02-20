---
title: "Part 5: Evidence"
bookCollapseSection: false
weight: 250
---
# Part 5: Evidence

Whenever an event related to required evidence happens you should report it to Kosli. 
You can report evidence to either a git commit or an artifact. 
Evidence reported against a git commit will be automatically 
attached to any artifact produced from that git commit. 

Currently we support following types of evidence:

## Pull request evidence

If you use GitHub, Bitbucket or Gitlab you can use Kosli to verify if the merge commit you used to build your artifact comes from a pull request. Remember to add the pull request evidence to your [pipeline template](/kosli_overview/what_is_kosli/#template) and use the same label for `--name` you provided in a `template` 

> note that -currently- the status of the PR does NOT impact the compliance status of the evidence.

If there is no pull request for a given commit, the evidence will be reported as incompliant and the pipeline will continue. You can choose to fail the pipeline altogether in case pull request is missing by using the `--assert` flag.

There are six different pull request commands

For GitHub: [report PR evidence to an artifact](/client_reference/kosli_report_evidence_artifact_pullrequest_github/) 
or [report PR evidence to a commit](/client_reference/kosli_commit_report_evidence_github-pullrequest/) along with the regular flags, you need to provide:
* `--github-org`
* `--github-token` your	Github personal access token with permissions to read PRs.


For Bitbucket: [report PR evidence to an artifact](/client_reference/kosli_report_evidence_artifact_pullrequest_bitbucket/)
or [report PR evidence to a commit](/client_reference/kosli_commit_report_evidence_bitbucket-pullrequest/) along with the regular flags, you need to provide:
*  `--bitbucket-password` - you need to use an api token which is the "App password" you create under "Personal Settings", keep in mind that api tokens you create under "Manage account" won't work for basic auth
* `--bitbucket-username` - you cannot user your email address you use to log in, you have an actual username under "Personal Settings" 
* `--bitbucket-workspace`

For Gitlab: 
[report PR evidence to an artifact](/client_reference/kosli_report_evidence_artifact_pullrequest_gitlab/) 
or [report PR evidence to a commit](/client_reference/kosli_commit_report_evidence_gitlab-mergerequest/) along with the regular flags, you need to provide:
* `--gitlab-org`
* `--gitlab-token` your	Gitlab personal access token with permissions to read Merge requests.
### Example

{{< tabs "gh-pr-example" "col-no-wrap" >}}

{{< tab "Artifact v2" >}}
```
$ kosli report evidence artifact pullrequest github project-a-app.bin \
	--artifact-type file \
	--build-url https://exampleci.com \
	--name pull-request \
	--flow project-a \
	--github-token *** \
	--github-org ProjectA \
	--repository repoB \
	--commit e67f2f2b121f9325ebf166b7b3c707f73cb48b14 

github pull request evidence is reported to artifact: 53c97572093cc107c0caa2906d460ccd65083a4c626f68689e57aafa34b14cbf
```
For more details see:  
[kosli report evidence artifact pullrequest github](/client_reference/kosli_report_evidence_artifact_pullrequest_github/)  
[kosli report evidence artifact pullrequest bitbucket](/client_reference/kosli_report_evidence_artifact_pullrequest_bitbucket/)  
[kosli report evidence artifact pullrequest gitlab](/client_reference/kosli_report_evidence_artifact_pullrequest_gitlab/) 
{{< /tab >}}

{{< tab "Artifact v0.1.x" >}}
```
$ kosli pipeline artifact report evidence github-pullrequest project-a-app.bin \
	--artifact-type file \
	--build-url https://exampleci.com \
	--name pull-request \
	--pipeline project-a \
	--github-token *** \
	--github-org ProjectA \
	--repository repoB \
	--commit e67f2f2b121f9325ebf166b7b3c707f73cb48b14 

github pull request evidence is reported to artifact: 53c97572093cc107c0caa2906d460ccd65083a4c626f68689e57aafa34b14cbf
```
For more details see:  
[kosli pipeline artifact report evidence github-pullrequest](/legacy_ref/v0.1.35/kosli_pipeline_artifact_report_evidence_github-pullrequest/)  
[kosli pipeline artifact report evidence bitbucket-pullrequest](/legacy_ref/v0.1.35/kosli_pipeline_artifact_report_evidence_bitbucket-pullrequest/)
{{< /tab >}}

{{< tab "Commit" >}}
```
$ kosli commit report evidence github-pullrequest \
	--build-url https://exampleci.com \
	--name pull-request \
	--pipelines project-a \
	--github-token *** \
	--github-org ProjectA \
	--repository repoB \
	--commit e67f2f2b121f9325ebf166b7b3c707f73cb48b14 

github pull request evidence is reported to commit: e67f2f2b121f9325ebf166b7b3c707f73cb48b14
```
For more details see:  
[kosli commit report evidence github-pullrequest](/client_reference/kosli_commit_report_evidence_github-pullrequest/)  
[kosli commit report evidence bitbucket-pullrequest](/client_reference/kosli_commit_report_evidence_bitbucket-pullrequest/)
{{< /tab >}}

{{< /tabs >}}

## JUnit test evidence

If you produce your test results in JUnit format, you can [report JUnit evidence to an artifact](/client_reference/kosli_pipeline_artifact_report_evidence_junit/) or
[report JUnit evidence to a commit](/client_reference/kosli_commit_report_evidence_junit/). These commands will analyze the JUnit results and determine if the evidence is compliant or not.
Remember to add the JUnit test evidence to your [pipeline template](/kosli_overview/what_is_kosli/#template) and use the same label for `--name` you provided in a `template`.

Use `--results-dir` flag to provide the location of the folder with your XML JUnit test results

### Example

{{< tabs "junit-example" "col-no-wrap" >}}

{{< tab "Artifact" >}}
```
$ kosli pipeline artifact report evidence junit project-a-app.bin \
	--pipeline project-a \
	--artifact-type file \
	--build-url https://exampleci.com \
	--name unit-test \
	--results-dir tests

junit test evidence is reported to artifact: 53c97572093cc107c0caa2906d460ccd65083a4c626f68689e57aafa34b14cbf
```
{{< /tab >}}

{{< tab "Commit" >}}
```
$ kosli commit report evidence junit \
	--pipelines project-a \
	--build-url https://exampleci.com \
	--name unit-test \
	--results-dir tests \
	--commit e67f2f2b121f9325ebf166b7b3c707f73cb48b14

junit test evidence is reported to commit: e67f2f2b121f9325ebf166b7b3c707f73cb48b14
```
{{< /tab >}}

{{< /tabs >}}

## Snyk scan evidence

To report results of a Snyk security scan, you can [report Snyk evidence to an artifact](/client_reference/kosli_pipeline_artifact_report_evidence_snyk/) or
[report Snyk evidence to a commit](/client_reference/kosli_commit_report_evidence_snyk/). These commands will analyze the Snyk scan results and determine if the evidence is compliant or not.
Remember to add the snyk scan evidence to your [pipeline template](/kosli_overview/what_is_kosli/#template) and use the same label for `--name` you provided in a `template`.

Use `--scan-results` flag to provide the location of the json file with your snyk scan results

### Example

{{< tabs "snyk-example" "col-no-wrap" >}}

{{< tab "Artifact" >}}
```
$ kosli pipeline artifact report evidence snyk project-a-app.bin \
	--pipeline project-a \
	--artifact-type file \
	--build-url https://exampleci.com \
	--name snyk \
	--scan-results snyk_scam.json 

snyk scan evidence is reported to artifact: 53c97572093cc107c0caa2906d460ccd65083a4c626f68689e57aafa34b14cbf
```
{{< /tab >}}

{{< tab "Commit" >}}
```
$ kosli commit report evidence snyk \
	--pipelines project-a \
	--build-url https://exampleci.com \
	--name snyk \
	--scan-results snyk_scam.json \
	--commit e67f2f2b121f9325ebf166b7b3c707f73cb48b14

snyk scan evidence is reported to commit: e67f2f2b121f9325ebf166b7b3c707f73cb48b14
```
{{< /tab >}}

{{< /tabs >}}

## Generic evidence

If Kosli doesn't support the type of the evidence you'd like to attach, you can [report Generic evidence to an artifact](/client_reference/kosli_pipeline_artifact_report_evidence_generic/) or
[report Generic evidence to a commit](/client_reference/kosli_commit_report_evidence_generic/).
Remember to add the evidence to your [pipeline template](/kosli_overview/what_is_kosli/#template) and use the same label for `--name` you provided in a `template`.

Use `--compliant=false` if you want to report a given evidence as non-compliant.
### Example

{{< tabs "generic-example" "col-no-wrap" >}}

{{< tab "Artifact v2">}}
```
$ kosli report evidence artifact generic project-a-app.bin \
	--flow project-a \
	--artifact-type file \
	--build-url https://exampleci.com \
	--name code-coverage \
	--compliant=false

generic evidence 'code-coverage' is reported to artifact: 53c97572093cc107c0caa2906d460ccd65083a4c626f68689e57aafa34b14cbf
```
See [kosli report evidence artifact generic](/client_reference/kosli_report_evidence_artifact_generic/) for more details
{{< /tab >}}

{{< tab "Artifact v0.1.x">}}
```
$ kosli pipeline artifact report evidence generic project-a-app.bin \
	--pipeline project-a \
	--artifact-type file \
	--build-url https://exampleci.com \
	--name code-coverage \
	--compliant=false

generic evidence 'code-coverage' is reported to artifact: 53c97572093cc107c0caa2906d460ccd65083a4c626f68689e57aafa34b14cbf
```
See [kosli pipeline artifact report evidence generic](/legacy_ref/v0.1.35/kosli_pipeline_artifact_report_evidence_generic/) for more details
{{< /tab >}}

{{< tab "Commit" >}}
```
$ kosli commit report evidence generic \
	--pipelines project-a \
	--build-url https://exampleci.com \
	--name code-coverage \
	--compliant=false \
	--commit e67f2f2b121f9325ebf166b7b3c707f73cb48b14

generic evidence 'code-coverage' is reported to commit: e67f2f2b121f9325ebf166b7b3c707f73cb48b14
```
See [kosli commit report evidence generic](/client_reference/kosli_commit_report_evidence_generic/) for more details
{{< /tab >}}

{{< /tabs >}}
