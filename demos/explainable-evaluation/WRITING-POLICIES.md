# Writing explainable policies — a guide

This guide is for the Customer Success team. It explains how to write Rego
policies that produce **explainable decisions** — JSON output where each
compliance check has a human name, the inputs it actually read, and the
rule as it ran with values substituted.

Read alongside the demos in this folder. Each pattern here corresponds to
one of the demo policies, so you can run an example before reaching for
your own data.

> **Status.** This is the experimental `--decision` flag on
> `kosli evaluate input`. The flag is opt-in and the JSON shape is
> versioned (`schema_version: "0.1.0"`). The default `kosli evaluate`
> behaviour is unchanged.

---

## 1. The two questions

A policy evaluation answers two questions:

1. **Is this compliant?** — `allow` / `deny`.
2. **How do you know?** — which checks ran, what data they read, and the
   rule the way it actually evaluated.

Today's `kosli evaluate` answers (1) well — it prints `RESULT: ALLOWED` or
`RESULT: DENIED` and any violation strings the policy emitted. Answering
(2) is largely the customer's problem: they look at the policy source,
they look at the input, and they reconstruct what happened.

`--decision` is a way of answering (2) automatically. You write Rego in a
particular way (named checks + a few annotations) and the tool emits a
structured JSON decision describing exactly what happened.

---

## 2. What the output looks like

For this policy:

```rego
# METADATA
# title: Bakery batch compliance
# description: A batch is compliant when temperature and time are in range.
package policy

import rego.v1

default allow := false

allow if {
	temp_ok
	time_ok
}

# METADATA
# title: Temperature in range
temp_ok if {
	input.bake.temp_c >= 175
	input.bake.temp_c <= 200
}

# METADATA
# title: Time in range
time_ok if {
	input.bake.minutes >= 25
	input.bake.minutes <= 40
}
```

…with input `{"bake": {"temp_c": 180, "minutes": 32}}`, the command:

```bash
kosli evaluate input --decision \
  --input-file bake.json \
  --policy    bakery.rego
```

…produces:

```json
{
  "schema_version": "0.1.0",
  "result": "allow",
  "policy": {
    "title": "Bakery batch compliance",
    "description": "A batch is compliant when temperature and time are in range."
  },
  "items": [{
    "result": "allow",
    "checks": [
      {
        "name": "temp_ok",
        "title": "Temperature in range",
        "result": "pass",
        "inputs_used": { "input.bake.temp_c": 180 },
        "evaluated": "180 >= 175 and 180 <= 200"
      },
      {
        "name": "time_ok",
        "title": "Time in range",
        "result": "pass",
        "inputs_used": { "input.bake.minutes": 32 },
        "evaluated": "32 >= 25 and 32 <= 40"
      }
    ]
  }]
}
```

Where each piece comes from:

| JSON field | Comes from |
|---|---|
| `schema_version` | Fixed by the CLI (currently `0.1.0`) |
| `result` | The boolean returned by `allow` |
| `policy.title` / `.description` | The package-level `# METADATA` block |
| `items[*].checks[*].name` | The Rego rule name |
| `items[*].checks[*].title` | That rule's `# METADATA / # title` |
| `items[*].checks[*].result` | Whether the rule evaluated to true |
| `inputs_used` | Every `input.*` / `data.params.*` the rule body actually read, with the resolved value |
| `evaluated` | The rule's predicates with values substituted in |

The `items` array is what makes the shape generic. A single-batch bakery
policy produces one item; an iterating policy (one rule applied to N
elements of an array) produces N items. The shape is the same.

---

## 3. Writing a policy — building it up

### 3.1 The minimum

A policy needs three things, same as today:

```rego
package policy

import rego.v1

default allow := false

allow if {
	# checks go here
	true
}
```

This passes any input. The decision is `{ "result": "allow", "items": [{...}] }`
with no checks listed.

### 3.2 Adding a named check

Pull a condition out into its own rule and annotate it:

```rego
package policy

import rego.v1

default allow := false

allow if { temp_ok }

# METADATA
# title: Temperature in range
temp_ok if { input.bake.temp_c >= 175 }
```

Now the decision has one check:

```json
"checks": [{
  "name": "temp_ok",
  "title": "Temperature in range",
  "result": "pass",
  "inputs_used": { "input.bake.temp_c": 180 },
  "evaluated": "180 >= 175"
}]
```

**The rule about checks.** A rule appears in the decision's `checks` list
if it's annotated with `# METADATA` and *reachable* from `allow`:

- Non-iterating policy: directly called from `allow`'s body.
- Iterating policy: the per-element rule called inside `every`.

Helpers that aren't annotated stay collapsed — they're the policy's
internals, not part of the audit story. Annotate the rules you want an
auditor to see; leave the rest alone.

### 3.3 Adding package metadata

Put a `# METADATA` block above the `package` declaration to give the
policy a name and description:

```rego
# METADATA
# title: Bakery batch compliance
# description: A batch is compliant when temperature and time are in range.
package policy
...
```

That fills in `policy.title` and `policy.description` in the decision.

### 3.4 Parameterising

When thresholds shouldn't be baked into the policy, move them to
`data.params`:

```rego
# METADATA
# title: Temperature in range
temp_ok if {
	input.bake.temp_c >= data.params.min_temp_c
	input.bake.temp_c <= data.params.max_temp_c
}
```

Pass values at evaluation time:

```bash
kosli evaluate input --decision \
  --input-file bake.json \
  --policy    bakery.rego \
  --params    '{"min_temp_c": 175, "max_temp_c": 200}'
```

`inputs_used` now wraps each param with its source:

```json
"inputs_used": {
  "input.bake.temp_c": 180,
  "data.params.min_temp_c": { "value": 175, "source": "params" },
  "data.params.max_temp_c": { "value": 200, "source": "params" }
}
```

The `source: "params"` tag tells an auditor that the operator supplied
this value at evaluation time — useful when policies have multiple
configurations across environments.

### 3.5 Iterating

When the input contains a list and the same check applies to each item,
use `every`:

```rego
# METADATA
# title: All batches compliant
package policy

import rego.v1

default allow := false

allow if {
	every batch in input.batches {
		batch_ok(batch)
	}
}

# METADATA
# title: Batch baked within range
batch_ok(batch) if {
	batch.temp_c >= 175
	batch.temp_c <= 200
}
```

Input:

```json
{ "batches": [
  { "id": "b-01", "temp_c": 180 },
  { "id": "b-02", "temp_c": 200 },
  { "id": "b-03", "temp_c": 150 }
] }
```

You get one item in the decision per element:

```json
"items": [
  { "result": "allow", "checks": [{ "name": "batch_ok", "result": "pass" }] },
  { "result": "allow", "checks": [{ "name": "batch_ok", "result": "pass" }] },
  { "result": "deny",  "checks": [{ "name": "batch_ok", "result": "fail" }] }
]
```

Top-level `result` is `"deny"` because at least one item failed.

**The pattern the tool recognises.** Iteration detection looks for a body
of the form `every <x> in input.<some.path> { <name>(<x>) }` — a single
function-rule call in the `every` body, with the iteration variable as
the argument. Other shapes work fine for Rego but won't be split into
per-element items.

### 3.6 Alternatives — when one of several paths can apply

Sometimes a control can be satisfied in more than one way. Source-code
review is the classic example: a commit is acceptable if it has
independent PR approval *or* if it was authored by a service account.

Rego expresses this naturally with multiple definitions of the same rule:

```rego
# METADATA
# title: Pull-request compliance
package policy

import rego.v1

default allow := false

allow if { pr_compliant }

# METADATA
# title: PR is compliant
# scope: document

# METADATA
# title: bot-authored PR
# scope: rule
pr_compliant if {
	input.pr.author == "bot"
}

# METADATA
# title: human-authored PR has an approver
# scope: rule
pr_compliant if {
	count(input.pr.approvers) > 0
}
```

Two things to notice here:

- **`# scope: document`** above the first definition gives the umbrella
  name for the rule. That becomes the check's `title`.
- **`# scope: rule`** on each definition names that specific alternative.

The decision tracks every alternative the engine tried and which one
fired. For input `{ "pr": { "author": "alice", "approvers": ["bob"] } }`:

```json
{
  "name": "pr_compliant",
  "title": "PR is compliant",
  "result": "pass",
  "inputs_used": { "input.pr.approvers": ["bob"] },
  "evaluated": "count([\"bob\"]) > 0",
  "alternatives_applied": [
    {
      "rule": "pr_compliant",
      "title": "bot-authored PR",
      "result": "fail",
      "reason": "\"alice\" == \"bot\""
    },
    {
      "rule": "pr_compliant",
      "title": "human-authored PR has an approver",
      "result": "pass"
    }
  ]
}
```

The auditor reads: *the PR is compliant; the bot-authored alternative
didn't apply because the author was "alice", not "bot"; the
human-authored alternative applied because there's at least one
approver.* No `violations` rules needed.

When neither alternative applies, both get `reason` strings showing
which predicate ruled them out:

```json
"alternatives_applied": [
  { "title": "bot-authored PR", "result": "fail",
    "reason": "\"alice\" == \"bot\"" },
  { "title": "human-authored PR has an approver", "result": "fail",
    "reason": "count([]) > 0" }
]
```

---

## 4. `# METADATA` reference

The tool reads three fields from `# METADATA` blocks:

| Field | Where to use it | Effect |
|---|---|---|
| `title` | Anywhere | Human name. Surfaced as `policy.title`, `checks[*].title`, or `alternatives_applied[*].title`. |
| `description` | Anywhere | Longer prose. Surfaced as `policy.description`. |
| `scope` | Multi-definition rules | `scope: document` for the umbrella; `scope: rule` on each definition. |

A `# METADATA` block attaches to **the next statement** in the file. If
you put a blank line between the `# METADATA` block and the thing it
should annotate, OPA will still attach it — but if you put another
statement between them, the annotation lands on the wrong thing.

```rego
# METADATA
# title: This attaches to temp_ok
temp_ok if { ... }
```

```rego
# METADATA
# title: This does NOT attach to temp_ok
default temp_ok := false
temp_ok if { ... }
```

For multi-definition rules, the convention is:

```rego
# METADATA
# title: Umbrella title for the whole rule
# scope: document

# METADATA
# title: First alternative
# scope: rule
my_rule if { ... }

# METADATA
# title: Second alternative
# scope: rule
my_rule if { ... }
```

The first `# METADATA` block declares the umbrella. OPA's compiler
applies `scope: document` annotations to *every* definition of a rule
with that name, so it doesn't matter which definition the block is
adjacent to — the umbrella title becomes the check title regardless of
which alternative fires.

---

## 5. Patterns for Kosli flow data

The demos use bakery-style data because it's small and self-contained.
Real Kosli policies typically work over a trail JSON. Two patterns worth
knowing:

### Single trail

When you capture a trail's data with `kosli evaluate trail … --show-input
--output json | jq '.input' > trail.json`, the resulting input has a
`trail` object at the root:

```json
{
  "trail": {
    "name": "<commit-sha>",
    "compliance_status": { ... },
    ...
  }
}
```

Reference it from a policy as `input.trail.<...>`:

```rego
allow if { pr_review_present }

# METADATA
# title: Trail has a PR review attestation
pr_review_present if {
    input.trail.compliance_status.attestations_statuses["pr-review"]
}
```

### Multiple trails

`kosli evaluate trails` produces an input with `trails` as an array.
Iterate the same way as section 3.5:

```rego
allow if {
    every trail in input.trails {
        trail_compliant(trail)
    }
}

# METADATA
# title: Trail has a compliant author
# scope: document

# METADATA
# title: service-account commit
# scope: rule
trail_compliant(trail) if {
    trail.author == "ci-bot"
}

# METADATA
# title: human commit with PR approval
# scope: rule
trail_compliant(trail) if {
    count(trail.pr.approvers) > 0
}
```

You get one decision item per trail, each carrying its own
`alternatives_applied`. See `demos/explainable-evaluation/policies/scr-trails.rego`
for the full demo.

---

## 6. Known limitations

These are real and worth knowing as you try the prototype:

- **Live API not wired yet.** `--decision` only works on `kosli evaluate
  input` (local JSON). `kosli evaluate trail` / `evaluate trails` (which
  fetch live from Kosli) still produce the old output. To experiment with
  trail data today, capture it with `--show-input --output json | jq
  '.input'`, save to a file, and run `evaluate input --decision` against
  that.

- **Iteration items lose `inputs_used` / `evaluated`** when the per-item
  check is a function rule (the usual shape — `trail_compliant(trail)`).
  The trace sees the body referencing `trail.x`, but `trail` is a
  parameter, not `input.*`, so it doesn't get substituted. The
  `alternatives_applied` per item still works correctly; only the
  Check-level evidence is missing on iterated items.

- **Comprehensions and inner `every`/`some` blocks** inside a check body
  aren't substituted in `evaluated`. Simple predicates (`a >= b`,
  `count(x) > 0`, `x == "foo"`) substitute and render in infix form;
  set/array/object comprehensions and nested `every` expressions fall
  through to OPA's literal Rego rendering instead.

- **`reason` for definitions the OPA indexer skipped.** When two
  definitions of the same rule could match and one succeeds, OPA's index
  often skips evaluating the other entirely. In that case we don't have a
  trace to pinpoint the specific failing predicate — we fall back to
  rendering the whole body with values substituted. Still readable, just
  not as precise.

- **No `# custom: { item_id: ... }` support yet.** Items in the decision
  don't have IDs. Adding one was an open question in the original
  proposal and is on the list of post-prototype work.

- **No `# METADATA` linting.** If you put a `# METADATA` block in the
  wrong place (typo, blank line in the wrong spot, attached to the wrong
  rule), the tool won't tell you. The annotation just doesn't appear in
  the output. A `kosli policy lint` command is on the roadmap.

---

## 7. What we'd like to learn

The reason this is going to CS first: we want to know whether the JSON
reads naturally to an auditor and to the engineers writing the policies.
Specific things we'd love to hear:

- **Did `# METADATA` placement ever surprise you?** Especially around
  multi-definition rules, where `scope: document` and `scope: rule` need
  to be in the right places.
- **What did you try to express that the prototype didn't handle?**
  Patterns that we should support but currently don't are the best
  feedback we can get.
- **Is the JSON shape what you'd want to record as an attestation?** Are
  any of the field names wrong, missing, or in the wrong place?
- **Did any of the limitations in section 6 hurt?** If yes, which one
  most, and on what kind of policy?
- **Would you rather write policies this way for new controls?** Versus
  the existing pattern with `violations` rules.

Send feedback to <internal channel TBD> or open an issue against the
branch. Bonus points for pasting a policy + input + the decision JSON
you got, even (or especially) when it surprised you.
