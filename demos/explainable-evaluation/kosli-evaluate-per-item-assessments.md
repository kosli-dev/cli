# Proposal: a third kosli evaluate output carrying per-item assessments

This describes a change to `kosli evaluate` that does not exist. It is written
down because the gap it would close is general, and because this repo's parity
test currently pays the interest on it.

"Item" here means whatever a policy judges: the elements of the collection its
input carries. In this repo those are Snyk vulns, but they could equally be
tests, commits, licences or deployments.

## The general problem: output granularity does not match input granularity

A policy is given a collection and emits two things:

- `allow`, one boolean for the whole input
- `violations`, a set of messages, populated only for the elements that failed

So an element that passed contributes nothing to the output, and an element that
failed contributes prose. Any caller that needs to know per-item outcomes, or how
close a passing item is to failing, has to reconstruct that from the input plus
whatever it can parse back out of the messages.

That reconstruction is the duplication. It is not specific to vulns: it appears
whenever a policy's caller wants a report, a dashboard, a forecast or a label per
item rather than a single verdict.

## What kosli evaluate emits today

Two keys, selected by name rather than by surfacing the whole policy document.

`package policy` in this repo's `snyk-vuln-compliance.rego` defines four things
at the top level: `max_days_by_severity`, `seconds_per_day`, `allow` and
`violations`. Only the last two appear:

    $ echo '{"vulns":[]}' | kosli evaluate input \
        --policy snyk-vuln-compliance.rego \
        --params @rego.params.aws-beta.json \
        --output json
    {
      "allow": true,
      "violations": null
    }

If the document were being surfaced wholesale, `seconds_per_day` and the severity
limits would be in that JSON. They are not. This was established by running the
command, not from Kosli documentation, which has not been consulted.

Rego itself has no notion of either name. `allow` and `violations` are ordinary
rule names; a policy could call them anything. The two-key shape is Kosli's
contract, not the language's.

## The proposal: look up one more rule name

The smallest possible change, and the most general: `kosli evaluate` looks up a
third top-level rule, say `assessments`, and includes it in the output when the
policy defines one. Policies that do not define it are unaffected.

Kosli fixes only the name. The policy decides the contents, exactly as it already
decides its violation strings. That keeps the feature useful to policies about
things nobody has thought of yet.

## A recommended convention for the contents

Worth recommending rather than enforcing, so that tooling can rely on a shape
without the feature being narrowed to one domain:

    {
      "allow": false,
      "violations": ["<id>: <human readable reason>"],
      "assessments": {
        "<item id>": {
          "compliant": false,
          "mechanism": "<name of the rule that decided>",
          "<policy specific facts>": "..."
        }
      }
    }

- keyed by the same id the violation messages lead with, so the two outputs join
- `compliant` gives the per-item outcome that callers currently reconstruct
- `mechanism` names which rule decided, so callers stop reimplementing the
  taxonomy in order to label things
- everything else is the policy's own business: the facts it measured, in
  whatever units it measured them

An object rather than an array makes duplicate ids inexpressible and turns a
caller's join into a lookup.

## The three outputs differ in kind

- `allow` is the decision
- `violations` is why something was refused
- `assessments` is what was assessed, for every item, pass or fail

The third is the only one whose granularity matches the input's. That is what
makes it able to describe an item that passed.

## Worked example: this repo

Two scripts here exist because the third output is missing.

`bin/vuln_annotations.py` reconstructs the per-item view by subtracting the ids
named in violations from the vuln list. It carries a reconciliation guard that
exits non-zero when a denial names no vuln, or names one outside the list,
because that reconstruction can otherwise present a failing vuln as passing.
With `assessments` it would read `compliant`.

`bin/find_expiring_vulns.py` recomputes the age, the severity limit and the
distance to the boundary, because the report answers a question the policy output
cannot: how many days remain. `tests/test_rego_report_verdict_parity.sh` holds
the two implementations to each other, which treats the symptom rather than
removing the second implementation.

Some of those numbers already leave the policy, lossily: the Case 1 violation
message is built from `floor(age_days(vuln))` and
`max_days_by_severity[vuln.severity]`, so the age and the limit are exported
floored and embedded in an English sentence. A structured third output is the
principled version of what that message text does accidentally.

Filled in for this policy, an assessment would carry `age_days`, `limit_days`,
`days_remaining`, `first_seen_at` and `measured_at` alongside `compliant` and
`mechanism`. `limit_days` matters more than it looks: the report reads
`rego.params.<env>.json` itself today, so publishing the limit the evaluation
actually applied removes any chance of the two reading different params.

Absent by design: `severity`, `vuln_url`, `artifact`. Those live in the vuln file
the report already reads, and echoing inputs invites them to disagree.

The mechanism vocabulary would be finer than the report's. This policy
distinguishes five cases (age limit, active ignore, forever ignore, expired
ignore, unmeasurable age) where the report collapses the three ignore cases into
`dot_snyk_expiry` and drops forever-ignores entirely. Adopting `assessments`
means deciding whether to map fine to coarse for display or to let the summary
get more specific.

| Consumer | Today | With assessments |
|---|---|---|
| `vuln_annotations.py` | subtracts violation ids from the vuln list, plus a reconciliation guard | reads `compliant` |
| `find_expiring_vulns.py` | recomputes age, limit, distance, mechanism | merges and sorts |
| `print_slack_status_symbol.py` | re-decides via `days_remaining <= 0` | reads `compliant` |
| `print_expiring_vulns_summary.py` | renders | renders, unchanged |

`print_slack_status_symbol.py` is worth calling out. It decides compliance by
testing a number against zero, which is why the exactly-zero boundary needs
pinning in two places: its own test, and the parity test's exactly-at-limit
scenario. Reading the policy's own per-item outcome would remove that hazard
rather than guarding it.

## In Rego

A partial object rule, with one definition per case:

    assessments[item.id] := assessment(item) if some item in input.items

    assessment(item) := {...} if <first case>
    assessment(item) := {...} if <second case>

The definitions must be mutually exclusive, which is a bonus rather than a
burden: if two ever matched and produced different objects, Rego raises a
conflict instead of silently picking one. Overlapping cases become an error.

`allow` could then be defined over `assessments`, with `every` requiring
`compliant`. That keeps one source for the decision and stays a positive
assertion, so a missing param leaves `allow` false rather than vacuously true.
See `docs/rego-default-allow-false-safety.md`.

## Constraints on adopting it

**Assessments are measurements, never the verdict.** `allow` and `violations`
stay the decision. If a caller starts deriving compliance from an assessment's
numbers instead of from its `compliant` field or from `allow`, the second
implementation has been reintroduced somewhere new.

**The win only lands if the caller stops computing.** A caller that keeps its own
arithmetic as a fallback for when assessments are absent has three
implementations instead of two. It has to delete the arithmetic and fail loudly
when the facts are missing.

**Output size scales with the input.** The document grows with the number of
items rather than the number of failures, on something attached to every
attestation. A policy over many passing items produces a much larger evaluation
document than today.

**Granularity seams in the caller.** In this repo the policy evaluates one
artifact at a time while the report spans every artifact in an environment, so
assessments would have to be persisted and re-associated with the per-item files
the notify job downloads. Any caller whose reporting scope is wider than its
evaluation scope has the same plumbing to do.

**Timing is not a constraint here**, which is worth stating because it looks like
one. This repo's report measures against the `now_ts` stamped into each vuln file
rather than a clock of its own, so the policy's numbers and the report's are for
the same instant.

## Effect on the parity test

`tests/test_rego_report_verdict_parity.sh` would change character rather than go
away. With one implementation there are no two verdicts to compare, but the test
still has a job: confirming that the report's rendering of the policy's facts
does not contradict `allow` and `violations`.
