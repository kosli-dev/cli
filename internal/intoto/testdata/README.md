# in-toto test fixtures

Everything here works offline. The signed fixtures are checked against
`trusted_root.json`, not against Sigstore's public-good instance, so no test
fetches a trust root.

## Regenerating

```bash
REGENERATE_INTOTO_FIXTURES=1 go test ./internal/intoto -run TestRegenerateFixtures
```

`TestRegenerateFixtures` in `fixtures_test.go` signs the statements in
`statements/` with a throwaway Sigstore: a Fulcio authority issuing
certificates that name a GitHub Actions workflow, and the Rekor log and
timestamp authority of sigstore-go's `testing/ca`. Every run makes new keys,
so commit the regenerated bundles and `trusted_root.json` together.

The bundles stay valid after the throwaway certificates expire, because a
Sigstore signature is checked against the time Rekor logged it.

## What each fixture is for

| Fixture | What it is |
| --- | --- |
| `artifact.bin` | The artifact every statement is about (sha256 `e889...60a8`). |
| `statements/*.json` | Unsigned statements, as buildx writes by default. The signed fixtures are made from these. |
| `trusted_root.json` | The trust root the signed fixtures verify against. |
| `github-provenance.sigstore.json` | SLSA v1 provenance from GitHub Actions (`actions/attest-build-provenance`), signed by `https://github.com/acme/widget/.github/workflows/build.yml@refs/heads/main`. |
| `rerun-provenance.sigstore.json` | The same build's provenance from attempt 2 of the run, logged a minute after attempt 1's. Its certificate names attempt 2 as the run. |
| `buildkit-provenance.sigstore.json` | SLSA v1 provenance from BuildKit, signed by the same workflow. |
| `slsa-v0.2-provenance.sigstore.json` | SLSA v0.2 provenance from the SLSA GitHub generator. |
| `spdx-sbom.sigstore.json`, `cyclonedx-sbom.sigstore.json` | SBOM attestations (`actions/attest-sbom`). |
| `wrong-subject-provenance.sigstore.json` | Validly signed, but about another artifact. |
| `wrong-signer-provenance.sigstore.json` | Validly signed by `https://github.com/mallory/...`. |
| `tampered-provenance.sigstore.json` | The commit in the statement was changed after signing. |
| `attestations.intoto.jsonl` | Three signed bundles, one per line: a vulnerability scan, provenance and a CycloneDX SBOM. |
| `newest-first.intoto.jsonl` | Two signed provenance bundles; the first was logged later. |
| `vuln-only.intoto.jsonl` | Only vulnerability scans, one signed and one not. |
| `wrong-signer-and-unsigned.intoto.jsonl` | Provenance signed by `https://github.com/mallory/...`, then the same provenance unsigned. |
| `key-signed-provenance.dsse.json`, `key.pub` | A DSSE envelope signed with a key rather than a certificate, and that key. |
| `unsigned-spdx-sbom.dsse.json` | A DSSE envelope with no signatures. |
