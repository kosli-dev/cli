# Plan: `kosli evaluate trail|trails --server-side` (hidden flag)

> **Ticket:** https://github.com/kosli-dev/server/issues/6700
> **Status:** plan, not started. Written 2026-09-14 against CLI `main` @ `11306cde` and server `main` @ `9239abaf0`.
> **Audience:** the agent or engineer who implements this. Follow the repo's TDD and thin-slice workflow (`CLAUDE.md`). Create a `## feat(evaluate): --server-side` section in `TODO.md` from the slice list below before coding.
> **Out of scope (moved to #6832):** shadow mode. Nothing here runs a server-side evaluation unless the flag is present.

---

## 1. What we build

Today `kosli evaluate trail` and `kosli evaluate trails` fetch the trail, transform it, fetch every attestation one by one, and run the Rego policy in-process (`cmd/kosli/evaluateHelpers.go`, `internal/evaluate/rego.go`).

With the hidden flag `--server-side`, the same commands instead:

1. Read the policy source exactly as today (local file or `http(s)://` URL).
2. `POST /api/v2/evaluations/{org}` with the policy source and the trail references.
3. Poll `GET /api/v2/evaluations/{org}/{id}` until the status is terminal, bounded by a wait budget.
4. Print the verdict with the existing table/JSON printers and drive the exit code from `result.allow` exactly as today.

Without the flag, behaviour is byte-for-byte unchanged.

---

## 2. Server contract (verified in `../server`)

Source of truth: `server/src/fastapi_app/v2/evaluations.py`, `server/src/fastapi_app/models/evaluations.py`, `server/src/fastapi_app/common/evaluations.py`, `server/src/tasks/opa_evaluation_tasks.py`.

### 2.1 Create

```
POST /api/v2/evaluations/{org}
Authorization: Bearer <api token>          (same as every other CLI call)
```

```json
{
  "context": { "trails": [ { "flow": "release", "trail": "my-trail" } ] },
  "policy":  { "files": { "policy.rego": "package policy\n\nallow := true\n" } },
  "params":  { }
}
```

- All models are `extra="forbid"`. Send nothing else. Do **not** send `decision` (that is #6628).
- `context.trails`: 1..100 entries. Duplicate `{flow, trail}` pairs are stored once, not refused.
- `policy.files`: 1..100 entries, relative path → source. Byte cap **1 MiB = sum of (path bytes + source bytes)**. Paths must not be absolute or contain `..`.
- `params`: optional, any JSON object, becomes `data.params`.
- Response `201`:
  ```json
  { "id": "<server id>", "status": "pending", "requested_at": 1757.0, "recorded_at": 1757.0 }
  ```
- Errors (envelope `{"message": "...", "errors": {...}?}`):

  | Code | Meaning | Message shape |
  |---|---|---|
  | 400 | payload validation (cap, path, name regex, extra field) | pydantic errors |
  | 403 | org lacks `is-server-side-evaluation-enabled` | `Server-side evaluation is not enabled for this organization` |
  | 404 | trail(s) not found | `These trails do not exist in org '<org>': <flow>/<trail>, ...` |
  | 404 | endpoint missing (old server) | plain FastAPI 404 |
  | 503 | enqueue refused; a `failed` result with kind `enqueue_failed` is already recorded | `Evaluation '<id>' could not be queued` |

### 2.2 Read

```
GET /api/v2/evaluations/{org}/{id}     200
```

```json
{ "id": "...", "status": "pending|completed|failed",
  "requested_at": 1757.0, "recorded_at": 1757.0,
  "result": { "allow": false, "violations": ["..."] },   // only when completed
  "error":  { "kind": "compile", "message": "..." } }     // only when failed
```

- `response_model_exclude_none=True`: absent fields are **omitted**, never `null`. Branch on `status`, never on field presence.
- **Unknown status values are non-terminal** (the server may add `running` later). Keep polling.
- A denial is `completed` with `allow: false`. It is never `failed`.
- `result` is passthrough from the evaluator: `{"allow": bool, "violations": [string]}`. `violations` may be absent or empty.
- `error.kind` is one of the six evaluator kinds `no_policy`, `entrypoint`, `compile`, `result_shape`, `input_shape`, `evaluate`, or the server's own `enqueue_failed`. Treat the set as open: print any kind verbatim.
- 404: `Evaluation '<id>' does not exist in org '<org>'`.

### 2.3 Timing

- Create P95 target 300 ms. Enqueue-to-terminal ceiling **30 s** (#6621/#6622). Expected ~1 s.
- The Lambda's own evaluation timeout is 50 s and surfaces as kind `evaluate`, so a run that hits it will exceed our 30 s wait and appear to us as "still pending".

### 2.4 Feature flag behaviour in tests

`is_server_side_evaluation_enabled()` returns **True** whenever the server runs `in_cli_tests()` or on localhost. The CLI's local test server therefore **never returns 403**. The 403 path must be tested with a fake HTTP server.

### 2.5 The local CLI test server cannot complete an evaluation

`docker-compose.yml` in this repo runs server, mongo and minio only. There is **no Redis broker, no Celery worker and no OPA Lambda**. The server's Celery app defaults to `redis://localhost:6379/0` with a 30 s `broker_connection_timeout`, so a `POST` against `localhost:8001` will block up to 30 s and then return 503 with an `enqueue_failed` result.

Consequence: **every `--server-side` command test uses an `httptest.NewServer` fake** for the evaluations endpoints, passing `--host <fake url> --max-api-retries 0`. This is the pattern already used by `TestEvaluateTrailRehydrationError` in `cmd/kosli/evaluateTrail_test.go` and is permitted by `docs/adr/20260421-fakes-and-contract-tests.md`. Adding redis + worker + opa-lambda to the CLI compose is a separate follow-up and is not required for this ticket.

---

## 3. Known contract mismatches (do not "fix" in the CLI, document them)

These are the reasons the flag is hidden. Record them in the handover, not in code.

1. **Policy contract.** `validatePolicy` (`internal/evaluate/rego.go:67`) requires `package policy` with an `allow` rule. The Lambda accepts any package and finds the entrypoint from an OPA `entrypoint: true` annotation (single-file bundles need no annotation). Under the flag the CLI **must not** run `validatePolicy`; the server classifies a broken policy as `failed`.
2. **Input shape.** The server input is built from the trail *moment*, not the trail read model:
   - top-level keys: `trails` (always, an array) and `trail` (only when exactly one trail);
   - each element has `moment_number`, `created_at`, `template_id`, `compliance_status`, `flow_name`, `trail_name`;
   - no `events`, `name`, `description`, `git_commit_info`, `user_data`, `origin_url`, `html_url`, `flow` object.
   The CLI's `TransformTrail` output keeps the whole trail document and uses `flow.name` / `name`. **A policy that reads any of the missing fields will behave differently under the flag.** The CLI's own `--show-input` cannot show the server's input (see §4.3).
3. **`--attestations`.** Filtering is client-side only. The create payload has no filter field.
4. **Size caps.** CLI remote policy cap is 5 MiB (`policyMaxBytes`); the server cap is 1 MiB. A policy the CLI reads can be refused by the server with 400.
5. **Exit codes.** Today every failure exits 1 (`logger.Error` → `log.Fatalf`). The ticket wants deny, broken policy, our fault, and still-pending to be distinguishable. That needs a new mechanism (slice 8).

---

## 4. Design decisions (assumptions for the implementer)

These are refinement choices the ticket leaves open. They are chosen here so work can start. Change them only with the ticket owner.

### 4.1 Flag

- Name `--server-side`, `bool`, default `false`, **hidden** via `cmd.Flags().MarkHidden("server-side")` (pattern: `cmd/kosli/attestSbom.go:165`). Add it in a small helper `addServerSideFlag(cmd)` called only from `newEvaluateTrailCmd` and `newEvaluateTrailsCmd`. **Do not add it to `commonEvaluateOptions.addFlags`**, so `evaluate input` never has it.
- Field `serverSide bool` on `commonEvaluateOptions` (it is shared by trail/trails and read in shared helpers).
- Env var `KOSLI_SERVER_SIDE` works automatically through Viper binding. Hidden flags still bind.

### 4.2 Flag interactions (validate in `PreRunE`/`run` before any network call)

| Combination | Behaviour |
|---|---|
| `--server-side --attestations ...` | error: `--attestations is not supported with --server-side` |
| `--server-side --show-input` | error: `--show-input is not supported with --server-side` (server does not return the input) |
| `--server-side --params ...` / `@file` | supported; parsed with `parseParams`, sent verbatim as `params`. `nil` params → send `{}` or omit; the server defaults to `{}`. |
| `--server-side --no-assert` / `--assert` | unchanged semantics; exit code driven by `result.allow` |
| `--server-side --output json\|table` | unchanged printers |
| `--server-side --dry-run` | `kosliClient.Do` returns `(nil, nil)`; print the payload (the client already logs it) and exit 0 without polling |

### 4.3 Policy upload

- Read with the existing `loadPolicy(ref)`. Never run `validatePolicy` under the flag.
- `policy.files` is a single entry. Key = `filepath.Base(ref)` for a local file; for a URL, the last path segment of the URL; fall back to `policy.rego` when the base is empty, `.`, `/` or has no `.rego` suffix. Never send an absolute path or `..`.
- Pre-flight size check: if `len(key) + len(source) > 1 MiB` return `policy bundle is N bytes, over the 1048576 byte limit` before the POST (mirrors the server message so the two paths read alike). Keep `policyMaxBytes` (5 MiB) for the remote read itself.

### 4.4 Wait

- Package `internal/evaluations` (new), no OPA import. Exposes:
  ```go
  type TrailRef struct { Flow, Trail string }
  type CreateRequest struct { Trails []TrailRef; Files map[string]string; Params map[string]interface{} }
  type Evaluation struct { ID, Status string; Result *Result; Error *EvaluationError; RequestedAt, RecordedAt float64 }
  type Result struct { Allow bool; Violations []string }
  type EvaluationError struct { Kind, Message string }
  func (c *Client) Create(ctx, org string, req CreateRequest) (*Evaluation, error)
  func (c *Client) Get(ctx, org, id string) (*Evaluation, error)
  func (c *Client) WaitForTerminal(ctx, org, id string, opts WaitOptions) (*Evaluation, error)
  ```
  `Client` wraps `*requests.Client` plus host and token (so `--dry-run`, retries and proxy behave as everywhere else). Constructed in `cmd/kosli` from `kosliClient` and `global`.
- `WaitOptions{Timeout: 30 * time.Second, Initial: 500 * time.Millisecond, Max: 5 * time.Second}`; exponential backoff doubling. Package-level defaults so tests can shrink them.
- Terminal = `status == "completed" || status == "failed"`. Anything else keeps polling.
- On timeout return a sentinel `ErrStillPending` wrapping the id: `evaluation <id> is still pending after 30s; read it later with GET /api/v2/evaluations/<org>/<id>`. This is never printed as a verdict.
- GET errors during polling: a 5xx/network error is already retried by `retryablehttp` per `--max-api-retries`; after that, fail with the error (our fault), do not keep polling.

### 4.5 Outcome mapping in the command

| Server outcome | CLI output | Exit code (after slice 8; before it every error is 1) |
|---|---|---|
| `completed`, `allow: true` | existing `RESULT: ALLOWED` / JSON | 0 |
| `completed`, `allow: false`, assert (default) | existing `RESULT: DENIED` + violations; error `policy denied` | 1 |
| `completed`, `allow: false`, `--no-assert` | existing output | 0 |
| `failed`, any kind | error `server-side evaluation failed (<kind>): <message>`. **Never prints `DENIED`.** | 2 |
| wait expired, still `pending` | error from `ErrStillPending` with the id | 3 |
| 403 on create | error `server-side evaluation is not enabled for org '<org>' (is-server-side-evaluation-enabled); remove --server-side to evaluate locally` | 4 |
| 404 on create (endpoint missing on old server) | error `this Kosli server does not support server-side evaluation; remove --server-side` | 4 |
| 404 trails not found, 400 validation | error with the server message verbatim | 1 |
| 503 / 5xx / network | error with the server message verbatim | 4 |

Exit codes 2/3/4 are proposals. They are hidden behind a hidden flag, so they can change before publication. The client-side path keeps exit 1 for everything, unchanged.

### 4.6 Reuse of printers

Split `evaluateAndPrintResult` so the printing half takes an `*evaluate.Result` and does not know where it came from. Both paths then share `printEvaluateResult(out, result, outputFormat, showInput, input, params, assertOnDeny)`. Map `evaluations.Result` → `evaluate.Result` in the command (two fields).

---

## 5. Slices

Each slice is one PR-sized change, independently mergeable, with its own test list. All slices share the branch `6700-evaluate-server-side`, one commit or more per slice, so the whole issue stays reviewable as one history. Mark the active slice in `TODO.md`, which is git-ignored and therefore local to your machine.

### Slice 0: refactor the printer seam (no behaviour change)

Goal: `evaluateAndPrintResult` becomes `evaluate` + `printEvaluateResult(out, *evaluate.Result, ...)`.

Tests (all existing; they must stay green):
- [ ] `make test_integration_single TARGET=EvaluateTrailCommandTestSuite`
- [ ] `make test_integration_single TARGET=EvaluateTrailsCommandTestSuite`
- [ ] `make test_integration_single TARGET=EvaluateInputCommandTestSuite`

Files: `cmd/kosli/evaluateHelpers.go`.

### Slice 1: `internal/evaluations` client, `Create` only

Goal: a typed client that POSTs the create payload and decodes the 201 body and the error envelope.

Tests (`internal/evaluations/client_test.go`, `httptest.NewServer`, `t.Run` style):
- [ ] `Create` sends `POST /api/v2/evaluations/{org}` with bearer token and `Content-Type: application/json`
- [ ] payload JSON is exactly `{context:{trails:[{flow,trail}]}, policy:{files:{...}}, params:{...}}` and nothing else (assert with a decoded map and key set)
- [ ] nil params serialise as `{}` (or are omitted; pick one and pin it)
- [ ] 201 body decodes into `Evaluation{ID, Status: "pending", RequestedAt, RecordedAt}`
- [ ] 403 returns a typed `*requests.APIError` (or a wrapped sentinel `ErrNotEnabled`) with the server message
- [ ] 404 and 400 return the server `message` verbatim
- [ ] 503 returns the server `message` verbatim
- [ ] `--dry-run` (client `DryRun: true`) returns `(nil, nil)` and sends nothing

Files: `internal/evaluations/client.go`, `internal/evaluations/client_test.go`.

### Slice 2: `Get` and `WaitForTerminal`

Tests:
- [ ] `Get` decodes `completed` with `result.allow` / `result.violations`
- [ ] `Get` decodes `failed` with `error.kind` / `error.message`
- [ ] `Get` decodes `completed` with no `violations` key → empty slice
- [ ] `WaitForTerminal` returns on the first `completed` after N `pending` responses (fake server with a response sequence)
- [ ] `WaitForTerminal` returns on `failed`
- [ ] `WaitForTerminal` treats an unknown status (`running`) as non-terminal and keeps polling
- [ ] `WaitForTerminal` returns `ErrStillPending` carrying the id when the timeout expires (use tiny `WaitOptions`)
- [ ] `WaitForTerminal` backs off: second interval ≥ first, capped at `Max` (assert on request timestamps loosely, or on a injected sleeper)
- [ ] `WaitForTerminal` stops and returns the error when `Get` returns a non-2xx after retries
- [ ] context cancellation stops the wait

Files: `internal/evaluations/client.go`, `internal/evaluations/wait.go`, tests.

### Slice 3: hidden `--server-side` on `evaluate trail`, happy path

Goal: first end-to-end path. Fake server serves create + get.

Tests (`cmd/kosli/evaluateTrail_test.go`, new suite or new test methods; fake server helper `newFakeEvaluationsServer(t, ...)` in a test helper file so slice 4 can reuse it):
- [ ] `--help` does **not** list `--server-side`
- [ ] `evaluate trail T --flow F --policy allow-all.rego --server-side` → fake receives one POST with `{flow: F, trail: T}` and one file entry keyed `allow-all.rego`; output `RESULT: ALLOWED`; exit 0
- [ ] same with `--output json` → JSON `{allow: true, violations: []}` identical to the client-side shape
- [ ] deny with violations → `RESULT: DENIED`, violations rows, error `policy denied: [...]` (assert default)
- [ ] deny with `--no-assert` → output printed, no error
- [ ] deny with `--output json` → JSON printed then `policy denied`
- [ ] without `--server-side` the fake evaluations endpoint is **never** called (fake asserts zero hits) and the existing client-side tests still pass against localhost:8001
- [ ] `--dry-run --server-side` → exit 0, no GET, payload logged
- [ ] a policy the local path rejects (`testdata/policies/no-package-policy.rego`) is uploaded, not refused, under the flag

Files: `cmd/kosli/evaluateHelpers.go` (new `runServerSide(...)`), `cmd/kosli/evaluateTrail.go`, `cmd/kosli/root.go` (flag help constant `serverSideFlag`), tests.

### Slice 4: `evaluate trails --server-side`

Tests (`cmd/kosli/evaluateTrails_test.go`):
- [ ] `evaluate trails T1 T2 --flow F --policy p.rego --server-side` → **one** POST with two `{flow: F, trail: Tn}` entries in argument order
- [ ] 100 trail names → one POST with 100 entries, accepted
- [ ] 101 trail names → client-side error before any request (`at most 100 trails per server-side evaluation`); pin the message
- [ ] allow / deny / `--no-assert` behave as slice 3
- [ ] duplicate trail names are sent as given (server dedupes); no client error

Files: `cmd/kosli/evaluateTrails.go`, tests.

### Slice 5: flag interaction validation

Tests:
- [ ] `--server-side --attestations x` → error `--attestations is not supported with --server-side`, no request sent
- [ ] `--server-side --show-input` → error `--show-input is not supported with --server-side`, no request sent
- [ ] `--server-side --params '{"a":1}'` → POST `params` equals `{"a":1}`
- [ ] `--server-side --params @testdata/evaluate/params-low-threshold.json` → POST `params` equals the file content
- [ ] `--server-side` with no `--params` → POST `params` is `{}` (pinned from slice 1)
- [ ] `evaluate input --server-side` → cobra `unknown flag` error (flag not registered there)

Files: `cmd/kosli/evaluateHelpers.go`, tests.

### Slice 6: classified failures and server errors

Tests:
- [ ] `status: failed, kind: compile` → stderr/err `server-side evaluation failed (compile): <message>`; stdout contains neither `ALLOWED` nor `DENIED`
- [ ] one test per kind in a table: `no_policy`, `entrypoint`, `compile`, `result_shape`, `input_shape`, `evaluate`, `enqueue_failed`, and an unknown kind `future_kind` (printed verbatim)
- [ ] 403 on create → error names the org, the flag `is-server-side-evaluation-enabled`, and says to remove `--server-side`
- [ ] 404 on create with a non-trail message (endpoint missing) → error `this Kosli server does not support server-side evaluation`
- [ ] 404 on create with `These trails do not exist ...` → that message verbatim
- [ ] 400 (server cap) → message verbatim
- [ ] 503 on create → message verbatim, no polling
- [ ] wait expires → error contains `still pending` and the evaluation id; no verdict printed (shrink wait via package var in the test)

Files: `cmd/kosli/evaluateHelpers.go`, tests.

### Slice 7: policy upload edge cases

Tests:
- [ ] `--policy https://<fake>/policies/pr.rego --server-side` → the CLI fetches and uploads the source under key `pr.rego`; the evaluations fake receives it; the server never receives the URL
- [ ] URL with no file name (`https://host/`) → key `policy.rego`
- [ ] local path `./dir/../p.rego` → key `p.rego` (basename only; never `..`)
- [ ] local policy of 1 MiB + 1 byte → client error `policy bundle is N bytes, over the 1048576 byte limit` before any request
- [ ] a remote policy between 1 MiB and 5 MiB → same client error (fetched, then refused)

Files: `cmd/kosli/evaluateHelpers.go` (a `policyBundleEntry(ref, source)` helper), tests.

### Slice 8: distinct exit codes

Goal: deny (1), broken policy (2), still pending (3), our fault (4) are distinguishable, under the flag only.

Design: add `type exitCodeError struct { code int; err error }` in `cmd/kosli` with `Error()` and `Unwrap()`; `main()` checks `errors.As(err, &exitCodeError{})` and calls `logger.Error` then `os.Exit(code)`. Keep `logger.Error`'s existing `Fatalf` path for every other error so nothing else changes. Note `enrichError` in `main.go` wraps errors: make sure it uses `%w` so `errors.As` still finds the code.

Tests:
- [ ] unit test on `exitCodeFor(err) int`: plain error → 1; deny → 1; `failed` → 2; `ErrStillPending` → 3; 403/404-endpoint/5xx/network → 4
- [ ] `enrichError` preserves the exit code through wrapping
- [ ] client-side deny still yields exit 1 (regression guard)
- [ ] `--dry-run` swallows the error and exits 0 as today (`innerMain` returns nil)

Files: `cmd/kosli/main.go`, `cmd/kosli/exitcode.go`, tests. This slice is optional for the first shadow comparison; land it last.

### Slice 9: wrap-up

- [ ] Run `make lint`, `make test_integration`.
- [ ] Manual check against **staging** with an org that has `is-server-side-evaluation-enabled`: allow, deny, `--no-assert`, broken policy, `trails` with several names. Record wall-clock for a many-attestation trail with and without the flag (first latency comparison the ticket asks for).
- [ ] Update `docs/adr/20260302-client-side-policy-evaluation.md` with a short "Status 2026-09" note pointing at the flag and at §3 above.
- [ ] Create `docs/handover/6700-evaluate-a-trail-server-side-from-the-cli.md` via the `handover` skill; copy §3 and §4 into its Decisions section.
- [ ] Remove the `TODO.md` section when the last slice merges.

---

## 6. Test strategy summary

- **Unit (`internal/evaluations`)**: `httptest.NewServer`, `t.Run`, `testify/require`. No OPA, no Kosli server.
- **Command tests (`cmd/kosli`)**: testify suites as today. For `--server-side` tests use a fake server for **both** the evaluations endpoints and nothing else (the command makes no other calls under the flag). Pass `--host <fake> --max-api-retries 0`. Reuse `cmdTestCase` with `golden`, `goldenRegex`, `goldenJson`, `wantError`.
- **Fake server helper**: one function returning `*httptest.Server` plus a recorder of received requests (method, path, decoded body) and a scripted list of GET responses. Put it in `cmd/kosli/evaluateFake_test.go` so slices 3–7 share it.
- **Do not** attempt a `--server-side` happy path against `localhost:8001` (see §2.5).
- Existing client-side tests must never see the evaluations fake; keep them on localhost:8001 to prove "without the flag, nothing changes".

---

## 7. File touch list

| File | Change |
|---|---|
| `internal/evaluations/client.go` | new: types, `Create`, `Get` |
| `internal/evaluations/wait.go` | new: `WaitForTerminal`, `WaitOptions`, `ErrStillPending` |
| `internal/evaluations/*_test.go` | new |
| `cmd/kosli/evaluateHelpers.go` | split printer; `runServerSide`; validation; bundle entry; size pre-check |
| `cmd/kosli/evaluateTrail.go` | `addServerSideFlag`, branch on `o.serverSide` |
| `cmd/kosli/evaluateTrails.go` | same, plus 100-trail guard |
| `cmd/kosli/root.go` | `serverSideFlag` help constant (`"[hidden] Evaluate on the Kosli server instead of locally."`) |
| `cmd/kosli/main.go`, `cmd/kosli/exitcode.go` | slice 8 only |
| `cmd/kosli/evaluateTrail_test.go`, `evaluateTrails_test.go`, `evaluateFake_test.go` | tests |
| `docs/adr/20260302-client-side-policy-evaluation.md` | status note |
| `docs/handover/6700-evaluate-a-trail-server-side-from-the-cli.md` | via `handover` skill |
| `TODO.md` | slice tracking |

No generated docs change: hidden flags are omitted by cobra and by `kosli docs`.

---

## 8. Open questions for the ticket owner (do not block slices 0–7)

1. Exit code values for broken policy / still pending / our fault (§4.5 proposes 2/3/4).
2. Should `--attestations` with `--server-side` be an error (chosen here) or silently ignored?
3. Should `--show-input --server-side` fetch `GET /api/v2/trails/{org}/{flow}/{trail}/moments/latest` for a single trail instead of erroring? It shows a *later* moment than the one evaluated, so it is misleading; hence the error.
4. Honeycomb wants evaluations split by source (flag vs shadow). The create payload has no `source` field and is `extra="forbid"`. The CLI already sends `User-Agent: Kosli/<version>`; the server can key on that, or #6832 adds a field. Nothing to do in this ticket.
5. Wait budget: 30 s fixed here. A hidden `--server-side-wait` flag is cheap to add later if staging measurements need it.
