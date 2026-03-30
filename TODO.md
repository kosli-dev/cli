# TODO

<!-- Each feature gets its own ## section below. -->
<!-- Only edit YOUR feature's section. Delete it after merging to main. -->

## CLI v3 — Breaking Changes & Cleanup

### Deprecated Commands to Remove

| Command | Replacement | File |
|---------|-------------|------|
| `kosli report artifact` | `kosli attest` commands | `cmd/kosli/reportArtifact.go` |
| `kosli snapshot server` (alias `directories`) | `kosli snapshot paths` | `cmd/kosli/snapshotServer.go` |

### Deprecated Flags to Remove

| Flag | Command | Replacement | File |
|------|---------|-------------|------|
| `--visibility` | `create flow` | Remove (org-level concern, not flow-level) | `cmd/kosli/createFlow.go` |
| `--registry-provider` | fingerprint commands | None ("no longer used") | `cmd/kosli/flags.go` |
| `--require-provenance` | `create environment` | Policies | `cmd/kosli/createEnvironment.go` |
| `--cluster` / `-C` | `snapshot ecs` | `--clusters` | `cmd/kosli/snapshotECS.go` |
| `--service-name` / `-s` | `snapshot ecs` | `--services` | `cmd/kosli/snapshotECS.go` |
| `--function-name` | `snapshot lambda` | `--function-names` | `cmd/kosli/snapshotLambda.go` |
| `--function-version` | `snapshot lambda` | None (non-functional, kept for compat) | `cmd/kosli/snapshotLambda.go` |
| `-e` (exclude shorthand) | `fingerprint`, `snapshot server` | `-x` | `cmd/kosli/fingerprint.go`, `cmd/kosli/snapshotServer.go` |

### Legacy Flow Creation Path

- [ ] Remove the `--template` string-slice flag and the legacy code path in `createFlow.go` (lines 138–153)
- [ ] Only keep the `--template-file` / `--use-empty-template` path, which hits `PUT /api/v2/flows/<org>/template_file`
- [ ] Remove the deprecated server endpoint `PUT /api/v2/flows/<org>` (see Legacy Flow Creation Endpoint below)

### API Compatibility Shim in `getArtifact.go`

- [ ] Remove `printArtifactAsTableWrapper()` — exists because the API returns an array for commit queries but a map for sha256 queries
- [ ] Fix API to consistently return an array, then remove the wrapper

### Config File Backward Compatibility

- [ ] Remove old default config location fallback in `root.go` (comment: "for backward compatibility with old default config location")

### Legacy Flow Handling in `assertArtifact.go`

- [ ] Remove legacy flow detection code (lines 220–221) that checks for "legacy_flow" and skips new attestation logic
- [ ] Note: the `legacy_flow` resolution type in the server (`src/model/policy_compliance_checker.py`) is **defined but never created** — it's dead code on both sides

### Tests & Docs

- [ ] Remove all test cases that verify deprecation warnings for removed commands/flags
- [ ] Update golden files / test output expectations
- [ ] Update doc generation if deprecation handling changes

---

### Server-Side Cleanup

#### Flask-to-FastAPI Migration Infrastructure

- [ ] Remove `RegexDispatcherMiddleware` (`src/fastapi_app/middleware.py`) — marked with `FASTAPI-POST-MIGRATION-CLEANUP` comment
- [ ] Remove bearer token prepending workaround (middleware lines 68–73)
- [ ] Remove PR attestation test-environment special case (middleware lines 25–34)
- [ ] Remove `is_api_path_migrated` / `is_public_resource` helpers (`src/auth/register_check_route_access.py`)
- [ ] Remove `migrated_apis` list and `api-migration-index` feature flag logic (`src/lib/feature_flags.py`)

#### Legacy Flow Creation Endpoint

- [ ] Remove deprecated `PUT /api/v2/flows/<org>` endpoint (Flask: `src/apis/v2/flow.py`, FastAPI: `src/fastapi_app/v2/flow.py`)
- [ ] Remove `LegacyCreateFlow` model and schema (`src/apis/schemas/flows/flows_declare_v2.json`)
- [ ] Remove `FlowTemplateConverter` (`src/model/flow_template_converter.py`) — only needed for inline-to-YAML conversion

#### V1 Flow Migration Code

- [ ] Remove migration scripts in `src/migrations/completed_migrations/migrate_v1_flows.py` and related undo/verify scripts (once all orgs confirmed migrated)
- [ ] Remove `has_trail` and `is_in_migration` properties from artifact models
- [ ] Remove `_legacy_compliance_state()` method and legacy status calculations

#### Flow Visibility Field

- [ ] Remove `visibility` from the flow model (`src/model/flow.py`) — access control is org-level only
- [ ] Remove `visibility` from flow creation payloads and API schemas
- [ ] Consider a migration to drop the field from existing flow documents

#### Rename `pipeline_id` → `flow_id` Throughout

- [ ] Rename `pipeline_id` to `flow_id` in artifact model (`src/model/artifact.py`), attestation model, trail model
- [ ] Remove `pipeline_name` → `flow_name` converter in FastAPI Pydantic models (`src/fastapi_app/models/artifacts.py`)
- [ ] Update database indexes in `src/documentdb/indexes.py` (many still reference `pipeline_id`)
- [ ] Requires a data migration for existing documents

#### Deprecated Event Types

- [ ] Remove legacy `started` and `changed` event types (`src/model/environment_consts.py`)
- [ ] Only keep the specific types: `started-compliant`, `started-non-compliant`, `started-unknown`, `became-compliant`, `became-non-compliant`, `scaled`, `updated-provenance`
- [ ] Update display priority logic and event validation

#### Legacy Attestation Types

- [ ] Remove `pull_request_legacy` attestation type support (`src/model/attestations_model/attestation.py`)
- [ ] Remove `generic_legacy` type conversion logic
- [ ] Update compliance checker to remove legacy type handling (`src/model/compliance_checker.py`)

#### Dead Code: `legacy_flow` Resolution Type

- [ ] Remove `legacy_flow` from `RuleResolution` type in `src/model/policy_compliance_checker.py` — defined but never created anywhere

#### Expect Deployment Remnants

- [ ] Remove deployment feature flag definitions in `src/lib/feature_flags.py` (indices 38–40)
- [ ] Review whether `ArtifactDeploymentEvent` in `src/model/trail_events.py` is still needed or can be cleaned up
- [ ] Clean up any orphaned deployment data models if no longer referenced by active code

#### Backward Compatibility Shims

- [ ] Remove deprecated `interval` parameter for environment snapshots (`src/fastapi_app/common/environment.py`) — replaced by `start_index`/`end_index`
- [ ] Remove `parse_snapshot_interval()` function
- [ ] Remove `legacy_approval_format` conversion code (`src/model/approvals.py`)

#### Authentication: Userfront Removal

- [ ] Remove `legacy_userfront_signin_page()` fallback (`src/auth/register_descope_auth.py`)
- [ ] Remove Userfront-related feature flags (`is-domain-using-descope-sso`, `is_using_descope`, `is-descope-session-validation-enabled`)
- [ ] Consolidate auth to Descope only

#### Experimental Features Flag

- [ ] Review `experimental_features_enabled` org field and `/api/v2/organizations/{org}/experimental_features` endpoint — remove if all orgs migrated

#### Legacy Snapshot Test Helpers

- [ ] Remove `test/helpers/unit/lib/legacy_k8s_snapshots.py` and legacy v1 snapshot test parametrization

---

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
