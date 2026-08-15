# Extending the audit to ask the same question of the API

The empty-flag audit answers "does the CLI refuse an empty value". This is about
answering the other half: when the CLI does not refuse it, does the server? The
answer is yes, the audit can be extended to ask that, and a prototype has done
it for one command and found something with it.

## TL;DR

- The question cannot be asked through the CLI, because an empty value mostly
  produces the same request as omitting the flag.
- It can be asked by capturing the request and replaying it, which needs no new
  instrumentation: `--debug` already logs the outgoing URL and JSON body.
- The flag-to-field mapping can be measured rather than written by hand, by
  diffing the payload of a run with the flag set against one with it omitted.
- A prototype on `kosli create flow` works end to end.
- It found that `--template ""` is refused by the CLI and accepted by the API,
  which stores a flow requiring an attestation whose name is empty.
- That is the first measured instance of the gap the decision document argues
  about: a CLI rule does not cover a customer calling the API directly.

## Why the CLI cannot ask this

`2026-08-13-empty-value-decision.md` already establishes the constraint. For 162
of the 166 combinations nothing refuses, an empty value produces a request
identical to the one omitting the flag sends, so there is nothing empty in the
payload for the server to reject. Running the CLI harder cannot get at the
server's behaviour, because the CLI never puts the question to it.

The only way to ask is to send the request the CLI would have sent, with the
field emptied, and read what comes back. That means talking HTTP directly rather
than through the CLI.

## What makes it possible

`--debug` logs the outgoing request, not only the response. `internal/requests/requests.go:265-276`
logs the URL and the pretty-printed JSON body of every request before it is
sent. The audit already runs every command with `--debug` (`empty-flag-audit/audit.py:212`)
and uses only the status line, discarding the payload.

So the extension needs no change to the CLI to see what is being sent.

## The mapping does not have to be written by hand

Which payload field a flag controls is not recorded anywhere, and a hand-written
table of 152 flags would be another thing to keep in step with the code. It can
be measured instead: capture the payload of the run with the flag set and the
run with the flag omitted, and the field that differs is the field that flag
controls.

Two details make the diff clean. The three runs must use the same resource name,
or the name shows up as a field the flag controls. And the comparison is between
`omitted` and `set`, not `empty` and `set`, because the empty run may not reach
the point of sending anything.

## The prototype

`/tmp/claude-501/empty-api-probe.py`, in the session scratchpad rather than the
repo. It probes `kosli create flow`, and for each flag it:

1. runs the CLI three times with `--debug`, capturing URL and payload for the
   flag omitted, set, and empty
2. derives the field that flag controls by diffing the omitted and set payloads
3. replays the captured request twice against the local server: once unmodified
   as a control, and once with that field emptied

The control matters. Without it a rejection cannot be told apart from a replay
that was malformed for some unrelated reason. It is the same three-run shape the
audit already uses, moved down a layer.

## What it found

| Flag | Field | What the CLI does | What the API does |
|---|---|---|---|
| `--description` | `description` | accepts it, and sends a payload identical to the one omitting the flag sends | 201 |
| `--template` | `template` | refuses it, exit 1 | **201** |

The first row is the expected case, and it confirms by measurement what the
decision document argues: there is nothing for the server to reject, because
nothing empty ever reaches it.

The second row is new. `--template ""` is refused by the CLI, by the
`nonEmptyStringSlice` type added in #1092 (`cmd/kosli/nonEmptyStringSlice.go:99`,
used on `--template` at `cmd/kosli/createFlow.go:90` and `--attachments` at
`cmd/kosli/flags.go:113`). Sending `template: [""]` straight to the API is
accepted, and reading the flow back shows it stored:

```
version: 1
trail:
  attestations: []
  artifacts:
  - name: artifact
    attestations:
    - name: ''
      type: '*'
```

A flow requiring an attestation whose name is empty, which no attestation can be
made to match by name. The CLI-side type closes this for people using the CLI.
It does nothing for anyone calling the API.

That is the argument in the "What a CLI rule cannot reach" section, with a
measured instance behind it instead of a worked example.

The probe also confirmed the ordering constraint in step 2 of the release plan
from the other direction. A `create flow` payload with no `description` is
refused with `400 Input payload validation failed`, `description: Field
required`. The server must accept an absent description before the CLI can stop
sending one.

## What has to be settled before this generalises

None of these is hard, but none is free either.

- **The method is not logged.** `--debug` logs the URL but not the HTTP method,
  so the prototype read `http.MethodPut` out of `cmd/kosli/createFlow.go:138`.
  Either the method joins the log line, which is a one-line change, or the
  extension keeps a per-command table, which is the hand-written mapping this
  approach otherwise avoids. Logging it is the better of the two.
- **Multipart requests log only their JSON fields** (`internal/requests/requests.go:334-344`),
  so the commands carrying attachments need their own handling.
- **Not every flag becomes a payload field.** Query parameters on the `list` and
  `get` commands, and flags that never leave the machine at all (`--exclude`,
  `--repo-root`, `--output`), have no body field to empty. The query-parameter
  ones are still reachable by emptying the parameter in the captured URL.
- **Replaying mutates the server**, so each replay needs a fresh resource name
  and the existing reset discipline. The probe refuses to send anywhere but the
  local host, which is what keeps a run of it away from app.kosli.com.
- **The third-party services stay out of reach.** AWS, Azure, Jira, SonarQube
  and the rest are unaffected by any of this. What they do with an empty value
  is still only visible in customers' runs.

## Why it is worth doing

The decision document's proposal fixes the CLI, and says plainly that a customer
calling the API directly keeps every one of these holes. Today that claim rests
on reasoning about what the payloads must look like. This turns it into the same
kind of measurement the rest of the audit is built on, and it does so for the
layer that covers every client rather than one of them.

It also gives the server-side work a test list. The rule for the server is not
"refuse an empty value" but "an absent field means what the verb says it means",
and a probe that replays every captured request with one field emptied, and
again with it removed, is how that rule gets checked rather than asserted.
