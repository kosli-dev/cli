# Bug: re-running create flow, begin trail or create policy wipes the description

These three commands are upserts. Running one a second time without
`--description` erases the description the object already had. No empty value is
involved and no variable needs to be unset: omitting the flag is enough.

Anyone driving these from CI, passing `--description` on some runs and not
others, is losing descriptions now.

## Reproducing it

Against a freshly reset local server:

```
$ kosli create flow wf --use-empty-template --description "the payments flow"
$ kosli get flow wf --output json | jq .description
"the payments flow"

$ kosli create flow wf --use-empty-template          # no --description
$ kosli get flow wf --output json | jq .description
""
```

The same for the other two:

| Command | after the first run | after a second run with no `--description` |
|---|---|---|
| `kosli create flow wf` | `the payments flow` | `""` |
| `kosli begin trail wt --flow wf` | `the release trail` | `""` |
| `kosli create policy wp policy.yml` | `the env policy` | `""` |

Every run exits 0 and prints what success prints.

## What is happening

The CLI puts `description` in the payload whether or not the flag was passed, so
an absent flag arrives at the server as an empty string, and the server writes
it. That also means `--description ""` and no `--description` produce the same
request: there is currently no way to say "leave the description alone".

## The fix, and the order it has to happen in

The CLI must send `description` only when the flag was actually passed. That is
one change in the CLI and a different amount of work per command on the server:

| Command | What the server needs | If the CLI stopped sending it today |
|---|---|---|
| `begin trail` | nothing - the update is already guarded by `if "description" in payload` | already correct |
| `create policy` | make `description` optional, and skip it when absent | accepted, but the description is still cleared |
| `create flow` | make `description` optional, and skip it when absent | rejected with a 422 - the field is required |

**The server change has to land before the CLI stops sending the field**, or
every `kosli create flow` fails with a 422. `begin trail` is the pattern the
other two need, already written.

## Why it matters beyond the lost text

A description is not compliance data, so this does not change a verdict. It does
mean the CLI has no way to express "leave this alone", which is the same gap that
makes an empty `--description` indistinguishable from an absent one. Fixing this
is a precondition for refusing empty values on those flags - see
`2026-08-13-empty-value-decision.md`.

## How it was found

While auditing what the CLI does with empty flag values
(`hack/empty-flag-audit`). This one turned up because an empty `--description`
and an absent `--description` produced identical results, which is only possible
if the absent case is also sending something.
