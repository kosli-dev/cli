---
title: "20260302 - Client-side enrichment pipeline for trail data"
description: "Transform, filter, and rehydrate trail data client-side to compensate for missing server API"
status: "Proposed"
date: "2026-03-02"
---

# 20260302 - Client-side enrichment pipeline for trail data

## Overview

Transform trail API responses client-side through a three-stage pipeline (transform, filter, rehydrate) to make attestation data accessible and useful for Rego policy evaluation.

## Context

The trail API (`/api/v2/trails/{org}/{flow}/{trail}`) returns compliance data with attestations as arrays. This shape is awkward for Rego policies — array indexing requires iteration, whereas map lookups by attestation name are natural and readable.

Additionally, the trail response only contains summary attestation data (name, compliance status, ID). Full attestation details (e.g. `html_url`, `origin_url`, `user_data`) require separate API calls per attestation.

There is no "trail status", "trail slots", or "trail moment" API that provides attestation data in a policy-friendly shape with full detail. Until such an API exists, the CLI must bridge the gap.

## Decision Drivers

- Rego policies need to reference attestations by their literal names (e.g. `input.trail.compliance_status.attestations_statuses["pull-request"]` for an attestation named `pull-request`)
- Policies may need full attestation detail data (not just the summary in the trail response)
- No server-side API provides this data in a policy-ready, map-based shape yet
- The enrichment should be transparent — users write policies against the enriched shape without knowing the pipeline exists

## Options Considered

### Option 1: Wait for server-side API

Wait for a dedicated API endpoint that returns trail data in a policy-friendly shape with full attestation details.

**Pros:**

- Clean solution — no client-side transformation needed
- Server can optimise the query (single round-trip)
- All clients get the same data shape

**Cons:**

- Blocks the evaluate feature entirely
- Server-side API timeline is unknown

### Option 2: Client-side enrichment pipeline

Transform the trail API response in the CLI before passing it to OPA.

**Pros:**

- Unblocks the feature now
- Users get a clean policy authoring experience
- Pipeline stages are independently testable
- Can be simplified or removed when a better API exists

**Cons:**

- N+1 API calls for rehydration (one per attestation with an ID)
- Client bears the transformation cost
- Data shape for policies is defined by CLI code, not the API contract

### Option 3: Require users to work with raw API shape

Pass the trail API response directly to Rego without transformation.

**Pros:**

- No client-side logic needed
- Policy input matches the API exactly

**Cons:**

- Poor developer experience — array iteration in Rego is verbose
- No access to attestation detail data
- Policies would break if the API shape changes anyway

## Decision

Implement a three-stage client-side enrichment pipeline:

1. **Transform** — Convert `attestations_statuses` arrays to maps keyed by `attestation_name`, at both trail and artifact levels. This enables Rego policies to use dot-notation access.

   The trail API returns attestations as an array:

   ```json
   {
     "compliance_status": {
       "attestations_statuses": [
         {"attestation_name": "pull-request", "is_compliant": true, "attestation_id": "abc123"},
         {"attestation_name": "unit-test", "is_compliant": true, "attestation_id": "def456"}
       ]
     }
   }
   ```

   After transformation, this becomes a map keyed by name:

   ```json
   {
     "compliance_status": {
       "attestations_statuses": {
         "pull-request": {"attestation_name": "pull-request", "is_compliant": true, "attestation_id": "abc123"},
         "unit-test": {"attestation_name": "unit-test", "is_compliant": true, "attestation_id": "def456"}
       }
     }
   }
   ```

   This allows Rego policies to reference attestations directly by name (e.g. `input.trail.compliance_status.attestations_statuses.pull_request.is_compliant`) instead of iterating an array.

2. **Filter** — When `--attestations` is specified, remove attestations not in the filter list. Plain names filter trail-level attestations; dot-qualified names (e.g. `artifact.attestation`) filter artifact-level attestations. Filtering happens before rehydration so we only fetch details for attestations that survive the filter.

3. **Rehydrate** — Collect `attestation_id` values from the filtered data, fetch full details via `/api/v2/attestations/{org}?attestation_id={id}`, and merge detail fields into each attestation entry (without overwriting existing summary fields).

**Rationale:** This is a workaround for missing server-side API. The pipeline makes policy authoring natural while keeping the transformation contained in `internal/evaluate/transform.go`. When a policy-friendly API exists, the pipeline can be simplified or removed without changing the policy contract.

**Trade-offs:** We accept N+1 API calls and client-side transformation cost in exchange for unblocking the feature with a good policy authoring experience.

## Consequences

**Positive:**

- Policies can reference attestations by name with dot-notation
- Full attestation detail data is available to policies
- The `--attestations` flag lets users scope policy evaluation to relevant attestations
- Pipeline stages are independently unit-tested

**Negative:**

- N+1 API calls per trail for rehydration (one per attestation)
- The policy input shape is defined by client-side code, not a stable API contract
- Pipeline adds complexity that should eventually be removed

**Neutral:**

- `walkTrailAttestations()` provides a single traversal abstraction used by all pipeline stages
- The pipeline is encapsulated in `internal/evaluate/transform.go` with `fetchAndEnrichTrail()` orchestrating it in `cmd/kosli/evaluateHelpers.go`

## Related Decisions

- [20260302-client-side-policy-evaluation](20260302-client-side-policy-evaluation.md) — the enrichment pipeline exists to serve client-side OPA evaluation
