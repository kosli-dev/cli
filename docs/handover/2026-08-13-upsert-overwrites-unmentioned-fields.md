# Bug: an upsert overwrites fields the command never mentioned

Several commands are upserts. Running one a second time without `--description`
replaces the description the object already had with an empty one. No empty value
is involved and no variable needs to be unset: omitting the flag is enough. The
same happens to `--user-data`, and to `create policy --comment`.

Anyone driving these from CI, passing the flag on some runs and not others, has
objects whose description or user data no longer says what they set it to.

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

Everything found so far, by running each command twice and reading the object
back:

| Command and flag | after the first run | after a second run without the flag |
|---|---|---|
| `create flow --description` | `the payments flow` | `""` |
| `create policy --description` | `the env policy` | `""` |
| `create policy --comment` | `the review note` | `""` |
| `begin trail --description` | `the release trail` | `""` |
| `begin trail --user-data` | the document | `{}` |
| `attest generic --user-data` | the document | `{}` |
| `attest custom --user-data` | the document | `{}` |
| `attest junit --user-data` | the document | `{}` |
| `attest decision --user-data` | the document | `{}` |
| `attest override --user-data` | the document | `{}` |

Every run exits 0 and prints what success prints.

`create environment` sends `"description": ""` in the same way but the
description survives, so sending an empty field is not on its own enough - the
server decides. Only running each one settles it, which is why the list above is
what was observed rather than what the payloads suggest.

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

## How much is actually lost

Less than "wipes" suggests, and it differs by object. Worth being accurate about,
because it decides how urgent this is.

| Object | What keeps the earlier value | Recoverable? |
|---|---|---|
| trail | `events[]`, each carrying the `trail_data_json` that was reported | yes |
| attestation | the earlier attestation event on its trail | yes |
| flow | nothing on the flow record itself, but every trail created from it snapshots the flow state it used | the flow's own description, no |
| policy | `events[]`, but each entry records only that a `metadata_update` happened, with `data: {}` | no |

So on trails and attestations the current view stops matching what was reported,
while the report itself is still there. On flows and policies the previous text is
gone.

None of it changes a compliance verdict. A trail keeps the template it was created
with, so editing a flow afterwards does not reach back into trails already
created: state is snapshotted, not shared. What is lost is what someone wrote to
explain the thing, not what any policy evaluates.

The reason it belongs with the empty-value work is narrower than the data: these
commands have no way to say "leave this alone", which is exactly why an empty
`--description` and an absent one are indistinguishable. Fixing that is a
precondition for refusing empty values on those flags - see
`2026-08-13-empty-value-decision.md`.

## How it was found

While auditing what the CLI does with empty flag values
(`hack/empty-flag-audit`). This one turned up because an empty `--description`
and an absent `--description` produced identical results, which is only possible
if the absent case is also sending something.
