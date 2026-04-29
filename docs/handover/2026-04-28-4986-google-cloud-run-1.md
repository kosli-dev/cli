# Handover: 4986-google-cloud-run-1

> **Last updated:** 2026-04-28
> **Branch:** `4986-google-cloud-run-1`
> **Ticket:** https://github.com/kosli-dev/server/issues/4986
> **Collaborators:** Tore Martin Hagen (engineer), Claude (claude-opus-4-7)

---

## Problem Definition

Add Google Cloud Run as a runtime environment type for `kosli snapshot`, mirroring the existing `kosli snapshot ecs` pattern. Customers deploying on GCP Cloud Run cannot currently use `kosli snapshot` to report what's running in their environments.

The first branch (`-1`) covers the CLI skeleton only — registering `kosli snapshot cloud-run` as a hidden command with arg/flag validation, no GCP API calls or HTTP yet. Later branches will add the GCP client wrapper, the end-to-end happy path, filtering flags, multi-revision handling, auth UX, and docs.

Constraints / acceptance criteria for the overall feature:
- Authentication via Application Default Credentials (ADC) — `GOOGLE_APPLICATION_CREDENTIALS`, `gcloud auth application-default login`, GCE/GKE metadata, Workload Identity Federation.
- Scope: project + region (Cloud Run has no cluster concept). Services only — Jobs are out of scope for this feature.
- Filtering flags: `--services`, `--services-regex`, `--exclude`, `--exclude-regex`.
- Server-side support for the `cloud-run` env type is a separate workstream.

GCP test environment (provisioned, ADC already configured on this machine):
- Project: `hello-world-cli-demo` (#429671251962)
- Region: `europe-west1`
- Service: `hello-world` at `https://hello-world-saxojpsd4a-ew.a.run.app`

---

## Decisions Made

- The CLI command stays `Hidden: true` and forces `global.DryRun = true` until the feature is complete, so we can iterate without exposing it to customers and without risk of accidental writes against the server. Both locks are removed in a later slice once the end-to-end path and tests are in place.
- The first branch ships only the cobra skeleton (no GCP calls, no HTTP, no payload). The thinnest possible end-to-end wiring slice is deferred to a later branch so the GCP client wrapper can be designed and tested in isolation.
- Server-side `cloud-run` env type work is tracked separately. The CLI proceeds against a dry-run-only command until that lands; the CLI does not need to coordinate releases with the server side for this branch.
- Package named `internal/cloudrun` (not `gcprun`) to mirror the user-facing command name `snapshot cloud-run`. If GCP integrations expand later we'll either rename or split into `internal/gcp/run`.
- `internal/cloudrun` reports all revisions referenced in `service.traffic[]` regardless of percent (including 0%). Trade-off: matches the user's framing of "running or could run" without dragging in retired revisions that aren't currently configured for traffic; canary 90/10 splits surface both revisions naturally.
- Digest extraction follows the ECS pattern (`internal/aws/aws.go:670-693`): use a `@sha256:` substring if present, else leave the digest empty rather than calling Artifact Registry. Registry-lookup mode (analogous to Azure's `--digests-source acr`) is deferred until customers ask for it.
- Wire payload follows the server's `out-snapshot-examples.txt` reference: top-level `{"type": "cloud-run", "artifacts": [...]}` with camelCase per-artifact fields (`revisionName`, `serviceName`, `digests`, `creationTimestamp`). Endpoint is `report/cloud-run` (kebab-case, parallels `report/azure-apps`). Server-side endpoint does not yet exist; the forced dry-run means no network call is made. Initial design (Slice 3) added `project`/`region` per artifact mirroring ECS's `cluster_name`; reverted in Slice 3.5 because the doc specifies `extra="forbid"` on every Pydantic model and project/region are derivable from the URL + flags.
- Command depends on a local `cloudRunLister` interface and a package-level `newCloudRunClient` variable so tests can substitute a stub without touching ADC. The seam stays in `cmd/kosli/snapshotCloudRun.go` rather than being exposed from `internal/cloudrun` — keeps the public package surface minimal.
- Error classification (`Classify`) lives in `internal/cloudrun` (GCP knowledge belongs to the package) but is *applied* at the command layer, not inside `Client.ListServices`. Why: applying it inside `ListServices` would double-wrap real-call errors when the command also classified them, and bypass the friendly path entirely for stub-driven tests. Calling it once at the command boundary covers both real and stub error sources.

---

## Next Steps

Slice plan (each slice is a separate, independently-mergeable branch):

- [x] **Slice 1 (this branch):** Skeleton command — `cmd/kosli/snapshotCloudRun.go` (Hidden, forced dry-run, stub `RunE`), register in `snapshot.go`, arg/flag validation tests. Done 2026-04-28: 5 cmdTestCase tests passing, `make lint` clean, hidden from `snapshot --help` but reachable directly.
- [x] **Slice 2:** Internal `internal/cloudrun` package — wraps `cloud.google.com/go/run/apiv2` to list services in project+region; unit-tested with a fake. Done 2026-04-28: `Client.ListServices` returns `Service{Name, URI, Revisions}` with one `Revision{Name, Digests, CreatedAt}` per traffic-configured revision (any percent including 0%, with `LATEST` resolved via `LatestReadyRevision` and dupes removed). Digest extraction mirrors the ECS fallback (`@sha256:` parse, else empty string). 9 unit tests passing.
- [x] **Slice 3:** End-to-end happy path — wire the package into `RunE`, build the snapshot payload, POST to the server `cloud-run` endpoint (still dry-run only). Done 2026-04-28: command now calls `cloudrun.New` + `ListServices`, builds an `EnvRequest` via `ToEnvRequest(services)`, and submits PUT `report/cloud-run` via `kosliClient.Do` (dry-run forced, so no network call leaves the client). Tested against the real `hello-world-cli-demo` GCP project — emits a digest-pinned artifact for the running `hello-world` service.
- [x] **Slice 3.5:** Align payload with the server's snapshot-examples doc. Done 2026-04-28: added top-level `"type": "cloud-run"`; first renamed `service_name` → `serviceName` (matching the *doc*'s K8S/ECS examples); then on 2026-04-29 reverted the rename to `service_name` because the actual wire format used by the existing ECS CLI is snake_case (the doc disagrees with reality on ECS). Final convention: unique-ID field camelCase (`revisionName`), grouping field snake_case (`service_name`), `digests` and `creationTimestamp` as universal. Per-artifact `project`/`region` stay dropped (would be rejected by `extra="forbid"` once the server defines a `CloudRunReport` model).
- [x] **Slice 4:** Filtering flags — `--services`, `--services-regex`, `--exclude`, `--exclude-regex`. Done 2026-04-28: backed by `filters.ResourceFilterOptions` (same struct ECS uses); 4 mutex pairs validated in `PreRunE`. Filter is applied in the command after `cloudrun.ListServices` returns — services excluded by name still cost their revision-fetch round-trips. If that becomes a bottleneck, push the filter into `cloudrun.ListServices` so excluded services skip the per-revision API calls.
- [x] **Slice 5:** ~~Multi-revision / traffic splitting — handle services with multiple active revisions and services with no active revisions.~~ Dropped 2026-04-28: multi-revision (traffic splitting), `LATEST` resolution, and dedup were all completed in Slice 2; the only remaining edge case (services with no active revisions emitting a placeholder artifact) was deferred until the server-side wire contract is defined, since picking the format unilaterally now risks rework. Re-open as a small slice once the server contract lands.
- [x] **Slice 6:** Auth error UX — clear messages for ADC / `GOOGLE_APPLICATION_CREDENTIALS` failures and for missing project/region. Done 2026-04-28: `cloudrun.Classify(err, project, region)` maps gRPC `Unauthenticated` → ADC advice, `PermissionDenied` → `roles/run.viewer` advice, `NotFound` → "project or region not found"; other codes pass through. Auth message names all three credential sources (env var, `gcloud auth application-default login`, GCE/GKE metadata server / Workload Identity) since the production deployment is a GKE cron job. `cloudrun.New(ctx)` errors get a generic "GCP client setup failed" prefix.
- [ ] **Slice 7:** Unhide the command, lift the forced dry-run, update CLI reference docs and examples.
