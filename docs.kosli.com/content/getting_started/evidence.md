---
title: "Part 6: Evidence"
bookCollapseSection: false
weight: 250
---
# Part 6: Evidence

Whenever an event related to required evidence happens you should report it to Kosli. 
You can report evidence to either a git commit or an artifact. 

{{< hint info >}}

For Kosli to know which evidence you report you need to provide evidence name (using `--name` flag) that is matching one of the names defined in a [flow template](/getting_started/flows/#create-a-flow).

{{< /hint >}}

## Commit evidence vs Artifact evidence

Some types of evidence naturally belong to an **artifact** - e.g. *unit test* or *snyk scan*. Some relate to the source code itself and you can report these on a **commit** - e.g. *code coverage* or *pull request*.  

It's up to you to decide and if you want to attach it all to an artifact it'll work fine. But if you produce multiple artifacts from the same commit you have a possibility to report a commit evidence that will be **automatically** attached to all **artifacts** reported to **ALL or selected flows** built from that commit. That way you won't have to report the same evidence multiple times to each artifact separately.

{{< hint info >}}

Evidence reported against a git commit will be automatically attached to:
* either **ALL** artifacts (in **ALL flows**) produced from that git commit (when `--flows` flag is **not** provided)
* or **only** artifacts produced from that git commit **reported to flows** provided in `--flows` flag (in a comma separated list format).  

If a given named evidence is reported multiple times, it is the compliance status of the 
last reported version of the evidence that is considered the compliance state of that evidence.

{{< /hint >}}

## Does Kosli store evidence?

When you report evidence to Kosli, we store related files in the Evidence Vault. That way you will always have an easy access to the evidence whenever you need it, e.g. in case of an audit. Fingerpints of each evidence file stored in vault will be saved alongside the evidence, which lets you confirm at any time that the evidence wasn't tampered with (or detect tampering).

### What exactly Kosli stores?

Depending on the evidence type we will store:
* for **junit** evidence type: archived directory containing junit test result and the fingerprint of the directory; you can provide the path to the directory using `--results-dir` flag
* for **snyk** evidence type: archived snyk results (in json format) and the fingerprint of the json file; you can provide the path to the snyk results json file using `--snyk-results` flag
* for **generic** evidence type: archived directories and/or files containing evidence and the fingerprint of the directories/files; you can provide a comma-separated list of paths to evidence directories/files using `--evidence-path` 

### Can I opt out from storing evidence? 

You can opt out from storing evidence in Kosli Vault by reporting a **generic** evidence without using the optional `--evidence-paths` flag. Kosli won't be able to determine compliance status, so the responsibility of determining that falls on you. To report evidence as non-compliant use `--compliant=false`, otherwise - for compliant evidence - you can skip the flag (it is set to compliant by default).

You can record the location of evidence files, e.g. if you store them on your own or use another external service to do that, using `--evidence-url` flag, and record the fingerprint of evidence files using `--evidence-fingerprint`

{{< hint info >}}

`--evidence-url` and `--evidence-fingerprint` are only useful if you didn't use `--results-dir`, `--snyk-results` or `--evidence-path` to upload evidence to Kosli Vault

{{< /hint >}}


Currently we support following types of evidence:

## Pull request evidence

If you use GitHub, Bitbucket or Gitlab you can use Kosli to verify if the merge commit you used to build your artifact comes from a pull request. Remember to add the pull request evidence to your [flow template](/kosli_overview/what_is_kosli/#template) and use the same label for `--name` you provided in a `template` 

> note that -currently- the status of the PR does NOT impact the compliance status of the evidence.

If there is no pull request for a given commit, the evidence will be reported as incompliant and the pipeline will continue. You can choose to fail the pipeline altogether in case pull request is missing by using the `--assert` flag.

There are six different pull request commands

For GitHub: [report PR evidence to an artifact](/client_reference/kosli_report_evidence_artifact_pullrequest_github/) 
or [report PR evidence to a commit](/client_reference/kosli_report_evidence_commit_pullrequest_github/) along with the regular flags, you need to provide:
* `--github-org`
* `--github-token` your	Github personal access token with permissions to read PRs.


For Bitbucket: [report PR evidence to an artifact](/client_reference/kosli_report_evidence_artifact_pullrequest_bitbucket/)
or [report PR evidence to a commit](/client_reference/kosli_report_evidence_commit_pullrequest_bitbucket/) along with the regular flags, you need to provide:
*  `--bitbucket-password` - you need to use an api token which is the "App password" you create under "Personal Settings", keep in mind that api tokens you create under "Manage account" won't work for basic auth
* `--bitbucket-username` - you cannot user your email address you use to log in, you have an actual username under "Personal Settings" 
* `--bitbucket-workspace`

For Gitlab: 
[report PR evidence to an artifact](/client_reference/kosli_report_evidence_artifact_pullrequest_gitlab/) 
or [report PR evidence to a commit](/client_reference/kosli_report_evidence_commit_pullrequest_gitlab/) along with the regular flags, you need to provide:
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
[kosli pipeline artifact report evidence github-pullrequest](/legacy_ref/v0.1.41/kosli_pipeline_artifact_report_evidence_github-pullrequest/)  
[kosli pipeline artifact report evidence bitbucket-pullrequest](/legacy_ref/v0.1.41/kosli_pipeline_artifact_report_evidence_bitbucket-pullrequest/)  
[kosli pipeline artifact report evidence gitlab-mergerequest](/legacy_ref/v0.1.41/kosli_pipeline_artifact_report_evidence_gitlab-mergerequest/)
{{< /tab >}}

{{< tab "Commit v2" >}}
```
$ kosli report evidence commit  github-pullrequest \
	--build-url https://exampleci.com \
	--name pull-request \
	--flow project-a \
	--github-token *** \
	--github-org ProjectA \
	--repository repoB \
	--commit e67f2f2b121f9325ebf166b7b3c707f73cb48b14 

github pull request evidence is reported to commit: e67f2f2b121f9325ebf166b7b3c707f73cb48b14
```
For more details see:  
[kosli report evidence commit pullrequest github](/client_reference/kosli_report_evidence_commit_pullrequest_github/)  
[kosli report evidence commit pullrequest bitbucket](/client_reference/kosli_report_evidence_commit_pullrequest_bitbucket/)
[kosli report evidence commit pullrequest github](/client_reference/kosli_report_evidence_commit_pullrequest_gitlab/)
{{< /tab >}}

{{< tab "Commit v0.1.x" >}}
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
[kosli commit report evidence github-pullrequest](/legacy_ref/v0.1.41/kosli_commit_report_evidence_github-pullrequest/)  
[kosli commit report evidence bitbucket-pullrequest](/legacy_ref/v0.1.41/kosli_commit_report_evidence_bitbucket-pullrequest/)  
[kosli commit report evidence gitlab-mergerequest](/legacy_ref/v0.1.41/kosli_commit_report_evidence_gitlab-mergerequest/)
{{< /tab >}}

{{< /tabs >}}

## JUnit test evidence

If you produce your test results in JUnit format, you can [report JUnit evidence to an artifact](/client_reference/kosli_report_evidence_artifact_junit/) or
[report JUnit evidence to a commit](/client_reference/kosli_report_evidence_commit_junit/). These commands will analyze the JUnit results and determine if the evidence is compliant or not.
Remember to add the JUnit test evidence to your [flow template](/kosli_overview/what_is_kosli/#template) and use the same label for `--name` you provided in a `template`.

Use `--results-dir` flag to provide the location of the folder with your XML JUnit test results

### Example

{{< tabs "junit-example" "col-no-wrap" >}}

{{< tab "Artifact v2" >}}
```
$ kosli report evidence artifact junit project-a-app.bin \
	--flow project-a \
	--artifact-type file \
	--build-url https://exampleci.com \
	--name unit-test \
	--results-dir tests

junit test evidence is reported to artifact: 53c97572093cc107c0caa2906d460ccd65083a4c626f68689e57aafa34b14cbf
```
See [kosli report evidence artifact junit](/client_reference/kosli_report_evidence_artifact_junit/) for more details
{{< /tab >}}

{{< tab "Artifact v1.0.x" >}}
```
$ kosli pipeline artifact report evidence junit project-a-app.bin \
	--pipeline project-a \
	--artifact-type file \
	--build-url https://exampleci.com \
	--name unit-test \
	--results-dir tests

junit test evidence is reported to artifact: 53c97572093cc107c0caa2906d460ccd65083a4c626f68689e57aafa34b14cbf
```
See [kosli pipeline artifact report evidence junit](/legacy_ref/v0.1.41/kosli_pipeline_artifact_report_evidence_junit/) for more details
{{< /tab >}}

{{< tab "Commit v2" >}}
```
$ kosli report evidence commit junit \
	--flow project-a \
	--build-url https://exampleci.com \
	--name unit-test \
	--results-dir tests \
	--commit e67f2f2b121f9325ebf166b7b3c707f73cb48b14

junit test evidence is reported to commit: e67f2f2b121f9325ebf166b7b3c707f73cb48b14
```
See [kosli report evidence commit junit](/client_reference/kosli_report_evidence_commit_junit/) for more details
{{< /tab >}}

{{< tab "Commit v0.1.x" >}}
```
$ kosli commit report evidence junit \
	--pipelines project-a \
	--build-url https://exampleci.com \
	--name unit-test \
	--results-dir tests \
	--commit e67f2f2b121f9325ebf166b7b3c707f73cb48b14

junit test evidence is reported to commit: e67f2f2b121f9325ebf166b7b3c707f73cb48b14
```
See [kosli commit report evidence junit](/legacy_ref/v0.1.41/kosli_commit_report_evidence_junit/) for more details
{{< /tab >}}


{{< /tabs >}}

## Snyk scan evidence

To report results of a Snyk security scan, you can [report Snyk evidence to an artifact](/client_reference/kosli_report_evidence_artifact_snyk/) or
[report Snyk evidence to a commit](/client_reference/kosli_report_evidence_commit_snyk/). These commands will analyze the Snyk scan results and determine if the evidence is compliant or not.
Remember to add the snyk scan evidence to your [flow template](/kosli_overview/what_is_kosli/#template) and use the same label for `--name` you provided in a `template`.

Use `--scan-results` flag to provide the location of the json file with your snyk scan results

### Example

{{< tabs "snyk-example" "col-no-wrap" >}}

{{< tab "Artifact v2" >}}
```
$ kosli report evidence artifact snyk project-a-app.bin \
	--pipeline project-a \
	--artifact-type file \
	--build-url https://exampleci.com \
	--name snyk \
	--scan-results snyk_scam.json 

snyk scan evidence is reported to artifact: 53c97572093cc107c0caa2906d460ccd65083a4c626f68689e57aafa34b14cbf
```
See [kosli report evidence artifact snyk](/client_reference/kosli_report_evidence_artifact_snyk/) for more details
{{< /tab >}}

{{< tab "Artifact v1.0.x" >}}
```
$ kosli pipeline artifact report evidence snyk project-a-app.bin \
	--pipeline project-a \
	--artifact-type file \
	--build-url https://exampleci.com \
	--name snyk \
	--scan-results snyk_scam.json 

snyk scan evidence is reported to artifact: 53c97572093cc107c0caa2906d460ccd65083a4c626f68689e57aafa34b14cbf
```
See [kosli pipeline artifact report evidence snyk](/legacy_ref/v0.1.41/kosli_pipeline_artifact_report_evidence_snyk/) for more details
{{< /tab >}}

{{< tab "Commit v2" >}}
```
$ kosli report evidence commit snyk \
	--flow project-a \
	--build-url https://exampleci.com \
	--name snyk \
	--scan-results snyk_scam.json \
	--commit e67f2f2b121f9325ebf166b7b3c707f73cb48b14

snyk scan evidence is reported to commit: e67f2f2b121f9325ebf166b7b3c707f73cb48b14
```
See [kosli report evidence commit snyk](/client_reference/kosli_report_evidence_commit_snyk/) for more details
{{< /tab >}}

{{< tab "Commit v0.1.x" >}}
```
$ kosli commit report evidence snyk \
	--pipelines project-a \
	--build-url https://exampleci.com \
	--name snyk \
	--scan-results snyk_scam.json \
	--commit e67f2f2b121f9325ebf166b7b3c707f73cb48b14

snyk scan evidence is reported to commit: e67f2f2b121f9325ebf166b7b3c707f73cb48b14
```
See [kosli commit report evidence snyk](/legacy_ref/v0.1.41/kosli_commit_report_evidence_snyk/) for more details
{{< /tab >}}

{{< /tabs >}}

## Jira evidence 

To verify that Jira issue reference is a part of a commit message or a branch name, and report it to Kosli, you can use [kosli report evidence commit jira](/client_reference/kosli_report_evidence_commit_jira/) command. 

If Jira reference is found in a commit message, that reference will be reported as evidence. If the reference is not found in the commit message, Kosli CLI will check if it's a part of a branch name.

Kosli CLI will also verify and report if the detected issue reference is found and accessible on Jira (reported as compliant) or not (reported as non compliant). 


### Example 

{{< tabs "jira-example" "col-no-wrap" >}}

{{< tab "Commit v2" >}}
```
$ kosli report evidence commit jira \
	--commit yourGitCommitSha1 \
	--name yourEvidenceName \
	--jira-base-url https://kosli.atlassian.net \
	--jira-username user@domain.com \
	--jira-api-token yourJiraAPIToken \
	--flows yourFlowName \
	--build-url https://exampleci.com \
	--api-token yourAPIToken \
	--org yourOrgName

snyk scan evidence is reported to commit: e67f2f2b121f9325ebf166b7b3c707f73cb48b14
```
See [kosli report evidence commit jira](/client_reference/kosli_report_evidence_commit_jira/) for more details
{{< /tab >}}

{{< /tabs >}}

## Generic evidence

If Kosli doesn't support the type of the evidence you'd like to attach, you can [report Generic evidence to an artifact](/client_reference/kosli_report_evidence_artifact_generic/) or
[report Generic evidence to a commit](/client_reference/kosli_report_evidence_commit_generic/).
Remember to add the evidence to your [flow template](/kosli_overview/what_is_kosli/#template) and use the same label for `--name` you provided in a `template`.

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
See [kosli pipeline artifact report evidence generic](/legacy_ref/v0.1.41/kosli_pipeline_artifact_report_evidence_generic/) for more details
{{< /tab >}}

{{< tab "Commit v2" >}}
```
$ kosli report evidence commit generic \
	--flow project-a \
	--build-url https://exampleci.com \
	--name code-coverage \
	--compliant=false \
	--commit e67f2f2b121f9325ebf166b7b3c707f73cb48b14

generic evidence 'code-coverage' is reported to commit: e67f2f2b121f9325ebf166b7b3c707f73cb48b14
```
See [kosli report evidence commit generic](/client_reference/kosli_report_evidence_commit_generic/) for more details
{{< /tab >}}

{{< tab "Commit v0.1.x" >}}
```
$ kosli commit report evidence generic \
	--pipelines project-a \
	--build-url https://exampleci.com \
	--name code-coverage \
	--compliant=false \
	--commit e67f2f2b121f9325ebf166b7b3c707f73cb48b14

generic evidence 'code-coverage' is reported to commit: e67f2f2b121f9325ebf166b7b3c707f73cb48b14
```
See [kosli commit report evidence generic](/legacy_ref/v0.1.41/kosli_commit_report_evidence_generic/) for more details
{{< /tab >}}

{{< /tabs >}}
