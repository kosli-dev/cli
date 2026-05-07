# TODO

<!-- Each feature gets its own ## section below. -->
<!-- Only edit YOUR feature's section. Delete it after merging to main. -->

## Clean up old docs generation (kosli-dev/docs#42)

- [x] Slice 1: Remove old doc workflows (`publish_docs.yml`, `publish_branch_docs.yml`)
- [x] Slice 2: Remove `docs-gen` job from `release.yml` + helm-chart.yml docs.kosli.com line
- [x] Slice 3: Remove Makefile targets + scripts (`bin/test_docs_cmds.sh`, `hack/generate-old-versions-docs.sh`)
- [x] Slice 4: Remove `docs.kosli.com/` directory
- [x] Slice 5: Remove Hugo formatter (`internal/docgen/hugo.go`, `hugo_test.go`), update `docs.go` default
- [x] Slice 6: Clean up references (`.gitignore`, `.clear-files`, `.htmltest.yml`, README, dev-guide, release-guide) — closes kosli-dev/docs#42

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
- [x] Slice 3: Help text and examples
  - [x] Update `evaluateLongDesc` and `evaluateInputLongDesc` exit-code section
  - [x] Add `--no-assert` example to each command's `Example` block
  - [x] Verify `kosli evaluate trail --help` shows new flags

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

## perf: compact JSON for non-multipart request bodies (#825)

### Slice 1: Switch MarshalIndent → Marshal and verify

- [x] Test: non-multipart JSON request body is compact (no indentation)
- [x] Fix: change `json.MarshalIndent` → `json.Marshal` on line 122 of requests.go
- [x] Fix: update `PayloadOutput` to pretty-print non-multipart body for debug/dry-run logging
- [x] Verify: all request tests pass

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

## `kosli snapshot cloud-run`: enumerate idle Cloud Run Jobs (kosli-dev/server#4986)

Goal: in addition to Cloud Run Services + revisions, list Cloud Run Jobs in
the project/region and emit one artifact per Job, identified by the image at
`Template.Template.Containers[0].Image`. Idle Jobs (no running Execution) must
appear. Both CLI and server API are hidden, so the wire format is being
restructured at the same time.

Decisions locked in during planning:

- Payload kept flat to match the convention used by ECS, K8S, Lambda, etc.
  (server team rejected an earlier nested-`cloud_run_context` proposal).
  Each artifact carries top-level `kind: "service" | "job"` discriminator
  alongside `service_name` / `revision_name` / `job_name`.
- Filter flags renamed: `--services` → `--include`, `--services-regex` →
  `--include-regex`. `--exclude` / `--exclude-regex` unchanged. Filter applies
  uniformly to Service names and Job names.
- New `Job` struct mirrors `Revision` shape: `Name`, `Digests`, `CreatedAt`.
- Success log message changed from "[N] revisions were reported" to
  "[N] artifacts were reported".

Server-side schema change to accept the new payload shape needs to land
before live test-bed snapshots succeed; CLI tests fake HTTP and stay green
independently.

### Slice 1: Payload schema refactor (services-only, pure refactor)

Restructure `RevisionData` → `ArtifactData` with flat top-level fields
(`Kind`, `ServiceName`, `RevisionName`, `JobName`). No new behaviour; existing
Service-only output keeps working with the new wire format.

- [x] `internal/cloudrun/payload_test.go`: rewrite existing tests to assert on
      flat `art.Kind == "service"`, `art.ServiceName`, `art.RevisionName`
- [x] `internal/cloudrun/payload_test.go`: add
      `TestToEnvRequest_SerializesFlatFields_JSON` to lock the wire format
- [x] `cmd/kosli/snapshotCloudRun_test.go`: update `goldenRegex` in test 05
      to assert the flat `kind` / `service_name` shape
- [x] `cmd/kosli/snapshotCloudRun_test.go`: skip
      `TestSnapshotCloudRunCmd_HappyPathReportsToServer` pending server-side
      schema update (kosli-dev/server#4986)

### Slice 2: Internal Jobs primitives (`internal/cloudrun`)

Add `Job` domain type, `listJobs` on `apiClient`, `ListJobs` on `Client`,
production `gcpAPI.listJobs` using `run.NewJobsClient`. `Close()` extended to
also close the jobs client. No command wiring yet.

### Slice 3: Wire Jobs end-to-end + log message

Extend `cloudRunLister` with `ListJobs`. In `run()`, list Jobs, apply the same
name filter, pass into `ToEnvRequest(services, jobs)`. Change success log
from "[N] revisions were reported" to "[N] artifacts were reported".

### Slice 4: Flag rename

`--services` → `--include`, `--services-regex` → `--include-regex`. Update
mutually-exclusive pairs and error messages.

### Slice 5: Help text / long description

Update `snapshotCloudRunShortDesc` / `snapshotCloudRunLongDesc` to mention
Jobs alongside Services. Verify `kosli snapshot cloud-run --help` reads
correctly.
