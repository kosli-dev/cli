# Extending the audit to ask the same question of the API

`audit.py` answers "does the CLI refuse an empty value". `replay.py`
answers the other half: when the CLI does not refuse it, does the server? This
document sets out how `replay.py` works, what it found, and what it cannot
reach.

## TL;DR

- The question cannot be asked through the CLI, because an empty value mostly
  produces the same request as omitting the flag.
- It can be asked by capturing the request and replaying it, which needs no new
  instrumentation: `--debug` already logs what the CLI sends.
- What a flag controls can be measured rather than written by hand, by diffing
  the request of a run with the flag set against one with it omitted - a payload
  field for a command that writes, a query parameter for one that reads.
- `replay.py` does it, reading the same spec.json and writing results-api.tsv.
- Of 417 rows, 77 are an answer about the server: 25 it accepts empty and 52 it
  refuses.
- Among them, `create flow --template` is refused by the CLI and accepted
  by the API, which stores a flow requiring an attestation whose name is empty.
- Also among them, an artifact can be stored with no name at all, leaving anyone
  who investigates it holding a fingerprint and a blank record.
- Both are measured instances of the gap the decision document argues about:
  refusing empty flag values in the CLI does nothing for a customer calling the
  API directly, so neither of these closes.
- Every emptied filter is accepted: `tag=`, `search=`, `name=` are answered 200
  rather than refused. What that means then differs by endpoint - an empty tag
  matches nothing, an empty repo name matches everything - and the status does
  not say which.
- Most remaining rows are not an answer for good reasons: some flags never leave
  the machine, and multipart requests are not replayed.

## Why the CLI cannot ask this

`2026-08-13-empty-value-decision.md` already establishes the constraint. For 165
of the 183 combinations nothing refuses, an empty value produces a request
identical to the one omitting the flag sends, so there is nothing empty in the
payload for the server to reject. Running the CLI harder cannot get at the
server's behaviour, because the CLI never puts the question to it.

The only way to ask is to send the request the CLI would have sent, with the
field emptied, and read what comes back. That means talking HTTP directly rather
than through the CLI.

## What makes it possible

`--debug` logs the outgoing request, not only the response.
`internal/requests/requests.go:265-276` logs the method, the URL and the
pretty-printed JSON body of every request that carries one, and every request is
logged again with its full URL when the answer comes back - which is where a
read's query string is found, since a read carries no body. The audit already
runs every command with `--debug` and uses only the status line.

## The mapping does not have to be written by hand

What a flag controls is not recorded anywhere, and a hand-written table of 165
flags would be another thing to keep in step with the code. It can
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

A sweep of the 74 commands that can produce a request gives 417 rows, of which
77 are an answer about the server:

| Outcome | Rows |
|---|---|
| no field - the flag changed nothing the request carries | 175 |
| nothing to read - no request was captured | 116 |
| unusable - the captured request was not valid | 49 |
| **the server refuses the emptied value** | **52** |
| **the server accepts the emptied value** | **25** |

Two kinds of request are asked, and they answer differently. The commands that
write carry their flags in a JSON body: 329 rows, 20 refusals and 11
acceptances. The commands that read carry theirs in the query string: 88 rows,
32 refusals and 14 acceptances, and only 2 controls that failed.

The eleven it accepts are worth reading as a list, because each is a field a
customer can empty by calling the API however the CLI behaves:

`create flow --description`, `--template`; `create environment --description`;
`create service-account --description`; `update control --description`;
`attest override --description`; `attest artifact --artifact-type`,
`--display-name`; `report artifact --artifact-type`, `--name`; `tag --unset`.

`create flow --template` is the one to look at. `--template ""` is refused by the
CLI, by the wrapper every flag's value carries (`cmd/kosli/nonEmptyValue.go`,
applied by the walk in `cmd/kosli/root.go`).
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
document's proposal, which says a CLI rule can only reach CLI traffic, with a
measured instance behind it instead of a worked example.

The four rows reaching `filename` say the same thing about a different field,
and it is the one a customer would feel. The server accepts an artifact whose
`filename` is empty, stores it, and serves it back that way:

```
$ kosli get artifact FLOW@aaaa...
Name:
Flow:              empty-filename-flow
Fingerprint:       aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa
```

An artifact is identified by its fingerprint, so no compliance verdict changes.
What is lost is the only human-readable handle on the record. Someone who finds
something important about an artifact and has its fingerprint opens the record
to learn what the thing actually was, and there is nothing there. The trail is
intact and unreadable, which for a product whose value is the auditability of
the trail is not a small defect - and it is Kosli that accepted the record
without saying anything.

That puts it in the same family as `--redact-commit-info` and `repo_info` in the
decision document: the record is poorer than the customer believes it to be, and
nothing said so at the time.

No route through the CLI produces it, which is why it is here rather than in the
decision document's table. But nothing in the CLI checks this field, and the
distinction matters: it is not protected, it is only never constructed empty.

`attest artifact` decides the name like this (`cmd/kosli/attestArtifact.go:174`):

```go
if o.displayName != "" {
    o.payload.Filename = o.displayName
} else {
    ...
    o.payload.Filename = filepath.Base(args[0])
}
```

An empty `--display-name` fails that test, so the else branch runs and the name
comes from the artifact named on the command line. A workflow step written as

```
kosli attest artifact ./target/my-app.jar --artifact-type file \
  --display-name "$RELEASE_NAME" ...
```

sends `"filename": "release-2026.8"` when the variable holds one, and
`"filename": "my-app.jar"` when it is unset - identical to what omitting the
flag sends, with nothing said about the difference. The artifact is recorded, and
recorded under the wrong name.

`report artifact --name ""` does the same. `--artifact-type ""` on both commands
fails earlier, and not on account of the filename either: `either
--artifact-type or --fingerprint must be specified`
(`cmd/kosli/cli_utils.go:488`) is a rule about which of two flags is set, and an
empty value reads there as unset.

So the CLI never sends an empty `filename` because the positional argument is
required and the fallback always has it, not because anyone validates it. Only a
caller writing the request themselves can send `""`, and the empty-value rule
proposed for the CLI would not close it.

The twenty it refuses are the reassuring half, over 14 distinct fields, and
they are mostly types rather than emptiness: `include_scaling`,
`require_provenance` and `privilege` are a boolean and an enum, `user_data` and
`artifacts` are objects, and an empty string is not one. `origin_url` is refused
as a URL.

The probe also confirmed, from the other direction, the ordering constraint in
step 2 of the staged release in the decision document's Appendix 0. A `create
flow` payload with no `description` is refused with `400 Input payload
validation failed`, `description: Field required`. The server must accept an
absent description before the CLI can stop sending one.

## Every emptied filter is accepted

The fourteen acceptances among the reads are all the same kind of flag:

`list controls --search`, `--tag`; `list environments --name`, `--tag`,
`--space-id`; `list repos --name`, `--search`, `--tag`, `--repo-id`; `list
trails --fingerprint`, `--flow-tag`; `list snapshots --interval`; `log
environment --interval`; `get artifact --trail`.

Every one is a filter, and the server takes all of them: it is sent `tag=` where
it was sent `tag=probe`, and answers 200 rather than refusing.

What it then means is not the same from one endpoint to the next, which is worth
more than the acceptance itself. An empty `tag` on `list environments` is
applied and matches nothing - the server returns zero environments where
omitting the parameter returns all of them. An empty `repo_name` on `list
artifacts` is ignored and every artifact comes back. Two filters, both accepted,
one narrowing to nothing and the other widening to everything, and a caller
cannot tell which from the status.

The narrowing one is the API-side twin of `kosli list environments --tag "$VAR"`
in the decision document, which answers "No environments were found" and exits
0: a filter that matched nothing, indistinguishable from a real no-match. The
widening one is worse in a compliance setting, because a query meant to be
narrow silently answers about everything.

The refusals are the structural parameters rather than the filtering ones -
`per_page`, `page`, `sort_direction`, `reverse`, `snappish1` and `snappish2` -
which the server type-checks and rejects when they arrive empty.

So the two halves agree, and unhelpfully. The decision document's appendix 2
already says of the filters that an empty value means nothing for them - "there
is no artifact called ''" - and records that 21 of the 24 filter flags are never
refused by the CLI. The server does not refuse them either. A customer calling
the API directly gets the same silently-empty answer, which refusing empty flag
values in the CLI would not change.

The two unusable controls are both `--repo`, and they show the same parameter
treated a third and fourth way. On `list artifacts` an unknown `repo_name` 404s
while an empty one returns 200 and every artifact. On `log environment` an empty
one is looked up and refused: `{"message":"Repo '' not found"}`. Neither can be
claimed as a verdict while the control 404s - the audit has no repository that
exists, for want of a connected git provider - but the inconsistency is
recorded.

## Why most rows are not an answer

The 175 and the 116 are not failures of the probe, and most of them are cases
where there is nothing for it to ask:

- **Some flags never leave the machine.** `--output`, `--repo-root`, `--exclude`
  and the like change what the CLI does, not what it sends. "The flag changes
  nothing the request carries" is the true answer for them, not a gap.
- **`kosli fingerprint` sends nothing at all**, by design, and neither does any
  `--dry-run` run.
- **A command whose run with a real value fails sends nothing to capture.**
  That is most of the 116, and it is the audit's own doing rather than the
  CLI's.

The 49 unusable controls are the ones worth fixing, and they are a symptom of
something larger: the audit invents the values it gives flags. Its own control
runs failed 114 times for the same reason, now 29. That is written up in
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

- **Query parameters are asked too.** A read carries its flags in the URL rather
  than a body, so the same diff is taken over the query string and the parameter
  is emptied there. The method is not logged on the line that survives a read,
  so only commands whose verb settles it - `get`, `list`, `log`, `diff`,
  `search` - are read this way. Replaying a read is also the only replay that
  is safe to repeat, since nothing is created.

Not settled:

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

It also gives the server-side work a test list. What the server has to enforce
is not "refuse an empty value" but "an absent field means what the verb says it
means", and a probe that replays every captured request with one field emptied,
and again with it removed, is how that gets checked rather than asserted.
