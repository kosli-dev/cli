# The audit invents its input values, and 114 of its control runs fail

Every combination the audit measures is three runs: the flag omitted, the flag
set to a real value, and the flag empty. The first two are the controls the
third is compared against. For 114 of the 412 combinations on commands that work
here, the run with a real value **fails**, because the value was invented.

A flag the audit knows nothing about is given `probe-<flag>`. That is a working
value for a name and a poor one for anything else: `--expires-at
probe-expires-at` where a timestamp is wanted, `--visibility probe-visibility`
where the server accepts only `public` or `private`.

## Why it matters

The audit still answers its question when a control fails - it compares three
runs and reports what it saw - but the answer rests on less than it appears to.
When the empty run and the set run both fail, "the empty value behaves like a
real one" means only that both were refused, possibly for different reasons.

That is how `attest override --commit` hid. Its control run had been returning
500 since before this branch, the row read `accepted`, and nothing said the
comparison had only one working leg. It took a timing check to find it.

## The remedy that looked obvious, measured

The server publishes its own schema at `/api/v2/openapi.json`. It describes 202
fields, so deriving the audit's values from it rather than inventing them looks
like the fix.

It is not, and the numbers are worth recording so nobody re-proposes it:

| | |
|---|---|
| flags whose real value fails | 40 |
| of those, named by the schema | 9 |
| of those, with an enum | 1 |
| **failing rows the schema could fix** | **6** |

The six are `--expires-at` (2 rows), `--visibility`, `--start-ts`, `--end-ts`
and `--grace-period-hours`. Real, and worth taking as a by-product of any schema
work, but not a project.

## What actually causes the other 108

Not malformed values. Rules about how flags combine, and what a value means,
which live in the CLI and never reach a payload the server describes:

| Rows | Flags | Why the real value fails |
|---|---|---|
| 20 | `--registry-password`, `--registry-username` | they need a container registry and an image to read from, not a better string |
| 14 | `--external-fingerprint`, `--external-url` | the two are only accepted together |
| 7 | `--redact-commit-info` | takes the names of commit fields, not free text |
| 7 | `--repo-provider` | one of a fixed set the CLI knows and the schema does not |
| 7 | `--repo-url` | must agree with the repository the command is run in |
| 6 | `--annotate` | takes `key=value`, and `probe-annotate` is neither |

This is why 31 of the 40 flags are absent from the schema: the server never sees
them. An audit of the API first, which is where this question started, would not
have supplied these either.

## What would help

- **Hand-written values in `COMMAND_VALUES`**, which is the mechanism bootstrap
  already has for exactly this. Unglamorous, and it is where the ~60 rows in the
  table above live.
- **The schema, for fields rather than values.** What it knows that nothing else
  does is what each endpoint accepts: names, types, whether a field is required.
  That is what `replay.py` wants in order to empty a field, and it would let the
  probe ask about a field the CLI never sends at all - which no amount of driving
  the CLI can reach.

## A check is possible, and would fail loudly today

The audit records `set_exit` and `omitted_exit` in every row and nothing reads
them. A check on the control runs, of the kind that already exists for slow runs
and for exhausted retries, would name these 114 immediately. That is not a
reason to avoid it, but it is a reason to fix the values first: a check that
fails on 114 rows the day it lands teaches everyone to ignore it.

## How it was found

By `replay.py`, the API-side probe. It replays each captured request twice, once
unmodified as a control, and 35 of its controls came back 4xx. Asking why led to
the CLI audit's own controls, where the same defect is four times larger and had
been sitting in a column nobody read.
