---
title: Flow Template Specification
---
# Flow Template Specification

This document describes the specification for how to write your Flow Template files in [YAML](http://yaml.org/). The template file contains the following fields:

```yml
version: The version of the specification schema. Allowed values are [1]. (required)
trail: # the trail specification (optional)
  attestations: # what attestations are required for the trail to be compliant (optional)
  - name: the attestation name (required)
    type: the attestation type. One of [generic, jira, junit, pull-request, snyk] (required)
  artifacts: # what artifacts are expected to be produced in the trail (optional)
  - name: reference name for the artifact (e.g. frontend-app) (required)
    attestations: # what attestations are required for the artifact to be compliant
    - name: the attestation name (required)
      type: the attestation type. One of [generic, jira, junit, pull-request, snyk] (required)
```
 
## Example:

```yaml
version: 1
trail:
  attestations:
  - name: jira-ticket
    type: jira
  - name: risk-level-assessment
    type: generic
  artifacts:
  - name: backend
    attestations:
    - name: unit-tests
      type: junit
    - name: security-scan
      type: snyk
  - name: frontend
    attestations:
    - name: manual-ui-test
      type: generic
```
