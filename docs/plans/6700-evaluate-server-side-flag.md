# Plan: `kosli evaluate trail|trails --server-side` (hidden flag)

> **Ticket:** https://github.com/kosli-dev/server/issues/6700
> **Status:** slices 0 to 7 done, slice 8 deferred by decision, slice 9 wrap-up remaining. Written 2026-09-14 against CLI `main` @ `11306cde`. Boxes are ticked as slices land, and a box proved wrong is struck through rather than deleted.
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

## 2. Server contract

Source of truth: the evaluations endpoints in the Kosli API, as published in its OpenAPI schema. Everything below was verified against the API rather than assumed; this repository is public, so the notes stay at the contract a caller can see.

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

- Every object is closed: an unknown field is a 400, not something ignored. Send nothing else. Do **not** send `decision` (that is #6628).
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
  | 400 | payload validation (cap, path, name regex, unknown field) | per-field validation errors |
  | 403 | org lacks `is-server-side-evaluation-enabled` | `Server-side evaluation is not enabled for this organization` |
  | 404 | trail(s) not found | `These trails do not exist in org '<org>': <flow>/<trail>, ...` |
  | 404 | endpoint not served (an older Kosli) | a 404 carrying no message of the API's |
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

- Absent fields are **omitted**, never `null`. Branch on `status`, never on field presence.
- **Unknown status values are non-terminal** (the server may add `running` later). Keep polling.
- A denial is `completed` with `allow: false`. It is never `failed`.
- `result` is passthrough from the evaluator: `{"allow": bool, "violations": [string]}`. `violations` may be absent or empty.
- `error.kind` is one of the six evaluator kinds `no_policy`, `entrypoint`, `compile`, `result_shape`, `input_shape`, `evaluate`, or the server's own `enqueue_failed`. Treat the set as open: print any kind verbatim.
- 404: `Evaluation '<id>' does not exist in org '<org>'`.

### 2.3 Timing

- Create P95 target 300 ms. Enqueue-to-terminal ceiling **30 s** (#6621/#6622). Expected ~1 s.
- The evaluator has a timeout of its own, longer than our wait, which surfaces as kind `evaluate`. A run that hits it will exceed our wait and appear to us as "still pending".

### 2.4 Feature flag behaviour in tests

The local test server does not exercise the entitlement check, so it **never returns 403** whatever org is used. The 403 path must therefore be tested against a stub.

### 2.5 The local CLI test server cannot complete an evaluation

`docker-compose.yml` in this repo runs the server, its database and object storage only. The pieces that actually run an evaluation -- a queue, a worker and the evaluator -- are not among them, so a create against `localhost:8001` waits for a queue that is not there and then fails.

Consequence: **every `--server-side` command test uses an `httptest.NewServer` fake** for the evaluations endpoints, passing `--host <fake url> --max-api-retries 0`. This is the pattern already used by `TestEvaluateTrailRehydrationError` in `cmd/kosli/evaluateTrail_test.go` and is permitted by `docs/adr/20260421-fakes-and-contract-tests.md`. Adding the queue, worker and evaluator to this repo's compose file is a separate follow-up and is not required for this ticket.

---

## 3. Known contract mismatches (do not "fix" in the CLI, document them)

These are the reasons the flag is hidden. Record them in the handover, not in code.

1. **Policy contract.** `validatePolicy` (`internal/evaluate/rego.go:67`) requires `package policy` with an `allow` rule. The API accepts any package and finds the entrypoint from an OPA `entrypoint: true` annotation (single-file bundles need no annotation). Under the flag the CLI **must not** run `validatePolicy`; the server classifies a broken policy as `failed`.
2. **Input shape.** The server input is built from the trail *moment*, not the trail read model:
   - top-level keys: `trails` (always, an array) and `trail` (only when exactly one trail);
   - each element has `moment_number`, `created_at`, `template_id`, `compliance_status`, `flow_name`, `trail_name`;
   - no `events`, `name`, `description`, `git_commit_info`, `user_data`, `origin_url`, `html_url`, `flow` object.
   The CLI's `TransformTrail` output keeps the whole trail document and uses `flow.name` / `name`. **A policy that reads any of the missing fields will behave differently under the flag.** The CLI's own `--show-input` cannot show the server's input (see §4.3).
3. **`--attestations`.** Filtering is client-side only. The create payload has no filter field.
4. **Size caps.** CLI remote policy cap is 5 MiB (`policyMaxBytes`); the server cap is 1 MiB. A policy the CLI reads can be refused by the server with 400.
5. **Exit codes.** Every failure exits 1, here and everywhere else in the CLI. The ticket wants deny, broken policy and our fault to be distinguishable, and they are, by what each one says rather than by a number. Distinct codes were considered and deferred; see slice 8 for why.

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
| dry run with `--server-side` | The create call returns nothing, so the payload is logged and the command exits 0 without polling. **There is no `--dry-run` flag on the evaluate commands**: it is added per command and these only ever read until now. Dry run is reached by setting the API token to the `DRY_RUN` sentinel, which the root command turns into dry-run mode. Do not add the flag here; that widens a published command's surface for a hidden feature |

### 4.3 Policy upload

- Read with the existing `loadPolicy(ref)`. Never run `validatePolicy` under the flag.
- `policy.files` is a single entry. Key = `filepath.Base(ref)` for a local file; for a URL, the last path segment. Fall back to `policy.rego` only when that leaves nothing usable, which is an empty path, a bare `.`, a bare `..` or a bare `/`. Never send an absolute path or `..`.
- **No extension rule.** An earlier draft here said to fall back when the name does not end in `.rego`. That was wrong, and checking settled it: every bundle entry is parsed as a Rego module whatever it is called, and the name only labels the error messages. Imposing an extension would rename a user's file for no reason and make the server's errors cite something the user does not have on disk.
- Pre-flight size check: if `len(key) + len(source) > 1 MiB` return `policy bundle is N bytes, over the 1048576 byte limit` before the POST (mirrors the server message so the two paths read alike). Keep `policyMaxBytes` (5 MiB) for the remote read itself.

### 4.4 Wait

- Package `internal/evaluations` (new), no OPA import. **Built in slice 1**, actual shape:
  ```go
  const StatusPending, StatusCompleted, StatusFailed = "pending", "completed", "failed"
  type TrailRef struct { Flow, Trail string }
  type CreateRequest struct { Trails []TrailRef; Files map[string]string; Params map[string]interface{} }
  type Evaluation struct { ID, Status string; RequestedAt, RecordedAt float64; Result *Result; Failure *Failure }
  func (e *Evaluation) IsTerminal() bool
  type Result struct { Allow bool; Violations []string }
  type Failure struct { Kind, Message string }
  func NewClient(httpClient *requests.Client, host, token string, dryRun bool) *Client
  func (c *Client) Create(org string, req CreateRequest) (*Evaluation, error)
  func (c *Client) Get(org, id string) (*Evaluation, error)                                  // slice 2
  func (c *Client) WaitForTerminal(ctx, org, id string, opts WaitOptions) (*Evaluation, error) // slice 2
  ```
  `Client` wraps `*requests.Client` plus host, token and the dry-run setting (so `--dry-run`, retries and proxy behave as everywhere else). Constructed in `cmd/kosli` from `kosliClient` and `global`.
- Two names differ from the original sketch. The failure type is `Failure`, not `EvaluationError`: it does not implement `error`, and a struct field spelled `Error` that is not an error misleads every reader. And `Create` takes no context, because the shared HTTP client accepts none and an ignored context would promise a cancellation that does not happen; the context arrives in slice 2, where it genuinely controls the poll loop.
- `WaitOptions{Timeout: 30 * time.Second, Initial: 500 * time.Millisecond, Max: 5 * time.Second}`; exponential backoff doubling. Package-level defaults so tests can shrink them.
- Terminal = `status == "completed" || status == "failed"`. Anything else keeps polling.
- On timeout return a sentinel `ErrStillPending` wrapping the id: `evaluation <id> is still pending after 30s; read it later with GET /api/v2/evaluations/<org>/<id>`. This is never printed as a verdict.
- GET errors during polling: a 5xx/network error is already retried by `retryablehttp` per `--max-api-retries`; after that, fail with the error (our fault), do not keep polling.

### 4.5 Outcome mapping in the command

**Every failure exits 1, as it always has.** The outcomes are told apart by what they say, not by a number, and the distinct codes originally sketched here are deferred; see slice 8.

| Server outcome | CLI output | Exit code |
|---|---|---|
| `completed`, `allow: true` | existing `RESULT: ALLOWED` / JSON | 0 |
| `completed`, `allow: false`, assert (default) | existing `RESULT: DENIED` + violations; error `policy denied` | 1 |
| `completed`, `allow: false`, `--no-assert` | existing output | 0 |
| `failed`, any kind | error `server-side evaluation failed (<kind>): <message>`. **Never prints `DENIED`.** | 1 |
| wait expired, still `pending` | error naming the evaluation and saying it is still pending. **Never a verdict.** | 1 |
| 403 on create | error naming the org, the feature flag, and how to carry on without the flag | 1 |
| 404 on create (endpoint missing on old server) | error `this Kosli server does not support server-side evaluation; remove --server-side ...` | 1 |
| 404 trails not found, 400 validation | error with the server message verbatim | 1 |
| 503 / 5xx / network | error with a sentence of our own. The server's message is **not** available, see below | 1 |

**A 5xx never carries the server's message**, proven in slice 1 and pinned by a test. Any retryable status, which is every 5xx plus 429 and 409, is consumed by the shared HTTP client: it exhausts the retries and reports giving up, discarding the response body. So a 400, 403 or 404 arrives as an API error carrying the server's sentence, and a 503 arrives as a plain error reading `giving up after N attempt(s)`. Two consequences. The enqueue failure the API documents at 503 cannot be shown to a user in its own words, so slice 6 must supply a sentence of its own. And a 503 is retried `--max-api-retries` times by default even though the evaluation behind it is already recorded as failed, which wastes the user's wall clock; leave that alone unless it shows up in practice, since it is behaviour of the shared client rather than of this command.

### 4.6 Reuse of printers

Done in slice 0. Both paths share `printEvaluateResult(out, result, input, outputFormat, showInput, params, assertOnDeny)`, which takes an `*evaluate.Result` and does not know where it came from. Slice 3 maps `evaluations.Result` onto it, two fields.

**The empty-violations shape decides whether the two paths print the same page.** Verified by running the binary: with nothing to report the local path prints `"violations": null`, because its collector returns a nil slice and a nil slice marshals to null. So the mapping must send an empty or absent list through as nil, not as `[]string{}`, or JSON output differs between the paths while the verdict agrees. Table output is unaffected either way. Slice 2 deliberately left the decoded shape alone so this choice is made once, here, where both paths meet.

---

## 5. Slices

Each slice is one PR-sized change, independently mergeable, with its own test list. All slices share the branch `6700-evaluate-server-side`, one commit or more per slice, so the whole issue stays reviewable as one history. Mark the active slice in `TODO.md`, which is git-ignored and therefore local to your machine.

### Slice 0: refactor the printer seam (no behaviour change)

Goal: `evaluateAndPrintResult` becomes `evaluate` + `printEvaluateResult(out, *evaluate.Result, ...)`.

**Done.** The payload build, marshal and format dispatch moved into `printEvaluateResult`, which takes a verdict instead of a policy reference. `evaluateAndPrintResult` kept its signature, so no caller changed.

Tests (all existing; they must stay green):
- [ ] `make test_integration_single TARGET=EvaluateTrailCommandTestSuite` — **not run**, needs the local server
- [ ] `make test_integration_single TARGET=EvaluateTrailsCommandTestSuite` — **not run**, needs the local server
- [x] `make test_integration_single TARGET=EvaluateInputCommandTestSuite` — covers every branch of the moved code: allow, deny, violations present and absent, table, json, show-input, show-input with params, assert and no-assert
- [x] `make lint` clean, build and vet clean

Files: `cmd/kosli/evaluateHelpers.go`.

### Slice 1: `internal/evaluations` client, `Create` only

Goal: a typed client that POSTs the create payload and decodes the 201 body and the error envelope.

**Done.** No sentinel error type was needed: `*requests.APIError` already carries the status code, so a caller distinguishes a refused feature flag from a missing trail without one.

Tests (`internal/evaluations/client_test.go`, `httptest.NewServer`, `t.Run` style):
- [x] `Create` sends `POST /api/v2/evaluations/{org}` with bearer token and `Content-Type: application/json`
- [x] payload JSON is exactly `{context:{trails:[{flow,trail}]}, policy:{files:{...}}, params:{...}}` and nothing else (assert with a decoded map and key set)
- [x] absent params serialise as `{}`, pinned; null would fail the server's validation
- [x] several trails go in one request, in the order given
- [x] 201 body decodes into `Evaluation{ID, Status: "pending", RequestedAt, RecordedAt}` with no result and no failure
- [x] 403 returns a `*requests.APIError` carrying the status and the server message
- [x] 404 and 400 return the server `message` verbatim
- [x] ~~503 returns the server `message` verbatim~~ — **disproved.** A 5xx is consumed by the retry layer and its body is discarded; a test now pins that loss instead. See §4.5.
- [x] `--dry-run` (client `DryRun: true`) returns `(nil, nil)` and sends nothing

Files: `internal/evaluations/client.go`, `internal/evaluations/client_test.go`.

### Slice 2: `Get` and `WaitForTerminal`

**Done.** Two shape changes from the sketch. `ErrStillPending` is a typed `*StillPendingError` rather than a sentinel, so it carries the org, the id and how long was waited and the caller can name the evaluation that is still running; match it with `errors.As`. And the wait needs no injected clock: the budget is a real deadline raced against the poll in a `select`, while the backoff arithmetic is a pure function tested on its own.

Tests:
- [x] `Get` decodes `completed` with `result.allow` / `result.violations`
- [x] `Get` decodes `failed` with `error.kind` / `error.message`
- [x] ~~`Get` decodes `completed` with no `violations` key → empty slice~~ — **deliberately not normalised.** Both shapes are pinned instead, absent staying absent and empty staying empty, because the local path prints an absent list as `null`. Reconciling them is slice 3's job, see below.
- [x] `WaitForTerminal` returns on the first `completed` after N `pending` responses (fake server with a response sequence)
- [x] `WaitForTerminal` returns on `failed`
- [x] `WaitForTerminal` treats an unknown status (`running`) as non-terminal and keeps polling
- [x] `WaitForTerminal` returns `*StillPendingError` carrying the id when the timeout expires, with no evaluation alongside it
- [x] `WaitForTerminal` backs off: `nextInterval` doubles and caps, tested as a pure function rather than by timing
- [x] `WaitForTerminal` stops and returns the error when `Get` returns a non-2xx after retries
- [x] context cancellation stops the wait, before the first read rather than after it
- [x] `WaitOptions{}` falls back to the 30 s budget
- [x] `go test -race -count=5`, `golangci-lint`, vet and build clean

Files: `internal/evaluations/client.go`, `internal/evaluations/wait.go`, tests.

### Slice 3: hidden `--server-side` on `evaluate trail`, happy path

Goal: first end-to-end path. Fake server serves create + get.

**Done.** The tests live in their own suite in `cmd/kosli/evaluateServerSide_test.go`, not in the existing trail suite, and that suite has no setup step. The trail suites build a flow and a trail on the local server before every test, so a test added there could not run without it; this one needs nothing but its own stub and runs anywhere. Slices 4 to 7 extend the same file.

Tests:
- [x] `--help` does **not** list `--server-side`, while the rest of the help still renders
- [x] `evaluate trail T --flow F --policy allow-all.rego --server-side` → fake receives one POST with `{flow: F, trail: T}` and one file entry keyed `allow-all.rego`; output `RESULT: ALLOWED`; exit 0; the trail endpoint is never called
- [x] same with `--output json` → identical bytes to the client-side shape
- [x] deny with violations → `RESULT: DENIED`, violations rows, error `policy denied` (assert default)
- [x] deny with `--no-assert` → output printed, no error
- [x] deny with `--output json` → JSON printed then `policy denied`
- [x] without `--server-side` no evaluation is created, and the trail is read locally exactly as before
- [x] dry run → exit 0, nothing created, nothing polled. Reached through the `DRY_RUN` API token, since these commands carry no `--dry-run` flag; see §4.2
- [x] a policy the local path rejects (`testdata/policies/no-package-policy.rego`) is uploaded, not refused, under the flag
- [x] an allow with no violations prints the same JSON as the local path does, `"violations": null`, whether the server sent an empty list or none at all (see §4.6)

Files: `cmd/kosli/evaluateHelpers.go` (`evaluateServerSide`, `serverVerdict`, `policyBundleKey`, `addServerSideFlag`), `cmd/kosli/evaluateTrail.go`, `cmd/kosli/root.go` (flag help constant `serverSideFlag`), `cmd/kosli/evaluateServerSide_test.go`.

A minimal guard for a `failed` evaluation landed here too, because a failure carries no result and printing one would dereference nothing. It names the kind and the message and never prints a verdict. Slice 6 owns the per-kind tests.

### Slice 4: `evaluate trails --server-side`

**Done.** The trail ceiling is checked in the shared entry point rather than in the multi-trail command, so a limit that belongs to the API is stated once and cannot drift between the two commands.

Tests (in the slice 3 suite):
- [x] `evaluate trails T1 T2 T3 --flow F --policy p.rego --server-side` → **one** POST with the entries in argument order, and no trail read of our own
- [x] 100 trail names → one POST with 100 entries, accepted
- [x] 101 trail names → refused before anything is sent, naming the limit and the count
- [x] allow, deny and `--no-assert` behave as slice 3
- [x] duplicate trail names are sent as given, since the server stores a repeat once rather than refusing it
- [x] the flag is hidden on **both** trail commands
- [x] `evaluate input --server-side` is an unknown flag, since that command names no trails

Files: `cmd/kosli/evaluateTrails.go`, `cmd/kosli/evaluateHelpers.go` (`maxServerSideTrails`), `cmd/kosli/evaluateServerSide_test.go`.

Note that the last box was slice 5's, and is ticked here because the flag reaching `evaluate input` is a property of where the flag is registered, which this slice finished. Slice 5 still owns the combinations that need refusing.

### Slice 5: flag interaction validation

**Done.** The two refusals are explicit errors rather than a cobra exclusion group. Cobra's own message states only that two flags conflict; each of these says *why*, which is what someone reaching for an undocumented flag actually needs. They live in the shared entry point, so the two commands cannot drift apart.

Tests:
- [x] `--server-side --attestations x` → refused, naming the flag, nothing sent
- [x] `--server-side --show-input` → refused, naming the flag, nothing sent
- [x] the same refusal applies on the multi-trail command
- [x] `--server-side --params '{"a":1}'` → POST `params` equals it
- [x] `--server-side --params @testdata/evaluate/params-low-threshold.json` → POST `params` equals the file content
- [x] `--server-side` with no `--params` → POST `params` is `{}`
- [x] unreadable `--params` is refused before anything is sent
- [x] `evaluate input --server-side` → cobra `unknown flag` error (ticked in slice 4, where the flag registration was finished)

Files: `cmd/kosli/evaluateHelpers.go` (`refuseWhatTheServerCannotDo`), `cmd/kosli/evaluateServerSide_test.go`.

### Slice 6: classified failures and server errors

**Done.** The rule is that the server's own words are passed on unchanged wherever it explained itself, and only two cases get wording of their own: a refused feature flag, whose message says nothing about what to do next, and a server too old to serve the route, which sends no words at all.

**How an old server is told apart from a missing trail**, both being 404. Verified by probing the client rather than assumed: a Kosli refusal always carries a message field, while an unrouted request does not, and the shared client renders the latter as a Go map. So the test is whether the message field survived, which keys off the error envelope rather than off any English wording. If the shared client ever stops rendering an envelope-less body that way, `TestAnOlderServerIsNamedAsSuch` fails, which is the point of pinning it.

Tests:
- [x] one test per kind: `no_policy`, `entrypoint`, `compile`, `result_shape`, `input_shape`, `evaluate`, `enqueue_failed`, and an unknown `future_kind`. Each names the kind and the message, and prints neither `ALLOWED` nor `DENIED`
- [x] a `failed` evaluation carrying no reason at all still names the evaluation and prints no verdict
- [x] 403 on create → names the org, the feature flag, and how to carry on regardless
- [x] 404 with no message field → `this Kosli server does not support server-side evaluation`, with no internal rendering leaking into it
- [x] 404 with `These trails do not exist ...` → that message verbatim
- [x] 400 over the byte cap → message verbatim
- [x] ~~503 on create → message verbatim~~ — **impossible**, per slice 1. It gets a sentence of our own, and a test asserts the server's words are gone rather than pretending otherwise
- [x] wait expires → names the evaluation, says still pending, prints no verdict. The budget is shrunk through the exported default

Files: `cmd/kosli/evaluateHelpers.go` (`serverSideRequestError`, `hasKosliErrorEnvelope`), `cmd/kosli/evaluateServerSide_test.go`.

### Slice 7: policy upload edge cases

**Done.** The naming was already right from slice 3, so only the size cap was new. The extension rule in §4.3 turned out to be unnecessary and is struck out there.

Tests:
- [x] `--policy http://<fake>/policies/pr.rego --server-side` → the CLI fetches and uploads the source under key `pr.rego`, and the URL never travels
- [x] URL with no file name → key `policy.rego`
- [x] a local path that climbs out and back, and one with a leading dot → base name only, with no `..` and no `/` in any bundle key
- [x] a local policy one byte over → refused before any request, naming the limit
- [x] a local policy exactly at the cap → accepted, which pins that the name counts toward it as the API counts
- [x] a remote policy the 5 MiB fetch allows but the 1 MiB cap does not → fetched, then refused before any request

Files: `cmd/kosli/evaluateHelpers.go` (`policyBundle`, `serverPolicyMaxBytes`), `cmd/kosli/evaluateServerSide_test.go`.

### Slice 8: distinct exit codes — DEFERRED, not done

**Decided against doing this here.** Every failure keeps exit code 1, which is what the CLI has always done.

Why it was dropped rather than built:

- **It is not this ticket's change to make.** The whole CLI has exactly one failure exit path, the logger's error call, which ends in a fatal log and exits 1. Nothing anywhere chooses a code. Introducing one invents a convention that every future command inherits, through the single function every command exits by. That is a far wider blast radius than the seven slices before it, all of which stayed inside the evaluate commands.
- **The requirement is already met, in the words.** The ticket asks that a denial, a broken policy and a fault of ours be distinguishable. They are: slice 6 gives each its own sentence, and tests assert that a broken policy and an expired wait never print a verdict. Distinct codes would only serve a script branching on them, and nothing can branch on a flag that is hidden and unpublished.
- **Nothing this ticket exists to measure depends on it.** The number wanted here is whether the two evaluation paths agree, and that is read off the verdicts, not off exit codes.
- **Code 1 for a denial is already promised.** `evaluate` and `evaluate input` both state in their help text that a denial exits with code 1, so that one was never free to change anyway.

If it is wanted later it should be its own ticket, deciding the convention for the CLI as a whole rather than for one hidden flag. The sketch that was here, an error type carrying a code plus a check in `main`, is a reasonable starting point, and `enrichError` already wraps with `%w` so a code would survive it.

### Slice 9: wrap-up

- [x] `make lint` clean across the whole repository.
- [x] Every test that can run on a machine without the local stack passes: the whole of `internal/evaluations`, the offline command suites, and `internal/evaluate`.
- [ ] **`make test_integration` has not been run.** It needs the local Kosli server, which needs a production API token to pull the server image. One test in `internal/requests` already fails without it on a clean checkout, so that failure is not from this work. The two client-side trail suites are the part of this change it would exercise, and they are the only thing still unverified.
- [ ] Manual check against **staging** with an org that has `is-server-side-evaluation-enabled`: allow, deny, `--no-assert`, broken policy, `trails` with several names. Record wall-clock for a many-attestation trail with and without the flag (first latency comparison the ticket asks for). Cannot be done from here; needs staging credentials.
- [x] `docs/adr/20260302-client-side-policy-evaluation.md` carries a status note naming the flag and both contract mismatches.
- [x] `docs/handover/6700-evaluate-a-trail-server-side-from-the-cli.md` exists and carries every decision.
- [ ] Remove the `TODO.md` section when the work merges. It is git-ignored, so it is local to whoever did the work.

---

## 5b. Review findings

A code review of the finished branch raised findings in two rounds. All but two were fixed; those two cannot be, without changing behaviour every command inherits.

**Fixed**

1. **A 404 without a Kosli message was only recognised when the body was JSON.** The tell used was the Go map rendering, which the shared client produces only for a JSON body carrying no message field. A body that is not JSON at all, such as an HTML page from a proxy or an empty response, leaves the decoder's complaint instead, so the old-server branch never ran and the user saw `invalid character '<' looking for beginning of value`. Both renderings are now treated as "no message from the server", and three bodies are pinned by tests.
2. **The verdict was chosen by the presence of a result rather than by the status.** An evaluation reporting a failure while also carrying a result would have printed as a verdict, which is the one outcome none of this may produce. Status decides now, and a test sends exactly that answer.
3. **The refusal at 403 threw the server's reason away** and always blamed the feature flag, so a token without rights on the org was sent after the wrong thing. Both reasons travel now.
4. **An accepted evaluation with no id was polled anyway**, fetching a different resource and then blaming the answer. It is refused with a clear message, and nothing is read.

**Fixed in a second pass, after review on the pull request**

5. **The envelope sniff still missed a shape**, a body that is a bare JSON string, which arrives as a plain word and looked like a sentence the server wrote. Replaced entirely: the shared client now records whether the body carried a message field, at the one place that can know, and the 404 branch reads that fact instead of guessing from how the message was rendered. The string coupling between two packages is gone, and four body shapes are pinned.
6. **An expired wait was dressed as a transport failure.** It reached the user behind a sentence saying Kosli could not be reached, over one saying Kosli answered and the evaluation is still running. It now passes through untouched, and the test asserts the transport wording is absent rather than merely that the right words appear somewhere.
7. **A read could have panicked on a suppressed request.** Unreachable today, guarded anyway.
8. **Policy parameters are parsed before the policy is fetched**, so a typo in them no longer costs a remote fetch first.
9. **The wait budget's test seam moved out of the package default.** A test was reassigning an exported variable in another package, which is safe only while nothing runs in parallel and leaves a shortened budget behind if a restore is ever missed.
10. **Both mirrored limits now name the server constant they copy**, so drift is findable.

**Caught by CI, not by me**

13. **A new flag has to be declared to the empty-flag audit.** `TestEmptyFlagAuditCoversEveryCommandAndFlag` walks the cobra tree and compares it against `cmd/kosli/testdata/empty-flag-audit-coverage.json`, so adding a flag turns the build red until that fixture is regenerated with `UPDATE_AUDIT_COVERAGE=1`. The audit in `hack/empty-flag-audit/` then refuses to run until `spec.json` covers it too, which is not checked by CI but would have been left broken for the next person.

    The flag is registered there with the value `false`, not `true`. The audit stops any combination that waits more than five seconds, and a real server-side evaluation needs the queue, worker and evaluator its server does not run, so `true` would break every audit run. The cost is that the audit exercises the flag's parsing rather than its behaviour, which is the same gap as everything else waiting on a fuller local stack.

    **Process note:** I only ever ran the evaluate suites in `cmd/kosli`, never the whole package, so this reached CI. Running the package once before pushing would have caught it in seconds.

**Third review round**

14. **The refusal branch had the very fault the 404 branch had just been fixed for.** It quoted the server's message unconditionally, so a refusal from a proxy in front of Kosli would have shown a decoder complaint dressed as the server's reason, alongside a feature-flag hint that was not the cause. It now reads the same recorded fact the 404 branch does, and four body shapes are pinned.
15. **An expired wait reported the budget rather than the time spent.** The two can differ, as this code's own comments say, so a read that overran by ten seconds would have understated the wait to the one person asking whether the platform met its promise. It is measured now.
16. **A non-text message panicked the whole CLI.** Pre-existing in the shared client, and confirmed by probe for a null, a number and an object. Only a message that is text is now read as one, which also makes the recorded fact honest: a body carrying the key with something unusable under it has no sentence to pass on. A stack trace in place of an error is worse than anything else on this list, and it was two lines.

**Fourth review round**

17. **A message key is not a sentence.** The recorded fact says only that the key was there, so a body carrying it empty, or one whose text is trimmed away entirely by the client's own phrase-stripping, passed the check and printed a stray colon and full stop where the reason should have been. Both routes confirmed by probe. The refusal now asks whether there is anything to quote, not whether a key existed.
18. **A refusal with no message printed as nothing at all**, since an API error renders as its message alone. The status is now named when there is nothing else to say.
19. **The 404 wording claimed to know which kind of 404 it was.** An unmatched route and a route that has since moved look identical from here, and a sentence that must be right about the difference is one a support thread quotes back. It now names both possibilities and still says what to do.

**Fifth review round**

20. **A failure to read the verdict back was worded for the create, and dropped the evaluation.** Once the create has succeeded the evaluation exists and the server holds its answer, so a refusal blamed a feature flag that had already let the create through, and a 404 denied support for the route the create had just used. Worse, every read failure except an expired wait discarded the id, which is the one thing that makes the outcome recoverable. Read failures now go through their own mapping, which names the evaluation and passes an expired wait through untouched.
21. **The measured wait was reported to the nanosecond.** Rounding it to the millisecond keeps the overrun visible, which is why it is measured, without putting six digits of noise in a sentence whose point is that the budget was spent.

**Sixth round, from a human reviewer**

22. **The error mapping had grown to twenty-nine lines of branching, and mostly did not show the status code.** Each branch was reasonable when the previous round asked for it; the accumulation was not. It is fourteen lines now and reports the status the API answered with plus whatever it said about why. Six tests collapsed into one table.
23. **The refusal no longer names the feature flag.** This was a stated requirement of the ticket, and dropping it was the reviewer's call, not an oversight. The consequence is that an organisation without the entitlement sees the API's own sentence and the status, with no hint that a hidden flag exists or that removing it evaluates locally instead. Worth raising with the ticket owner rather than leaving buried here.
24. **A message the API did not write is still not quoted.** A proxy's page, or the decoder's complaint about one, would otherwise print as the server's reason. One condition, not a branch.

**Known limitations, recorded rather than fixed**

11. **The wait budget governs the polling, not a single read.** A read carries no context and the shared HTTP client sets no overall deadline, so a very slow server can overrun the budget by one read, and cancelling stops the loop only at its next turn. Holding to the budget exactly means giving the read a deadline of its own, which needs `internal/requests` to accept a context. That is a change every command inherits, so it belongs in its own ticket rather than here. Stated on `WaitOptions` so nobody reads the budget as a hard bound.
12. **Creating an evaluation is not idempotent and the shared client retries it.** A create that succeeded but whose answer was lost is sent again, so one command can leave more than one evaluation behind. They duplicate each other rather than disagreeing, since each is deterministic for the same policy and instant, so the cost is wasted work rather than a wrong answer. Preventing it needs a caller-supplied key, which the API does not take. Stated on `Create`.

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
| `cmd/kosli/evaluateTrail_test.go`, `evaluateTrails_test.go`, `evaluateFake_test.go` | tests |
| `docs/adr/20260302-client-side-policy-evaluation.md` | status note |
| `docs/handover/6700-evaluate-a-trail-server-side-from-the-cli.md` | via `handover` skill |
| `TODO.md` | slice tracking |

No generated docs change: hidden flags are omitted by cobra and by `kosli docs`.

---

## 8. Open questions for the ticket owner (do not block slices 0–7)

1. ~~Exit code values for broken policy, still pending and our fault.~~ Settled: every failure keeps exit code 1 and slice 8 is deferred. Reopen as its own ticket if a caller ever needs to branch on the outcome.
2. Should `--attestations` with `--server-side` be an error (chosen here) or silently ignored?
3. Should `--show-input --server-side` fetch `GET /api/v2/trails/{org}/{flow}/{trail}/moments/latest` for a single trail instead of erroring? It shows a *later* moment than the one evaluated, so it is misleading; hence the error.
4. Honeycomb wants evaluations split by source (flag vs shadow). The create payload has no `source` field and is `extra="forbid"`. The CLI already sends `User-Agent: Kosli/<version>`; the server can key on that, or #6832 adds a field. Nothing to do in this ticket.
5. Wait budget: 30 s fixed here. A hidden `--server-side-wait` flag is cheap to add later if staging measurements need it.
