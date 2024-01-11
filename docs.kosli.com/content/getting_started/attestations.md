---
title: "Part 7: Attestations"
bookCollapseSection: false
weight: 270
---
# Part 7: Attestations

Attestations lets you declare whether an Artifact or a Trail adhere to a certain requirement or not. The attestation often includes evidence proofing the claim.

Kosli allows you to report different types of attestations about an Artifact or a Trail. For some types, Kosli will process the evidence you provide and conclude whether the evidence proofs compliance or otherwise. 


## Trail attestation vs Artifact attestation

Depending on your process requirements, some attestations will belong to an Artifact while others will belong to a Trail. When you report an attestation, you have the choice of where to attach it:

1. **To a trail**: the attestation belongs to a single trail and is not linked to a specific artifact.
2. **To an Artifact**: the attestation belongs to a specific artifact.

## Binding attestations to an artifact

To bind an attestation to an artifact, you have two options:
1. **Binding with the fingerprint**: the attestation belongs only to the artifact with that fingerprint. This requires that the artifact has already been reported to Kosli.
2. **Binding with template name and git commit**: the attestation belongs to any artifact that **has been or will be** reported with the specified template name and a matching git commit. For instance, if multiple artifacts are reported as `backend` from the same trail, and an attestation has been reported targeting template name `backend`. The attestation will be bound to the `backend` artifacts that has the same git commit as the attestation.


{{< hint info >}}

Attestations are append-only immutable records. You can report the same attestation multiple times, and each report will be recorded. However, only the latest version of the attestation is  considered when evaluating trail or artifact compliance.

{{< /hint >}}


## Evidence Vault

Along with attestations data, you can attach additional supporting evidence files. These will be securely stored in Kosli's **Evidence Vault** and can easily be retrieved when needed. Alternatively, you can store the evidence files in your own preferred storage and only attach links to it in the Kosli attestation.

{{< hint info >}}

For `JUnit` attestations (see below), Kosli automatically stores the JUnit XML results files in the Evidence Vault. You can disable this by setting `--upload-results=false` 

{{< /hint >}}

## Attestation types

Currently we support the following types of evidence:

### Pull requests

If you use GitHub, Bitbucket, Gitlab or Azure DevOps you can use Kosli to verify if a given git commit comes from a pull/merge request. 

{{< hint info >}}
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

See [attest JUnit results o an artifact or a trail](/client_reference/kosli_attest_junit/) for usage details and examples.

### Snyk security scans 

You can report results of a Snyk security scan to Kosli and it will analyze the Snyk scan results and determine the compliance status based on whether vulnerabilities where found or not.

{{< hint warning >}}
Currently, only `snyk container scan` results are supported for this type. For `snyk code scan` please use the generic attestation type.
{{< /hint >}}

See [attest Snyk results o an artifact or a trail](/client_reference/kosli_attest_snyk/) for usage details and examples.


### Jira issues

You can use the Jira attestation to verify that a git commit or branch contains a reference to a Jira issue and that an issue with the same reference does exist in Jira.

If Jira reference is found in a commit message, that reference will be reported as evidence. If the reference is not found in the commit message, Kosli CLI will check if it's a part of a branch name.

Kosli CLI will also verify and report if the detected issue reference is found and accessible on Jira (reported as compliant) or not (reported as non compliant). 

See [attest Jira issue to an artifact or a trail](/client_reference/kosli_attest_jira/) for usage details and examples.


### Generic

If Kosli doesn't support the type of the attestation you'd like to attach, you can use the generic type.

Use `--compliant=false` if you want to report a given evidence as non-compliant.

See [report generic attestation to an artifact or a trail](/client_reference/kosli_attest_generic/) for usage details and examples.