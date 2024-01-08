---
title: "Part 4: Flows"
bookCollapseSection: false
weight: 240
---
# Part 4: Flows

A Kosli Flow represents a business or software process that requires change tracking. It allows you to monitor changes across all steps within a process or focus specifically on a subset of essential steps.

Prior to creating a Flow, it's essential to establish a `Flow Template`. This template defines the necessary steps within the business or software process represented by a Kosli Flow. The compliance of Flow trails and artifacts is assessed based on the template.

{{< hint info >}}
In all the commands below we skip the required `--api-token` and `--org` flags for brevity. These can be set as described [here](/getting_started/install#assigning-flags-via-config-files).
{{< /hint >}}

## Create a Flow Template

A Flow template is a YAML file following the syntax outlined in the [flow template spec](/template_ref).

Here is an example:

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
```

## Create a flow

To create a Flow, you can run:

```shell
$ kosli create flow process-a --description "My SW delivery process" \
 --template-file sw-delivery-template.yml
```

Rerunning the command with different description or template file will update the Flow. 

See [kosli create flow](/client_reference/kosli_create_flow/) for more details. 
