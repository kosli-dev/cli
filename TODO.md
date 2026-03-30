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
