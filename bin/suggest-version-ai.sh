#!/usr/bin/env bash
# Suggest next semver and changelog by sending the git diff to Claude.
# Does not rely on commit messages. Changelog will be attached as a header to release Changelog.
#
# Calls Claude through the Claude Code CLI (`claude -p`), so it runs under the
# releaser's own Claude login rather than a shared API key.
#
# Optional: CLAUDE_MODEL (default: claude-sonnet-5)
#
# Requires: claude (Claude Code CLI, logged in), jq
# Usage: bin/suggest-version-ai.sh [base_ref] [-o release_notes.md]
#   base_ref defaults to the latest git tag.
#   -o FILE  write changelog markdown to FILE (default: dist/release_notes.md)
#
# Output: bump (major|minor|patch), next_version (e.g. v1.3.0), and changelog file.

set -euo pipefail

BASE_REF=""
RELEASE_NOTES_FILE="dist/release_notes.md"
while [[ $# -gt 0 ]]; do
  case "$1" in
    -o) RELEASE_NOTES_FILE="$2"; shift 2 ;;
    *)  BASE_REF="$1"; shift ;;
  esac
done
BASE_REF="${BASE_REF:-$(git describe --tags --abbrev=0 2>/dev/null)}"
SUGGESTED_VERSION_FILE="$(dirname "$RELEASE_NOTES_FILE")/suggested_version"

if [ -z "$BASE_REF" ]; then
  echo "ERROR: No base ref. Pass a tag or branch, or create a tag first." >&2
  exit 1
fi

# Cap diff size to stay within context (avoid SIGPIPE when diff is large: head exits first)
MAX_DIFF_CHARS=50000
set +o pipefail
DIFF="$(git diff "$BASE_REF"..HEAD 2>/dev/null | head -c "$MAX_DIFF_CHARS")"
set -o pipefail

if ! command -v claude >/dev/null 2>&1; then
  echo "ERROR: Claude Code CLI (claude) not found. Install it and log in, or use: make release tag=vX.Y.Z" >&2
  exit 1
fi

# Remove stale outputs from a previous run so a failure partway through doesn't mislead the next invocation.
trap 'rm -f "$RELEASE_NOTES_FILE" "$SUGGESTED_VERSION_FILE"' EXIT

if [ -z "$DIFF" ]; then
  echo "No changes since $BASE_REF. Bump: patch (no change)." >&2
  CHANGELOG="No code changes since $BASE_REF."
  mkdir -p "$(dirname "$RELEASE_NOTES_FILE")"
  echo "$CHANGELOG" > "$RELEASE_NOTES_FILE"
  echo "patch"
  CURRENT="${BASE_REF#v}"
  if [[ "$CURRENT" =~ ^[0-9]+\.[0-9]+\.[0-9]+ ]]; then
    MAJOR="${CURRENT%%.*}"; REST="${CURRENT#*.}"; MINOR="${REST%%.*}"; PATCH="${REST#*.}"; PATCH="${PATCH%%[-+]*}"
    NEXT="v${MAJOR}.${MINOR}.$((PATCH+1))"
    if [ -n "$(git tag -l "$NEXT")" ]; then
      echo "ERROR: tag $NEXT already exists. Push it or use: make release tag=$NEXT" >&2
      exit 1
    fi
    echo "$NEXT" > "$SUGGESTED_VERSION_FILE"
    echo "$NEXT"
  fi
  trap - EXIT
  exit 0
fi

PROMPT="You are a release engineer. Given the following git diff for a CLI application (Kosli CLI), do two things.

Scope: Consider ONLY changes to the CLI itself—i.e. code under cmd/ and internal/ that affects user-facing commands, flags, and behavior. IGNORE all other changes when deciding the version and when writing the changelog:
- Ignore: documentation (docs*, *.md), Helm charts (charts/), CI/workflows (.github/), scripts (bin/, scripts/), tests (*_test.go, testdata/), Makefile, config files, and any other non-CLI code.
- If the diff contains only ignored changes, recommend a patch bump and write a single short line for the changelog (e.g. \"No user-facing CLI changes.\").

1) Suggest the semantic version bump (based only on CLI changes):
   - major: Breaking changes (removed/renamed commands or flags, changed default behavior).
   - minor: New commands, flags, subcommands, or features.
   - patch: Bug fixes, refactors, internal or dependency updates; or no user-facing CLI changes.

2) Write a short changelog in markdown for the GitHub release body. Include only user-facing CLI changes. Use bullet points; be concise; no preamble.
   - Structure the changelog with section headers (e.g. \"# Breaking changes\", \"# New features\", \"# Bug fixes\" or \"# Improvements\") and list items under each header. Use only headers that have at least one change—omit any section that would be empty.
   - Do not write placeholder lines under any header (no \"No other changes\", \"No user-facing CLI changes in this release\", or similar). If there are no CLI changes at all, output a single short line only (no headers).

Reply in this exact format (no other text before or after):
BUMP: major|minor|patch
---CHANGELOG---
<markdown changelog here>"

CLAUDE_MODEL="${CLAUDE_MODEL:-claude-sonnet-5}"
WORK_DIR=$(mktemp -d)
trap 'rm -rf "$WORK_DIR"; rm -f "$RELEASE_NOTES_FILE" "$SUGGESTED_VERSION_FILE"' EXIT
printf '%s\n\n--- diff ---\n\n%s' "$PROMPT" "$DIFF" > "$WORK_DIR/input"

# Run from an empty directory with no tools so the repo's CLAUDE.md and the
# working tree don't leak into what should be a pure diff-in, text-out call.
RESPONSE=$(cd "$WORK_DIR" && claude -p \
  --model "$CLAUDE_MODEL" \
  --tools "" \
  --output-format json \
  --no-session-persistence \
  --strict-mcp-config \
  --disable-slash-commands \
  < input) || true

CONTENT=$(echo "$RESPONSE" | jq -r 'select(.is_error == false) | .result // empty' 2>/dev/null) || true
if [ -z "$CONTENT" ]; then
  echo "ERROR: claude -p failed or returned no content. Response:" >&2
  echo "$RESPONSE" >&2
  exit 1
fi

set +o pipefail
BUMP=$(echo "$CONTENT" | tr '[:upper:]' '[:lower:]' | grep -oE 'major|minor|patch' | head -1)
set -o pipefail
case "$BUMP" in
  major|minor|patch) ;;
  *)
    echo "WARN: Could not parse bump (got: $CONTENT). Defaulting to patch." >&2
    BUMP=patch
    ;;
esac

CHANGELOG_MARKER='---CHANGELOG---'
set +o pipefail
if echo "$CONTENT" | grep -qF -- "$CHANGELOG_MARKER"; then
  CHANGELOG=$(echo "$CONTENT" | sed -n "/${CHANGELOG_MARKER}/,\$ p" | tail -n +2)
else
  CHANGELOG=$(echo "$CONTENT" | sed -n '2,$ p')
fi
set -o pipefail
mkdir -p "$(dirname "$RELEASE_NOTES_FILE")"
echo "$CHANGELOG" > "$RELEASE_NOTES_FILE"

# Compute next version from current tag
CURRENT="${BASE_REF#v}"
if [[ "$CURRENT" =~ ^[0-9]+\.[0-9]+\.[0-9]+ ]]; then
  MAJOR="${CURRENT%%.*}"; REST="${CURRENT#*.}"; MINOR="${REST%%.*}"; PATCH="${REST#*.}"; PATCH="${PATCH%%[-+]*}"
  case "$BUMP" in
    major) NEXT="v$((MAJOR+1)).0.0" ;;
    minor) NEXT="v${MAJOR}.$((MINOR+1)).0" ;;
    patch) NEXT="v${MAJOR}.${MINOR}.$((PATCH+1))" ;;
  esac
else
  NEXT=""
fi

if [ -n "$NEXT" ]; then
  if [ -n "$(git tag -l "$NEXT")" ]; then
    echo "ERROR: tag $NEXT already exists. Push it or use: make release tag=$NEXT" >&2
    exit 1
  fi
  echo "$NEXT" > "$SUGGESTED_VERSION_FILE"
fi
echo "Suggested bump: $BUMP (from diff since $BASE_REF)" >&2
echo "Next version:  $NEXT" >&2
echo "Changelog:     $RELEASE_NOTES_FILE" >&2
trap - EXIT
echo "$BUMP"
echo "$NEXT"
