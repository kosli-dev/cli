# Issue #6920: Evaluate an inline policy and record its decision in one request: `kosli evaluate policy`

> **Last updated:** 2026-09-17
> **Issue:** https://github.com/kosli-dev/server/issues/6920
> **Implementation plan:** [docs/plans/6920-evaluate-policy.md](../plans/6920-evaluate-policy.md)
> **Collaborators:** Simon Castagna (engineer), Claude (claude-opus-5)

---

## Problem Definition

Server-side evaluation reaches the CLI today only through the hidden `--server-side` flag on `kosli evaluate trail` (#6700). It evaluates, and it records nothing. The decision Kosli stores still comes from a second command: a pipeline step reads `allow` out of a local report and asserts it back with `kosli attest decision --compliant=<that value>`. Kosli may have done the evaluating, but the decision in the audit record is asserted by the pipeline, from a value the pipeline could have typed by hand.

`kosli evaluate policy` is one command and one request: the caller sends the policy, the trail references, a control and a destination, and the decision is recorded where the evaluation runs. The decision value never passes through this CLI.

**Why it matters.** This is the first command whose output *is* the governance record, with no human-asserted value in between. It is what lets our own controls stop being self-reported.

**Constraints and acceptance criteria.**

- One request evaluates and records. No second command, no client-asserted compliance value.
- Without `--control` the evaluation runs, the verdict prints, and nothing is recorded.
- A denial is a decision, recorded as non-compliant with its violations. A policy that cannot run is not: it records nothing and must never print a denial.
- Destination flags without `--control` are refused, and a control or a destination the caller cannot write to is refused before the evaluation is queued.
- The command always waits for a terminal status; `--assert` exits non-zero on denial and changes nothing else.
- `--policy` accepts a single file or a directory, within the published bundle caps.
- Output and `--output json` match `kosli evaluate trail`, so a caller switching commands does not re-parse.
- An organisation without the server-side evaluation entitlement is refused in words that name it.
- `--name` defaults to `<control>-decision`, so the common case names only the control.

**Scope.** Asynchronous evaluation is out of the first cut: the command always waits. Versioned and published policies are out of scope for the ticket, and again by decision for this round: an inline policy is the only policy there is until publishing exists. The policy-bundle digest goes with them. Also out: migrating our own controls onto this command, `kosli evaluate input`, and removing the client-side evaluator or the hidden `--server-side` flag.

**Public repository.** This repository is public. The plan, the code and the help text stay at the contract a caller can see — the published API schema, the requests sent and the answers received.

---

## Plan

Transcribed from [docs/plans/6920-evaluate-policy.md](../plans/6920-evaluate-policy.md), committed on this branch. Read it before starting any slice: it holds the contract this ticket adds, the command surface, the outcome mapping and a per-slice test list ready to copy into `TODO.md`. It builds on the #6700 plan rather than repeating it. Each slice is independently mergeable.

- [x] Slice 1 — the command exists, evaluates one trail and prints the verdict.
- [x] Slice 2 — `--assert` exits non-zero on a denial.
- [x] Slice 3 — `--context` names what is evaluated, and is always required.
- [ ] Slice 4 — `--control` records a decision, with `--flow` and `--trail` as its destination.
- [ ] Slice 5 — server refusals and classified failures, one test per shape, each with a sentence of its own.
- [ ] Slice 6 — a directory of policy files as one bundle, with the caps refused here.
- [ ] Slice 7 — help text, docs, changelog, lint, full test run, and a staging check against an entitled organisation.

---

## PRs & Branches

| Branch | PR | Status |
|--------|----|--------|
| `6920-evaluate-policy` | — | in progress |

---

## Decisions Made

- The command declares its own options rather than inheriting the evaluate commands' shared ones, because four of those flags have no meaning here and inheriting them only to hide them is how two commands drift apart.
- Asserting is opt-in on this command and the default is silent, which is the reverse of `evaluate trail`. A command that records a decision should not fail a pipeline unless the caller asked it to, and the flag that asks is the one the tutorial already publishes.
- What is evaluated and where a decision lands are named separately: `--context` is the only way to say what to evaluate and is always required, while `--flow` and `--trail` name the destination alone. Neither is refused for being present without `--control`, because a pipeline sets them as environment variables for every command it runs, and refusing them would refuse an ordinary run that asked for no decision. The ticket's example predates this split.
- The first cut of the command is synchronous only, and `--sync` is not offered: with nothing to opt into, the flag would name the one behaviour there is. A command whose purpose is recording a decision should not return before the decision exists. An asynchronous mode, and the `--sync` flag that would pair with it, belong to a later ticket if anyone asks for them.
- `--name` defaults to `<control>-decision` rather than being required beside `--control`, so the common case names the control once. The default is computed before the request is sent, because a name the caller can predict is worth more than one chosen further away.
- Nothing is recorded without `--control`, and a destination flag without one is refused rather than ignored: accepting it would read as a decision having been recorded when none was.
- The verdict printer is reused untouched, so a caller moving from `evaluate trail` does not re-parse. Anything this command has to add — the evaluation id, the recorded decision id — is said around the verdict rather than inside its payload.
- Versioned policies are excluded from this round by the engineer, on top of the ticket's own exclusion of policy publishing. Until publishing exists, an inline policy is the only policy there is.
- Every outcome keeps the one exit code this CLI has always used, against the ticket's request for three. A single failure exit path is a product-wide convention and not this command's to change, as #6700 also found. The obligation moves to the wording instead: a denial, a broken policy, an unfinished evaluation and a refused destination each get a sentence of their own, and none may read as another.
- A slice of its own for refused flag combinations was dropped once the command became synchronous and `--name` gained a default: the only refusal left is a destination flag without `--control`, which belongs with the decision block that gives it meaning.
- The policy-bundle digest is not built. It belongs with versioned policies, so a decision recorded now cites the policy it ran by the evaluation that ran it and nothing more.
- Tests drive a stubbed server, as under #6700 and for the same reason: this repository's test environment cannot complete a server-side evaluation. That an evaluation really records a decision has to be checked on staging, which is why the wrap-up slice carries a manual check rather than a test.

---

## Next Steps

- [ ] Slice 4: the decision block, the `--name` default, and the refusal of `--name` or `--fingerprint` without `--control`.
- [ ] Check what the create endpoint refuses for each destination failure, against staging, so Slice 5's messages are written from real answers rather than guessed.
