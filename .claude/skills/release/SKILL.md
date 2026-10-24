---
name: release
description: Cut a Kosli CLI release - suggest the next semver tag from the changes since the last tag, draft the release notes with the user, then tag and push via `make release`. Use when asked to "release", "cut a release", "tag a new version", or "draft release notes".
---

## Overview

A release is an annotated tag on `main`. Its message becomes the top of the GitHub Release:
`release.yml` reads it and hands it to GoReleaser, which appends the "What's Changed" list
of merged PRs. This skill does the judgment work (version bump, notes) and leaves the
mechanics to `make release tag=vX.Y.Z`, which creates the annotated tag from
`dist/release_notes.md` and pushes it. `make release` is the only way the tag gets
created; see the Releasing section of `CLAUDE.md` for why.

## Steps

**1. Check the starting point.** Run `git fetch origin --tags`. Confirm the current
branch is `main`, it matches `origin/main`, and the working tree has no tracked changes.
Find the base with `git describe --tags --abbrev=0`. If any check fails, stop and tell
the user what to fix. Done when `main` == `origin/main` and the base tag is known.

**2. Classify every change since the base.** List commits with
`git log --oneline <base>..HEAD`, then read the diffs of the ones that touch `cmd/` or
`internal/` (`git show <sha> -- cmd internal`). Only user-facing CLI changes count:
commands, flags, output, defaults, error behaviour, and fixes a user would notice.
Tests, `testdata/`, docs, `charts/`, `.github/`, `bin/`, scripts, the Makefile and
dependency bumps without behaviour change do not. Done when every commit is marked
user-facing or not, with a one-line reason.

**3. Pick the bump.** Base it only on the user-facing changes:

- **major**: a removed or renamed command or flag, or a changed default.
- **minor**: a new command, subcommand, flag or feature.
- **patch**: bug fixes, refactors, dependency updates, or no user-facing changes.

Compute the next tag from the base (`v2.48.2` + minor -> `v2.49.0`) and confirm it does
not exist yet with `git tag -l <next>`. A pre-release tag (`v2.49.0-rc1`) only when the
user asks for one.

**4. Draft `dist/release_notes.md`.** Write user-facing CLI changes only, as concise
bullets under `#` headings, using only the headings that have entries, in this order:
`# Breaking changes`, `# New features`, `# Improvements`, `# Bug fixes`. Each bullet
names the command or flag and what changed for the user. When there are no user-facing
changes, the file is a single line with no heading: `No user-facing CLI changes.`
Create `dist/` if needed and overwrite any file already there.

**5. Review with the user.** Show the suggested tag, the reasoning for the bump (the
user-facing changes that drove it), and the notes. Apply every edit they ask for to
`dist/release_notes.md` and show the result again. Done when the user explicitly
approves both the tag and the notes.

**6. Tag and push.** Pushing the tag starts the release pipeline, so ask once more:
"Create and push `<tag>`?". On a yes, run `make release tag=<tag>`. Then report the run
with `gh run list --workflow release.yml --limit 1`.
