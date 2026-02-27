# TODO

<!-- Each feature gets its own ## section below. -->
<!-- Only edit YOUR feature's section. Delete it after merging to main. -->

## kosli evaluate trail

- [x] Slice 1: Skeleton `evaluate` parent + `evaluate trail` fetches trail JSON
- [ ] Slice 2: `--policy` flag + OPA Rego evaluation ← active
  - internal/evaluate unit tests:
    - [x] allow-all policy returns allow=true
    - [x] deny-all policy returns allow=false with violation
    - [x] policy missing `package policy` returns error
    - [x] policy missing `allow` rule returns error
    - [x] policy with syntax error returns error
  - command tests:
    - [x] --policy with allow-all policy exits 0
    - [x] --policy with deny-all policy exits 1
    - [x] --policy with non-existent file fails
    - [x] --policy with invalid rego fails
    - [x] without --policy prints wrapped JSON (existing behaviour preserved)
- [ ] Slice 3: JSON audit result output + `--format` flag
- [ ] Slice 4: `--show-input` flag
- [ ] Slice 5: `--attestations` flag + attestation enrichment
- [ ] Slice 6: `--output` file flag
- [ ] Slice 7: `kosli evaluate trails` (collection mode)
