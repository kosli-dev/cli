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

## Add `--assert` / `--no-assert` to kosli evaluate commands

Goal: make evaluate commands a policy decision point — print the verdict but
let callers choose whether a deny becomes a non-zero exit. Today's default
stays "assert" (non-zero on deny); the next major release flips the default
to "no-assert" by changing one line.

- [x] Slice 1: Plumb `assertOnDeny` bool through `evaluateAndPrintResult` and the two printers (always passed `true`)
  - [x] Existing `wantError: true` deny-all cases stay green
- [x] Slice 2: Add `--assert` / `--no-assert` flags to `commonEvaluateOptions`, mark mutually exclusive, default = assert
  - [x] `evaluate input --policy deny-all --no-assert` exits 0, prints `RESULT: DENIED`
  - [x] `evaluate input --policy deny-all --assert` exits non-zero
  - [x] `evaluate input --policy deny-all` (neither flag) exits non-zero (default unchanged)
  - [x] `evaluate input --assert --no-assert ...` fails with cobra mutual-exclusion error
  - [x] `evaluate input --policy deny-all --no-assert --output json` emits `"allow": false`, exits 0
  - [x] Smoke test in `evaluate trail` and `evaluate trails` suites (`--no-assert` exit 0 + mutual exclusion); deferred run pending local Kosli server
- [ ] Slice 3: Help text and examples ← active
  - [ ] Update `evaluateLongDesc` and `evaluateInputLongDesc` exit-code line
  - [ ] Add `--no-assert` example to each command's `Example` block
  - [ ] Verify `kosli evaluate trail --help` shows new flags

## Fakes & contract tests for GitHub API integration

### Slice 1: FakeGitHubClient + contract tests (`internal/github`) ← active

- [x] `TestGitHubContract_Fake`: V2 returns PRs for commit with PRs
- [x] `TestGitHubContract_Fake`: V2 returns empty for commit with no PRs
- [x] `TestGitHubContract_Fake`: V2 returns error when Err is injected
- [x] `TestGitHubContract_Fake`: V1 returns PRs for commit with PRs
- [x] `TestGitHubContract_Fake`: V1 returns empty for commit with no PRs
- [x] `TestGitHubContract_Fake`: V1 returns error when Err is injected
- [x] `TestGitHubContract_RealGitHub`: same contract, env-gated on `KOSLI_GITHUB_TOKEN`

### Slice 2: Thread fake through command layer ← active

- [x] Add `ProviderAndLabel() (string, string)` to `types.PRRetriever` interface
- [x] Implement on `GithubConfig` → `("github", "pull request")`
- [x] Implement on `GitlabConfig` → `("gitlab", "merge request")`
- [x] Implement on `AzureConfig` → `("azure", "pull request")`
- [x] Implement on bitbucket `Config` → `("bitbucket", "pull request")`
- [x] Implement on `FakeGitHubClient` → `("github", "pull request")`
- [x] Replace reflection in `getGitProviderAndLabel` with `retriever.ProviderAndLabel()`
- [x] Inject fake in `assertPRGithub_test.go`
- [x] Inject fake in `attestPRGithub_test.go`

## Fakes & contract tests for cloud provider integrations (#758)

### Lambda (done — this PR)

- [x] Slice 1: Define `LambdaAPI` interface and refactor signatures
- [x] Slice 2: Contract test suite against real AWS
- [x] Slice 3: Build `FakeLambdaClient` that passes the contract
- [x] Slice 4: Fake-backed unit tests for filtering and pagination
- [x] Slice 5: Fake-backed unit tests for orchestration
- [x] Slice 6: Trim existing integration tests
- [x] Slice 7: Package-level factory + fake-backed command tests

### ECS (next)

- [ ] Define `ECSAPI` interface (`ListClusters`, `DescribeClusters`, `ListServices`, `ListTasks`, `DescribeTasks`) and refactor signatures
- [ ] Contract test suite against real AWS (env-gated)
- [ ] Build `FakeECSClient` that passes the contract (nested pagination: clusters → services → tasks)
- [ ] Fake-backed unit tests for filtering (cluster names, service names, regex, exclude patterns)
- [ ] Fake-backed unit tests for orchestration (concurrent cluster/service/task fetching, error propagation)
- [ ] `NewECSClientFunc` factory + inject fake into `snapshotECS_test.go` command tests
- [ ] Trim existing ECS integration tests to smoke tests
- [ ] Add ECS to `make test_contract_aws`

### S3

- [ ] Define `S3API` interface (decide: fake at paginator level or raw `ListObjectsV2` level)
- [ ] Contract test suite against real AWS (env-gated)
- [ ] Build `FakeS3Client` that passes the contract
- [ ] Fake-backed unit tests for path include/exclude filtering and digest computation
- [ ] `NewS3ClientFunc` factory + inject fake into `snapshotS3_test.go` command tests
- [ ] Trim existing S3 integration tests to smoke tests
- [ ] Add S3 to `make test_contract_aws`

### Azure Apps

- [ ] Define interfaces for ARM AppService + Azure Container Registry clients
- [ ] Contract test suite against real Azure (env-gated)
- [ ] Build fakes that pass the contracts
- [ ] Fake-backed unit tests for app listing, image fingerprinting, error propagation
- [ ] Factory + inject fakes into `snapshotAzureApps_test.go` command tests
- [ ] Trim existing Azure integration tests to smoke tests

### Docker

- [ ] Define `DockerAPI` interface (Pull, Push, Tag, Remove, Run, container operations)
- [ ] Contract test suite against real Docker daemon
- [ ] Build `FakeDockerClient` that passes the contract
- [ ] Fake-backed unit tests
- [ ] Factory + inject fake into `snapshotDocker_test.go` command tests
- [ ] Trim existing Docker integration tests to smoke tests

### Kubernetes

- [ ] Define interface for Kubernetes clientset operations (pod listing, namespace listing)
- [ ] Contract test suite against real cluster (KIND, env-gated)
- [ ] Build fake that passes the contract (semaphore pattern, namespace filtering)
- [ ] Fake-backed unit tests for filtering, large-scale concurrency, error propagation
- [ ] Factory + inject fake into `snapshotK8S_test.go` command tests
- [ ] Trim existing Kube integration tests to smoke tests
