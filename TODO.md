# TODO

<!-- Each feature gets its own ## section below. -->
<!-- Only edit YOUR feature's section. Delete it after merging to main. -->

## kosli evaluate trail

- [x] Slice 1: Skeleton `evaluate` parent + `evaluate trail` fetches trail JSON
- [x] Slice 2: `--policy` flag + OPA Rego evaluation
- [x] Slice 3: JSON audit result output + `--format` flag
- [x] Slice 4: `--show-input` flag
- [ ] Slice 5: `--attestations` flag + attestation enrichment
  - [x] Slice 5a: Array-to-map transform for `attestations_statuses`
  - [ ] Slice 5b: Rehydration ← active
    ### CollectAttestationIDs
    - [ ] nil input returns empty slice
    - [ ] trail with no compliance_status returns empty slice
    - [ ] collects ID from trail-level attestation
    - [ ] collects IDs from artifact-level attestation
    - [ ] skips entries with null/missing attestation_id
    - [ ] collects from both trail-level and artifact-level
    ### RehydrateTrail
    - [ ] nil details map leaves trail unchanged
    - [ ] empty details map leaves trail unchanged
    - [ ] merges detail fields into trail-level attestation
    - [ ] does not overwrite existing fields
    - [ ] merges detail fields into artifact-level attestation
    - [ ] attestation with no matching detail is left unchanged
    ### Integration
    - [ ] rehydrated trail-level attestation has html_url from detail
    - [ ] rehydrated artifact-level attestation has html_url from detail
    - [ ] Rego policy can reference rehydrated field
  - [ ] Slice 5c: `--attestations` filtering
- [ ] Slice 6: `--output` file flag
- [ ] Slice 7: `kosli evaluate trails` (collection mode)
