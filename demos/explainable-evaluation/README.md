# Explainable policy evaluation — demo walkthrough

Seven demos that show what the experimental `kosli evaluate input --decision`
flag produces, from a single-batch bakery policy up to a Kosli-shaped trails
policy with iteration and alternatives.

## Setup

From the repo root:

```bash
make build              # produces ./kosli
```

Each demo is self-contained. Run them individually with the commands below,
or run them all in sequence with:

```bash
./demos/explainable-evaluation/run.sh
```

The runner pauses between demos. Hit return to advance.

## Reading the output

The `--decision` flag produces a single JSON document with the shape:

```
{
  schema_version, result, policy: {title, description},
  items: [
    { id?, result, checks: [
      { name, title, result, inputs_used, evaluated, alternatives_applied? }
    ]}
  ]
}
```

The whole document is intentionally one canonical artifact — what you'd record
as an attestation alongside the policy source.

---

## Demo 1 — bakery, passing

The simplest case. Single item, single-definition predicate rules, all checks
pass.

```bash
./kosli evaluate input \
  --input-file demos/explainable-evaluation/inputs/bakery-pass.json \
  --policy    demos/explainable-evaluation/policies/bakery.rego \
  --decision
```

**What to point at:**

- `policy.title` / `policy.description` come from the package-level `# METADATA`.
- Each check has a human title from its own `# METADATA`.
- `inputs_used` lists every `input.*` the rule body read, with the value at
  the time.
- `evaluated` shows the predicate with values substituted: `180 >= 175 and
  180 <= 200`. Same logic, but readable.

---

## Demo 2 — bakery, failing (short-circuit)

Same policy, with `temp_c = 165` (below the lower bound).

```bash
./kosli evaluate input \
  --input-file demos/explainable-evaluation/inputs/bakery-fail.json \
  --policy    demos/explainable-evaluation/policies/bakery.rego \
  --decision --no-assert
```

**What to point at:**

- `result: "deny"` at top level and on the failing check.
- `evaluated: "165 >= 175"` — *only* the predicate that ran. The upper bound
  was never evaluated (short-circuit), and the rendering reflects that.
- The passing `time_ok` check still shows both predicates because both ran.
- For comparison, the default output (no `--decision`) just prints
  `RESULT: DENIED` with no diagnostic; uncomment the second command in
  `run.sh` to see it side-by-side.

---

## Demo 3 — bakery, parameterised

The thresholds move from policy constants to operator-supplied parameters.

```bash
./kosli evaluate input \
  --input-file demos/explainable-evaluation/inputs/bakery-pass.json \
  --policy    demos/explainable-evaluation/policies/bakery-params.rego \
  --params    '{"min_temp_c": 175, "max_temp_c": 200}' \
  --decision
```

**What to point at:**

- `inputs_used` now contains entries like
  `"data.params.min_temp_c": {"value": 175, "source": "params"}` — the
  auditor can see exactly which parameter values the operator supplied at
  evaluation time.
- `evaluated` substitutes through the params the same way as inputs.

---

## Demo 4 — iteration

One annotated rule applied to every element of `input.batches`.

```bash
./kosli evaluate input \
  --input-file demos/explainable-evaluation/inputs/batches-mixed.json \
  --policy    demos/explainable-evaluation/policies/batches.rego \
  --decision --no-assert
```

**What to point at:**

- Three batches in, three `items` out — one per element.
- First two batches pass; the third fails (`temp_c: 150`). Top-level
  `result` reflects the conjunction.
- Each item carries its own `check` for `batch_ok` with its own pass/fail.
- The same JSON shape covers "one batch" (demo 1) and "many batches" (this
  one); only the cardinality changes.

---

## Demo 5 — alternatives, passing

A multi-definition rule: a PR is compliant either because it's bot-authored
OR because it has at least one approver. Input has an alice-authored PR with
bob as approver — the human branch fires.

```bash
./kosli evaluate input \
  --input-file demos/explainable-evaluation/inputs/pr-human-approved.json \
  --policy    demos/explainable-evaluation/policies/pr-approval.rego \
  --decision
```

**What to point at:**

- `alternatives_applied` lists both definitions, in source order.
- The failed alternative ("bot-authored PR") carries `reason: "\"alice\" ==
  \"bot\""` — the substituted predicate that ruled it out.
- The passing alternative has no `reason` field — it didn't need to defend
  itself.
- The Check itself carries `inputs_used` and `evaluated` *hoisted from the
  winning alternative* — one place to read "what did this check do?"

---

## Demo 6 — alternatives, all failing

Same policy, an alice-authored PR with no approvers. Neither alternative
applies.

```bash
./kosli evaluate input \
  --input-file demos/explainable-evaluation/inputs/pr-no-approval.json \
  --policy    demos/explainable-evaluation/policies/pr-approval.rego \
  --decision --no-assert
```

**What to point at:**

- `result: "fail"` at the Check level; no `inputs_used`/`evaluated` hoisted
  because there's no winning alternative.
- Both alternatives carry a `reason`. The auditor sees the full picture:
  - `"alice" == "bot"` (failed)
  - `count([]) > 0` (failed)
- No `violations` rule was written. The trace + annotations are doing the
  work that prose used to do.

---

## Demo 7 — Kosli-shaped: trails with per-trail alternatives

The capstone. Multiple trails in `input.trails`, each one evaluated by a
multi-definition `trail_compliant(trail)` rule. Mirrors the SCR-style policy
shape a real Kosli customer would write.

```bash
./kosli evaluate input \
  --input-file demos/explainable-evaluation/inputs/trails.json \
  --policy    demos/explainable-evaluation/policies/scr-trails.rego \
  --decision --no-assert
```

**What to point at:**

- One `item` per trail. Per-trail `result` is independent.
- Each item's check is `trail_compliant`, which itself is multi-def, so each
  item carries its own `alternatives_applied`.
- Trails where the bot-author branch matches show that alternative passing;
  trails where the human-approval branch matches show that one. Trails
  where neither matches show *both* failed alternatives with their reasons.
- This is the proposal's claim, end-to-end: one structured artifact, named
  controls, evidence per item, and per-alternative attribution — without
  the customer writing parallel violations prose.

---

## What's not in these demos

These limitations are real and worth flagging when discussing the prototype:

- **`evaluate trail` / `evaluate trails`** (live API) doesn't take `--decision`
  yet. Only `evaluate input` (local JSON) does.
- **Comprehensions inside `every`/`some`/aggregation** render via OPA's
  fallback string form, not value-substituted.
- **Iteration items don't get their own `inputs_used`/`evaluated`** when the
  per-item check is a function rule (its body references the parameter, not
  `input.*`).
- **`reason` for definitions OPA's indexer skipped entirely** falls back to
  the whole substituted body, not the specific failing predicate — OPA
  doesn't trace defs it considers redundant.

Everything else in the proposal lands.
