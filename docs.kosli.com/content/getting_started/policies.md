---
title: "Part 9: Environment Policies"
bookCollapseSection: false
weight: 290
summary: "Environment Policies enable you to define and enforce compliance requirements for artifact deployments across different environments."
---

# Part 9: Environment Policies

Environment Policies enable you to define and enforce compliance requirements for artifact deployments across 
different environments. With Environment Policies, you can:

- Define specific requirements for each environment (e.g, dev, staging, prod)
- Enforce consistent compliance standards across your deployment pipeline
- Prevent non-compliant artifacts from being deployed (via admission controllers)

Policies are written in YAML and are immutable (updating a policy creates a new version). They can be attached to 
one or more environments, and an environment can have one or more policies attached to it.

## Create a Policy

You can create a policy via CLI or via the API. Here is a basic policy that requires provenance and specific 
attestations:

```yaml {.command}
# prod-policy.yaml
_schema: https://kosli.com/schemas/policy/environment/v1
artifacts: # the rules apply to artifacts in an environment snapshot
  provenance:
    required: true # all artifacts must have provenance
  attestations:
    - name: dependency-scan # all artifacts must have dependency-scan attestation
      type: "*" # any attestation type
    - name: unit-test # all artifacts must have unit-test attestation
      type: junit # must be a 'junit' attestation type
```

You can create and manage policies using the Kosli CLI (global flags like org and api-token are omitted for brevity):

```shell {.command}
kosli create policy prod-requirements prod-policy.yaml
```

```shell {.command}
kosli get policy prod-requirements
```

See [kosli create policy](/client_reference/kosli_create_policy/) for usage details and examples.

{{% hint info %}}
Once you create a policy, you will be able to see it in the UI under `policies` in the left navigation menu.
{{% /hint %}}

## Declarative Policy Syntax

A Policy is declaratively defined according to the following schema:

```yaml {.command}
_schema: https://kosli.com/schemas/policy/environment/v1

artifacts:
  provenance:
    required: true | false (default = false)
    exceptions: (default [])
    - if: ${{ expression }}

  trail-compliance:
    required: true | false (default = false)
    exceptions: (default [])
    - if: ${{ expression }}

  attestations: (default [])
    - if: ${{ expression }} (default = true)
      name: str (default = "*") # cannot have both name and type as *
      type: oneOf ['*', 'junit', 'jira', 'pull_request', 'snyk', 'sonar', 'generic', 'custom:<custom-type-name>'] (default = "*") # cannot have both name and type as *
```

### Policy Rules

A policy consists of `rules` which are applied to artifacts in an environment snapshot.

#### Provenance

```yaml {.command}
artifacts:
  provenance:
    required: true # Requires artifact to be part of a Kosli Flow
```

#### Trail Compliance

```yaml {.command}
artifacts:
  trail-compliance:
    required: true # Requires the trail in which the artifact is attested to be compliant
```

#### Specific Attestations

```yaml {.command}
artifacts:
  attestations:
    - name: "*" # attestation name can be anything
      type: pull-request
    - name: acceptance-test
      type: "*" # attestation type can be any built-in or existing custom type
    - name: security-scan
      type: snyk
    - name: coverage-metrics
      type: custom:my-coverage-metrics # custom attestation type
```

### Policy Rules Exceptions

You can add exceptions to policy rules using expressions.

```yaml
_schema: https://kosli.com/schemas/policy/environment/v1

artifacts
  provenance:
    required: true
    exceptions:
    # provenance is required except when one of the expressions evaluates to true
    - if: ${{ expression1 }}
    - if: ${{ expression2 }}

  trail-compliance:
    required: true
    exceptions:
    # trail-compliance is required except when one of the expressions evaluates to true
    - if: ${{ expression1 }}
    - if: ${{ expression2 }}

  attestations:
    - if: ${{ expression }} # this attestation is only required when expression evaluates to true
      name: unit-tests
      type: junit
```

#### Policy Expressions

Policy expressions allow you to create conditional rules using a simple and powerful syntax. Expressions are wrapped
in `${{ }}` and can be used in policy rules to create dynamic conditions. An expression consists of operands
and operators:

**Operators**

Expressions support these operators:

- Comparison: `==, !=, <, >, <=, >=`
- Logical: `and, or, not`
- List membership: `in`

**Operands**

Operands can be:

- Literal string
- List
- Context variable
- Function call

**Available Contexts**

Contexts are built-in objects which are accessible from an expression. Expressions can access two main contexts:

- `flow` - Information about the Kosli Flow:
  - `flow.name` - Name of the flow
  - `flow.tags` - Flow tags (accessed via flow.tags.tag_name)
- `artifact` - Information about the artifact:
  - `artifact.name` - Name of the artifact
  - `artifact.fingerprint` - SHA256 fingerprint

**Functions**

Functions are helpers that can be used when constructing conditions. They may or may not accept arguments. Arguments 
can be literals or context variables. Expressions can use following functions:

- `exists(arg)` : checks whether the value of arg is not None/Null
- `matches(input, regex)` : checks if input matches regex

**Example Expressions**

- ${{ exists(flow) }}
- ${{ flow.name in ["runner", 'saver', differ] }}
- ${{ matches(artifact.name, "^datadog:.*") }}
- ${{ flow.name == "runner" and matches(artifact.name, "^runner:.*") }}
- ${{ flow.tags.risk-level == "high" or matches(artifact.name, "^runner:.*") }}
- ${{ not flow.tags.risk-level == "high"}}
- ${{ flow.tags.risk-level != "high"}}
- ${{ flow.tags.key.with.dots == "value"}}
- ${{ flow.tags.risk-level >= 2 }}
- ${{ flow.name == 'prod' and (flow.tags.key_name == "value" or artifact.name == 'critical-service') }}
- ${{ flow.name == 'HIGH-RISK' and artifact.fingerprint == "37193ba1f3da2581e93ff1a9bba523241a7982a6c01dd311494b0aff6d349462" }}

## Attaching/Detaching Policies to/from Environments

Once you define your policies, you can attach them to environments via CLI or API:

```shell {.command}
kosli attach-policy prod-requirements --environment=aws-production
```

To detach a policy from an environment:

```shell {.command}
kosli detach-policy prod-requirements --environment=aws-production
```

Any attachment/detachment operation automatically triggers an evaluation of the latest environment snapshot and 
creates a new one with an updated compliance status.

{{% hint info %}}
If you detach all attached policies from an environment, the environment compliance state will become **unknown**
since there are no longer any defined requirements for artifacts running in it. The environment will continue to
track snapshots, but compliance cannot be evaluated without policies.
{{% /hint %}}


## Policy Enforcement Gates

Environment policies enable you to proactively block deploying a non-compliant artifact into an environment. This 
can be done as a deployment gate in your delivery pipeline or as an admission controller in your environment.

Regardless of where you place your policy enforcement gate, it will be using the `assert artifact` Kosli CLI command 
or its equivalent API call.

```shell {.command}
kosli assert artifact --fingerprint=$SHA256 --environment=aws-production
```

An artifact can also be asserted directly against one or more policies.
```shell {.command}
kosli assert artifact --fingerprint=$SHA256 --policy=has-approval,has-been-integration-tested
```
