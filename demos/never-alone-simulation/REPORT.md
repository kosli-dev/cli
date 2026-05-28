# Decision output diagnostics — findings and proposed improvements

This report documents what the `--decision` flag currently produces when
run against the never-alone simulation trails, why certain information is
missing, and what changes would fix it.

---

## What the output looks like today

Running:

```bash
./kosli evaluate input \
  --policy four-eyes.rego \
  --input-file trails/v2.11.44-fail.json \
  --decision
```

produces (abbreviated — actual output is the full JSON below):

```json
{
  "schema_version": "0.1.0",
  "result": "deny",
  "policy": { "title": "Four-eyes principle", "description": "..." },
  "items": [
    {
      "result": "allow",
      "checks": [{ "name": "trail_compliant", "result": "pass",
        "alternatives_applied": [
          { "rule": "trail_compliant", "title": "service-account commit — exempt from PR review", "result": "fail", "reason": "is_service_account(trail)" },
          { "rule": "trail_compliant", "title": "human commit with independent PR approval",      "result": "pass" }
        ]
      }]
    },
    {
      "result": "allow",
      "checks": [{ "name": "trail_compliant", "result": "pass",
        "alternatives_applied": [
          { "rule": "trail_compliant", "title": "service-account commit — exempt from PR review", "result": "pass" },
          { "rule": "trail_compliant", "title": "human commit with independent PR approval",      "result": "fail",
            "reason": "not is_service_account(trail) and assign(attest, pr_attest(trail)) and some pr in attest.pull_requests and all_authors_resolved(pr) and has_independent_approval(trail, pr)" }
        ]
      }]
    },
    { "result": "allow", "checks": [{ "name": "trail_compliant", "result": "pass", "alternatives_applied": ["..."] }] },
    { "result": "allow", "checks": [{ "name": "trail_compliant", "result": "pass", "alternatives_applied": ["..."] }] },
    {
      "result": "deny",
      "checks": [{ "name": "trail_compliant", "result": "fail",
        "alternatives_applied": [
          { "rule": "trail_compliant", "title": "service-account commit — exempt from PR review", "result": "fail", "reason": "is_service_account(trail)" },
          { "rule": "trail_compliant", "title": "human commit with independent PR approval",      "result": "fail", "reason": "has_independent_approval(trail, pr)" }
        ]
      }]
    }
  ]
}
```

The input `v2.11.44-fail.json` has five trails in order:

| # | Commit (short) | Author | Expected result |
| --- | --- | --- | --- |
| 1 | `96368697` | Dan Grøndahl | pass — human, approved |
| 2 | `6589e8bb` | ci-signed-commit-bot[bot] | pass — service account |
| 3 | `5959d21d` | Faye | pass — human, approved |
| 4 | `ee748df6` | Tore Martin Hagen | pass — human, approved |
| 5 | `81a88f9d` | Tore Martin Hagen | **deny — no independent approval** |

**The problem:** there is no identifier on any item. You can only tell
which trail failed by counting items and cross-referencing the input
file — the fifth item is the deny, so you look up the fifth trail. With
a longer list this becomes error-prone, and in a recorded attestation the
link is lost entirely.

---

## Root causes

### 1. `Item.ID` is never populated — CLI gap

The `Item` struct in `internal/evaluate/decision.go` has an `id` field:

```go
type Item struct {
    ID     string  `json:"id,omitempty"`   // always empty today
    Result string  `json:"result"`
    Checks []Check `json:"checks"`
}
```

The `evaluateIteration()` function already holds each array element as
`elem map[string]interface{}` at the point where `Item` is constructed,
but never reads a name or id from it. This was explicitly deferred in
the original implementation ("no item id extraction yet — follow-up once
the schema settles").

**Fix:** in `evaluateIteration()`, try `elem["name"]` then `elem["id"]`
before constructing the `Item`. Trails always carry a `name` field (the
commit SHA), so this would immediately produce:

```json
{ "id": "a7573bcb0efbb25949ac04826b3991d860293e9a", "result": "deny", ... }
```

This is a small, self-contained change to `internal/evaluate/decision.go`.

---

### 2. `inputs_used` and `evaluated` are empty per item — function-rule trace limitation

The policy evaluates trails via:

```rego
every trail in input.trails {
    trail_compliant(trail)
}
```

Because `trail_compliant(trail)` is a **function rule** (takes `trail`
as a parameter), OPA's trace records predicate outcomes relative to the
function argument, not to `input.*` paths. The `--decision` tracer only
substitutes values for paths that begin with `input.` or `data.params.`,
so `inputs_used` and `evaluated` end up empty at the check level on every
item.

`alternatives_applied` is unaffected — the `result` and `reason` fields
on each alternative still work. However, because `trail` is a function
parameter (not an `input.*` path), OPA's trace can't substitute its
value either. The `reason` strings show the predicate form with the
parameter name left in, not the actual commit data:

- `"reason": "is_service_account(trail)"` — not the author string
- `"reason": "has_independent_approval(trail, pr)"` — not the approver list

**This limitation cannot be resolved in Rego alone.** It is a known
constraint of the current prototype, documented in
`demos/explainable-evaluation/WRITING-POLICIES.md` section 6.

---

## Proposed changes

### Change 1 — CLI: populate `Item.ID` from the iteration element *(high value)*

**File:** `internal/evaluate/decision.go`

In `evaluateIteration()`, at the point where `Item` is constructed,
extract an identifier from the element:

```go
func itemID(elem map[string]interface{}) string {
    for _, key := range []string{"name", "id"} {
        if v, ok := elem[key]; ok {
            if s, ok := v.(string); ok && s != "" {
                return s
            }
        }
    }
    return ""
}
```

Then:

```go
items = append(items, Item{
    ID:     itemID(elem),
    Result: itemResult,
    Checks: checks,
})
```

`name` is tried first because Kosli trail objects use `name` for the
commit SHA. `id` is the fallback for other array shapes (e.g. the bakery
demo's `{ "id": "batch-01", ... }`). When neither is present the field
is omitted (existing `omitempty` behaviour is preserved).

After this change, each item in the never-alone output would be labelled
by its commit SHA, making the decision fully traceable.

### Change 2 — no Rego-only fix for `reason` legibility

A Rego restructure was considered to make the `reason` strings more
readable. However, because `trail` is a function parameter throughout
the entire `trail_compliant` call chain, OPA's value substitution does
not apply regardless of how the predicates are written. Inlining
`is_service_account`'s body, for example, would change the reason from:

```text
"is_service_account(trail)"
```

to:

```text
"regex.match(\"svc_.*\", trail.git_commit_info.author)"
```

The parameter reference `trail.git_commit_info.author` still appears
unsubstituted — no improvement. The only fix is the CLI change (item 1)
which gives each item an `id`, making the parameter names irrelevant for
identification.

---

## What the output looks like after Change 1 (CLI)

The `reason` strings are unchanged — function parameters stay
unsubstituted. What changes is that every item gains an `id`, making the
decision self-contained and directly traceable to a commit.

```json
{
  "result": "deny",
  "policy": { "title": "Four-eyes principle", "description": "..." },
  "items": [
    {
      "id": "96368697cb7cd4c2f8c45be29600dd6e759337e8",
      "result": "allow",
      "checks": [{ "name": "trail_compliant", "result": "pass",
        "alternatives_applied": [
          { "title": "service-account commit — exempt from PR review", "result": "fail", "reason": "is_service_account(trail)" },
          { "title": "human commit with independent PR approval",      "result": "pass" }
        ]
      }]
    },
    {
      "id": "6589e8bb6159d6a6bc4ee6ceda3b46edbde0f9fe",
      "result": "allow",
      "checks": [{ "name": "trail_compliant", "result": "pass",
        "alternatives_applied": [
          { "title": "service-account commit — exempt from PR review", "result": "pass" },
          { "title": "human commit with independent PR approval",      "result": "fail",
            "reason": "not is_service_account(trail) and assign(attest, pr_attest(trail)) and some pr in attest.pull_requests and all_authors_resolved(pr) and has_independent_approval(trail, pr)" }
        ]
      }]
    },
    { "id": "5959d21ddffecbb3cc3c06a0f4431aec3398107c", "result": "allow", "checks": ["..."] },
    { "id": "ee748df6acd8e75c241c43b6f05cee336c9f2c47", "result": "allow", "checks": ["..."] },
    {
      "id": "81a88f9d3fc855a93bdb99ec7682e855cc719d5c",
      "result": "deny",
      "checks": [{ "name": "trail_compliant", "result": "fail",
        "alternatives_applied": [
          { "title": "service-account commit — exempt from PR review", "result": "fail", "reason": "is_service_account(trail)" },
          { "title": "human commit with independent PR approval",      "result": "fail", "reason": "has_independent_approval(trail, pr)" }
        ]
      }]
    }
  ]
}
```

An auditor can now read: commit `81a88f9d` (Tore Martin Hagen) failed
four-eyes — it is not a service account, and the human-approval check
failed at `has_independent_approval`. The other four commits all pass:
`6589e8bb` was approved via the service-account exemption; the other
three passed through the human-approval path.

---

## Limitations that remain after Change 1

- `inputs_used` and `evaluated` at the Check level stay empty for
  function-rule iteration. This is a prototype constraint; addressing it
  would require the tracer to follow function-argument bindings into the
  OPA evaluation tree, which is a larger change.
- The `reason` strings remain function-parameter references (`is_service_account(trail)`,
  `has_independent_approval(trail, pr)`) rather than substituted values.
  They identify *which predicate failed* but not the underlying data values.
  This too requires a tracer-level fix in the CLI.
- `--decision` only works with `kosli evaluate input` (local JSON). Live
  `kosli evaluate trail(s)` calls still produce the old output.
