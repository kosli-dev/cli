# Plan: `kosli evaluate policy` — evaluate and record a decision in one request

> **Ticket:** https://github.com/kosli-dev/server/issues/6920
> **Status:** written 2026-09-17 against CLI `main` @ `f4f57577`. No slice started.
> **Audience:** the agent or engineer who implements this. Follow the repo's TDD and thin-slice workflow (`CLAUDE.md`). Create a `## feat(evaluate): kosli evaluate policy` section in `TODO.md` from the slice list below before coding.
> **Builds on:** [docs/plans/6700-evaluate-server-side-flag.md](6700-evaluate-server-side-flag.md), which established the evaluations client, the wait, and the shared verdict printer. Read its sections 2 and 4 first: the create/read contract, the timing, and the outcome mapping are not repeated here.
> **Out of scope:** versioned and published policies, and the policy-bundle digest that belongs with them; asynchronous evaluation; migrating our own controls onto the command; removing the client-side evaluator or the hidden `--server-side` flag.

**Public repository.** This repository is public. Everything below stays at the contract a caller can see: the published API schema, the requests the CLI sends, and the answers it gets. Nothing about how the platform is built belongs in this plan, in code comments, or in help text.

---

## 1. What we build

One published command that evaluates a policy and records the outcome in the same request:

```shell
kosli evaluate policy \
    --flow my-release-flow \
    --trail "$GITHUB_SHA" \
    --policy ./policies/SDLC-CTRL-0007-code-review/policy.rego \
    --control SDLC-CTRL-0007 \
    --fingerprint "$ARTIFACT_FINGERPRINT" \
    --params '{"protected_branch": "master"}' \
    --assert
```

What it replaces is a two-step pipeline pattern: evaluate, read `allow` out of a local report, then assert that value back with `kosli attest decision --compliant=<value>`. In that shape the recorded decision is whatever the pipeline typed. Here the caller asks for the decision and never carries its value.

Three things follow from that, and they set every rule below:

1. **The decision value never passes through this CLI.** The command sends the policy, the trail references and a destination. It does not read a verdict and then write it.
2. **A denial is a decision; a broken policy is not.** A policy that cannot run has decided nothing and must never print a denial or leave a non-compliant decision behind.
3. **Without `--control` nothing is recorded.** The evaluation runs, the verdict prints, the destination flags are not required and nothing is written.

---

## 2. Contract

Source of truth: the evaluations endpoints in the published Kosli API schema. Sections 2.1–2.3 of the #6700 plan still hold. This ticket adds one optional block to the create body and one optional field to the read response.

### 2.1 Create, with a decision

```
POST /api/v2/evaluations/{org}
```

```json
{
  "context": { "trails": [ { "flow": "release", "trail": "my-trail" } ] },
  "policy":  { "files": { "policy.rego": "package policy\n\nallow := true\n" } },
  "params":  { },
  "decision": {
    "control": "SDLC-CTRL-0007",
    "name": "SDLC-CTRL-0007-decision",
    "flow": "release",
    "trail": "my-trail",
    "fingerprint": "<sha256>"
  }
}
```

- `decision` is optional. Absent, the evaluation stands alone and writes nothing. Every object in this body still forbids unknown fields, so the block is sent only when asked for, never as nulls.
- `control`, `name`, `flow`, `trail` are required inside the block; `fingerprint` is optional, and absent the decision is recorded against the trail itself.
- The destination is checked when the evaluation is created, not after it has run: a control that does not exist, a destination the token cannot write to, a flow that cannot be attested to, and a fingerprint that is not in that trail are all refused at create time. The caller is never told "queued" for a decision that will never land.
- Refusals keep the envelope and the status codes listed in the #6700 plan; the messages name what was refused. The CLI passes them on untouched — see 4.4.

### 2.2 Read

The evaluation resource gains `decision_attestation_id`: the id of the decision this evaluation wrote, present only where one was asked for and has been written. Everything else about the read is unchanged.

---

## 3. Command surface

| Flag | Required | Meaning |
|---|---|---|
| `--flow`, `-f` | yes | Flow of the trail to evaluate, and of the decision's destination. |
| `--trail` | yes | Trail to evaluate, and the decision's destination. |
| `--policy`, `-p` | yes | A `.rego` file, a directory, or an `http(s)://` URL. |
| `--params` | no | Inline JSON or `@file.json`, unchanged, read by the policy as `data.params`. |
| `--control` | no | The control the decision answers. Present, a decision is recorded; absent, nothing is. |
| `--name` | no | The attestation name the decision is recorded under. Defaults to `<control>-decision`. |
| `--fingerprint` | no | The artifact the decision is about. Absent, the decision is about the trail. |
| `--context` | no | Repeatable `trail=<flow>/<trail>`, to evaluate several trails at one instant. |
| `--assert` | no | Exit non-zero when the policy denies. |
| `--output`, `-o` | no | `table` (default) or `json`, the same shapes `evaluate trail` prints. |

Not offered, and why: `--sync` (the command is always synchronous, so there is nothing to opt into), `--no-assert` (asserting is opt-in here, so its opposite is the default and needs no flag), `--attestations` and `--show-input` (filtering and input display happen where the evaluation runs, and this command never evaluates locally), `--server-side` (this command has no other side).

**Waiting.** The command always waits for a terminal status and prints the verdict. `--assert` changes the exit code on a denial and nothing else. There is no asynchronous mode: a command whose purpose is recording a decision should not return before the decision exists. If one is ever wanted, it is a flag and a ticket of its own.

---

## 4. Design decisions (assumptions for the implementer)

### 4.1 One command, its own options

`evaluate policy` declares its own options struct rather than embedding `commonEvaluateOptions`. It shares the policy load, the params parse, the evaluations client and the verdict printer, and nothing else: the shared struct carries four flags this command must not offer, and inheriting them to hide them again is how the two commands drift apart.

### 4.2 The verdict printer is reused unchanged

Output shape and `--output json` must match `evaluate trail`, so a caller switching commands does not re-parse. That is a constraint on this command, not a licence to change the printer. Anything new this command has to say — the recorded decision id — is said around the verdict, not inside its payload.

### 4.3 Flag combinations refused before any request

- `--name` or `--fingerprint` without `--control`: refused, naming `--control`. They are meaningless alone, and accepting them would look like a decision was recorded.
- More than 100 trails, or a policy bundle over the published caps: refused here, naming the cap, rather than sent to be rejected.

Each refusal says why, in its own sentence, rather than relying on cobra's "these flags conflict" wording.

`--name` is not among these: with `--control` and no `--name`, the name is `<control>-decision`. The default is computed where the request is built and sent explicitly, because the field is required on the wire and a name the caller can predict is worth more than one the API chose.

### 4.4 Outcome mapping

Four outcomes, and they must never be confused with each other:

| Outcome | What prints | Exit |
|---|---|---|
| Completed, allowed | the verdict | 0 |
| Completed, denied | the verdict and its violations | non-zero under `--assert`, else 0 |
| Failed (a classified policy failure) | the failure and its kind, never a verdict | non-zero |
| Unfinished when the wait expires | the evaluation id, never a verdict | non-zero |

Where the server explained itself, its words are passed on untouched. Two refusals keep wording of our own, exactly as in #6700: an organisation that is not entitled to server-side evaluation, and a server too old to serve the route.

### 4.5 One exit code, and messages that tell the outcomes apart

The ticket asks for denial, a broken policy and a fault of ours to be three distinguishable exit codes. They stay one code, as everywhere else in this CLI, and the outcomes are told apart by what they say. #6700 made the same call: a single failure exit path is a product-wide convention, and it is not this command's to change. The obligation that remains is on the wording — a broken policy, an unfinished evaluation and a refused destination each need their own sentence, and none of them may read as a denial. Slice 5 is where that is proved, one test per shape.

### 4.6 A directory of policy files

`--policy` pointing at a directory uploads every file below it as one bundle, keyed by path relative to that directory, within the published 100-file and 1 MiB caps. Paths stay inside the bundle. A single file keeps today's behaviour: one entry named after the file, no extension imposed.

### 4.7 Repeating `--context`

`--context trail=<flow>/<trail>` adds trails to the evaluation. `--flow`/`--trail` always name one of them, and always name where the decision lands. The form is `key=value` so that other kinds of context can be added later without a second flag.

### 4.8 Tests drive a stubbed server

Unchanged from #6700, and the reason is unchanged: this repository's test environment cannot complete a server-side evaluation. Every test of this command drives a stubbed HTTP server that answers the documented shapes. The cost is that these tests pin our half of the contract only; that a decision really lands has to be checked on staging, and that check is in the wrap-up slice.

### 4.9 A new command costs more than the command

Adding a command and its flags means updating the command-surface fixture and the empty-flag audit's own specification, or the build goes red. Budget for it in the first slice, not the last.

---

## 5. Slices

Each is independently mergeable and leaves no half-finished behaviour exposed.

### Slice 1: the command exists, evaluates one trail and prints the verdict

The thinnest end-to-end path: `--flow`, `--trail`, `--policy`, `--params`, no decision, no asserting. Waits for a terminal status and prints the verdict, exiting 0 whatever it is.

- Command registered under `evaluate`, `--help` reads correctly, required flags enforced
- One trail sent as the context, policy uploaded as a one-file bundle, params passed through
- Output matches `evaluate trail` byte for byte, table and json
- A dry run sends nothing and says so
- The command-surface fixture and the audit specification updated

### Slice 2: `--assert`

- `--assert` exits non-zero on a denial, and 0 on an allow
- Without it, a denial still prints in full and exits 0
- An expired wait names the evaluation, prints no verdict, and fails whether or not `--assert` was given

### Slice 3: `--control` records a decision

- `--control` (+ optional `--name`, `--fingerprint`) sends the decision block
- Absent `--name`, the name sent is `<control>-decision`
- Absent `--control`, no decision block is sent at all
- The recorded decision id is read back and reported once the evaluation completes
- A denial records a decision too — nothing about the decision block depends on the verdict

### Slice 4: refusing what cannot be asked for

One test per refusal in 4.3, each checked before any network call.

### Slice 5: server refusals and classified failures

One test per shape: unknown control, unwritable or archived destination, unknown fingerprint, unresolvable trail, organisation not entitled, server too old, enqueue refused, and a policy that does not compile. Each has a sentence of its own; a classified failure never prints a denial and never leaves a decision behind. This slice is where 4.5 is made good: one exit code, and no two outcomes that read alike.

### Slice 6: a directory of policy files

Relative keys, the file and byte caps refused here with the cap named, paths that would climb out of the bundle refused.

### Slice 7: several trails in one evaluation

Repeating `--context`, order preserved, the ceiling refused here, the decision still landing on the one `--flow`/`--trail`.

### Slice 8: wrap-up

Help text and documentation, the changelog entry, `make lint`, the full integration run, and a manual check against staging with an entitled organisation: allow, deny, a broken policy, a decision recorded and read back, and a destination the token cannot write to.

---

## 6. Test strategy

- Command tests in a suite of their own, needing no running server, driving a stub that answers the documented shapes — as `evaluateServerSide_test.go` does today.
- Client tests in `internal/evaluations` for the new block and the new field: sent only when asked for, absent otherwise, and decoded when read back.
- The existing evaluate suites stay on the local test server and must stay green: nothing here changes them.
- Golden output comparisons against `evaluate trail` for the verdict, so the two cannot drift.

---

## 7. Settled by the ticket owner

Recorded here so no slice reopens them:

1. **The policy digest is not built.** It belongs with versioned policies, and nothing in this command sends or shows one.
2. **One exit code.** Clear messages carry the difference between a denial, a broken policy and a fault of ours — see 4.5.
3. **Synchronous only.** No asynchronous mode and no `--sync` flag; the command always waits.
4. **`--name` defaults to `<control>-decision`.**
