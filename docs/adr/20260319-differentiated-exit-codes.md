---
title: "20260319 - Differentiated exit codes for the Kosli CLI"
description: "Differentiate the existing uniform exit-1-on-all-errors into structured exit codes that distinguish compliance failures, server errors, config errors, and usage errors"
status: "Proposed"
date: "2026-03-19"
---

# 20260319 - Differentiated exit codes for the Kosli CLI

## Overview

Introduce structured, semantic exit codes to the Kosli CLI so that callers (CI/CD pipelines, shell scripts) can distinguish between a compliance/policy violation and other failure modes without parsing stderr output.

## Context

Prior to this change, the Kosli CLI exits with code `1` for **all** errors. This is because `logger.Error()` in `internal/logger/logger.go` calls `log.Fatalf()`, which unconditionally calls `os.Exit(1)`:

```go
// internal/logger/logger.go — current behaviour
func (l *Logger) Error(format string, v ...interface{}) {
    format = fmt.Sprintf("Error: %s\n", format)
    l.errLog.Fatalf(format, v...)  // always exits with code 1
}
```

In `cmd/kosli/main.go`, every error flows through this path:

```go
if err != nil {
    logger.Error(err.Error())  // exits 1 via Fatalf — always
}
```

This means:
- A script running `kosli assert artifact` gets exit 1 whether the artifact is non-compliant, the server is down, or the API token is wrong.
- A CI pipeline cannot distinguish a network outage (retry-able) from a policy denial (actionable) from a bad token (fix config).
- The only way to differentiate failures is to parse stderr output.

## Decision Drivers

- CI/CD pipelines need to fail fast on compliance violations without parsing stderr
- Operators need to distinguish transient server failures (retry) from auth failures (fix token) from compliance violations (fix the artifact/policy)
- Shell scripts using `$?` should be able to act on the specific class of failure
- The change should preserve the existing exit-1 behaviour for compliance failures (the most common case) to minimise disruption

## Options Considered

### Option A: Keep uniform exit 1 for all errors (status quo)
- **Pro**: No change; no migration needed
- **Con**: Cannot distinguish failure types without parsing stderr; limits CI/CD automation

### Option B: Structured exit codes — 0/1/2/3/4 with semantic meaning (chosen)
- **Pro**: Rich signalling; enables CI/CD to route failures to the right handler; aligns with how tools like `git`, `curl`, and policy engines behave
- **Con**: Scripts checking `[ $? -eq 1 ]` as a proxy for "any error" will miss exit codes 2, 3, 4
- **Mitigation**: Exit 1 remains the most common failure code (compliance violations AND the default for unclassified errors), so most existing `[ $? -eq 1 ]` checks will continue to work for the dominant use case

## Decision

**Option B — structured exit codes.**

| Code | Meaning | Error type |
|------|---------|------------|
| 0 | Success | — |
| 1 | Compliance/policy violation (also: unclassified errors) | `ErrCompliance` |
| 2 | Server unreachable or 5xx | `ErrServer` |
| 3 | Auth/config error (401/403) | `ErrConfig` |
| 4 | CLI usage error (unknown flag, missing required flag, wrong arg count) | `ErrUsage` |

Unknown/unclassified errors fall back to exit 1 (same as before).

Commands affected:
- **Assert commands** (8): `assert artifact`, `assert approval`, `assert snapshot`, `assert pullrequest *` — exit 1 on compliance failure (unchanged)
- **Attest commands with `--assert`** (5): `attest pullrequest *`, `attest jira` — exit 1 when assertion fails (unchanged)
- **Evaluate commands** (2): `evaluate trail`, `evaluate trails` — exit 1 when policy denies (unchanged)
- **All commands**: exit 2 on server/network failure, exit 3 on 401/403, exit 4 on usage errors (NEW — previously these all exited 1)
- Special case: `assert status` exits 0 (responsive) or 2 (unreachable) — never 1

## Consequences

### Positive
- Compliance assertions remain usable as hard gates (exit 1, same as before)
- Operators can write scripts that retry on exit 2, fix config on exit 3, and triage policy violations on exit 1
- Exit code meaning is documented per-command in generated CLI docs (via `commandExitCodes` in `cmd/kosli/docs.go`)
- Dry-run mode is preserved: errors in `--dry-run` still exit 0

### Negative
- Scripts checking `[ $? -eq 1 ]` as a proxy for "any error" will stop catching server errors (now exit 2), auth errors (now exit 3), and usage errors (now exit 4). Scripts using `[ $? -ne 0 ]` are unaffected.
- Whether this warrants a major version bump is debatable — exit code 1 was never documented as meaning "any error", but it was the de facto implicit contract via `Fatalf`.

### Neutral
- Cobra built-in usage errors (unknown flags, missing required flags, wrong argument count) are auto-detected via string pattern matching and classified as `ErrUsage` without requiring explicit wrapping in each command.
- HTTP 4xx errors other than 401/403 (e.g. 400, 404) return plain `fmt.Errorf` and fall through to exit 1 — consistent with treating them as data/compliance failures.

## Implementation

Error types are thin wrappers in `internal/errors/errors.go`:

```go
type ErrCompliance struct{ msg string }
type ErrServer     struct{ msg string }
type ErrConfig     struct{ msg string }
type ErrUsage      struct{ msg string }

func ExitCodeFor(err error) int {
    if err == nil { return 0 }
    var e1 ErrCompliance; if errors.As(err, &e1) { return 1 }
    var e2 ErrServer;     if errors.As(err, &e2) { return 2 }
    var e3 ErrConfig;     if errors.As(err, &e3) { return 3 }
    var e4 ErrUsage;      if errors.As(err, &e4) { return 4 }
    if isCobraUsageError(err)                    { return 4 }
    return 1 // safe default
}
```

`errors.As` correctly unwraps chained errors (`fmt.Errorf("...: %w", ErrCompliance{...})`).

The exit code dispatch in `cmd/kosli/main.go` bypasses `logger.Error()` (which calls `log.Fatalf` → `os.Exit(1)`) and writes to stderr directly:

```go
if err != nil {
    fmt.Fprintf(os.Stderr, "Error: %s\n", err.Error())
    os.Exit(kosliErrors.ExitCodeFor(err))
}
```

This is necessary because ~40 command factory functions call `logger.Error()` for fatal setup errors and rely on `Fatalf` to halt — changing `logger.Error` itself would require auditing all those call sites.

## Review Process

This branch was reviewed by a five-agent code review swarm (Opus 4.6) before the fixes were applied.

### Bugs found and fixed

1. **Exit code dispatch was dead code** — `logger.Error()` in `main.go` called `log.Fatalf` → `os.Exit(1)` before `ExitCodeFor` could run. All errors exited 1 regardless of type. Tests didn't catch this because they call `ExitCodeFor(err)` on the returned error directly, never going through `main()`. Fixed by replacing `logger.Error` with `fmt.Fprintf` in `main.go` only.

2. **Same issue in `innerMain()`** — `logger.Error` at the unknown-subcommand path exited before `return ErrUsage` could execute. Fixed the same way.

3. **HTTP 5xx not classified as `ErrServer`** — In `internal/requests/requests.go`, 5xx responses (when retries are exhausted) were returned as plain `fmt.Errorf`, falling through to exit 1 instead of exit 2. Fixed by adding a `resp.StatusCode >= 500` check.

### Verified after fixes

```
$ ./kosli version                                          # exit 0
$ KOSLI_API_TOKEN=bad ./kosli list environments --org x    # exit 3
$ ./kosli list environments --host http://localhost:1 ...  # exit 2
$ ./kosli list environments --bogus-flag                   # exit 4
```

### Remaining work — test coverage

55 `wantExitCode` annotations now exist across 14 test files, covering all assert, attest `--assert`, and evaluate command families with compliance failure (exit 1), usage error (exit 4), and success regression guard (exit 0) cases.

No end-to-end test verifies the actual compiled binary's process exit code. The current tests validate error classification logic in isolation.

## Related Decisions

- [20260302 - Client-side policy evaluation](./20260302-client-side-policy-evaluation.md) — the evaluate command whose exit behaviour this ADR formalises
