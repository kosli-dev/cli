---
title: "Evaluate trails with OPA policies"
bookCollapseSection: false
weight: 509
draft: true
summary: "Learn how to use kosli evaluate trail and kosli evaluate trails to check your Kosli trails against custom OPA/Rego policies. This tutorial walks through writing a policy that verifies pull requests have been approved."
---

# Evaluate trails with OPA policies

The `kosli evaluate` commands let you evaluate Kosli trails against custom policies written in [Rego](https://www.openpolicyagent.org/docs/latest/policy-language/), the policy language used by the [Open Policy Agent (OPA)](https://www.openpolicyagent.org/) project. This is useful for enforcing rules like "every artifact must have an approved pull request" or "all security scans must pass" — and for gating deployments in CI/CD pipelines based on those rules.

In this tutorial, we'll write a policy that checks whether pull requests on a trail have been approved, then evaluate it against real trails in public Kosli orgs.

## Step 1: Prerequisites

To follow this tutorial, you need to:

* [Install Kosli CLI](/getting_started/install/).
* [Get a Kosli API token](/getting_started/service-accounts/).
* Set the `KOSLI_API_TOKEN` environment variable to your token:
  ```shell {.command}
  export KOSLI_API_TOKEN=<your-api-token>
  ```

{{<hint info>}}
You don't need OPA installed — the Kosli CLI has a built-in Rego evaluator. You just need to write a `.rego` policy file.
{{</hint>}}

## Step 2: Write a policy

Create a file called `pr-approved.rego` with the following content:

```rego
package policy

import rego.v1

default allow = false

violations contains msg if {
    some trail in input.trails
    some pr in trail.compliance_status.attestations_statuses["pull-request"].pull_requests
    count(pr.approvers) == 0
    msg := sprintf("trail '%v': pull-request %v has no approvers", [trail.name, pr.url])
}

allow if {
    count(violations) == 0
}
```

Let's break down what this policy does:

- **`package policy`** — every evaluate policy must use the `policy` package.
- **`import rego.v1`** — use Rego v1 syntax (the `if`/`contains` keywords).
- **`default allow = false`** — trails are denied unless explicitly allowed.
- **`violations`** — a set of messages describing why the policy failed. The rule iterates over trails, then over pull requests within the `pull-request` attestation, looking for PRs where `approvers` is empty.
- **`allow`** — trails are allowed only when there are no violations.

{{<hint info>}}
The policy contract requires `package policy` and an `allow` rule. The `violations` rule is optional but recommended — it provides human-readable reasons when a trail is denied.
{{</hint>}}

## Step 3: Evaluate multiple trails

Let's evaluate several trails from the public `cyber-dojo` org against our policy. The `kosli evaluate trails` command fetches trail data from Kosli and passes it to the policy as `input.trails`:

```shell {.command}
kosli evaluate trails \
  --policy pr-approved.rego \
  --org cyber-dojo \
  --flow dashboard-ci \
  9978a1ca82c273a68afaa85fc37dd60d1e394f84 \
  b334d371eb85c9a5c811776de1b65fb80b52d952 \
  5abd63aa1d64af7be5b5900af974dc73ae425bd6 \
  cb3ec71f5ce1103779009abaf4e8f8a3ed97d813
```

The cyber-dojo project doesn't require PR approvals, so you'll see violations:

```plaintext {.light-console}
RESULT:      DENIED
VIOLATIONS:  trail '5abd63aa1d64af7be5b5900af974dc73ae425bd6': pull-request https://github.com/cyber-dojo/dashboard/pull/342 has no approvers
             trail '9978a1ca82c273a68afaa85fc37dd60d1e394f84': pull-request https://github.com/cyber-dojo/dashboard/pull/344 has no approvers
             trail 'b334d371eb85c9a5c811776de1b65fb80b52d952': pull-request https://github.com/cyber-dojo/dashboard/pull/343 has no approvers
             trail 'cb3ec71f5ce1103779009abaf4e8f8a3ed97d813': pull-request https://github.com/cyber-dojo/dashboard/pull/341 has no approvers
```

Now try the `kosli-public` org, where PRs do have approvers:

```shell {.command}
kosli evaluate trails \
  --policy pr-approved.rego \
  --org kosli-public \
  --flow cli \
  5a0f3c0 \
  167ed93 \
  030cc31
```

```plaintext {.light-console}
RESULT:  ALLOWED
```

## Step 4: Evaluate a single trail

To evaluate just one trail, use `kosli evaluate trail` (singular). The data is passed to the policy as `input.trail` instead of `input.trails`, so you need a slightly different policy. Save this as `pr-approved-single.rego`:

```rego
package policy

import rego.v1

default allow = false

violations contains msg if {
    some pr in input.trail.compliance_status.attestations_statuses["pull-request"].pull_requests
    count(pr.approvers) == 0
    msg := sprintf("trail '%v': pull-request %v has no approvers", [input.trail.name, pr.url])
}

allow if {
    count(violations) == 0
}
```

```shell {.command}
kosli evaluate trail \
  --policy pr-approved-single.rego \
  --org cyber-dojo \
  --flow dashboard-ci \
  9978a1ca82c273a68afaa85fc37dd60d1e394f84
```

```plaintext {.light-console}
RESULT:      DENIED
VIOLATIONS:  trail '9978a1ca82c273a68afaa85fc37dd60d1e394f84': pull-request https://github.com/cyber-dojo/dashboard/pull/344 has no approvers
```

{{<hint info>}}
When writing a policy for `kosli evaluate trail`, reference `input.trail` (a single object) instead of `input.trails` (an array). You can write one policy that handles both by checking for both keys, or keep separate policies for each command.
{{</hint>}}

## Step 5: Explore the policy input with --show-input

When writing policies, it helps to see exactly what data is available. Use `--show-input` combined with `--output json` to see the full input that gets passed to the policy:

```shell {.command}
kosli evaluate trail \
  --policy pr-approved-single.rego \
  --org cyber-dojo \
  --flow dashboard-ci \
  --show-input \
  --output json \
  9978a1ca82c273a68afaa85fc37dd60d1e394f84
```

This outputs the evaluation result along with the complete `input` object. You can pipe it through `jq` to explore the structure:

```shell {.command}
kosli evaluate trail \
  --policy pr-approved-single.rego \
  --org cyber-dojo \
  --flow dashboard-ci \
  --show-input \
  --output json \
  9978a1ca82c273a68afaa85fc37dd60d1e394f84 2>/dev/null | jq '.input.trail.compliance_status | keys'
```

```plaintext {.light-console}
[
  "artifacts_statuses",
  "attestations_statuses",
  "evaluated_at",
  "flow_template_id",
  "is_compliant",
  "status"
]
```

{{<hint info>}}
Use the `--attestations` flag to limit which attestations are enriched with full detail. The flag filters by **attestation name** (not type). For example, `--attestations pull-request` fetches only details for attestations named `pull-request`, which speeds up evaluation and reduces noise when exploring the input.
{{</hint>}}

## Step 6: Use in CI/CD

The `kosli evaluate` commands exit with code 0 when the policy allows and code 1 when it denies. This makes them straightforward to use as gates in CI/CD pipelines:

```shell {.command}
# Example: gate a deployment on policy evaluation
if kosli evaluate trail \
  --policy policies/pr-approved.rego \
  --org "$KOSLI_ORG" \
  --flow "$FLOW_NAME" \
  "$GIT_COMMIT"; then
  echo "Policy passed — proceeding with deployment"
  # ... deploy commands ...
else
  echo "Policy denied — blocking deployment"
  exit 1
fi
```

This pattern lets you enforce custom compliance rules as part of your delivery pipeline, using the same trail data that Kosli already collects.
