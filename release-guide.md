# Releasing

This document describes the CI/CD pipelines and release process for the Kosli CLI.

## CI pipelines overview

### On every push to any branch (`main.yml`)

Runs on every push to every branch:

1. **Test** — lint (golangci-lint), integration tests (full suite), Snyk code and dependency scans
2. **Docker** — builds and pushes `ghcr.io/kosli-dev/cli:<short-sha>` (linux/amd64 only), runs Snyk container scan and smoke test
3. **Kosli trail** — attestations are reported to the `cli` flow (on `main` and tags only)
4. **Slack** — notifies `#ci-failures` on failure (main branch only)

### On push to `main` with Helm chart changes (`helm-chart.yml`)

Triggers when files under `charts/` change on `main`:

1. Lints the Helm chart
2. Generates Helm docs (chart README)
3. Packages and uploads the chart to the S3-hosted Helm repo (`charts.kosli.com`)
4. Opens a PR to merge the generated Helm docs back into `main`

## Release process

### Default: release with the Claude Code `release` skill

In Claude Code, on an up-to-date `main`, run `/release` (or ask Claude to cut a release). The skill at `.claude/skills/release/SKILL.md`:

1. Reads the commits and diffs since the last tag and picks out the user-facing CLI changes.
2. Suggests the next semver tag, with the changes that justify the bump.
3. Drafts `dist/release_notes.md` and revises it with you until you approve it.
4. After a final confirmation, runs `make release tag=vX.Y.Z`, which creates an annotated tag with the release notes as the tag body and pushes it. The `release.yml` workflow reads the notes from the tag body and passes them to GoReleaser, which puts them at the top of the GitHub Release.

It runs under your own Claude login; no Anthropic API key is involved.

### Release body layout

The release body has two parts:

1. The notes carried in the tag body — drafted with the `release` skill or written by hand, describing user-facing CLI changes.
2. GitHub's **What's Changed** list of merged PRs and the **Full Changelog** compare link, from GoReleaser's `github-native` changelog. Dependency bumps are filtered out by `.github/release.yml`.

A tag with no notes gets part 2 only.

### Without Claude Code: release with an explicit tag

Without Claude Code, run:

```bash
make release tag=v2.x.y
```

This checks the branch is up to date, creates an annotated tag (using `dist/release_notes.md` as the tag body if that file exists, otherwise the version string), and pushes it. No prompt, so check that any `dist/release_notes.md` left over from an earlier release is gone or rewritten first. The release workflow runs as above; if the tag has no notes, the release body is the auto-generated **What's Changed** list alone.

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

### 9. Docs repo dispatch

Sends a `cli-release` event to `kosli-dev/docs` so the Mintlify docs site can regenerate CLI reference pages.

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

Documentation is maintained in the [kosli-dev/docs](https://github.com/kosli-dev/docs) repo and served via Mintlify. CLI reference pages are regenerated automatically when a `cli-release` dispatch event is received.
