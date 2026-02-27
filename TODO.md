# TODO

<!-- Each feature gets its own ## section below. -->
<!-- Only edit YOUR feature's section. Delete it after merging to main. -->

## kosli evaluate trail

- [x] Slice 1: Skeleton `evaluate` parent + `evaluate trail` fetches trail JSON
- [x] Slice 2: `--policy` flag + OPA Rego evaluation
- [ ] Slice 3: JSON audit result output + `--format` flag ← active
  - [ ] `--policy allow-all --format json` prints JSON with `allow: true` and empty violations
  - [ ] `--policy deny-all --format json` prints JSON with `allow: false` and violations, exits 1
  - [ ] `--policy allow-all --format text` prints human-readable allowed text
  - [ ] `--policy deny-all --format text` prints human-readable denied text with violations, exits 1
  - [ ] `--policy allow-all` (no --format) defaults to text format
  - [ ] `--format json` without `--policy` is ignored (prints trail JSON as before)
  - [ ] `--format invalid` returns an error
- [ ] Slice 4: `--show-input` flag
- [ ] Slice 5: `--attestations` flag + attestation enrichment
- [ ] Slice 6: `--output` file flag
- [ ] Slice 7: `kosli evaluate trails` (collection mode)
