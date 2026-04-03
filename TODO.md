# TODO

<!-- Each feature gets its own ## section below. -->
<!-- Only edit YOUR feature's section. Delete it after merging to main. -->

## Fix: git worktree HEAD resolution

- [x] Slice 1: Enable `EnableDotGitCommonDir` in `gitview.New()` so HEAD resolves in worktrees
  - [x] Test: `New()` succeeds when called from a git worktree path
  - [x] Test: `BranchName()` returns correct branch when called from a worktree
  - [x] Test: `GetCommitInfoFromCommitSHA("HEAD", ...)` works from a worktree

## kosli evaluate trail

- [x] Slice 1: Skeleton `evaluate` parent + `evaluate trail` fetches trail JSON
- [x] Slice 2: `--policy` flag + OPA Rego evaluation
- [x] Slice 3: JSON audit result output + `--format` flag
- [x] Slice 4: `--show-input` flag
- [x] Slice 5: `--attestations` flag + attestation enrichment
  - [x] Slice 5a: Array-to-map transform for `attestations_statuses`
  - [x] Slice 5b: Rehydration
  - [x] Slice 5c: `--attestations` filtering
- [x] Slice 6: Replace `--format` with `--output` flag
- [x] Slice 7: `kosli evaluate trails` (collection mode)
- [x] Slice 8: Make `--policy` required, remove no-policy code path
- [x] Slice 9: Extract shared enrichment pipeline
- [x] Slice 10: Extract shared options struct
- [x] Slice 11: Use `output.FormattedPrint` for output dispatch
- [x] Slice 12: Debug logging for swallowed errors during attestation detail fetching
- [x] Slice 13: Standardise `assert` vs `require` in `transform_test.go`
- [x] Slice 14: Make `--output table` produce actual tabular output
- [x] Slice 15: DRY up command flag registration
- [x] Slice 16: Extract tree-traversal duplication in transform.go
- [x] Slice 17: Align test method naming
- [x] Slice 18: Fail evaluation when rehydration errors occur (instead of silently swallowing them)
- [x] Slice 19: Add Long descriptions, Example blocks, and docs feedback (policy contract hint, snyk trail example)

## kosli evaluate input

- [x] Slice 1: `evaluate input --input-file` with a file path
- [x] Slice 2: stdin support (omit --input-file to read stdin; `-` not supported by cobra)
- [x] Slice 3: help text and examples
- [x] Slice 4: PR review feedback
  - [x] Remove "using OPA" from all evaluate command long descriptions
  - [x] Add test cases for policy validation errors (missing package policy, missing allow rule, deny without violations)
  - [x] Update help text examples with fixture-capture workflow
  - [x] Refactor: use `cmd.InOrStdin()` for testable stdin
  - [x] Refactor: embed `commonEvaluateOptions` to remove flag duplication
- [x] Slice 5: Detect terminal stdin and error when no input is piped

## Add `--params` flag to `kosli evaluate` commands

- [x] Slice 1: `evaluate.Evaluate()` accepts params, passes via OPA data store
- [x] Slice 2: Add `--params` flag across all three commands
- [x] Slice 3: Show params in `--show-input` output
- [x] Slice 4: Update help text and examples

## Fakes & contract tests for cloud provider integrations (#758)

- [x] Slice 1: Define `LambdaAPI` interface and refactor signatures
- [x] Slice 2: Contract test suite against real AWS
- [x] Slice 3: Build `FakeLambdaClient` that passes the contract
- [x] Slice 4: Fake-backed unit tests for filtering and pagination
- [x] Slice 5: Fake-backed unit tests for orchestration
- [x] Slice 6: Trim existing integration tests
- [ ] Slice 7: Package-level factory + fake-backed command tests ← active
  - [ ] Add `NewLambdaClientFunc` factory variable to `internal/aws`
  - [ ] `GetLambdaPackageData` uses factory instead of direct client creation
  - [ ] `snapshotLambda_test.go` injects fake in SetupTest, resets in TearDownTest
  - [ ] Remove `requireAuthToBeSet` from command test cases — they always run
  - [ ] Add Makefile target for AWS smoke tests
