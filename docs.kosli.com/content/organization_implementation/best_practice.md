---
title: 'Technical best practices'
weight: 140
summary: ""
---

## Naming convention

* Flows:
  * CI part: name your flows like "repo"/"application".
  * CD part: name your flow "release-name"
* Trail:
  * Default to SHA
  * Others could be release version (Git tag), merge number etc.

## Flow grouping

Centralized pipelines means that you will get a lot of flows based on the same template.

Solution; Add tags to your flows so that all with the same "template" will have the same tag. In that way you do not need to know the name, but can evaluate based on tags.

## Governance

* Streamline your pipelines

## Automatic creation of elements

The creation commands (`kosli create attestation-type`, `environment`, `flow` etc) have a "create or update" behaviour. That means that you can set creation of environments, policies, flows etc. inside your pipeline. If the information is the identical to what is stored in Kosli, no action will be taken. If the information differes, it will update the current object.

## Too many custom attestation types for test formats

* Use [Common Test Report Format](https://ctrf.io/). Create one custom attestation type:
`kosli create attestation-type ctrf --jq ".results.summary.failed > 0" --schema schema.json`
* Use that across your test report types to make a uniform JSON structure to query afterwards.

## Environment policies

When creating policies, we have found it best to split up policies into their logical units. That way, it is easier to maintain and understand the logical connection between requirements.

An example of that would be the Cyber-Dojo environment policy repository on [Github](https://github.com/cyber-dojo/kosli-environment-policies).

In that repository we have four poilicices attached:

* artifact-provenance.yml to make sure that all artifacts going in to this environment has provenance.
* compliant-build-process.yml
* pull-request.yml The trail the artifact is attested to needs to be merged through a pull request
* security-scan.yml The trail the artifact is attested to has a security scan attestation associated.
