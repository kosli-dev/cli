# Bug: the CLI and the server disagree about what an absent flag means

`kosli create flow` and `kosli begin trail` are both create-or-update, and both
describe themselves that way. They do not agree on what happens to a field you
did not mention.

- **`create flow` uses the apply model.** "Create or update a Kosli flow. You can
  specify flow parameters in flags." What you pass is what the flow becomes, so
  an absent `--description` meaning "no description" is that model working. There
  is nothing to fix in it.
- **`begin trail` says the same thing and behaves differently.** Its update on
  the server is guarded by `if "description" in payload`, so an absent
  description leaves the stored one alone - the patch model. The CLI never lets
  that guard fire, because it puts `description` in the payload whether or not
  the flag was passed.

So a trail's description is replaced by an empty one by a command that never
mentioned it, and the server was written to prevent exactly that.

(The guard was read in the server repo earlier in this work. It cannot be checked
from this repo, and it cannot be checked through the CLI either, because the CLI
always sends the field.)

## The fields it applies to

Running each command twice and reading the object back afterwards:

| Command and flag | after the first run | after a second run without the flag |
|---|---|---|
| `begin trail --description` | `the release trail` | `""` |
| `begin trail --user-data` | the document | `{}` |
| `attest generic --user-data` | the document | `{}` |
| `attest custom --user-data` | the document | `{}` |
| `attest junit --user-data` | the document | `{}` |
| `attest decision --user-data` | the document | `{}` |
| `attest override --user-data` | the document | `{}` |
| `create flow --description` | `the payments flow` | `""` |
| `create policy --description` | `the env policy` | `""` |
| `create policy --comment` | `the review note` | `""` |

The last three are the apply model doing its job, and are listed only so the set
is complete. The rest are the disagreement.

`create environment` sends `"description": ""` in the same way and the
description survives, so what the CLI sends does not settle it on its own - the
server decides. Only running each one shows which is which.

## How much is at stake

Less than "data loss" would suggest. A trail keeps its earlier value in its
`events[]`, each carrying the `trail_data_json` that was reported, and an
attestation's earlier value stays on the attestation event that carried it. What
changes is the current view, not the record of what was reported.

Nor does it reach backwards. A trail snapshots the flow state it was created
with: changing a flow's template afterwards leaves existing trails requiring
exactly what they always did. State is snapshotted, not shared.

One thing is unmeasured. A policy can read `user_data`, so a trail or attestation
whose `user_data` has become `{}` may evaluate differently from one that still
carries it. Whether any policy does read it in practice has not been tested here.

## The fix

Decide, per command, which model it uses, and make both layers say the same
thing:

- **apply** - the CLI sends every field, and the server writes what arrives.
  `create flow` is already this, on both sides.
- **patch** - the CLI sends a field only when its flag was passed, and the server
  leaves absent fields alone. `begin trail`'s server side is already this; the
  CLI is not.

Whichever is chosen for `begin trail`, the two sides have to be changed together.
If the CLI stops sending `description` while the server still requires it,
`kosli create flow` fails with a 422; if the server starts ignoring an absent
description while the CLI still sends `""`, nothing changes at all.

## Why it belongs with the empty-value work

Until a command says what an absent flag means, it cannot say what an empty one
means either: today `--description ""` and no `--description` produce the same
request, so the two cannot be told apart. See
`2026-08-13-empty-value-decision.md`.

## How it was found

While auditing what the CLI does with empty flag values
(`hack/empty-flag-audit`). An empty `--description` and an absent `--description`
produced identical results, which is only possible if the absent case is sending
something too. Running every write command with `--debug` and none of its
optional flags showed which fields arrive empty when nobody asked for them.
