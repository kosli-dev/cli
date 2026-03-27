---
title: "20260320 - Error reporting strategy: exit codes vs structured JSON envelope"
description: "Evaluate whether the CLI should report errors via numeric exit codes, structured JSON envelopes, or a hybrid — informed by codebase analysis and industry research"
status: "Proposed"
date: "2026-03-20"
---

# 20260320 - Error reporting strategy: exit codes vs structured JSON envelope

## Context

[ADR 20260319](./20260319-differentiated-exit-codes.md) introduced structured exit codes (0/1/2/3/4) to replace the uniform exit-1-on-all-errors behaviour. Before merging that work, we want to evaluate whether a fundamentally different approach — structured JSON error output — would serve consumers better, and choose one path before the exit code contract becomes load-bearing.

### What the swarm investigation found

**Error surface analysis** — The CLI has ~80 commands across 6 categories. Attestation commands have the highest error diversity: a single `attest pullrequest` invocation can fail in 5+ distinct ways (flag validation, fingerprint calculation, git access, VCS provider API, Kosli server, compliance assertion). Several commands already emit **multiple errors per invocation** (e.g. a server error followed by an assertion failure). A numeric exit code can only represent one of these.

**Consumption pattern analysis** — Today, all consumers rely on `$?` and plain-text stderr:
- GitHub Actions workflows use `kosli-dev/setup-cli-action@v2` and rely on step-level failure (non-zero = step fails)
- Shell scripts check `$?` (e.g. `bin/test_install_script_over_homebrew.sh`)
- No SDK wrappers or programmatic JSON parsers exist
- `--output json` only affects success output; errors are always plain text on stderr
- The server returns structured JSON errors (`{"message": "...", "errors": [...]}`) but the CLI flattens them to plain strings before printing

**Industry research** — How comparable tools handle it:

| Tool | Exit codes | Structured errors | Notes |
|------|-----------|-------------------|-------|
| curl | ~100 codes | No | Most granular; CI uses `case $?` |
| git | 0, 1, 128, 129 | No | Exit 1 overloaded (error vs "differences found") |
| kubectl | 0, 1 | Partial (K8s Status object on stderr) | Programmatic users prefer client-go |
| gh | 0, 1, 2, 4 | Partial (some JSON on stderr) | Inconsistent — not all errors are JSON |
| AWS CLI | 0, 1, 2, 130, 252-255 | Yes (JSON envelope with `Error.Code`) | Gold standard for programmatic use |
| gcloud | 0, 1, 2 | No | Widely seen as a weakness |
| Terraform | 0, 1, 2 | Yes (`-json` flag covers everything) | Best hybrid: exit code + NDJSON stream |

No RFC or standard exists for CLI error output. The closest references are BSD `sysexits.h` (largely ignored by modern tools) and RFC 9457 (Problem Details for HTTP APIs, adaptable to CLI).

## Decision Drivers

1. **Who consumes errors programmatically today?** Shell scripts and CI pipelines — both understand `$?` natively.
2. **Who might consume errors programmatically in the future?** SDK wrappers, GitHub Actions composites, Terraform providers, or orchestration tools calling `kosli` as a subprocess.
3. **Can a single exit code represent the error surface?** For most commands, yes. For commands with multi-error paths (attest with `--assert`, evaluate), a single code loses information.
4. **What is the migration cost?** Exit codes are already partially shipped (ADR 20260319). JSON envelopes would require changes to every consumer.
5. **What does the Kosli server already return?** Structured JSON errors with `message` and `errors` fields — the CLI currently discards this structure.

## Open Design Questions

These questions apply to any option that uses exit codes (A, C, and the ranged variant A2). They were raised during [PR #714 review](https://github.com/kosli-dev/cli/pull/714#issuecomment-4097867565).

### Should exit codes use spaced ranges for future subdivision?

The current scheme packs codes tightly (1, 2, 3, 4). An alternative uses ranges with gaps:

```
┌───────┬───────────────────┐
│ Range │      Domain       │
├───────┼───────────────────┤
│ 51-59 │ Compliance/policy │
├───────┼───────────────────┤
│ 61-69 │ Server/infra      │
├───────┼───────────────────┤
│ 71-79 │ Auth/config       │
├───────┼───────────────────┤
│ 81-89 │ Usage             │
└───────┴───────────────────┘
```

This would allow later subdivision (e.g. 51 = assertion failed, 52 = approval missing, 53 = artifact not found) without renumbering.

**Trade-off**: More future-proof, but harder to remember and document. CI scripts with `case $?` become range checks (`if [ $? -ge 51 ] && [ $? -le 59 ]`). The compact 1-4 scheme is easier to use today but paints us into a corner if we ever need finer granularity within a category.

**Note**: POSIX reserves exit codes 126-128 for shell use, and 128+N for signal termination. Any scheme must stay below 126.

### Where does HTTP 404 belong?

Currently, 404 responses return a plain `fmt.Errorf` and fall through to exit 1. But "environment not found" and "server broken" are very different for scripts.

Options:
- **Stay in exit 1** (current): 404 is a data/application error — the server responded correctly, the thing just doesn't exist. This is analogous to `git diff` returning 1 for "differences found."
- **Own category**: A "not found" exit code (e.g. 5, or 55 in the ranged scheme) for resources that should exist but don't.
- **Context-dependent**: 404 during `assert artifact` is a compliance signal (exit 1); 404 during `get environment` is a usage/config issue (exit 4).

### Should 429 (rate limiting) be separate from 5xx?

Currently, `retryablehttp` retries both 429 and 5xx, and both ultimately exit 2 (server). But they have different causes:
- 5xx = server broken → escalate to ops
- 429 = client sending too fast → back off and retry

A separate exit code for rate limiting would let CI pipelines add longer backoff instead of escalating.

### Timeouts vs connection refused

Both are currently exit 2 (server). But:
- Connection refused = server is down → check infrastructure
- Timeout = server is slow or network is lossy → retry with longer timeout

Separating these helps operators triage, but adds code proliferation.

### Partial failures in multi-host mode

`snapshot k8s` can operate across multiple environments. If 2 of 3 succeed, what exit code?
- Exit 0 (partial success) loses the failure signal
- Exit 2 (server error) overstates the problem
- A dedicated "partial failure" code could work but adds complexity

This is the strongest argument for JSON envelopes: partial failures need per-item status.

### Dry-run: should it return "what would have been" the exit code?

Currently `--dry-run` always exits 0. An alternative: dry-run exits with the code that *would have* resulted, allowing CI to validate pipeline logic without side effects. This would break the current contract where `--dry-run` is guaranteed safe (exit 0).

### Signal handling (SIGTERM, SIGINT)

If the CLI receives SIGTERM mid-request, should it:
- Exit immediately with 128+signal (standard Unix convention)?
- Attempt graceful shutdown and exit with the original error category?
- This is currently unhandled — the process just dies.

### Exit code stability contract

**This is the most important question.** Should exit codes be stable across minor versions?

If yes: CI pipelines can hardcode `case $?` and rely on semver. Any new code or renumbering is a major version bump.

If no: exit codes are advisory and can change. But then CI pipelines can't depend on them, which undermines the whole point.

**Recommendation**: Treat exit codes as a stable contract from the moment they ship. Document this explicitly. New codes can be added (additive change = minor version), but existing codes must not change meaning (= major version if they do).

## Options Considered

### Option A: Compact exit codes (current implementation from ADR 20260319)

Keep the 0/1/2/3/4 scheme. Errors are plain text on stderr. No structured error output.

**Pros:**
- Already implemented (55 test annotations, 14 files, docs generated)
- Universal — every shell, CI system, and language can check `$?`
- Simple to document: 5 codes, one table per command
- No parsing required for the dominant use case (CI gate: zero vs non-zero)
- `set -e` and GitHub Actions step failure work out of the box

**Cons:**
- Loses context: exit 1 means "compliance failure" but not *which* assertion, *which* artifact, or *what* the policy said
- Multi-error commands (e.g. server error + assertion failure) can only report one exit code
- Programmatic consumers must parse stderr text to get details
- No versioning story — if we add exit code 5 later, existing `case` statements silently miss it
- Tightly packed: no room to subdivide categories without renumbering (breaking change)

### Option A2: Ranged exit codes (spaced scheme)

Same as Option A but with spaced ranges (51-59, 61-69, 71-79, 81-89) to allow future subdivision within each category.

**Pros:**
- All the pros of Option A
- Future-proof: can add 52 = "approval missing", 53 = "artifact not found" without renumbering
- Each range can be checked with a simple bash range test
- Room for the edge cases above (429, timeout, partial failure) in their respective ranges

**Cons:**
- Harder to remember: "what does exit 63 mean?" vs "what does exit 2 mean?"
- More complex `case $?` in shell scripts (range checks vs exact matches)
- YAGNI risk: we may never need the granularity, and the gaps become wasted design space
- No industry precedent — curl uses sequential codes, not ranges
- Still can't represent multi-error scenarios

### Option B: Structured JSON envelope (replacing exit codes)

All commands write a JSON envelope to stdout on both success and failure. Exit code is always 0 (success) or 1 (failure). The envelope carries the error details:

```json
{
  "version": "1",
  "status": "error",
  "exit_code": 3,
  "category": "auth",
  "message": "Invalid API token or unauthorized access",
  "errors": [
    {
      "code": "AUTH_FAILED",
      "message": "401 Unauthorized",
      "detail": "The provided API token is not valid for org 'acme'"
    }
  ]
}
```

**Pros:**
- Rich, machine-readable context: category, code, message, detail, multiple errors
- Versioned schema (`"version": "1"`) — clients can adapt without breaking
- Can represent multi-error scenarios (server error + compliance failure)
- Forward-compatible: adding new fields or error codes doesn't break consumers
- Preserves structured data the server already returns (instead of flattening it)

**Cons:**
- **Breaking change for every consumer**: `$?` stops being meaningful; all scripts must parse JSON
- **`set -e` and CI step failure stop working** if exit is always 0/1 — or the envelope duplicates what the exit code already says
- **Requires a JSON parser** in every consumer context (not always available in minimal CI environments)
- Every command's output contract changes, including success output for commands that currently print plain text
- Human readability suffers — `Error: bad token` becomes a JSON blob unless you add a `--human` flag
- Significantly more implementation work than what's already done

### Option C: Hybrid — exit codes + optional structured JSON errors (recommended evaluation)

Keep exit codes as the primary error signal. Add an opt-in `--output json` mode that wraps **both success and error output** in a structured envelope. When `--output json` is not set, behaviour is identical to Option A.

```
# Default: human-readable, exit code carries the signal
$ kosli assert artifact --fingerprint abc123 ...
Error: Artifact is not compliant
$ echo $?
1

# JSON mode: same exit code, but stderr is structured
$ kosli assert artifact --fingerprint abc123 ... --output json 2>&1
{
  "version": "1",
  "status": "error",
  "exit_code": 1,
  "category": "compliance",
  "message": "Artifact is not compliant",
  "errors": [
    {"code": "COMPLIANCE_VIOLATION", "message": "Artifact is not compliant"}
  ]
}
$ echo $?
1
```

**Pros:**
- Exit codes work today for all existing consumers (no migration)
- Programmatic consumers get structured errors when they opt in
- Versioned schema for the JSON envelope
- Multi-error representation in JSON mode
- Can be added incrementally (start with `list`/`get` commands that already support `--output json`, extend to all commands later)

**Cons:**
- Two error output paths to maintain and test
- Risk of the two paths diverging (JSON says one thing, text says another)
- More implementation surface than Option A alone
- Commands that don't currently support `--output json` need it added

## Analysis

### Mapping the decision to actual consumers

| Consumer | Needs exit codes? | Needs JSON errors? | Notes |
|----------|:-:|:-:|-------|
| GitHub Actions step failure | Yes | No | Non-zero = red step |
| Shell script with `case $?` | Yes | No | Routes to retry/alert/fail |
| `kosli-dev/setup-cli-action` | Yes | No | Wraps CLI in a GH Action |
| Terraform provider | No | No | Already exists; calls the Kosli API directly, not the CLI |
| MCP server wrapping the CLI | Yes | **Yes** | See below |
| Future SDK wrapper | No | Yes | Would want structured errors to map to exceptions |
| Human at terminal | Yes | No | Reads plain text |
| Log aggregation (Datadog, etc.) | No | Yes | Structured fields for indexing |

Exit codes serve **all current consumers**. JSON envelopes serve **near-term future consumers**, most notably an MCP server.

### MCP server as a consumer

An MCP (Model Context Protocol) server wrapping the Kosli CLI would invoke commands as subprocesses and translate results into MCP tool responses. This consumer has specific needs:

- **Needs to distinguish error categories** to return appropriate MCP error types (e.g. retryable vs permanent failure)
- **Needs structured error detail** to include in the MCP response content — an LLM calling the tool needs to understand *why* something failed, not just *that* it failed
- **Cannot reliably parse stderr text** — error message formats may change between CLI versions; an MCP server needs a stable contract
- **Multi-error context is valuable** — if an attestation partially succeeds (reported to server) but assertion fails (compliance), the MCP server should convey both facts to the calling model
- **Exit codes alone are insufficient** — exit 1 tells the MCP server "compliance failure" but not which policy, which artifact, or what the remediation is. The server would need to forward raw stderr as unstructured text, losing the ability to provide structured tool error responses.

This makes the MCP server the most concrete near-term consumer that would benefit from Option C (hybrid). An MCP server could work with exit codes + stderr text today, but structured JSON would make it significantly more reliable and version-resilient.

### Multi-error problem

The strongest argument for JSON envelopes is multi-error commands. Today, `attest pullrequest --assert` can produce:
1. A compliance error (no PRs found) — exit 1
2. A server error (artifact doesn't exist) — exit 2

The CLI currently prints both to stderr but can only exit with one code. With exit codes, the last/most-severe error wins. With JSON, both are captured. However, this affects only ~7 commands (attest with `--assert` + evaluate), not the full ~80.

### Implementation cost

| Option | New code | Test changes | Consumer migration |
|--------|----------|-------------|-------------------|
| A (compact exit codes) | Done | Done (55 annotations) | None |
| A2 (ranged exit codes) | Moderate (renumber + add codes) | All `wantExitCode` annotations change | Scripts using exact code checks break |
| B (JSON envelope only) | Major rewrite | All tests change | All consumers must update |
| C (hybrid) | Moderate (add JSON error path) | Add JSON error tests | None (opt-in) |

### How each option handles the edge cases

| Edge case | A (compact) | A2 (ranged) | B (JSON) | C (hybrid) |
|-----------|:-:|:-:|:-:|:-:|
| 404 not found | Falls to exit 1 | Could get own code (e.g. 54) | Distinct error code in envelope | Both |
| 429 rate limit | Exit 2 (lumped with 5xx) | Could get own code (e.g. 62) | Distinct error code | Both |
| Timeout vs refused | Both exit 2 | Could split (61 vs 62) | Distinct per-error | Both |
| Partial multi-host failure | Last error wins | Last error wins | Per-item status array | Per-item in JSON mode |
| Multi-error (assert + server) | Last error wins | Last error wins | All errors in array | All errors in JSON mode |
| Dry-run exit code | Always 0 | Always 0 | Could include "would_be" field | Could include "would_be" field |

## Decision

**Deferred — this ADR is open for discussion.**

The swarm analysis and [PR #714 review feedback](https://github.com/kosli-dev/cli/pull/714#issuecomment-4097867565) suggest:

1. **Exit codes (Option A) are sufficient for all current consumers.** No consumer today parses error output. The 0/1/2/3/4 scheme gives CI pipelines exactly the routing signal they need.

2. **Ranged exit codes (Option A2) are future-proof but speculative.** The gaps allow subdivision, but no consumer today needs "approval missing" vs "artifact not found" as distinct exit codes. The renumbering cost is non-trivial (all tests, all docs, all CI scripts), and there's no industry precedent for ranged CLI exit codes.

3. **JSON envelopes (Option B) as a full replacement are premature.** There are no programmatic consumers that would benefit, and the migration cost is high. The gh CLI and gcloud both attempted partial JSON errors and ended up with inconsistency — a cautionary tale.

4. **The hybrid (Option C) is the natural evolution path** — and is the only option that cleanly handles partial failures, multi-error commands, and the 404/429/timeout distinctions without proliferating exit codes. But it should be driven by demand, not speculation.

5. **Exit code stability must be an explicit contract.** Whichever option ships first, document that exit codes are stable across minor versions. New codes = minor version bump. Changed codes = major version bump.

**Recommendation**: Merge the compact exit codes (Option A / ADR 20260319) as the foundation. Explicitly document the stability contract. Design the JSON error envelope schema now (so it's ready when needed), but don't implement it until a real consumer demands it. If subdivision within categories becomes necessary, prefer adding JSON output (Option C) over renumbering to ranges (Option A2), since JSON handles the edge cases that drive the subdivision need in the first place.

## Appendix: Draft JSON error envelope schema

For future reference, if/when Option C is implemented:

```json
{
  "version": "1",
  "status": "error | success",
  "exit_code": 1,
  "category": "compliance | server | auth | usage | unknown",
  "message": "Human-readable summary",
  "errors": [
    {
      "code": "COMPLIANCE_VIOLATION | SERVER_ERROR | AUTH_FAILED | USAGE_ERROR",
      "message": "Short description",
      "detail": "Extended context (optional)",
      "source": {
        "flag": "--fingerprint",
        "command": "assert artifact"
      }
    }
  ],
  "metadata": {
    "command": "kosli assert artifact",
    "cli_version": "2.12.0",
    "request_id": "optional-server-request-id"
  }
}
```

Design notes:
- `version` field enables schema evolution without breaking parsers
- `errors` is an array to support multi-error commands
- `exit_code` is included so JSON parsers don't need to also check `$?`
- `source.flag` helps users fix usage errors programmatically
- `metadata.request_id` preserves server-side correlation (currently discarded)
- Inspired by RFC 9457 (Problem Details) and AWS CLI's `Error.Code` pattern

## Related Decisions

- [20260319 - Differentiated exit codes](./20260319-differentiated-exit-codes.md) — the exit code scheme this ADR evaluates
- [20260302 - Client-side policy evaluation](./20260302-client-side-policy-evaluation.md) — the evaluate command, one of the multi-error commands that would benefit most from structured output
