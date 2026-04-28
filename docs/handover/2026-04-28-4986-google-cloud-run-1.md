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

---

## Next Steps

Slice plan (each slice is a separate, independently-mergeable branch):

- [x] **Slice 1 (this branch):** Skeleton command — `cmd/kosli/snapshotCloudRun.go` (Hidden, forced dry-run, stub `RunE`), register in `snapshot.go`, arg/flag validation tests. Done 2026-04-28: 5 cmdTestCase tests passing, `make lint` clean, hidden from `snapshot --help` but reachable directly.
- [ ] **Slice 2:** Internal `internal/gcprun` package — wraps `cloud.google.com/go/run/apiv2` to list services in project+region; unit-tested with a fake.
- [ ] **Slice 3:** End-to-end happy path — wire the package into `RunE`, build the snapshot payload, POST to the server `cloud-run` endpoint (still dry-run only).
- [ ] **Slice 4:** Filtering flags — `--services`, `--services-regex`, `--exclude`, `--exclude-regex`.
- [ ] **Slice 5:** Multi-revision / traffic splitting — handle services with multiple active revisions and services with no active revisions.
- [ ] **Slice 6:** Auth error UX — clear messages for ADC / `GOOGLE_APPLICATION_CREDENTIALS` failures and for missing project/region.
- [ ] **Slice 7:** Unhide the command, lift the forced dry-run, update CLI reference docs and examples.
