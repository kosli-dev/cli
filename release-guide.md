# Releasing

This document describes the CI/CD pipelines and release process for the Kosli CLI.

## CI pipelines overview

### On every push to any branch (`main.yml`)

Runs on every push to every branch:

1. **Test** — lint (golangci-lint), integration tests (full suite), Snyk code and dependency scans
2. **Docker** — builds and pushes `ghcr.io/kosli-dev/cli:<short-sha>` (linux/amd64 only), runs Snyk container scan and smoke test
3. **Kosli trail** — attestations are reported to the `cli` flow (on `main` and tags only)
4. **Slack** — notifies `#ci-failures` on failure (main branch only)

### On push to `main` with docs changes (`publish_docs.yml`)

Triggers when files under `docs.kosli.com/` change on `main` (excluding `content/client_reference/` and `content/helm/`):

- Copies the `docs.kosli.com/` directory to the `docs-main` branch using `git-publish-subdir-action`
- Preserves generated CLI reference docs and `metadata.json` from the last release (via `.clear-files`)

### On push to `main` with Helm chart changes (`helm-chart.yml`)

Triggers when files under `charts/` change on `main`:

1. Lints the Helm chart
2. Generates Helm docs (README + docs site content)
3. Packages and uploads the chart to the S3-hosted Helm repo (`charts.kosli.com`)
4. Opens a PR to merge the generated Helm docs back into `main`

### Manual: publish branch docs (`publish_branch_docs.yml`)

Manually triggered from any branch — deploys docs to `staging-docs--kosli-docs.netlify.app` for preview.

## Release process

A release is triggered by pushing a semver tag:

```bash
make release tag=v2.x.y
```

This validates the working tree is clean and up to date with the remote, creates an annotated tag, and pushes it. The `release.yml` workflow then runs:

### 1. Pre-build

Extracts the tag version and prepares Kosli trail metadata.

### 2. Kosli init

Creates a trail on the `cli-release` flow with the release version as the trail name.

### 3. Never-alone trail

Verifies the "never alone" policy — ensures commits had peer review.

### 4. Test

Full test suite (same as the `main` pipeline): lint, integration tests, Snyk code and dependency scans.

### 5. GoReleaser

Builds multi-platform binaries (darwin/linux/windows x amd64/arm64/arm) and creates a GitHub Release with:
- Archive files (`.tar.gz`, `.zip` for Windows)
- Linux packages (`.deb`, `.rpm`)
- Publishes packages to [Gemfury](https://gemfury.com/) (`push.fury.io/kosli/`)

### 6. Binary provenance (per-platform matrix)

For each binary artifact:
- Generates GitHub Sigstore build provenance attestation
- Generates SBOM (both SPDX and CycloneDX formats)
- Reports artifact + SBOM attestations to Kosli

### 7. Docker

Builds and pushes `ghcr.io/kosli-dev/cli:<tag>` (linux/amd64 + linux/arm64):
- GitHub Sigstore build provenance attestation
- SBOM generation and attestation
- Snyk container scan
- Smoke test (verifies the image can connect to Kosli)
- All results attested to Kosli

### 8. Homebrew

Opens a PR to `Homebrew/homebrew-core` to update the `kosli-cli` formula (skipped for pre-releases).

### 9. Docs generation

- Runs `make legacy-ref-docs` then `make cli-docs` to generate CLI reference markdown
- Writes `metadata.json` with the release version
- Pushes the full `docs.kosli.com/` directory to the `docs-main` branch (no `.clear-files` — replaces everything)

### 10. Reporter dispatches

Triggers downstream repository workflows to update:
- `kosli-dev/terraform-aws-evidence-reporter` — evidence reporter Lambda package
- `kosli-dev/terraform-aws-kosli-reporter` — environment reporter package
- Uploads a new CLI Lambda layer to AWS

### 11. Failure notification

Sends a Slack notification to `#ci-failures` if any job fails.

## Pre-release tags

Tags containing a hyphen (e.g., `v2.12.0-rc1`) are treated as pre-releases:
- The Homebrew formula PR is skipped
- GoReleaser marks the GitHub Release as a pre-release

## Docs hosting

The docs site is served from the `docs-main` branch (likely via Netlify). Content reaches that branch through two paths:

1. **Static content** — pushed on every merge to `main` that touches `docs.kosli.com/` (preserving generated files)
2. **Generated CLI reference + version metadata** — pushed only during a release (full replacement)

## Local development

```bash
make hugo          # Build CLI docs + Helm docs + run Hugo dev server (port 1515)
make hugo-local    # Build CLI docs only + run Hugo dev server (port 1515)
make cli-docs      # Regenerate CLI reference markdown from the built binary
make helm-docs     # Regenerate Helm chart docs
make check-links   # Build site and check for broken links
```
