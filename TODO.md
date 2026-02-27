# TODO

<!-- Each feature gets its own ## section below. -->
<!-- Only edit YOUR feature's section. Delete it after merging to main. -->

## kosli evaluate trail

- [x] Slice 1: Skeleton `evaluate` parent + `evaluate trail` fetches trail JSON
- [x] Slice 2: `--policy` flag + OPA Rego evaluation
- [x] Slice 3: JSON audit result output + `--format` flag
- [ ] Slice 4: `--show-input` flag ← active
  - [ ] `--policy allow-all --format json --show-input` includes `input` key in JSON output
  - [ ] `--policy deny-all --format json --show-input` includes `input` key alongside allow/violations
  - [ ] `--policy allow-all --format text --show-input` prints input JSON after evaluation result
  - [ ] `--policy allow-all --format json` without `--show-input` does NOT include `input` key
  - [ ] `--show-input` without `--policy` is ignored (prints trail JSON as before)
- [ ] Slice 5: `--attestations` flag + attestation enrichment
- [ ] Slice 6: `--output` file flag
- [ ] Slice 7: `kosli evaluate trails` (collection mode)
