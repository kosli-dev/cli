---
title: "Part 7: Attestations"
bookCollapseSection: false
weight: 270
---
# Part 7: Attestations

Attestations are how you record the facts your care about in your software supply chain. They are the evidence that you have performed certain activities, such as running tests, security scans, or ensuring that a certain requirement is met.

Kosli allows you to report different types of attestations about artifacts and trails. For some types, Kosli will process the evidence you provide and conclude whether the evidence proves compliance or otherwise. 

Let's take a look at how to make attestations to Kosli.

The following template is expecting 4 attestations, one for each `name`.

```yml
version: 1
trail:
  attestations:
  - name: jira-ticket
    type: jira
  artifacts:
  - name: backend
    attestations:
    - name: unit-tests
      type: junit
    - name: security-scan
      type: snyk
```

It expects `jira-ticket` on the trail, the `backend` artifact, with `unit-tests` and `security-scan` attached to it. When you make an attestation, you have the choice of what `name` to attach it to:

## Make the `jira-ticket` attestation to a trail

The `jira-ticket` attestation belongs to a single trail and is not linked to a specific artifact. In this example, the id of the trail is the git commit.

```shell
$ kosli attest jira \
    --flow backend-ci \
	--trail $(git rev-parse HEAD) \	
    --name jira-ticket 
    ...
```

## Make the `unit-test` attestation to the `backend` artifact

Some attestations are attached to a specific artifact, like the unit tests for the `backend` artifact. Often, evidence like unit tests are created before the artifact is built. To attach the evidence to the artifact before its creation, use `backend` (the artifact's `name` from the template), as well as `unit-tests` (the attestation's `name` from the template).

```shell
$ kosli attest junit \
    --name backend.unit-tests \
    --flow backend-ci \
    --trail $(git rev-parse HEAD) \
    ...
```

This attestation belongs to any artifact attested with the matching `name` from the template (in this example `backend`) and a matching git commit. 

## Make the `backend` artifact attestation

Once the artifact has been built, it can be attested with the following command.

```shell
$ kosli attest artifact my_company/backend:latest \
	--artifact-type docker \
    --flow backend-ci \
	--trail $(git rev-parse HEAD) \	
    --name backend 
    ...
```

The Kosli CLI will calculate the fingerprint of the docker image called `my_company/backend:latest` and attest it as the `backend` artifact `name` in the trail.

{{< hint info >}}
### Automatically gather git commit and CI environment information
In all attestation commands the Kosli CLI will automatically gather the git commit and other information from the current git repository and the [CI environment](https://docs.kosli.com/integrations/ci_cd/). This is how the git commit is used to match attestation to artifacts.
{{< /hint >}}

## Make the `security-scan` attestation to the `backend` artifact

Often, evidence like snyk reports are created after the artifact is built. In this case, you can attach the evidence to the artifact after its creation. Use `backend` (the artifact's `name` from the template), as well as `security-scan` (the attestation's `name` from the template) to name the attestation.

The following attestation will only belong to the artifact `my_company/backend:latest` attested above and its fingerprint, in this case calculated by the Kosli CLI.

```shell
$ kosli attest snyk \
    --artifact-type docker my_company/backend:latest \
    --name backend.security-scan \
    --flow backend-ci \
    --trail $(git rev-parse HEAD)
    ...
```

{{< hint info >}}
### Attestation immutability

Attestations are append-only immutable records. You can report the same attestation multiple times, and each report will be recorded. However, only the latest version of the attestation is  considered when evaluating trail or artifact compliance.
{{< /hint >}}

## Evidence Vault

Along with attestations data, you can attach additional supporting evidence files. These will be securely stored in Kosli's **Evidence Vault** and can easily be retrieved when needed. Alternatively, you can store the evidence files in your own preferred storage and only attach links to it in the Kosli attestation.

{{< hint info >}}

For `JUnit` attestations (see below), Kosli automatically stores the JUnit XML results files in the Evidence Vault. You can disable this by setting `--upload-results=false` 

{{< /hint >}}

## Attestation types

Currently, we support the following types of evidence:

### Pull requests

If you use GitHub, Bitbucket, Gitlab or Azure DevOps you can use Kosli to verify if a given git commit comes from a pull/merge request. 

{{< hint warning >}}
Currently, the status of the PR does NOT impact the compliance status of the attestation.
{{< /hint >}}

If there is no pull request for the commit, the attestation will be reported as `non-compliant`. You can choose to short-circuit execution in case pull request is missing by using the `--assert` flag.

See the CLI reference for the following commands for more details and examples:

- [attest Github PR ](/client_reference/kosli_attest_pullrequest_github/) 
- [attest Bitbucket PR ](/client_reference/kosli_attest_pullrequest_bitbucket/)
- [attest Gitlab PR ](/client_reference/kosli_attest_pullrequest_gitlab/)
- [attest Azure Devops PR ](/client_reference/kosli_attest_pullrequest_azure/)


### JUnit test results

If you produce your test results in JUnit format, you can attest the test results to Kosli. Kosli will analyze the JUnit results and determine the compliance status based on whether any tests have failed and/or errored or not.

See [attest JUnit results to an artifact or a trail](/client_reference/kosli_attest_junit/) for usage details and examples.

### Snyk security scans 

You can report results of a Snyk security scan to Kosli and it will analyze the Snyk scan results and determine the compliance status based on whether vulnerabilities where found or not.

See [attest Snyk results to an artifact or a trail](/client_reference/kosli_attest_snyk/) for usage details and examples.

### Jira issues

You can use the Jira attestation to verify that a git commit or branch contains a reference to a Jira issue and that an issue with the same reference does exist in Jira.

If Jira reference is found in a commit message, that reference will be reported as evidence. If the reference is not found in the commit message, Kosli CLI will check if it's a part of a branch name.

Kosli CLI will also verify and report if the detected issue reference is found and accessible on Jira (reported as compliant) or not (reported as non compliant). 

See [attest Jira issue to an artifact or a trail](/client_reference/kosli_attest_jira/) for usage details and examples.

### Generic

If Kosli doesn't support the type of the attestation you'd like to attach, you can use the generic type.

Use `--compliant=false` if you want to report a given evidence as non-compliant.

See [report generic attestation to an artifact or a trail](/client_reference/kosli_attest_generic/) for usage details and examples.