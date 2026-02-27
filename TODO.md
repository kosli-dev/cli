# TODO

<!-- Each feature gets its own ## section below. -->
<!-- Only edit YOUR feature's section. Delete it after merging to main. -->

## kosli evaluate trail

- [ ] Slice 1: Skeleton `evaluate` parent + `evaluate trail` fetches trail JSON ← active
  - [x] missing trail name argument fails
  - [x] providing more than one argument fails
  - [x] missing --flow flag fails
  - [x] missing --api-token fails
  - [x] evaluating a non-existing trail fails
  - [x] evaluating an existing trail prints wrapped JSON with trail key
- [ ] Slice 2: `--policy` flag + OPA Rego evaluation
- [ ] Slice 3: JSON audit result output + `--format` flag
- [ ] Slice 4: `--show-input` flag
- [ ] Slice 5: `--attestations` flag + attestation enrichment
- [ ] Slice 6: `--output` file flag
- [ ] Slice 7: `kosli evaluate trails` (collection mode)
