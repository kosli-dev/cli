# Extending the audit to ask the same question of the API

The empty-flag audit answers "does the CLI refuse an empty value". `replay.py`
answers the other half: when the CLI does not refuse it, does the server? This
is what it does, what it found, and what it cannot reach.

## TL;DR

- The question cannot be asked through the CLI, because an empty value mostly
  produces the same request as omitting the flag.
- It can be asked by capturing the request and replaying it, which needs no new
  instrumentation: `--debug` already logs the outgoing URL and JSON body.
- The flag-to-field mapping can be measured rather than written by hand, by
  diffing the payload of a run with the flag set against one with it omitted.
- `replay.py` does it, reading the same spec.json and writing results-api.tsv.
- Of 412 rows, 19 are an answer about the server: 11 fields it accepts empty and
  8 it refuses.
- Among the eleven, `create flow --template` is refused by the CLI and accepted
  by the API, which stores a flow requiring an attestation whose name is empty.
- That is a measured instance of the gap the decision document argues about: a
  CLI rule does not cover a customer calling the API directly.
- Most rows are not an answer, and mostly for good reasons: the read commands
  send no body, and some flags never leave the machine.

## Why the CLI cannot ask this

`2026-08-13-empty-value-decision.md` already establishes the constraint. For 172
of the 189 combinations nothing refuses, an empty value produces a request
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

## The tool

`replay.py`, beside `audit.py` and reading the same `spec.json`. Not a third
mode of the audit: `--ci` asks the audit's question in another environment and
its rows line up with the plain run's, while this asks a different question and
its rows are about a captured request. It writes `results-api.tsv`.

For each command-and-flag pair it:

1. runs two of the audit's own invocations with `--debug`, capturing method, URL
   and payload with the flag omitted and with it set
2. derives the field that flag controls by diffing those two payloads
3. replays the captured request twice: once unmodified as a control, and once
   with that field emptied

It reuses `invocation_for`, so it captures the very invocations the audit
measures, and `normalise`, so two runs' fixture names do not read as a field the
flag controls - without that, `--description` appeared to control `name`.

The control matters, and earned itself on the first command it ran. Without it a
`400` beside an emptied `400` reads as a refusal when the request was never
valid. Each row says which it was.

## What it found

A sweep of the 74 commands that can produce a request gives 412 rows, of which
19 are an answer about the server:

| Outcome | Rows |
|---|---|
| nothing to read - no JSON body was captured | 254 |
| no field - a body was sent, the flag changed nothing in it | 104 |
| unusable - the captured request was not valid | 35 |
| **the server accepts the emptied field** | **11** |
| **the server refuses the emptied field** | **8** |

The eleven it accepts are worth reading as a list, because each is a field a
customer can empty by calling the API however the CLI behaves:

`create flow --description`, `--template`; `create environment --description`;
`create service-account --description`; `update control --description`;
`attest override --description`; `attest artifact --artifact-type`,
`--display-name`; `report artifact --artifact-type`, `--name`; `tag --unset`.

`create flow --template` is the one to look at. `--template ""` is refused by the
CLI, by the `nonEmptyStringSlice` type added in #1092
(`cmd/kosli/nonEmptyStringSlice.go:99`, used on `--template` at
`cmd/kosli/createFlow.go:90` and `--attachments` at `cmd/kosli/flags.go:113`).
Sending `template: [""]` straight to the API is accepted, and reading the flow
back shows it stored:

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
It does nothing for anyone calling the API. That is the argument in the decision
document's "What a CLI rule cannot reach" section, with a measured instance
behind it instead of a worked example.

The two `--artifact-type` rows say the same thing about a different field: the
server accepts an artifact whose `filename` is empty.

The eight it refuses are the reassuring half, and they are mostly types rather
than emptiness: `include_scaling`, `require_provenance` and `privilege` are a
boolean and an enum, `user_data` and `artifacts` are objects, and an empty
string is not one. `origin_url` is refused as a URL.

The probe also confirmed the ordering constraint in step 2 of the release plan
from the other direction. A `create flow` payload with no `description` is
refused with `400 Input payload validation failed`, `description: Field
required`. The server must accept an absent description before the CLI can stop
sending one.

## Why most rows are not an answer

The 254 and the 104 are not failures of the probe, and three quarters of them
are cases where there is nothing for it to ask:

- **The reads send no body.** `list` (51 rows), `get` (19), `log` (10), `diff`
  and `search` put their flags in the query string. Reachable, by emptying the
  parameter in the captured URL, which is a second mechanism this does not have.
- **Some flags never leave the machine.** `--output`, `--repo-root`, `--exclude`
  and the like change what the CLI does, not what it sends. "The flag changes no
  field of the payload" is the true answer for them, not a gap.
- **`kosli fingerprint` (6 rows) sends nothing at all**, by design.
- **`--dry-run` runs send nothing**, also by design.

The 35 unusable controls are the ones worth fixing, and they are a symptom of
something larger: the audit invents the values it gives flags, and 114 of its
own control runs fail for the same reason. That is written up in
`2026-08-15-the-audit-invents-its-input-values.md`.

## What is settled, and what is not

Settled while building it:

- **The method is logged.** `--debug` prints it beside the URL
  (`internal/requests/requests.go:268`), so nothing here keeps a hand-written
  table of which endpoint takes which verb.
- **Replays do not collide.** A payload field that names the resource is made
  unique per replay, so a control and an emptied replay are two creates rather
  than a create and an update. The probe refuses to send anywhere but the local
  host, which is what keeps a run of it away from app.kosli.com.

Not settled:

- **Query parameters.** The read commands put their flags in the URL, which is
  85 of the rows that produced no answer. Emptying a parameter in the captured
  URL is the same idea one layer over, and it is the largest single thing
  missing.
- **Multipart requests log only their JSON fields**
  (`internal/requests/requests.go:334-344`), so `--attachments` and the
  template-file endpoints cannot be replayed as JSON. Two `create flow` rows
  show this as a control that fails.
- **Fields the CLI never sends.** The schema at `/api/v2/openapi.json` describes
  202 fields, and the probe can only ask about the ones some flag puts in a
  payload. A field no flag reaches is a hole no amount of driving the CLI finds.
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
