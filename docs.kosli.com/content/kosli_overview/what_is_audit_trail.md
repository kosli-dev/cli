---
title: 'Kosli Audit trail'
beta: true
weight: 130
---
# Kosli Audit trail

{{< hint warning >}}
**Kosli Audit trail** is a beta feature.
Beta features provide early access to product functionality. These
features may change between releases without warning, or can be removed from a 
future release. You can enable beta features by using the `kosli enable beta` command.
{{< /hint >}}

An audit trail is a system for automating collecting documentation for each step of a business process.

{{< hint info >}}
Note that **all** CLI command flags can be set as environment variables by adding the the `KOSLI_` prefix and capitalizing them. 
In the command examples below both `--api-token` and `--org` flags were set from environment variables.
{{< /hint >}}

## Audit trail
For each business process you create a Kosli audit trail.
When creating an audit trail you define a list of steps required for completing the process.

An example could be a bank which has a process for **Large money transfer**,
with steps **user-verification** and **risk-assessment**

To create an audit trail for such a process you could run the following command:
```
kosli create audit-trail LargeMoneyTransfer --steps user-verification,risk-assessment
```

## Workflow
A workflow is a single run of a business process. For example for **Large money transfer**
we would report a new workflow every time a new transfer attempt occurs. To distinguish
the different workflows we rely on unique IDs. ID can be anything you use to identify
processes in your system.

To report a workflow for such the **Large money transfer** you could run the following command:
```
kosli report workflow --audit-trail LargeMoneyTransfer --id lmf-56
```

## Evidence
For each step in a process you can report evidence of failed or successful steps,
and upload evidence files if needed.

To report the execution of the **user-verification** for such the **Large money transfer**
you could run the following command:
```
kosli report evidence workflow --audit-trail LargeMoneyTransfer --id lmf-56 --step user-verification
```
