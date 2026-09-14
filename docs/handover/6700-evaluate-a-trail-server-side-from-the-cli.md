# Issue #6700: Evaluate a trail server-side from the CLI behind a hidden flag

> **Last updated:** 2026-09-14
> **Issue:** https://github.com/kosli-dev/server/issues/6700
> **Implementation plan:** [docs/plans/6700-evaluate-server-side-flag.md](../plans/6700-evaluate-server-side-flag.md)
> **Collaborators:** Simon Castagna (engineer), Claude (claude-fable-5-1, claude-opus-5)

---

## Problem Definition

`kosli evaluate trail` and `kosli evaluate trails` evaluate entirely on the user's machine today. The CLI reads the trail, reshapes it, fetches every attestation one at a time, then compiles and runs the Rego policy in its own process. Three sibling tickets (#6620, #6621, #6622, all closed) built the server-side path: an endpoint that creates an evaluation from an inline policy and trail references, an input shaper that derives the policy input from the stored trail moment, and a worker that invokes the OPA Lambda and records a terminal result.

This issue joins the two. A hidden flag on the two trail commands sends the policy and the trail references to the evaluations endpoint, waits for a terminal status, and prints the same verdict with the same exit-code behaviour. The flag is hidden rather than published because the point is to run the two paths against each other while neither is a documented contract.

**Why it matters.** The two evaluators do not agree today, and the disagreement is invisible until something runs both. The server accepts policies the local path refuses, and it feeds the policy a different input document. This flag is the first place either difference can be observed.

**Constraints and acceptance criteria.**

- Without the flag, the commands behave exactly as they do today.
- A classified policy failure is never reported as a denial. A policy that does not compile must not print a denial verdict.
- A denial exits non-zero and the no-assert flag exits zero, driven by the server's verdict.
- Several trail names go in one evaluation, so they all resolve at a single instant.
- An organisation without the feature flag gets a refusal that names the flag.
- The wait is bounded, and expiring is reported as an unfinished evaluation, never as a verdict.

**Scope note.** The issue originally carried a second half: creating the server-side evaluation even when the flag is absent, discarding the result, and measuring the disagreement rate at real volume. That half is now #6832, which also carries the data-egress question about uploading customer-authored policy source to Kosli. #6831 covers dogfooding our own controls through the server-side path. Nothing in this issue sends a policy to Kosli unless the user asks for it.

**Out of scope:** publishing the flag, removing the local evaluator, the input subcommand, recorded decisions, policy publishing.

---

## Plan

Transcribed from [docs/plans/6700-evaluate-server-side-flag.md](../plans/6700-evaluate-server-side-flag.md), committed on this branch. Read it before starting any slice: it holds the verified server contract, the known mismatches between the two evaluation paths, the design decisions behind each choice, and a per-slice test list ready to copy into `TODO.md`. Each slice is independently mergeable.

- [x] Slice 0 — split the verdict printer away from the evaluation, so both paths share it. No behaviour change. Two trail suites could not be run: they need the local test server, which needs a production token.
- [ ] Slice 1 — new client package: create an evaluation, decode the accepted response and every error shape.
- [ ] Slice 2 — read an evaluation, and wait for a terminal status with backoff and a bounded budget.
- [ ] Slice 3 — the hidden flag on the single-trail command, happy path end to end.
- [ ] Slice 4 — the same flag on the multi-trail command, all names in one evaluation.
- [ ] Slice 5 — refuse the flag combinations the server has no equivalent for.
- [ ] Slice 6 — classified failures and server refusals, one test per failure kind.
- [ ] Slice 7 — policy upload edge cases: remote policies, file naming, the size cap.
- [ ] Slice 8 — distinct exit codes for denial, broken policy, expired wait and our own faults.
- [ ] Slice 9 — lint, full test run, a staging check against an entitled organisation, and the first latency comparison between the two paths.

---

## PRs & Branches

| Branch | PR | Status |
|--------|----|--------|
| `6700-evaluate-server-side` | — | in progress |

---

## Decisions Made

- The work is planned as ten slices with the plan committed to the repository before any code, because the ticket's hard part is the contract between two evaluators that disagree, and that had to be established from the server source rather than assumed.
- The two paths are knowingly not equivalent, and the flag stays hidden for exactly that reason: the server accepts policies the local evaluator refuses, and it feeds the policy the trail's recorded moment rather than the live trail document. A policy reading trail metadata outside the compliance slots will answer differently under the flag. Reconciling them is not this issue's job; making the difference observable is.
- The local policy validity check is deliberately skipped under the flag, so a broken policy is classified by the evaluator that will actually run it rather than pre-judged by a stricter rule the server does not share.
- Input filtering and input display are refused under the flag rather than approximated. The server offers no equivalent for either, and showing the caller a later moment than the one evaluated would mislead more than refusing.
- This repository's test environment cannot complete a server-side evaluation at all: the local stack runs a server, a database and object storage, with no queue broker, no worker and no evaluator. The local server also treats every local caller as entitled, so the feature-flag refusal cannot be reproduced there either. Every test of the new path therefore drives a stubbed HTTP server, and the existing tests stay on the real local server to prove the unflagged path is untouched. The cost accepted is that these tests pin our side of the contract only; real agreement between the two paths has to be measured on staging.
- A denial, a broken policy, an expired wait and a fault of ours become four distinguishable exit codes, which this CLI has never had. They stay provisional while the flag is hidden, and the unflagged path keeps its single failure code unchanged.
- Every slice shares one branch for the whole issue rather than a branch each, so the history reads as one piece of work and a reviewer can follow the plan through to the code without walking a stack of branches.
- The wait is bounded at the platform's stated enqueue-to-terminal ceiling. Expiring reports an unfinished evaluation carrying its identifier, so a slow evaluation can never be mistaken for a denial.

---

## Next Steps

- [ ] Confirm the four proposed exit-code values with the ticket owner before Slice 8; everything up to it is unaffected.
- [ ] Start Slice 0 on a branch of its own: separate the verdict printer from the evaluation, with the existing tests as the only guard.
- [ ] Build the stubbed-server test helper early in Slice 3, since Slices 3 to 7 all depend on it.
- [ ] Decide whether the local test stack should later gain a broker, a worker and an evaluator, which would turn the stubbed tests into a real end-to-end check. Not required here, and a ticket of its own.
- [ ] Measure both paths on staging against a trail with many attestations, which is the latency comparison the ticket predicts and the first number it can produce.
