# Bug: the CLI and the server disagree about what an absent flag means

`kosli create flow` and `kosli begin trail` are both create-or-update, and both
describe themselves that way. They do not agree on what happens to a field you
did not mention.

- **`create flow` uses the apply model.** "Create or update a Kosli flow. You can
  specify flow parameters in flags." It reads its flags the way `kubectl apply`
  reads a manifest: what you pass is what the flow becomes, so an absent
  `--description` meaning "no description" is that model working. There is
  nothing to fix in it.
- **`begin trail` says the same thing and behaves differently.** Its update on
  the server is guarded by `if "description" in payload`, so an absent
  description leaves the stored one alone - the patch model. The CLI never lets
  that guard fire, because it puts `description` in the payload whether or not
  the flag was passed.

So a trail's description is replaced by an empty one by a command that never
mentioned it, and the server was written to prevent exactly that.

Which of the two is right is the decision here, and it is worth making
deliberately rather than leaving one layer to win by accident. What is not in
doubt is that both layers should not be answering it differently.

(The guard was read in the server repo earlier in this work. It cannot be checked
from this repo, and it cannot be checked through the CLI either, because the CLI
always sends the field.)

## What the CLI does today

Running each command twice and reading the object back afterwards. This is the
current behaviour, not a list of ten defects - under the apply model most of it
is correct, and the point of the table is to show how much rides on the choice:

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

The `create` rows are the apply model doing its job. The `begin trail` and
`attest` rows are the ones the choice decides: keep apply and they are correct
too, adopt patch and every one of them is wrong today.

`create environment` sends `"description": ""` in the same way and the
description survives, so what the CLI sends does not settle it on its own - the
server decides. Only running each one shows which is which.

## The server half, with line numbers (2026-08-19)

Read in the server repo, which is what lines 25-27 above could not do. The three
`create` commands are three different shapes, and only one of them behaves the
way this document assumed.

**`create flow`.** `description: str` is required (`models/flows.py:17`), and the
handler passes it straight to `flow.update` with no presence guard
(`common/flow.py:155`). So a re-run empties the stored description. (`visibility`
on the same endpoint is `None`-tolerant, handled at `common/flow.py:147-149` and
`:164`. Recorded as a fact about the code, not as a decision about description.)

**`create environment`.** Required twice: `description: str`
(`models/environments.py:22`), and a second gate below it raising `BadRequest`
when the key is absent (`model/environments.py:179`). But the update path drops a
falsy description before applying it (`_upsert_data`,
`model/environments.py:266-275`) and writes only `if "description" in data`
(`_apply_update`, `:282`). That is why the description survives, as line 52 above
records. Clearing is done through the PATCH endpoint, whose docstring names the
contrast (`v2/environment.py:87-89`).

**`create policy`.** `description: str = Field("")` (`models/policies.py:25`), so
an absent description has already become `""` by the time the handler sees it,
and the update writes the field whenever it differs from the stored value
(`model/policies/commands.py:180-181`). `comment` is declared the same way
(`models/policies.py:27`). Adding `omitempty` to the CLI changes nothing here on
its own: the model default has to become `None` as well.

## Two questions the line numbers separate

They were one question while the server was unread, and answering only the first
looks like a fix.

1. **Will the server accept an absent field?** Today, no: `create flow` answers
   422 and `create environment` answers 400. So a clean CLI payload is blocked on
   the server whatever is decided about meaning.
2. **What does an absent field mean once accepted?** Under the apply rule in this
   document, "use the default", so a re-run still empties a flow's description and
   only the wire payload gets tidier. Under patch, "preserve". Undecided.

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

Not a decision per command - that is ninety-one decisions nobody will remember,
and it is how this happened. The verbs already say it, and the rest of the CLI
already follows them:

| Verb | What an absent field means | Evidence it is already so |
|---|---|---|
| `create`, `begin` | apply - the CLI sends every field, the server writes what arrives | `create flow` replaces the description on both sides |
| `update` | patch - the CLI sends a field only when its flag was passed, the server leaves absent fields alone | `kosli update control vc --name "second name"` leaves the description untouched and takes the record's `version` from 1 to 2 |

By that rule `begin trail` is apply, so the CLI is already right and the server's
guard is the part out of step. Nothing else moves.

The one thing to weigh before agreeing to it: re-running a workflow re-runs
`begin trail`, and under apply a re-run that does not pass `--user-data` leaves
the trail with none. The earlier value is still in the trail's events, so nothing
is lost, but the current view stops showing it. If that is unwanted, the answer
is that trails are patch after all - and then `begin` and `create` mean different
things, and the CLI needs to stop sending unmentioned fields for trails and
attestations.

Whichever way it is settled, the two sides have to change together. If the CLI
stops sending `description` while the server still requires it, `kosli create
flow` fails with a 422; if the server starts ignoring an absent description while
the CLI still sends `""`, nothing changes at all.

## Why it belongs with the empty-value work

Until a command says what an absent flag means, it cannot say what an empty one
means either: today `--description ""` and no `--description` produce the same
request, so the two cannot be told apart. See
`hack/empty-flag-audit/docs/2026-08-13-empty-value-decision.md`.

That document proposes refusing an empty flag value, and it does not reach this.
`kosli begin trail` sends `"description": ""` whether or not `--description` was
passed, so there is no empty value given for such a rule to refuse. This one is
settled by deciding what an absent flag means, not by policing what a present
one carries.

## How it was found

While auditing what the CLI does with empty flag values, in the
`hack/empty-flag-audit/` directory of the CLI repository. An empty `--description`
and an absent `--description`
produced identical results, which is only possible if the absent case is sending
something too. Running every write command with `--debug` and none of its
optional flags showed which fields arrive empty when nobody asked for them.
