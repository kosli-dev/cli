# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Workflow Preferences

- **NEVER** run, suggest, or offer to run `git add`, `git commit`, or `git push` commands. Do not prompt the user to commit. The user handles all git staging, committing, and pushing themselves.

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
