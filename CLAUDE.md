# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Workflow Preferences

- Claude **may** run `git checkout -b`, `git add`, and `git commit` as part of the TDD and slice workflow described below.
- **NEVER** run `git push` or `git push --force`. The user handles all pushing themselves.

## Project Overview

Kosli CLI — a Go command-line tool for recording and querying software delivery events to the [Kosli](https://www.kosli.com) platform. Built with Cobra/Viper, it supports artifact fingerprinting, environment snapshots, and compliance attestations across AWS, Azure, Docker, Kubernetes, and various CI/CD systems.

## Build & Development Commands

```bash
make build                  # Build binary → ./kosli
make lint                   # Run golangci-lint (requires brew-installed golangci-lint)
make fmt                    # Format code (go fmt)
make vet                    # Run go vet
make deps                   # Download and tidy Go modules
```

## Testing

Tests require a local Kosli server running via Docker Compose. The `test_setup` target handles this automatically.

```bash
make test_integration                              # Run tests (--short, skips slow ones, -p=8)
make test_integration_full                         # Run ALL tests including slow K8s tests
make test_integration_single TARGET=TestSuiteName  # Run a single test suite
```

Some tests are skipped without these env vars: `KOSLI_GITHUB_TOKEN`, `KOSLI_GITLAB_TOKEN`, `KOSLI_BITBUCKET_ACCESS_TOKEN`, `KOSLI_AZURE_TOKEN`, `KOSLI_SONAR_API_TOKEN`. Running any tests requires `KOSLI_API_TOKEN_PROD` (reader rights for the `kosli` org on app.kosli.com).

## Architecture

### Entry Point & Command System

- **`cmd/kosli/main.go`** — entry point; initializes logger and Kosli HTTP client, creates root Cobra command
- **`cmd/kosli/root.go`** — root command setup, global flags (api-token, org, host, dry-run, debug, etc.), Viper config binding
- **`cmd/kosli/*.go`** — ~80+ command files, each following the pattern `new<Command>Cmd()` factory function returning a `*cobra.Command`
- **`GlobalOpts`** struct in root.go holds shared config (ApiToken, Org, Host, HttpProxy, DryRun, MaxAPIRetries, etc.)

### Internal Packages (`internal/`)

| Package | Purpose |
|---------|---------|
| `requests` | HTTP client with retryable requests, multipart uploads, dry-run mode |
| `digest` | SHA256 fingerprinting for files, directories, OCI/Docker images |
| `server` | Server-side artifact handling and fingerprinting |
| `aws`, `azure`, `docker`, `kube` | Cloud environment integrations (ECS, Lambda, S3, Azure Apps, K8S) |
| `github`, `gitlab`, `bitbucket` | VCS provider API integrations |
| `jira`, `snyk`, `sonar` | Issue tracking and security scanning integrations |
| `gitview` | Git repository operations (via go-git) |
| `security` | Credentials encryption and keyring storage |
| `logger` | Structured logging with configurable output streams |
| `output` | Table/JSON output formatting |
| `filters` | Data filtering and query operations |

### Configuration

- Config file: `~/.kosli.yml` (YAML)
- All flags can be set via env vars with `KOSLI_` prefix (e.g., `--api-token` → `KOSLI_API_TOKEN`)
- Setting `KOSLI_API_TOKEN=DRY_RUN` enables dry-run mode

### Test Patterns

All command tests in `cmd/kosli/` use the **testify suite** pattern:

- Each test file defines a suite struct embedding `suite.Suite`
- `SetupTest()` creates test flows/trails against the local server (localhost:8001)
- Tests use `cmdTestCase` structs with `cmd` (CLI args string), `golden` (expected output), and `wantError` fields
- Golden file comparisons validate command output
- Test suites are run via `suite.Run(t, new(SuiteStruct))`

### Build Details

- `CGO_ENABLED=0` — static binary
- Version info injected via `-ldflags` (version, git commit, build metadata)
- Multi-platform releases via GoReleaser (darwin/linux/windows × amd64/arm64)
- Semantic versioning; releases created with `make release tag=v<version>`

## Working Style: TDD

Follow a strict Red-Green-Refactor loop for every change:

1. **Check `TODO.md`** — read the current test list for the active slice. If the list is empty, write one before coding.
2. **Pick the next test** — choose the smallest, most informative failing test from the list.
3. **Write the test** — add only the test; run it and confirm it fails (red).
4. **Make it pass** — write the minimum production code to turn the test green.
5. **Refactor** — clean up duplication and improve names while all tests stay green.
6. **Commit** — commit the green state with a message like `green: <what the test proves>`.
7. **Update `TODO.md`** — check off the passing test; note any new tests discovered during the step.

Repeat steps 2–7 until the slice is complete, then commit any remaining cleanup.

### Test list discipline

Maintain the test list in `TODO.md` under the active slice (see Thin Vertical Slices below). Before writing code, brainstorm the tests you expect to write. During the loop, add new tests as you discover them. The list is a living document — it grows and shrinks as understanding deepens.

In this repo, tests are organised using `suite.Run()` for suite-based command tests (the dominant pattern in `cmd/kosli/`) and `t.Run()` for standalone function tests in `internal/` packages.

### Testability

Keep production code testable by following the patterns already established in this repo:

- **Command tests**: `newRootCmd(stdout, stderr, args)` accepts `io.Writer` for output capture; tests use `executeCommandC()` which wires a `bytes.Buffer`. Test cases are `cmdTestCase` structs with `cmd`, `golden`, and `wantError` fields.
- **Logging**: Internal functions accept `logger.Logger` as a parameter rather than using a global.
- **HTTP dependencies**: Tests use `httpfake.HTTPFake` to stub HTTP endpoints.
- **Fixtures**: Test data lives in `testdata/` subdirectories (e.g. `cmd/kosli/testdata/`, `internal/server/testdata/`).

If a function is hard to test, that's a design signal — restructure so dependencies are injected.

## Working Style: Branches and PRs

- Create a feature branch before the first commit: `git checkout -b feature-name`.
- Each feature gets its own branch ready to push to a PR. User decides when to make a PR.
- We can consider rebaseing to squash commits for a slice if needed (see below)

## Working Style: Thin Vertical Slices

Break every feature into the smallest slices that are **independently mergeable and useful**. Each slice must:

- Be a complete vertical cut (test + production code + help text if applicable)
- Leave the codebase in a working state
- Be small enough to review in minutes

### Principles

1. **Start with the thinnest possible end-to-end path** — prove the wiring works before adding polish.
2. **Each slice adds exactly one capability** — one new flag, one new output format, one new validation rule.
3. **Slices build on each other** — later slices refine, extend, or optimise earlier ones.
4. **If a slice feels too big, split it further.**

### Example: Adding a `kosli get trail` subcommand

- **Slice 1**: `kosli get trail` accepts trail name + flow, calls GET endpoint, prints raw JSON — proves the API call and output path.
- **Slice 2**: Format output as a table using the existing `output` package.
- **Slice 3**: Add `--output json` flag for structured output.
- **Slice 4**: Add filtering flags (e.g. `--status`).

### Example: Adding a new field to an existing command

- **Slice 1**: Add the new flag to the command, pass it through to the API request; add a `cmdTestCase` that exercises it.
- **Slice 2**: Add validation for the new flag; add error-case tests.
- **Slice 3**: Update docs/help text.

### Tracking slices in `TODO.md`

Use a fenced-section approach in `TODO.md` at the repo root:

- Each feature is an `## H2` heading — contributors only edit their own section.
- Track slices as a checklist; mark the active slice with `← active`.
- During TDD, nest the test list under the active slice.
- Trim the test list after the slice is merged (keep only the slice title).
- Delete the entire section when the feature merges to main.

### Slice checklist

Before merging a slice, verify:

- [ ] All tests pass (`make test_integration` or `make test_integration_single TARGET=...`)
- [ ] `make lint` passes
- [ ] Does `kosli <command> --help` reflect the change?
- [ ] The slice works independently — no half-finished behaviour is exposed
- [ ] `TODO.md` is updated (test list trimmed, next slice marked active)
