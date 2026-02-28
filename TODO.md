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
  - [x] Slice 5b: Rehydration
  - [ ] Slice 5c: `--attestations` filtering ← active
    ### Unit tests — `FilterAttestations`
    - [ ] nil filters returns trail unchanged
    - [ ] empty filters slice returns trail unchanged
    - [ ] plain name keeps only matching trail-level attestation
    - [ ] plain name removes non-matching trail-level attestations
    - [ ] dot-qualified name keeps only matching artifact-level attestation
    - [ ] dot-qualified name: unmentioned artifact gets empty attestations_statuses
    - [ ] mixed filters: trail-level and artifact-level both applied
    - [ ] filters with no matches leave all attestations_statuses empty
    ### Integration tests
    - [ ] `--attestations trail-att` output only has trail-att in trail-level
    - [ ] `--attestations cli.art-att` output only has art-att in cli's attestations
    - [ ] Rego policy referencing filtered-in attestation passes
- [ ] Slice 6: `--output` file flag
- [ ] Slice 7: `kosli evaluate trails` (collection mode)
