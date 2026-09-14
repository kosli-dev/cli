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
- [x] Slice 1 — new client package: create an evaluation, decode the accepted response and every error shape.
- [x] Slice 2 — read an evaluation, and wait for a terminal status with backoff and a bounded budget.
- [x] Slice 3 — the hidden flag on the single-trail command, happy path end to end.
- [x] Slice 4 — the same flag on the multi-trail command, all names in one evaluation.
- [x] Slice 5 — refuse the flag combinations the server has no equivalent for.
- [x] Slice 6 — classified failures and server refusals, one test per failure kind.
- [x] Slice 7 — policy upload edge cases: remote policies, file naming, the size cap.
- [x] Slice 8 — distinct exit codes. **Deferred by decision, not built.** Every failure keeps exit code 1.
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
- Two limits of the waiting are accepted rather than engineered around, and both are stated in the code so nobody reads more into it than it does. The budget governs how long the evaluation is asked about, not a single read, because a read carries no cancellation and the shared HTTP client sets no overall deadline; a very slow server can therefore overrun the budget by one read. And creating an evaluation is not idempotent while that same client retries it, so one command can leave a duplicate behind if an answer is lost in transit. Duplicates agree with each other, since each is deterministic for the same policy and instant, so the cost is wasted work rather than a wrong verdict. Fixing either means changing code every command shares, which is a decision for its own ticket.
- Every failure still exits with the one code this CLI has always used, and the outcomes are told apart by what they say instead. Giving them separate codes was planned and then dropped: the tool has a single failure exit path that every command shares, so inventing codes would set a convention for the whole product from inside one hidden flag. The ticket's requirement is met without them, since each outcome already has its own sentence and a broken policy or an unfinished evaluation is never worded as a denial. A denial exiting with that code is also already promised in published help text. Worth its own ticket if a caller ever needs to branch on the outcome.
- The policy keeps the name of the file it came from rather than being renamed to a fixed one, and no file extension is imposed. Reading the evaluator settled that: it parses every entry in a bundle as a policy module whatever the entry is called, so the name only labels the error messages. Renaming would make the server cite a file the author does not have on disk.
- Where the server explained itself, its words are passed on untouched, and only two refusals get wording of their own: an organisation that is not entitled, whose message says nothing about what to do next, and a server too old to serve the route, which sends no words at all. Those two are told apart from an ordinary refusal by whether the error envelope carried a message field, a structural tell rather than a match on English wording, and a test pins it so the day that stops being true is a failure rather than a silent misdiagnosis.
- The two options with no server-side answer are refused outright rather than quietly ignored, and each refusal says why. Honouring either halfway would be worse than refusing it: a filter that was dropped would evaluate more than the caller asked about, and an input printed from here would not be the input the server judged. They are stated as errors rather than as a flag-exclusion rule, because the built-in wording says only that two flags conflict, and someone reaching for an undocumented flag needs the reason.
- The tests for the flag sit in a suite of their own that needs no running server, rather than joining the existing trail suites. Those build a flow and a trail before every test, so anything added there inherits a dependency this feature cannot satisfy anyway: the local stack has no evaluator. The cost is that these tests prove only our half of the contract, and agreement with the real server still has to be measured on staging.
- Whether the two paths print the same page turns on how an empty violation list is rendered, which was not foreseen. Running the binary showed the local path prints a null rather than an empty list when there is nothing to report. The reading side therefore keeps whichever shape it was sent instead of tidying it, so the choice is made once, where the two paths meet, rather than hidden in a decoder.
- An expired wait is a distinct outcome carrying the evaluation's identity, not a sentinel and never a verdict, so a caller can tell the user which evaluation is still running and where to read it later.
- A server error cannot be reported to the user in the server's own words, which was assumed possible when the work was planned. Every retryable status is consumed by the shared HTTP client, which exhausts its retries and reports giving up, discarding the response body. So a refused feature flag or a missing trail keeps its sentence, while the enqueue failure the API documents does not, and the command has to supply one of its own. A test pins the loss so nobody plans around it again.
- Every slice shares one branch for the whole issue rather than a branch each, so the history reads as one piece of work and a reviewer can follow the plan through to the code without walking a stack of branches.
- The wait is bounded at the platform's stated enqueue-to-terminal ceiling. Expiring reports an unfinished evaluation carrying its identifier, so a slow evaluation can never be mistaken for a denial.

---

## Next Steps

- [ ] Run the full integration suite once the local test server is available. It needs a production token to pull the server image. The two client-side trail suites are the only part of this change still unverified, and one failure in the HTTP client package already exists on a clean checkout without that server.
- [ ] Check the flag by hand against staging, with an org that has the feature enabled: allow, deny, no-assert, a broken policy, and several trails at once.
- [ ] Time both paths on staging against a trail with many attestations. That is the latency comparison the ticket predicts, and the first number it can produce.
- [ ] Open a ticket for distinct exit codes if anyone ever needs to branch on the outcome. Deliberately not done here; see the decision above.
- [ ] Open a ticket for letting the shared HTTP client take a cancellation, which would let the wait hold its budget exactly and stop a read already in flight. It would serve every command, not just this one.
- [ ] Run the two trail suites once the local test server is available; they are the only part of Slice 0 still unverified.
- [ ] Decide whether the local test stack should later gain a broker, a worker and an evaluator, which would turn the stubbed tests into a real end-to-end check. Not required here, and a ticket of its own.
- [ ] Measure both paths on staging against a trail with many attestations, which is the latency comparison the ticket predicts and the first number it can produce.
